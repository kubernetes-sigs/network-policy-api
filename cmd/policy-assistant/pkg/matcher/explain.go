package matcher

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/json"
	"github.com/mattfenwick/collections/pkg/slice"
	"golang.org/x/exp/slices"
	v1 "k8s.io/api/core/v1"
	"strings"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// peerProtocolGroup groups all anps and banps in single struct
type peerProtocolGroup struct {
	port     string
	subject  string
	policies map[string]*anpGroup
}

// dummy implementation of the interface so we add the struct to the targer peers
func (p *peerProtocolGroup) Matches(subject, peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return false
}

type anpGroup struct {
	name     string
	priority int
	effects  []string
	kind     PolicyKind
}

type SliceBuilder struct {
	Prefix   []string
	Elements [][]string
}

func (s *SliceBuilder) Append(items ...string) {
	s.Elements = append(s.Elements, append(s.Prefix, items...))
}

func (p *Policy) ExplainTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)

	table.SetHeader([]string{"Type", "Subject", "Source rules", "Peer", "Action", "Port/Protocol"})

	builder := &SliceBuilder{}
	ingresses, egresses := p.SortedTargets()
	builder.TargetsTableLines(ingresses, true)

	builder.Elements = append(builder.Elements, []string{"", "", "", "", "", ""})
	builder.TargetsTableLines(egresses, false)

	table.AppendBulk(builder.Elements)

	table.Render()
	return tableString.String()
}

func (s *SliceBuilder) TargetsTableLines(targets []*Target, isIngress bool) {
	var ruleType string
	if isIngress {
		ruleType = "Ingress"
	} else {
		ruleType = "Egress"
	}
	for _, target := range targets {
		sourceRules := target.SourceRules
		sourceRulesStrings := make([]string, 0, len(sourceRules))
		for _, rule := range sourceRules {
			sourceRulesStrings = append(sourceRulesStrings, string(rule))
		}
		slices.Sort(sourceRulesStrings)
		rules := strings.Join(sourceRulesStrings, "\n")
		s.Prefix = []string{ruleType, target.TargetString(), rules}

		if len(target.Peers) == 0 {
			s.Append("no pods, no ips", "NPv1: All peers allowed", "no ports, no protocols")
			continue
		}

		peers := groupAnbAndBanp(target.Peers)
		for _, p := range slice.SortOn(func(p PeerMatcher) string { return json.MustMarshalToString(p) }, peers) {
			switch t := p.(type) {
			case *AllPeersMatcher:
				s.Append("all pods, all ips", "NPv1: All peers allowed", "all ports, all protocols")
			case *PortsForAllPeersMatcher:
				pps := PortMatcherTableLines(t.Port, NetworkPolicyV1)
				s.Append("all pods, all ips", "NPv1: All peers allowed", strings.Join(pps, "\n"))
			case *IPPeerMatcher:
				s.IPPeerMatcherTableLines(t)
			case *PodPeerMatcher:
				s.Append(resolveSubject(t), "NPv1: All peers allowed", strings.Join(PortMatcherTableLines(t.Port, NewV1Effect(true).PolicyKind), "\n"))
			case *peerProtocolGroup:
				s.peerProtocolGroupTableLines(t)
			default:
				panic(errors.Errorf("invalid PeerMatcher type %T", p))
			}
		}

	}
}

func (s *SliceBuilder) IPPeerMatcherTableLines(ip *IPPeerMatcher) {
	peer := ip.IPBlock.CIDR + "\n" + fmt.Sprintf("except %+v", ip.IPBlock.Except)
	pps := PortMatcherTableLines(ip.Port, NetworkPolicyV1)
	s.Append(peer, "NPv1: All peers allowed", strings.Join(pps, "\n"))
}

func (s *SliceBuilder) peerProtocolGroupTableLines(t *peerProtocolGroup) {
	actions := []string{}

	anps := make([]*anpGroup, 0, len(t.policies))
	for _, v := range t.policies {
		if v.kind == AdminNetworkPolicy {
			anps = append(anps, v)
		}
	}
	if len(anps) > 0 {
		actions = append(actions, "ANP:")
		slices.SortFunc(anps, func(a, b *anpGroup) bool {
			return a.priority < b.priority
		})
		for _, v := range anps {
			if len(v.effects) > 1 {
				actions = append(actions, fmt.Sprintf("   pri=%d (%s): %s (ineffective rules: %s)", v.priority, v.name, v.effects[0], strings.Join(v.effects[1:], ", ")))
			} else {
				actions = append(actions, fmt.Sprintf("   pri=%d (%s): %s", v.priority, v.name, v.effects[0]))
			}
		}
	}

	banps := make([]*anpGroup, 0, len(t.policies))
	for _, v := range t.policies {
		if v.kind == BaselineAdminNetworkPolicy {
			banps = append(banps, v)
		}
	}
	if len(banps) > 0 {
		actions = append(actions, "BANP:")
		for _, v := range banps {
			if len(v.effects) > 1 {
				actions = append(actions, fmt.Sprintf("   %s (ineffective rules: %s)", v.effects[0], strings.Join(v.effects[1:], ", ")))
			} else {
				actions = append(actions, fmt.Sprintf("   %s", v.effects[0]))
			}
		}
	}

	s.Append(t.subject, strings.Join(actions, "\n"), t.port)
}

func PortMatcherTableLines(pm PortMatcher, kind PolicyKind) []string {
	switch port := pm.(type) {
	case *AllPortMatcher:
		return []string{"all ports, all protocols"}
	case *SpecificPortMatcher:
		var lines []string
		for _, portProtocol := range port.Ports {
			if portProtocol.Port == nil {
				lines = append(lines, "all ports on protocol "+string(portProtocol.Protocol))
			} else if portProtocol.Port.StrVal != "" {
				if kind == NetworkPolicyV1 {
					lines = append(lines, fmt.Sprintf("namedport '%s' on protocol %s", portProtocol.Port.StrVal, portProtocol.Protocol))
				} else {
					lines = append(lines, fmt.Sprintf("namedport '%s'", portProtocol.Port.StrVal))
				}
			} else {
				lines = append(lines, fmt.Sprintf("port %d on protocol %s", portProtocol.Port.IntVal, portProtocol.Protocol))
			}
		}
		for _, portRange := range port.PortRanges {
			lines = append(lines, fmt.Sprintf("ports [%d, %d] on protocol %s", portRange.From, portRange.To, portRange.Protocol))
		}
		return lines
	default:
		panic(errors.Errorf("invalid PortMatcher type %T", port))
	}
}

func groupAnbAndBanp(p []PeerMatcher) []PeerMatcher {
	result := make([]PeerMatcher, 0, len(p))
	groups := map[string]*peerProtocolGroup{}

	for _, v := range p {
		switch t := v.(type) {
		case *PeerMatcherAdmin:
			k := t.Port.GetPrimaryKey() + t.Pod.PrimaryKey() + t.Namespace.PrimaryKey()
			if _, ok := groups[k]; !ok {
				groups[k] = &peerProtocolGroup{
					port:     strings.Join(PortMatcherTableLines(t.PodPeerMatcher.Port, t.effectFromMatch.PolicyKind), "\n"),
					subject:  resolveSubject(t.PodPeerMatcher),
					policies: map[string]*anpGroup{},
				}
			}
			kg := t.Name
			if _, ok := groups[k].policies[kg]; !ok {
				groups[k].policies[kg] = &anpGroup{
					name:     t.Name,
					priority: t.effectFromMatch.Priority,
					effects:  []string{},
					kind:     t.effectFromMatch.PolicyKind,
				}
			}
			groups[k].policies[kg].effects = append(groups[k].policies[kg].effects, string(t.effectFromMatch.Verdict))
		default:
			result = append(result, v)
		}
	}

	groupResult := make([]*peerProtocolGroup, 0, len(groups))
	for _, v := range groups {
		groupResult = append(groupResult, v)
	}
	slices.SortFunc(groupResult, func(a, b *peerProtocolGroup) bool {
		if a.port == b.port {
			return a.subject < b.subject
		}
		return a.port < b.port
	})

	for _, v := range groupResult {
		result = append(result, v)
	}

	return result
}

func resolveSubject(nsPodMatcher *PodPeerMatcher) string {
	var namespaces string
	var pods string
	switch ns := nsPodMatcher.Namespace.(type) {
	case *AllNamespaceMatcher:
		namespaces = "all"
	case *LabelSelectorNamespaceMatcher:
		namespaces = kube.LabelSelectorTableLines(ns.Selector)
	case *SameLabelsNamespaceMatcher:
		namespaces = fmt.Sprintf("Same labels - %s", strings.Join(ns.labels, ", "))
	case *NotSameLabelsNamespaceMatcher:
		namespaces = fmt.Sprintf("Not Same labels - %s", strings.Join(ns.labels, ", "))
	case *ExactNamespaceMatcher:
		namespaces = ns.Namespace
	default:
		panic(errors.Errorf("invalid NamespaceMatcher type %T", ns))
	}

	switch p := nsPodMatcher.Pod.(type) {
	case *AllPodMatcher:
		pods = "all"
	case *LabelSelectorPodMatcher:
		pods = kube.LabelSelectorTableLines(p.Selector)
	default:
		panic(errors.Errorf("invalid PodMatcher type %T", p))
	}

	return fmt.Sprintf("Namespace:\n   %s\nPod:\n   %s", strings.TrimSpace(namespaces), strings.TrimSpace(pods))
}
