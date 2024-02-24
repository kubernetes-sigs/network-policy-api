package matcher

import (
	"fmt"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

// string of the form "[policyKind] namespace/name"
type NetPolID string

func netPolID(p interface{}) NetPolID {
	switch p := p.(type) {
	case *networkingv1.NetworkPolicy:
		ns := p.Namespace
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		return NetPolID(fmt.Sprintf("[%s] %s/%s", NetworkPolicyV1, ns, p.Name))
	case *v1alpha1.AdminNetworkPolicy:
		ns := p.Namespace
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		return NetPolID(fmt.Sprintf("[%s] %s/%s", AdminNetworkPolicy, ns, p.Name))
	case *v1alpha1.BaselineAdminNetworkPolicy:
		ns := p.Namespace
		if ns == "" {
			ns = metav1.NamespaceDefault
		}
		return NetPolID(fmt.Sprintf("[%s] %s/%s", BaselineAdminNetworkPolicy, ns, p.Name))
	default:
		panic(fmt.Sprintf("invalid policy type %T", p))
	}
}

// Target represents ingress or egress for one or more NetworkPolicies.
// It can represent either:
// a) one or more v1 NetPols sharing the same Namespace and Pod Selector
// b) one or more ANPs/BANPs sharing the same Namespace Selector and Pod Selector.
type Target struct {
	SubjectMatcher
	SourceRules []NetPolID
	// Peers contains all matchers for a Target.
	// Order matters for rules in the same ANP or BANP.
	// Priority matters for rules in different ANPs.
	Peers map[string][]PeerMatcher
}

func (t *Target) String() string {
	return t.GetPrimaryKey()
}

func (t *Target) Simplify() {
	//t.Peers = Simplify(t.Peers)
}

// Combine creates a new Target combining the egress and ingress rules
// of the two original targets.  Neither input is modified.
// The Primary Keys of the two targets must match.
func (t *Target) Combine(other *Target) *Target {
	myPk := t.GetPrimaryKey()
	otherPk := other.GetPrimaryKey()
	if myPk != otherPk {
		panic(errors.Errorf("cannot combine targets: primary keys differ -- '%s' vs '%s'", myPk, otherPk))
	}

	peers := map[string][]PeerMatcher{}

	peers = combinePeers(peers, t.Peers)
	peers = combinePeers(peers, other.Peers)

	// ensure that we have only one banp after combining
	for k := range peers {
		var banp bool
		peers[k] = slice.Filter(func(a PeerMatcher) bool {
			switch t := a.(type) {
			case *PeerMatcherAdmin:
				if t.effectFromMatch.PolicyKind == BaselineAdminNetworkPolicy {
					if banp {
						return false
					}
					banp = true
					return true

				}
				return true
			default:
				return true
			}
		}, peers[k])

	}

	return &Target{
		SubjectMatcher: t.SubjectMatcher,
		Peers:          peers,
		SourceRules:    sets.New(t.SourceRules...).Insert(other.SourceRules...).UnsortedList(),
	}
}

func (t *Target) CombineCommonPeers(other *Target) {
	if other == nil || len(other.Peers) == 0 {
		return
	}

	rules := sets.New(t.SourceRules...)
	for k := range t.Peers {
		if _, ok := other.Peers[k]; ok {
			t.Peers[k] = append(t.Peers[k], other.Peers[k]...)
			rules.Insert(other.SourceRules...)
		}
	}
	t.SourceRules = rules.UnsortedList()
}

func combinePeers(dest map[string][]PeerMatcher, source map[string][]PeerMatcher) map[string][]PeerMatcher {
	for i, v := range source {
		if _, ok := dest[i]; !ok {
			dest[i] = []PeerMatcher{}
		}
		dest[i] = append(dest[i], v...)
	}
	return dest
}

// CombineTargetsIgnoringPrimaryKey creates a new v1 target from the given namespace and pod selector,
// and combines all the edges and source rules from the original targets into the new target.
func CombineTargetsIgnoringPrimaryKey(namespace string, podSelector metav1.LabelSelector, targets []*Target) *Target {
	if len(targets) == 0 {
		return nil
	}
	target := &Target{
		SubjectMatcher: NewSubjectV1(namespace, podSelector),
		Peers:          targets[0].Peers,
		SourceRules:    targets[0].SourceRules,
	}
	//for _, t := range targets[1:] {
	//	target.Peers = append(target.Peers, t.Peers...)
	//	target.SourceRules = append(target.SourceRules, t.SourceRules...)
	//}
	return target
}

// SubjectMatcher defines which Pods a ANP, BANP, or v1 NetPol applies to
type SubjectMatcher interface {
	// Matches returns true if the candidate satisfies the subject selector
	Matches(candidate *InternalPeer) bool
	// TargetString is used for printing in tables
	TargetString() string
	// GetPrimaryKey serializes the subject selector into a json-like string
	GetPrimaryKey() string
}

// SubjectV1 implements SubjectSelector for v1 NetPols
type SubjectV1 struct {
	primaryKey  string
	namespace   string
	podSelector metav1.LabelSelector
}

func NewSubjectV1(namespace string, podSelector metav1.LabelSelector) *SubjectV1 {
	return &SubjectV1{
		namespace:   namespace,
		podSelector: podSelector,
		primaryKey:  fmt.Sprintf(`{"Namespace": "%s", "PodSelector": %s}`, namespace, kube.SerializeLabelSelector(podSelector)),
	}
}

func (s *SubjectV1) Matches(candidate *InternalPeer) bool {
	return s.namespace == candidate.Namespace && kube.IsLabelsMatchLabelSelector(candidate.PodLabels, s.podSelector)
}

func (s *SubjectV1) TargetString() string {
	pods := kube.LabelSelectorTableLines(s.podSelector)
	if pods == "all" {
		pods = "all pods"
	}
	return fmt.Sprintf("namespace: %s\n%s", s.namespace, pods)
}

func (s *SubjectV1) GetPrimaryKey() string {
	return s.primaryKey
}

// SubjectAdmin implements SubjectSelector for ANPs/BANPs
type SubjectAdmin struct {
	subject    *v1alpha1.AdminNetworkPolicySubject
	primaryKey string
}

func NewSubjectAdmin(subject *v1alpha1.AdminNetworkPolicySubject) *SubjectAdmin {
	s := &SubjectAdmin{subject: subject}

	if (s.subject.Namespaces == nil && s.subject.Pods == nil) || (s.subject.Namespaces != nil && s.subject.Pods != nil) {
		// unexpected since there should be exactly one of Namespaces or Pods
		s.primaryKey = "invalid"
	} else if s.subject.Namespaces != nil {
		s.primaryKey = fmt.Sprintf(`{"Namespaces": "%s"}`, kube.SerializeLabelSelector(*s.subject.Namespaces))
	} else {
		s.primaryKey = fmt.Sprintf(`{"NamespaceSelector": %s, "PodSelector": %s}`, kube.SerializeLabelSelector(s.subject.Pods.NamespaceSelector), kube.SerializeLabelSelector(s.subject.Pods.PodSelector))
	}

	return s
}

func (s *SubjectAdmin) Matches(candidate *InternalPeer) bool {
	if (s.subject.Namespaces == nil && s.subject.Pods == nil) || (s.subject.Namespaces != nil && s.subject.Pods != nil) {
		// unexpected since there should be exactly one of Namespaces or Pods
		return false
	}

	if s.subject.Namespaces != nil {
		return kube.IsLabelsMatchLabelSelector(candidate.NamespaceLabels, *s.subject.Namespaces)
	}

	return kube.IsLabelsMatchLabelSelector(candidate.NamespaceLabels, s.subject.Pods.NamespaceSelector) &&
		kube.IsLabelsMatchLabelSelector(candidate.PodLabels, s.subject.Pods.PodSelector)
}

func (s *SubjectAdmin) TargetString() string {
	if s.subject.Namespaces != nil {
		return fmt.Sprintf("Namespace labels:\n%s", kube.LabelSelectorTableLines(*s.subject.Namespaces))
	} else {
		return fmt.Sprintf("Namespace labels:\n%s\nPod labels:\n%s", kube.LabelSelectorTableLines(s.subject.Pods.NamespaceSelector), kube.LabelSelectorTableLines(s.subject.Pods.PodSelector))
	}
}

func (s *SubjectAdmin) GetPrimaryKey() string {
	return s.primaryKey
}
