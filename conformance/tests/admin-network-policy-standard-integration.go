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

	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "sigs.k8s.io/network-policy-api/apis/v1beta1"
	"sigs.k8s.io/network-policy-api/conformance/utils/kubernetes"
	"sigs.k8s.io/network-policy-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		CNPAdminTierIntegration,
	)
}

var CNPAdminTierIntegration = suite.ConformanceTest{
	ShortName:   "CNPAdminTierIntegration",
	Description: "Tests integration support for gress traffic between admin CNP, NP and baseline CNP using pod selectors with PASS and ACCEPT actions based on a server and client model",
	Features: []suite.SupportedFeature{
		suite.SupportClusterNetworkPolicy,
	},
	Manifests: []string{"base/api_integration/standard-anp-np-banp.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// The policies in api_integration/standard-anp-np-banp.yaml single out two pods by the conformance-role label:
		// draco-malfoy-0 ('delegated') is the slytherin pod selected by the admin CNP rules at index 0, which start as
		// "deny" and are updated to "pass" below, while draco-malfoy-1 is left to the "accept" rules at index 1;
		// harry-potter-0 ('isolated') is the only pod selected by the network policy, so harry-potter-1 is left to
		// the baseline CNP.
		// The base pods are shared by every test, so always strip the labels added below.
		t.Cleanup(func() {
			kubernetes.RemovePodLabel(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", conformanceRoleLabel, s.TimeoutConfig.GetTimeout)
		})
		t.Cleanup(func() {
			kubernetes.RemovePodLabel(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", conformanceRoleLabel, s.TimeoutConfig.GetTimeout)
		})
		kubernetes.SetPodLabel(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", conformanceRoleLabel, "delegated", s.TimeoutConfig.GetTimeout)
		kubernetes.SetPodLabel(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", conformanceRoleLabel, "isolated", s.TimeoutConfig.GetTimeout)

		t.Run("Should Deny traffic from slytherin to gryffindor respecting admin CNP", func(t *testing.T) {
			// This test uses `pass-example` admin CNP from api_integration/standard-anp-np-banp.yaml
			// harry-potter-0 is our server pod in gryffindor namespace, selected by the network policy
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			// ingress rule at index 0 (deny-ingress-from-delegated-slytherin-pods) selects draco-malfoy-0 by pod label
			// and takes effect ahead of the network policy, which allows TCP port 80 -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
			// harry-potter-1 is our server pod in gryffindor namespace, not selected by the network policy
			serverPod = kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-1", s.TimeoutConfig.GetTimeout)
			// ingress rule at index 0 applies to every gryffindor pod -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig, false)
		})

		t.Run("Should Deny traffic to slytherin from gryffindor respecting admin CNP", func(t *testing.T) {
			// This test uses `pass-example` admin CNP from api_integration/standard-anp-np-banp.yaml
			// draco-malfoy-0 is our server pod in slytherin namespace
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", s.TimeoutConfig.GetTimeout)
			// egress rule at index 0 (deny-egress-to-delegated-slytherin-pods) selects draco-malfoy-0 by pod label
			// and takes effect ahead of the network policy, which allows TCP port 80 for harry-potter-0 -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
			// egress rule at index 0 applies to every gryffindor pod, harry-potter-1 included -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig, false)
		})

		t.Run("Should support a 'pass-ingress' policy for admin CNP and respect the match for network policy", func(t *testing.T) {
			// This test uses `pass example` admin CNP from api_integration/standard-anp-np-banp.yaml
			// and alters the ingress rule action to "pass"
			cnp := kubernetes.GetClusterNetworkPolicy(t, s.Client, "pass-example", s.TimeoutConfig.GetTimeout)
			mutate := cnp.DeepCopy()
			// change ingress rule from "deny" to "pass"
			mutate.Spec.Ingress[0].Action = api.ClusterNetworkPolicyRuleActionPass
			kubernetes.PatchClusterNetworkPolicy(t, s.Client, cnp, mutate, s.TimeoutConfig.GetTimeout)

			// harry-potter-0 is our server pod in gryffindor namespace, selected by the network policy
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			// ingress rule at index 0 passes traffic from draco-malfoy-0 on to the network policy, which allows TCP port 80 -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
			// the network policy does not allow TCP port 8080, so harry-potter-0 is isolated on it -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig, false)
			// ingress rule at index 1 (accept-all-ingress-from-slytherin) accepts traffic from draco-malfoy-1, so the
			// network policy is not consulted -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "tcp",
				serverPod.Status.PodIP, int32(8080), s.TimeoutConfig, true)

			// harry-potter-1 is our server pod in gryffindor namespace, not selected by the network policy
			serverPod = kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-1", s.TimeoutConfig.GetTimeout)
			// ingress rule at index 0 passes traffic from draco-malfoy-0 and no network policy selects harry-potter-1, so
			// the `default` baseline CNP ingress rule at index 0 (deny-all-ingress-from-slytherin) applies -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
			// ingress rule at index 1 accepts traffic from draco-malfoy-1, so the baseline CNP is not consulted -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-1", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should support a 'pass-egress' policy for admin CNP and respect the match for network policy", func(t *testing.T) {
			// This test uses `pass example` admin CNP from api_integration/standard-anp-np-banp.yaml
			// and alters the egress rule action to "pass"
			cnp := kubernetes.GetClusterNetworkPolicy(t, s.Client, "pass-example", s.TimeoutConfig.GetTimeout)
			mutate := cnp.DeepCopy()
			// change egress rule from "deny" to "pass"
			mutate.Spec.Egress[0].Action = api.ClusterNetworkPolicyRuleActionPass
			kubernetes.PatchClusterNetworkPolicy(t, s.Client, cnp, mutate, s.TimeoutConfig.GetTimeout)

			// draco-malfoy-0 and draco-malfoy-1 are our server pods in slytherin namespace
			delegatedPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", s.TimeoutConfig.GetTimeout)
			acceptedPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-1", s.TimeoutConfig.GetTimeout)

			// harry-potter-0 is our client pod in gryffindor namespace, selected by the network policy
			// egress rule at index 0 passes traffic to draco-malfoy-0 on to the network policy, which allows TCP port 80 -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				delegatedPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
			// the network policy does not allow TCP port 8080, so harry-potter-0 is isolated on it -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				delegatedPod.Status.PodIP, int32(8080), s.TimeoutConfig, false)
			// egress rule at index 1 (accept-all-egress-to-slytherin) accepts traffic to draco-malfoy-1, so the
			// network policy is not consulted -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				acceptedPod.Status.PodIP, int32(8080), s.TimeoutConfig, true)

			// harry-potter-1 is our client pod in gryffindor namespace, not selected by the network policy
			// egress rule at index 0 passes traffic to draco-malfoy-0 and no network policy selects harry-potter-1, so
			// the `default` baseline CNP egress rule at index 0 (deny-all-egress-to-slytherin) applies -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				delegatedPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
			// egress rule at index 1 accepts traffic to draco-malfoy-1, so the baseline CNP is not consulted -> ALLOWED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-1", "tcp",
				acceptedPod.Status.PodIP, int32(80), s.TimeoutConfig, true)
		})

		t.Run("Should support a 'pass-ingress' policy for admin CNP and respect the match for baseline cluster network policy", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
			defer cancel()
			// This test uses `default` baseline CNP from api_integration/standard-anp-np-banp.yaml
			np := &networkingv1.NetworkPolicy{}
			err := s.Client.Get(ctx, client.ObjectKey{
				Namespace: "network-policy-conformance-gryffindor",
				Name:      "allow-gress-from-to-slytherin-to-gryffindor",
			}, np)
			require.NoErrorf(t, err, "unable to fetch the network policy")
			// delete network policy so that baseline CNP takes effect on harry-potter-0 as well
			err = s.Client.Delete(ctx, np)
			require.NoErrorf(t, err, "unable to delete the network policy")
			// harry-potter-0 is our server pod in gryffindor namespace
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-gryffindor", "harry-potter-0", s.TimeoutConfig.GetTimeout)
			// ingress rule at index 0 passes traffic from draco-malfoy-0 and no network policy selects harry-potter-0 anymore,
			// so the `default` baseline CNP ingress rule at index 0 (deny-all-ingress-from-slytherin) applies -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-slytherin", "draco-malfoy-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
		})

		t.Run("Should support a 'pass-egress' policy for admin CNP and respect the match for baseline cluster network policy", func(t *testing.T) {
			// This test uses `default` baseline CNP from api_integration/standard-anp-np-banp.yaml
			// draco-malfoy-0 is our server pod in slytherin namespace
			serverPod := kubernetes.GetPod(t, s.Client, "network-policy-conformance-slytherin", "draco-malfoy-0", s.TimeoutConfig.GetTimeout)
			// egress rule at index 0 passes traffic to draco-malfoy-0 and no network policy selects harry-potter-0 anymore,
			// so the `default` baseline CNP egress rule at index 0 (deny-all-egress-to-slytherin) applies -> DENIED
			kubernetes.PokeServer(t, s.ClientSet, &s.KubeConfig, "network-policy-conformance-gryffindor", "harry-potter-0", "tcp",
				serverPod.Status.PodIP, int32(80), s.TimeoutConfig, false)
		})
	},
}
