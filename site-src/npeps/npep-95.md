# NPEP-95: NPEP template

* Issue: [#95](https://github.com/kubernetes-sigs/network-policy-api/issues/95)
* Status: Provisional

## TLDR

(1-2 sentence summary of the proposal)

## Goals

(Primary goals of this proposal.)

## Non-Goals

(What is explicitly out of scope for this proposal.)

## Introduction

(Can link to external doc -- but we should bias towards copying
the content into the NPEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

## User-Stories/Use-Cases

(What new user-stories/use-cases does this proposal introduce?)

A user story should typically have a summary structured this way:

1. **As a** [user concerned by the story]
2. **I want** [goal of the story]
3. **so that** [reason for the story]

The “so that” part is optional if more details are provided in the description.
A story can also be supplemented with examples, diagrams, or additional notes.

e.g

Story 1: Deny traffic at a cluster level

As a cluster admin, I want to apply non-overridable deny rules to certain pod(s)
and(or) Namespace(s) that isolate the selected resources from all other cluster
internal traffic.

For Example: The admin wishes to protect a sensitive namespace by applying an
AdminNetworkPolicy which denies ingress from all other in-cluster resources
for all ports and protocols.

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
