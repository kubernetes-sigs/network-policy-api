# Policy Assistant (derived from Cyclonus)

Policy Assistant is a project to assist users regarding the three APIs for network policies (NetworkPolicy, AdminNetworkPolicy, BaselineAdminNetworkPolicy).

## Overview

Policy Assistant is a CLI (command-line interface) designed to help users:
1. ***Develop/understand policy configurations***.
1. ***Prevent pitfalls*** while developing policies.
1. ***Troubleshoot*** network policy issues.

Policy Assistant is a static analysis tool which ***simulates the action of network policies*** for the given traffic. ***No cluster is needed.*** Policy Assistant can read resources either from your cluster or from config files.

CNI developers may benefit from Policy Assistant as well.
Policy Assistant is capable of providing a fuzz testing framework which CNI developers could run as a second conformance profile (to ensure the CNI's implementation is compliant with API specifications).

### Roadmap

We are maintaining the project via GitHub issues.
Parent issue: [#150](https://github.com/kubernetes-sigs/network-policy-api/issues/150).

### KubeCon EU 2024

For a presentation and discussion on Policy Assistant and the admin policy APIs, see [this talk](https://youtu.be/riSv0g-TNtI?si=jiRy2mAKB0OVMFJF&t=1232).

## Installation

### Option 1: Download from Releases

*Work in progress.*

### Option 2: Make from Source

> [!NOTE]
> The CLI binary is still called "cyclonus". This will soon be renamed.

1. Clone the repo.
2. `cd cmd/policy-assistant`
3. `make cyclonus`
4. The `cyclonus` binary will be produced at *cmd/cyclonus/cyclonus*.

## Example Usage

### Analyze

> [!NOTE]
> The CLI binary is still called "cyclonus". This will soon be renamed.

#### "explain" mode

Visualize all your policies in a table.

```shell
$ ./cyclonus analyze --policy-path examples/demos/kubecon-eu-2024 --mode explain
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
> "walkthrough" mode displays more granular info in a more intuitive way.

Visualize how traffic would be allowed/denied.

```shell
$ ./cyclonus analyze --policy-path examples/demos/kubecon-eu-2024 --mode probe --probe-path examples/demos/kubecon-eu-2024/demo-probe.json
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

```shell
$ ./cyclonus analyze --policy-path examples/demos/kubecon-eu-2024 --mode walkthrough             
verdict walkthrough:
+---------------------------------------+---------+-------------------------------------------------------------+------------------------------+
|                TRAFFIC                | VERDICT |                     INGRESS WALKTHROUGH                     |      EGRESS WALKTHROUGH      |
+---------------------------------------+---------+-------------------------------------------------------------+------------------------------+
| demo/[pod=a] -> demo/[pod=b]:80 (TCP) | Allowed | [ANP] Allow (allow-80)                                      | no policies targeting egress |
+---------------------------------------+---------+-------------------------------------------------------------+                              +
| demo/[pod=a] -> demo/[pod=b]:81 (TCP) | Denied  | [ANP] Pass (pass-81) -> [BANP] Deny (baseline-deny)         |                              |
+---------------------------------------+---------+-------------------------------------------------------------+                              +
| demo/[pod=b] -> demo/[pod=a]:80 (TCP) | Allowed | [ANP] Allow (allow-80)                                      |                              |
+---------------------------------------+---------+-------------------------------------------------------------+                              +
| demo/[pod=b] -> demo/[pod=a]:81 (TCP) | Denied  | [ANP] Pass (pass-81) -> [NPv1] Dropped (demo/deny-to-pod-a) |                              |
+---------------------------------------+---------+-------------------------------------------------------------+------------------------------+
```

## Development

See [Make from Source](#option-2-make-from-source).

### Testing

Run `go test ./...` in the *cmd/policy-assistant/* directory.

Integration tests located at *test/integration/integration_test.go*.
The tests verify:

1. Building/translating Policy specs into interim data structures (matchers).
2. Simulation of expected connectivity for ANP, BANP, and v1 NetPols.
