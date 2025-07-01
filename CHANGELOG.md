# Changelog

## Table of Contents

- [v0.1.0](#v010)
- [v0.1.1](#v011)
- [v0.1.7](#v017)

# v0.1.0

API Version: v1alpha1

This is the initial release of the network-policy-api. It includes two
main resources geared towards cluster admins:

- AdminNetworkPolicy
- BaselineAdminNetworkPolicy

Please check out the [network-policy-api website](https://network-policy-api.sigs.k8s.io/) for more information.

# v0.1.1

API Version: v1alpha1

This is a patch release of the network-policy-api. It includes two
main resources geared towards cluster admins:

- AdminNetworkPolicy
- BaselineAdminNetworkPolicy

Additionally it includes many conformance test updates and fixes:

- Ingress/Egress Traffic conformance for TCP/UDP/SCTP
- Movement of base testing yamls
- Variable renaming and comment improvements
- Increased default timeout
- Removal of K8s.io/kubernetes dependency

Please check out the [network-policy-api website](https://network-policy-api.sigs.k8s.io/) for more information.

# v0.1.7

API Version: v1alpha1

This is a patch release of the network-policy-api. It includes two
main resources geared towards cluster admins:

- AdminNetworkPolicy
- BaselineAdminNetworkPolicy

The new aspects of the API being released here since v0.1.1 that are worth highlighting include:

- A new type of egress peer `networks` is supported to be able to express CIDR ranges as peers
- An experimental egress peer `nodes` is supported to be able to express Kubernetes nodes as peers
- An experimental egress peer `domainNames` is supported to be able to express FQDNs as peers
- Docs text change around calling out that host-networked pods are not selected as part of subject or peers
- More conformance tests specially for the new fields

Another noteworthy change is the removal of `sameLabels` and `notSameLabels` fields from the API.
Originally these fields were added to be able to express a form of tenancy that was relative to
the selected subject of the policy. Given the selection based on sameness and not-sameness of labels
could compound to many possible ways of expressing relations that would exceed cardinality, the
community is working on a better API proposal for tenancy. See [NPEP-122](https://github.com/kubernetes-sigs/network-policy-api/pull/178)
for more details.

Please check out the [network-policy-api website](https://network-policy-api.sigs.k8s.io/) for more information.