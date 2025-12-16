# NPEP-248: Add explicit match all traffic option to peers

* Issue: [#248](AdminNetworkPolicy Core can't express allow/deny all)
* Status: Provisional

## TLDR

Add a new, explicit "all", option to `ClusterNetworkPolicyXXXPeer` in order to
explicitly match all traffic.

## Goals

Make CNP more useful by extending it to be able to write blanket allow/deny 
rules covering all traffic, not just pod-ot-pod traffic.

## Non-Goals

Using tenancy as an example to show flexibility of combining priority, pass
action and a "deny all", but not trying to fully solve tenancy in this NPEP. 

## Introduction

In order to avoid back-compatibility headaches, the SIG has adopted a policy
that ClusterNetworkPolicyXXXPeer structs must have exactly one field set.  This
allows implementations that see an empty peer struct to infer that there is an
unknown field, and then "fail closed".

Since we've made the empty struct invalid, there's no way to match "all" traffic, 
this NPEP aims to add an explicit way to say "all".

## User-Stories/Use-Cases

### Story 1: Deny all traffic that is not explicitly allowed by higher precedence rules

As a cluster admin, I want to adopt a default deny security posture, where the
only traffic allowed to my cluster's pods is allowed by CNP (or v1 NP) so that
I minimise my cluster's attack surface.

This goes beyond the original KEP, which only talks about in-cluster traffic.
However, it seems like a very natural thing to need as soon as CNP is put into
practice!

(I.e. I need to be able to write deny all in the baseline tier.)

### Story 2: Tenancy via automated CNP creation

As a cluster admin, I want to enforce tenancy boundaries between groups of 
namespaces by auto-creating per-namespace CNPs; I want this to fail closed 
(block all traffic) in the window between namespace creation and per-namespace 
CNP creation.

(I.e. I need to be able to write deny all in the middle of the admin tier, after my 
"pass same namespace" policies.)

### Story 3: Avoid confusion between 0.0.0.0/0 and ::/0

As a cluster admin, I want a way to write "all traffic" that is less error-prone
than writing both IPv4 and IPv6 CIDRs so that I don't accidentally leave a whole
plane of traffic open (especially when upgrading cluster from IPv4 to dual stack).

### Story 4: Allow all traffic to a pod on a particular port from anywhere

As a cluster admin, I want to allow all traffic to, say, kube-dns port 53 from
inside and outside the cluster so that I can make kubernetes services available
over DNS from anywhere.

## API

Starting with YAML examples for discussion.  A new field "all" is added to the
ingress/egress peer objects; it takes a sigil `{}` value

```yaml
spec:
  ingress:
  - action: Deny
    from:
    - all: {}
  egress:
  - action: Allow
    from:
    - all: {}
```

**Note:** Since we're explicitly using the empty struct here, we won't be able
to add new fields inside the struct later without a breaking change.  "Old"
implementations won't see the new fields, and they'd act like they weren't present.

I think this is OK in this unique case because "all" is naturally a unique "top" 
value for the set of matches.

### What does "all" cover?

This is slightly slippery due to the differences in networking approaches that 
are out there.

I think "all" should match all reasonable workload traffic that the network and
policy engines both support.  In practice, that currently means "all IP traffic"
with the following exceptions:

* In line with v1 NetworkPolicy, Kubelet to Pod traffic is always allowed to 
  allow health checks.
* Implementations may block or restrict  certain traffic to maintain security
  or network integrity. For example: blocking IP spoofing or overlay packets
  from pods.

### Examples: 

#### Story 1: Deny all traffic that is not explicitly allowed by higher precedence rules

This can be done with a policy at the end of the baseline tier.

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: default
spec:
  tier: Baseline
  priority: 1000
  ingress:
  - action: Deny
    from:
    - all: {}
  egress:
  - action: Deny
    from:
    - all: {}
```

### Story 2: Tenancy via automated CNP creation

Being able to write "deny all" allows for closing the policy creation race for
new tenants.

When provisioning the cluster, we create a low-precedence admin-tier policy
that denies all traffic to/from tenant-labelled namespaces.

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: default
spec:
  tier: Admin
  priority: 1000
  subject:
    namespaces:
      matchExpressions: 
        key: "tenant"
        operator: Exists
  ingress:
  - action: Deny
    from:
    - all: {}
  egress:
  - action: Deny
    from:
    - all: {}
```

Then, via to-be-written controller, we auto-generate a higher precedence policy per unique tenant 
at namespace creation time.  For example, a policy that delegates intra-tenant traffic 
to v1 NetworkPolicy and does not allow traffic outside the cluster.

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: tenant-foo
spec:
  tier: Admin
  priority: 900
  subject:
    namespaces:
      matchLabels: 
        tenant: foo
  ingress:
  - action: Pass
    from:
    - namespaces:
      matchLabels:
        tenant: foo
  egress:
  - action: Pass
    to:
    - namespaces:
      matchLabels:
        tenant: foo
```

Result:

* When new tenant namespaces are created, their traffic is denied by default.
* Then, the controller creates the per-tenant policy, opening up the desired access.
  If the controller is down / slow / fails, the system fails closed.

## Conformance Details

TBD

## Alternatives

### Add networks to ingress peers and use CIDRs

Egress peers already have networks and this could be used to write "all IP traffic"
by writing two rules: one for 0.0.0.0/0 and one for ::/0.  

Adding networks to ingress peers has been proposed in [NPEP-127](https://github.com/kubernetes-sigs/network-policy-api/pull/249/files).

Pros:

* Already have this for egress.

Cons:

* Writing two rules is error-prone.  It's easy to forget the IPv6 one if you're
  IPv4-minded, leaving a whole plane unsecured.
* When upgrading to dual stack, it's especially easy to forget to add the 
  extra rule.
* Standardising network matches for ingress is non-trivial due to SNAT at ingress.
  While the 0.0.0.0/0 and ::/0 CIDRs are unambiguous, different CNI plugins and 
  load balancers do varying amounts of SNAT, making the general feature hard to use.

### Alternative YAML representations:

#### Use a bool

```yaml
spec:
  ingress:
  - action: Deny
    from:
    - all: true
  egress:
  - action: Deny
    from:
    - all: true
```

Pros:

* `all: true` reads well

Cons:

* `all: false` becomes possible and meaningless

#### Add a sigil field inside the object

```yaml
spec:
  ingress:
  - action: Deny
    from:
    - all:
        all: true
  egress:
  - action: Deny
    from:
    - all: 
        all: true
```

Pros:

* Can extend the new object later.

Cons:

* It stutters badly and ends up fiddly to write.

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
