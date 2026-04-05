# NPEP-133: FQDN Selector for Egress Traffic

* Issue:
  [#133](https://github.com/kubernetes-sigs/network-policy-api/issues/133)
* Status: Implementable

## TLDR

This enhancement proposes adding a new optional selector to specify egress peers
using [Fully Qualified Domain
Names](https://www.wikipedia.org/wiki/Fully_qualified_domain_name) (FQDNs).

## Goals

* Provide a selector to specify egress peers using a Fully Qualified Domain Name
  (for example `kubernetes.io`).
* Support basic wildcard matching capabilities when specifying FQDNs (for
  example `*.cloud-provider.io`)
* Currently only `ACCEPT` type rules are proposed.
  * Safely enforcing `DENY` rules based on FQDN selectors is difficult as there
    is no guarantee a Network Policy plugin is aware of all IPs backing a FQDN
    policy. If a Network Policy plugin has incomplete information, it may
    accidentally allow traffic to an IP belonging to a denied domain. This would
    constitute a security breach.
    
    By contrast, `ACCEPT` rules, which may also have an incomplete list of IPs,
    would not create a security breach. In case of incomplete information, valid
    traffic would be dropped as the plugin believes the destination IP does not
    belong to the domain. While this is definitely undesirable, it is at least
    not an unsafe failure.

* DomainNames may only be used with Accept action (not Pass or Deny).
  * The rationale for excluding Deny is described above.
  * Pass is excluded for now because no existing implementations support
    DomainNames with Pass and there is no strong use case for it. This
    restriction may be relaxed in the future.

## Non-Goals

* This enhancement does not include a FQDN selector for allowing ingress
  traffic.
* This enhancement only describes enhancements to the existing L4 filtering as
  provided by ClusterNetworkPolicy. It does not propose any new L7 matching or
  filtering capabilities, like matching HTTP traffic or URL paths.
  * This selector should not control what DNS records are resolvable from a
    particular workload.
  * This selector provides no capability to detect traffic destined for
    different domains backed by the same IP (e.g. CDN or load balancers).
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

* As a cluster admin, I want to allow Pods in the cluster to send traffic to a
  entire tree of domains. For example, our CDN has domains of the format
  `<session>.<random>.<region>.my-app.cdn.com`. I want to be able to use a
  wild-card selector to allow the full tree of subdomains below
  `*.my-app.cdn.com`.

### Future User Stories

These are some user stories we want to keep in mind, but due to limitations of
the existing Network Policy API, cannot be implemented currently. The design
goal in this case is to ensure we do not make these unimplementable down the
line.

* As a cluster admin, I want to switch the default disposition of the cluster to
  be default deny. This is enforced using a Baseline-tier
  `ClusterNetworkPolicy`. I also want individual namespace owners to be able to
  specify their egress peers. Namespace admins would then use a FQDN selector
  in the Kubernetes `NetworkPolicy` objects to allow `my-service.com`.
  
## API

This NPEP proposes adding a `DomainNames` field to
`ClusterNetworkPolicyEgressPeer` which allows specifying domain names as
egress peers. DomainNames is only available with Accept rules.

This restriction is enforced via CEL validation on
`ClusterNetworkPolicyEgressRule`:

```go
// +kubebuilder:validation:XValidation:rule="self.action != 'Accept' ? !self.to.exists(peer, has(peer.domainNames)) : true",message="domainNames may only be used with Accept action"
type ClusterNetworkPolicyEgressRule struct { ... }
```

```golang

// DomainName describes one or more domain names to be used as a peer.
//
// DomainName can be an exact match, or use the wildcard specifier '*' to match
// one or more labels.
//
// '*', the wildcard specifier, matches one or more entire labels. It does not
// support partial matches. '*' may only be specified as a prefix.
// 
//  Examples:
//    - `kubernetes.io` matches only `kubernetes.io`.
//      It does not match "www.kubernetes.io", "blog.kubernetes.io",
//      "my-kubernetes.io", or "wikipedia.org".
//    - `blog.kubernetes.io` matches only "blog.kubernetes.io".
//      It does not match "www.kubernetes.io" or "kubernetes.io".
//    - `*.kubernetes.io` matches subdomains of kubernetes.io.
//      "www.kubernetes.io", "blog.kubernetes.io", and
//      "latest.blog.kubernetes.io" match, however "kubernetes.io", and
//      "wikipedia.org" do not.
//
// +kubebuilder:validation:Pattern=`^(\*\.)?([a-zA-z0-9]([-a-zA-Z0-9_]*[a-zA-Z0-9])?\.)+[a-zA-z0-9]([-a-zA-Z0-9_]*[a-zA-Z0-9])?\.?$`
type DomainName string

type ClusterNetworkPolicyEgressPeer struct {
    <snipped>
    // DomainNames provides a way to specify domain names as peers.
    //
    // DomainNames is only supported for Accept rules.
    // In order to control access, DomainNames Accept rules should be used
    // with a lower precedence egress deny -- this allows the admin to
    // maintain an explicit "allowlist" of reachable domains.
    // 
    // Support: Extended
    //
    // <network-policy-api:experimental>
    // +optional
    // +listType=set
    // +kubebuilder:validation:MinItems=1
    // +kubebuilder:validation:MaxItems=25
    DomainNames []DomainName `json:"domainNames,omitempty"`
}
```

### Examples

#### Pods in `monitoring` namespace can talk to `my-service.com` and `*.cloud-provider.io`

```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: allow-my-service-egress
spec:
  tier: Admin
  priority: 55
  tier: Admin
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: "monitoring"
  egress:
  - name: "allow-to-my-service"
    action: "Accept"
    to:
    - domainNames:
      - "my-service.com"
      - "*.cloud-provider.io"
    protocols:
    - tcp:
        destinationPort:
          number: 443
```

#### Maintaining an allowlist of domains

There are a couple ways to maintain an allowlist:

This example includes the Deny rule in the same ClusterNetworkPolicy object.
It's also possible to use another ClusterNetworkPolicy object with a lower
priority (e.g. `100` in this example):
```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: allow-my-service-egress
spec:
  tier: Admin
  priority: 55
  tier: Admin
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: "monitoring"
  egress:
  - name: "allow-to-my-service"
    action: "Accept"
    to:
    - domainNames:
      - "my-service.com"
      - "*.cloud-provider.io"
    protocols:
    - tcp:
        destinationPort:
          number: 443
  - name: "default-deny"
    action: "Deny"
    to:
    - networks:
      - "0.0.0.0/0"
      - "::/0"
```

This example uses a Baseline-tier default-deny ClusterNetworkPolicy to create
the allowlist:
```yaml
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: allow-my-service-egress
spec:
  tier: Admin
  priority: 55
  tier: Admin
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: "monitoring"
  egress:
  - name: "allow-to-my-service"
    action: "Accept"
    to:
    - domainNames:
      - "my-service.com"
      - "*.cloud-provider.io"
    protocols:
    - tcp:
        destinationPort:
          number: 443
---
apiVersion: policy.networking.k8s.io/v1alpha2
kind: ClusterNetworkPolicy
metadata:
  name: default-deny
spec:
  tier: Baseline
  priority: 0
  subject:
    namespaces: {}
  egress:
  - name: "default-deny"
    action: "Deny"
    to:
    - networks:
      - "0.0.0.0/0"
      - "::/0"
```

### Expected Behavior

1. A FQDN egress policy does not grant the workload permission to communicate
   with any in-cluster DNS services (like `kube-dns`). A separate rule needs to
   be configured to allow traffic to any DNS servers. FQDN policies are not
   expected to work if the pod cannot reach DNS.
1. FQDN policies should not affect the ability of workloads to resolve domains,
   only their ability to communicate with the IP backing them. Put another way,
   FQDN policies should not result in any form of DNS filtering.
   *  For example, if a policy allows traffic to `kubernetes.io`, any selected
      Pods can still resolve `wikipedia.org` or
      `my-services.default.svc.cluster.local`, but can not send traffic to them
      unless allowed by a different rule.
1. Each implementation will provide guidance on which DNS name-server is
   considered authoritative for resolving domain names. This could be the
   `kube-dns` Service or potentially some other DNS provider specified in the
   implementation's configuration.
1. Pods are expected to use the DNS configuration provided via `resolv.conf`
   (i.e. the canonical cluster DNS server). If a pod uses a different DNS
   server (e.g. hardcoded `8.8.8.8`), FQDN rule processing is not guaranteed
   to work.
1. DNS record querying and lifetimes:
   *  Pods are expected to make a DNS query for a domain before sending traffic
      to it. If the Pod fails to send a DNS request and instead just sends
      traffic to the IP (either because of caching or a static config), traffic
      is not guaranteed to flow.
   *  Pods should respect the TTL of DNS records they receive. Trying to
      establish new connection using DNS records that are expired is not
      guaranteed to work.
   *  When the TTL for a DNS record expires, the implementor should stop
      allowing new connections to that IP. Existing connections will still be
      allowed (that's consistent with NetworkPolicy behavior on long-running
      connections).
   *  If a DNS record is refreshed before the TTL expires (e.g., due to a new
      DNS query from the workload), the TTL timer should be reset based on the
      new response.
1. Implementations must support at least 100 unique IPs (either IPv4 or IPv6)
   for each domain. This is true for both explicitly specified domains, as well
   as for each domain selected by a wild-card rule. For example, the rule
   `*.kubernetes.io` supports 100 IPs each for both `docs.kubernetes.io` and
   `blog.kubernetes.io`.
1. PTR records are not required to properly configure a FQDN selector. For
   example, as long as an A record exists mapping `my-hostname` to `1.2.3.4`,
   the Network Policy implementation should allow traffic to `1.2.3.4`. There is
   no requirement that a PTR record for `1.2.3.4.in-addr.arpa` exist or that it
   points to `my-hostname` (it is allowed to point to `other-host`).
1. Targeting in-cluster endpoints with FQDN selector is not recommended. There
   are other selectors which can more precisely capture intent. However, if
   in-cluster endpoints are selected:
   *  ✅︎ Supported:
      *  Selecting Pods using their [generated DNS
         record](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pods)
         (for example `pod-ip-address.my-namespace.pod.cluster.local`). This is
         analogous to selecting the Pod by its IP address using the Network
         selector.
      * Headless Services can be selected using their [generated DNS
        record](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#services)
        because the generated DNS records contain a list of all the Pod IPs that
        back the service.
   *  ❌ Not Supported:
      * ClusterIP Services can not be selected using their [generated DNS
        record](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#services)
        (for example `my-svc.my-namespace.svc.cluster.local`). This is
        consistent with the behavior when selecting the Service VIP using the
        Network selector.
      * ExternalName Services return a `CNAME` record. See the entry below
        about CNAME support.
      * Any record which points to the IPs used for `LoadBalancer` type
        services. This includes the `externalIPs` and the
        `.status.loadBalancer.ingress` fields
1. If the specified domain in a FQDN selector resolves to a [CNAME
   record](#cname-records) the behavior of the implementor depends on the
   returned response.

   If the upstream resolver used [CNAME chasing](#cname-chasing) to fully
   resolve the domain to a A/AAAA record and returns the resulting chain, the
   implementor can use this information to allow traffic to the specified IPs.
   However the implementor does not need to perform their own CNAME chasing or
   to understand resolutions across multiple DNS requests.

   For example, if the FQDN selector is allowing traffic to `www.kubernetes.io`:
     *  If a DNS query to the upstream resolver returns *a single response* with
        the following records:
        ```
        www.kubernetes.io  --  CNAME to kubernetes.io
        kubernetes.io      --  A     to 1.2.3.4
        ```
        The implementor can use this response to allow traffic to `1.2.3.4`
     *  If DNS query only responds with a CNAME record, the resolver is not
        required to allow traffic even if subsequent requests resolve the full
        chain:
        ```
        # REQUEST 1

        www.kubernetes.io  --  CNAME to kubernetes.io
        
        # REQUEST 2

        kubernetes.io      --  A     to 1.2.3.4
        ```
        The implementer can still deny traffic to `1.2.3.4` because no single
        response contained the full chain required to resolve the domain.

### Recommended Behavior

The following recommendations are based on operational experience from existing
implementations and feedback gathered during KubeCon Atlanta 2025.

1. Pods should only make DNS requests to the canonical DNS server, because
   FQDN rules for a pod are only guaranteed to work after that pod makes an
   appropriate DNS request to the canonical DNS server (see [Expected
   Behavior](#expected-behavior)). If a pod queries an alternate DNS server
   (e.g. hardcoded `8.8.8.8`), the implementation may not observe the DNS
   response, and the FQDN Accept rule will simply not take effect. This is
   a functionality issue, not a security breach -- the pod's traffic is
   denied, not accidentally allowed. For example, when a pod queries the
   canonical DNS server for `my-service.com` and receives `1.2.3.4`, the
   implementation observes that response and learns that `my-service.com`
   maps to `1.2.3.4`. It can then enforce the FQDN Accept rule by allowing
   traffic to `1.2.3.4`. If the pod instead queries `8.8.8.8`, the
   implementation never sees the response and has no IP to allow -- the rule
   simply doesn't take effect.
1. Administrators should create a high-precedence Admin-tier rule allowing
   egress DNS traffic to the canonical DNS server (e.g. `kube-dns` in
   `kube-system`), to ensure that DNS is not blocked by other deny rules.
   For example:
   ```yaml
   apiVersion: policy.networking.k8s.io/v1alpha2
   kind: ClusterNetworkPolicy
   metadata:
     name: allow-dns
   spec:
     tier: Admin
     priority: 50
     subject:
       namespaces:
         matchLabels:
           requires-fqdn-policy: "true"
     egress:
     - name: "allow-dns"
       action: "Accept"
       to:
       - namespaces:
           matchLabels:
             kubernetes.io/metadata.name: "kube-system"
       protocols:
       - udp:
           destinationPort:
             number: 53
       - tcp:
           destinationPort:
             number: 53
   ```
1. Implementations that operate by snooping DNS responses on the wire MUST
   only trust responses originating from the canonical DNS server. Trusting
   responses from arbitrary sources is a security concern: a malicious actor
   could forge DNS responses to trick the implementation into allowing
   traffic to unintended IPs.
1. Implementations MAY provide a configurable grace period beyond the TTL to
   accommodate DNS propagation delays and client-side caching.
1. Although securing DNS resolution is a non-goal of this NPEP, implementations
   are recommended to consider mitigations for DNS cache poisoning (e.g.,
   DNSSEC validation by the canonical DNS server) when documenting their trust
   model.

## Connection Lifecycle on Policy Updates

When FQDN policies change, implementations must handle in-flight connections
gracefully. See the [Expected Behavior](#expected-behavior) section for DNS
record updates.

* **Policy addition**: New connections matching the FQDN are allowed once DNS
  resolution for the domain has been observed. Connections to IPs that haven't
  been observed via DNS are not guaranteed to be allowed.
* **Policy removal**: Existing established connections SHOULD be allowed to
  complete gracefully. New connections SHOULD be denied after the policy is
  removed. This is consistent with the general NetworkPolicy behavior for
  long-running connections.

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
| Wildcards      | ✅︎                              | ️✅︎           | ✅︎            | ❌                         | ✅︎                        |
| Egress Rules   | ✅︎                              | ️✅︎           | ✅︎            | ✅︎                         | ✅︎                        |
| Ingress Rules  | ❌                              | ️❌           | ❌            | ❌                         | ❌                        |
| Allow Rules    | ✅︎                              | ️✅︎           | ✅︎            | ✅︎                         | ✅︎                        |
| Deny Rules     | ✅︎                              | ️❌(?)        | ❌            | ✅︎                         | ❌(?)                     |

## Appendix

### CNAME Records

CNAME records are a type of DNS record (like a `A` or `AAAA`) that direct the
resolver to query another name to retrieve actual A/AAAA records.

For example:
```
$ dig www.kubernetes.io

... Omitted output ...

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;www.kubernetes.io.		IN	A

;; ANSWER SECTION:
www.kubernetes.io.	3600	IN	CNAME	kubernetes.io.
kubernetes.io.		3600	IN	A	147.75.40.148

... Omitted Output ...
```
### CNAME Chasing

CNAME chasing refers to an optional behavior for DNS resolvers whereby they
perform subsequent lookups to resolve CNAMEs returned for a particular query. In
the above example, querying for `www.kubernetes.io.` returned a CNAME record for
`kubernetes.io.`. When CNAME chasing is enabled, the DNS server will
automatically resolve `kubernetes.io.` and return both records as the DNS
response.
