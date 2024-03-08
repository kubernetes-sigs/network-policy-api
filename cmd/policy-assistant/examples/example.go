package examples

import (
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

var CoreGressRulesCombinedANB = []*v1alpha1.AdminNetworkPolicy{
	{
		ObjectMeta: v1.ObjectMeta{
			Name: "example-anp",
		},
		Spec: v1alpha1.AdminNetworkPolicySpec{
			Priority: 20,
			Subject: v1alpha1.AdminNetworkPolicySubject{
				Namespaces: &v1.LabelSelector{
					MatchExpressions: []v1.LabelSelectorRequirement{
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: v1.LabelSelectorOpExists,
						},
					},
				},
			},
			Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
				{
					Name:   "allow-to-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-to-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "pass-to-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-to-slytherin-at-ports-80-53-9003",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "pass-to-slytherin-at-port-80-53-9003",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "allow-to-hufflepuff-at-ports-8080-5353",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 8080},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "deny-to-hufflepuff-everything-else",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
				},
			},
			Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
				{
					Name:   "allow-from-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-from-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "pass-from-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-from-slytherin-at-port-80-53-9003",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "pass-from-slytherin-at-port-80-53-9003",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "allow-from-hufflepuff-at-port-80-5353-9003",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "deny-from-hufflepuff-everything-else",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name: "example-anp-2",
		},
		Spec: v1alpha1.AdminNetworkPolicySpec{
			Priority: 16,
			Subject: v1alpha1.AdminNetworkPolicySubject{
				Namespaces: &v1.LabelSelector{
					MatchExpressions: []v1.LabelSelectorRequirement{
						{
							Key:      "kubernetes.io/metadata.name",
							Operator: v1.LabelSelectorOpExists,
						},
					},
				},
			},
			Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
				{
					Name:   "allow-to-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-to-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "pass-to-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-to-slytherin-at-ports-80-53-9003-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "pass-to-slytherin-at-port-80-53-9003-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "allow-to-hufflepuff-at-ports-8080-5353-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 8080},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "deny-to-hufflepuff-everything-else-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
				},
			},
			Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
				{
					Name:   "allow-from-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-from-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "pass-from-ravenclaw-everything-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
				{
					Name:   "deny-from-slytherin-at-port-80-53-9003-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "pass-from-slytherin-at-port-80-53-9003-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "allow-from-hufflepuff-at-port-80-5353-9003-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
					Ports: &[]v1alpha1.AdminNetworkPolicyPort{
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
						},
						{
							PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
						},
					},
				},
				{
					Name:   "deny-from-hufflepuff-everything-else-2",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					From: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

var CoreGressRulesCombinedBANB *v1alpha1.BaselineAdminNetworkPolicy = &v1alpha1.BaselineAdminNetworkPolicy{
	ObjectMeta: v1.ObjectMeta{
		Name: "default",
	},
	Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
		Subject: v1alpha1.AdminNetworkPolicySubject{
			Namespaces: &v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "kubernetes.io/metadata.name",
						Operator: v1.LabelSelectorOpExists,
					},
				},
			},
		},
		Egress: []v1alpha1.BaselineAdminNetworkPolicyEgressRule{
			{
				Name:   "allow-to-ravenclaw-everything",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow,
				To: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							SameLabels: []string{"Test"},
						},
					},
				},
			},
			{
				Name:   "deny-to-ravenclaw-everything",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
				To: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NotSameLabels: []string{"Test1"},
						},
					},
				},
			},
			{
				Name:   "deny-to-slytherin-at-ports-80-53-9003",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
				To: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchExpressions: []v1.LabelSelectorRequirement{
									{
										Key:      "kubernetes.io/metadata.name",
										Operator: v1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
				Ports: &[]v1alpha1.AdminNetworkPolicyPort{
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
					},
				},
			},
			{
				Name:   "allow-to-hufflepuff-at-ports-8080-5353",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow,
				To: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
								},
							},
						},
					},
				},
				Ports: &[]v1alpha1.AdminNetworkPolicyPort{
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 8080},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
					},
				},
			},
			{
				Name:   "deny-to-hufflepuff-everything-else",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
				To: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
								},
							},
						},
					},
				},
			},
		},
		Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
			{
				Name:   "allow-from-ravenclaw-everything",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow,
				From: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
								},
							},
						},
					},
				},
			},
			{
				Name:   "deny-from-slytherin-at-port-80-53-9003",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
				From: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-slytherin",
								},
							},
						},
					},
				},
				Ports: &[]v1alpha1.AdminNetworkPolicyPort{
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 53},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
					},
				},
			},
			{
				Name:   "allow-from-hufflepuff-at-port-80-5353-9003",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow,
				From: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
								},
							},
						},
					},
				},
				Ports: &[]v1alpha1.AdminNetworkPolicyPort{
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolTCP, Port: 80},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolUDP, Port: 5353},
					},
					{
						PortNumber: &v1alpha1.Port{Protocol: v12.ProtocolSCTP, Port: 9003},
					},
				},
			},
			{
				Name:   "deny-from-hufflepuff-everything-else",
				Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
				From: []v1alpha1.AdminNetworkPolicyPeer{
					{
						Namespaces: &v1alpha1.NamespacedPeer{
							NamespaceSelector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "network-policy-conformance-hufflepuff",
								},
							},
						},
					},
				},
			},
		},
	},
}

var SimpleANPs = []*v1alpha1.AdminNetworkPolicy{
	{
		ObjectMeta: v1.ObjectMeta{
			Name: "simple-anp-1",
		},
		Spec: v1alpha1.AdminNetworkPolicySpec{
			Priority: 34,
			Subject: v1alpha1.AdminNetworkPolicySubject{
				Namespaces: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "test",
					},
				},
			},
			Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
				{
					Name:   "allow-to-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name: "simple-anp-2",
		},
		Spec: v1alpha1.AdminNetworkPolicySpec{
			Priority: 50,
			Subject: v1alpha1.AdminNetworkPolicySubject{
				Namespaces: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "test",
					},
				},
			},
			Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
				{
					Name:   "allow-to-ravenclaw-everything",
					Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
					To: []v1alpha1.AdminNetworkPolicyPeer{
						{
							Namespaces: &v1alpha1.NamespacedPeer{
								NamespaceSelector: &v1.LabelSelector{
									MatchLabels: map[string]string{
										"kubernetes.io/metadata.name": "network-policy-conformance-ravenclaw",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
