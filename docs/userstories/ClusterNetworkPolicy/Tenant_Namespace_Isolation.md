### Summary

As a cluster admin, I want to isolate all the tenants (modeled as Namespaces)
on my cluster from each other by default.

### Description

Many enterprises are creating shared Kubernetes clusters that are managed by a
centralized platform team. Each internal team that wants to run their workloads
gets assigned a Namespace on the shared clusters. Naturally, the platform team
will want to make sure that, by default, all intra-namespace traffic is allowed
and all inter-namespace traffic is denied.

### Acceptance Criteria

If I apply a tenant isolation policy to certain Namespaces in the cluster, these
Namespace should only allow intra-namespace traffic, unless there are higher-order
policy rules applied.

### Notes

Comments in the [ClusterNetworkPolicy KEP](https://github.com/kubernetes/enhancements/pull/2522#discussion_r632866263)
and disucssion in the sig-network group seems to suggest that tenants are often
modeled as a set of Namespaces in addition to single Namespace (two Namespaces
with label app=coke below to a single tenant and three Namespaces with label
app=pepsi below to another tenant). In this case, tenant isolation would imply
traffic is only allowed to/from Namespaces that has the same label value as self,
for some label key used for tenant identification.
