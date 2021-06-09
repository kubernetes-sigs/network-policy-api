# Node Selector

## Owners

- @jayunit100 (Jay Vyas)

## User Story

As a policy creator, 

- I need to limit ingress from and egress to nodes (and specifically, their Kubernetes recognized network interfaces), so that 'special' nodes which may scale up and down dynamically over time, can be protected from sending or recieveing traffic.
- Since these nodes can increase/decrease, targetting them via metadata rather then IP addresses is required, and the obvious metadata to use would be `labels`.

## Notes

- This may involve changing the semantics around health checks, which are always ALLOW policies (nodes sending traffic to pods is always allowed).
- This may be considerd a Cluster scoped policy, since it can effect many different target namespaces, and since it is administrative in nature (as opposed to application centered).
- This was originally considered as a proposal for the broader SIG as a first pass improvement to v1 policies, but then we realized it was quite complicated in that it might break semantics of ALLOW for node -> pods, and that it also might need to be collaboratively built alongside the broader cluster network policy.


