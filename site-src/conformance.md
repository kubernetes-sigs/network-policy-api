Adapted with :blue_heart: from the [gateway api project release documentation](https://gateway-api.sigs.k8s.io/)

# Conformance

This API covers a broad set of features and use cases and has been implemented
widely. This combination of both a large feature set and variety of
implementations requires clear conformance definitions and tests to ensure the
API provides a consistent experience wherever it is used.

When considering Network Policy API conformance, there are three important concepts:

## 1. Release Channels

Within Network Policy API, release channels are used to indicate the stability of a
field or resource. The "standard" channel of the API includes fields and
resources that have graduated to "beta". The "experimental" channel of the API
includes everything in the "standard" channel, along with experimental fields
and resources that may still be changed in breaking ways **or removed
altogether**. For more information on this concept, refer to our
[versioning](./versioning.md) documentation.

## 2. Support Levels

Unfortunately some implementations of the API will not be able to support every
feature that has been defined. To address that, the API defines a corresponding
support level for each feature:

* **Core** features will be portable and we expect that there is a reasonable
  roadmap for ALL implementations towards support of APIs in this category.
* **Extended** features are those that are portable but not universally
  supported across implementations. Those implementations that support the
  feature will have the same behavior and semantics. It is expected that some
  number of roadmap features will eventually migrate into the Core. Extended
  features will be part of the API types and schema.
* **Implementation-specific** features are those that are not portable and are
  vendor-specific. Implementation-specific features will not have API types and
  schema except via generic extension points.

Behavior and feature in the Core and Extended set will be defined and validated
via behavior-driven conformance tests. Implementation-specific features will not
be covered by conformance tests.

By including and standardizing Extended features in the API spec, we expect to
be able to converge on portable subsets of the API among implementations without
compromising overall API support. Lack of universal support will not be a
blocker towards developing portable feature sets. Standardizing on spec will
make it easier to eventually graduate to Core when support is widespread.

### Overlapping Support Levels

It is possible for support levels to overlap for a specific field. When this
occurs, the minimum expressed support level should be interpreted. For example,
an identical struct may be embedded in two different places. In one of those
places, the struct is considered to have Core support while the other place only
includes Extended support. Fields within this struct may express separate Core
and Extended support levels, but those levels must not be interpreted as
exceeding the support level of the parent struct they are embedded in.

## 3. Conformance Tests

Network Policy API includes a set of conformance tests. These create a series of
`ClusterNetworkPolicy` resources in a cluster — in both the admin and baseline
tiers — and test that the implementation enforces them according to the API
specification.

Each release contains a set of conformance tests, these will continue to
expand as the API evolves. Currently conformance tests cover the majority
of Core capabilities in the standard channel, in addition to some Extended
capabilities.

### Running Tests

Conformance tests run against a live Kubernetes cluster with an implementation
of the Network Policy API deployed in it. To run them you will need:

* A **multi-node** Kubernetes cluster. Some tests exercise node-level and
  cross-node traffic; the upstream CI uses a [kind][kind] cluster with one
  control-plane node and two worker nodes.
* The `ClusterNetworkPolicy` CRD installed from the release channel you want
  to test:

    ```shell
    # standard channel
    kubectl apply -f config/crd/standard/policy.networking.k8s.io_clusternetworkpolicies.yaml

    # OR experimental channel (adds experimental fields such as named ports
    # and node peers)
    kubectl apply -f config/crd/experimental/policy.networking.k8s.io_clusternetworkpolicies.yaml
    ```

* An implementation running in the cluster that enforces
  `ClusterNetworkPolicy`, for example [kube-network-policies][knp] — the
  reference implementation used by this project's CI.
* A kubeconfig for the cluster: the suite runs against the active context of
  your `KUBECONFIG` (or `~/.kube/config`).

The standard conformance suite can then be run from the root of this
repository with:

```shell
make conformance
```

Suite flags are passed via the `CONFORMANCE_FLAGS` Makefile variable, or after
`-args` when invoking `go test` directly — the following two commands are
equivalent:

```shell
make conformance CONFORMANCE_FLAGS="--all-features"
go test -v ./conformance -run '^TestConformance$' -timeout 20m -args --all-features
```

The tests create their own client and server workloads in a set of
`network-policy-conformance-*` namespaces and remove them when the run
completes.

#### Running Specific Tests

The suite registers each conformance test as a subtest named after its
`ShortName`, so an individual test can be selected with a regular `go test`
run filter:

```shell
go test -v ./conformance -run '^TestConformance$/CNPAdminTierIngressTCP'
```

The available short names are listed in the test definitions under
[`conformance/tests`][tests-dir].

#### Selecting Features

Every test declares which features it exercises, and only tests whose
features are all enabled will run:

| Channel      | Feature                               |
|--------------|---------------------------------------|
| Standard     | `ClusterNetworkPolicy`                |
| Experimental | `ClusterNetworkPolicyNamedPorts`      |
| Experimental | `ClusterNetworkPolicyEgressNodePeers` |

Standard features are always enabled. Experimental features are opt-in —
enable specific ones with `--supported-features`, or everything at once with
`--all-features` (remember that experimental features also require the
experimental channel CRD):

```shell
make conformance CONFORMANCE_FLAGS="--supported-features=ClusterNetworkPolicyNamedPorts"
make conformance CONFORMANCE_FLAGS="--all-features"
```

Features can also be excluded with `--exempt-features`, e.g. to run against
an experimental-channel cluster while skipping a feature the implementation
does not support:

```shell
make conformance CONFORMANCE_FLAGS="--all-features --exempt-features=ClusterNetworkPolicyEgressNodePeers"
```

The authoritative feature list lives in [`features.go`][features].

#### Suite Level Options

* `--cleanup-base-resources=false` leaves the test workloads and namespaces
  in place after the run, which is useful for inspecting cluster state after
  a failure.
* `--debug` enables debug logging.

The full set of flags is defined in [`flags.go`][cflags].

#### Trying It Out Locally

The project's conformance CI job is a complete working recipe for a local
run: it creates a multi-node kind cluster, installs the experimental-channel
CRD, deploys kube-network-policies as the enforcing implementation, and runs
the suite with a full set of report flags. See the
[conformance workflow][ci-workflow] — the same steps work on a developer
machine with `kind`, `kubectl`, and Go installed.

### Conformance Profiles

Conformance profiles bundle the features that make up a certifiable unit of
the API, so that implementations can run them together and generate a
[conformance report](#conformance-reports) certifying their support. The
design is described in [NPEP-137](./npeps/npep-137-conformance-profiles.md).

One profile is currently defined, covering the `ClusterNetworkPolicy` API:

| Profile                | Standard features      | Experimental features                                                   |
|------------------------|------------------------|-------------------------------------------------------------------------|
| `ClusterNetworkPolicy` | `ClusterNetworkPolicy` | `ClusterNetworkPolicyNamedPorts`, `ClusterNetworkPolicyEgressNodePeers` |

Profile runs use the `TestConformanceProfiles` entrypoint (or the
`make conformance-profiles` target, which forwards `CONFORMANCE_FLAGS` the
same way) together with the `--conformance-profiles` flag. Because their
purpose is to produce a conformance report, they also require the
implementation metadata flags described in the next section. Profile
definitions live in [`conformance_profiles.go`][profiles].

## Conformance Reports

Conformance reports are how implementations demonstrate — rather than just
claim — their support for the API. A report is generated by the test suite
itself and records the profile that was run, per-channel
passed/failed/skipped statistics with the names of any failed or skipped
tests, and exactly which experimental features the implementation declared
support for. Together with a link to the CI job that produced it, this gives
users verifiable, reproducible evidence of feature support that a
self-maintained feature list cannot provide.

If your implementation is listed on the
[implementations page](./implementations.md), please substantiate the entry
by submitting a conformance report as described below.

To generate a report, run the profiles suite with your implementation's
details (this mirrors the invocation this project's CI uses to test
[kube-network-policies][knp] on every pull request — see the
[conformance workflow][ci-workflow] for a complete working example):

```shell
go test -v ./conformance -run TestConformanceProfiles -timeout 20m -args \
    --conformance-profiles=ClusterNetworkPolicy \
    --organization=my-org \
    --project=my-network-policy-implementation \
    --url=https://github.com/my-org/my-project \
    --version=v1.2.3 \
    --contact=@my-org/maintainers \
    --additional-info=https://my-org.example/ci-run-that-generated-this-report \
    --all-features \
    --report-output=my-report.yaml
```

The implementation metadata flags are all required. In particular,
`--additional-info` must link to the CI integration that shows how the report
was generated; maintainers use it to verify a report's provenance before
accepting it. The report is printed at the end of the test log and, with
`--report-output`, written to the given file.

If the implementation does not support every experimental feature, replace
`--all-features` with an explicit `--supported-features` selection: features
that are not enabled are recorded in the report as `unsupportedFeatures` and
their tests are skipped, which is expected. Skipped or failing **standard**
tests, on the other hand, mean the implementation is not conformant, and the
report will say so.

To submit a report, open a pull request adding the generated YAML under
[`conformance/reports`][reports-dir] as
`conformance/reports/<API release version>/<implementation>.yaml` (see
[NPEP-137](./npeps/npep-137-conformance-profiles.md)). Merged reports are the
project's official record of which implementations conform, and at what
feature granularity.

## Contributing to Conformance

Many implementations run conformance tests as part of their full e2e test
suites. Contributing conformance tests means that implementations can share
the investment in test development and ensure that we're providing a
consistent experience.

All code related to conformance lives in the [`conformance`][conformance-dir]
directory of the project. Each test pairs a Go file in `conformance/tests/`
with a YAML manifest of `ClusterNetworkPolicy` resources under
`conformance/base/`: the Go file declares a `suite.ConformanceTest` with a
unique `ShortName`, the `Features` it depends on, and its `Manifests`, and
appends itself to the shared test list from an `init()` function. The base
workloads that the policies select over (client and server pods spread across
the `network-policy-conformance-*` namespaces) are defined once for the whole
suite in `conformance/base/manifests.yaml`.

Issues related to conformance are
[labeled with "area/conformance"][conformance-issues]. These often cover
adding new tests to improve our coverage or fixing flaws or limitations in
the existing tests.

[kind]: https://kind.sigs.k8s.io/
[knp]: https://github.com/kubernetes-sigs/kube-network-policies
[tests-dir]: https://github.com/kubernetes-sigs/network-policy-api/tree/main/conformance/tests
[features]: https://github.com/kubernetes-sigs/network-policy-api/blob/main/conformance/utils/suite/features.go
[cflags]: https://github.com/kubernetes-sigs/network-policy-api/blob/main/conformance/utils/flags/flag.go
[profiles]: https://github.com/kubernetes-sigs/network-policy-api/blob/main/conformance/utils/suite/conformance_profiles.go
[ci-workflow]: https://github.com/kubernetes-sigs/network-policy-api/blob/main/.github/workflows/conformance.yml
[reports-dir]: https://github.com/kubernetes-sigs/network-policy-api/tree/main/conformance/reports
[conformance-dir]: https://github.com/kubernetes-sigs/network-policy-api/tree/main/conformance
[conformance-issues]: https://github.com/kubernetes-sigs/network-policy-api/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fconformance
