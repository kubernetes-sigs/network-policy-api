package matcher

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

// PeerMatcherAdmin models an ANP or BANP rule, incorporating an ANP/BANP action and an ANP priority.
// NOTE: we only use the PodPeerMatcher out of all the PeerMatcher imlementations.
// This is because ANP and BANP only deal with Pod to Pod traffic, and do not deal with external IPs.
type PeerMatcherAdmin struct {
	*PodPeerMatcher
	Name            string
	effectFromMatch Effect
}

// NewPeerMatcherANP creates a PeerMatcherAdmin for an ANP rule
func NewPeerMatcherANP(peer *PodPeerMatcher, v Verdict, priority int, source string) *PeerMatcherAdmin {
	return &PeerMatcherAdmin{
		PodPeerMatcher: peer,
		Name:           source,
		effectFromMatch: Effect{
			PolicyKind: AdminNetworkPolicy,
			Priority:   priority,
			Verdict:    v,
		},
	}
}

// NewPeerMatcherBANP creates a new PeerMatcherAdmin for a BANP rule
func NewPeerMatcherBANP(peer *PodPeerMatcher, v Verdict, source string) *PeerMatcherAdmin {
	return &PeerMatcherAdmin{
		PodPeerMatcher: peer,
		Name:           source,
		effectFromMatch: Effect{
			PolicyKind: BaselineAdminNetworkPolicy,
			Verdict:    v,
		},
	}
}

// Effect models the effect of one or more v1/v2 NetPol rules on a peer
type Effect struct {
	PolicyKind
	// Priority is only used for ANP (there can only be one BANP)
	Priority int
	Verdict
}

type PolicyKind string

const (
	NetworkPolicyV1            PolicyKind = "NPv1"
	AdminNetworkPolicy         PolicyKind = "ANP"
	BaselineAdminNetworkPolicy PolicyKind = "BANP"
)

func NewV1Effect(allow bool) Effect {
	if allow {
		return Effect{NetworkPolicyV1, 0, Allow}
	}
	return Effect{NetworkPolicyV1, 0, None}
}

type Verdict string

// Verdicts for v1 NetPols are Allow or None.
// Verdicts for ANP are Allow, Deny, Pass, or None.
// Verdicts for BANP are Allow, Deny, or None.
const (
	// None is used to indicate that the peer did not match
	None Verdict = "None"
	// Allow is used to indicate that the peer allowed the traffic
	Allow Verdict = "Allow"
	// Deny is used for ANP/BANP to indicate that the peer explicitly denied the traffic
	Deny Verdict = "Deny"
	// Pass is used for ANP to indicate that the peer passes the traffic down to v1 NetPol
	Pass Verdict = "Pass"
)

func AdminActionToVerdict(action v1alpha1.AdminNetworkPolicyRuleAction) Verdict {
	switch action {
	case v1alpha1.AdminNetworkPolicyRuleActionAllow:
		return Allow
	case v1alpha1.AdminNetworkPolicyRuleActionDeny:
		return Deny
	case v1alpha1.AdminNetworkPolicyRuleActionPass:
		return Pass
	default:
		panic(errors.Errorf("unsupported ANP action %s", action))
	}
}

func BaselineAdminActionToVerdict(action v1alpha1.BaselineAdminNetworkPolicyRuleAction) Verdict {
	switch action {
	case v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow:
		return Allow
	case v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny:
		return Deny
	default:
		panic(errors.Errorf("unsupported BANP action %s", action))
	}
}
