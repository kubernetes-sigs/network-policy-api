# API Object Overview

Prior to the Network Policy API the original NetworkPolicy V1 Resource was the only
way for k8s users to apply security rules to their kubernetes workloads. One of the
main drawbacks to this API was that it was designed exclusively for use by the
Application Developer, although in reality it is used by many different cluster
personas, sometimes creating a complex web of objects to be maintained. In
contrast, each resource in the Network Policy API is designed to be used by a
specific persona.

With the advent of the ClusterNetworkPolicy resource Cluster Admins will now have 
the ability to apply policy globally with only a few simple policy objects.

## Roles and personas

In this documentation we refer to three primary personas:

- Application Developer
- Namespace Administrator
- Cluster Administrator

## Resource Model

!!! note
    Network Policy API resources are in the `policy.networking.k8s.io` API group as
    Custom Resource Definitions (CRDs). Unqualified resource names below will
    implicitly be in this API group.

Currently, there is one main object in the Network Policy API resource model:

- **ClusterNetworkPolicy (CNP)**

The diagram below demonstrates how these new API objects interact with
each-other and existing NetworkPolicy Objects:

<img src="/images/CNP.drawio.svg" width="541" alt="Cluster Network Policy API model">

## General Notes

- Any well-defined ClusterNetworkPolicy rules should
be read as-is, i.e. there will not be any implicit isolation effects for the Pods
selected by the ClusterNetworkPolicy, as opposed to what NetworkPolicy rules imply.

- We now have multiple API versions, see the [ClusterNetworkPolicy blog post](blog/posts/ClusterNetworkPolicy.md) for more details

## The ClusterNetworkPolicy Resource

The ClusterNetworkPolicy (CNP) resource will help administrators set cluster-wide security
rules for the cluster, which are evaluated before or after the NetworkPolicies defined by the
namespace owners.

### Tiers

Tier is used as the top-level grouping for network policy prioritization.

Policy tiers are evaluated in the following order:
* `Admin` tier
* NetworkPolicy tier
* `Baseline` tier

ClusterNetworkPolicy can use 2 of these tiers: `Admin` and `Baseline`.

The `Admin` tier will help administrators set strict security rules for the cluster, 
i.e. a developer CANNOT override these rules by creating NetworkPolicies that apply 
to the same workloads as the ClusterNetworkPolicy.

The `Baseline` tier will allow administrators to set baseline security rules that 
describe default connectivity for cluster workloads, which CAN be overridden by 
developer NetworkPolicies if needed. 
The major use case for `Baseline` tier is the ability to flip the [default security stance of the
cluster](user-stories.md#story-5-cluster-wide-default-guardrails).

### Actions

Unlike the NetworkPolicy resource in which each rule represents an allowed
traffic, ClusterNetworkPolicy will enable administrators to set `Pass`,
`Deny` or `Accept` as the action of each rule. ClusterNetworkPolicy rules should
be read as-is, i.e. there will not be any implicit isolation effects for the Pods
selected by the ClusterNetworkPolicy, as opposed to implicit deny NetworkPolicy rules imply.

- **Accept**: Accepts the selected traffic, allowing it into
  the destination. No further ClusterNetworkPolicy or
  NetworkPolicy rules will be processed.

- **Deny**: Drops the selected traffic. No further
  ClusterNetworkPolicy or NetworkPolicy rules will be
  processed.

- **Pass**: Skips all further ClusterNetworkPolicy rules in the
  current tier for the selected traffic, and passes
  evaluation to the next tier.

ClusterNetworkPolicy `Deny` rules are useful for administrators to explicitly
block traffic with malicious in-cluster clients, or workloads that pose security risks.
Those traffic restrictions can only be lifted once the `Deny` rules are deleted,
modified by the admin, or overridden by a higher priority rule.

On the other hand, the `Accept` rules can be used to call out traffic in the cluster
that needs to be allowed for certain components to work as expected (egress to
CoreDNS for example). This traffic should not be blocked when developers apply
NetworkPolicy to their Namespaces which isolates the workloads.

ClusterNetworkPolicy `Pass` rules in the `Admin` tier allow an admin to delegate security posture for
certain traffic to the Namespace owners by overriding any lower precedence Allow
or Deny rules. For example, intra-namespace traffic management can be delegated to namespace
admins explicitly with the use of `Pass` rules. More specifically traffic selected by a `Pass` rule
will skip any lower precedence `Admin` tier rules and proceed to be evaluated by `NetworkPolicy` and
`Baseline` tier policies. When the `Pass` action is matched at the `Admin` tier, `NetworkPolicy` will 
apply next or if there is no `NetworkPolicy` match, `Baseline` policies will be evaluated.

### Priorities 

Integer priority values were added to the ClusterNetworkPolicy API to allow Cluster 
Admins to express direct and intentional ordering between various CNP Objects.  
The `Priority` field in the CNP spec is defined as a non-negative integer value 
where rules with lower priority values have higher precedence, and are checked 
before policies with higher priority values in the same tier.

### Rules 

Each CNP should define at least one `Ingress` or `Egress` relevant in-cluster traffic flow 
along with the associated Action that should occur. In each `gress` rule the user 
should AT THE MINIMUM define an `Action`, and at least one `ClusterNetworkPolicyPeer`.
Optionally the user may also define select `Ports` to filter traffic on and also 
a name for each rule to make management and reporting easier for Admins.

### Status 

For `v1alpha2` of this API the ANP status field is simply defined as a list of 
[`metav1.condition`](https://github.com/kubernetes/apimachinery/blob/v0.25.0/pkg/apis/meta/v1/types.go#L1464)s. Currently there are no rules as to what these conditions should display,
and it is up to each implementation to report what they see fit. For further 
API iterations the community may consider standardizing these conditions based on 
the usefulness they provide for various implementors.
