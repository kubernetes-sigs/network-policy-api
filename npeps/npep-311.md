# NPEP-311: Best Practices for Multi-Cluster NetworkPolicy in a Flat Network 

* Issue:
  [#311](https://github.com/kubernetes-sigs/network-policy-api/issues/311)
* Status: Informational

## TLDR

This NPEP documents a recommended set of practices for applying standard
NetworkPolicy resources in a multi-cluster, flat network environment. It
proposes a conventional labeling scheme, aligned with SIG-Multicluster, to
enable consistent and predictable cross-cluster policy enforcement without
changes to the API.

## Goals

* To establish and document a common, intuitive operational model for
  multi-cluster network policy.

* To provide clear, reusable patterns for administrators to secure applications
  that span multiple clusters.

* To align these practices with the cluster identification conventions
  established by SIG-Multicluster in
  [KEP-2149](https://github.com/kubernetes/enhancements/tree/master/keps/sig-multicluster/2149-clusterid).

## Non-Goals

* This proposal does not introduce any changes to the Kubernetes Network Policy
  API specification or any extension based on labels and/or annotations.

* This proposal does not mandate any specific CNI implementation or
  multi-cluster architecture beyond the prerequisite of a flat network.

## Introduction

In a multi-cluster architecture with a flat network, every pod IP is unique and
directly routable from any other pod, regardless of its origin cluster. This
topology allows for a simplified security model where NetworkPolicy selectors
can be evaluated against a global inventory of all pods and namespaces within a
defined ClusterSet.

The core principles of this model are:

* Policies are Local: A NetworkPolicy resource is applied only to the cluster
  where it is created. Policies are not replicated, which contains the impact of
  changes and allows for per-cluster rollout strategies.

* Selectors are Global: The policy engine within each cluster evaluates policy
  rules against all known pods and namespaces in the entire ClusterSet.

* Administrator-Defined Identity: The ability to differentiate between clusters
  is important for cross-cluster communication. This model places the
  responsibility on the cluster administrator to implement a consistent labeling
  strategy that can be used for identity and policy selection.

This document formalizes these principles as a set of best practices for the
community.

## User-Stories/Use-Cases

### Story 1: Securing a "Stretched" Application

As a platform operator running an application across multiple clusters,

I want to write a single policy to allow traffic from all frontend pods to my
database pods,

so that I don't have to manage separate, IP-based policies for frontend
instances that are spread across different clusters.

### Story 2: Enforcing Cross-Cluster Security Boundaries

As a security administrator,

I want to allow a database application in a production cluster to receive
traffic only from a specific billing application in a separate PCI-compliant
cluster,

so that I can enforce strict, auditable cross-cluster communication paths.

### Story 3: Maintaining Local Policy Scope

As an application developer,

I want to apply a NetworkPolicy to my application and be confident that it only
affects traffic within my local cluster,

so that I don't accidentally expose my service to other clusters or break
connectivity by applying a policy that is too broad.

## API

No changes are proposed to the NetworkPolicy v1 API neither extensions based on
labels and/or annotations. This document describes a set of practices that
leverage the existing API.

The central recommendation is to adopt a consistent labeling scheme for cluster
identification. To effectively manage policies in a multi-cluster environment,
it is highly recommended to align with the conventions outlined in [KEP-2149:
ClusterId](https://github.com/kubernetes/enhancements/tree/master/keps/sig-multicluster/2149-clusterid).

### Recommendation

Each namespace within a cluster should be labeled by the cluster administrator with a key that identifies its
parent cluster. The recommended label is:

`cluster.clusterset.k8s.io/cluster-name: <cluster-id>`

The <cluster-id> value should correspond to the unique name of the cluster
within the ClusterSet (e.g., cluster-a, us-west-2). This practice enables policy
authors to create selectors that precisely target peers from specific clusters.

## Conformance Details

Not applicable, as this is an informational proposal that does not introduce new
API features.

## Alternatives

The primary alternative is the absence of a documented best practice. This leads
to fragmented, implementation-specific approaches to multi-cluster policy,
reducing portability and creating a confusing experience for users. By
documenting a common pattern, we provide a consistent model that both users and
CNI implementers can reference.

## References

* KEP-2149: ClusterId for ClusterSet identification:
  https://github.com/kubernetes/enhancements/tree/master/keps/sig-multicluster/2149-clusterid

* Calico Multi-Cluster Flat model: Calico is one CNI that supports a flat
  network model where policies can be applied across clusters, as described in
  https://kubernetes.slack.com/archives/C01CWSHQWQJ/p1756283784627849