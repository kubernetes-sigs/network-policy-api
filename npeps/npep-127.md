# NPEP-127: Add southbound traffic support in ClusterNetworkPolicy API

* Issue: [#127](https://github.com/kubernetes-sigs/network-policy-api/issues/127)
* Status: Provisional

## TLDR

This NPEP proposes adding support for cluster ingress (southbound) traffic
control in the `ClusterNetworkPolicy` API by adding `networks` and `nodes`
fields to the ingress peer, mirroring what already exists for egress.

## Goals

* Add `networks` (CIDR) peer to `ClusterNetworkPolicyIngressPeer`, enabling
  ingress traffic control from external destinations (outside the cluster)
* Add `nodes` peer to `ClusterNetworkPolicyIngressPeer`, enabling ingress
  traffic control from cluster nodes

## Non-Goals

* N/A

## Introduction

Currently, `ClusterNetworkPolicyIngressPeer` only supports `namespaces` and
`pods` selectors, which only match non-hostNetwork cluster pods. This
enhancement aims to add `networks` (CIDR) and `nodes` peers to ingress rules,
allowing administrators to match incoming traffic by source IP address or
originating node. A similar API already exists for
[egress](https://github.com/kubernetes-sigs/network-policy-api/issues/126)
direction (see [NPEP-126](npep-126-egress-traffic-control.md)).

Most of the time, cluster workloads are not directly accessible from
outside the cluster. However, the
[Kubernetes networking model](https://kubernetes.io/docs/concepts/services-networking/)
does not define this -- it is up to each implementation, and there are
many solutions and use cases where pod
IPs are directly routable from external networks. Additionally, workloads
can be exposed through k8s APIs like
[Service](https://kubernetes.io/docs/concepts/services-networking/service/)
(NodePort, LoadBalancer),
[Ingress API](https://kubernetes.io/docs/concepts/services-networking/ingress/),
[Gateway API](https://gateway-api.sigs.k8s.io/), or downstream
implementations like
[Openshift's Route API](https://docs.openshift.com/container-platform/4.13/rest_api/network_apis/route-route-openshift-io-v1.html).
Some of these APIs may provide functionality to filter incoming
connections, but Network Policy API is different because it is applied
directly to the cluster workloads, and should provide the required level
of security regardless of how or whether a cluster workload is exposed
to the outer world.

### Source IP preservation

A key consideration for `networks` ingress matching is whether the
original source IP is visible to the policy enforcement point.

In some deployments, pod IPs are directly routable from external
networks via BGP, and source IP is naturally preserved:

- Calico
  [BGP peering](https://docs.projectcalico.org/networking/bgp) with
  native routing (no overlay) makes pod IPs directly routable from
  external networks
- ovn-kubernetes
  [Route Advertisements](https://ovn-kubernetes.io/features/bgp-integration/route-advertisements/)
  exports pod network routes to BGP peers, making pod IPs directly
  reachable from external networks

In other deployments where traffic goes through a Service (NodePort,
LoadBalancer), source IP of incoming connections is not guaranteed to
be preserved. But many implementations have a way to request source
IP preservation:

- Calico
  [eBPF dataplane](https://docs.tigera.io/calico/latest/operations/ebpf/enabling-ebpf#value)
- ovn-kubernetes preserves source IP for
  [ExternalTrafficPolicy=local](https://github.com/ovn-org/ovn-kubernetes/blob/master/docs/design/service-traffic-policy.md)
- Cilium
  [Client Source IP Preservation](https://docs.cilium.io/en/stable/network/kubernetes/kubeproxy-free/#client-source-ip-preservation)

When the source IP is preserved, it can be used for matching in the
ClusterNetworkPolicy ingress rules. When it is not preserved (e.g.
after SNAT by a load balancer or kube-proxy), the `networks` peer
will match against the translated source IP, which may be a node IP
rather than the original client IP. See the
[Expected Behavior](#expected-behavior) section for how
implementations should handle this.

### User Stories

1. **As a** cluster administrator running VMs and Kubernetes on the
cluster network where pod IPs are directly routable (e.g. via BGP),
**I want** to restrict which external VM subnets can send ingress
traffic to my cluster workloads **so that** I can control what
traffic from outside the cluster reaches my pods.

2. **As a** cluster administrator exposing services via LoadBalancer
with `externalTrafficPolicy: Local`, **I want** to deny ingress
traffic from specific external CIDRs to my cluster workloads **so
that** I can block known-bad source IP ranges while the original
client IP is preserved.

3. **As a** cluster administrator implementing a zero trust policy,
**I want** to deny all ingress traffic from outside the cluster
unless explicitly allowed **so that** no external connection can
reach cluster workloads without an explicit Accept rule. Note:
without `networks` on ingress, there is currently no way to express
a blanket deny-all ingress rule since at least one peer is required
(see [#248](https://github.com/kubernetes-sigs/network-policy-api/issues/248)).

4. **As a** cluster administrator **I want** to deny ingress
connections from outside the cluster on well-known ports like ftp
(21), telnet (23), SNMP (161) to cluster workloads **so that** I
can block insecure protocols from external sources without affecting
internal pod-to-pod communication on those same ports.

5. **As a** cluster administrator **I want** to create a Baseline
tier ClusterNetworkPolicy that denies ingress from specific external
CIDRs **so that** namespaces without a NetworkPolicy `ipBlock`
ingress rule still have a default security posture against known-bad
external sources, while namespace admins can override this with their
own NetworkPolicy if needed.

6. **As a** cluster administrator **I want** to deny ingress
connections from specific cluster nodes to sensitive workloads **so
that** if a worker node is compromised, it cannot affect workloads
running on other nodes.

## API

The proposal is to add `networks` and `nodes` fields to
`ClusterNetworkPolicyIngressPeer`, mirroring the existing fields on
`ClusterNetworkPolicyEgressPeer`:

```go
type ClusterNetworkPolicyIngressPeer struct {
	// Namespaces defines a way to select all pods within a set of Namespaces.
	// Note that host-networked pods are not included in this type of peer.
	//
	// +optional
	Namespaces *metav1.LabelSelector `json:"namespaces,omitempty"`

	// Pods defines a way to select a set of pods in
	// a set of namespaces. Note that host-networked pods
	// are not included in this type of peer.
	//
	// +optional
	Pods *NamespacedPod `json:"pods,omitempty"`

	// Nodes defines a way to select a set of nodes in
	// the cluster (based on the node's labels). It selects
	// the nodeIPs as the peer type by matching on the IPs
	// present in the node.Status.Addresses field of the node.
	// This field follows standard label selector
	// semantics; if present but empty, it selects all Nodes.
	//
	// TBD: What about secondary node IPs or internal management
	// IPs that are not in node.Status.Addresses? It's pretty hard
   // to solve the case of ignoring externalIPs, egressIPs, KVIPs...and other corner cases
	//
	// <network-policy-api:experimental>
	//
	// +optional
	Nodes *metav1.LabelSelector `json:"nodes,omitempty"`

	// Networks defines a way to select peers via CIDR blocks.
	// This is intended for representing entities that live outside the cluster,
	// which can't be selected by pods, namespaces and nodes peers, but note
	// that cluster-internal traffic will be checked against the rule as
	// well. So if you Accept or Deny traffic from "0.0.0.0/0", that will allow
	// or deny all IPv4 pod-to-pod traffic as well. If you don't want that,
	// add a rule that Passes all pod traffic before the Networks rule.
	//
	// Each item in Networks should be provided in the CIDR format and should be
	// IPv4 or IPv6, for example "10.0.0.0/8" or "fd00::/8".
	//
	// Networks can have up to 25 CIDRs specified.
	//
	// <network-policy-api:experimental>
	//
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=25
	Networks []CIDR `json:"networks,omitempty"`
}
```

### Example: Allow ingress from a known external CIDR

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: allow-from-external-database
spec:
  tier: Admin
  priority: 100
  subject:
    namespaces:
      matchLabels:
        app: backend
  ingress:
  - name: allow-from-external-db
    action: Accept
    from:
    - networks:
      - "203.0.113.0/24"
    protocols:
    - tcp:
        destinationPort: 5432
```

### Example: Deny all external ingress (zero trust)

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: deny-external-ingress
spec:
  tier: Admin
  priority: 200
  subject:
    namespaces: {}
  ingress:
  - name: pass-cluster-internal
    action: Pass
    from:
    - namespaces: {}
  - name: deny-all-external
    action: Deny
    from:
    - networks:
      - "0.0.0.0/0"
      - "::/0"
```

## Expected Behavior

1. The `networks` ingress peer matches when the source IP of the
   packet is contained within one of the specified CIDR blocks. The
   source IP used for matching is the one seen at the policy
   enforcement point (i.e. the IP present in the packet where the network
   plugin enforces network policy). Implementations are not required
   to reconstruct the original client IP if it has been translated
   or masqueraded.

2. In deployments where pod IPs are directly routable (e.g. via BGP
   with no overlay), the source IP of external traffic is the original
   client IP. The `networks` peer will match against this IP as
   expected.

3. In deployments where traffic passes through a Service (NodePort,
   LoadBalancer), the source IP may be translated (SNAT'd) to a node
   IP by kube-proxy or the load balancer. In this case:
   - If `externalTrafficPolicy: Local` or equivalent source IP
     preservation is configured, the original client IP is preserved
     and can be matched by the `networks` peer.
   - If source IP is not preserved, the `networks` peer will match
     against the translated (e.g. node) IP, not the original client
     IP. This is expected behavior -- the implementation matches what
     it sees.

4. Implementations should document their source IP preservation
   behavior and any configuration required to preserve source IPs for
   ingress `networks` matching.

5. The `nodes` ingress peer matches traffic originating from the
   node's IP addresses as reported in `node.status.addresses`
   (including host-networked pods on that node). **TBD**: How should
   secondary node IPs (e.g. multiple `InternalIP` entries) be
   handled?

6. Both `networks` and `nodes` ingress peers also match
   cluster-internal traffic. A rule that denies traffic from
   `"0.0.0.0/0"` will deny all IPv4 ingress including pod-to-pod
   traffic. To target only external traffic, add a higher-precedence
   rule that Passes all pod/namespace traffic before the `networks`
   rule.

7. Kubelet health probes (liveness, readiness, startup) originate
   from the node IP. A broad `networks` or `nodes` Deny rule may
   inadvertently block kubelet probes, causing pods to be restarted
   or removed from Service endpoints.
   [KEP-4559: Redesigning Kubelet Probes](https://github.com/kubernetes/enhancements/pull/4558)
   aims to solve this problem at the Kubernetes level by making
   kubelet probes work correctly in the presence of network policies.
   Until that KEP is implemented, implementations will need to handle
   kubelet probe traffic themselves (e.g. by automatically exempting
   probe traffic or documenting that administrators must add explicit
   Accept rules for probe traffic). See also
   [#314](https://github.com/kubernetes-sigs/network-policy-api/issues/314)
   for ongoing discussion around CNP and kubelet probes.
   **TBD**: There has been discussion (at KubeCon) about providing a
   well-known "allow same-node" semantic that would let administrators
   easily permit all traffic from the node a pod is running on,
   which would naturally cover kubelet probes.

## Conformance Details

The following conformance features should be defined:

- `ClusterNetworkPolicyIngressNetworks`: Tests that `networks` peer works
  correctly in ingress rules for ClusterNetworkPolicy.
- `ClusterNetworkPolicyIngressNodes`: Tests that `nodes` peer works
  correctly in ingress rules for ClusterNetworkPolicy.

## Alternatives

- **Rely on external firewalls or ingress controllers**: This shifts security
  outside the Network Policy API, which may not cover all traffic paths
  (e.g. NodePort, hostPort, direct pod IP access).

## References

- [NPEP-126: Egress traffic control](npep-126-egress-traffic-control.md) - the
  equivalent egress feature that introduced `networks` and `nodes` peers
- [Issue #127](https://github.com/kubernetes-sigs/network-policy-api/issues/127) -
  tracking issue for southbound traffic support
- [PR #249](https://github.com/kubernetes-sigs/network-policy-api/pull/249) -
  original NPEP-127 proposal
