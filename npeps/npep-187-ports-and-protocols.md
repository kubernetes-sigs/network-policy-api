# NPEP-187: More protocols support

* Issue: [#187](https://github.com/kubernetes-sigs/network-policy-api/issues/187)
* Status: Provisional

## TLDR

Change existing `ports` selector to allow more protocols in the future.

## Goals

Change existing `ports` selector to allow more protocols in the future.

## Non-Goals

Adding new protocols to the API.

## Introduction

Currently only TCP, UDP, and SCTP protocols are supported, but there are more protocols (ICMP, ICMPv6 are the most popular requests) 
that may be useful.

## User-Stories/Use-Cases

Story 1: Protocols extension

As a cluster admin I want to be able to use protocols other than TCP, UDP, and SCTP in my ANP.
For example, I want to only allow ICMP connections to implement health monitoring and deny everything else.

## API

As we only support port-based protocols (TCP, UDP, and SCTP), current protocol matching is expressed as
```yaml
ports:
  - portNumber:
      protocol: TCP
      port: 80
  - portRange:
      protocol: UDP
      start: 1
      end: 100
```
To enable more protocols support in the future, we need to change the high-level field name from `port` to `protocol` to be more generic.

In the `protocols` filter design we decided to start from the YAML that we want (i.e. YAML that is easy to write and understand).
We want to make popular filters (TCP and UDN ports) easy and less popular filters (like ICMP and named ports) possible.
This brings us to the following design:

```yaml
protocols:
  - protocol: TCP
    port: 8080
  - protocol: UDP
    portRange: 
      start: 8080
      end: 9090 
  - namedPort: http
    # protocol: "" # must be empty string: defaulted / omitempty
```

with the validation like so:
- empty `protocols` list is not allowed
- at least 1 field for each `protocol` element must be set
- if `namedPort` is set, `protocol` must be unset
- if `protocol` is TCP/UDP/SCTP, `port: int` or `portRange: {start: int, end: int}` must be set, but not at the same time

Using ICMP as a potential example for a pure protocol match (`protocol: ICMP`) or protocol with extra fields (e.g. ICMP type and code),
a future extension could look like this:

```yaml
protocols:
  # option 1:  A little inconsistent with port match as it has an extra `icmp` field instead of flat parameters
  - protocol: ICMP
    icmp:
      type: 7
      code: 3
  # option 2: Flat parameters, may conflict with the future protocol that also have type or code
  - protocol: ICMP
    type: 7
    code: 3
  # option 3: Same as option 1, but without a discriminator (protocol)
  - icmp:
      type: 7
      code: 3
```

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
