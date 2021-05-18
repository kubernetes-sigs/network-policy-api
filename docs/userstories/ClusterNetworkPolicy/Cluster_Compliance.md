### Summary

As a cluster admin, I want all workloads to start with a network/security
model that meets the needs of my company.

### Description

A platform admin may want to factor out policies that each namespace would have
to write individually in order to make deployment and auditability easier.
Common examples include allowing all workloads to be able to talk to the cluster
DNS service and, similarly, allowing all workloads to talk to the logging/monitoring
pods running on the cluster.
