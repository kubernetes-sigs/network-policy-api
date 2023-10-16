/*
Copyright 2023 The Kubernetes Authors.
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

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ens,scope=Cluster
// +kubebuilder:printcolumn:name="Networks",type=string,JSONPath=".spec.networks"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ExternalNetworkSet is a cluster level resource that is used to define
// a set of networks outside the cluster which can be referred to from
// the AdminNetworkPolicy && BaselineAdminNetworkPolicy APIs as an external peer
type ExternalNetworkSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Specification of the desired behavior of ExternalNetworkSet.
	Spec ExternalNetworkSetSpec `json:"spec"`
}

// ExternalNetworkSetSpec defines the desired state of ExternalNetworkSet.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type ExternalNetworkSetSpec struct {
	// Networks is the list of NetworkCIDR (both v4 & v6) that can be used to define
	// external destinations.
	// A total of 100 CIDRs will be allowed in each NetworkSet instance.
	// ANP & BANP APIs may use the .spec.in(e)gress.from(to).externalNetworks selector
	// to select a set of external networks
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	Networks []string `json:"networks,omitempty" validate:"omitempty,dive,cidr"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ExternalNetworkSetList contains a list of ExternalNetworkSet
type ExternalNetworkSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExternalNetworkSet `json:"items"`
}
