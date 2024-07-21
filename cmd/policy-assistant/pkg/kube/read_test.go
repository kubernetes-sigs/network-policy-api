package kube

import (
	"context"
	"errors"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha12 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"testing"
)

func TestReadNetworkPoliciesFromKube(t *testing.T) {
	scenarios := map[string]struct {
		AdminNetworkPolicies         []v1alpha12.AdminNetworkPolicy
		BaselineAdminNetworkPolicies []v1alpha12.BaselineAdminNetworkPolicy
		NetworkPolicies              []v1.NetworkPolicy

		expectedNetErr  error
		expectedAnpErr  error
		expectedBanpErr error
	}{
		"parse error on admin network policies retrieval": {
			expectedAnpErr: context.DeadlineExceeded,
		},
		"return admin network policies": {
			AdminNetworkPolicies: []v1alpha12.AdminNetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "admin-network-policy",
					},
				},
			},
		},
		"parse error on base admin network policies retrieval": {
			expectedAnpErr: context.DeadlineExceeded,
		},
		"return base admin network policies": {
			BaselineAdminNetworkPolicies: []v1alpha12.BaselineAdminNetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "base-admin-network-policy"},
				},
			},
		},
		"parse error on network policies retrieval": {
			expectedNetErr: context.DeadlineExceeded,
		},
		"return network policies": {
			NetworkPolicies: []v1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "network_policy",
						Namespace: "default",
					},
				},
			},
		},
	}

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
			k := &MockKubernetes{
				AdminNetworkPolicies:        scenario.AdminNetworkPolicies,
				AdminNetworkPolicyError:     scenario.expectedAnpErr,
				BaseNetworkPolicies:         scenario.BaselineAdminNetworkPolicies,
				BaseAdminNetworkPolicyError: scenario.expectedBanpErr,
				Namespaces:                  map[string]*MockNamespace{},
				NetworkPolicyError:          scenario.expectedNetErr,
			}

			for _, np := range scenario.NetworkPolicies {
				k.Namespaces[np.Namespace] = &MockNamespace{
					Netpols: map[string]*v1.NetworkPolicy{
						np.Name: &np,
					},
				}
			}

			netpol, anps, banp, netErr, anpErr, banpErr := ReadNetworkPoliciesFromKube(context.TODO(), k, []string{"default"}, true, true)
			if scenario.expectedNetErr != nil {
				if !errors.Is(netErr, scenario.expectedNetErr) {
					t.Fatalf("Unexpected error: %v, expected %v", netErr, scenario.expectedNetErr)
				}
			}
			if scenario.expectedAnpErr != nil {
				if !errors.Is(anpErr, scenario.expectedAnpErr) {
					t.Fatalf("Unexpected error: %v, expected %v", anpErr, scenario.expectedAnpErr)
				}
			}
			if scenario.expectedBanpErr != nil {
				if !errors.Is(banpErr, scenario.expectedBanpErr) {
					t.Fatalf("Unexpected error: %v, expected %v", banpErr, scenario.expectedBanpErr)
				}
			}

			if len(scenario.AdminNetworkPolicies) > 0 {
				if anps[0].Name != scenario.AdminNetworkPolicies[0].Name {
					t.Fatalf("Unexpected ANP: %v, expected %v", anps[0].Name, scenario.AdminNetworkPolicies[0].Name)
				}
			}

			if scenario.BaselineAdminNetworkPolicies != nil {
				if banp.Name != scenario.BaselineAdminNetworkPolicies[0].Name {
					t.Fatalf("Unexpected BANP: %v, expected %v", banp.Name, banp.Name)
				}
			}
			if len(scenario.NetworkPolicies) > 0 {
				if netpol[0].Name != scenario.NetworkPolicies[0].Name {
					t.Fatalf("Unexpected NetworkPolicy: %v, expected %v", netpol[0].Name, scenario.NetworkPolicies[0].Name)
				}
			}
		})
	}

}
