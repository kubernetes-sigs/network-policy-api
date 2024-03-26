# NPEP-122: Tenancy API

* Issue: [#122](https://github.com/kubernetes-sigs/network-policy-api/issues/122)
* Status: Implementable

## TLDR

Tenancy definition and the API (SameLabels/NotSameLabels) is confusing and ambiguous as of now. 
We want to rethink the tenancy use cases and the API to keep it simple and easy to understand, but
flexible enough to cover all defined use cases.

## Goals

- Clarify tenancy use cases.
- Provide a complete definition of a tenant in respect to ANP.
- Avoid unneeded tenancy configs that only exist because of ANP use cases.

## Non-Goals

- Define new use cases for tenancy (as opposed to clarify the ones we considered initially, but didn't explain well enough)
- Define user stories for multiple tenancy policies in the same cluster.

## Introduction

The KEP doesn’t say a whole lot about tenants…

From the Goals:

    As a cluster administrator, I want to have the option to enforce in-cluster network level access controls that 
    facilitate network multi-tenancy and strict network level isolation between multiple teams and tenants sharing 
    a cluster via use of namespaces or groupings of namespaces per tenant.
    
    Example: I would like to define two tenants in my cluster, one composed of the pods in foo-ns-1 and foo-ns-2 
    and the other with pods in bar-ns-1, where inter-tenant traffic is denied.

From the User Stories:

    Story 4: Create and Isolate multiple tenants in a cluster
    
    As a cluster admin, I want to build tenants in my cluster that are isolated from each other by default. 
    Tenancy may be modeled as 1:1, where 1 tenant is mapped to a single Namespace, or 1:n, where a single tenant 
    may own more than 1 Namespace.

Elsewhere:

    AdminNetworkPolicy Pass rules allows an admin to delegate security posture for certain traffic to the Namespace 
    owners by overriding any lower precedence Allow or Deny rules. For example, intra-tenant traffic management can be 
    delegated to tenant admins explicitly with the use of Pass rules.
    
So really, the only solidly-agreed-upon use case is that you should be able to create enforced isolation 
between particular sets of namespaces.

### Clarifying user stories

#### Story 4.1: Create and Isolate multiple tenants in a cluster by default, overridable isolation

Here is the existing tenancy related user story from our website:

    As a cluster admin, I want to build tenants in my cluster that are isolated from each other by default. 
    Tenancy may be modeled as 1:1, where 1 tenant is mapped to a single Namespace, or 1:n, where a single tenant 
    may own more than 1 Namespace.

The wording used here has led to much confusion. Specifically, the "by default" part actually means this policy should 
be at the BANP priority, and solves the "Engineering org vs Marketing org" use case where you want to keep orgs 
from accidentally interfering with each other. By default, cross-tenant traffic should be dropped. 
However, namespace owners can override this behavior by applying their own policies as needed. 

In reality, the CR attached to this story defines strict tenancy, which should be a separate use case (see next section).

To make the use case more obvious we can add some details to it like the following

    As a cluster admin, I want to build tenants for different departments (e.g. Engineering vs Marketing) 
    in my cluster that are isolated from each other by default.
    By default, cross-tenant traffic is dropped. However, namespace owners can override this behavior by applying 
    their own policies as needed.
    Tenancy may be modeled as 1:1, where 1 tenant is mapped to a single Namespace, or 1:n, where a single tenant 
    may own more than 1 Namespace.

#### Story 4.2: Create and Isolate multiple tenants in a cluster, strict isolation

Strict tenancy is the "Coke vs Pepsi" sort of thing where you want each tenant to feel like it has its own cluster, 
and be totally independent of the other tenants. We can write it down like this

    As a cluster admin, I want to build tenants for different organizations (e.g. Coke vs Pepsi) 
    in my cluster that are isolated from each other, where this isolation
    can't be overridden by namespace owners. This policy should make every tenant completely independent and isolated 
    from other tenants. Tenancy may be modeled as 1:1, where 1 tenant is mapped to a single Namespace, or 1:n, where a single tenant 
    may own more than 1 Namespace.

#### Story 4.3: Allow internal connections for tenants

    BANP
    As a cluster admin, I want to setup an overridable deny-all policy to protect namespaces by default.
    At the same time I want to build tenants in my cluster and allow connections inside one tenant by default.
    
    ANP
    As a cluster admin, I want to setup a deny-all policy to only allow connections that are explicitly specified.
    Besides allowing required cluster services (like kube-api, dns, etc.) with ANP, I want to build tenants 
    and allow connections inside one tenant.

Both user stories have zero-trust policy in mind, where every allowed connection should be explicitly specified,
and everything else is denied. It may be set up as "by default"/overridable/BANP or strict/ANP.
Allow connection inside one tenant in this context means skip deny rules for same tenant, and delegate same-tenant 
policies to the namespaces NetworkPolicy. We assume that there is no reason for cluster admin to forcefully allow connections
inside one tenant instead of delegating to NetworkPolicy.

#### Story 4.4: Tenants interaction with (B)ANP

    As a cluster admin, I want to be able to setup policies with higher and lower priority than tenancy policy.
    I want to deny inter-tenant connections, but I want to allow ingress traffic from the monitoring namespace
    to all namespaces. Therefore, I need "allow from monitoring" rule to have higher priority than "deny from
    other tenants".
    I want to setup "deny all" BANP to protect cluster workloads, but I want to allow internal connections
    within tenant. Therefore, I need "allow from same tenant" rule to have higher priority than "deny all" BANP.

#### What I couldn't figure out user stories for

- Ports *[]AdminNetworkPolicyPort

### Existing API

AdminNetworkPolicy has the “SameLabels” and “NotSameLabels” fields to support the use cases involving tenancy. For example:

**Use case**

Traffic should be disallowed by default between namespaces owned by different users (defined by labels i.e `user=foo` or `user=bar`).

**Policy**
```
kind: BaselineAdminNetworkPolicy
apiVersion: policy.networking.k8s.io/v1alpha1
metadata:
  name: user-based-tenancy
spec:
  subject:
    namespaces:
      matchExpressions:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces:
            notSameLabels:
              - user
```

**Meaning**

In namespaces that have a “user” label, by default, deny ingress from namespaces that have a different “user” label.

There are several major problems with this implementation of tenancy as it pertains to the user stories.

First, There is no explicit definition of "tenancy" anywhere. The administrator has an idea of 
"tenants are defined by the user label", but that's only true because this particular ANP happens to include that
particular rule, and there's no way to find the ANP(s) that defines tenancy if you don't already know what they are.

Second, the SameLabels/NotSameLabels selectors behave really differently from other peers, causing multiple underlying 
rules to be created, and the syntax doesn't make that obvious.

Third, the syntax is very general purpose / powerful. ANP has subjects and peers, which are different, 
and currently Tenancy is defined on the peers side. Tenancy by itself has the same subject and peer, 
at least for the existing use cases, and having separate selectors for subject and peer allows for more
configurations than needed. 

Fourth, the ANP subject allows using pod selectors, while tenancy use cases only need namespace selectors.

## API

### Tenant definition

For the purposes of this NPEP we define a Tenant as a set of namespaces.
`tenancyLabels` is a set of label keys, based on which all Namespaces affected by Tenancy are be split into Tenants.
A Tenant is identified by values of `tenancyLabels`, which are shared by all namespaces in a given Tenant.

There are 2 ways to select which namespaces should be affected by Tenancy rules:

1. Use a distinct new `namespaceSelector` field to define which namespaces should be affected by Tenancy, 
while the `tenancyLabels` field can be used to define how the selected namespaces are split into Tenants.

    **Cons**: selected namespace may not have some of the `tenancyLabels`, which will likely result in introducing "None" value
    for `tenancyLabels`. It brings more complexity and is not required for any of the listed use cases.

```yaml
spec:
  matchExpression:
    - key: system-namespace
      operator: DoesNotExist
    - key: user
      operator: Exists
  tenancyLabels:
    - user
```

2. Overload `tenancyLabels` and implicitly apply tenancy rules only to namespaces where `tenancyLabels` are present.

    **Cons**: the simplest option that covers all given use cases.

```yaml
spec:
  tenancyLabels:
    - user
```

### Peers and actions

Based on the existing User Stories, Tenancy only needs action to "skip same tenant" and "deny not same tenant".
Using both actions at the same time doesn't make a lot of sense, since "skip same tenant" action only makes sense if
there is a (B)ANP that will deny tenancy connections. Otherwise, "deny not same tenant" is sufficient.

Tenancy rules don't need to specify separate rules for ingress and egress, because
- for `SameTenant` if ingress is denied, then egress to the same tenant may be allowed when leaving the source pod, but will
  be denied by ingress rule when coming to the destination pod. The same applies to egress.
- for `NotSameTenant` if ingress is denied, then egress from `Tenant1` to `Tenant2` will be allowed by `Tenant1`, but
  considered ingress from another tenant by `Tenant2`, and denied. The same applies to egress.

It means that deny in at least one direction automatically denies the other direction.
Therefore, the only extra parameter Tenancy needs is priority/precedence, and Tenancy rules may look something like:

```yaml
spec:
  action: SkipSameTenant/DenyNotSameTenant
```

### Priorities

Based on User Story 4.4, we need to have Tenancy in the same priority range as ANP and BANP. There are multiple ways to do so:

1. Reuse ANP and BANP to define priority, insert Tenancy rules to `ingress` and `egress`.
Current (B)ANP semantics is: `subject` selects pod to which all `ingress` and `egress` rules are applied.
Tenancy can't be used in this semantics, because it has its own "subject" defined by the tenancy labels.
There are multiple ways to "turn on" tenancy mode in B(ANP) and allow tenancy rules.

Exclusive fields `subject`/`tenancySubject`
```yaml
kind: AdminNetworkPolicy
spec:
  # this field turns on tenancy
  tenancySubject:
    labels: ["user"]
  ingress:
    - action: Allow
      from:
        - SameTenant
        - pods:
            podSelector: {}
```

A similar, but slightly different option with pushing `tenantNamespace` as an alternative subject selector to existing 
`pods` and `namespaces`
```yaml
kind: AdminNetworkPolicy
spec:
  # this field turns on tenancy
  subject:
    tenantNamespaces:
      labels: ["user"]
  ingress:
    - action: Allow
      from:
        - SameTenant
        - pods:
            podSelector: {}
```
**CONS** 
- "switch" that enables some peer types is not the best API design
- gives more flexibility than intended (multiple tenancy definitions)
- conflicts with singleton BANP, meaning that if Tenancy is defined on BANP level, general-purpose BANP selecting different
pods can't be created.

2. Reuse ANP and BANP to define priority, insert Tenancy definition and Rules to `ingress` and `egress`.

It splits the tenancy labels and namespace selector which defines the namespace that are affected by the tenancy.

```yaml
kind: AdminNetworkPolicy
spec:
  priority: 10
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: sensitive-ns
  ingress:
    - action: Allow
      from:
      - tenant:
          labels: ["user"]
          tenant: Same
```

**CONS** complicates selection or affected namespaces, allows strange configuration, where namespaces selected by 
`subject` have no tenancy labels. Has the same problem as the original design.

3. Create 2 objects with ANP and BANP priorities (let's say NetworkTenancy and BaselineNetworkTenancy)

```yaml
kind: NetworkTenancy
spec:
  priority: 10
  tenancyLabels: ["user"]
  action: SkipSameTenant
---
kind: BaselineNetworkTenancy
spec:
  priority: 10
  tenancyLabels: ["user"]
  action: SkipSameTenant
```

While multiple ANPs with the same priority are allowed, we probably can allow multiple Tenancies or Tenancy and ANP
with the same priority, but if we decide to only allow ANP per priority, Tenancy needs to be accounted for in the same range.

**CONS**: BANP doesn't have a priority, to use this method we would need to define a priority for BANP.

4. Create 1 new object with implicit priorities.

`precedence` field + reserved highest-priority rule before (B)ANP
Similar to the previous one, but a bit more flexible:
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels: ["user"]
  precedence: ANP/BANP
  action: SkipSameTenant
```
**CONS** Priorities are implicit and need to be added as extra layers between ANP/NP/BANP.
**PROS** 
- No changes to the existing ANP/BANP objects
- Limited to the use cases we designed it for (smaller chance to shoot yourself in a foot)
- Users that don't care about tenancy can just ignore this CRD
- We can throw it away if we want to change API again


#### Final suggestion

Considering we agree with option 4 from the previous section on priorities specification, we can outline further details here.

For
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels: ["user"]
  precedence: ANP/BANP
  action: SkipSameTenant/DenyNotSameTenant
```

To implement user story 4.3, Tenancy rules should have higher priority than ANP/BANP.
Considering the following priority precedence: ANP Tenancy->ANP->NP->BANP Tenancy->BANP, we can express all mentioned user stories.

<details>
<summary>Full yaml examples (with the initial fields, will be updates as we agree on the final CRD format)</summary>

* 4.1 "overridable isolation"
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: BANP
  action: DenyNotSameTenant
```
OR (second option may be more useful if there is a deny BANP in the cluster)
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: BANP
  action: SkipSameTenant
---
kind: BaselineAdminNetworkPolicy
spec:
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces: {}
  egress:
    - action: Deny
      to:
        - namespaces: {}
```
BANP can also be replaced with deny-all BANP
```yaml
kind: BaselineAdminNetworkPolicy
spec:
  subject:
    namespaces: {}
  ingress:
    - action: Deny
      from:
        - namespaces: {}
  egress:
    - action: Deny
      to:
        - namespaces: {}
```

* 4.2 strict isolation
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: ANP
  action: DenyNotSameTenant
```
OR (second option may be more useful if there is a deny ANP in the cluster)
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: ANP
  action: SkipSameTenant
---
kind: AdminNetworkPolicy
spec:
  priority: 1
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces: {}
  egress:
    - action: Deny
      to:
        - namespaces: {}
```

* 4.3 Allow internal connections for tenants
BANP-level
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: BANP
  action: SkipSameTenant
---
kind: BaselineAdminNetworkPolicy
spec:
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces: {}
  egress:
    - action: Deny
      to:
        - namespaces: {}
```

ANP-level
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: ANP
  action: SkipSameTenant
---
kind: AdminNetworkPolicy
spec:
  priority: 1
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces: {}
  egress:
    - action: Deny
      to:
        - namespaces: {}
```

* 4.4 Tenants interaction with (B)ANP
* 4.4.1 allow from monitoring + deny from not same tenant
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: ANP
  action: SkipSameTenant
---
kind: AdminNetworkPolicy
spec:
  priority: 1
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Allow
      from:
        - namespaces:
            namespaceSelector:
              matchLabels:
                kubernetes.io/metadata.name: monitoring-ns
    - action: Deny
      from:
        - namespaces: {}
```
* 4.4.2 allow from same tenant + BANP deny all
```yaml
kind: NetworkTenancy
spec:
  tenancyLabels:
    - "user"
  precedence: BANP
  action: SkipSameTenant
---
kind: BaselineAdminNetworkPolicy
spec:
  subject:
    namespaces:
      matchExpression:
        - key: user
          operator: Exists
  ingress:
    - action: Deny
      from:
        - namespaces: {}
```
</details>

## Conformance Details

TBD
<!---
(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)
-->

## Alternatives

Other alternatives were mentioned https://docs.google.com/document/d/113xBe7VMK7hMYdIdB9gobp7JwVkWQLnqdMPNkamfaK8/edit,
but none of them cover all the Goals defined in this NPEP.

There are 2 main problems with leaving Tenancy as a (B)ANP peer:
1. tenancy is only based on namespace labels, but (B)ANP subject allows using pod selector too
2. tenancy definition is less obvious, since it is a part of the peers list. Usually peer rules are the same for
all (B)ANP subject pods, but for tenancy that is not true.
3. There are actually more that 2 problems, mainly about allowing much more configurations for tenancy than we have
user stories for, but it is covered in the previous sections :)

Therefore, creating a new object seems like a more clear way to implement Tenancy.

## References

- https://docs.google.com/document/d/113xBe7VMK7hMYdIdB9gobp7JwVkWQLnqdMPNkamfaK8
