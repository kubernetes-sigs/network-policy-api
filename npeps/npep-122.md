# NPEP-122: Tenancy API

* Issue: [#122](https://github.com/kubernetes-sigs/network-policy-api/issues/122)
* Status: Provisional

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
    
    As a cluster admin, I want to build tenants in my cluster and always allow connections inside one tenant.
    At the same time I want to setup an overridable deny-all policy to protect namespaces by default.
    This policy should make sure internal connectivity for a tenant is always allowed, in case there are
    lower-priority deny rules.

#### Story 4.4: Tenants interaction with (B)ANP

    As a cluster admin, I want to be able to setup policies with higher and lower priority than tenancy policy.
    I want to deny inter-tenant connections, but I want to allow ingress traffic from the monitoring namespace
    to all namespaces. Therefore, I need "allow from monitoring" rule to have higher priority than "deny from
    other tenants".
    I want to setup "deny all" BANP to protect cluster workloads, but I want to allow internal connections
    within tenant. Therefore, I need "allow from same tenant" rule to have higher priority than "deny all" BANP.

#### What I couldn't figure out user stories for

- Skip action
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

TBD

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
