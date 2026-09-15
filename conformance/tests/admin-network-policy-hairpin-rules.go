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
		CNPAdminTierHairpin,
	)
}

var CNPAdminTierHairpin = suite.ConformanceTest{
	ShortName:   "CNPAdminTierHairpin",
	Description: "Tests support for ClusterNetworkPolicy enforcement on service hairpin traffic",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
		suite.SupportClusterNetworkPolicyHairpin,
	},
	Manifests: []string{"base/admin_tier/hairpin-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		t.Run("Should enforce CNP ingress deny policy on hairpin service traffic", func(t *testing.T) {
			// Get hairpin service in gryffindor namespace
			svc := kubernetes.GetService(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-hairpin-service", s.TimeoutConfig.GetTimeout)
			// harry-potter-0 connects to the Service IP (which resolves back to gryffindor pods)
			// Under CNP, this hairpin traffic is subject to ingress policy and should be denied.
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				svc.Spec.ClusterIP, int32(80), s.TimeoutConfig, false)
		})
	},
}
