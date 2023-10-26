# NPEP-136: DeveloperNetworkPolicy

* Issue: [#136](https://github.com/kubernetes-sigs/network-policy-api/issues/136)
* Status: Provisional

## TLDR

I want to have the next generation of NetworkPolicies for Application Developers as an addition to Admin-focused APIs.

## Goals

Add the missing part of "NetworkPolicy v2": ANP API is the cluster-scoped and is created for cluster admins, 
DeveloperNetworkPolicy (DNP for short) is namespace-scoped and is created for application developer/namespace owners.

## Non-Goals

This is not the next version of an existing NetworkPolicy API, therefore it doesn't need to be compatible with
NetworkPolicy, but it needs to be fully interoperable (i.e. have well-defined behaviour when used with NetworkPolicy).

## Introduction

Mostly coming from https://docs.google.com/document/d/10L9_ZV7nmabObFSKs4rR_t8ATtOTBPwsZKKSB6GFYp0/edit

We want to build a new API, starting from scratch, and not building on top of the NetworkPolicy. Therefore, this NPEP
will define the developer-focused user stories from scratch. There are some general properties that are desired for the new API
(that we mostly learned thanks to NetworkPolicy)

**Note** the following points apply to ANP/BANP too

1. It needs to be extensible

    We keep running into the problem in NetworkPolicy that we want to add new “stuff”, but the new stuff would be 
    invisible to old implementations. We need to fix this in DNP.

2. It needs to be explicit

   - It should not have “implicit deny on first policy”. Denies, like accepts, should be explicit.
   This implies that you should be able to create a policy that says “allow X”, where if X wasn’t already denied 
   (eg by BaselineAdminNetworkPolicy) then the “allow X” would just be a no-op (rather than causing other traffic to be denied like with NPv1).
   - It should not have the automatic “connections from the pod’s node IP are always accepted” exception.
   (but you shouldn’t have to manually add that rule either… we probably need to make probes work differently)
   TODO: add link to the probes discussion
   - It should not have tricky “arrays of objects where multiple fields can be set but none are required” 
   such that adding or removing a single “-” drastically changes the meaning of the policy.
   - “Present but empty” / “Present but zero length” should not mean something different from “not present” / nil. 
   Likewise, the difference between an array of length 0 and an array of length 1 should be the same as the difference 
   between an array of length 1 and an array of length 2.

DeveloperNetworkPolicy should work well together with ANP and BANP as a part of the same API.

## User-Stories/Use-Cases

### User Story 1: Isolate namespace for ingress

As a namespace owner, I want to restrict all incoming connections from outside the namespace, because I don't trust
pods from the other namespaces.

### User Story 2: Isolate namespace for egress

As a namespace owner, I want to restrict all outgoing connections outside the namespace. I control
what pods in the namespace are doing, and I know that they are not configured to connect to the endpoints outside the namespace.
I want to make sure that if hackers or bugs will change that behaviour, outgoing connections won't succeed.

### User Story 3: Allowlist egress external endpoints, deny ingress

As a namespace owner, I want to allow pods in my namespace to only access specific external endpoints 
that I define by CIDR and/or FQDN and deny everything else, including incoming connections.

### User Story 4: Explicitly allow required connectivity with deny-all BANP

Cluster administrator defined a deny-all egress BANP to require all namespace explicitly declaring what egress connectivity they need.
As a namespace owner, I want to list required egress endpoints by CIDR and/or FQDN rules. 

### User Story 5: Communication between namespaces

As an owner of multiple namespaces, I want to allow and deny connections between namespaces.
I have 3 namespaces: `db`, `backend`, `server`, and I want to define the following policies:
1. namespace `db` only allows incoming connection from the `backend` namespace, all outgoing connection are denied.
2. namespace `backend` only allows incoming connections from the `server` namespace, and only allows outgoing connections
to the `db` namespace.
3. `server` namespace allows all incoming connections, and only allows outgoing connections to the `backend` namespace.

### User Story 6: Communication within a namespace

As a namespace owner, I have 3 pods in my namespace: `db`, `backend`, `server`. I want to define the same policy as in Story 5,
but inside one namespace.

### User Story 7: TODO: there are still more cases to cover

### Less obvious User Stories (may be better postponed to the next iteration, after we agree on the most basic cases)
- external ingress to the namespace (requires integration with the Ingress/Gateway API)
- using nodes as peers (struggling to figure out the use case, may be something related to hostNetwork pods)
- using services as peers (was never properly discussed before, but sounds like a valid case)


## API

Interoperability with NetworkPolicy will be explained here, since it is not a use case.

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

Don't do anything, just keep using NP. (that means we can't improve anything)

## References
