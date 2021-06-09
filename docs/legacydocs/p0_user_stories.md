# NetworkPolicy User Stories

Initial proposal or priorities, based on feasabiliy and overall user engagement... reprioritization welcome, please PR changes and in case of any major conflicts we can discuss in the broader group.  

Taken from 
- conversations and zoom meetings
- https://docs.google.com/document/d/10t4q5XO1ED2PnK3ishn4y3G4Tma7uMYgesG-itQHMiU/edit# 
- https://docs.google.com/document/d/1AtWQy2fNa4qXRag9cCp5_HsefD7bxKe3ea2RPn8jnSs/edit . 

This document attempts to formalize and consolidate the many docs and spreadsheets from the original meetings into a single browsable artifact, which eventually can translate into KEP-like-documents, KEPs, and possibly example CRDs and other artifacts which can be used to capture the overall evolution of the network policy working group over time.

## Contributing

Please file an issue, issue a PR, or cast a vote.  The more Feedback we get, the better !

## Terminology

- *app*: a program or collection of programs running in a namespace that comprise a end-user consumed service.  Common "guestbook", "blackduck", "backend API services", or the "ELK" stack, which usually is consumed through Kibana.
- *tier-1 policies*: this is a division between user stories that we've made based on feedback over several months.  Some policies were generally agreed on as more relevant to a broader group, and thus categorized that way (tier-1).  
- *tier-2 policies* Others were considered interesting, but either out of scope or simply very difficult to "nail down" in terms of precise semantics.  These are in tier-2.  They aren't necessarily "bad ideas" but rather, will require deeper ownership and commitment then the tier-1 policies to drive forward.
- *modifying* these are policies which are likely to modify the API in a way that might make them harder to implement/propose to the broader sig.
- *non-modifying* these are policies which likely wont break existing semantics of v1 policies.

-----------------------------------------------------------

# All User Stories

## Application scoped User stories

### tier-1

#### non v1 modifying
- I want 2 apps in different namespaces to *connect via namespace NAME*, because I
  can't read (or write) the labels for those namespaces from my Kubernetes
  client.   I cannot select pods for this because I don't know all the labels
  and names of pod/services that I want to contact ahead of time.
  - [Name as Policy Target](stories/name_as_policy_target.md)

- I want different pods (with potentially different labels) that fulfill a service to be able talk to each other, without talking to other services but I don't know the labels on these services.
  - [Ingress Rules Targeting Services](stories/ingress_rules_targeting_services.md)

- I dont want to look at 10 or 100 policies to figure out whether I have the right allow rules.
  - [Pod Reachability Query](stories/pod_reachability_query.md)

- I wrote a policy, but am not sure if my policy did the right thing, nor if it has taken effect on the pods that I'm concerned with, yet.
  - [Policy Query](stories/policy_query.md)
  
- I want to restrict egress from my pods based on FQDNs, not IP addresses. For example, "allow pods to egress to www.my-trusted-company.com".
  - [FQDN policy](stories/fqdn_policy.md) (see also https://docs.google.com/document/d/1Htcy4UXKZytUe-lWJIIEJZzoa3MtCMr-Ms_KONaXirM/edit?ts=5f84ec53)

#### v1 modifying

- I want to select *host traffic* from or to a Kubernetes node using the Node resource and its labels, because a I want to protect a set of critical services on a dynamic pool of nodes in my cluster.
  - *TBD ... Do we need parameters to specify when to evaulate policy (pre/post SNAT for example).* ?
  - [Node Selector](stories/node_selector.md)
  
- I don't want to directly update my CIDR rules for a policy every time I add a new node or other group of IPs, which need to have policies associated with them.
  - We can use a slice instead of a string for CIDR (KEP-able)
  - [Named Endpoint Set](stories/named_endpoint_set.md)

- I want to allow my application to communicate with high level ports of another "legacy" application, which is not accessed via a service, and which binds to a random port or binds a a random port (like passive FTP)
  - This can be added as an additional portsRange field, which might be an array of portRange object that contains a from and a to integer field
  - [Port Set/Port Range](stories/port_ranges.md)
  
- I want to select a service to my Network Policy instead of selecting a namespace or a Pod
  - * As a User/App developer, I define a service with an external name, and I want to limit who can have access to this service.
  - * As a User/App developer, I exposed a service using node port/fall-through LB and I want to limit who can access to my services/backend pods from external.  

- Writing network policies is hard, I forget what the defaults for ports, ingress/egress, and nil/empty collections (for label selectors and policy structs) are (see https://github.com/kubernetes/kubernetes/issues/51726 for official issue pointing to this).  This might result in revising or removing policy types.

### tier-2

- I want to be able to scrape from pod endpoints for every pod in my cluster, but can't afford to make new policies for each one given the large rate of pod churn.
  - [Cluster Scoped Policy](stories/cluster_scoped_policy.md)

## Cluster scoped user stories

### tier-1

- I only want pods on nodes that are labeled as “infrastructure” nodes to be able to access the Kubernetes apiserver; 
other pods should be blocked from accessing it via any IP.

- I want all namespaces matching X to be completely 100% locked-down by default; no pods can talk to any other pods until the developers specify policies allowing it - I may also want the converse situation for namespace Y.

- I want to target all apps with network policies, but most of my apps need apiserver access, and i dont know the k8s service IP because all of my clusters have different Service CIDRs.

- I want to block all pods from being able to reach ports on nodes. The rule should automatically cover all nodes, and all IPs on those nodes, except for port 53.

- I want to write policies which deny specific traffic, and I want to assign priorities to my policies, and have the CNI use the priorities to resolve conflicts (recently formally added, mattfennwick)

### tier-2

- I want all namespaces to be “isolated” (don’t accept traffic from other Namespaces) by default; developers must specify if they want other namespaces to be able to access their services.

- I want to administratively put a choke point (gateway) between pods that arent in the same app so i can audit cross-app dependencies and implement ingress controls by default in my cluster.  (Workaround: Create a namespace selector around the targetted namespace, which only allows traffic from a predetermined ingress point.  See https://github.com/jayunit100/network-policy-subproject/issues/14) for further discussion. 

- I want to block all pods from being able to reach the AWS metadata server (169.254.169.254)
except for pods in Namespace X, except I want to run kube2iam as a metadata proxy, so pods trying to access the metadata server should be redirected to that instead.

- I want pods in Namespace X to only have cluster egress via a proxy server provided by Service Y.

- I want all pods to be blocked from accessing the internet, no matter what, without having to explicitly add my K8s CIDR ranges allow rules .  This would be a convenience story (otherwise, CIDR ranges are sufficient to satisfy this user stories). 

- I want all pods to be blocked from accessing the internet by default, with exception allowing case-by-case allow-listing of certain IP ranges.  Convenience/Sugar story.

## Descoped from NetworkPolicy API

These still might be explored by this group but are descoped from the primary use cases reported to the SIG.

### Tier 1

- Prioritize Network Policy - I want to create a rule that I’m sure will be executed before anything else.

- I want to log all the times that someone from the outside world was deined access to a pod in my cluster by a NetworkPolicy rule.

- I want a pool of DMZ nodes in my cluster to prevent all outgoing traffic to my cassandra cluter which has highly sensitive data on it which is known to be vulnerable due to a recently discovered CVE.

- I want a developer in Namespace X to build an app that automatically requests access to a database in Namespace Y, but I don't want to *grant* that access until an admin approves it.

- I want visualization of my network policies, as a tool to help me create accurate network policies as well as to help me verify existing network policies for correctness

### Tier 2

- I want to restrict certain processes in a pod without restricting others, so that  some processes are not able to make certain network calls in a cluster where certain set pods running an old version of Nginx are at risk of being comprimised.

- I want to have a named way to add a policy for containers that can or cannot access a MySQL instance in my data center, without knowing that services IP address.

- I want to have a way to restrict incoming traffic based on the source port of the packet. Something like "deny any incoming traffic comming from 0.0.0.0/0:1234", where 1234 is the source port I want to block, would work.

----------------------------------------------------

# CHANGELOG and VOTES

As we move things via PRs, lets note the context so that we can detect cycles and or changes that are reversing previous user requests.  If we notice any obvious disagreements, we can resolve it as a group.  This is an alterantive to voting which might collect 'passive' opinions which havent been deeply thought out.

## CHANGELOG

- *Ricardo* added service selector policy
- *Mattfennwick* added a tier 1 policy for priorization / resolving 
- *Cody* added links v1modifying generified stories (policy target and so on)
- *Jay* addressed clarity issues in the 'cant connect to the internet' and 'namespace by NAME' user stories.  
- *Abhishek, Chris, Jay* linking to tim hockins issue around defining `empty from` as `none`.


## VOTES (please only upvote a maximum of 3 policies)

- *Matt Fenwick* requested adding a visualization story to tier-1 .   (Interpretted as an upvote)
- *Jay vyas* Moved the abov request for tier-1 to "Descoped" since its not an API thing, but it is a valid user story. 
- *Eric Bannon* strongly upvotes the need for "connect via namespace NAME" (seeing customers run into confusion here a lot) 
- *David Byron* upvotes the "I want all namespaces matching X to be completely 100% locked-down by default" story.
- *Ricardo* upvotes "Port Ranges"
- *Tim Downey* upvotes "Policy Target/connect via Namespace Name" 
- *Tim Downey* upvotes "I dont want to look at 10 or 100 policies to figure out whether I have the right allow rules"
- *Anish Ramasekar* upvotes "I don't want to directly update my CIDR rules for a policy every time I add a new node or other group of IPs, which need to have policies associated with them."
- *Gobind Johar* upvotes "Name as Policy Target", "Cluster Scoped Policy" and "FQDN Policy"
- *Andrew Sy Kim* upvotes "Name as Policy Target/connect via Namespace Name"
- *Andrew Sy Kim* upvotes "Node Selector", and "Cluster Scoped Policy"
- *Jay Vyas* upvotes "Connect via namespace name"
- *Jay Vyas* upvotes "Port Ranges"
