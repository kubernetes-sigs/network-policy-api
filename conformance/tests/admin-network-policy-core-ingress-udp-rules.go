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
		AdminNetworkPolicyIngressUDP,
		AdminNetworkPolicyIngressNamedPort,
		AdminNetworkPolicyIngressPodSelectorUDP,
	)
}

var AdminNetworkPolicyIngressPodSelectorUDP = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyIngressPodSelectorUDP",
	Description: "Tests support for ingress traffic (UDP protocol) targeting specific pods using admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
	},
	Manifests: []string{"base/admin_network_policy/core-ingress-udp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for UDP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-0 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// Within namespace hogwarts-staff, we target pods by their podSelector
			// ingressRule at index10 will ALLOW traffic from dumbledore pod to hufflepuff on port 53
			// ingressRule at index11 will ALLOW traffic from dumbledore pod to hufflepuff on port 5353
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)

		})

		t.Run("Should support an 'deny-ingress' policy for UDP protocol at the specified pod", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-0 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// Within namespace hogwarts-staff, we target pods by their podSelector
			// ensure ingress from professor-quirrell pod to hufflepuff is DENIED on both ports
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-quirrell-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-quirrell-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})
	},
}

var AdminNetworkPolicyIngressUDP = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyIngressUDP",
	Description: "Tests support for ingress traffic (UDP protocol) using admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
	},
	Manifests: []string{"base/admin_network_policy/core-ingress-udp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for UDP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
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
			assert.True(t, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)

		})

		t.Run("Should support an 'allow-ingress' policy for UDP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure ingress is ALLOWED from gryffindor to hufflepuff at port 53; ingressRule at index5
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// harry-potter-1 is our client pod in gryfindor namespace
			// ensure ingress is DENIED from gryffindor to hufflepuff for rest of the traffic; ingressRule at index6 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

		t.Run("Should support an 'deny-ingress' policy for UDP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-udp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index0 and index1
			allowRule := anp.DeepCopy().Spec.Ingress[0]
			anp.Spec.Ingress[0] = anp.DeepCopy().Spec.Ingress[1]
			anp.Spec.Ingress[1] = allowRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure ingress is DENIED from ravenclaw to hufflepuff
			// ingressRule at index0 will take precedence over ingressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

		t.Run("Should support a 'deny-ingress' policy for UDP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
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
			assert.True(t, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to hufflepuff for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)

		})

		t.Run("Should support an 'pass-ingress' policy for UDP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-udp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index0 and index2
			denyRule := anp.DeepCopy().Spec.Ingress[0]
			anp.Spec.Ingress[0] = anp.DeepCopy().Spec.Ingress[2]
			anp.Spec.Ingress[2] = denyRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure ingress is PASSED from ravenclaw to hufflepuff
			// ingressRule at index0 will take precedence over ingressRule at index1&index2; thus PASS takes precedence over ALLOW/DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
		})

		t.Run("Should support a 'pass-ingress' policy for UDP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-udp` ANP
			// cedric-diggory-0 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-udp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index3 and index4
			denyRule := anp.DeepCopy().Spec.Ingress[3]
			anp.Spec.Ingress[3] = anp.DeepCopy().Spec.Ingress[4]
			anp.Spec.Ingress[4] = denyRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// draco-malfoy-0 is our client pod in slytherin namespace
			// ensure ingress from slytherin is PASSED to hufflepuff at port 5353; ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to hufflepuff for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
		})
	},
}
