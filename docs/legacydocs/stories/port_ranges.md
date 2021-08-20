# Port Ranges/PortSet

## Owners

- @rikatz (Ricardo Katz)

## User Story

As the owner of a web application I want to allow the traffic to my Pods in different ports without having to create multiple network policies. As an example, my application is exposed in ports 80, 443 and 8443 and I want to group this in a single rule

Also, I have another application that communicates with NodePorts of a different cluster and I want to allow the egress of the traffic only the NodePort range (eg. 30000-32767) as I don't know which port is going to be allocated in the other side, but don't want to create a rule for each of them.

*Linked Issue*: https://github.com/kubernetes/kubernetes/issues/67526

## Implementation suggestion (not a user story, will be moved to the KEP later)

PortSet:

 ```
 egress:
  - ports:
    - PortSet: 30000-37267
      protocol: TCP
```

and 

```
ingress
 - ports:
   - PortSet: 80,443,56000
```

Will need some validation, like:

* Are all the Ports within a range a uint16?
* Is the specification valid? Contains only numbers, ":" and "-"
* Starts with a number, ends with a number.

