### Summary

As a cluster admin, I want to explicitly deny traffic from certain source IPs
that I know to be bad.

### Description

Many admins maintain lists of IPs that are known to be bad actors, especially
to curb DoS attacks. A cluster admin could use ClusterNetworkPolicy to codify all
the source IPs that should be denied in order to prevent that traffic from
accidentally reaching workloads.
Note that the inverse of this (allow traffic from well known source IPs) is also
a valid use case.

### Acceptance Criteria

If I apply a deny specific IPs rule to a set of Pods, ingress traffic from those
IPs (as seen by the Pod) should be dropped.

### Notes

An active discussion on the [ClusterNetworkPolicy KEP](https://github.com/kubernetes/enhancements/pull/2522#discussion_r601729952)
is exploring whether the source IP should/can consider the various ingress
mechanism which may obscure the origin of the packet due to NATing.
