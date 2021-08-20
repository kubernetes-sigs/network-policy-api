# Ingress Network Policy Rules Targeting Services

## User Story

I as network policy creator, I want ingress rules to target destination Services
by name or label.

Egress rules should not target Services because a pod could implement multiple
services. In this case, there could be a conflict between the egress rules for
each service the pod implements.

Related: [Name As Policy Target](./named_endpoint_set.md)

