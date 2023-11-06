# Policy Assistant (derived from Cyclonus)

Explains your configuration of (Baseline)AdminNetworkPolicy and v1 NetworkPolicy. Additionally, can test conformance of (B)ANP and v1 NetworkPolicy via a connectivity matrix. Derived from the great work of @mattfenwick et al. in [Cyclonus](https://github.com/mattfenwick/cyclonus).

More details here: [Cyclonus](https://github.com/mattfenwick/cyclonus).

## Usage

CLI currently under development. Will build off of `cyclonus analyze` (visualization) and `cyclonus generate` (conformance tests).

## Development

Integration tests located at *test/integration/integration_test.go*. The tests verify:

1. Building/translating Policy specs into interim data structures (matchers).
2. Simulation of expected connectivity for ANP, BANP, and v1 NetPols.
