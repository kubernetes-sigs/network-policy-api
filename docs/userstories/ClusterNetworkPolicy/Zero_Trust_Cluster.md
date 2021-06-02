### Summary

As a cluster admin, I want all Pods in my cluster to have a default zero-trust
deny-all security posture, so that all traffic allowed in the cluster must be
explicitly called out, by separate higher precedence policies. Common examples of
such explicitly allowed traffic include allowing all workload Pods to be able to
talk to the cluster DNS service, and similarly, allowing all workload Pods to talk
to logging/monitoring Pods running on the cluster, such as Prometheus.

### Description

Kubernetes imposes the fundamental requirement on CNIs that, all Pods should be
able to communicate with all other Pods by default, if there's no intentional
network segmentation policies in the cluster. However, a cluster admin often
needs a stricter default security posture, a model where all workloads start with
no ingress and egress access from/to anywhere at all, unless explicitly allowed
by separate policies.

This model can ensure that all allowed traffic flows in the cluster will match
at least one policy rule. In practice, the zero-trust cluster policy should be
deployed with a set of guardrail rules that ensures the minimum communication
available for cluster add-on services, such as DNS and logging/monitoring. In
the cluster example above, those policies combined would effectively translate
into: isolate all Pods in the cluster, except that all Pods can communicate with
<app=kube-dns> Pods on DNS ports, and all Pods can communicate with
<app=prometheus> Pods on 9090.

### Acceptance Criteria

If I apply a zero-trust deny-all policy on a Kubernetes cluster, all Pods will
be isolated, unless there are higher-order policy rules applied.

If I apply a zero-trust deny-all policy together with a set of higher-order
allow rules, only the traffic explicitly listed in the higher-order allow rules
will be permitted. All other ingress/egress traffic from/to Pods in the cluster
will be denied.
