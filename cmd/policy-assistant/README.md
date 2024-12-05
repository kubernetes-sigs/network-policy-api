# Policy Assistant (derived from Cyclonus)

Policy Assistant is a project to assist users regarding all APIs for network policies.
Currently, the APIs are:

- [NetworkPolicy (v1)](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [AdminNetworkPolicy and BaselineAdminNetworkPolicy](https://network-policy-api.sigs.k8s.io/api-overview/)

## Overview

Policy Assistant is a CLI (command-line interface) designed to help users:
1. ***Develop/understand policy configurations***.
1. ***Prevent pitfalls*** while developing policies.
1. ***Troubleshoot*** network policy issues.

Policy Assistant is a static analysis tool which ***simulates the action of network policies*** for the given traffic. Policy Assistant can read resources either from your cluster or from config files, so no cluster is needed.

For instance, Policy Assistant can simulate and walk through which policies impact cluster traffic:

```shell
$ policy-assistant analyze --mode walkthrough --policy-path policies/ --traffic-path traffic.json
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

### Quick Install

Download the latest `policy-assistant` release either from GitHub ([web page](https://github.com/kubernetes-sigs/network-policy-api/releases/v0.0.1-policy-assistant)) or via these bash commands:

```bash
curl -LO https://github.com/kubernetes-sigs/network-policy-api/releases/download/v0.0.1-policy-assistant/policy-assistant_linux_amd64.tar.gz
# optionally verify check sum
tar -xvf policy-assistant_linux_amd64.tar.gz
sudo mv policy-assistant_linux_amd64/policy-assistant /usr/bin/policy-assistant
policy-assistant analyze --help
```

Alternatively, [install from source](#make-from-source).

See [example usage](#example-usage) below.

### Fuzz Testing Capability

CNI developers may benefit from Policy Assistant as well.
Policy Assistant is capable of providing a fuzz testing framework (see [#154](https://github.com/kubernetes-sigs/network-policy-api/issues/154)) which CNI developers could run as a second conformance profile (to ensure the CNI's implementation is compliant with API specifications).

### Roadmap

Planning is currently via GitHub issues.

- Original issue for Policy Assistant: [#150](https://github.com/kubernetes-sigs/network-policy-api/issues/150).
- First CLI release: [#255](https://github.com/kubernetes-sigs/network-policy-api/issues/255)

### KubeCon EU 2024

For a presentation and discussion on Policy Assistant and the admin policy APIs, see [this talk](https://youtu.be/riSv0g-TNtI?si=jiRy2mAKB0OVMFJF&t=1232).

## Example Usage

### Analyze

#### "explain" mode

Visualize all your policies in a table.

```shell
$ policy-assistant analyze --mode explain --policy-path cmd/policy-assistant/examples/demos/kubecon-eu-2024/policies/
explained policies:
+---------+---------------------------------------+---------------------------+------------+----------------------------+--------------------------+
|  TYPE   |                SUBJECT                |       SOURCE RULES        |    PEER    |           ACTION           |      PORT/PROTOCOL       |
+---------+---------------------------------------+---------------------------+------------+----------------------------+--------------------------+
| Ingress | Namespace:                            | [NPv1] demo/deny-to-pod-a | no peers   | NPv1:                      | none                     |
|         |    demo                               |                           |            |    Allow any peers         |                          |
|         | Pod:                                  |                           |            |                            |                          |
|         |    pod = a                            |                           |            |                            |                          |
+         +---------------------------------------+---------------------------+------------+----------------------------+--------------------------+
|         | Namespace:                            | [ANP] default/anp1        | Namespace: | BANP:                      | all ports, all protocols |
|         |    kubernetes.io/metadata.name = demo | [ANP] default/anp2        |    all     |    Deny                    |                          |
|         |                                       | [ANP] default/anp3        | Pod:       |                            |                          |
|         |                                       | [BANP] default/default    |    all     |                            |                          |
+         +                                       +                           +            +----------------------------+--------------------------+
|         |                                       |                           |            | ANP:                       | port 80 on protocol TCP  |
|         |                                       |                           |            |    pri=1 (allow-80): Allow |                          |
|         |                                       |                           |            |                            |                          |
|         |                                       |                           |            |                            |                          |
+         +                                       +                           +            +----------------------------+--------------------------+
|         |                                       |                           |            | ANP:                       | port 81 on protocol TCP  |
|         |                                       |                           |            |    pri=2 (pass-81): Pass   |                          |
|         |                                       |                           |            |    pri=3 (deny-81): Deny   |                          |
|         |                                       |                           |            |                            |                          |
+---------+---------------------------------------+---------------------------+------------+----------------------------+--------------------------+
```

#### "probe" mode

> [!NOTE]
> "walkthrough" mode is more intuitive and informative than "probe" mode.

Visualize how traffic would be allowed/denied.

```shell
$ policy-assistant analyze --mode probe --probe-path examples/demos/kubecon-eu-2024/demo-probe.json --policy-path cmd/policy-assistant/examples/demos/kubecon-eu-2024/policies/
probe (simulated connectivity):
INFO[2024-08-07T17:26:28-07:00] probe on port 80, protocol TCP               
Ingress:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | X      | #      |
+--------+--------+--------+

Egress:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | .      | #      |
+--------+--------+--------+

Combined:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | X      | #      |
+--------+--------+--------+



INFO[2024-08-07T17:26:28-07:00] probe on port 81, protocol TCP               
Ingress:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | X      | #      |
+--------+--------+--------+

Egress:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | .      | #      |
+--------+--------+--------+

Combined:
+--------+--------+--------+
|        | DEMO/A | DEMO/B |
+--------+--------+--------+
| demo/a | #      | .      |
| demo/b | X      | #      |
+--------+--------+--------+
```

#### "walkthrough" mode

Visualize how traffic would be allowed/denied and which policies are causing the verdict.

For more information, see the [walkthrough demo](./examples/demos/walkthrough/README.md).

```shell
$ policy-assistant analyze --mode walkthrough --policy-path policies/ --traffic-path traffic.json
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

## Development

### Make from Source

1. Clone the repo.
2. `cd cmd/policy-assistant`
3. `make policy-assistant`
4. The `policy-assistant` binary will be produced at *cmd/policy-assistant/policy-assistant*.

### Testing

Run `go test ./...` in the *cmd/policy-assistant/* directory.

Integration tests located at *test/integration/integration_test.go*.
The tests verify:

1. Building/translating Policy specs into interim data structures (matchers).
2. Simulation of expected connectivity for ANP, BANP, and v1 NetPols.

#### GitHub Action

PRs must pass the GitHub Action for Policy Assistant specified under *.github/*.
