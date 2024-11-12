## v0.0.1-policy-assistant

This release contains the `policy-assistant` Command-Line Interface (CLI) and its source code.

Policy Assistant is a project to help users develop/troubleshoot upstream network policies. Current APIs: NetworkPolicy (v1), AdminNetworkPolicy and BaselineAdminNetworkPolicy.

`policy-assistant` is a static analysis tool which can simulate policy verdicts for traffic.
`policy-assistant` can either read policies/pods from file or from a Kubernetes cluster.

For more information, see the [Policy Assistant README](https://github.com/kubernetes-sigs/network-policy-api/blob/main/cmd/policy-assistant/README.md) or [this demo](https://github.com/kubernetes-sigs/network-policy-api/blob/main/cmd/policy-assistant/examples/demos/walkthrough/README.md).

### What's New

Inaugural release for `policy-assistant`.

### Supported APIs

- NetworkPolicy v1 (networking.k8s.io/v1)
- AdminNetworkPolicy and BaselineAdminNetworkPolicy [v1alpha1 (policy.networking.k8s.io/v0.1.1)](https://github.com/kubernetes-sigs/network-policy-api/releases/tag/v0.1.1)

### Special Notes

We will be iterating on how we version policy assistant.
It's possible that future releases will not follow the same release version format.
