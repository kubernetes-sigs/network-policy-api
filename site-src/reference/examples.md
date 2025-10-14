# Examples

The following contains example yamls for all of the resources which makeup the
Network Policy API.

## Sample Resources

These examples will start with the object yaml defintions used to implement the
[core use cases](../user-stories.md). Please feel
free to contribute more examples that may seem relevant to other users :-).

### Sample Spec for Story 1: Deny traffic at a cluster level

![Alt text](../images/explicit_deny.png?raw=true "Explicit Deny")

```yaml
--8<-- "user-story-examples/user-story-1.yaml"
```

### Sample Spec for Story 2: Allow traffic at a cluster level

![Alt text](../images/explicit_allow.png?raw=true "Explicit Allow")

```yaml
--8<-- "user-story-examples/user-story-2.yaml"
```

### Sample Spec for Story 3: Explicitly Delegate traffic to existing K8s Network Policy

![Alt text](../images/delegation.png?raw=true "Delegate")

```yaml
--8<-- "user-story-examples/user-story-3.yaml"
```

### Sample Spec for Story 4: Create and Isolate multiple tenants in a cluster

![Alt text](../images/tenants.png?raw=true "Tenants")

```yaml
--8<-- "user-story-examples/user-story-4-v1.yaml"
```

This can also be expressed in the following way:

```yaml
--8<-- "user-story-examples/user-story-4-v2.yaml"
```

### Sample Spec for Story 5: Cluster Wide Default Guardrails

![Alt text](../images/baseline.png?raw=true "Default Rules")

```yaml
--8<-- "user-story-examples/user-story-5.yaml"
```
