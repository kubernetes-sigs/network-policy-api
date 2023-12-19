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
		BaselineAdminNetworkPolicyIngressUDP,
		BaselineAdminNetworkPolicyIngressPodSelectorUDP,
	)
}

var BaselineAdminNetworkPolicyIngressPodSelectorUDP = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyIngressPodSelectorUDP",
	Description: "Tests support for ingress traffic (UDP protocol) at the specific targeted pod using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-ingress-udp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		t.Run("Should support an 'allow-ingress' policy for UDP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-ingress' policy for UDP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-quirrell-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})
	},
}

var BaselineAdminNetworkPolicyIngressUDP = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyIngressUDP",
	Description: "Tests support for ingress traffic (UDP protocol) using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-ingress-udp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for UDP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// cedric-diggory-0 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure ingress is ALLOWED from ravenclaw to hufflepuff
			// ingressRule at index0 will take precedence over ingressRule at index1; thus ALLOW takes precedence over DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'allow-ingress' policy for UDP protocol at the specified port", func(t *testing.T) {
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
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure ingress is ALLOWED from gryffindor to hufflepuff at port 53; ingressRule at index5 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryfindor namespace
			// ensure ingress is DENIED from gryffindor to hufflepuff for rest of the traffic; ingressRule at index6 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'deny-ingress' policy for UDP protocol; ensure rule ordering is respected", func(t *testing.T) {
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
			banp := &v1alpha1.BaselineAdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "default",
			}, banp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index0 and index1
			allowRule := banp.DeepCopy().Spec.Ingress[0]
			banp.Spec.Ingress[0] = banp.DeepCopy().Spec.Ingress[1]
			banp.Spec.Ingress[1] = allowRule
			err = s.Client.Update(ctx, banp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure ingress is DENIED from ravenclaw to hufflepuff
			// ingressRule at index0 will take precedence over ingressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-ingress' policy for UDP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// cedric-diggory-0 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// draco-malfoy-0 is our client pod in slytherin namespace
			// ensure ingress from slytherin is DENIED to hufflepuff at port 80; ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to hufflepuff for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})
	},
}
