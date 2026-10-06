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
		CNPAdminTierPodSelector,
	)
}

var CNPAdminTierPodSelector = suite.ConformanceTest{
	ShortName:   "CNPAdminTierPodSelector",
	Description: "Tests support for pod selectors in the subject and peers of ClusterNetworkPolicy on Admin Tier, used alone and combined with namespace selectors",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
	},
	Manifests: []string{"base/admin_tier/standard-pod-selector-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support pod selectors in egress peers", func(t *testing.T) {
			// This test uses `pod-selector-rules` admin CNP whose subject selects gryffindor pods
			// by pod label alone (namespaceSelector omitted).

			// egress rule at index 0 (deny-to-ravenclaw-pods) selects ravenclaw pods by pod label alone -> DENIED
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-ravenclaw", "luna-lovegood-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// egress rule at index 1 (deny-to-cedric-diggory-0) selects a single pod by name within hufflepuff -> DENIED
			serverPod = kubernetes.GetPod(t, s.Client, "network-policy-conformance-hufflepuff", "cedric-diggory-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// cedric-diggory-1 lives in the same namespace but no rule selects it -> ALLOWED
			// (if either egress rule ignored its podSelector, this would be denied)
			serverPod = kubernetes.GetPod(t, s.Client, "network-policy-conformance-hufflepuff", "cedric-diggory-1", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should support pod selectors in ingress peers", func(t *testing.T) {
			// This test uses `pod-selector-rules` admin CNP; harry-potter-0 is our server pod in the gryffindor subject
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)

			// ingress rule at index 0 (deny-from-slytherin-pods) selects slytherin pods by pod label alone -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// ingress rule at index 1 (deny-from-cedric-diggory-0) selects a single pod by name within hufflepuff -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// cedric-diggory-1 lives in the same namespace but no rule selects it -> ALLOWED
			// (if either ingress rule ignored its podSelector, this would be denied)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-1", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should support pod selectors in the subject", func(t *testing.T) {
			// draco-malfoy-0 is not selected by the `pod-selector-rules` subject, even though its
			// namespaceSelector is omitted, so deny-to-cedric-diggory-0 must not apply to it -> ALLOWED
			// (if the subject podSelector were ignored, every pod would be a subject and this would be denied)
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-hufflepuff", "cedric-diggory-0", s.TimeoutConfig.GetTimeout)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)

			// This part uses `pod-name-subject-rules` admin CNP whose subject is the single pod luna-lovegood-0 in ravenclaw
			serverPod = kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", s.TimeoutConfig.GetTimeout)

			// luna-lovegood-0 is the subject, so egress rule at index 0 (deny-to-slytherin) applies -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)

			// luna-lovegood-1 lives in the same namespace but is not the subject -> ALLOWED
			// (if the subject podSelector were ignored, the whole namespace would be the subject and this would be denied)
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})
	},
}
