# NPEP-137: Conformance Profiles

* Issue: [#137](https://github.com/kubernetes-sigs/network-policy-api/issues/137)
* Status: Implementable

## TLDR

Add conformance profiles for existing conformance tests which implementations
can leverage when running the conformance test suite in their downstream repos
to report the results of the tests back to the Network Policy API project and
receive recognition (eg. "conformance badge")

* NOTE: Adapted with love from https://gateway-api.sigs.k8s.io/geps/gep-1709/ (Then
a question may arise, why do we need our own NPEP? Answer: Gateway API Project has
lots of resources and features and it gets very complicated to adopt their same API.
We need to tune the profiles to be compatible with what works for Network Policy API
project, thus the need for this NPEP). Our profiles are much simpler than what GatewayAPI
supports.

## Goals

* Add conformance profiles for existing tests that downstream implementations
can subscribe to to run tests associated to the supported API feature sets

* Add a reporting mechanism via a new CRD where conformance results can be
reported back to the Network Policy API project and provide "conformance badges"
via Network Policy official website to recognize these implementations.

* Expand existing conformance testing framework to account for profiles

## Non-Goals

* The first iteration will just have a simple repo under which the report
results will be stored, no automated reporting infrastructure will be built for this now.
Implementations will need to open PRs against network-policy-api repo to upload
the results.

## Introduction

Currently, the conformance tests are grouped into `CoreFeatures` and
`ExtendedFeatures`. The support for features that fall under the `CoreFeatures`
are a requirement for conformant implementations, while the support for features that fall
under `ExtendedFeatures` are optional and do not gate API conformance for implementations.

In this NPEP, we will add a concept called `named profiles` which will indicate a
"level of conformance" for a given resource. We are picking profiles on a resource
level because that is what makes sense for the project as of today. The initial decision
to report conformance profiles on a per-resource level is a choice which may be re-evaluated
by the NetworkPolicy API working group in the future. Using these `named profiles`,
implementations can run the conformance tests, prove they satisfy the requirements for each profile,
and receive conformance badges from the Network Policy API project.

This NPEP aims to provide an API (CRD) for reporting results from these conformance
tests. CLI tooling will also be provided for invoking the conformance tests with
the profile names so that the results can be easily submitted back upstream.

## User-Stories/Use-Cases

Story 1: Receive an API conformance badge

As a CNI maintainer, I want to receive recognition from Network Policy API subgroup
that my plugin implements the `AdminNetworkPolicy` API. This badge will enable us
to advertise upstream feature support in our plugin easily to our users.

Story 2: Track Implementations for each API resource

As a Network Policy project maintainer, I would like to easily track which
implementations are using the defined API resources in our project so that it
paves way for easier collaboration and feedback loops when changes are introduced
to the API in each release.

## API

We will add a new API resource called `ConformanceReport` which will be at
the center of our test result reporting workflow. The implementors running
the tests can then:

1. Choose a `named profile`
2. Integrate the tests required by a given profile in their downstream project
3. Report the results to Network Policy API project using this new API resource
4. Get a conformance badge recognition via our official website

## Profiles

Named Profiles for Network Policy API project will be tied to each API resource.
We will start with two named profiles and expand this as the project evolves.

1. AdminNetworkPolicy
2. BaselineAdminNetworkPolicy

Each of these profiles may have a combination of conformance tests that fall under
`CoreFeatures` and `ExtendedFeatures`. Example; if you pick the profile
`AdminNetworkPolicy`, all tests like `AdminNetworkPolicyEgressSCTP` and
`AdminNetworkPolicyPriorityField` fall under the `SupportAdminNetworkPolicy` feature
which is under the `CoreFeatures` subset. So these tests must pass for the
`AdminNetworkPolicy` profile conformance. Whereas tests (we don't have any today)
that fall under `AdminNetworkPolicySameLabels` feature which is under the
`ExtendedFeatures` subset are not mandatory for `AdminNetworkPolicy`
profile conformance.

## Integration

The conformance profile test suite can be integrated, invoked and run from your implementation
using two methods:

* The `go test` CLI commands specifying the required information need to generate
a conformance test report. Sample:
```
go test  -v ./conformance -run TestConformanceProfiles -args --conformance-profiles=AdminNetworkPolicy,BaselineAdminNetworkPolicy --organization=ovn-org -project=ovn-kubernetes -url=<project-url> -version=0.1.1 -contact=<> -additionalinfo=<link-to-implementation>
```
* Using the conformance profile test suite `TestConformanceProfiles` by directly customizing it by providing
the correct arguments. Sample:
```
cpSuite := suite.NewConformanceProfileTestSuite(
            suite.ConformanceProfileOptions{
                    suite.Options: cSuiteDefaultOptions,
                    Implementation: confv1alpha1.Implementation{
                            Organization:          "ovn-org",
                            Project:               "OVN-Kubernetes CNI",
                            URL:                   "https://github.com/ovn-org/ovn-kubernetes",
                            Version:               "0.1.1",
                            Contact:               []string{"@tssurya"},
                            AdditionalInformation: "https://github.com/ovn-org/ovn-kubernetes/blob/1c9f73dc8a755c07b22858c7404a7884970d1989/test/conformance/network_policy_v2_test.go"
                    },
                    ConformanceProfiles: sets.New(
                            suite.ANPConformanceProfileName,
                            suite.BANPConformanceProfileName,
                    ),
            })
```

## Reporting process and certification

The reporting process is related to a specific API's version and channel (core and experimental).
There are fields in the ConformanceReport CRD that includes such information. Any implementation
can run the existing conformance test suite specifying the profiles they support and that will
generate an output that looks like this:

```
    Conformance report:
        apiVersion: policy.networking.k8s.io/v1alpha1
        date: "2023-10-03T08:15:25+02:00"
        implementation:
          contact:
          - "@tssurya"
          organization: ovn-org
          project: OVN-Kubernetes CNI
          url: "https://github.com/ovn-org/ovn-kubernetes"
          version: 0.1.1
          additionalInformation: "https://github.com/ovn-org/ovn-kubernetes/blob/1c9f73dc8a755c07b22858c7404a7884970d1989/test/conformance/network_policy_v2_test.go"
        kind: ConformanceReport
        networkPolicyV2APIVersion: v0.1.1
        profiles:
        - core:
            failedTests:
            - AdminNetworkPolicyIngressSCTP
            - AdminNetworkPolicyEgressUDP
            result: failure
            statistics:
              Failed: 2
              Passed: 5
              Skipped: 0
            summary: ""
          name: AdminNetworkPolicy
        - core:
            result: success
            statistics:
              Failed: 0
              Passed: 7
              Skipped: 0
            summary: ""
          name: BaselineAdminNetworkPolicy
```

This can then be uploaded to network-policy-api/conformance/reports/v.x.x/cni-name.yaml by
opening a PR. That will then be reviewed and approved by maintainers thus recognizing the
implementations that are conformant. It is recommended for the implementations to use the
`additionalInformation` field to provide links to the implementation or github actions or
jenkins or any other CI/CD job definitions that helped generate this report. This will
help maintainers make an informed decision on merging the report PR.

## Alternatives

N/A - We just went with what Gateway API project already has implemented without
having to reinvent the wheel.

## References

1. https://github.com/kubernetes-sigs/network-policy-api/pull/142
