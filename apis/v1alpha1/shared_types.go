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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdminNetworkPolicySubject defines what resources the policy applies to.
// Exactly one field must be set.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AdminNetworkPolicySubject struct {
	// Namespaces is used to select pods via namespace selectors.
	// +optional
	Namespaces *metav1.LabelSelector `json:"namespaces,omitempty"`
	// Pods is used to select pods via namespace AND pod selectors.
	// +optional
	Pods *NamespacedPodSubject `json:"pods,omitempty"`
}

// NamespacedPodSubject allows the user to select a given set of pod(s) in
// selected namespace(s).
type NamespacedPodSubject struct {
	// NamespaceSelector follows standard label selector semantics; if empty,
	// it selects all Namespaces.
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector"`

	// PodSelector is used to explicitly select pods within a namespace; if empty,
	// it selects all Pods.
	PodSelector metav1.LabelSelector `json:"podSelector"`
}

// AdminNetworkPolicyPort describes how to select network ports on pod(s).
// Exactly one field must be set.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AdminNetworkPolicyPort struct {
	// Port selects a port on a pod(s) based on number.
	//
	// Support: Core
	//
	// +optional
	PortNumber *Port `json:"portNumber,omitempty"`

	// NamedPort selects a port on a pod(s) based on name.
	//
	// Support: Extended
	//
	// <network-policy-api:experimental>
	// +optional
	NamedPort *string `json:"namedPort,omitempty"`

	// PortRange selects a port range on a pod(s) based on provided start and end
	// values.
	//
	// Support: Core
	//
	// +optional
	PortRange *PortRange `json:"portRange,omitempty"`
}

type Port struct {
	// Protocol is the network protocol (TCP, UDP, or SCTP) which traffic must
	// match. If not specified, this field defaults to TCP.
	//
	// Support: Core
	//
	Protocol v1.Protocol `json:"protocol"`

	// Number defines a network port value.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	//
	// Support: Core
	//
	Port int32 `json:"port"`
}

// PortRange defines an inclusive range of ports from the the assigned Start value
// to End value.
type PortRange struct {
	// Protocol is the network protocol (TCP, UDP, or SCTP) which traffic must
	// match. If not specified, this field defaults to TCP.
	//
	// Support: Core
	//
	Protocol v1.Protocol `json:"protocol,omitempty"`

	// Start defines a network port that is the start of a port range, the Start
	// value must be less than End.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	//
	// Support: Core
	//
	Start int32 `json:"start"`

	// End defines a network port that is the end of a port range, the End value
	// must be greater than Start.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	//
	// Support: Core
	//
	End int32 `json:"end"`
}

// AdminNetworkPolicyIngressPeer defines an in-cluster peer to allow traffic from.
//
// Note that presence of a Service object with this policy subject as its backend
// has no impact on the behavior of the policy applied to the peer
// trying to talk to the Service. It will work in the same way as if the
// Service didn't exist since policy is applied after ServiceVIP (clusterIP,
// externalIP, loadBalancerIngressIP) is rewritten to the backendIPs.
//
// Exactly one of the selector pointers must be set for a given peer. If a
// consumer observes none of its fields are set, they must assume an unknown
// option has been specified and fail closed.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AdminNetworkPolicyIngressPeer struct {
	// Namespaces defines a way to select all pods within a set of Namespaces.
	// Note that host-networked pods are not included in this type of peer.
	//
	// Support: Core
	//
	// +optional
	Namespaces *NamespacedPeer `json:"namespaces,omitempty"`
	// Pods defines a way to select a set of pods in
	// a set of namespaces. Note that host-networked pods
	// are not included in this type of peer.
	//
	// Support: Core
	//
	// +optional
	Pods *NamespacedPodPeer `json:"pods,omitempty"`
}

// AdminNetworkPolicyEgressPeer defines a peer to allow traffic to.
//
// Note that presence of a Service object with this peer as its backend
// has no impact on the behavior of the policy applied to the subject
// trying to talk to the Service. It will work in the same way as if the
// Service didn't exist since policy is applied after ServiceVIP (clusterIP,
// externalIP, loadBalancerIngressIP) is rewritten to the backendIPs.
//
// Exactly one of the selector pointers must be set for a given peer. If a
// consumer observes none of its fields are set, they must assume an unknown
// option has been specified and fail closed.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AdminNetworkPolicyEgressPeer struct {
	// Namespaces defines a way to select all pods within a set of Namespaces.
	// Note that host-networked pods are not included in this type of peer.
	//
	// Support: Core
	//
	// +optional
	Namespaces *NamespacedPeer `json:"namespaces,omitempty"`
	// Pods defines a way to select a set of pods in
	// a set of namespaces. Note that host-networked pods
	// are not included in this type of peer.
	//
	// Support: Core
	//
	// +optional
	Pods *NamespacedPodPeer `json:"pods,omitempty"`
	// Nodes defines a way to select a set of nodes in
	// the cluster. This field follows standard label selector
	// semantics; if present but empty, it selects all Nodes.
	//
	// Support: Extended
	//
	// <network-policy-api:experimental>
	// +optional
	Nodes *metav1.LabelSelector `json:"nodes,omitempty"`
	// Networks defines a way to select peers via CIDR blocks.
	// This is intended for representing entities that live outside the cluster,
	// which can't be selected by pods, namespaces and nodes peers, but note
	// that cluster-internal traffic will be checked against the rule as
	// well. So if you Allow or Deny traffic to `"0.0.0.0/0"`, that will allow
	// or deny all IPv4 pod-to-pod traffic as well. If you don't want that,
	// add a rule that Passes all pod traffic before the Networks rule.
	//
	// Note that because policies are applied after Service VIPs (clusterIPs, externalIPs,
	// load balancer IPs) are rewritten to endpoint IPs, a Networks selector cannot match
	// such a VIP. For example, a Networks selector that denies traffic to the entire
	// service CIDR will not actually block any service traffic.
	//
	// Each item in Networks should be provided in the CIDR format and should be
	// IPv4 or IPv6, for example "10.0.0.0/8" or "fd00::/8".
	//
	// Networks can have upto 25 CIDRs specified.
	//
	// Support: Extended
	//
	// <network-policy-api:experimental>
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=25
	Networks []CIDR `json:"networks,omitempty"`
}

// NamespacedPeer defines a flexible way to select Namespaces in a cluster.
// Exactly one of the selectors must be set.  If a consumer observes none of
// its fields are set, they must assume an unknown option has been specified
// and fail closed.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type NamespacedPeer struct {
	// NamespaceSelector is a labelSelector used to select Namespaces, This field
	// follows standard label selector semantics; if present but empty, it selects
	// all Namespaces.
	//
	// Support: Core
	//
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// SameLabels is used to select a set of Namespaces that share the same values
	// for a set of labels.
	// To be selected a Namespace must have all of the labels defined in SameLabels,
	// AND they must all have the same value as the subject of this policy.
	// If Samelabels is Empty then nothing is selected.
	//
	// Support: Extended
	//
	// <network-policy-api:experimental>
	// +optional
	// +kubebuilder:validation:MaxItems=100
	SameLabels []string `json:"sameLabels,omitempty"`

	// NotSameLabels is used to select a set of Namespaces that do not have certain
	// values for a set of label(s).
	// To be selected a Namespace must have all of the labels defined in NotSameLabels,
	// AND at least one of them must have different values than the subject of this policy.
	// If NotSameLabels is empty then nothing is selected.
	//
	// Support: Extended
	//
	// <network-policy-api:experimental>
	// +optional
	// +kubebuilder:validation:MaxItems=100
	NotSameLabels []string `json:"notSameLabels,omitempty"`
}

// NamespacedPodPeer defines a flexible way to select Namespaces and pods in a
// cluster. The `Namespaces` and `PodSelector` fields are required.
type NamespacedPodPeer struct {
	// Namespaces is used to select a set of Namespaces.
	//
	// Support: Core
	//
	Namespaces NamespacedPeer `json:"namespaces"`

	// PodSelector is a labelSelector used to select Pods, This field is NOT optional,
	// follows standard label selector semantics and if present but empty, it selects
	// all Pods.
	//
	// Support: Core
	//
	PodSelector metav1.LabelSelector `json:"podSelector"`
}

// CIDR is an IP address range in CIDR notation (for example, "10.0.0.0/8" or "fd00::/8").
// The regex for the IPv4 and IPv6 CIDR range was taken from
// https://blog.markhatton.co.uk/2011/03/15/regular-expressions-for-ip-addresses-cidr-ranges-and-hostnames/
// The resulting regex is an OR of both regexes. IPv4 address embedded in IPv6 addresses are not supported.
// TODO: Change the CIDR's validation regex to use CEL isCIDR() in Kube 1.31 when it is available.
// +kubebuilder:validation:Pattern=`(^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(3[0-2]|[1-2][0-9]|[0-9]))$)|(^s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:)))(%.+)?s*(\/(12[0-8]|1[0-1][0-9]|[1-9][0-9]|[0-9]))$)`
// +kubebuilder:validation:XValidation:rule="self.contains(':') != self.contains('.')",message="CIDR must be either an IPv4 or IPv6 address. IPv4 address embedded in IPv6 addresses are not supported"
// +kubebuilder:validation:MaxLength=43
type CIDR string
