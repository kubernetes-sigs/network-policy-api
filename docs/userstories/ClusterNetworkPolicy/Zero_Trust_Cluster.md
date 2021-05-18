### Summary

As a cluster admin, I want all Pods in my cluster to have a default zero-trust
deny-all security posture, so that all traffic allowed in the cluster must be
explicitly called out (by separate higher precedence policies).

### Description

Kubernetes imposes the fundamental requirement on CNIs that, all Pods should be
able to communicate with all other Pods by default, if there's no intentional
network segmentation policies in the cluster. However, a cluster admin often
needs a stricter default security posture, a model where all workloads start with
no ingress and egress access from/to anywhere at all, unless explicitly allowed
by separate policies. This model can ensure that all allowed traffic flows in the
cluster will match at least one policy rule.

### Acceptance Criteria

If I apply a zero-trust deny-all policy on a Kubernetes cluster, all Pods will
be isolated, unless there are higher-order policy rules applied.
