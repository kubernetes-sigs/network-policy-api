---
date: 2025-10-09
authors:
  - npinaeva
---

# API update for v1alpha2: `ClusterNetworkPolicy` replaces `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy`

We have merged `v1alpha1.AdminNetworkPolicy` and `v1alpha1.BaselineAdminNetworkPolicy` into a single API in `v1alpha2.ClusterNetworkPolicy`.

AdminNetworkPolicy (ANP) and BaselineAdminNetworkPolicy (BANP) were the first APIs created by the Network Policy API working group.
They are `v1alpha1` resources, which means that they are not stable and are mostly used to get early feedback on the API design from
the community.

If you have seen these APIs, you probably have noticed that they are quite similar. If you have written controllers to reconcile these resources,
you probably also found some code duplication. The original reason for having two separate resources was just a lack of
use cases for multiple instances of BANP, which made it a singleton. However, with more feedback from the community, we have realized that
those use cases do exist and that having two separate resources adds more burden than value.

<!-- more -->

### What has changed

As a reminder, here is how the original Admin Network Policy API Model looked like:

![image](/images/ANP-api-model.png)

The model stays the same with the new API, the only difference is that evaluation order is now defined by the `tier` field
and not by the resource type. `AdminNetworkPolicy` becomes a `ClusterNetworkPolicy` with `tier=Admin` and 
`BaselineAdminNetworkPolicy` becomes a `ClusterNetworkPolicy` with `tier=Baseline`.

<img src="/images/CNP.drawio.svg" width="541" alt="Cluster Network Policy API model">

Other changes bring the functionality of `BaselineAdminNetworkPolicy` to parity with `AdminNetworkPolicy`. This includes:

- Allowing multiple `ClusterNetworkPolicy` resources with `tier=Baseline` by using the same `priority` field as for `tier=Admin`.
- Supporting `Pass` action in `ClusterNetworkPolicy` with `tier=Baseline` to allow skipping all further rules in the `Baseline` tier.
- Supporting `domainNames` matching for `egress` rules in `ClusterNetworkPolicy` with `tier=Baseline`.

The [enhancement proposal](https://github.com/kubernetes-sigs/network-policy-api/pull/289) has some more details for those interested.

### Examples

Let's take a look at some examples based on the original [user stories](../../user-stories.md).

#### Story 1:Deny traffic at a cluster level

To deny all traffic at the cluster level, the following `AdminNetworkPolicy` was used:

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: cluster-wide-deny-example
spec:
  priority: 10
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: sensitive-ns
  ingress:
    - action: Deny
      from:
      - namespaces:
         namespaceSelector: {}
      name: select-all-deny-all
```

which looks like this with the new `ClusterNetworkPolicy` API:

```yaml
--8<-- "user-story-examples/user-story-1.yaml"
```

#### Story 5: Cluster Wide Default Guardrails

To deny all traffic in a cluster by default (in an overridable manner), the following `BaselineAdminNetworkPolicy` was used:

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: BaselineAdminNetworkPolicy
metadata:
  name: default
spec:
  subject:
    namespaces: {}
  ingress:
    - action: Deny   # zero-trust cluster default security posture
      from:
      - namespaces:
          namespaceSelector: {}
```

which looks like this with the new `ClusterNetworkPolicy` API:

```yaml
--8<-- "user-story-examples/user-story-5.yaml"
```

### Migration

We appreciate all early adopters of the `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` APIs. 
You can continue using them at their latest released version [v0.1.7](https://github.com/kubernetes-sigs/network-policy-api/releases/tag/v0.1.7)
At the same time, we encourage you to plan your migration to the new `ClusterNetworkPolicy` API, which can be done once its first version is available.
We plan to base our `beta` release on the `ClusterNetworkPolicy` API.
