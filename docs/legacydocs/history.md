(note this is open to PRs ! just an attempt at collecting history, which is of course subject to curation and interpretation)


# How did we get here? 

Many interesting points were brought up in the 2016-2018 time frame as the API has evolved, which give us alot of insight into the user stories we see today.

- It's intersection with Service Mesh technologies hasn't been addressed, but its clear that theres a documentation hole on wether you can implement multitenant security w networkpolicies alone vs service mesh's, and what the tradeoffs are. 
- The creation of 'exceptions', in CIDR blocks, was debated for its consitency 
- The notion of L7 has been addressed in the past as, out of scope and ideal for a CRD 
- The creation of a namespace and pod `and` filter was also seen as possibly adding confusion to the API
- At one point, a very interesting suggestion - that ALL network policies should operate at the service level, was even made !
- A 'namespace skeleton' API was also proposed (and a provisional KEP for default policies was withdrawn).
- In general API promotion from beta, which we propose, will require first getting feedback from CNI providers on wether those changes are working well or not.

# Possible conclusions to draw from these past conversations

- We should trust our instincts around *service* boundaries for policies, which many folks have suggested and clearly has been suggested in the past as well.
- We should look at L7 as a corner case which likely cannot be solved by this API
- We should consider, in v2, revising the API from scratch to have less sharp edges
- An operator or CRD which extends this API, rather then forcing changes on the broader SIG, translating 1-1 network policies, is probably a great goal for us to work towards, given that there is alot of work required to get any changes coordinated to network policies.
- We probably should be looking at operators or other sig collaboration for default policies or gateways, given that this has been discussed and decided against in the past as a sig-network owned construct

# Notable commentary, on the history of the evolution of the NetworkPolicy API

- Early commentary on namespaces as names.  From a post in https://groups.google.com/g/kubernetes-sig-network/c/GzSGt-pxBYQ/m/Rbrxve-gGgAJ , the original motivation for namespaces as name selectors was suggested in 2016 by dan.  Note this re-emerged recently again (see reference in the bottom of this timeline).

```
2) we need to clarify how namespaces are matched. I'm pretty 
uncomfortable with using labels here, since anyone can add labels at 
any time to any namespace and thus send traffic to your pod if they 
know the label you're using. If it were simply a namespace name, it's 
much easier to specifically allow only namespaces you control since you 
know your own namespace names. 
```

- Early commentary on null / empty fields, and cleaning them up from tim hockins: https://github.com/kubernetes/kubernetes/issues/51726

- June 20, 2017 Namespace by name (a contraversial topic that alot of folks want) was suggested, https://github.com/kubernetes/kubernetes/issues/47797.  This was again re-suggested recently (see end of this timeline)

- Aug, 2017, on the introduction of IPBlocks to network policies.  An interesting comment on how IPBlocks can be used to "deny" things, and how "deny" functionality wasnt in the original interest of the NetworkPolicy API.

  - the 'deny' semantic here was that this field had an `except` clause
  - namespace and pod selectors *dont* have an except clause.
```
This backdoor-introduces "deny" logic, and I worry that this turns into a firewall.
(timh)
```
- January 2, 2018: proposal to extend network policy with L7 concepts https://github.com/kubernetes/kubernetes/issues/58624 mostly decided against by the community. 
```
Perhaps in the short term it would be best to follow a CRD approach.
- e.g. Vendor/project specific CRDs that mirror the Kubernetes Network Policy structure
- but with the addition of any extensions the project supports. 
- move at its own velocity without dependency on broader community agreement. 

(alex pollit)
```

- January 15, 2018: A discussion on how egress and ipblock evolved.
  - moving from beta to GA, dan asked wether CNI implementers had any feedback on these new api constructs
  - calico folks mentioned, no major feedback , no major issue
  - An insightful comment from tim on *WHY* the network policy API is bounded to applications, and clusters, and not higher a level security tooling,  
```
  CIDR is a poor primitive for this use case, but ACK that we don't have 
  a better one just yet. NetPolicy is a cluster-centric abstraction 
  (for now?) and istio is a better fit for this because it acknowledges 
  that the problem of networking is larger than a cluster. 
(tim h)
```
- January 15, same thread as above... a VERY STRONG user story for service network policies proposed, serendipitiously, by tim:

```
(policy is specified against pods, not services) is the biggest mistake NetPolicy made. Maybe we should  EOL NetPolicy 
and rebuild it as ServiceNetPolicy. Only half joking. 
(tim h)
```
  Note that, this service policy is possible because kube-proxy by default preserves source IPs.
- Another interesting qoute on service mesh and networkpolicy integration (tim)

```
Honestly, widespread adoption of Istio (and Istio-ish things) may just 
obsolete NetworkPolicy). 
(tim h)
```
- January 22, 2018: Proposal to add The concept of pod+namespace rule.
  - `and` rules and the difficulty of combining AND vs OR rules was brought up in https://github.com/kubernetes/kubernetes/issues/58637 by dan winship and others.
  - the idea of a *generic* and was also proposed  
  - To qoute tim(h):
```
I'm concerned about how complex this becomes over time, if we add more fields here. Today it is simple - exactly one. 
If we allow one or two, but not any combination, it's awkward at best.
(tim h)
```
  This enabled the new concept of pods and namespaces, both in the SAME from block, allowing people to block / allow traffic from a particular pod in a separate namespace.
```
  - from
    - namespaceSelector: {}
    - podSelector:
        matchLabels:
          environment: production
```
- December 4, 2017, one of the first user storys for a **default policy** was created, by ahmetb
```
   I've been thinking about writing a controller that deploys default-deny-all-[ingress|egress] rules to each namespace.
(ahmetb)
```
  - an interesting response in this thread, from tim, `I think we need a more generalized "skeleton namespace" design, which could contain arbitrary policies`
- September, 2018 SCTP added to the network policy API
- November, 2018: It appears that a potential user story for denying local-host-network traffic was written by Dan winship.
```
But that then implies that a pod also always 
accepts traffic from any hostNetwork pods on the same node, which 
probably isn't desired. Does anyone do anything clever to avoid this? 
(dan w)
```
  - This is an interesting discussion because, as we know, health checks need to work from the kubelet -> pod.  But this of course automatically whitelists all hostnetwork->pod traffic.
  - nobody responded to the 'does anyone do anything clever to avoid this' question.  
  - Maybe the idea of denying local node traffic is a user story we forgot to add.
- June 4, 2018:  Another insightful comment on the boundaries of network policies made by dan winship:
```
OpenShift's multitenancy does not meet your "must not be able to see, 
hear, taste, touch, or smell each others presence" definition. It's more 
like "must not be able to spy on, eavesdrop on, or harass each other". 

I don't think anyone has asked for stronger tenant isolation, but then, 
in most OpenShift deployments the "tenants" are just different groups 
within a single company, not actually different companies. 
(dan w)
```
- June 25, 2018: Proposal to add calico globalnetworkpolicy to the K8s api was put on the mailing list https://groups.google.com/g/kubernetes-sig-network/c/pFyfse3njD4/m/vuhdqo4nAwAJ.
  - No major opposition.
  - Nobody seemed to take up the offer to add it to the core API either
- Oct 21, 2019: A KEP by vallery for default restricted network behaviour was closed, based on the idea that it isnt so uch a network policy issue, but rather a potentially different API.
- Feb 17, 2020: Namespace by Name (namespace as `name` rather then label selectors) proposed *again*  https://github.com/kubernetes/kubernetes/issues/88253.
- Nov 9, 2020: NetworkPolicy microversioning proposed by Dan winship
- Dec 22, 2020: Conversation w/ bryan boreham on why service policies dont exist yet ~ partially because of the desire to avoid biasing against iptables, which is easy to do service policies on
- March 15, 2021: Namespace as Name and Port ranges were both added to upstream Kubernetes by the Network Policy subproject.  Matt fennwicks cyclonus tool for fuzz testing identified a few issues in different CNI providers, demonstrating that the NetworkPolicy API continues to need refinement and enforcement by different CNIs.

Note most of this history comes from https://groups.google.com/g/kubernetes-sig-network/ and https://github.com/kubernetes/kubernetes/issues/ .  
There might of course be other back channel discussions or PR debates that are also relevant.
