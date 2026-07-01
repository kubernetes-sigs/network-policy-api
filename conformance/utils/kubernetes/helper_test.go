/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/network-policy-api/conformance/utils/config"
)

// TestPokeServerDryRun verifies that, in dry-run mode, PokeServer short-circuits
// before contacting the dataplane. It is deliberately called with a nil client
// and nil kubeConfig: if the probe were actually attempted it would panic or
// fail, so a passing subtest proves no dataplane access occurred.
func TestPokeServerDryRun(t *testing.T) {
	SetDryRun(true)
	defer SetDryRun(false)

	ok := t.Run("probe is a no-op in dry-run mode", func(st *testing.T) {
		PokeServer(st, nil, nil, "client-ns", "client-pod", "tcp", "192.0.2.1", 80, config.TimeoutConfig{}, false)
	})
	assert.True(t, ok, "dry-run PokeServer must not fail or contact the dataplane")
}
