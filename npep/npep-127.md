# NPEP-127: Add southbound traffic support in (B)ANP API

* Issue: [#127](https://github.com/kubernetes-sigs/network-policy-api/issues/127)
* Status: Provisional

## TLDR

This NPEP proposes adding support for cluster ingress (southbound) traffic control in the `AdminNetworkPolicy` and `BaselineAdminNetworkPolicy` API objects.

## Goals

* Implement ingress traffic control from external destinations (outside the cluster)

## Non-Goals

(What is out of scope for this proposal.)

## Introduction

Currently, (B)ANP only allows to specify [peers](https://github.com/kubernetes-sigs/network-policy-api/blob/46be90096f23934f56342bc408b94a1aa79a159c/apis/v1alpha1/shared_types.go#L98-L103) 
based on pod and namespace selector, that will only apply to non-hostNetwork cluster pods. This enhancement aims to
add more types of incoming connections that may be specified. A similar effort exists for the [egress](https://github.com/kubernetes-sigs/network-policy-api/issues/126)
direction.

By default, cluster workloads are not accessible from outside the cluster, but there are many ways to expose these workloads
(either with k8s APIs like [Ingress API](https://kubernetes.io/docs/concepts/services-networking/ingress/) or
[Gateway API](https://gateway-api.sigs.k8s.io/), or with downstream implementations,
like [Openshift's Route CR](https://docs.openshift.com/container-platform/4.13/rest_api/network_apis/route-route-openshift-io-v1.html)).
Some of the APIs may provide functionality to filter incoming connections, but Network Policy API is different, because
it is applied to the cluster workloads, and should provide the required level of security regardless of how/if a cluster workload
is exposed to the outer world.

### Use cases
Most of the use cases refer to ANP priority, meaning the policy should be non-overridable by namespace owners, unless
other clarification are mentioned in the use case itself.

 - zero trust policy

    As a cluster administrator when implementing a zero trust policy I want to make sure no ingress connection 
    (unless explicitly allowed) will get to the cluster workloads. This should apply to all connecitons, including traffic from outside the cluster.

 - block well-known ports

    As a cluster administrator I decided that some well-known services like ftp, telnet, SNMP, etc. 
    should not be allowed for cluster workloads. To implement this policy I want to explicitly deny all ingress connections
    for TCP ports 21, 23, 161, etc.

 - compromised node protection
 
    My cluster has a set of nodes with very sensitive workloads, as a cluster administrator I want to make sure that if 
    a worker node A is compromised, it will not be able to affect other worker nodes. To do so, I want to create a cluster-wide policy to
    deny ingress connections from some of the cluster nodes. I may also need to explicitly allow access from the control plane nodes.
    This security model assumes all pods from the same namespace are scheduled to the same node, and ingress connections
    from pods located on the compromised node may be denied with existing namespace selectors.

 - allow external endpoints

    I have a cluster (management cluster) that creates other clusters on request (hosted clusters).
    Management cluster assigns a subnet for hosted cluster workloads. As a cluster administrator I want to allow
    incoming connections from the hosted clusters identifying them with assigned CIDRs


- external services

    I have an external service to collect telemetry and metrics from multiple clusters.
    As a cluster administrator I want to make sure ingress connections from this service can't be denied by the namespace owners 
    with NetworkPolicy.

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
