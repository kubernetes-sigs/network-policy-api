---
date: 2025-10-05
authors:
  - frozenprocess
---

# The Road to a More Scalable and Unified Network Policy API

The Admin and Baseline Network Policies (ANP and BANP) have been available in `v1alpha1` for some time, with several open source [implementations](/implementations) giving security enthusiasts a sneak peek into how they can experiment with and adopt them in real-world environments. In this blog, we’ll go through some of the new changes introduced in v1alpha2 and 🤞 the upcoming beta!

In particular, the sub-group has been addressing complexities in the existing two APIs, along with making terminology changes to better represent their functionality. A central part of this effort is rethinking how cluster-wide and namespace-level policies interact, ensuring the API evolves into a more consistent and powerful tool for Kubernetes users. To achieve this, the separate APIs are being unified into a single, modular, and extensible design.

![Image](./Resources/1to2-transition.svg)

Keep in mind, that we always welcome volunteers to help accelerate this effort. If you’re interested, you can learn more about how to get involved [here](https://github.com/kubernetes-sigs/network-policy-api?tab=readme-ov-file#community-discussion-contribution-and-support).

<!-- more -->

## Quick reminder

Regardless of the names or the fancy technical terms, the ultimate goal here is to provide a more robust and comprehensive way to secure your environment.
The `v1alpha1` API addressed this by introducing two resources, `ANP` and `BANP`, which brought features like **Action, Priority, and multi-namespace reach** to Kubernetes security, among others. These were capabilities many of us felt were missing when working with traditional Kubernetes network policies.

To give you a clear sense of what to expect from this blog, we’ll dive into the following bullet points in detail:

1. [Unifying Admin and Baseline Network Policies](#unifying-admin-and-baseline-network-policies)
    1. Changes to Action
    1. Policy Types (Tier)
    1. Docs and Versioning Strategy
1. [A Modular Architecture for Network Policies](#a-modular-architecture-for-network-policies)
1. [The Path to Beta (Get involved!)](#the-path-beta-get-involved)


## Unifying Admin and Baseline Network Policies

Both `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` are being replaced by the new `ClusterNetworkPolicy` resource. Introduced in `v1alpha2`, this resource consolidates and enhances the functionality of its predecessors, making cluster scaled network policy management more streamlined and efficient.

### Changes to Action

`ClusterNetworkPolicy` supports the **Accept, Deny, and Pass** actions. Note that the previous **Allow** action has been renamed to **Accept** to make its behavior clearer to users.

### Policy Types (Tier)

The `ClusterNetworkPolicySpec` now includes a `Tier` field, which can be set to either `Admin` or `Baseline`.
The `Admin tier` corresponds to the old `AdminNetworkPolicy` and is evaluated **first**. The `Baseline tier` corresponds to the old `BaselineAdminNetworkPolicy` and is **evaluated after** the `NetworkPolicy tier`.

***

## A Modular Architecture for Network Policies

The group is refactoring its reference implementation to improve modularity and extensibility. This new architecture employs a plugin-based system where new policy evaluators can be added without modifying the core logic. This design is intended to simplify the process of adding new features and experimenting with new ideas. The new design also incorporates a **Cube IP tracker** to help with scalability issues by abstracting pod information and reducing the load on the API server. This approach is seen as a way to address issues with network policies that depend on IPs and pods, which have high churn and can heavily impact the API server.

***

## The Path Beta (Get involved!)

The immediate priority for the subgroup is to finalize v1alpha2, which will pave the way for a stable v1beta1 candidate. This timeline, however, depends on receiving timely API reviews from the busy Kubernetes API approvers.

An important part of this process is updating the documentation to reflect all the new API changes and the project’s updated name. The team plans to complete these documentation updates once the core API changes are merged.

# What’s Next

The subgroup aims to:

- Finalize `v1alpha2` implementation.
- Prepare a beta release. [details](https://github.com/orgs/kubernetes-sigs/projects/32)
- Continue tightening conformance, documentation, and image publishing automation. ([Get involved we can always use some help](https://github.com/kubernetes-sigs/network-policy-api?tab=readme-ov-file#community-discussion-contribution-and-support))
