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
		BaselineAdminNetworkPolicyIngressTCP,
		BaselineAdminNetworkPolicyIngressNamedPort,
	)
}

var BaselineAdminNetworkPolicyIngressTCP = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyIngressTCP",
	Description: "Tests support for ingress traffic (TCP protocol) using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-ingress-tcp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for TCP protocol; ensure rule ordering is respected", func(t *testing.T) {
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
			// ensure ingress is ALLOWED from ravenclaw to gryffindor
			// ingressRule at index0 will take precedence over ingressRule at index1; thus ALLOW takes precedence over DENY since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
		})

		t.Run("Should support an 'allow-ingress' policy for TCP protocol at the specified port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// harry-potter-1 is our server pod in gryffindor namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-gryffindor",
				Name:      "harry-potter-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			// cedric-diggory-0 is our client pod in hufflepuff namespace
			// ensure ingress is ALLOWED from hufflepuff to gryffindor at port 80; ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// cedric-diggory-1 is our client pod in hufflepuff namespace
			// ensure ingress is DENIED from hufflepuff to gryffindor for rest of the traffic; ingressRule at index4 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

		t.Run("Should support an 'deny-ingress' policy for TCP protocol; ensure rule ordering is respected", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// harry-potter-1 is our server pod in gryffindor namespace
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
			allowRule := banp.DeepCopy().Spec.Ingress[0]
			banp.Spec.Ingress[0] = banp.DeepCopy().Spec.Ingress[1]
			banp.Spec.Ingress[1] = allowRule
			err = s.Client.Update(ctx, banp)
			require.NoErrorf(t, err, "unable to update the baseline admin network policy")
			// luna-lovegood-0 is our client pod in ravenclaw namespace
			// ensure ingress is DENIED from ravenclaw to gryffindor
			// ingressRule at index0 will take precedence over ingressRule at index1; thus DENY takes precedence over ALLOW since rules are ordered
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
			// luna-lovegood-1 is our client pod in ravenclaw namespace
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-ravenclaw", "luna-lovegood-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

		t.Run("Should support a 'deny-ingress' policy for TCP protocol at the specified port", func(t *testing.T) {
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
			// draco-malfoy-0 is our client pod in slytherin namespace
			// ensure ingress from slytherin is DENIED to gryffindor at port 80; ingressRule at index2 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
			// draco-malfoy-1 is our client pod in slytherin namespace
			// ensure ingress from slytherin is ALLOWED to gryffindor for rest of the traffic; matches no rules hence allowed
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
		})
	},
}

var BaselineAdminNetworkPolicyIngressNamedPort = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyIngressNamedPort",
	Description: "Tests support for ingress traffic on a named port using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
		suite.SupportBaselineAdminNetworkPolicyNamedPorts,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-ingress-tcp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-ingress' policy for named port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` BANP
			// harry-potter-1 is our server pod in gryffindor namespace
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
			namedPortRule := banp.DeepCopy().Spec.Ingress[3]
			webPort := "web"
			// rewrite the tcp port 80 rule as named port rule
			namedPortRule.Ports = &[]v1alpha1.AdminNetworkPolicyPort{
				{
					NamedPort: &webPort,
				},
			}
			banp.Spec.Ingress[3] = namedPortRule
			err = s.Client.Update(ctx, banp)
			require.NoErrorf(t, err, "unable to update the baseline admin network policy")
			// cedric-diggory-0 is our client pod in hufflepuff namespace
			// ensure ingress is ALLOWED from hufflepuff to gryffindor at at the web port, which is defined as TCP at port 80 in pod spec
			// ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// cedric-diggory-1 is our client pod in hufflepuff namespace
			// ensure ingress is DENIED from hufflepuff to gryffindor for rest of the traffic; ingressRule at index4 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

	},
}
