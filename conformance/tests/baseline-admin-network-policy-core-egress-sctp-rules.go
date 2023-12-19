/*
Copyright 2022 The Kubernetes Authors.

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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"sigs.k8s.io/network-policy-api/conformance/utils/kubernetes"
	"sigs.k8s.io/network-policy-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		BaselineAdminNetworkPolicyEgressSCTP,
		BaselineAdminNetworkPolicyEgressPodSelectorSCTP,
	)
}

var BaselineAdminNetworkPolicyEgressPodSelectorSCTP = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyEgressPodSelectorSCTP",
	Description: "Tests support for egress traffic (SCTP protocol) at specific pods using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-egress-sctp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-egress' policy for SCTP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()

			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hogwarts-staff",
				Name:      "professor-dumbledore-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// ensure egress is ALLOWED to Professor Dumbledore from ravenclaw
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-egress' policy for SCTP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()

			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hogwarts-staff",
				Name:      "professor-quirrell-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// ensure egress is DENIED to professor quirrell from ravenclaw
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})
	},
}

var BaselineAdminNetworkPolicyEgressSCTP = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyEgressSCTP",
	Description: "Tests support for egress traffic (SCTP protocol) using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-egress-sctp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-egress' policy for SCTP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// harry-potter-0 is our server pod in gryffindor namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-gryffindor",
				Name:      "harry-potter-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure egress is ALLOWED to gryffindor from ravenclaw
			// egressRule at index0 will take precedence over egressRule at index1; thus ALLOW takes precedence over DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'allow-egress' policy for SCTP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure egress is ALLOWED to hufflepuff from ravenclaw at port 9003; egressRule at index5 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			// ensure egress is DENIED to hufflepuff from ravenclaw for rest of the traffic; egressRule at index6 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'deny-egress' policy for SCTP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// harry-potter-0 is our server pod in gryffindor namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-gryffindor",
				Name:      "harry-potter-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			banp := &v1alpha1.BaselineAdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "default",
			}, banp)
			require.NoErrorf(t, err, "unable to fetch the baseline admin network policy")
			// swap rules at index0 and index1
			allowRule := banp.DeepCopy().Spec.Egress[0]
			banp.Spec.Egress[0] = banp.DeepCopy().Spec.Egress[1]
			banp.Spec.Egress[1] = allowRule
			err = s.Client.Update(ctx, banp)
			require.NoErrorf(t, err, "unable to update the baseline admin network policy")
			// luna-lovegood-0 is our client pod in gryffindor namespace
			// ensure egress is DENIED to gryffindor from ravenclaw
			// egressRule at index0 will take precedence over egressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-egress' policy for SCTP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// draco-malfoy-0 is our server pod in slytherin namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-slytherin",
				Name:      "draco-malfoy-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure egress to slytherin is DENIED from ravenclaw at port 9003; egressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			// ensure egress to slytherin is ALLOWED from ravenclaw for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})
	},
}
