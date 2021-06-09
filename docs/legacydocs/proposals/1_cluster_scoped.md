# A Cluster Scoped extension to the NetworkPolicy API

To get more context on this document please refer to this original [proposal](https://docs.google.com/document/d/10t4q5XO1ED2PnK3ishn4y3G4Tma7uMYgesG-itQHMiU/edit#).

## Summary

This proposal outlines a plan for implementing *cluster level* NetworkPolicies
in the Kubernetes API.

## Motivation

Current NetworkPolicy API is tailored towards application developers to protect
their workloads from unintended usage. Thus, the NetworkPolicy API is unable to
satisfy some of the use cases which are relevant to cluster administrators.

### Goals

The goal of this initiative is to develop a new security model with the
following in mind:

- Must be able to express the intent of a cluster administrator
- Ability to set the default security policy for the cluster and Namespaces

### Non-Goals

This document does not aim to achieve the following goals:

- Enable traffic policies that affect things outside the cluster
- Secure traffic between multiple clusters
- Cluster level monitoring or threat detection
- CLI/devOps tool to gain more visibility in the cluster

This does not mean that the above said goals will be ignored, in fact they
will be worked upon and documented in a separate KEP within this repository.

## User stories

Please refer to [this doc](../p0_user_stories.md) for detailed administrator
focused user stories collected from the community.

## Design Details

<TODO abhishek...>

## Alternatives

<TODO abhishek...>
