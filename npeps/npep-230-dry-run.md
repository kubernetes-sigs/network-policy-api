# NPEP-95: NPEP template

* Issue: [#230](https://github.com/kubernetes-sigs/network-policy-api/issues/230)
* Status: Provisional

## TLDR

Add dry-run mode for (B)ANP to allow "disabling" policies without deleting them.
This should allow implementations add extra audit/monitoring/logging capabilities on top of this feature.

## Goals

A (B)ANP dry-run mode should not affect any connections, i.e. it should be treated as if did not exist.
Networking plugins can interpret the contents of the object to provide feedback (e.g. via logging or observability tools) 
to see which connections will be dropped/allowed once this (B)ANP is enforced.

## Non-Goals

Define exact logging/observability formats for network plugins.

## Introduction

Users may want to ensure no unexpected connections will be denied by a new (B)ANP.
A dry-run mode should not affect any connections, and allow the networking plugin to provide feedback 
(e.g. via logging or observability tools) to see which connections will be dropped/allowed once this (B)ANP is enforced.

## User-Stories/Use-Cases

As a cluster admin, I am designing new ANPs for my cluster and want to make sure applying them won't
have any unexpected effects. To do so, I want to apply ANPs in a dry-run mode and get feedback from the network
plugin on which connections would be dropped/allowed.

As a cluster admin, I want to get feedback from the network plugin for currently allowed/denied connections
with (B)ANP. 
This one needs more discussion as it is already do-able with extra labels/annotations, because adding
logging/audit/observability while enforcing (B)ANP is a side effect and doesn't break (B)ANP semantics.
But since we are considering adding a new flag, it may be useful to add this option with vague description, like 
"Enables logging/audit/observability as defined by the networking plugin. May also have no effect is the plugin doesn't support it."

### Dry-run for changing an existing (B)ANP

Calico's StagedPolicy allows analyzing the effect of changing an existing policy. That approach requires a new CRD (StagedPolicy),
that is an exact copy of the original policy, then you can create a StagedPolicy with the same name as an already existing policy,
and that would meat that you intend to replace the existing policy with the staged one.

Besides a new CRD, this approach also requires some extra simulation logic, because existing policy should be applied on
one hand, and be replaced in the simulation on the other hand.

Simulating the (B)ANP change is possible with the simple dry-run mode, but it requires creating a new (B)ANP representing 
the "diff" between existing and a new policy config. This approach is less user-friendly though.

## API

(... details, can point to PR with changes)

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

- Use Policy Assistant https://github.com/kubernetes-sigs/network-policy-api/issues/221, may have some limitations. Needs additional discussion.
- Leave it to be implementation-specific. This required a copy of ANP CRD that changes behaviour (doesn't really apply) with dry-run flag.

## References

Similar features:
- calico: https://docs.tigera.io/calico-cloud/network-policy/staged-network-policies
- cilium Issue: https://github.com/cilium/cilium/issues/9580

