# Examples

The following contains example yamls for all of the resources which makeup the
Network Policy API.

## Sample AdminNetworkPolicy and BaseLineAdminNetworkPolicy Resources

These examples will start with the object yaml defintions used to implement the
[core use cases](../intro.md#adminnetworkpolicy-api-user-stories). Please feel
free to contribute more examples that may seem relevant to other users :-).

### Sample Spec for Story 1: Deny traffic at a cluster level

![Alt text](../../images/explicit_deny.png?raw=true "Explicit Deny")

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: cluster-wide-deny-example
spec:
  priority: 10
  subject:
    namespaces:
      matchLabels:
        kubernetes.io/metadata.name: sensitive-ns
  ingress:
    - action: Deny
      from:
      - namespaces:
         namespaceSelector: {}
      name: select-all-deny-all
```

### Sample Spec for Story 2: Allow traffic at a cluster level

![Alt text](../../images/explicit_allow.png?raw=true "Explicit Allow")

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: cluster-wide-allow-example
spec:
  priority: 30
  subject:
    namespaces: {}
  ingress:
    - action: Allow
      from:
      - namespaces:
          namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: monitoring-ns
  egress:
    - action: Allow
      to:
      - namespaces:
          namespaceSelector:
            matchlabels:
              kubernetes.io/metadata.name: kube-system
        pods:   
          podSelector:
            matchlabels:
              app: kube-dns
```

### Story 3: Explicitly Delegate traffic to existing K8s Network Policy

![Alt text](../../images/delegation.png?raw=true "Delegate")

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: pub-svc-delegate-example
spec:
  priority: 20
  subject:
    namespaces: {}
  egress:
  - action: Pass
    to:
    - namespaces:
        namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: bar-ns-1
      pods:
        podSelector:
          matchLabels:
            app: svc-pub
    ports:
      - portNumber: 
          protocol: TCP
          port: 8080
```

### Story 4: Create and Isolate multiple tenants in a cluster  

![Alt text](../../images/tenants.png?raw=true "Tenants")

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: tenant-creation-example
spec:
  priority: 50
  subject:
    namespaces:
      matchExpressions: {key: "tenant"; operator: Exists}
  ingress:
    - action: Deny
      from:
      - namespaces:
          notSameLabels:
          - tenant
```

This can also be expressed in the following way: 

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: tenant-creation-example
spec:
  priority: 50
  subject:
    namespaces:
      matchExpressions: {key: "tenant"; operator: Exists}
  ingress:
    - action: Pass # Pass inter-tenant traffic to any defined NetworkPolicies
      from:
      - namespaces:
          sameLabels:
          - tenant
    - action: Deny   # Deny everything else other than same tenant traffic
      from:
      - namespaces:
          namespaceSelector: {}
```

### Story 5: Cluster Wide Default Guardrails

![Alt text](../../images/baseline.png?raw=true "Default Rules")

```yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: BaselineAdminNetworkPolicy
metadata:
  name: baseline-rule-example
spec:
  subject:
    namespaces: {}
  ingress:
    - action: Deny   # zero-trust cluster default security posture
      from:
      - namespaces:
          namespaceSelector: {}
```
