# NPEP-133: FQDN Selector for Egress Traffic

* Issue:
  [#133](https://github.com/kubernetes-sigs/network-policy-api/issues/133)
* Status: Provisional

## TLDR

This enhancement proposes adding a new optional selector to specify egress peers
using [Fully Qualified Domain
Names](https://www.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDNs).

## Goals

* Provide a selector to specify egress peers using a Fully Qualified Domain Name
  (for example `kubernetes.io`).
* Support basic wildcard matching capabilities when specifying FQDNs (for
  example `*.cloud-provider.io`)
* Currently only `ALLOW` type rules are proposed.
  * Safely enforcing `DENY` rules based on FQDN selectors is difficult as there
    is no guarantee a Network Policy plugin is aware of all IPs backing a FQDN
    policy. If a Network Policy plugin has incomplete information, it may
    accidentally allow traffic to an IP belonging to a denied domain. This would
    constitute a security breach.
    
    By contrast, `ALLOW` rules, which may also have an incomplete list of IPs,
    would not create a security breach. In case of incomplete information, valid
    traffic would be dropped as the plugin believes the destination IP does not
    belong to the domain. While this is definitely undesirable, it is at least
    not an unsafe failure.

* Currently only AdminNetworkPolicy is the intended scope for this proposal.
  * Since Kubernetes NetworkPolicy does not have a FQDN selector, adding this
    capability to BaselineAdminNetworkPolicy could result in writing baseline
    rules that can't be replicated by an overriding NetworkPolicy. For example,
    if BANP allows traffic to `example.io`, but the namespace admin installs a
    Kubernetes Network Policy, the namespace admin has no way to replicate the
    `example.io` selector using just Kubernetes Network Policies.

## Non-Goals

* This enhancement does not include a FQDN selector for allowing ingress
  traffic.
* This enhancement only describes enhancements to the existing L4 filtering as
  provided by AdminNetworkPolicy. It does not propose any new L7 matching or
  filtering capabilities, like matching HTTP traffic or URL paths.
  * This selector should not control what DNS records are resolvable from a
    particular workload.
* This enhancement does not provide a mechanism for selecting in-cluster
  endpoints using FQDNs. To select Pods, Nodes, or the API Server,
  AdminNetworkPolicy has other more specific selectors.
  * Using the FQDN selector to refer to other Kubernetes endpoints, while not
    explicitly disallowed, is not defined by this spec and left up to individual
    providers. Trying to allow traffic to the following domains is NOT
    guaranteed to work:
    * `my-svc.my-namespace.svc.cluster.local` (the generated DNS record for a
      Service as defined
      [here](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#services))
    * `pod-ip-address.my-namespace.pod.cluster.local` (the generated DNS record
      for a Pod as defined
      [here](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pods))
* This enhancement does not add any new mechanisms for specifying how traffic is
  routed to a destination (egress gateways, alternative SNAT IPs, etc). It just
  adds a new way of specifying packets to be allowed or dropped on the normal
  egress data path.
* This enhancement does not require any mechanism for securing DNS resolution
  (e.g. DNSSEC or DNS-over-TLS). Unsecured DNS requests are expected to be
  sufficient for looking up FQDNs.

## Introduction

FQDN-based egress controls are a common enterprise security practice.
Administrators often prefer to write security policies using DNS names such as
“www.kubernetes.io” instead of capturing all the IP addresses the DNS name might
resolve to. Keeping up with changing IP addresses is a maintenance burden, and
hampers the readability of the network policies.

## User Stories

* As a cluster admin, I want to allow all Pods in the cluster to send traffic to
  an external service specified by a well-known domain name. For example, all
  Pods must be able to talk to `my-service.com`.

* As a cluster admin, I want to allow Pods in the "monitoring" namespace to be
  able to send traffic to a logs-sink, hosted at `logs-storage.com`

* As a cluster admin, I want to allow all Pods in the cluster to send traffic to
  any of the managed services provided by my Cloud Provider. Since the cloud
  provider has a well known parent domain, I want to allow Pods to send traffic
  to all sub-domains using a wild-card selector -- `*.my-cloud-provider.com`

### Future User Stories

These are some user stories we want to keep in mind, but due to limitations of
the existing Network Policy API, cannot be implemented currently. The design
goal in this case is to ensure we do not make these unimplementable down the
line.

* As a cluster admin, I want to block all cluster egress traffic by default, and
  require namespace admins to create NetworkPolicies explicitly allowing egress
  to the domains they need to talk to.

  The Cluster admin would use a `BaselineAdminNetworkPolicy` object to switch
  the default disposition of the cluster. Namespace admins would then use a FQDN
  selector in the Kubernetes `NetworkPolicy` objects to allow `my-service.com`.
  
## API

TODO: https://github.com/kubernetes-sigs/network-policy-api/issues/133

## Alternatives

### IP Block Selector

IP blocks are an important tool for specifying Network Policies. However, they
do not address all user needs and have a few short-comings when compared to FQDN
selectors:

* IP-based selectors can become verbose if a single logical service has numerous
  IPs backing it.
* IP-based selectors pose an ongoing maintenance burden for administrators, who
  need to be aware of changing IPs.
* IP-based selectors can result in policies that are difficult to read and
  audit.

### L4 Proxy

Users can also configure a L4 Proxy (e.g. using SOCKS) to inspect their traffic
and implement egress firewalls. They present a few trade-ofs when compared to a
FQDN selector:

* Additional configuration and maintenance burden of the proxy application
  itself
* Configuring new routes to direct traffic leaving the application to the L4
  proxy.  

### L7 Policy

Another alternative is to provide a L7 selector, similar to the policies
provided by Service Mesh providers. While L7 selectors can offer more
expressivity, they often come trade-offs that are not suitable for all users:

* L7 selectors necessarily support a select set of protocols. Users may be
  using a custom protocol for application-level communication, but still want
  the ability to specify endpoints using DNS.
* L7 selectors often require proxies to perform deep packet inspection and
  enforce the policies. These proxies can introduce un-desireable latencies in
  the datapath of applications.

## References

* [NPEP #126](https://github.com/kubernetes-sigs/network-policy-api/issues/126):
  Egress Control in ANP

### Implementations

* [Antrea](https://antrea.io/docs/main/docs/antrea-network-policy/#fqdn-based-filtering)
* [Calico](https://docs.tigera.io/calico-enterprise/latest/network-policy/domain-based-policy)
* [Cilium](https://docs.cilium.io/en/latest/security/policy/language/#dns-based)
* [OpenShift](https://docs.openshift.com/container-platform/latest/networking/openshift_sdn/configuring-egress-firewall.html)

The following is a best-effort breakdown of capabilities of different
NetworkPolicy providers, as of 2023-09-25. This information may be out-of-date,
or inaccurate.

|                | Antrea                         | Calico       | Cilium       | OpenShift <br/> (current) | OpenShift <br/> (future) |
| -------------- | ------------------------------ | ------------ | ------------ | ------------------------- | ------------------------ |
| Implementation | DNS Snooping <br/> + Async DNS | DNS Snooping | DNS Snooping | Async DNS                 | DNS Snooping             |
| Wildcards      | ✔                              | ️✔           | ✔            | ❌                         | ✔                        |
| Egress Rules   | ✔                              | ️✔           | ✔            | ✔                         | ✔                        |
| Ingress Rules  | ❌                              | ️❌           | ❌            | ❌                         | ❌                        |
| Allow Rules    | ✔                              | ️✔           | ✔            | ✔                         | ✔                        |
| Deny Rules     | ✔                              | ️❌(?)        | ❌            | ✔                         | ❌(?)                     |