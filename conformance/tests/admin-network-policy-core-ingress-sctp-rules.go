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
		AdminNetworkPolicyIngressSCTP,
		AdminNetworkPolicyIngressPodSelector,
	)
}

var AdminNetworkPolicyIngressPodSelector = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyIngressPodSelector",
	Description: "Tests suppot for ingress traffic (SCTP) on a pod selector using admin network policy api based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
	},
	Manifests: []string{"base/admin_network_policy/core-ingress-sctp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for the specified pod (SCTP)", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// ensure ingress is ALLOWED from dumbledore pods, ingressRule at index9 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)

			// ensure ingress is DENIED from dumbledore pods for everything but sctp
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)

		})

		t.Run("Should support an 'deny-ingress' policy for the specified pod (SCTP)", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")

			// ensure ingress is DENIED from professor-quirrel pod, ingressRule at index8 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-quirrell-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)

			// ingress is ALLOWED from professor-quirrell for all other protocols since we didn't explicitly DENY them
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-quirrell-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

	},
}

var AdminNetworkPolicyIngressSCTP = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyIngressSCTP",
	Description: "Tests support for ingress traffic (SCTP protocol) using admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
	},
	Manifests: []string{"base/admin_network_policy/core-ingress-sctp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for SCTP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure ingress is ALLOWED from gryffindor to ravenclaw
			// ingressRule at index0 will take precedence over ingressRule at index1; thus ALLOW takes precedence over DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)

			// ensure ingress is ALLOWED from dumbledore pods, ingressRule at index9 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)

			// ensure ingress is DENIED from dumbledore pods for everything but sctp
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hogwarts-staff", "professor-dumbledore-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)

		})

		t.Run("Should support an 'allow-ingress' policy for SCTP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-1 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// cedric-diggory-0 is our client pod in hufflepuff namespace
			// ensure ingress is ALLOWED from hufflepuff to ravenclaw at port 9003; ingressRule at index5 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// cedric-diggory-1 is our client pod in hufflepuff namespace
			// ensure ingress is DENIED from hufflepuff to ravenclaw for rest of the traffic; ingressRule at index6 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support an 'deny-ingress' policy for SCTP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-1 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-sctp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index0 and index1
			allowRule := anp.DeepCopy().Spec.Ingress[0]
			anp.Spec.Ingress[0] = anp.DeepCopy().Spec.Ingress[1]
			anp.Spec.Ingress[1] = allowRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure ingress is DENIED from gryffindor to ravenclaw
			// ingressRule at index0 will take precedence over ingressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'deny-ingress' policy for SCTP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// draco-malfoy-0 is our client pod in slytherin namespace
			// ensure ingress from slytherin is DENIED to ravenclaw at port 9003; ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, false)
			assert.Equal(t, true, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to ravenclaw for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)

		})

		t.Run("Should support an 'pass-ingress' policy for SCTP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-1 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-sctp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index0 and index2
			denyRule := anp.DeepCopy().Spec.Ingress[0]
			anp.Spec.Ingress[0] = anp.DeepCopy().Spec.Ingress[2]
			anp.Spec.Ingress[2] = denyRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure ingress is PASSED from gryffindor to ravenclaw
			// ingressRule at index0 will take precedence over ingressRule at index1&index2; thus PASS takes precedence over ALLOW/DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})

		t.Run("Should support a 'pass-ingress' policy for SCTP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `ingress-sctp` ANP
			// luna-lovegood-0 is our server pod in ravenclaw namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-ravenclaw",
				Name:      "luna-lovegood-0",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "ingress-sctp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			// swap rules at index3 and index4
			denyRule := anp.DeepCopy().Spec.Ingress[3]
			anp.Spec.Ingress[3] = anp.DeepCopy().Spec.Ingress[4]
			anp.Spec.Ingress[4] = denyRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// draco-malfoy-0 is our client pod in slytherin namespace
			// ensure ingress from slytherin is PASSED to ravenclaw at port 9003; ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "sctp",
				serverPod.Status.PodIP, int32(9003), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to ravenclaw for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "sctp",
				serverPod.Status.PodIP, int32(9005), s.TimeoutConfig.RequestTimeout, true)
			assert.Equal(t, true, success)
		})
	},
}
