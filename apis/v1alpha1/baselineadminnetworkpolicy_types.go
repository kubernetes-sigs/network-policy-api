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

// BaselineAdminNetworkPolicy is a cluster level resource that is part of the
// adminNetworkPolicy API.
type BaselineAdminNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Specification of the desired behavior of BaselineAdminNetworkPolicy.
	Spec BaselineAdminNetworkPolicySpec `json:"spec"`

	// Status is the status to be reported by the implementation.
	// +optional
	Status BaselineAdminNetworkPolicyStatus `json:"status,omitempty"`
}

// BaselineAdminNetworkPolicyStatus defines the observed state of
// BaselineAdminNetworkPolicy.
type BaselineAdminNetworkPolicyStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// BaselineAdminNetworkPolicySpec defines the desired state of
// BaselineAdminNetworkPolicy.
type BaselineAdminNetworkPolicySpec struct {
	// Subject defines the pods to which this BaselineAdminNetworkPolicy applies.
	Subject AdminNetworkPolicySubject `json:"subject"`

	// List of Ingress rules to be applied to the selected pods if they are not
	// matched by any AdminNetworkPolicy or NetworkPolicy rules.
	// A total of 100 Ingress rules will be allowed in each BANP instance.
	// BANPs with no ingress rules do not affect ingress traffic.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Ingress []BaselineAdminNetworkPolicyIngressRule `json:"ingress,omitempty"`

	// List of Egress rules to be applied to the selected pods if they are not
	// matched by any AdminNetworkPolicy or NetworkPolicy rules.
	// A total of 100 Egress rules will be allowed in each BANP instance. BANPs
	// with no egress rules do not affect egress traffic.
	// +optional
	// +kubebuilder:validation:MaxItems=100
	Egress []BaselineAdminNetworkPolicyEgressRule `json:"egress,omitempty"`
}

// BaselineAdminNetworkPolicyIngressRule describes an action to take on a particular
// set of traffic destined for pods selected by a BaselineAdminNetworkPolicy's
// Subject field. The traffic must match both ports and from.
type BaselineAdminNetworkPolicyIngressRule struct {
	// Name is an identifier for this rule, that may be no more than 100 characters
	// in length. This field should be used by the implementation to help
	// improve observability, readability and error-reporting for any applied
	// BaselineAdminNetworkPolicies.
	// +optional
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name,omitempty"`

	// Action specifies the affect this rule will have on matching traffic,
	// currently the following actions are supported:
	// Allow: allows the selected traffic
	// Deny: denies the selected traffic
	// This field is mandatory.
	Action BaselineAdminNetworkPolicyRuleAction `json:"action"`

	// Ports allows for matching traffic based on port and protocols.
	// If Ports is not set then traffic is not filtered via port.
	// +optional
	Ports []AdminNetworkPolicyPort `json:"ports,omitempty"`

	// List of sources whose traffic this AdminNetworkPolicyRule applies to.
	// If any adminNetworkPolicyPeer matches the source of incoming
	// traffic then the specified action is applied.
	// This field must be defined and contain at least one item.
	// +kubebuilder:validation:MinItems=1
	From []AdminNetworkPolicyPeer `json:"from"`
}

// AdminNetworkPolicyEgressRule describes an action to take on a particular
// set of traffic originating from pods selected by a BaselineAdminNetworkPolicy's
// Subject field. The traffic must match both ports and to.
type BaselineAdminNetworkPolicyEgressRule struct {
	// Name is an identifier for this rule, that may be no more than 100 characters
	// in length. This field should be used by the implementation to help
	// improve observability, readability and error-reporting for any applied
	// BaselineAdminNetworkPolicies.
	// +optional
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name,omitempty"`

	// Action specifies the affect this rule will have on matching traffic,
	// currently the following actions are supported:
	// Allow: allows the selected traffic
	// Deny: denies the selected traffic
	// This field is mandatory.
	Action BaselineAdminNetworkPolicyRuleAction `json:"action"`

	// Ports allows for matching traffic based on port and protocols.
	// If Ports is not set then traffic is not filtered via port.
	// +optional
	Ports []AdminNetworkPolicyPort `json:"ports,omitempty"`

	// List of destinations whose traffic this adminNetworkPolicyRule applies to.
	// If any adminNetworkPolicyPeer matches the destination of outgoing
	// traffic then the specified action is applied.
	// This field must be defined and contain at least one item.
	// +kubebuilder:validation:MinItems=1
	To []AdminNetworkPolicyPeer `json:"to"`
}

// BaselineAdminNetworkPolicyRuleAction string describes the BaselineAdminNetworkPolicy
// action type.
// +enum
type BaselineAdminNetworkPolicyRuleAction string

const (

	// BaselineAdminNetworkPolicyRuleActionDeny enables admins to deny specific traffic.
	BaselineAdminNetworkPolicyRuleActionDeny BaselineAdminNetworkPolicyRuleAction = "Deny"
	// BaselineAdminNetworkPolicyRuleActionAllow enables admins to specifically allow certain traffic.
	BaselineAdminNetworkPolicyRuleActionAllow BaselineAdminNetworkPolicyRuleAction = "Allow"
)

//+kubebuilder:object:root=true

// BaselineAdminNetworkPolicyList contains a list of BaselineAdminNetworkPolicy
type BaselineAdminNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BaselineAdminNetworkPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BaselineAdminNetworkPolicy{}, &BaselineAdminNetworkPolicyList{})
}
