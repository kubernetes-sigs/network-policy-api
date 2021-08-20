# The NetworkPolicy++ Subproject 'super-kep' Copied from [Jay's Repo](https://github.com/jayunit100/network-policy-subproject)


# NEWS

Currently were actively working on 4 KEPs to improve the 
policy API, summarized here, https://docs.google.com/document/d/1_clStao-uM3OblOTsA4Kgx2y4C9a6KXmdOywW4tFSLY/edit#
our agenda as always is here https://docs.google.com/document/d/1AtWQy2fNa4qXRag9cCp5_HsefD7bxKe3ea2RPn8jnSs/edit#. 


----------------------------

This repository attempts to capture all of the work done by the NetworkPolicy working group as a series of documents which
can be reused, and formulated into a series of KEPs, CRDs, and proposals over time, based on ideas the communtiy came up with on how to improve the NetworkPolicy API.

## Context

After about 4 months of deliberation over various network policy user stories, the Sig-network network-policy subgroup (retroactively formed in July of 2020) decided that it was time to

- start looking at how the networkpolicy api could be improved by actively designing changes against it
- create a combination of hacking on KEP-like docs and possibly CRD(s)
- while formulating an overall strategy to tie those docs together

These documents are formatted similarly to KEPs for the sake of uniformity.  Some of these documents may be submitted as full blown KEPs once concepts are clarified.

## Motivation

Kubernetes is becoming the API equivalent of the OS for the cloud. The
combinations of cloud fabrics, host operating systems and Kubernetes
distributions make this easier.


- Kubernetes is becoming the OS of the cloud
- Users prefer vendor specific APIs due to lack of Kubernetes Network Policy API
  deficiencies
- Cluster administrators need APIs beyond current devloper focused APIs for securing the cluster, not just apps
- Unified network security approach can simplify auditing
- APIs are confusing for basic use cases like default deny
- People want to secure apps using other abstractions in addition to label selectors

### Goals



- Don't break anything
- Simplify migration when using new capabilities
- Develop a new administrator API so that RBAC can be used to separate roles if desired
- Develop a new set of APIs for application policies
- Deliver value as quickly as possible
- Better documentation!
- Generalize the concept of endpoint (not just pods)
- Comprehensive error reporting

### Non-Goals

- Graphical User Interfaces (GUIs) are out of scope
- Protecting non Kubernetes workloads
- Intra-pod restrictions (i.e., process restriction, kubectl exec)
- Support for encryption of traffic
- Support NetworkPolicies spanning multiple clusters

## Deliverables

The following are a set of deliverables that will probably have value over time to people designing a V2 API.

### Concepts

The first work the group did is outline the basic conceptual *needs* that users have when administrating
security across clusters.

#### Part 1: User Stories

Owners: Cody, Jay, Andrew, Ricardo

The user stories proposed by various members of the SIG are described here.

Please pull request any new user stories along with a suggested priority if you have new ideas.

1) [prioritized user stories](p0_user_stories.md)

#### Part 2: Concepts

Owners: Abhishek (maybe dan?)

The following set of conceptual docs outlines the progression from user stories towards
high level actionable goals (no KEPs or detailed API changes suggested here).

1) [Security model proposal](proposals/1_model.md)
2) [Create/update developer focused API](proposals/1_dev_scoped.md)
3) [Creation of a cluster-scoped API](proposals/1_cluster_scoped.md)

### Implementation

#### Part 3a: KEPS

Owners: Ricardo, Chris

Although not all of the concepts above have a straight forward implementation, some very clear and obvious improvements to the NetworkPolicy API can be conservatively made.  These will be iterated on over time.

<TODO>

#### Part 3b: CRDs

Owners: Anish, Jay, Abhishek, Yang

Some people just feel the need to hack.  Given that many concepts can be implemented over time as CRDs which create or manage large numbers of network policy objects, as an R&D project, experimental iteration on a NetworkPolicy Operator, similar to that of https://github.com/rushabh268/kubernetes-network-policy-controller, will be considered as part of the ongoing R&D of this group.

<TODO>

#### Templates

Owners: Gobind

Since many people simple want easy ways to iterate use policies, now, a collection of links to templates and tutorials on how to better use policies in their existing state [templates](templates.md).

#### Suggestion

After evaluating these docs, this group proposes *blah*.
