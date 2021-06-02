### Summary

As a cluster admin, I want to ensure that all traffic coming into (going out of)
my cluster always goes through my ingress (egress) gateway.

### Description

It is common practice in enterprises to setup checkpoints in their clusters at
ingress/egress. These checkpoints usually perform advanced checks such as
firewalling, authentication, packet/connection logging, etc.

This is a big request for compliance reasons, and ClusterNetworkPolicy can ensure
that all the traffic is forced to go through ingress/egress gateways.
It is worth noting that the Cluster-scoped NetworkPolicy APIs will not redirect
traffic, rather it can ensure that no traffic is allowed in/out except traffic
via the gateways.

### Acceptance Criteria

If I apply a policy to the cluster for ingress/egress gateway enforcement, there
can be no ingress traffic allowed from outside of the cluster except for the ingress
gateway, and no egress traffic to outside of the cluster except for the egress
gateway.
