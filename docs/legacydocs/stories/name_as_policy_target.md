# Name as Policy Target

## Owners

- Abhishek Raut

## User Story

As a network policy creator, I want to target source and destination endpoints in
network policy rules by referring to named resources rather than resource
labels.

*List specific resources that could be targeted

## v1 Incremental Addition (MVP)

Start with Namespace name as predicate for namespace selector.

We can add this either in ClusterNetworkPolicy only to reduce friction or it can be added to NetworkPolicy and ClusterNetworkPolicy.
