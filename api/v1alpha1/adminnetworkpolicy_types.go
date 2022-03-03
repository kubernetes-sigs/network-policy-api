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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AdminNetworkPolicySpec defines the desired state of AdminNetworkPolicy
type AdminNetworkPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AdminNetworkPolicy. Edit adminnetworkpolicy_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AdminNetworkPolicyStatus defines the observed state of AdminNetworkPolicy
type AdminNetworkPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AdminNetworkPolicy is the Schema for the adminnetworkpolicies API
type AdminNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdminNetworkPolicySpec   `json:"spec,omitempty"`
	Status AdminNetworkPolicyStatus `json:"status,omitempty"`
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
