## Use Cases

1. Test a new policy before applying it to your cluster.
1. Understand which policies are affecting traffic in your cluster.
1. Without a cluster, understand how policies would affect traffic for "fake" Pods.

## Overview

### Walkthrough

How do policies impact traffic?

```bash
# single source/destination read from cluster. policies read from YAML files
policy-assistant analyze --mode walkthrough \
  --policy-path policies/ \
  --src-workload demo/deployment/a \
  --dst-workload demo/pod/b \
  --port 81 \
  --protocol TCP

# multiple traffic tuples (not necessarily read from cluster). policies read from cluster
policy-assistant analyze --mode walkthrough \
  --namespace demo \
  --traffic-path traffic.json
```

Example output:

```bash
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
|                     TRAFFIC                     | VERDICT |                             INGRESS WALKTHROUGH                             |      EGRESS WALKTHROUGH      |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
| demo/deployment/a -> demo/pod/b:80 (TCP)        | Allowed | [ANP] Allow (allow-80)                                                      | no policies targeting egress |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+                              +
| demo/deployment/a -> demo/pod/b:81 (TCP)        | Denied  | [ANP] No-Op -> [BANP] Deny (baseline-deny)                                  |                              |
+-------------------------------------------------+         +-----------------------------------------------------------------------------+                              +
| demo2/[app=nginx] -> demo/deployment/a:81 (TCP) |         | [ANP] Pass (development-ns) -> [NPv1] Dropped (demo/deny-anything-to-pod-a) |                              |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
```

### Explain

You can also summarize your policies in a table:

```bash
$ policy-assistant analyze --mode explain --policy-path policies/
explained policies:
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
|  TYPE   |  SUBJECT   |            SOURCE RULES            |         PEER          |             ACTION              |      PORT/PROTOCOL       |
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
| Ingress | Namespace: | [NPv1] demo/deny-anything-to-pod-a | no peers              | NPv1:                           | none                     |
|         |    demo    |                                    |                       |    Allow any peers              |                          |
|         | Pod:       |                                    |                       |                                 |                          |
|         |    pod = a |                                    |                       |                                 |                          |
+         +------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
|         | Namespace: | [ANP] default/anp1                 | Namespace:            | BANP:                           | all ports, all protocols |
|         |    all     | [ANP] default/anp2                 |    all                |    Deny                         |                          |
|         |            | [BANP] default/default             | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+         +            +                                    +-----------------------+---------------------------------+                          +
|         |            |                                    | Namespace:            | ANP:                            |                          |
|         |            |                                    |    development = true |    pri=2 (development-ns): Pass |                          |
|         |            |                                    | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+         +            +                                    +-----------------------+---------------------------------+--------------------------+
|         |            |                                    | Namespace:            | ANP:                            | port 80 on protocol TCP  |
|         |            |                                    |    all                |    pri=1 (allow-80): Allow      |                          |
|         |            |                                    | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
```

## Demo

To try for yourself:

1. Download `policy-assistant` via the [Quick Install](../../../README.md#quick-install) guide.
1. This document references files from this folder. Feel free to clone the repo and change into this directory.

### Demo 1: Without a Cluster

You can also use `policy-assistant` with a cluster. See [Demo 2](#demo-2-using-a-cluster) for more info.

#### Specify Policies from File

Specify your policy YAML(s) with `--policy-path`.

`--mode explain` will explain the policy files:

```bash
$ policy-assistant analyze --mode explain --policy-path policies/
explained policies:
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
|  TYPE   |  SUBJECT   |            SOURCE RULES            |         PEER          |             ACTION              |      PORT/PROTOCOL       |
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
| Ingress | Namespace: | [NPv1] demo/deny-anything-to-pod-a | no peers              | NPv1:                           | none                     |
|         |    demo    |                                    |                       |    Allow any peers              |                          |
|         | Pod:       |                                    |                       |                                 |                          |
|         |    pod = a |                                    |                       |                                 |                          |
+         +------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
|         | Namespace: | [ANP] default/anp1                 | Namespace:            | BANP:                           | all ports, all protocols |
|         |    all     | [ANP] default/anp2                 |    all                |    Deny                         |                          |
|         |            | [BANP] default/default             | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+         +            +                                    +-----------------------+---------------------------------+                          +
|         |            |                                    | Namespace:            | ANP:                            |                          |
|         |            |                                    |    development = true |    pri=2 (development-ns): Pass |                          |
|         |            |                                    | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+         +            +                                    +-----------------------+---------------------------------+--------------------------+
|         |            |                                    | Namespace:            | ANP:                            | port 80 on protocol TCP  |
|         |            |                                    |    all                |    pri=1 (allow-80): Allow      |                          |
|         |            |                                    | Pod:                  |                                 |                          |
|         |            |                                    |    all                |                                 |                          |
+---------+------------+------------------------------------+-----------------------+---------------------------------+--------------------------+
```

#### Walk through Traffic ("Fake" Pods)

Now let's walk through how these policies impact traffic.

Specify your traffic in JSON like in *traffic-no-cluster.json*.

Then use `--mode walkthrough` with an extra argument for `--traffic-path`:

```bash
$ policy-assistant analyze --mode walkthrough --policy-path policies/ --traffic-path traffic-no-cluster.json
verdict walkthrough:
+--------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
|                  TRAFFIC                   | VERDICT |                             INGRESS WALKTHROUGH                             |      EGRESS WALKTHROUGH      |
+--------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
| demo/[pod=a] -> demo/[pod=b]:80 (TCP)      | Allowed | [ANP] Allow (allow-80)                                                      | no policies targeting egress |
+--------------------------------------------+---------+-----------------------------------------------------------------------------+                              +
| demo/[pod=a] -> demo/[pod=b]:81 (TCP)      | Denied  | [ANP] No-Op -> [BANP] Deny (baseline-deny)                                  |                              |
+--------------------------------------------+         +-----------------------------------------------------------------------------+                              +
| demo2/[app=nginx] -> demo/[pod=a]:81 (TCP) |         | [ANP] Pass (development-ns) -> [NPv1] Dropped (demo/deny-anything-to-pod-a) |                              |
+--------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
```

### Demo 2: Using a Cluster

Create the demo namespace:

```bash
kubectl create ns demo
```

Create deployment `a` and pod `b`:

```bash
kubectl apply -f demo-deployment-a.yaml
kubectl apply -f demo-pod-b.yaml
```

Create the policies (you can also keep referencing policies via `--policy-path` if you'd like):

```bash
# install the v0.1.1 version of AdminNetworkPolicy and BaselineAdminNetworkPolicy
wget https://github.com/kubernetes-sigs/network-policy-api/releases/download/v0.1.1/install.yaml
kubectl apply -f install.yaml

# apply policies
kubectl apply -f policies/
```

#### Specifying Policies from Cluster

You can still [specify policies from file](#specify-policies-from-file) if you'd like.

To specify policies from cluster, use `--namespace` or `--all-namespaces`.

Any AdminNetworkPolicy and BaselineAdminNetworkPolicy from the cluster will always be included since they are cluster-scoped objects.

Here's an example of `--explain mode` with `--namespace x` to get any NetworkPolicy from the `x` namespace (this namespace is not part of the demo resources actually):

```bash
$ policy-assistant analyze --mode explain -n x
explained policies:
+---------+------------+------------------------+-----------------------+---------------------------------+--------------------------+
|  TYPE   |  SUBJECT   |      SOURCE RULES      |         PEER          |             ACTION              |      PORT/PROTOCOL       |
+---------+------------+------------------------+-----------------------+---------------------------------+--------------------------+
| Ingress | Namespace: | [NPv1] x/base          | Namespace:            | NPv1:                           | port 80 on protocol TCP  |
|         |    x       |                        |    ns In [x y]        |    Allow any peers              |                          |
|         | Pod:       |                        | Pod:                  |                                 |                          |
|         |    pod = a |                        |    pod In [b c]       |                                 |                          |
+         +------------+------------------------+-----------------------+---------------------------------+--------------------------+
|         | Namespace: | [ANP] default/anp1     | Namespace:            | BANP:                           | all ports, all protocols |
|         |    all     | [ANP] default/anp2     |    all                |    Deny                         |                          |
|         |            | [BANP] default/default | Pod:                  |                                 |                          |
|         |            |                        |    all                |                                 |                          |
+         +            +                        +-----------------------+---------------------------------+                          +
|         |            |                        | Namespace:            | ANP:                            |                          |
|         |            |                        |    development = true |    pri=2 (development-ns): Pass |                          |
|         |            |                        | Pod:                  |                                 |                          |
|         |            |                        |    all                |                                 |                          |
+         +            +                        +-----------------------+---------------------------------+--------------------------+
|         |            |                        | Namespace:            | ANP:                            | port 80 on protocol TCP  |
|         |            |                        |    all                |    pri=1 (allow-80): Allow      |                          |
|         |            |                        | Pod:                  |                                 |                          |
|         |            |                        |    all                |                                 |                          |
+---------+------------+------------------------+-----------------------+---------------------------------+--------------------------+
|         |            |                        |                       |                                 |                          |
+---------+------------+------------------------+-----------------------+---------------------------------+--------------------------+
| Egress  | Namespace: | [NPv1] x/base          | 10.224.1.0/24         | NPv1:                           | port 80 on protocol TCP  |
|         |    x       |                        | except []             |    Allow any peers              |                          |
|         | Pod:       |                        |                       |                                 |                          |
|         |    pod = a |                        |                       |                                 |                          |
+         +            +                        +-----------------------+                                 +--------------------------+
|         |            |                        | all pods, all ips     |                                 | port 53 on protocol UDP  |
|         |            |                        |                       |                                 | port 53 on protocol TCP  |
|         |            |                        |                       |                                 |                          |
|         |            |                        |                       |                                 |                          |
+---------+------------+------------------------+-----------------------+---------------------------------+--------------------------+
```

#### Walk Through Traffic (Cluster Pods)

Reference Pods from cluster by workload name such as `demo/deployment/a` or `kube-system/daemonset/kube-proxy` etc.

You can still [specify traffic for "fake" Pods](#walk-through-traffic-fake-pods) if you'd like.

#### Option 1: single source/destination in CLI args

You can specify workloads, port, and protocol via the CLI:

```bash
$ policy-assistant analyze --mode walkthrough --all-namespaces \
  --src-workload demo/deployment/a \
  --dst-workload demo/pod/b \
  --port 81 \
  --protocol TCP
verdict walkthrough:
+------------------------------------------+---------+--------------------------------------------+------------------------------+
|                 TRAFFIC                  | VERDICT |            INGRESS WALKTHROUGH             |      EGRESS WALKTHROUGH      |
+------------------------------------------+---------+--------------------------------------------+------------------------------+
| demo/deployment/a -> demo/pod/b:81 (TCP) | Denied  | [ANP] No-Op -> [BANP] Deny (baseline-deny) | no policies targeting egress |
+------------------------------------------+---------+--------------------------------------------+------------------------------+
```

This example uses policies from cluster, but you could also [specify policies from file](#specify-policies-from-file).

#### Option 2: multiple source/destination pairs in JSON

Specify source/destination workload names in your traffic like in *traffic.json*.

Notice how you can still specify "fake" sources/destinations like above.

*traffic.json* has an example of mix and matching a "fake" source (nginx) and a destination workload from the cluster.

```bash
verdict walkthrough:
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
|                     TRAFFIC                     | VERDICT |                             INGRESS WALKTHROUGH                             |      EGRESS WALKTHROUGH      |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
| demo/deployment/a -> demo/pod/b:80 (TCP)        | Allowed | [ANP] Allow (allow-80)                                                      | no policies targeting egress |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+                              +
| demo/deployment/a -> demo/pod/b:81 (TCP)        | Denied  | [ANP] No-Op -> [BANP] Deny (baseline-deny)                                  |                              |
+-------------------------------------------------+         +-----------------------------------------------------------------------------+                              +
| demo2/[app=nginx] -> demo/deployment/a:81 (TCP) |         | [ANP] Pass (development-ns) -> [NPv1] Dropped (demo/deny-anything-to-pod-a) |                              |
+-------------------------------------------------+---------+-----------------------------------------------------------------------------+------------------------------+
```

Again, this example uses policies from cluster, but you could also [specify policies from file](#specify-policies-from-file).
