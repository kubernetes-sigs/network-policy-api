
### Summary

1. **As a** cluster administrator **I want** to restrict traffic from
specific cluster workloads to all or specific destinations outside the
cluster **so that** I can enforce security for northbound traffic.

2. **As a** cluster administrator **I want** to explicitly allow traffic
from cluster workloads to handpicked destinations outside the cluster and
deny the rest **so that** I can guarantee cluster wide egress traffic
flow irrespective of whatever network policies are defined by the
developer for specific namespaces.

### Description

Often times cluster administrators want to explicitly control the
egress traffic from workloads towards external destinations like allowing
some (not all) workloads to access a secure service outside the cluster.

They might also want to setup cluster wide safeguard rules for egress in
general so that they don't have to worry about when namespace specific
policies are created by users at any given point.

### Acceptance Criteria

1. If I add an admin network policy to deny traffic from a specific "subject"
of workloads towards an external destination (optionally provide port
restrictions) then those workloads should not be able to connect to that
external destination.

2. If I apply an admin network policy with lower precedence egress rules to deny
traffic from a specific "subject" of workloads towards all external
destinations and then apply higher precedence egress rules to allow traffic for
those workloads towards well defined destinations then those workloads should
not be able to connect to any other external destinations than the explicitly
allowed ones.

[
Also, here are a few points that need to be addressed:

1. Addressing ingress use cases for north-south traffic is out of scope.
If we have requests from users to restrict access INTO the cluster
workloads from external entities we can consider a new `ingress-control-to-cluster`
case in the future. That will be a separate effort altogether.
2. I propose to include this in our existing ANP & BANP objects and not create
a new object.
]

### Resources:

* https://github.com/kubernetes-sigs/network-policy-api/issues/28 

### Notes

* This is listed as one of the beta graduation criterias: https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/2091-admin-network-policy#alpha-to-beta-graduation. Let's try to get this in our alpha versions itself.
