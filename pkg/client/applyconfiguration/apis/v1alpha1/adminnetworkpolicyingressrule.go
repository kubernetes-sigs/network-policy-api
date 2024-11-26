/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

// AdminNetworkPolicyIngressRuleApplyConfiguration represents a declarative configuration of the AdminNetworkPolicyIngressRule type for use
// with apply.
type AdminNetworkPolicyIngressRuleApplyConfiguration struct {
	Name   *string                                           `json:"name,omitempty"`
	Action *v1alpha1.AdminNetworkPolicyRuleAction            `json:"action,omitempty"`
	From   []AdminNetworkPolicyIngressPeerApplyConfiguration `json:"from,omitempty"`
	Ports  *[]AdminNetworkPolicyPortApplyConfiguration       `json:"ports,omitempty"`
}

// AdminNetworkPolicyIngressRuleApplyConfiguration constructs a declarative configuration of the AdminNetworkPolicyIngressRule type for use with
// apply.
func AdminNetworkPolicyIngressRule() *AdminNetworkPolicyIngressRuleApplyConfiguration {
	return &AdminNetworkPolicyIngressRuleApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *AdminNetworkPolicyIngressRuleApplyConfiguration) WithName(value string) *AdminNetworkPolicyIngressRuleApplyConfiguration {
	b.Name = &value
	return b
}

// WithAction sets the Action field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Action field is set to the value of the last call.
func (b *AdminNetworkPolicyIngressRuleApplyConfiguration) WithAction(value v1alpha1.AdminNetworkPolicyRuleAction) *AdminNetworkPolicyIngressRuleApplyConfiguration {
	b.Action = &value
	return b
}

// WithFrom adds the given value to the From field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the From field.
func (b *AdminNetworkPolicyIngressRuleApplyConfiguration) WithFrom(values ...*AdminNetworkPolicyIngressPeerApplyConfiguration) *AdminNetworkPolicyIngressRuleApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithFrom")
		}
		b.From = append(b.From, *values[i])
	}
	return b
}

func (b *AdminNetworkPolicyIngressRuleApplyConfiguration) ensureAdminNetworkPolicyPortApplyConfigurationExists() {
	if b.Ports == nil {
		b.Ports = &[]AdminNetworkPolicyPortApplyConfiguration{}
	}
}

// WithPorts adds the given value to the Ports field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Ports field.
func (b *AdminNetworkPolicyIngressRuleApplyConfiguration) WithPorts(values ...*AdminNetworkPolicyPortApplyConfiguration) *AdminNetworkPolicyIngressRuleApplyConfiguration {
	b.ensureAdminNetworkPolicyPortApplyConfigurationExists()
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithPorts")
		}
		*b.Ports = append(*b.Ports, *values[i])
	}
	return b
}