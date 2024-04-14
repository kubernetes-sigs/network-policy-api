package kube

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/api/networking/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1alpha12 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"testing"
	"time"
)

func TestReadNetworkPoliciesFromKube(t *testing.T) {
	anpNotFound := errors2.NewNotFound(schema.GroupResource{v1alpha12.GroupName, "AdminNetworkPolicy"}, "AdminNetworkPolicy")
	banpNotFound := errors2.NewNotFound(schema.GroupResource{v1alpha12.GroupName, "BaselineAdminNetworkPolicy"}, "BaselineAdminNetworkPolicy")
	netpolNotFound := errors2.NewNotFound(schema.GroupResource{v1.GroupName, "NetworkPolicies"}, "NetworkPolicies")

	scenarios := map[string]struct {
		ctxCreator     func() (context.Context, context.CancelFunc)
		anpResponse    func() ([]v1alpha12.AdminNetworkPolicy, error)
		banpResponse   func() (v1alpha12.BaselineAdminNetworkPolicy, error)
		netPolResponse func() ([]v1.NetworkPolicy, error)

		expectedNetErr  error
		expectedAnpErr  error
		expectedBanpErr error
	}{
		"timeout error on admin network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.TODO(), 1*time.Nanosecond)
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				time.Sleep(1 * time.Millisecond)
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
			expectedAnpErr: context.DeadlineExceeded,
		},
		"resource not found on admin network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, anpNotFound
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
			expectedAnpErr: anpNotFound,
		},
		"return admin network policies": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "admin-network-policy",
						},
					},
				}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
		},
		"timeout error on base admin network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.TODO(), 1*time.Nanosecond)
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				time.Sleep(1 * time.Millisecond)
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
			expectedAnpErr: context.DeadlineExceeded,
		},
		"resource not found on base admin network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, banpNotFound
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
			expectedBanpErr: banpNotFound,
		},
		"return base admin network policies": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "base-admin-network-policy"},
				}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, nil
			},
		},
		"timeout error on network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.TODO(), 1*time.Nanosecond)
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				time.Sleep(10 * time.Millisecond)
				return nil, nil
			},
			expectedNetErr: context.DeadlineExceeded,
		},
		"resource not found on network policies retrieval": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return nil, netpolNotFound
			},
			expectedNetErr: netpolNotFound,
		},
		"return network policies": {
			ctxCreator: func() (context.Context, context.CancelFunc) {
				return context.TODO(), func() {}
			},
			anpResponse: func() ([]v1alpha12.AdminNetworkPolicy, error) {
				return []v1alpha12.AdminNetworkPolicy{}, nil
			},
			banpResponse: func() (v1alpha12.BaselineAdminNetworkPolicy, error) {
				return v1alpha12.BaselineAdminNetworkPolicy{}, nil
			},
			netPolResponse: func() ([]v1.NetworkPolicy, error) {
				return []v1.NetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "network_policy"},
					},
				}, nil
			},
		},
	}

	for name, scenario := range scenarios {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := scenario.ctxCreator()
			defer cancel()

			k := &MockKubernetes{
				AdminNetworkPolicies: scenario.anpResponse,
				BaseNetworkPolicies:  scenario.banpResponse,
				NetworkPolicies:      scenario.netPolResponse,
			}
			netpol, anps, banp, netErr, anpErr, banpErr := ReadNetworkPoliciesFromKube(ctx, k, []string{"test"})
			if scenario.expectedNetErr != nil {
				if !errors.Is(netErr, scenario.expectedNetErr) {
					t.Fatalf("Unexpected error1: %v, expected %v", netErr, scenario.expectedNetErr)
				}
			}
			if scenario.expectedAnpErr != nil {
				if !errors.Is(anpErr, scenario.expectedAnpErr) {
					t.Fatalf("Unexpected error1: %v, expected %v", anpErr, scenario.expectedAnpErr)
				}
			}
			if scenario.expectedBanpErr != nil {
				if !errors.Is(banpErr, scenario.expectedBanpErr) {
					t.Fatalf("Unexpected error1: %v, expected %v", banpErr, scenario.expectedBanpErr)
				}
			}

			if len(anps) > 0 {
				expected, _ := scenario.anpResponse()
				if anps[0].Name != expected[0].Name {
					t.Fatalf("Unexpected ANP: %v, expected %v", anps[0].Name, expected[0].Name)
				}
			}

			if banp != nil {
				expected, _ := scenario.banpResponse()
				if banp.Name != expected.Name {
					t.Fatalf("Unexpected BANP: %v, expected %v", banp.Name, banp.Name)
				}
			}
			if len(netpol) > 0 {
				expected, _ := scenario.netPolResponse()
				fmt.Println(netpol[0].Name, expected[0].Name+"1")
				if netpol[0].Name != expected[0].Name {
					t.Fatalf("Unexpected NetworkPolicy: %v, expected %v", netpol[0].Name, expected[0].Name)
				}
			}
		})
	}

}
