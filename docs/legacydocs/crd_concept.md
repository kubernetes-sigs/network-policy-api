Update: 

(9/13/2020) Matt fennwick took a crack at this here, https://github.com/mattfenwick/kube-prototypes/tree/master/pkg/netpol/crd .  In the process, he found a few corner cases in types.go that specifically warrant v2 api changes.  He'll be presenting at the next sig-network or networkpolicy API group meeting.  

# An EAV model for network policys

An EAV model (entity, attribute, value) gives you the ability to natively define graphs as the implementation for policies, as opposed to other 
concepts which dont match quite as naturally to security boundaries.

Givne that several use cases suggested expanding policy APIs to cover
higher level boundaries:

- services as a boundary for network policies
- namespaces as a boundary for network policies
- namespace names as a boundary for network policies
- service names as a boundary for network policies
- port ranges 

And additionally, there were discussions that

- ingress rules are eseentially the same as egress rules from a user perspective
- policies generally should be easier to write

It makes sense to ask - how can we accomodate abitrary policy definitions?

Using an 'entity' model for pods, we can arbitrarily attach attributes to pods, and then
rely on logical matching to find pods which are allowed 

## Drawbacks

Describing network policies in this way might not cover all 100% of the Kubernete networkpolicy api cases, but
likely covers most of them.

## Initial state

Controller looks up all pods in all namespaces and isolates them.
Controller labels all namesapces uniquely.
Controller labels all pods uniquely.
Controller lables all services with the pods it targets.

```
    # nothing
```

### First app, app and db, basic allow

Controller looks up all pods in ns1.
Makes ingress rules from myapp pod -> mydb pod.
Makes egress rules from mydb pod -> myapp pod.

```
    pod1 namespace myapp
    pod2 namespace mydb

    pod2 ingress pod1 0 allow # ns1 allow traffic from anywhere in ns2
    # egress is made automatically
```

### Security firedrill: lockdown app->db access

Controller looks up all ingress rules for mypod in ns myapp, checks
what their priority is.  Deletes any which match 'deny' rules, thus
blocking db access...

```
    pod1 pod mypod
    pod1 namespaces myapp

    pod1 ingress ns1 1 deny # infinite priority
```

### Service boundaries

In this case, we replicate the orignial rule 
```
    pod1 service myapp
    pod2 service mydb

    pod2 ingress pod1 0 allow ALL # any pods under svc myapp are allowed into pod2
```

### TODO

Add support for Ports
Add support for IP selectors

-------------------------------------------------------------------------------------

# A CRD that might be easier to build these sorts of policies on

Which translates to one or many networkpolicies under the hood. 

This is a conceptual CRD, psuedo code to be iterated on in Markdown.
```
NetpolExtended {
    Spec:
        Scope:
            AllNamespaces: true/false
            AllNamespacesRegex: true/false
            IncludedNamespaces: (labels)
            // ^ Pick one
        PortSpec:
            Range: min,max
            []Ports:
            Port
            // ^ Pick one, either a 
        NodeSelectorSpec:
             matchLabels
            // will allow traffic to all nodes in this sector
        ServiceSelectorSpec:
            // will allow traffic to pods in services of this selector
            matchLabels
        NamespacesSelectorSpec:
            // will allow all traffic to all namespaces in this selector
            matchLabels
        K8sDefaultService: true
            // helps to stop you from shooting yourself in the foot by blocking egress to dns
        KubeDNSService: true
            // core dns?
   Status:
       // current 'status' to show you what your policy is *currently* doing -- think truthtables: something easy to visually grok
        List[Pod,Pod] ConnectedViaNetworkPolicy
        List[Pod,Pod] ConnectedViaClusterPolicy
}
```
