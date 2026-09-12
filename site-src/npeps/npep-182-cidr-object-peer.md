# NPEP-182: Add new CIDR object peer for northbound traffic

* Issue: [#182](https://github.com/kubernetes-sigs/network-policy-api/issues/182)
* Status: Provisional

## Co-Authors
@joestringer and @networkop for raising relevant user stories

## TLDR

This NPEP proposes adding support for a new CIDRGroup object peer type for
cluster egress (northbound) traffic control that can be referred in the
`AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` API objects using selectors.
[NPEP-126](https://network-policy-api.sigs.k8s.io/npeps/npep-126-egress-traffic-control/#implementing-egress-traffic-control-towards-cidrs)
already adds support for inline CIDR peer type directly on the
`AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` API objects. This NPEP proposes
adding more extensibility by introducing a new CIDRGroup object in addition to the
inline CIDR peers so that users can choose either of these methods based on their needs.

## Goals

* Provide users with a way to group their CIDRs in a meaningful
manner which can then be referred to from ANP and BANP objects.

## Non-Goals

## Introduction

The current approach of defining inline CIDR peers works well
if the number of CIDR blocks involved in defining policies are
less in number and mostly static in nature. However in environments
where we could have a more dynamic setup, the management of inline CIDR
peers gets more tricker an cumbersome. In such cases having a way to
group CIDR blocks together to represent an entity or a group of
entities which the policy can refer to as a network peer can be useful.
This also ensures reference of same CIDR group peer from ANP and BANP
stays consistent and any changes to the list of CIDR blocks only involves
editing the object itself and not the rules in the policy.

### User Stories for CIDRGrouping

* As a cluster admin I want to be able to create admin network policies that
match a dynamic set of external IPs (e.g. set of VMs or set of directly reachable
Pods in another cluster). I may not be able to use FQDN rules for that due to
TTL being too long or simply lack of DNS service discovery in an external system.
As a cluster admin, I would create CIDR group resource and a BGP controller that
would manage it. The mapping between BGP communities and CIDR group resource names
is a BGP controller configuration (e.g. annotation on the CIDR group resource).
The speed of IP churn is bounded by the BGP advertisement interval and can be
further reduced by the BGP controller implementation.

* As a cluster administrator I want to to ensure that pods can reach
commonly-used databases under my control but outside Kubernetes. Many but
not all applications in my environment rely on these databases. I want to
delegate writing network policy for this traffic to namespace owners.

Example: As a cluster administrator I define a CIDR group that defines
a set of RDS instances that is used across multiple apps. The owners of
namespaceA and namespaceB can then define policies that allow traffic to
this group of RDS instances, and they reference the instances by CIDR group.
As a cluster administrator I can migrate the database infrastructure and
update the CIDR group independently of the namespace owners. The applications
in namespaceC do not use this infrastructure, so the cluster administrator
and the owners of namespaceC do not need to think about network policy
for apps in namespaceC.

NOTE: Second use case is not possible today using NetworkPolicy resource
since we only have `ipBlocks` as a peer however this is definitely a useful
case to keep in mind for having a CIDR Group.

## API

This NPEP Proposes to add a new `CIDRGroup` object:

```
// CIDRGroup defines a group of CIDR blocks that can be referred to from
// AdminNetworkPolicy and BaselineAdminNetworkPolicy resources.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type CIDRGroup struct {
	// cidrs is the list of network cidrs that can be used to define destinations.
	// A total of 25 CIDRs will be allowed in each CIDRGroup instance.
	// ANP & BANP APIs may use the .spec.egress.to.networks.cidrGroups selector
	// to select a set of cidrGroups.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=25
	cidrs []CIDR `json:"cidrs,omitempty"
}
```

In order to ensure it is coexisting with inline CIDR peers without confusion,
we propose to change the type of `networks` peer from `string` to a struct of type
`NetworkPeer`:

```
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type NetworkPeer struct {
	// cidrs represents a list of CIDR blocks
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=25
	CIDRs []CIDR `json:"cidrs,omitempty"
    	// cidrGroups defines a way to select cidrGroup objects
	// that consist of network CIDRs as a peer.
	// This field follows standard label selector semantics; if present
	// but empty, it selects all cidrGroups defined in the cluster.
	//
	// +optional
	CIDRGroups *metav1.LabelSelector `json:"cidrGroups,omitempty"
}
```

and this is referenced from an ANP or BANP Egress Peer in the following
manner:

```
type AdminNetworkPolicyEgressPeer struct {
    <snipped>
    // Networks defines a way to select peers via CIDR blocks. This is
    // intended for representing entities that live outside the cluster,
    // which can't be selected by pods and namespaces peers, but note
    // that cluster-internal traffic will be checked against the rule as
    // well, so if you Allow or Deny traffic to `"0.0.0.0/0"`, that will allow
    // or deny all IPv4 pod-to-pod traffic as well. If you don't want that,
    // add a rule that Passes all pod traffic before the Networks rule.
    //
    // Support: Core
    //
    // +optional
    // +kubebuilder:validation:MinItems=1
    // +kubebuilder:validation:MaxItems=100
    Networks []NetworkPeer `json:"networks,omitempty"
}
```

Define a `CIDRGroup` object, example:

```
apiVersion: policy.networking.k8s.io/v1alpha1
kind: CIDRGroup
metadata:
  name: cluster-wide-cidr-cloud-1
  labels:
    env: cloud-1
  annotations:
    "bgp.cidrmanager.k8s.io/is-managed": "true"
    "bgp.cidrmanager.k8s.io/32bit-community": "2147483647"
spec:
  cidrs:
  - 192.0.2.0/24
  - 203.0.113.0/24
  - 198.51.100.0/24
status:
  conditions:
  - lastTransitionTime: "2022-12-29T14:53:50Z"
    status: "True"
    type: Reconciled
```

Then refer to this object from an ANP:

```
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: networks-peer-example
spec:
  priority: 30
  subject:
    namespaces: {}
  egress:
    - action: Allow
      to:
      - networks:
          cidrGroups:
            matchLabels:
              env: cloud-1
    - action: Deny
      to:
      - networks:
          cidrs:
            - 0.0.0.0/0
```

## Alternatives

N/A

## References

See https://github.com/kubernetes-sigs/network-policy-api/pull/144#discussion_r1408175206 for details
