# How to secure a K8s cluster using NetworkPolicys

## Current

NetworkPolicy API is an application-centric construct allowing users to
specify how a Pod is allowed to communicate with various entities over the
network. This allows developers to secure their applications from unwanted
network access. However, this does not allow admins or other personas to secure
the cluster as a whole using K8s constructs. Admins must co-ordinate with
platform operators or other APIs outside of K8s to enforce additional security
measures. This is being done via either a set of config options for the cluster
or implicitly enforced by the provider which the developers may not be aware of,
or varies from provider to provider and the experience is not guaranteed to be
uniform across.

### Network Policy Use Cases

Network policy is used for the following scenarios:
1. Implementing microsegmentation in a multi-tenant cluster
2. Restricting pods from being able to talk to the internet/highly secure VMs
3. Forcing all egress traffic to go through an egress gateway
4. Protecting pods from DoS attacks

## Proposal

To get more context on this document please refer to this original [proposal](https://docs.google.com/document/d/10t4q5XO1ED2PnK3ishn4y3G4Tma7uMYgesG-itQHMiU/edit#).

### Security model

Before doing a deep dive in the design and implementation of the proposal, it
is worthwhile describing the personas and the roles that these personas play
in defining the security model of a K8s cluster.

The following roles may exist in a lifecycle of a K8s cluster:
- Platform: responsible for setting up and maintaing the cluster
- SecurityOps: responsible for outlining the security of the cluster as whole
- NetworkOps: responsible for maintaining and setting up the networking
  resources of the cluster
- Application Developer: responsible for bringing up Services within Namespaces

It is not necessary for the above roles to be mapped 1:1 to individuals. One or
more role may be shared by individuals or a group of users and how it is mapped
may differ based on the structure of the organization.

In the simplest of form, all the above roles may be shared by a single person,
in which case the security model is greatly simplified. However, in most cases,
there would be at least two or more roles, where in an application developer
group is responsible for the apps or Services that they maintain within their
Namespaces, and are only concerned with how their Services interact with the
network, while the other set of roles have a global view of the cluster and
are responsible for the security of the cluster as a whole, and outline the
policies that must be adhered to by all the Services inside the cluster, and/or
specify the default action to be taken within the cluster, if none of the
policies match.

Thus, the application developers intent differs from that of the cluster
administrator, and they may end up writing security policies which are not
alike each other. It can also be debated that a given cluster may have more
roles with different authorities and precedence, and as such the security
policy written by each of them may differ from each other, and there is a
need to group these policies together to form a hierarchy of security
policies, where in one group of policies is responsible for a subset of
firewall rules, and delegates the responsibility of other rules to another
group of policies, thereby introducing the concept of "ordered group of
policies" or "tiered policies".

In either case, it is clear that there is a need for another API, which
clearly expresses the intent of a role which has higher precedence than
the application developers. This API must satisfy the requirements of the
other cluster scoped roles, and have higher precedence than that of the
application developer written NetworkPolicies. Eventually, we could
introduce a concept to group like minded policies together to form a
hierarchy and assign each group to a role.

#### Administrator focused policies

For simplicity, let us consider all roles except for Application developers
role as a single "admin" role. The admin role is concerned with the following:

- Pods: Responsible for all the Pods in the cluster, as opposed to Pods backing
  individual Services. May also want to specify policy for group of Pods
  across Namespaces.
- Namespaces: Responsible for all Namespaces or group of Namespaces and dictate
  the interaction between Namespaces.
- Nodes: Responsible for all the Nodes and how Pods may be able to access these
  Nodes (or what ports can be accessed). Specify rules for Pods access to
  Services based on the Nodes on which they are hosted.
- Cluster external: Outline how Pods within the cluster are allowed to access
  the internet, i.e. block specific traffic, redirect traffic through a secure
  gateway etc.
- Precedence: Responsible to write policies which supersedes all developer
  written policies. In addition to that, also write default policies for the
  cluster which act as a failsafe mechanism for the cluster, when no rule,
  either set by the admin or by the developers NetworkPolicy, is matched.
- Auditing: Responsible to outline the logging policy for traffic matching
  the rules set by the admin, to carefully review and feed it to an analyzer
  or visualization component.

#### Developer focused policies

The developer focused policies can already be expressed with the current
NetworkPolicy APIs. The evolution of these APIs can either be additive, i.e.
add new backwards compatible fields, or deprecate existing APIs in favor of
new dev focused API which can encompass a wider set of features with a
consistent API for ingress and egress traffic.

The dev role is concerned with the following:
- Pods: responsible in securing Pods backing the developer's Services.
- Cluster external: After conforming to the admin policies, specify further
  security constraints on the developers Service w.r.t. to external sites.
- Services: Exposing Services externally via Ingress or any other supported
  means should automatically translate any network mappings that may exist
  in the path of the Service and desired external entities.
- Namespaces: responsible to control traffic to their Pods, to/from other
  Pods belonging to other users/Namespaces.

### Design and implementation

1. Detailed design for the cluster scoped policies can be found [here](1_cluster_scoped.md). 
2. Detailed design for the developer scoped policies can be found [here](1_dev_scoped.md). 
