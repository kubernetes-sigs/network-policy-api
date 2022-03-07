/*
Copyright 2022.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AdminNetworkPolicy is the Schema for the adminnetworkpolicies API
type AdminNetworkPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of AdminNetworkPolicy.
	// +optional
	Spec AdminNetworkPolicySpec `json:"spec,omitempty"`

	// ANPStatus is the status to be reported by the implementation, this is not
	// standardized in alpha and consumers should report what they see fit in
	// relation to their AdminNetworkPolicy implementation
	// +optional
	Status AdminNetworkPolicyStatus `json:"status,omitempty"`
}

// AdminNetworkPolicyStatus defines the observed state of AdminNetworkPolicy
type AdminNetworkPolicyStatus struct {
	Conditions []metav1.Condition `json:"status"`
}

// AdminNetworkPolicySubject defines what objects the policy selects.
// Exactly one of the `NamespaceSelector` or `NamespaceAndPodSelector` pointers
// should be set.
type AdminNetworkPolicySubject struct {
	NamespaceSelector       *metav1.LabelSelector `json:"namespaceselector"`
	NamespaceAndPodSelector *NamespacedPodSubject `json:"namespaceandpodselector"`
}

// NamespacedPodSubject allows the user to select a given set of pod(s) in
// selected namespace(s)
type NamespacedPodSubject struct {
	// This field follows standard label selector semantics; if present but empty,
	// it selects all Namespaces.
	NamespaceSelector *metav1.LabelSelector `json:"namespaceselector"`

	// Used to explicitly select pods within a namespace; if present but empty,
	// it selects all Pods.
	PodSelector *metav1.LabelSelector `json:"podselector"`
}

// AdminNetworkPolicySpec defines the desired state of AdminNetworkPolicy
type AdminNetworkPolicySpec struct {
	// Priority is an int32 value bound to 0 - 1000, the lowest positive priority,
	// "1" corresponds to the highest importance, while higher priorities have
	// lower importance. An ANP with a priority of "0" will be evaluated after all
	// positive priority AdminNetworkPolicies and standard NetworkPolicies.
	// This field is NOT optional.
	Priority *int32 `json:"priority"`

	// Subject defines the objects to which this AdminNetworkPolicy applies.
	// This field is NOT optional.
	Subject AdminNetworkPolicySubject `json:"subject"`

	// List of Ingress rules to be applied to the selected objects.
	// A total of 100 rules will be allowed per each network policy instance,
	// this rule count will be calculated as the total summation of the
	// Ingress and Egress rules in a single AdminNetworkPolicy Instance.
	// If this field is empty then this AdminNetworkPolicy has no effect on
	// ingress traffic.
	Ingress []AdminNetworkPolicyIngressRule `json:"ingress,omitempty"`

	// List of Egress rules to be applied to the selected objects.
	// A total of 100 rules will be allowed per each network policy instance,
	// this rule count will be calculated as the total summation of the
	// Ingress and Egress rules in a single AdminNetworkPolicy Instance.
	// If this field is empty then this AdminNetworkPolicy has no effect on
	// egress traffic.
	Egress []AdminNetworkPolicyEgressRule `json:"egress,omitempty"`
}

// AdminNetworkPolicyIngressRule describes an action to take on a particular
// set of traffic destined for pods selected by an AdminNetworkPolicy's
// Subject field. The traffic must match both ports and from.
type AdminNetworkPolicyIngressRule struct {
	// Name is an identifier for this rule, that should be no more than 100 characters
	// in length.
	// +optional
	Name string `json:"name,omitempty"`

	// Action specifies whether this rule must pass, allow or deny traffic.
	// Allow: allows the selected traffic
	// Deny: denies the selected traffic
	// Pass: allows the selected traffic to skip and remaining positive priority (non-zero)
	// ANP rules and be delegated by K8's Network Policy.
	// This field is mandatory.
	Action AdminNetPolRuleAction `json:"action"`

	// Ports allows for matching on traffic based on port and protocols.
	// This field is mandatory.
	Ports AdminNetworkPolicyPorts `json:"ports"`

	// List of sources from which traffic will be allowed/denied/passed to the entities
	// selected by this AdminNetworkPolicyRule. Items in this list are combined using a logical OR
	// operation. If this field is empty, this rule matches no sources.
	// If this field is present and contains at least one item, this rule
	// allows/denies/passes traffic from the defined AdminNetworkPolicyPeer(s)
	// If it is empty no traffic is matched by the AdminNetworkPolicyIngressRule.
	From []AdminNetworkPolicyPeer `json:"from"`
}

// AdminNetworkPolicyEgressRule describes an action to take on a particular
// set of traffic originating from pods selected by a AdminNetworkPolicy's
// Subject field. The traffic must match both ports and to.
type AdminNetworkPolicyEgressRule struct {
	// Name is an identifier for this rule, that should be no more than 100 characters
	// in length.
	// +optional
	Name string `json:"name,omitempty"`

	// Action specifies whether this rule must pass, allow or deny traffic.
	// Allow: allows the selected traffic
	// Deny: denies the selected traffic
	// Pass: allows the selected traffic to skip and remaining positive priority (non-zero)
	// ANP rules and be delegated by K8's Network Policy.
	// This field is mandatory.
	Action AdminNetPolRuleAction `json:"action"`

	// Ports allows for matching on traffic based on port and protocols.
	// This field is mandatory.
	Ports AdminNetworkPolicyPorts `json:"ports"`

	// List of destinations to which traffic will be allowed/denied/passed from the entities
	// selected by this AdminNetworkPolicyRule. Items in this list are combined using a logical OR
	// operation. If this field is empty, this rule matches no destinations.
	// If this field is present and contains at least one item, this rule
	// allows/denies/passes traffic to the defined AdminNetworkPolicyPeer(s)
	// If it is empty no traffic is matched by the AdminNetworkPolicyEgressRule.
	To []AdminNetworkPolicyPeer `json:"to"`
}

// AdminNetworkPolicyPorts handles selection of traffic based on port(s).
// Exactly one of the fields must be defined.
type AdminNetworkPolicyPorts struct {
	// AllPorts cannot be "false" when it is set
	// AllPorts allows the user to select all ports for all protocols, thus not
	// selecting traffic based on L4 principles.
	// If "true" then all ports are selected for all protocols.
	// +optional
	AllPorts *bool `json:"allports,omitempty"`

	// The list of ports to allow/deny/pass traffic on, each item in this list is
	// combined using a logical OR. When this field is present it should contain at
	// least one item, and this rule allows/denies/passes traffic only if the traffic
	// matches at least one port in the list.
	// +optional
	List []AdminNetworkPolicyPort `json:"list,omitempty"`
}

// AdminNetworkPolicyPort describes a port to select
type AdminNetworkPolicyPort struct {
	// The protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
	// field defaults to TCP.
	// +optional
	Protocol *v1.Protocol `json:"protocol,omitempty"`

	// The port on the given protocol. This can either be a numerical or named
	// port on a pod. If this field is not provided, this matches no port names and
	// numbers.
	// If present, only traffic on the specified protocol AND port will be matched.
	// +optional
	Port *intstr.IntOrString `json:"port,omitempty"`

	// If set, indicates that the range of ports from port to endPort, inclusive,
	// should be allowed by the policy. This field cannot be defined if the port field
	// is not defined or if the port field is defined as a named (string) port.
	// The endPort must be equal or greater than port.
	// +optional
	EndPort *int32 `json:"endport,omitempty"`
}

// AdminNetPolRuleAction string describes the AdminNetworkPolicy action type.
// +enum
type AdminNetPolRuleAction string

const (
	// RuleActionPass enables admins to provide exceptions to ClusterNetworkPolicies and delegate this rule to
	// K8s NetworkPolicies.
	AdminNetPolRuleActionPass AdminNetPolRuleAction = "Pass"
	// RuleActionDeny enables admins to deny specific traffic.
	AdminNetPolRuleActionDeny AdminNetPolRuleAction = "Deny"
	// RuleActionAllow enables admins to specifically allow certain traffic.
	AdminNetPolRuleActionAllow AdminNetPolRuleAction = "Allow"
)

// AdminNetworkPolicyPeer defines an in-cluster peer to allow traffic to/from.
// Exactly one of the selector pointers should be set for a given peer.
type AdminNetworkPolicyPeer struct {
	// +optional
	Namespaces *NamespaceSet `json:"namespaces,omitempty"`

	// +optional
	NamespacedPods *NamespaceAndPodSet `json:"namespacedpods,omitempty"`
}

// NamespaceSet defines a flexible way to select Namespaces in a cluster.
// Exactly one of the selectors should be set.  If a consumer observes none of
// its fields are set, they should assume an option they are not aware of has
// been specified and fail closed.
type NamespaceSet struct {
	// Self cannot be "false" when it is set.
	// If Self is "true" then all pods in the subject's namespace are selected.
	// +optional
	Self *bool `json:"self,omitempty"`

	// NotSelf cannot be "false" when it is set.
	// if NotSelf is "true" then all pods not in the subject's Namespace are selected.
	// +optional
	NotSelf *bool `json:"notself,omitempty"`

	// NamespaceSelector is a labelSelector used to select Namespaces, This field
	// follows standard label selector semantics; if present but empty, it selects
	// all Namespaces.
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// SameLabels is used to select a set of Namespaces that share the same values
	// for a set of labels.
	// To be selected a Namespace must have all of the labels defined in SameLabels,
	// and they must all have the same value as the subject of this policy.
	// If Samelabels is Empty then nothing is selected.
	// +optional
	SameLabels []string `json:"samelabels,omitempty"`

	// NotSameLabels is used to select a set of Namespaces that do not have a set
	// of label(s). To be selected a Namespace must have none of the labels defined
	// in NotSameLabels. If NotSameLabels is empty then nothing is selected.
	// +optional
	NotSameLabels []string `json:"notsamelabels,omitempty"`
}

// PodSet defines a flexible way to select pods in a cluster. Exactly one of the
// selectors should be set.  If a consumer observes none of its fields are set,
// they should assume an option they are not aware of has been specified and fail closed.
type PodSet struct {
	// PodSelector is a labelSelector used to select Pods, This field is NOT optional,
	// follows standard label selector semantics and if present but empty, it selects
	// all Pods.
	PodSelector *metav1.LabelSelector `json:"podselector"`
}

// NamespaceSetAndPod defines a flexible way to select Namespaces and pods in a
// cluster. The `Namespaces` and `Pods` fields are required and must not be empty.
type NamespaceAndPodSet struct {
	// Namespaces is used to select a set of Namespaces.  It must be defined and
	// non-empty.
	Namespaces NamespaceSet `json:"namespaces"`

	// Namespaces is used to select a set of Pods in the set of Namespaces. It must
	// must be defined and non-empty.
	Pods PodSet `json:"pods"`
}

//+kubebuilder:object:root=true

// AdminNetworkPolicyList contains a list of AdminNetworkPolicy
type AdminNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdminNetworkPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdminNetworkPolicy{}, &AdminNetworkPolicyList{})
}
