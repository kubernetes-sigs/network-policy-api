package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/exp/maps"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

// Policy represents ALL Policies in the cluster (i.e. all ANPs, BANPs, and v1 NetPols).
// A NetPol, ANP, or BANP is translated into an Ingress and/or Egress Target.
// The (primary) key for these Targets is a string representation of either:
// a) the Namespace and Pod Selector (for v1 NetPols)
// b) the Namespace Selector and Pod Selector (for ANPs and BANPs).
type Policy struct {
	Ingress map[string]*Target
	Egress  map[string]*Target
}

func NewPolicy() *Policy {
	return &Policy{Ingress: map[string]*Target{}, Egress: map[string]*Target{}}
}

func NewPolicyWithTargets(ingress []*Target, egress []*Target) *Policy {
	np := NewPolicy()
	np.AddTargets(true, ingress)
	np.AddTargets(false, egress)
	return np
}

func (p *Policy) SortedTargets() ([]*Target, []*Target) {
	key := func(t *Target) string { return t.GetPrimaryKey() }
	ingress := slice.SortOn(key, maps.Values(p.Ingress))
	egress := slice.SortOn(key, maps.Values(p.Egress))
	return ingress, egress
}

func (p *Policy) AddTargets(isIngress bool, targets []*Target) {
	for _, target := range targets {
		p.AddTarget(isIngress, target)
	}
}

func (p *Policy) AddTarget(isIngress bool, target *Target) {
	if target == nil {
		return
	}

	pk := target.GetPrimaryKey()
	var dict map[string]*Target
	if isIngress {
		dict = p.Ingress
	} else {
		dict = p.Egress
	}
	if prev, ok := dict[pk]; ok {
		combined := prev.Combine(target)
		dict[pk] = combined
	} else {
		dict[pk] = target
	}
}

func (p *Policy) TargetsApplyingToPod(isIngress bool, subject *InternalPeer) []*Target {
	var targets []*Target
	var dict map[string]*Target
	if isIngress {
		dict = p.Ingress
	} else {
		dict = p.Egress
	}
	for _, target := range dict {
		if target.Matches(subject) {
			targets = append(targets, target)
		}
	}
	return targets
}

// DirectionResult contains information about each rule of each v1/v2 NetPol on traffic in a single direction (ingress or egress)
type DirectionResult []Effect

// IsAllowed returns true if the traffic is allowed after accounting for all v1/v2 NetPols.
func (d DirectionResult) IsAllowed() bool {
	anp, npv1, banp := d.Resolve()
	if anp == nil && npv1 == nil && banp == nil {
		return true
	}

	if anp != nil {
		if anp.Verdict == Allow {
			return true
		}

		if anp.Verdict == Deny {
			return false
		}
	}

	if npv1 != nil {
		return npv1.Verdict == Allow
	}

	// if banp is nil, then anp was a pass and there are no npv1 rules
	return banp == nil || banp.Verdict != Deny
}

// Flow returns a string representation of the flow through ANP, v1 NetPol, and BANP.
// E.g. "[ANP] Pass -> [BANP] No-Op"
func (d DirectionResult) Flow() string {
	anp, npv1, banp := d.Resolve()

	flows := make([]string, 0)
	if anp != nil {
		if anp.Verdict == Allow {
			return "[ANP] Allow"
		}

		if anp.Verdict == Deny {
			return "[ANP] Deny"
		}

		if anp.Verdict == Pass {
			flows = append(flows, "[ANP] Pass")
		} else {
			flows = append(flows, "[ANP] No-Op")
		}
	}

	if npv1 != nil {
		if npv1.Verdict == Allow {
			flows = append(flows, "[NPv1] Allow")
		} else {
			flows = append(flows, "[NPv1] Dropped")
		}

		return strings.Join(flows, " -> ")
	}

	if banp != nil {
		if banp.Verdict == Allow {
			flows = append(flows, "[BANP] Allow")
		} else if banp.Verdict == Deny {
			flows = append(flows, "[BANP] Deny")
		} else {
			flows = append(flows, "[BANP] No-Op")
		}
	}

	return strings.Join(flows, " -> ")
}

// Resolve returns the final Effect on traffic for ANP, v1 NetPol, and BANP respectively.
// A nil Effect indicates that there are none of that PolicyKind
// or e.g. ANP allowed traffic before reaching v1 NetPol and BANP.
func (d DirectionResult) Resolve() (*Effect, *Effect, *Effect) {
	if d == nil {
		return nil, nil, nil
	}

	// 1. ANP rules
	var anpEffect *Effect
	for _, e := range d {
		if e.PolicyKind != AdminNetworkPolicy {
			continue
		}

		if anpEffect == nil {
			anpEffect = &Effect{
				PolicyKind: AdminNetworkPolicy,
				Verdict:    None,
				Priority:   maxInt,
			}
		}

		if e.Verdict != None && e.Priority < anpEffect.Priority {
			eCopy := e
			anpEffect = &eCopy
		}
	}

	if anpEffect != nil && (anpEffect.Verdict == Allow || anpEffect.Verdict == Deny) {
		return anpEffect, nil, nil
	}

	// 2. v1 NetPol rules
	haveV1NetPols := false
	for _, e := range d {
		if e.PolicyKind != NetworkPolicyV1 {
			continue
		}

		haveV1NetPols = true
		if e.Verdict == Allow {
			eCopy := e
			return anpEffect, &eCopy, nil
		}
	}

	if haveV1NetPols {
		v1NoMatch := NewV1Effect(false)
		return anpEffect, &v1NoMatch, nil
	}

	// 3. BANP rules
	var banpEffect *Effect
	for _, e := range d {
		if e.PolicyKind != BaselineAdminNetworkPolicy {
			continue
		}

		if banpEffect == nil {
			banpEffect = &Effect{
				PolicyKind: BaselineAdminNetworkPolicy,
				Verdict:    None,
			}
		}

		if e.Verdict != None {
			eCopy := e
			return anpEffect, nil, &eCopy
		}
	}

	return anpEffect, nil, banpEffect
}

// AllowedResult contains information to calculate the final result taken on traffic in a cluster
// taking into account all ANPs/BANPs and v1 NetPols.
type AllowedResult struct {
	Ingress DirectionResult
	Egress  DirectionResult
}

func (ar *AllowedResult) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	table.SetHeader([]string{"Type", "Action", "Target"})
	// NOTE: broken. This function only used by 1) "cyclonus analyze explain --mode=query-target" and 2) a trace-level log in "cyclonus generate"
	// Broken since we're changing the DirectionResult struct to use Effects instead of Allowing/DenyingTargets
	// addTargetsToTable(table, "Ingress", "Allow", ar.Ingress.AllowingTargets)
	// addTargetsToTable(table, "Ingress", "Deny", ar.Ingress.DenyingTargets)
	table.Append([]string{"", "", ""})
	// addTargetsToTable(table, "Egress", "Allow", ar.Egress.AllowingTargets)
	// addTargetsToTable(table, "Egress", "Deny", ar.Egress.DenyingTargets)
	table.SetFooter([]string{"Is allowed?", fmt.Sprintf("%t", ar.IsAllowed()), ""})

	table.Render()
	return tableString.String()
}

func (ar *AllowedResult) IsAllowed() bool {
	return ar.Ingress.IsAllowed() && ar.Egress.IsAllowed()
}

// IsTrafficAllowed returns:
// - whether the traffic is allowed
// - which rules allowed the traffic
// - which rules matched the traffic target
func (p *Policy) IsTrafficAllowed(traffic *Traffic) *AllowedResult {
	return &AllowedResult{
		Ingress: p.IsIngressOrEgressAllowed(traffic, true),
		Egress:  p.IsIngressOrEgressAllowed(traffic, false),
	}
}

func (p *Policy) IsIngressOrEgressAllowed(traffic *Traffic, isIngress bool) DirectionResult {
	var subject *TrafficPeer
	var peer *TrafficPeer
	if isIngress {
		subject = traffic.Destination
		peer = traffic.Source
	} else {
		subject = traffic.Source
		peer = traffic.Destination
	}

	// 1. if target is external to cluster -> allow
	//   this is because we can't stop external hosts from sending or receiving traffic
	if subject.Internal == nil {
		return nil
	}

	matchingTargets := p.TargetsApplyingToPod(isIngress, subject.Internal)

	// 2. No targets match => automatic allow
	if len(matchingTargets) == 0 {
		return nil
	}

	// 3. Check if any matching targets allow this traffic
	effects := make([]Effect, 0)
	for _, target := range matchingTargets {
		for _, m := range target.Peers {
			// check if m is a PeerMatcherAdmin
			e := NewV1Effect(true)
			matcherAdmin, ok := m.(*PeerMatcherAdmin)
			if ok {
				e = matcherAdmin.effectFromMatch
			}

			if !m.Matches(subject, peer, traffic.ResolvedPort, traffic.ResolvedPortName, traffic.Protocol) {
				e.Verdict = None
			}

			effects = append(effects, e)
		}
	}

	return effects
}

func (p *Policy) Simplify() {
	for _, ingress := range p.Ingress {
		ingress.Simplify()
	}
	for _, egress := range p.Egress {
		egress.Simplify()
	}
}
