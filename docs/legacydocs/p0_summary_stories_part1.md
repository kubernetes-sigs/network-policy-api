This document briefly describes the 3 user stories that the Network Policy API subproject subproject has determined:
- Are relatively high in priority and have been eagerly discussed over the last several months
- Can be added to the current networking.k8s.io/v1 API group
- xref: https://docs.google.com/document/d/1_5SpZwvi9KlNMT6psrK1yu4Eo7G2FxyMuHHz-JIHd5Y/edit?usp=sharing (original brainstorming doc)
- xref: https://github.com/jayunit100/network-policy-subproject/blob/master/p0_user_stories.md (superset of all user stories proposed)

------------------------------------------------

# Proposal to Sig-network

The following user stories, which roughly map to KEPs (small) or working group tasks (medium/large) which will result in KEPs, are listed below in order of relative difficulty and scope.

## 1. Port Ranges / Port Set (Small)

**Objective**: Allow a Network Policy to contemplate a set of ports in a single rule. 

**Limitations with v1 API**
* The ‘ports’ field in ingress and egress policies is an array that needs a declaration of each single port to be contemplated. e.g. to allow range 6000-9000 except port 6379 requires each and every port 6000, 6001 … 9000 with the exception of 6379 to be set explicitly.

**Problems to be solved**
* A user wants to allow egress to all node ports of another cluster (30000 - 32768). 
* A third group of users wants to allow their Pods to communicate with a range of ports like 6000-9000, except to port 6379 which they consider insecure

-------

## 2. Select Namespace by name in a NetworkPolicy (Medium)

**Objective**: Allow a NetworkPolicy to use namespace names as the selector, instead of labels

**Limitations with v1 API**
* API supports only selecting Pods by Label (which is fine) and Namespace by labels (which is not fine, due to the problems specified below).

**Problems to be solved**
* Namespaces are often “stand alone” and may not need to be logically categorized using labels (e.g. kube-system). Referencing by name prevents unnecessary labelling of namespaces to fit into the NetworkPolicy API.
* a group of users wants to allow ingress in their Pods from any Pods in another namespace, but they don’t want to trust the label as a selector of the namespace, because per cluster RBAC any user can put their own labels in their own namespaces and become a ‘rogue’ namespace.

-------

## 3. Cluster-Scoped Network Policy (Large)

**Objective**: Enforce NetworkPolicy rules for a set of (or all) namespaces in a cluster.

**Limitations with v1 API**
* The API only supports namespaced rules, but you cannot create a rule that applies to a set of namespaces or the whole cluster.
* The API only supports expressing the intent of a “developer” role and does not capture the requirements of an “administrator” role.

**Problems to be solved**: 
* The Cluster Admins wants that all the namespaces have a ‘default deny’ rule for ingress and egress, but it’s up to the developer to open a new namespace-specific network policy
* It’s well known that DNS is a required service that any Pod in the cluster needs to reach, so a Network Policy always allowing any Pod in any namespace to reach the CoreDNS Pods is necessary.
