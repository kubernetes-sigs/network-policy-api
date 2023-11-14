/*
Copyright 2023 The Kubernetes Authors.

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

var AdminNetworkPolicyEgressNamedPort = suite.ConformanceTest{
	ShortName:   "AdminNetworkPolicyEgressNamedPort",
	Description: "Tests support for egress traffic on a named port using admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportAdminNetworkPolicy,
		suite.SupportAdminNetworkPolicyNamedPorts,
	},
	Manifests: []string{"base/admin_network_policy/core-egress-tcp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-egress' policy for named port", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `egress-tcp` ANP
			// cedric-diggory-1 is our server pod in hufflepuff namespace
			serverPod := &v1.Pod{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-hufflepuff",
				Name:      "cedric-diggory-1",
			}, serverPod)
			require.NoErrorf(t, err, "unable to fetch the server pod")
			anp := &v1alpha1.AdminNetworkPolicy{}
			err = s.Client.Get(ctx, client.ObjectKey{
				Name: "egress-tcp",
			}, anp)
			require.NoErrorf(t, err, "unable to fetch the admin network policy")
			namedPortRule := anp.DeepCopy().Spec.Egress[5]
			webPort := "web"
			// replace the tcp port 8080 rule as named port rule which translate to tcp port 80 instead
			namedPortRule.Ports = &[]v1alpha1.AdminNetworkPolicyPort{
				{
					NamedPort: &webPort,
				},
			}
			anp.Spec.Egress[5] = namedPortRule
			err = s.Client.Update(ctx, anp)
			require.NoErrorf(t, err, "unable to update the admin network policy")
			// harry-potter-0 is our client pod in gryffindor namespace
			// ensure egress is ALLOWED to hufflepuff from gryffindor at the web port, which is defined as TCP at port 80 in pod spec
			// egressRule at index5 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// harry-potter-1 is our client pod in gryffindor namespace
			// ensure egress is DENIED to hufflepuff from gryffindor for rest of the traffic; egressRule at index6 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

	},
}
