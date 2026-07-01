/*
Copyright 2025 The Kubernetes Authors.

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
		CNPBaselineTierIngressRulesEgressUnaffected,
		CNPBaselineTierEgressRulesIngressUnaffected,
	)
}

// CNPBaselineTierIngressRulesEgressUnaffected is the Baseline-tier sibling of
// CNPAdminTierIngressRulesEgressUnaffected: a Baseline-tier ClusterNetworkPolicy that
// declares only ingress rules must not affect the subject's egress ("CNPs with no
// egress rules do not affect egress traffic"). This verifies the directional
// isolation guarantee holds in the Baseline code path too. See issue
// kubernetes-sigs/network-policy-api#103.
var CNPBaselineTierIngressRulesEgressUnaffected = suite.ConformanceTest{
	ShortName:   "CNPBaselineTierIngressRulesEgressUnaffected",
	Description: "A Baseline CNP with only ingress rules must not affect the subject's egress",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
	},
	Manifests: []string{"base/baseline_tier/standard-ingress-isolation.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should enforce the ingress rule: deny ingress from slytherin to gryffindor", func(t *testing.T) {
			// Positive control: confirms the ingress-only baseline policy is active.
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
		})

		t.Run("Should not affect egress: gryffindor can still egress to ravenclaw", func(t *testing.T) {
			// ravenclaw is not referenced by the policy, so its reply path cannot
			// confound this check. With zero egress rules, this egress must be allowed.
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-ravenclaw", "luna-lovegood-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should not affect egress: gryffindor can still egress to hufflepuff", func(t *testing.T) {
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-hufflepuff", "cedric-diggory-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})
	},
}

// CNPBaselineTierEgressRulesIngressUnaffected is the Baseline-tier sibling of
// CNPAdminTierEgressRulesIngressUnaffected: a Baseline-tier ClusterNetworkPolicy that
// declares only egress rules must not affect the subject's ingress ("CNPs with no
// ingress rules do not affect ingress traffic"). See issue
// kubernetes-sigs/network-policy-api#103.
var CNPBaselineTierEgressRulesIngressUnaffected = suite.ConformanceTest{
	ShortName:   "CNPBaselineTierEgressRulesIngressUnaffected",
	Description: "A Baseline CNP with only egress rules must not affect the subject's ingress",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
	},
	Manifests: []string{"base/baseline_tier/standard-egress-isolation.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should enforce the egress rule: deny egress from gryffindor to slytherin", func(t *testing.T) {
			// Positive control: confirms the egress-only baseline policy is active.
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
		})

		t.Run("Should not affect ingress: ravenclaw can still reach gryffindor", func(t *testing.T) {
			// ravenclaw is not referenced by the policy, so its reply path cannot
			// confound this check. With zero ingress rules, this ingress must be allowed.
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should not affect ingress: hufflepuff can still reach gryffindor", func(t *testing.T) {
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})
	},
}
