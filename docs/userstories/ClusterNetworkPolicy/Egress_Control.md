### Summary

As a cluster admin, I want to explicitly limit which workloads can connect to
well known destinations outside the cluster.

### Description

This user story is particularly relevant in hybrid environments where customers
have highly restricted databases running behind static IPs in their networks
and want to ensure that only a given set of workloads is allowed to connect to
the database for PII/privacy reasons.

### Acceptance Criteria

If I apply a policy that limit egress to certain addresses, it should be able to
guarantee that only the selected Pods can connect to those IPs.
