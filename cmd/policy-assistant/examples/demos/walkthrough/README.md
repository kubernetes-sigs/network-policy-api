## Use Cases

1. Test a new policy before applying it to your cluster.
2. Understand which policies are affecting traffic in your cluster.

## Overview

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

## Demo

To try for yourself:

1. Download `policy-assistant` via the [Quick Install](../../../README.md#quick-install) guide.
1. Leverage the JSON/YAML files in this folder.
1. Not required: create a Kubernetes cluster and apply any desired YAML files.

## Usage

### Specifying Policies

#### Option 1: reference policies from YAML files

Use this argument: `--policy-path <file/folder>`

#### Option 2: reference policies from cluster

Specify the `--namespace` or `--all-namespaces`.

### Specifying Pods

#### Option 1: specify single source/destination in CLI args

You can use the following arguments to reference Pods from cluster by workload name:

```bash
policy-assistant analyze --mode walkthrough \
  --src-workload demo/deployment/a \
  --dst-workload demo/pod/b \
  --port 81 \
  --protocol TCP
```

#### Option 2: specify multiple source/destination pairs in JSON

You can also reference Pods via JSON.
You can also specify Pods which are not running in a cluster in this JSON.

See the example *traffic.json* file.

```bash
policy-assistant analyze --mode walkthrough \
  --traffic-path traffic.json
```
