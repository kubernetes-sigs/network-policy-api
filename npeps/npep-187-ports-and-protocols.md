# NPEP-187: More protocols support

* Issue: [#187](https://github.com/kubernetes-sigs/network-policy-api/issues/187)
* Status: Provisional

## TLDR

Change existing `ports` selector to allow more protocols in the future.
Option3: has won the design discussion.

```yaml
protocols:
  - tcp:
      destinationPort:
        number: 8080
      flags: [syn] # future extension example
  - tcp:
      destinationPort:
        range: 
          start: 8080
          end: 9090
  - udp:
      destinationPort:
        number: 8080
  - udp:
      destinationPort:
        number: 9090
  - namedPort: http
  - namedPort: monitoring
  - icmp: # that doesn't exist yet, but may be added
      type: 7
      code: 3
```

## Goals

Change existing `ports` selector to allow more protocols in the future.

## Non-Goals

Adding new protocols to the API.

## Introduction

Currently only TCP, UDP, and SCTP protocols are supported, but there are more protocols (ICMP, ICMPv6 are the most popular requests) 
that may be useful in the future, so we want to leave space for future extensions.

## User-Stories/Use-Cases

Story 1: Protocols extension

As a cluster admin I want to be able to use protocols other than TCP, UDP, and SCTP in my CNP.
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

We have discussed an option to allow protocol-only match, like TCP any port, but we never had a good user story for it to
justify having special syntax for it. 
For now, it can be expressed as a port range that covers all ports (1-65535).

In the `protocols` filter design we decided to start from the YAML that we want (i.e. YAML that is easy to write and understand).
We want to make popular filters (TCP and UDN ports) **easy** and less popular filters (like ICMP and named ports) **possible**.
There are multiple options to support this:

### Option 1: Protocols list with protocol discriminator

```yaml
protocols:
  - protocol: TCP
    port: 
      number: 8080
  - protocol: UDP
    port:
      range: 
        start: 8080
        end: 9090 
  - namedPort: http
    # protocol: "" # must be empty string: defaulted / omitempty
```

with the validation like so:
- empty `protocols` list is not allowed
- at least 1 field for each `protocol` element must be set
- if `namedPort` is set, `protocol` must be unset
- if `protocol` is TCP/UDP/SCTP, `port` must be set
- exactly one of `number` or `range` must be set in `port`

### Option2: Join by protocol

```yaml
protocols:
  tcp:
    - port:
        number: 8080
    - port:
        range: 
          start: 8080
          end: 9090
  udp:
    - port:
        number: 8080
    - port:
        number: 9090
  namedPort:
    - http
    - monitoring
  icmp:
    - type: 7
      code: 3
```

with the validation like so:
- `protocols` must have at least 1 protocol set
- each protocol list must not be empty
- each protocol element must have at least 1 field set

### Option3: Support older clients

Supporting older clients that may not know about the newer API fields is an important consideration for APIs where
evolution is highly likely. The idea is to make it possible for older clients to understand/report that a
given CNP has fields they don't understand and "fail closed".

Suppose one day we add GRE support. The API understands an object with gre specified but we have an 
older client which does not. The client will deserialize the gre field and discard it. 
One approach to defending that is to elevate the list, like:

```go
type Protocols struct {
// exactly one should be set, if you see this empty it's something you don't know how to handle
    TCP *TCPProtocol
    UDP *UDPProtocol
    GRE *GREProtocol
}
```

Then rather than the parent having a single `Protocols Protocols` it would have a list: `Protocols []Protocols`. Rendering as YAML like:

```yaml
protocols
- tcp:
    port: 53
- udp:
    port: 53
- gre:
    somethingHere: true
```
and an older client would see:
```yaml
protocols
- tcp:
    port: 53
- udp:
    port: 53
- {}  // <- something is obviously wrong!
```

We have an open issue to solve this problem https://github.com/kubernetes-sigs/network-policy-api/issues/340
and think that requiring implementations to automatically detect unsupported API (eg validating that you can 
round-trip the JSON representation through the compiled representation without losing any fields) is much better 
than struggling to design the API in a way that just makes it possible to manually detect that something got dropped.

To see more details about the API design struggle for this requirement, 
<details>
<summary>Click here</summary>

If the old client can't even see that some protocol match is not recognized, it can't implement the `Deny` agreement.
So some traffic that should have been denied may be allowed, which is a security risk.

To support future extensions on the existing protocols match while keeping old clients aware of it
we need to keep a one-of behaviour for all possible match combinations. So if we ever introduce a TCP flag match,
that can be used both with `port` and `portRange`, we will have to express it like
```go
type TCPProtocol struct {
  // exactly one of these must be set 
  Port *int
  PortRange *PortRange
  PortWithFlag *PortWithFlag
  PortRangeWithFlag *PortRangeWithFlag
}
```
or
```go
type TCPProtocol struct {
  // exactly one of these must be set 
  Port *Port // includes both port and port range
  PortWithFlag *PortWithFlag
}
```

Every new field that can be used in combination with the other fields will generate new fields, e.g. if we imagine adding
a timeout field that can be used with all the other options, we will have

```go
type TCPProtocol struct {
  // exactly one of these must be set 
  Port *Port // includes both port and port range
  PortWithFlag *PortWithFlag
  PortWithTimeout *PortWithTimeout
  PortWithFlagWithTimeout *PortWithFlagWithTimeout
}
```
which can age poorly with more options, but it doesn't need any extra validations (except for one-of) and keeps old clients aware of the unknown fields.
</details>

```yaml
protocols:
  - tcp:
      port:
        number: 8080
  - tcp:
      port:
        range: 
          start: 8080
          end: 9090
  - udp:
      port:
        number: 8080
  - udp: 
      port:
        number: 9090
  - namedPort: http
  - namedPort: monitoring
  - icmp:
      type: 7
      code: 3
```

with the validation like so:
- empty `protocols` list is not allowed
- exactly 1 field for each `protocols` element must be set

### Comparison

Option2 is the easiest to write, takes less space and simplifies the validation compared to Option1, but doesn't
support old clients as opposed to Option3.
We have another issue that will solve the older client problem in a more generic way without the need to design the API
with that in mind https://github.com/kubernetes-sigs/network-policy-api/issues/340.

Option2 is slightly weird because it means the fields of the protocols struct have "OR" semantics rather than "AND" semantics.

Options 2 and 3 don't allow defaulting protocol to TCP 
(which is the most popular case). On the other hand, I am not convinced that defaulting protocol to TCP is
a) a good idea (less transparency on which protocol is used)
b) much simpler than adding `tcp:`

To compare, the first option for `tcp:8080` is
```yaml
protocols:
  - port: 
      number: 8080
```
and the second is
```yaml
protocols:
  tcp:
    - port: 
        number: 8080
```
and the third is
```yaml
protocols:
  - tcp:
      port: 
        number: 8080
```

Now we need to decide if taking care of the old clients is important enough to justify the extra complexity, see below some examples.

And if you want to match 3 TCP ports: 80, 8080, 443 (you can imagine how it looks with even more ports)

```yaml
protocols:
  - port: 
      number: 80
  - port:
      number: 8080
  - port:
      number: 443
```
vs
```yaml
protocols:
  tcp:
    - port: 
        number: 80
    - port:
        number: 8080
    - port:
        number: 443
```
vs
```yaml
protocols:
  - tcp:
      port: 
        number: 80
  - tcp:
      port:
        number: 8080
  - tcp:
      port:
        number: 443
```

Allow DNS (TCD and UDP port 53):

```yaml
protocols:
  - port: 
      number: 53
  - protocol: UDP
    port:
      number: 53
```
vs
```yaml
protocols:
  tcp:
    - port: 
        number: 53
  udp:
    - port: 
        number: 53
```
vs
```yaml
protocols:
  tcp:
    - port: 
        number: 53
  udp:
    - port: 
        number: 53
```

Allow a range of UDP ports (e.g. for telephony use cases):

```yaml
protocols:
  - protocol: UDP
    port:
      range:
        start: 10000
        end: 20000
```
vs
```yaml
protocols:
  udp:
    - port:
        range:
          start: 10000
          end: 20000
```
vs
```yaml
protocols:
  - udp:
      port:
        range:
          start: 10000
          end: 20000
```

### New protocols/fields support

Using ICMP as a potential example for a pure protocol match (`protocol: ICMP`) or protocol with extra fields (e.g. ICMP type and code),
a future extension could look like this:

Option 1:
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

Option 2:
```yaml
protocols:
  icmp:
    - type: 7
      code: 3
```

Option 3:
```yaml
protocols:
  - icmp:
      type: 7
      code: 3
```

A likely option is that we will start with the `ICMP: match all` option, as more details matching is not very popular,
but we also want to leave space for future extensions (adding more fields), so the struct should allow both

Option 1:
```yaml
protocols:
  - protocol: ICMP
    icmp:
      matchAll: true
```
vs
Option 2:
```yaml
protocols:
  icmp: 
    - matchAll: true
```
vs
Option 3:
```yaml
protocols:
  - icmp:
      matchAll: true
```

validation is the same for both cases: `matchAll` and other fields like `code` and `type` should not be used at the same time.

To add new match fields to the existing protocols (e.g. TCP flags), we can just add new fields to the existing protocol struct.
Option 2:
```yaml
protocols:
  tcp:
    - port: 
        number: 80
      flags:
        syn: true
```

## src vs dst port confusion

It has always been confusing that we use `port` to mean `destinationPort` in both egress and ingress peer types.
So for egress rules, we set both peers and port for destination matching, 
while for ingress rules, we set peers for source matching and port for destination matching.
To make this more obvious, we agreed to rename `port` to `destinationPort` in the new API version.
For consistency, we will also rename `namedPort` to `destinationNamedPort`.

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
