// crdtest tests the current CRDs against example valid and invalid CRDs. There
// are two ways to use this:
//
// # Test cases
//
// Each test case takes its name from the name of the resource being
// instantiated.
//
//   - pkg/crdtest/valid are valid resource definitions.
//   - pkg/crdtest/invalid are valid resource definitions.
//
// # Running
//
// Run as a standard Go-test:
//
//	$ go test ./pkg/crdtest
//
// Run in -watch mode for development:
//
//	$ go test -c -o crdtest ./pkg/crdtest # or `make crdtest`
//	$ ./crdtest -crdDir config/crd/standard -watch pkg/crdtest/valid 2>/tmp/out
//
// will watch changes to the *.yaml files in pkg/crdtest/valid and attempt to
// load them into an API server. Validation errors will be written to stdout.
package main
