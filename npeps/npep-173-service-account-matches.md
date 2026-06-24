# NPEP-173: NPEP template

* Issue: [#173](https://github.com/kubernetes-sigs/network-policy-api/issues/173)
* Status: Provisional

## TLDR

Ability to create policies that control network traffic based on workload identities a.ka.
[Kubernetes service accounts](https://kubernetes.io/docs/concepts/security/service-accounts/).

## Goals

* Use service accounts (identities) as a way of describing (admin) network policies for pods

## Non-Goals

* There might be other identity management constructs like SPIFFE IDs which
are outside the scope of this enhancement. We can only provide a way to select
constructs known to core Kubernetes.

## Introduction

Every pod in Kubernetes will have an associated service account. If user does
not explicitly provide a service account, a default service account is created
in the namespace. A given pod will always have 1 service account associated
with itself. So instead of using pod labels as a way of selecting pods sometimes
there might be use cases to use service accounts as a way of selecting the pods.
This NPEP tries to capture the design details for service accounts as selectors
for policies.

## User-Stories/Use-Cases

1. **As a** cluster admin **I want** to select pods using their service accounts
   instead of labels **so that** I can avoid having to setup webhooks and
   validation admission policies to prevent users from changing labels on their
   namespaces and pods that will make my policy start/stop matching the traffic
   which is undesired
2. **As a** cluster admin **I want** to select pods using their service accounts
   instead of labels **so that** I can avoid the scale impact caused due to
   mutation of pod/namespace labels that will cause a churn which makes my CNI
   implementation react to that every time user changes a label.
3. **As a** cluster admin my workloads have immutable identities and **I want**
   to apply policies per workloads using their service accounts instead of labels
   **since** I want to have eventual consistency of that policy in the cluster.

## Unresolved Discussions

* How to provide a standard way to configure/describe the service mesh behavior
  of intercepting traffic and deciding whether to allow it based on information
  in the TLS packets?
  * NetworkPolicies apply on L3 and L4 while Meshes operate at L7 mostly. So when
    trying to express "denyAll connections except the allowed serviceAccount to
    serviceAccount connections"; how do we split the enforcement in this scenario
    between the CNI plugin that implementations the network policy at L3/L4 and
    service mesh implementation that implements the policy at L7?
    * One way is to probably split the implementation responsibility:
      1. CNI plugin can take of implementing denyAll connections and
         allowed serviceAccount to serviceAccount connections upto L3/L4
      2. service mesh implementation can implement the allowed serviceAccount to
         serviceAccount connections on L7
    * NOTE: There are some service mesh implementations which have a
      CNI component in them that collapses the logic of enforcement - so maybe
      that might become a requirement that the implementation be able to handle
      end to end enforcement of the full policy
  * We might also want to express additional constraints at L7 (such as only
    allowing requests from a given identity on specific paths or with specific
    HTTP/gRPC methods) - Ideally we would want these extra fields to be coexisting
    with the L3/L4 Authorization policy
* Should this really be part of AdminNetworkPolicy and NetworkPolicy APIs or should
  this be a new CRD?
  * Making it part of existing APIs: Existing network policy APIs are pretty heavy
    with other types of selectors like namespaces, pods, nodes, networks, fqdns and
    expecting mesh implementations to implement the full CRD just to get the identity
    based selection might not be practical
  * Making it part of new API: We fall back to compatibility problem of the different
    layers and coexistence of this new API on a same cluster as existing network policy APIs

## API

(... details, can point to PR with changes)

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
