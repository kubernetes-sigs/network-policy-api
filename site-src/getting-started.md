# Getting started with Network Policy API

**1. Install a Network Policy API compatible CNI**

There are numerous Container Network Plugin projects that support or are actively working on
supporting the Network Policy API. Please refer to the [implementations](implementations.md)
doc for details on the supportability matrix.

**2. Install the Network Policy API CRDs**

To install the latest released version of the `ClusterNetworkPolicy` API, which is `v0.2.0`, use the following command:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/tags/v0.2.0/config/crd/standard/policy.networking.k8s.io_clusternetworkpolicies.yaml
```

The latest released version of the `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` is `v0.1.7`.
Use the following command to install it in your cluster:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/tags/v0.1.7/config/crd/standard/policy.networking.k8s.io_adminnetworkpolicies.yaml \
-f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/tags/v0.1.7/config/crd/standard/policy.networking.k8s.io_baselineadminnetworkpolicies.yaml
```

We are also going to preserve and maintain the [release-0.1](https://github.com/kubernetes-sigs/network-policy-api/tree/release-0.1) 
branch for the `v1alpha1` APIs, including potential updates and bug fixes until further notice.

**3. Try out one of the sample yamls for specific user stories**

- [Deny traffic at a cluster level](reference/examples.md#sample-spec-for-story-1-deny-traffic-at-a-cluster-level)
- [Allow traffic at a cluster level](reference/examples.md#sample-spec-for-story-2-allow-traffic-at-a-cluster-level)
- [Explicitly Delegate traffic to existing K8s Network Policy](reference/examples.md#sample-spec-for-story-3-explicitly-delegate-traffic-to-existing-k8s-network-policy)
- [Create and Isolate multiple tenants in a cluster](reference/examples.md#sample-spec-for-story-4-create-and-isolate-multiple-tenants-in-a-cluster)
- [Cluster Wide Default Guardrails](reference/examples.md#sample-spec-for-story-5-cluster-wide-default-guardrails)
