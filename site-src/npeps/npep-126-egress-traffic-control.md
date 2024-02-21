# NPEP-126: Add northbound traffic support in (B)ANP API

* Issue: [#126](https://github.com/kubernetes-sigs/network-policy-api/issues/126)
* Status: Implementable

## TLDR

This NPEP proposes adding support for cluster egress (northbound) traffic control
in the `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` API objects.

## Goals

* Implement egress traffic control towards external destinations (outside the cluster)
* Implement egress traffic control towards cluster nodes
  - Currently the behaviour for policies defined around traffic from cluster
    workloads (non-hostNetworked pods) towards nodes in the
    cluster is undefined. See https://github.com/kubernetes-sigs/network-policy-api/issues/73.
    - ANP currently supports only east-west traffic and this traffic flow cuts from
    overlay to underlay which makes this part of the egress (northbound) use case.
    - Let's provide a defined behaviour in ANP to explicitly achieve the use case.
    - NOTE: Traffic towards nodes here includes traffic towards host-networked pods on that node
      because a "node" resource encompasses all objects that share the host-networking resources
* Implement egress traffic control towards k8s-apiservers
  - An apiserver endpoint in this context is special in the sense that it can be any entity
    including but not limited to a host-networked pod within the cluster OR external VMs OR
    infrastructure nodes running outside the cluster. This is why its a separate category goal.

## Non-Goals

* Implementing southbound (ingress) traffic use cases is outside the scope of this NPEP
* Implementing egress traffic control towards arbitrary hostNetworked pods is outside the scope of this NPEP
  - Currently the behaviour for policies defined around traffic from cluster
  workloads (non-hostNetworked pods) towards hostNetworked pods in the
  cluster is undefined. See https://github.com/kubernetes-sigs/network-policy-api/issues/73.
  - ANP currently supports only east-west traffic and this traffic flow cuts from
  overlay to underlay which makes this part of the egress (northbound) use case.
  - NOTE: Currently there are no user stories for `CNI pod to arbitrarily chosen hostNetworked pods`.
    Let's provide a defined behaviour in ANP to explicitly achieve the use case in the future if we have
    user stories for this outside of the k8s-apiserver usecase which is already covered in the goals.
    If that happens, this can be moved to goals.

## Introduction

### User Stories for egress traffic control towards external destinations

1. **As a** cluster administrator **I want** to restrict traffic from
specific cluster workloads to all or specific destinations outside the
cluster **so that** I can enforce security for northbound traffic.
Example: Pods in namespaceA and namespaceB should not be able to talk
to the internet but they should be able to access company's intranet.

2. **As a** cluster administrator **I want** to to ensure that pods can
reach my cluster-external DNS server even if namespace admins create
NetworkPolicies that block cluster-external egress.
Example: As an owner of namespaceA I define policies that deny all
northbound egress traffic for that namespace. However the cluster-admin
can decide all namespaces in the cluster must be able to talk to the
EXTERNAL_DNS_SERVER_IP on port 53.

### User Stories for egress traffic control towards cluster nodes

1. **As a** cluster administrator **I want** to easily block access from
cluster workloads to specific ports on cluster nodes without having to block
access to those ports on external hosts, without having to manually list
the IP address of every node, and without having to change the policy when
new nodes are added to the cluster.

### User Stories for egress traffic control towards k8s-apiservers

1. **As a** cluster administrator **I want** to easily allow access to
k8s-apiservers from cluster workloads when there are other deny rules in place
for these workloads.

2. **As a** cluster administrator **I want** to easily block access from
selected cluster workloads to k8s-apiservers for securing the server.

## API

Proof of Concept for the API design details can be found here:

* https://github.com/kubernetes-sigs/network-policy-api/pull/143
* https://github.com/kubernetes-sigs/network-policy-api/pull/185

### Implementing egress traffic control towards cluster nodes

This NPEP proposes to add a new type of `AdminNetworkPolicyEgressPeer` called `Nodes`
to be able to explicitly select nodes (based on the node's labels) in the cluster.
This ensures that if the list of IPs on a node OR list of nodes change, the users
don't need to manually intervene to include those new IPs. The label selectors will
take care of this automatically. Note that the nodeIPs that this type of peer matches
on are the IPs present in `Node.Status.Addresses` field of the node.

```
// AdminNetworkPolicyEgressPeer defines a peer to allow traffic to.
// Exactly one of the selector pointers must be set for a given peer. If a
// consumer observes none of its fields are set, they must assume an unknown
// option has been specified and fail closed.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
type AdminNetworkPolicyEgressPeer struct {
    <snipped>
	// Nodes defines a way to select a set of nodes in
	// in the cluster. This field follows standard label selector
	// semantics; if present but empty, it selects all Nodes.
	//
	// Support: Core
	//
	// +optional
	Nodes *metav1.LabelSelector `json:"nodes,omitempty"`
}
```

Note that `AdminNetworkPolicyPeer` will be changed to
`AdminNetworkPolicyEgressPeer` and `AdminNetworkPolicyIngressPeer` since ingress and
egress peers have started to diverge at this point and it is easy to
maintain it with two sets of peer definitions.
This ensures nodes can be referred to only as "egress peers".

Example: Admin wants to deny egress traffic from tenants who don't have
`restricted`, `confidential` or `internal` level security clearance
to control-plane nodes at 443 and 6443 ports in the cluster

```
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: node-as-egress-peer
spec:
  priority: 55
  subject:
    namespaces:
      matchExpressions:
        - {key: security, operator: notIn, values: [restricted, confidential, internal]}
  egress:
  - name: "deny-all-egress-to-kapi-server"
    action: "Deny"
    to:
    - nodes:
        matchLabels:
          node-role.kubernetes.io/control-plane:
    ports:
      - portNumber:
          protocol: TCP
          port: 443
      - portNumber:
          protocol: TCP
          port: 6443
```

### Implementing egress traffic control towards CIDRs

This NPEP proposes to add a new type of `AdminNetworkPolicyEgressPeer` called `Networks`
to be able to select destination CIDRs. This is provided to be able to select entities
outside the cluster that cannot be selected using the other peer types.
This peer type will not be supported in `AdminNetworkPolicyIngressPeer`.

```
// AdminNetworkPolicyEgressPeer defines a peer to allow traffic to.
// Exactly one of the selector pointers must be set for a given peer. If a
// consumer observes none of its fields are set, they must assume an unknown
// option has been specified and fail closed.
// +kubebuilder:validation:MaxProperties=1
// +kubebuilder:validation:MinProperties=1
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
	Networks []string `json:"networks,omitempty" validate:"omitempty,dive,cidr"`
}
```

Note: It is recommended to use `networks` to select a set of CIDR range destinations
that represent entities outside the cluster. If a user puts a podCIDR, nodeCIDR,
serviceCIDR or other intra-cluster networks, it will work but it is better to use
namespaces, pods, nodes peers to express such entities. Not all implementations can
correctly define the boundary between "internal" and "external" destinations with respect
to a Kubernetes cluster which is why this field is generic enough to select any CIDR
destination.

Example: Let's define ANP and BANP that refer to some CIDR networks:
```
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: network-as-egress-peer
spec:
  priority: 70
  subject:
    namespaces: {}
  egress:
  - name: "deny-egress-to-external-dns-servers"
    action: "Deny"
    to:
    - networks:
      - 194.0.2.0/24
      - 205.0.113.15/32
      - 199.51.100.10/32
    ports:
      - portNumber:
          protocol: UDP
          port: 53
  - name: "allow-all-egress-to-intranet"
    action: "Allow"
    to:
    - networks:
      - 192.0.2.0/24
      - 203.0.113.0/24
      - 198.51.100.0/24
  - name: "allow-all-intra-cluster-traffic"
    action: "Allow"
    to:
    - networks:
      - POD_CIDR
      - NODE_CIDR
      - SERVICE_CIDR
  - name: "pass-all-egress-to-internet"
    action: "Pass"
    to:
    - networks:
      - 0.0.0.0/0
---
apiVersion: policy.networking.k8s.io/v1alpha1
kind: BaselineAdminNetworkPolicy
metadata:
  name: default
spec:
  subject:
    namespaces: {}
  egress:
  - name: "deny-all-egress-to-internet"
    action: "Deny"
    to:
    - networks:
      - 0.0.0.0/0
```
This allows admins to specify rules that define:

* all pods cannot talk to company's intranet DNS servers.
* all pods can talk to rest of the company's intranet.
* all pods can talk to other pods, nodes, services.
* all pods cannot talk to internet (using last ANP Pass rule + BANP guardrail rule)

## Alternatives

* Instead of adding CIDR peer directly into the main object, we can
define a new object called `NetworkSet` and use selectors or
name of that object to be referred to from AdminNetworkPolicy and
BaselineAdminNetworkPolicy objects. This is particularly useful
if CIDR ranges are prone to changes versus the current model is
is better if the set of CIDRs are mostly a constant and are only referred
to from one or two egress rules. It increases readability. However the
drawback is if the CIDRs do change, then one has to ensure to update all
the relevant ANPs and BANP accordingly. In order to see whether we need
a new object to be able to define CIDRs in addition to the in-line peer,
we have another NPEP where that is being discussed
https://github.com/kubernetes-sigs/network-policy-api/pull/183. The scope
of this NPEP is limited to inline CIDR peers.

## References

* https://github.com/danwinship/enhancements/blob/cluster-egress-firewall/keps/sig-network/20190917-cluster-egress-firewall.md#blocking-access-to-services-used-by-the-node
* https://github.com/kubernetes-sigs/network-policy-api/pull/86