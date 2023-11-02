package matcher

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodPeerMatcher matches a Peer in Pod to Pod traffic against an ANP, BANP, or v1 NetPol rule.
// It accounts for Namespace, Pod, and Port/Protocol.
type PodPeerMatcher struct {
	Namespace NamespaceMatcher
	Pod       PodMatcher
	Port      PortMatcher
}

func (ppm *PodPeerMatcher) PrimaryKey() string {
	return ppm.Namespace.PrimaryKey() + "---" + ppm.Pod.PrimaryKey()
}

func (ppm *PodPeerMatcher) Matches(subject, peer *TrafficPeer, portInt int, portName string, protocol v1.Protocol) bool {
	return !peer.IsExternal() &&
		ppm.Namespace.Matches(peer.Internal.Namespace, peer.Internal.NamespaceLabels, subject.Internal.NamespaceLabels) &&
		ppm.Pod.Matches(peer.Internal.PodLabels) &&
		ppm.Port.Matches(portInt, portName, protocol)
}

// PodMatcher possibilities:
// 1. PodSelector:
//   - empty/nil
//   - not empty
// 2. NamespaceSelector
//   - nil
//   - empty
//   - not empty
//
// Combined:
// 1. all pods in policy namespace
//   - empty/nil PodSelector
//   - nil NamespaceSelector
//
// 2. all pods in all namespaces
//   - empty/nil PodSelector
//   - empty NamespaceSelector
//
// 3. all pods in matching namespaces
//   - empty/nil PodSelector
//   - not empty NamespaceSelector
//
// 4. matching pods in policy namespace
//   - not empty PodSelector
//   - nil NamespaceSelector
//
// 5. matching pods in all namespaces
//   - not empty PodSelector
//   - empty NamespaceSelector
//
// 6. matching pods in matching namespaces
//   - not empty PodSelector
//   - not empty NamespaceSelector
//
// 7. everything
//   - don't have anything at all -- i.e. empty []NetworkPolicyPeer
//

type PodMatcher interface {
	Matches(podLabels map[string]string) bool
	PrimaryKey() string
}

type AllPodMatcher struct{}

func (p *AllPodMatcher) Matches(podLabels map[string]string) bool {
	return true
}

func (p *AllPodMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all pods",
	})
}

func (p *AllPodMatcher) PrimaryKey() string {
	return `{"type": "all-pods"}`
}

type LabelSelectorPodMatcher struct {
	Selector metav1.LabelSelector
}

func (p *LabelSelectorPodMatcher) Matches(podLabels map[string]string) bool {
	return kube.IsLabelsMatchLabelSelector(podLabels, p.Selector)
}

func (p *LabelSelectorPodMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "matching pods by label",
		"Selector": p.Selector,
	})
}

func (p *LabelSelectorPodMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "label-selector", "selector": "%s"}`, kube.SerializeLabelSelector(p.Selector))
}

// namespaces

// NamespaceMatcher detects if the peer namespace is a match
type NamespaceMatcher interface {
	Matches(namespace string, namespaceLabels, subjectNamespaceLabels map[string]string) bool
	PrimaryKey() string
}

type ExactNamespaceMatcher struct {
	Namespace string
}

func (p *ExactNamespaceMatcher) Matches(namespace string, _, _ map[string]string) bool {
	return p.Namespace == namespace
}

func (p *ExactNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":      "specific namespace",
		"Namespace": p.Namespace,
	})
}

func (p *ExactNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "exact-namespace", "namespace": "%s"}`, p.Namespace)
}

type LabelSelectorNamespaceMatcher struct {
	Selector metav1.LabelSelector
}

func (p *LabelSelectorNamespaceMatcher) Matches(_ string, namespaceLabels, _ map[string]string) bool {
	return kube.IsLabelsMatchLabelSelector(namespaceLabels, p.Selector)
}

func (p *LabelSelectorNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":     "matching namespace by label",
		"Selector": p.Selector,
	})
}

func (p *LabelSelectorNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "label-selector", "selector": "%s"}`, kube.SerializeLabelSelector(p.Selector))
}

type AllNamespaceMatcher struct{}

func (a *AllNamespaceMatcher) Matches(_ string, _, _ map[string]string) bool {
	return true
}

func (a *AllNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type": "all namespaces",
	})
}

func (a *AllNamespaceMatcher) PrimaryKey() string {
	return `{"type": "all-namespaces"}`
}

type SameLabelsNamespaceMatcher struct {
	labels []string
}

func (s *SameLabelsNamespaceMatcher) Matches(_ string, namespaceLabels, subjectNamespaceLabels map[string]string) bool {
	if len(s.labels) == 0 {
		return false
	}

	for _, k := range s.labels {
		v, ok := namespaceLabels[k]
		if !ok {
			return false
		}

		v2, ok := subjectNamespaceLabels[k]
		if !ok || v != v2 {
			return false
		}
	}

	return true
}

func (s *SameLabelsNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "same labels",
		"Labels": s.labels,
	})
}

func (s *SameLabelsNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "same-labels", "labels": "%s"}`, strings.Join(s.labels, ","))
}

type NotSameLabelsNamespaceMatcher struct {
	labels []string
}

func (s *NotSameLabelsNamespaceMatcher) Matches(_ string, namespaceLabels, subjectNamespaceLabels map[string]string) bool {
	if len(s.labels) == 0 {
		return false
	}

	different := false

	for _, k := range s.labels {
		v, ok := namespaceLabels[k]
		if !ok {
			return false
		}

		v2, ok := subjectNamespaceLabels[k]
		if !ok {
			return false
		}

		if v != v2 {
			different = true
		}
	}

	return different
}

func (s *NotSameLabelsNamespaceMatcher) MarshalJSON() (b []byte, e error) {
	return json.Marshal(map[string]interface{}{
		"Type":   "not same labels",
		"Labels": s.labels,
	})
}

func (s *NotSameLabelsNamespaceMatcher) PrimaryKey() string {
	return fmt.Sprintf(`{"type": "not-same-labels", "labels": "%s"}`, strings.Join(s.labels, ","))
}
