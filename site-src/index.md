## What is the AdminNetworkPolicy API?

The AdminNetworkPolicy API is an open source project managed by the [SIG-NETWORK][sig-network]
community. It is a collection of resources, which aim to make securing Kubernetes 
clusters easier for Administrators.  

![Gateway API Model](./images/ANP-api-model.png)

## Getting started

Whether you are a user interested in using the AdminNetworkPolicy API or an implementer 
interested in conforming to the API, the following resources will help give 
you the necessary background:

- [API overview](/guides/api-overview)
- [Examples](/guides/examples)
- [API Source](https://github.com/kubernetes-sigs/network-policy-api/tree/master/apis/v1alpha1)


## AdminNetworkPolicy API User Stories

The following user stories drive the concepts of the AdminNetworkPolicy API for the 
`v1alpha1` version of the api. More information on how the community ended up here 
can be found in the [API KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/2091-admin-network-policy)
and in the accompanying [KEP PR](https://github.com/kubernetes/enhancements/pull/2522)

Future API developments should all start with a **well-defined** and **intentional** 
user story.

### Story 1: Deny traffic at a cluster level

As a cluster admin, I want to apply non-overridable deny rules
to certain pod(s) and(or) Namespace(s) that isolate the selected
resources from all other cluster internal traffic.

For Example: In this diagram there is a AdminNetworkPolicy applied to the
`sensitive-ns` denying ingress from all other in-cluster resources for all
ports and protocols.

![Alt text](./images/explicit_deny.png?raw=true "Explicit Deny")

### Story 2: Allow traffic at a cluster level

As a cluster admin, I want to apply non-overridable allow rules to  
certain pods(s) and(or) Namespace(s) that enable the selected resources
to communicate with all other cluster internal entities.  

For Example: In this diagram there is a AdminNetworkPolicy applied to every
namespace in the cluster allowing egress traffic to `kube-dns` pods, and ingress
traffic from pods in `monitoring-ns` for all ports and protocols.

![Alt text](./images/explicit_allow.png?raw=true "Explicit Allow")

### Story 3: Explicitly Delegate traffic to existing K8s Network Policy

As a cluster admin, I want to explicitly delegate traffic so that it
skips any remaining cluster network policies and is handled by standard
namespace scoped network policies.

For Example: In the diagram below egress traffic destined for the service svc-pub
in namespace bar-ns-1 on TCP port 8080 is delegated to the k8s network policies
implemented in foo-ns-1 and foo-ns-2. If no k8s network policies touch the
delegated traffic the traffic will be allowed.

![Alt text](./images/delegation.png?raw=true "Delegate")

### Story 4: Create and Isolate multiple tenants in a cluster

As a cluster admin, I want to build tenants in my cluster that are isolated from
each other by default. Tenancy may be modeled as 1:1, where 1 tenant is mapped
to a single Namespace, or 1:n, where a single tenant may own more than 1 Namespace.

For Example: In the diagram below two tenants (Foo and Bar) are defined such that
all ingress traffic is denied to either tenant.  

![Alt text](./images/tenants.png?raw=true "Tenants")

### Story 5: Cluster Wide Default Guardrails

As a cluster admin I want to change the default security model for my cluster,
so that all intra-cluster traffic (except for certain essential traffic) is
blocked by default. Namespace owners will need to use NetworkPolicies to
explicitly allow known traffic. This follows a whitelist model which is
familiar to many security administrators, and similar
to how [kubernetes suggests network policy be used](https://kubernetes.io/docs/concepts/services-networking/network-policies/#default-policies).

For Example: In the following diagram all Ingress traffic to every cluster
resource is denied by a baseline deny rule.

![Alt text](./images/baseline.png?raw=true "Default Rules")

## Who is working on AdminNetworkPolicy?

The AdminNetworkPolicy API is a
[SIG-Network](https://github.com/kubernetes/community/tree/master/sig-network)
project being built to improve and standardize cluster-wide security policy in k8s. In-progress implementations include Antrea (VMware) and Openshift (RedHat),
If you are interested in contributing to or
building an implementation using the AdminNetworkPolicy API then don’t hesitate to [get
involved!](/contributing/community)

[sig-network]: https://github.com/kubernetes/community/tree/master/sig-network

