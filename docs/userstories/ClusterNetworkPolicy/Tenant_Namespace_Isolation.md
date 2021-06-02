### Summary

As a cluster admin, I want to isolate all the tenants (modeled as Namespaces)
in my cluster from each other. This means that all across-Namespace traffic is
denied if each tenant is assigned a single Namespace.

### Description

Many enterprises are creating shared Kubernetes clusters that are managed by a
centralized platform team. Each internal team that wants to run their workloads
gets assigned a Namespace on the shared clusters. Naturally, the platform team
will want to make sure that, inter-namespace traffic between these tenants is
denied, so that the tenant won't interfere with each other in unwanted way.

This cluster security posture could be achieved by applying a deny inter-Namespace
traffic policy to all the tenant Namespaces, with certain exceptions. Exceptions
may include: don't deny traffic from tenant Namespaces to kube-system Namespace,
or don't drop communication between tenant 1 and tenant 2 as requested by both
tenants.

### Acceptance Criteria

If I apply a tenant isolation policy to certain Namespaces in the cluster, these
Namespace should deny inter-namespace traffic, unless listed as exempt or
explicited allowed by higher-order policy rules.

### Notes

Comments in the [ClusterNetworkPolicy KEP](https://github.com/kubernetes/enhancements/pull/2522#discussion_r632866263)
and disucssion in the sig-network group seems to suggest that tenants are often
modeled as a set of Namespaces in addition to single Namespace (two Namespaces
with label app=coke below to a single tenant and three Namespaces with label
app=pepsi below to another tenant). In this case, tenant isolation would imply
traffic is only denied to/from Namespaces that has different label value as self,
for some label key used for tenant identification.

Another [comment thread](https://github.com/kubernetes/enhancements/pull/2522/files#r604694072)
from the same KEP explores whether the Namespace isolation should be enforced
(i.e. not overridable by K8s NetworkPolicy). Most reviewers agree that 1) This
tenant boundary policy should not be overriden by tenants; they can only be
amended by cluster admins 2) Tenant Namespace isolation should not imply that
intra-Namespace traffic is allowed; tenants should be able to segment their own
Namespaces as they wish via K8s NetworkPolicies.
