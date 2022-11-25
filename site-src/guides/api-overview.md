## API overview

Prior to the AdminNetworkPolicy API there was no native tooling for Cluster Admins
to apply security rules in a cluster-wide manner, and in some cases Network Policies
were being incorrectly used to do so, creating a complex web of objects to be
maintained.

With the advent of the AdminNetworkPolicy API Cluster Admins will now have the 
ability to apply policy on in-cluster workloads with only a few simple policy 
objects that can be applied globally.

## Roles and personas

In this documentation we refer to three primary personas:

- Application Developer
- Namespace Administrator
- Cluster Administrator

## Resource Model 

There are two main objects in the AdminNetworkPolicy API resource model

- **AdminNetworkPolicy (ANP)**

- **BaselineAdminNetworkPolicy (BANP)**

## General Notes

- Any well defined AdminNetworkPolicy rules should
be read as-is, i.e. there will not be any implicit isolation effects for the Pods
selected by the AdminNetworkPolicy, as opposed to what NetworkPolicy rules imply.

- As of `v1alpha1` of this API we focus primarily on E/W cluster traffic and 
do not address N/S (Ingress/Egress) use cases. However this is an issue the community 
would like to keep thinking about during further iterations, and a tracking issue 
can be found/ commented on here ---> [issue #28](https://github.com/kubernetes-sigs/network-policy-api/issues/28)

## The AdminNetworkPolicy Resource 

The AdminNetworkPolicy (ANP) resource will help administrators set strict security 
rules for the cluster, i.e. a developer CANNOT override these rules by creating 
NetworkPolicies that apply to the same workloads as the AdminNetworkPolicy.

### AdminNetworkPolicy Actions 

Unlike the NetworkPolicy resource in which each rule represents an allowed
traffic, AdminNetworkPolicies will enable administrators to set `Pass`,
`Deny` or `Allow` as the action of each rule. AdminNetworkPolicy rules should
be read as-is, i.e. there will not be any implicit isolation effects for the Pods
selected by the AdminNetworkPolicy, as opposed to implicit deny NetworkPolicy rules imply.

- **Pass**: Traffic that matches a `Pass` rule will skip all further rules from all
  numbered ANPs and instead be enforced by the K8s NetworkPolicies.
  If there is no K8s NetworkPolicy of BaselineAdminNetworkPolicy rule match
  traffic will be governed by the implementation. For most implementations,
  this means "allow", but there may be implementations which have their own policies 
  outside of the standard Kubernetes APIs.
- **Deny**: Traffic that matches a `Deny` rule will be dropped.
- **Allow**: Traffic that matches an `Allow` rule will be allowed.

AdminNetworkPolicy `Deny` rules are useful for administrators to explicitly
block traffic with malicious in-cluster clients, or workloads that pose security risks.
Those traffic restrictions can only be lifted once the `Deny` rules are deleted,
modified by the admin, or overridden by a higher priority rule.

On the other hand, the `Allow` rules can be used to call out traffic in the cluster
that needs to be allowed for certain components to work as expected (egress to
CoreDNS for example). This traffic should not be blocked when developers apply
NetworkPolicy to their Namespaces which isolates the workloads.

AdminNetworkPolicy `Pass` rules allow an admin to delegate security posture for
certain traffic to the Namespace owners by overriding any lower precedence Allow
or Deny rules. For example, intra-tenant traffic management can be delegated to tenant
admins explicitly with the use of `Pass` rules. More specifically traffic selected 
by a `Pass` rule will skip any further ANP rule selection, be evaluated against
any well defined NetworkPolicies, and if not terminated ultimately be evaluated against any 
BaselineAdminNetworkPolicies. 

### AdminNetworkPolicy Priorities 

Integer priority values were added to the AdminNetworkPolicy API to allow Cluster 
Admins to express direct and intentional ordering between various ANP Objects.  
The `Priority` field in the ANP spec is defined as an integer value 
within the range 0 to 1000 where rules with lower priority values have higher 
precedence. Regardless of priority value all ANPs have higher precedence than 
any defined NetworkPolicy or BaselineAdminNetworkPolicy objects.

### AdminNetworkPolicy Rules 

Each ANP should define at least one `Ingress` or `Egress` relevant in-cluster traffic flow 
along with the associated Action that should occur. In each `gress` rule the user 
should AT THE MINIMUM define an `Action`, and at least one `AdminNetworkPolicyPeer`.
Optionally the user may also define select `Ports` to filter traffic on and also 
a name for each rule to make management and reporting easier for Admins.

### AdminNetworkPolicy Status 

For `v1alpha1` of this API the ANP status field is simply defined as a list of 
[`metav1.condition`](https://github.com/kubernetes/apimachinery/blob/v0.25.0/pkg/apis/meta/v1/types.go#L1464)s. Currently there are no rules as to what these conditions should display,
and it is up to each implementation to report what they see fit. For further 
API iterations the community may consider standardizing these conditions based on 
the usefulness they provide for various implementors.

## The BaselineAdminNetworkPolicy Resource 

The BaselineAdminNetworkPolicy (BANP) resource will allow administrators to 
set baseline security rules that describes default connectivity for cluster workloads, 
which CAN be overridden by developer NetworkPolicies if needed. The major use case 
BANPs solve is the ability to flip the [default security stance of the 
cluster](../index.md#story-5-cluster-wide-default-guardrails).

### BaselineAdminNetworkPolicy Rule Actions 

BaselineAdminNetworkPolicies allow administrators to define two distinct actions
for each well defined rule, `Allow` and `Deny`. 

- **Deny**: Traffic that matches a `Deny` rule will be dropped.
- **Allow**: Traffic that matches an `Allow` rule will be allowed.

### BaselineAdminNetworkPolicy Rules 

BANP Rules are defined and behave the same (Except for the `Pass` action) as [ANP 
rules](#adminnetworkpolicy-rules).

### BaselineAdminNetworkPolicy Status

The BANP `status` field follows the same constructs as used by the
[AdminNetworkPolicy.Status](#adminnetworkpolicy-status) field.
