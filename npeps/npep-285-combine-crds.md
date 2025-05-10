# NPEP-285: NPEP template

* Issue: [#285](https://github.com/kubernetes-sigs/network-policy-api/issues/285)
* Status: Provisional

## TLDR

Combine ANP and BANP into a single CRD.

## Goals

- Outline existing differences between ANP and BANP.
- Merge ANP and BANP into a single CRD.

## Non-Goals

(What is explicitly out of scope for this proposal.)

## Introduction

ANP and BANP are 2 layers of policies around NetworkPolicy, evaluated as shown in the diagram below.

![image](../site-src/images/ANP-api-model.png)  
Currently, ANP and BANP are 2 separate CRDs with the following differences:
- BANP is a singleton, which is also the reason why it doesn't have `spec.priority` field.
- ANP supports `Pass` action, while BANP doesn't. `Pass` action could be considered a delegation to the NetworkPolicy, 
or it could be seen as return from the current policy layer .
- ANP supports `domainNames` matching for egress rules, while BANP doesn't.

## User-Stories/Use-Cases

Story 1: Multiple BANPs

As a cluster admin, I want to apply different BANPs to different sets namespaces, so that I can have different default trust levels.
For example, "system" namespaces, like kube-system, could have a less strict BANP than "user" namespaces.

Story 2: Single CRD

As a cluster admin, I prefer to have 1 CRD for both ANP and BANP, so that it is easier to manage, understand, and monitor.
As an API implementation developer, I prefer to have 1 CRD for both ANP and BANP, so that I can reduce the complexity of the code.

## API

### Combining ANP and BANP

To combine ANP and BANP, we need to have a way to specify policy layer. Currently, we have the following options:
- Add a new field `spec.tier` to distinguish ANP vs BANP. The values could be `Override` and `Baseline`.
- Change `spec.priority` field to have a pre-defined value for NetworkPolicy, use numbers below for ANP, and numbers above for BANP.
For example, if we set NetworkPolicy priority to 0, then negative priorities would mean ANP, and positive priorities would mean BANP. 

We prefer to go with the first option, because using negative numbers is more difficult to understand and explain, plus
the semantics of `Pass` action becomes too confusing and too hard to explain. 
"Pass skips over all remaining ANPs with priority greater than 0, then checks NPs, then comes back and processes the rest of the ANPs with priority less than 0."

#### Here we go again... naming

Trying to find the right name for the `spec.tier` field wasn't easy and brought up the discussion about the CRD name itself.

1. Persona-based naming

Current naming is persona/user-focused, AdminNetworkPolicy is a cluster-scoped policy and DeveloperNetworkPolicy was supposed to be
a replacement for the namespaced NetworkPolicy (instead of NetworkPolicy v2).
This may be confusing, because nothing prevents "developers" with the right RBAC from creating AdminNetworkPolicies.
And admins also could use DeveloperNetworkPolicies to enforce policies on some namespaces.
This naming also prevents (or makes more confusing) adding more personas in the future, like Platform/Infrastructure team that
could have its own `tier` in AdminNetworkPolicy not really being cluster admins.

We could do tier names based on the personas though, like Admin/BaselineAdmin/Platform.
NetworkPolicy v2 could be called ApplicationNetworkPolicy (as was originally proposed).

2. Changing name is a challenge

We have advertised ANP/BANP a lot during previous kubecons, and we also have some implementations with real customers 
that are using it. So changing the name of the CRD may bring more confusion about it being a totally new API, and potentiall
eternity of people mixing up old vs new name.  
On the other hand, this is our last chance to do what is right before beta.

#### Find a better name for `spec.priority`

While we are here, `priority` field name was brought to our attention multiple times as confusing. It should be mostly related to the fact
that "higher priority" means "lower number", which is counter-intuitive. Suggested alternatives are "order" and "sequence".

On the other hand "priority" is a well-known term in the networking world, and the live survey during [sig-network updates EU 2025](https://www.youtube.com/watch?v=lBOdQHNNgEU)
showed that a majority of people understands the `priority` field correctly. Some examples from the industry:
- [google cloud firewall](https://cloud.google.com/firewall/docs/firewalls#priority_order_for_firewall_rules)
- [AWS Network FIrewall](https://docs.aws.amazon.com/network-firewall/latest/developerguide/suricata-rule-evaluation-order.html)
- [Azure Firewall](https://learn.microsoft.com/en-us/azure/firewall/rule-processing)

### Resolving differences between ANP and BANP

1. BANP is a singleton, which is also the reason why it doesn't have `spec.priority` field.
There is no "deep" reason for this; it's just because, at the time, we didn't have user stories that would require more 
than one BANP, so we decided to initially only allow a single BANP, figuring that we could always relax that and allow 
multiple BANPs later if we found good use cases for it, whereas if we started off allowing multiple BANPs and then it 
turned out we had been right before and there really were no good use cases for it, it would be harder to retroactively
restrict it, because even if there weren't good use cases for it, people would have come up with bad use cases and wouldn't want us to change it.

Decision: Make BANP non-singleton and apply `spec.priority` field to it.

2. ANP supports `Pass` action, while BANP doesn't. `Pass` action could be considered a delegation to the NetworkPolicy,
  or it could be seen as return from the current policy layer.   
We have 2 options: allow using `Pass` for BANP, meaning that the rest od BANPs will be skipped or forbid this action for BANP.  
Unless we find strong reasons to forbid `Pass` for BANP, it makes sense to allow it to make the API more uniform.

3. ANP supports `domainNames` matching for egress rules, while BANP doesn't.  
The original reasoning from the 
[FQDN NPEP](https://github.com/kubernetes-sigs/network-policy-api/blob/feae59d056cf87131338a8449d31b85dc1d9790f/npeps/npep-133-fqdn-egress-selector.md?plain=1#L33-L38)
is
> Since Kubernetes NetworkPolicy does not have a FQDN selector, adding this
capability to BaselineAdminNetworkPolicy could result in writing baseline
rules that can't be replicated by an overriding NetworkPolicy. For example,
if BANP allows traffic to `example.io`, but the namespace admin installs a
Kubernetes Network Policy, the namespace admin has no way to replicate the
`example.io` selector using just Kubernetes Network Policies.

DNS Name selector is only [supported](https://github.com/kubernetes-sigs/network-policy-api/blob/feae59d056cf87131338a8449d31b85dc1d9790f/apis/v1alpha1/adminnetworkpolicy_types.go#L282)
with the `Allow` rule, which means the desired override by the NetworkPolicy could only be `Deny`. 
Due to the default-deny nature of NetworkPolicy, all required traffic in a namespace should be explicitly allowed, which means
the override of the BANP `Allow` DNS Name selector will happen automatically.

Similar to the `Pass` action, unless we find strong reasons to forbid `domainNames` for BANP, it makes sense to allow it to make the API more uniform.

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

The only alternative is to leave ANP and BANP as 2 CRDs, the main advantage on this approach is that implementations that already support
an alpha version of the API don't need to introduce major changes. (But we have alpha to allow big changes like this, right?)

## References

Was discussed in the [Network Policy API meeting](https://docs.google.com/document/d/1AtWQy2fNa4qXRag9cCp5_HsefD7bxKe3ea2RPn8jnSs) on 11th and 25th February 2025.

