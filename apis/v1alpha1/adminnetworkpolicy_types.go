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

// All fields in this package are required unless Explicitly marked optional
// +kubebuilder:validation:Required
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AdminNetworkPolicy is the Schema for the adminnetworkpolicies API.
type AdminNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Specification of the desired behavior of AdminNetworkPolicy.
	Spec AdminNetworkPolicySpec `json:"spec"`

	// Status is the status to be reported by the implementation.
	// +optional
	Status AdminNetworkPolicyStatus `json:"status,omitempty"`
}

// AdminNetworkPolicyStatus defines the observed state of AdminNetworkPolicy.
type AdminNetworkPolicyStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// AdminNetworkPolicySpec defines the desired state of AdminNetworkPolicy.
type AdminNetworkPolicySpec struct {
	// Priority is an int32 value bound to 0 - 1000, the lowest priority,
	// "0" corresponds to the highest importance, while higher priorities have
	// lower importance.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000
	Priority int32 `json:"priority"`

	// Subject defines the pods to which this AdminNetworkPolicy applies.
	Subject AdminNetworkPolicySubject `json:"subject"`

	// List of Ingress rules to be applied to the selected pods BEFORE any
	// NetworkPolicy or BaselineAdminNetworkPolicy rules have been applied.
	// A total of 100 rules will be allowed in each ANP instance.
	// ANPs with no ingress rules do not affect ingress traffic.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Ingress []AdminNetworkPolicyIngressRule `json:"ingress,omitempty"`

	// List of Egress rules to be applied to the selected pods BEFORE any
	// NetworkPolicy or BaselineAdminNetworkPolicy rules have been applied.
	// A total of 100 rules will be allowed in each ANP instance.
	// ANPs with no egress rules do not affect egress traffic.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Egress []AdminNetworkPolicyEgressRule `json:"egress,omitempty"`
}

// AdminNetworkPolicyIngressRule describes an action to take on a particular
// set of traffic destined for pods selected by an AdminNetworkPolicy's
// Subject field. The traffic must match both ports and from.
type AdminNetworkPolicyIngressRule struct {
	// Name is an identifier for this rule, that may be no more than 100 characters
	// in length. This field should be used by the implementation to help
	// improve observability, readability and error-reporting for any applied
	// AdminNetworkPolicies.
	// +optional
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name,omitempty"`

	// Action specifies the affect this rule will have on matching traffic,
	// currently the following actions are supported:
	// Allow: allows the selected traffic
	// Deny: denies the selected traffic
	// Pass: instructs the selected traffic to skip any remaining ANP rules, and
	// then pass execution to any NetworkPolicies that select the pod.
	// If the pod is not selected by any NetworkPolicies then execution
	// is passed to any BaselineAdminNetworkPolicies that select the pod.
	// This field is mandatory.
	Action AdminNetworkPolicyRuleAction `json:"action"`

	// Ports allows for matching traffic based on port and protocols.
	// If Ports is not set then traffic is not filtered via port.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Ports []AdminNetworkPolicyPort `json:"ports,omitempty"`

	// List of sources whose traffic this AdminNetworkPolicyRule applies to.
	// If any adminNetworkPolicyPeer matches the source of incoming
	// traffic then the specified action is applied.
	// This field must be defined and contain at least one item.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	From []AdminNetworkPolicyPeer `json:"from"`
}

// AdminNetworkPolicyEgressRule describes an action to take on a particular
// set of traffic originating from pods selected by a AdminNetworkPolicy's
// Subject field. The traffic must match both ports and to.
type AdminNetworkPolicyEgressRule struct {
	// Name is an identifier for this rule, that may be no more than 100 characters
	// in length. This field should be used by the implementation to help
	// improve observability, readability and error-reporting for any applied
	// AdminNetworkPolicies.
	// +optional
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name,omitempty"`

	// Action specifies the affect this rule will have on matching traffic,
	// currently the following actions are supported:
	// Allow: allows the selected traffic (even if it would otherwise have been denied by NetworkPolicy)
	// Deny: denies the selected traffic (even if it would otherwise have been denied by NetworkPolicy)
	// Pass: instructs the selected traffic to skip any remaining ANP rules, and
	// then pass execution to any NetworkPolicies that select the pod.
	// If the pod is not selected by any NetworkPolicies then execution
	// is passed to any BaselineAdminNetworkPolicies that select the pod.
	// This field is mandatory.
	Action AdminNetworkPolicyRuleAction `json:"action"`

	// Ports allows for matching traffic based on port and protocols.
	// If Ports is not set then traffic is not filtered via port.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Ports []AdminNetworkPolicyPort `json:"ports,omitempty"`

	// List of destinations whose traffic this adminNetworkPolicyRule applies to.
	// If any adminNetworkPolicyPeer matches the destination of outgoing
	// traffic then the specified action is applied.
	// This field must be defined and contain at least one item.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	To []AdminNetworkPolicyPeer `json:"to"`
}

// AdminNetworkPolicyRuleAction string describes the AdminNetworkPolicy action type.
// +enum
type AdminNetworkPolicyRuleAction string

const (
	// AdminNetworkPolicyRuleActionPass enables admins to provide exceptions to
	// AdminNetworkPolicies by passing rule execution directly to any matching
	// K8s networkPolicies.
	AdminNetworkPolicyRuleActionPass AdminNetworkPolicyRuleAction = "Pass"
	// AdminNetworkPolicyRuleActionDeny enables admins to deny specific traffic.
	AdminNetworkPolicyRuleActionDeny AdminNetworkPolicyRuleAction = "Deny"
	// AdminNetworkPolicyRuleActionAllow enables admins to specifically allow certain traffic.
	AdminNetworkPolicyRuleActionAllow AdminNetworkPolicyRuleAction = "Allow"
)

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
