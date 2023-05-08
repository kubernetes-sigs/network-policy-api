# User Stories

**ALL** Network Policy API resources and future API developments should start with
a **well-defined** and **intentional** user story(s).

## AdminNetworkPolicy + BaselineAdminNetworkPolicy

### v1alpha1 User Stories

The following user stories drive the concepts for the `v1alpha1` version of the
ANP and BANP resources. More information on how the community ended up here
can be found in the [API KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/2091-admin-network-policy)
and in the accompanying [KEP PR](https://github.com/kubernetes/enhancements/pull/2522)

#### Story 1: Deny traffic at a cluster level

As a cluster admin, I want to apply non-overridable deny rules
to certain pod(s) and(or) Namespace(s) that isolate the selected
resources from all other cluster internal traffic.

For Example: In this diagram there is a AdminNetworkPolicy applied to the
`sensitive-ns` denying ingress from all other in-cluster resources for all
ports and protocols.

![Alt text](./images/explicit_deny.png?raw=true "Explicit Deny")

#### Story 2: Allow traffic at a cluster level

As a cluster admin, I want to apply non-overridable allow rules to  
certain pods(s) and(or) Namespace(s) that enable the selected resources
to communicate with all other cluster internal entities.  

For Example: In this diagram there is a AdminNetworkPolicy applied to every
namespace in the cluster allowing egress traffic to `kube-dns` pods, and ingress
traffic from pods in `monitoring-ns` for all ports and protocols.

![Alt text](./images/explicit_allow.png?raw=true "Explicit Allow")

#### Story 3: Explicitly Delegate traffic to existing K8s Network Policy

As a cluster admin, I want to explicitly delegate traffic so that it
skips any remaining cluster network policies and is handled by standard
namespace scoped network policies.

For Example: In the diagram below egress traffic destined for the service svc-pub
in namespace bar-ns-1 on TCP port 8080 is delegated to the k8s network policies
implemented in foo-ns-1 and foo-ns-2. If no k8s network policies touch the
delegated traffic the traffic will be allowed.

![Alt text](./images/delegation.png?raw=true "Delegate")

#### Story 4: Create and Isolate multiple tenants in a cluster

As a cluster admin, I want to build tenants in my cluster that are isolated from
each other by default. Tenancy may be modeled as 1:1, where 1 tenant is mapped
to a single Namespace, or 1:n, where a single tenant may own more than 1 Namespace.

For Example: In the diagram below two tenants (Foo and Bar) are defined such that
all ingress traffic is denied to either tenant.  

![Alt text](./images/tenants.png?raw=true "Tenants")

#### Story 5: Cluster Wide Default Guardrails

As a cluster admin I want to change the default security model for my cluster,
so that all intra-cluster traffic (except for certain essential traffic) is
blocked by default. Namespace owners will need to use NetworkPolicies to
explicitly allow known traffic. This follows a whitelist model which is
familiar to many security administrators, and similar
to how [kubernetes suggests network policy be used](https://kubernetes.io/docs/concepts/services-networking/network-policies/#default-policies).

For Example: In the following diagram all Ingress traffic to every cluster
resource is denied by a baseline deny rule.

![Alt text](./images/baseline.png?raw=true "Default Rules")
