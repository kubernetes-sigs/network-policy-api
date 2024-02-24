package matcher

import (
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"strings"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

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

		for _, peers := range target.Peers {
			if len(peers) == 0 {
				s.Append("no pods, no ips", "NPv1: All peers allowed", "no ports, no protocols")
				continue
			}
			var subject string
			var ports string
			var action string

			anps := []*PeerMatcherAdmin{}
			banps := []*PeerMatcherAdmin{}

			for _, p := range peers {
				switch t := p.(type) {
				case *AllPeersMatcher:
					s.Append("all pods, all ips", "NPv1: All peers allowed", "all ports, all protocols")
				case *PortsForAllPeersMatcher:
					pps := PortMatcherTableLines(t.Port, NetworkPolicyV1)
					s.Append("all pods, all ips", "", strings.Join(pps, "\n"))
				case *IPPeerMatcher:
					s.IPPeerMatcherTableLines(t)
				case *PodPeerMatcher:
					s.PodPeerMatcherTableLines(t, NewV1Effect(true), "")
				case *PeerMatcherAdmin:
					subject = resolveSubject(t.PodPeerMatcher)
					ports = strings.Join(PortMatcherTableLines(t.PodPeerMatcher.Port, t.effectFromMatch.PolicyKind), "\n")

					switch t.effectFromMatch.PolicyKind {
					case AdminNetworkPolicy:
						anps = append(anps, t)
					case BaselineAdminNetworkPolicy:
						banps = append(banps, t)
					default:
						panic("This should not be possible")
					}
				}
			}

			if len(anps) > 1 {
				g := map[string]*AnpGroup{}
				for _, v := range anps {
					e := string(v.effectFromMatch.Verdict)
					if _, ok := g[v.Name]; !ok {
						g[v.Name] = &AnpGroup{
							name:     v.Name,
							priority: v.effectFromMatch.Priority,
							effects:  []string{},
						}
					}
					g[v.Name].effects = append(g[v.Name].effects, e)
				}

				groups := maps.Values(g)

				slices.SortFunc(groups, func(a, b *AnpGroup) bool {
					return a.priority > b.priority
				})

				r := []string{"ANP:"}
				for _, a := range groups {
					if len(a.effects[1:]) > 0 {
						r = append(r, fmt.Sprintf("   pri=%d (%s): %s (ineffective rules: %s)", a.priority, a.name, a.effects[0], strings.Join(a.effects[1:], ", ")))
					} else {
						r = append(r, fmt.Sprintf("   pri=%d (%s): %s", a.priority, a.name, a.effects[0]))
					}
				}
				action = strings.Join(r, "\n") + "\n"
			}

			if len(banps) >= 1 {
				g := map[string]*AnpGroup{}
				for _, v := range banps {
					e := string(v.effectFromMatch.Verdict)
					if _, ok := g[v.Name]; !ok {
						g[v.Name] = &AnpGroup{
							name:     v.Name,
							priority: v.effectFromMatch.Priority,
							effects:  []string{},
						}
					}
					g[v.Name].effects = append(g[v.Name].effects, e)
				}
				for _, a := range g {
					if len(a.effects[1:]) > 0 {
						action += fmt.Sprintf("BNP: %s (ineffective rules: %s)", a.priority, a.name, a.effects[0], strings.Join(a.effects[1:], ", "))
					} else {
						action += fmt.Sprintf("BNP: %s", a.effects[0])
					}
				}
			}
			s.Append(subject, action, ports)
		}

	}
}

type AnpGroup struct {
	name     string
	priority int
	effects  []string
}

func (s *SliceBuilder) IPPeerMatcherTableLines(ip *IPPeerMatcher) {
	peer := ip.IPBlock.CIDR + "\n" + fmt.Sprintf("except %+v", ip.IPBlock.Except)
	pps := PortMatcherTableLines(ip.Port, NetworkPolicyV1)
	s.Append(peer, "", strings.Join(pps, "\n"))
}

func (s *SliceBuilder) PodPeerMatcherTableLines(nsPodMatcher *PodPeerMatcher, e Effect, name string) {
	s.Append(resolveSubject(nsPodMatcher), priorityTableLine(e, name), strings.Join(PortMatcherTableLines(nsPodMatcher.Port, e.PolicyKind), "\n"))
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

func priorityTableLine(e Effect, name string) string {
	if e.PolicyKind == NetworkPolicyV1 {
		return "NPv1: All peers allowed"
	} else if e.PolicyKind == AdminNetworkPolicy {
		return fmt.Sprintf("   (pri=%d) %s: %s ", e.Priority, name, e.Verdict)
	} else if e.PolicyKind == BaselineAdminNetworkPolicy {
		return fmt.Sprintf("   (%s): %s", name, e.Verdict)
	} else {
		panic(errors.Errorf("Invalid effect %s", e.PolicyKind))
	}
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
