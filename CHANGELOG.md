# Changelog

## Table of Contents

- [v0.1.0](#v010)
- [v0.1.1](#v011)
- [v0.1.7](#v017)
- [v0.2.0](#v020)
- [v0.2.1](#v021)

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

# v0.2.0

API Version: v1alpha2

This is a major release of the network-policy-api that introduces v1alpha2 with
a single unified resource for cluster administrators:

- ClusterNetworkPolicy (CNP)

ClusterNetworkPolicy combines the previously separate AdminNetworkPolicy and
BaselineAdminNetworkPolicy resources into a single CRD with a `tier` field
(Admin or Baseline) to control policy precedence.
See [NPEP-285](https://network-policy-api.sigs.k8s.io/npeps/npep-285-combine-crds)
for more details.

The key changes since v0.1.7 include:

- **NPEP-285: Combine ANP and BANP into ClusterNetworkPolicy** - A single CRD
  with `tier` field replaces the two separate resources. The `Allow` action has
  been renamed to `Accept`.
- **NPEP-187: Ports and Protocols** - The `ports` field has been replaced with
  a new `protocols` field that provides explicit per-protocol matching for TCP,
  UDP, and SCTP with clearer semantics.
- Updated conformance tests for the new v1alpha2 API
- Reduced MaxItems from 100 to 25 for rules and peers
- Updated documentation and user stories for ClusterNetworkPolicy

Please check out the [network-policy-api website](https://network-policy-api.sigs.k8s.io/) for more information.

# v0.2.1

API Version: v1alpha2

This is a patch release of the network-policy-api on top of v0.2.0. It continues
to ship the single unified resource for cluster administrators:

- ClusterNetworkPolicy (CNP)

There are no breaking changes to the shape of the API in this release. The
changes are focused on API validation fixes, channel placement corrections, and
a large set of conformance test and documentation improvements.

The key changes since v0.2.0 include:

- **API validation fixes:**
  - CIDR fields must now be in canonical (masked) form, so a value with host
    bits set such as `10.1.2.3/8` is rejected (use `10.0.0.0/8`).
  - Fixed the `DomainName` validation pattern to reject invalid characters.
- **NamedPort is experimental again** - The `namedPort`/`destinationNamedPort`
  fields have been moved back to the experimental channel while the design is
  finalized, so they are no longer present in the standard channel CRDs.
- **Conformance tests:**
  - New coverage: CRD CEL validation, CNP pod and namespace selector changes,
    ingress/egress isolation and same-priority CNPs across the Admin and Baseline
    tiers, and PortRange.
  - Replaced the partial result status with optional tests.
  - The report's bundle version now comes from the installed CRDs'
    `bundle-version` annotation and must match the suite's; the
    `networkPolicyV2APIVersion` field was renamed to `networkPolicyBundleVersion`.
- **Documentation:**
  - Added a conformance reports README and submission process, instructions for
    running the suite, and the kube-network-policies v0.2.0 report.
  - Clarified hairpinning (service hairpin enforcement preferred for beta), ICMP
    handling, and the `Accept`/`Deny`/`Pass` action descriptions.

Please check out the [network-policy-api website](https://network-policy-api.sigs.k8s.io/) for more information.