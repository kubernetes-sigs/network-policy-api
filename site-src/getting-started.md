# Getting started with Network Policy API

**1. Install a Network Policy API compatible CNI**

There are numerous Container Network Plugin projects that support or are actively working on
supporting the Network Policy API. Please refer to the [implementations](implementations.md)
doc for details on the supportability matrix.

**2. Install the Network Policy API CRDs**

We don't have a released version of the `ClusterNetworkPolicy` API yet, but you can try out
the latest version from the `main` branch by applying the following manifest:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/heads/main/config/crd/standard/policy.networking.k8s.io_clusternetworkpolicies.yaml
```

The latest released version of the `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` is `v0.1.7`.
Use the following command to install it in your cluster:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/tags/v0.1.7/config/crd/standard/policy.networking.k8s.io_adminnetworkpolicies.yaml \
-f https://raw.githubusercontent.com/kubernetes-sigs/network-policy-api/refs/tags/v0.1.7/config/crd/standard/policy.networking.k8s.io_baselineadminnetworkpolicies.yaml
```

**3. Try out one of the sample yamls for specific user stories**

- [Deny traffic at a cluster level](reference/examples.md#sample-spec-for-story-1-deny-traffic-at-a-cluster-level)
- [Allow traffic at a cluster level](reference/examples.md#sample-spec-for-story-2-allow-traffic-at-a-cluster-level)
- [Explicitly Delegate traffic to existing K8s Network Policy](reference/examples.md#sample-spec-for-story-3-explicitly-delegate-traffic-to-existing-k8s-network-policy)
- [Create and Isolate multiple tenants in a cluster](reference/examples.md#sample-spec-for-story-4-create-and-isolate-multiple-tenants-in-a-cluster)
- [Cluster Wide Default Guardrails](reference/examples.md#sample-spec-for-story-5-cluster-wide-default-guardrails)
