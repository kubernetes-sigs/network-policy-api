# Changelog

## Table of Contents

- [v0.1.0](#v010)
- [v0.1.1](#v011)

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