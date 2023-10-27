# NPEP-126: Add northbound traffic support in (B)ANP API

* Issue: [#126](https://github.com/kubernetes-sigs/network-policy-api/issues/126)
* Status: Provisional

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

(... details, can point to PR with changes)


## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

* https://github.com/danwinship/enhancements/blob/cluster-egress-firewall/keps/sig-network/20190917-cluster-egress-firewall.md#blocking-access-to-services-used-by-the-node
* https://github.com/kubernetes-sigs/network-policy-api/pull/86