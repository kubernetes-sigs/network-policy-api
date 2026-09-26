# AdminNetworkPolicy recipes

These recipes use the policy.networking.k8s.io/v1alpha1
AdminNetworkPolicy and BaselineAdminNetworkPolicy resources. They are
cluster-scoped and require an implementation that supports the
AdminNetworkPolicy API. The current project is developing ClusterNetworkPolicy
as the successor API; for new deployments, compare these examples with the
[current API overview](api-overview.md) and your implementation's support
matrix.

Install the matching v1alpha1 CRDs before applying these examples. Replace
the namespace and label values with the conventions used by your cluster.
Always test a policy in a non-production environment first.

## Recipe 1: Allow essential monitoring and DNS traffic

Use a high-precedence AdminNetworkPolicy for traffic that must remain
available even when namespace owners add or change NetworkPolicy objects. The
empty namespace selector in subject selects all namespaces.

~~~yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: cluster-essentials
spec:
  priority: 10
  subject:
    namespaces: {}
  ingress:
    - name: allow-monitoring
      action: Allow
      from:
        - namespaces:
            matchLabels:
              kubernetes.io/metadata.name: monitoring
      ports:
        - portNumber:
            protocol: TCP
            port: 9090
  egress:
    - name: allow-dns-udp
      action: Allow
      to:
        - pods:
            namespaceSelector:
              matchLabels:
                kubernetes.io/metadata.name: kube-system
            podSelector:
              matchLabels:
                k8s-app: kube-dns
      ports:
        - portNumber:
            protocol: UDP
            port: 53
    - name: allow-dns-tcp
      action: Allow
      to:
        - pods:
            namespaceSelector:
              matchLabels:
                kubernetes.io/metadata.name: kube-system
            podSelector:
              matchLabels:
                k8s-app: kube-dns
      ports:
        - portNumber:
            protocol: TCP
            port: 53
~~~

The monitoring namespace and the DNS pod label vary by cluster. Update them
before applying the object.

## Recipe 2: Provide a baseline deny guardrail

BaselineAdminNetworkPolicy is evaluated after higher-precedence
AdminNetworkPolicy and namespace NetworkPolicy rules. A default policy named
default can provide a deny-by-default fallback while still allowing namespace
owners to define narrower NetworkPolicy rules.

~~~yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: BaselineAdminNetworkPolicy
metadata:
  name: default
spec:
  subject:
    namespaces: {}
  ingress:
    - name: deny-unspecified-ingress
      action: Deny
      from:
        - namespaces: {}
  egress:
    - name: deny-unspecified-egress
      action: Deny
      to:
        - namespaces: {}
~~~

Keep essential AdminNetworkPolicy Allow rules at a numerically lower priority
than any other AdminNetworkPolicy rules. Verify DNS, health checks,
control-plane access, and required external egress before enabling a baseline
deny rule.

## Recipe 3: Delegate selected traffic to namespace owners

Use the Pass action when administrators want a namespace's NetworkPolicy
objects to make the final decision for a specific flow. This example
delegates TCP traffic to port 8080 for the public-api pods in payments:

~~~yaml
apiVersion: policy.networking.k8s.io/v1alpha1
kind: AdminNetworkPolicy
metadata:
  name: delegate-public-api
spec:
  priority: 20
  subject:
    namespaces: {}
  egress:
    - name: delegate-public-api
      action: Pass
      to:
        - pods:
            namespaceSelector:
              matchLabels:
                kubernetes.io/metadata.name: payments
            podSelector:
              matchLabels:
                app: public-api
      ports:
        - portNumber:
            protocol: TCP
            port: 8080
~~~

The selected flow continues to NetworkPolicy evaluation. If no NetworkPolicy
applies, it can continue to the baseline policy. A Pass rule does not grant
access by itself, so pair it with an explicit namespace-scoped allow rule when
access is required:

~~~yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-client-to-public-api
  namespace: payments
spec:
  podSelector:
    matchLabels:
      app: public-api
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: clients
      ports:
        - protocol: TCP
          port: 8080
~~~

## Validate a recipe

After installing the CRDs, apply one recipe at a time and inspect the object:

~~~sh
kubectl apply --dry-run=server -f recipe.yaml
kubectl get adminnetworkpolicy
kubectl get baselineadminnetworkpolicy default -o yaml
~~~

Then test both a connection that should be allowed and one that should be
denied. Confirm the implementation's policy-status conditions and logs before
moving to the next recipe.
