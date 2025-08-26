# NPEP-308: Extending NetworkPolicy for Multi-Cluster support via `networking.k8s.io/multicluster-name` label

* Issue: [#308](https://github.com/kubernetes-sigs/network-policy-api/issues/308)
* Status: Provisional

## TLDR

This proposal introduces a standardized method for defining multi-cluster
`NetworkPolicies` by extending the existing Kubernetes `NetworkPolicy` API. This
is achieved by giving a new meaning to the well-defined label,
`networking.k8s.io/multicluster-name`, in the `podSelector` and
`namespaceSelector`, enabling policies to target resources across different
clusters.

## Goals

* To establish a standard, agnostic convention for multi-cluster network
  policies within the existing `NetworkPolicy` resource.
* To enhance portability of network security policies across various
  multi-cluster Kubernetes environments.
* To avoid the proliferation of custom CRDs for multi-cluster policy definition,
  hence simplifying the user experience.

## Non-Goals

* This proposal does **not** define the architecture or implementation of the
  multi-cluster control or data planes.
* It does **not** specify how projects should implement the underlying network
  connectivity and policy enforcement between clusters.

## Introduction

The current Kubernetes `NetworkPolicy` API is limited to intra-cluster traffic.
As multi-cluster architectures become more common, there is a growing need for a
standardized way to manage traffic flow between services in different clusters.
Today, this is handled by project-specific solutions, such as Cilium's
`CiliumNetworkPolicy` or Calico's `GlobalNetworkPolicy`. This leads to a
fragmented ecosystem where security policies are not portable and users are
locked into a specific project for multi-cluster functionality.

This proposal aims to address this by introducing a simple, backward-compatible
extension to the `NetworkPolicy` API that will provide a unified way to define
cross-cluster communication rules.

## User-Stories/Use-Cases

Story 1: Secure Applications Communication Across Clusters

* **As a** DevOps engineer,
* **I want to** define a `NetworkPolicy` that allows a `frontend` service in
  `cluster-1` to communicate with a `backend` service in `cluster-2`,
* **so that** I can enforce fine-grained security policies for my application,
  regardless of which cluster the services are running in.

Story 2: Isolate Development and Production Environments

* **As a** platform administrator,
* **I want to** create a default-deny policy for a namespace in my development
  cluster that allows egress traffic only to specific services in my production
  cluster,
* **so that** I can prevent unauthorized access from the development environment
  to sensitive production resources.

## API

We propose the introduction of a new reserved label,
`networking.k8s.io/multicluster-name`, to be used within the `podSelector` and
`namespaceSelector` of `NetworkPolicy` resources. When this label is present,
the CNI is responsible for interpreting it as a selector for resources in a
remote cluster.

**Importantly, the addition of this label will have no impact on CNIs or other
projects that do not implement this functionality.** If a CNI does not recognize
the `networking.k8s.io/multicluster-name` label, it will simply be ignored, and
the `NetworkPolicy` will be enforced as if the label was not present. This
ensures full backward compatibility and means that this change will not break
any existing `NetworkPolicy` implementations.

Here is an example of a `NetworkPolicy` that allows ingress traffic from pods
with the label `app: frontend` in the `default` namespace of `cluster-2` to pods
with the label `app: backend` in the `default` namespace of the local cluster:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-cross-cluster-frontend-to-backend
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
      namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: default
          networking.k8s.io/multicluster-name: cluster-2
```

## Conformance Details

The feature name for conformance testing will be `MultiClusterNetworkPolicyLabelExtensions`.

## Alternatives

* Per Project-Specific CRDs: This is the current approach, which leads to a lack
  of portability and a fragmented user experience. Each project has its own CRD
  and syntax for defining multi-cluster policies.

* New MultiClusterNetworkPolicy CRD: This would create a new, dedicated CRD for
  multi-cluster policies. While this would provide a standard, it would also add
  another API to learn and manage, and it would not be as seamlessly integrated
  as extending the existing NetworkPolicy API.

The proposed solution of extending the existing NetworkPolicy API is preferred
because it is the most lightweight, backward-compatible, and user-friendly
approach.

## References

* [Cilium Cluster Mesh Policy](https://docs.cilium.io/en/latest/network/clustermesh/policy/)
* [Calico Cluster Mesh](https://docs.tigera.io/use-cases/cluster-mesh)
* [Calico Federated Endpoints](https://github.com/tigera-cs/calico-cloud-unified-control/blob/main/modules/federatedendpoints-1.md)
