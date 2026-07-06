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

package tests

import (
	"testing"

	"sigs.k8s.io/network-policy-api/conformance/utils/kubernetes"
	"sigs.k8s.io/network-policy-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		CNPAdminTierSamePriority,
	)
}

var CNPAdminTierSamePriority = suite.ConformanceTest{
	ShortName:   "CNPAdminTierSamePriority",
	Description: "Tests that two disjoint ClusterNetworkPolicies in the Admin tier with the same priority are both applied and work as expected",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
	},
	Manifests: []string{
		"base/admin_tier/same-priority-1.yaml",
		"base/admin_tier/same-priority-2.yaml",
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// harry-potter-0 is our server pod in gryffindor namespace
		gryffindorServer := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
		// cedric-diggory-0 is our server pod in hufflepuff namespace
		hufflepuffServer := kubernetes.GetPod(t, s.Client, "network-policy-conformance-hufflepuff", "cedric-diggory-0", s.TimeoutConfig.GetTimeout)

		t.Run("Should allow traffic matching same priority disjoint policies", func(t *testing.T) {
			// Policy 1 (gryffindor) allows ravenclaw
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				gryffindorServer.Status.PodIP, int32(80), s.TimeoutConfig, true)

			// Policy 2 (hufflepuff) allows slytherin
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				hufflepuffServer.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should deny traffic not matching same priority disjoint policies", func(t *testing.T) {
			// Policy 1 (gryffindor) deny-all-else should block slytherin
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				gryffindorServer.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// Policy 2 (hufflepuff) deny-all-else should block ravenclaw
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				hufflepuffServer.Status.PodIP, int32(80), s.TimeoutConfig, false)
		})
	},
}
