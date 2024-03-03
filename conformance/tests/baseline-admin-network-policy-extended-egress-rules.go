/*
Copyright 2024 The Kubernetes Authors.

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
		BaselineAdminNetworkPolicyEgressNamedPort,
	)
}

var BaselineAdminNetworkPolicyEgressNamedPort = suite.ConformanceTest{
	ShortName:   "BaselineAdminNetworkPolicyEgressNamedPort",
	Description: "Tests support for egress traffic on a named port using baseline admin network policy API based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportBaselineAdminNetworkPolicy,
		suite.SupportBaselineAdminNetworkPolicyNamedPorts,
	},
	Manifests: []string{"base/baseline_admin_network_policy/core-egress-udp-rules.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		t.Run("Should support an 'allow-egress' policy for named port", func(t *testing.T) {
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
			dnsPortRule := banp.DeepCopy().Spec.Egress[3]
			dnsPort := "dns"
			// rewrite the udp port 53 rule as named port rule
			dnsPortRule.Ports = &[]v1alpha1.AdminNetworkPolicyPort{
				{
					NamedPort: &dnsPort,
				},
			}
			banp.Spec.Egress[3] = dnsPortRule
			err = s.Client.Update(ctx, banp)
			require.NoErrorf(t, err, "unable to update the baseline admin network policy")
			// cedric-diggory-0 is our client pod in hufflepuff namespace
			// ensure egress is ALLOWED to gryffindor from hufflepuff at the dns port, which is defined as UDP at port 53 in pod spec
			// modified ingressRule at index3 should take effect
			success := kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-0", "udp",
				serverPod.Status.PodIP, int32(53), s.TimeoutConfig.RequestTimeout, true)
			assert.True(t, success)
			// cedric-diggory-1 is our client pod in hufflepuff namespace
			// ensure egress is DENIED to gryffindor from hufflepuff for rest of the traffic; egressRule at index4 should take effect
			success = kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-hufflepuff", "cedric-diggory-1", "udp",
				serverPod.Status.PodIP, int32(5353), s.TimeoutConfig.RequestTimeout, false)
			assert.True(t, success)
		})

	},
}
