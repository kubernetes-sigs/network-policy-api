
### Summary

1. **As a** cluster administrator **I want** to easily block access from
cluster workloads to k8s-apiservers running on cluster nodes (host-net pods).
2. **As a** cluster administrator **I want** to easily allow access to
k8s-apiservers from cluster workloads when there are other deny rules in place
for these workloads.
3. **As a** cluster administrator **I want** to easily block access from
cluster workloads to specific ports on cluster nodes.

### Description

Currently the behaviour for policies defined around traffic from cluster
workloads (non-hostNetworked pods) towards hostNetworked pods or nodes in the
cluster is undefined. See https://github.com/kubernetes-sigs/network-policy-api/issues/73.

ANP currently supports only east-west traffic and this traffic flow cuts from
overlay to underlay which makes this part of the egress (south-north) use case.

Let's provide a defined behaviour in ANP to explicitly achieve the use cases.

### Acceptance Criteria

1. If I apply a policy that limits egress access from workloads identified
by label:A towards k8s-apiserver backends (master nodes at 6443 port), it should
be able to guarantee that those workloads cannot access the apiserver
endpoints:targetPort. Even if the master node IPs change, the policy should be
able to automatically pivot towards the new nodeIPs.

[
Also, here are a few points that need to be addressed:

1. Clarify that the existing `podSelector` in the ANP API are for regular
non-hostNetworked pods only. These don't apply to hostNetworked pods.
2. Add a new well defined optional field called `hostNetworkPodSelector` to both
`Subject` and `Peer` to be able to explicitly select host networked pods.
]

### Resources:

* https://github.com/danwinship/enhancements/blob/cluster-egress-firewall/keps/sig-network/20190917-cluster-egress-firewall.md#blocking-access-to-services-used-by-the-node
