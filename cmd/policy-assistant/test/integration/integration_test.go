package connectivity

/*
This integration test suite verifies:
1. Building/translating NetPol spec into interim data structures (matchers).
2. Simulation of expected connectivity for ANP, BANP, and v1 NetPols.

It displays connectivity matrixes for each test as well, which can be helpful for debugging.
*/

import (
	"testing"

	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// some copied vars needed to reference them as pointers
var (
	udp        = v1.ProtocolUDP
	serve80tcp = "serve-80-tcp"
)

type connectivityTest struct {
	name                   string
	args                   args
	defaultIngressBehavior probe.Connectivity
	defaultEgressBehavior  probe.Connectivity
	// nonDefaultIngress represents e.g. ingress flows that are allowed when defaultIngressBehavior is to deny
	nonDefaultIngress []flow
	// nonDefaultEgress is like nonDefaultIngress
	nonDefaultEgress []flow
}

type args struct {
	resources *probe.Resources
	netpols   []*networkingv1.NetworkPolicy
	anps      []*v1alpha1.AdminNetworkPolicy
	banp      *v1alpha1.BaselineAdminNetworkPolicy
}

type flow struct {
	from, to string
	port     int
	proto    v1.Protocol
}

func TestNetPolV1Connectivity(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "ingress port allowed",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/a", "x/a", 80, v1.ProtocolTCP},
				{"x/a", "x/a", 81, v1.ProtocolTCP},
				{"x/a", "x/a", 81, v1.ProtocolUDP},
				{"x/b", "x/a", 80, v1.ProtocolTCP},
				{"x/b", "x/a", 81, v1.ProtocolTCP},
				{"x/b", "x/a", 81, v1.ProtocolUDP},
				{"y/a", "x/a", 80, v1.ProtocolTCP},
				{"y/a", "x/a", 81, v1.ProtocolTCP},
				{"y/a", "x/a", 81, v1.ProtocolUDP},
				{"y/b", "x/a", 80, v1.ProtocolTCP},
				{"y/b", "x/a", 81, v1.ProtocolTCP},
				{"y/b", "x/a", 81, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				netpols: []*networkingv1.NetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "x",
							Name:      "base",
						},
						Spec: networkingv1.NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							Ingress: []networkingv1.NetworkPolicyIngressRule{
								{
									Ports: []networkingv1.NetworkPolicyPort{
										{
											Protocol: &udp,
											Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 80},
										},
									},
								},
							},
							PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
						},
					},
				},
			},
		},
	}

	runConnectivityTests(t, tests...)
}

func TestANPConnectivity(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "egress port number protocol unspecified",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultEgress: []flow{
				{"x/a", "x/b", 80, v1.ProtocolTCP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													NamespaceSelector: &metav1.LabelSelector{
														MatchLabels: map[string]string{"ns": "x"},
													},
												},
												PodSelector: metav1.LabelSelector{
													MatchLabels: map[string]string{"pod": "b"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port: 80,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ingress port number protocol unspecified",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/b", "x/a", 80, v1.ProtocolTCP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													NamespaceSelector: &metav1.LabelSelector{
														MatchLabels: map[string]string{"ns": "x"},
													},
												},
												PodSelector: metav1.LabelSelector{
													MatchLabels: map[string]string{"pod": "b"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port: 80,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ingress named port",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/b", "x/a", 80, v1.ProtocolTCP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													NamespaceSelector: &metav1.LabelSelector{
														MatchLabels: map[string]string{"ns": "x"},
													},
												},
												PodSelector: metav1.LabelSelector{
													MatchLabels: map[string]string{"pod": "b"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											NamedPort: &serve80tcp,
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ingress same labels port range",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/a", "x/a", 80, v1.ProtocolTCP},
				{"x/a", "x/a", 81, v1.ProtocolTCP},
				{"x/b", "x/a", 80, v1.ProtocolTCP},
				{"x/b", "x/a", 81, v1.ProtocolTCP},
				{"x/c", "x/a", 80, v1.ProtocolTCP},
				{"x/c", "x/a", 81, v1.ProtocolTCP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y", "z"}, []string{"a", "b", "c"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
									}),
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													SameLabels: []string{"ns"},
												},
												PodSelector: metav1.LabelSelector{},
											},
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
			name:                   "not same labels",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"y/a", "x/a", 80, v1.ProtocolUDP},
				{"y/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NotSameLabels: []string{"ns"},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port:     80,
												Protocol: v1.ProtocolUDP,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ordering matters for overlapping rules (order #1)",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"y/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													NamespaceSelector: &metav1.LabelSelector{
														MatchLabels: map[string]string{"ns": "y"},
													},
												},
												PodSelector: metav1.LabelSelector{
													MatchLabels: map[string]string{"pod": "a"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port:     80,
												Protocol: v1.ProtocolUDP,
											},
										},
									}),
								},
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NotSameLabels: []string{"ns"},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port:     80,
												Protocol: v1.ProtocolUDP,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ordering matters for overlapping rules (order #2)",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"y/a", "x/a", 80, v1.ProtocolUDP},
				{"y/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NotSameLabels: []string{"ns"},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port:     80,
												Protocol: v1.ProtocolUDP,
											},
										},
									}),
								},
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Pods: &v1alpha1.NamespacedPodPeer{
												Namespaces: v1alpha1.NamespacedPeer{
													NamespaceSelector: &metav1.LabelSelector{
														MatchLabels: map[string]string{"ns": "y"},
													},
												},
												PodSelector: metav1.LabelSelector{
													MatchLabels: map[string]string{"pod": "a"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Port:     80,
												Protocol: v1.ProtocolUDP,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "deny all egress",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityBlocked,
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "multiple ANPs (priority order #1)",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 99,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "multiple ANPs (priority order #2)",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityBlocked,
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 101,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Egress: []v1alpha1.AdminNetworkPolicyEgressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionDeny,
									To: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
	}

	runConnectivityTests(t, tests...)
}

func TestBANPConnectivity(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "egress",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultEgress: []flow{
				{"x/a", "x/b", 80, v1.ProtocolTCP},
				{"x/a", "x/b", 80, v1.ProtocolUDP},
				{"x/a", "x/b", 81, v1.ProtocolTCP},
				{"x/a", "x/b", 81, v1.ProtocolUDP},
				{"x/a", "y/b", 80, v1.ProtocolTCP},
				{"x/a", "y/b", 80, v1.ProtocolUDP},
				{"x/a", "y/b", 81, v1.ProtocolTCP},
				{"x/a", "y/b", 81, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Pods: &v1alpha1.NamespacedPodSubject{
								NamespaceSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"ns": "x"},
								},
								PodSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"pod": "a"},
								},
							},
						},
						Egress: []v1alpha1.BaselineAdminNetworkPolicyEgressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								To: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Pods: &v1alpha1.NamespacedPodPeer{
											Namespaces: v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
											PodSelector: metav1.LabelSelector{
												MatchLabels: map[string]string{"pod": "b"},
											},
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
			name:                   "ingress",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/b", "x/a", 80, v1.ProtocolTCP},
				{"x/b", "x/a", 80, v1.ProtocolUDP},
				{"x/b", "x/a", 81, v1.ProtocolTCP},
				{"x/b", "x/a", 81, v1.ProtocolUDP},
				{"y/b", "x/a", 80, v1.ProtocolTCP},
				{"y/b", "x/a", 80, v1.ProtocolUDP},
				{"y/b", "x/a", 81, v1.ProtocolTCP},
				{"y/b", "x/a", 81, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Pods: &v1alpha1.NamespacedPodSubject{
								NamespaceSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"ns": "x"},
								},
								PodSelector: metav1.LabelSelector{
									MatchLabels: map[string]string{"pod": "a"},
								},
							},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Pods: &v1alpha1.NamespacedPodPeer{
											Namespaces: v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
											PodSelector: metav1.LabelSelector{
												MatchLabels: map[string]string{"pod": "b"},
											},
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
			name:                   "ordering matters for overlapping rules (order #1)",
			defaultIngressBehavior: probe.ConnectivityBlocked,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"y/b", "x/a", 80, v1.ProtocolTCP},
				{"y/b", "x/b", 80, v1.ProtocolTCP},
				{"y/b", "y/a", 80, v1.ProtocolTCP},
				{"y/b", "y/b", 80, v1.ProtocolTCP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80}, []v1.Protocol{v1.ProtocolTCP}),
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Namespaces: &metav1.LabelSelector{},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionAllow,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Pods: &v1alpha1.NamespacedPodPeer{
											Namespaces: v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{
													MatchLabels: map[string]string{"ns": "y"},
												},
											},
											PodSelector: metav1.LabelSelector{
												MatchLabels: map[string]string{"pod": "b"},
											},
										},
									},
								},
							},
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Namespaces: &v1alpha1.NamespacedPeer{
											NamespaceSelector: &metav1.LabelSelector{},
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
			name:                   "ordering matters for overlapping rules (order #2)",
			defaultIngressBehavior: probe.ConnectivityBlocked,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80}, []v1.Protocol{v1.ProtocolTCP}),
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Namespaces: &metav1.LabelSelector{},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Namespaces: &v1alpha1.NamespacedPeer{
											NamespaceSelector: &metav1.LabelSelector{},
										},
									},
								},
							},
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Pods: &v1alpha1.NamespacedPodPeer{
											Namespaces: v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{
													MatchLabels: map[string]string{"ns": "y"},
												},
											},
											PodSelector: metav1.LabelSelector{
												MatchLabels: map[string]string{"pod": "b"},
											},
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

	runConnectivityTests(t, tests...)
}

func TestANPWithNetPolV1(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "ANP allow all over NetPol",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				netpols: []*networkingv1.NetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "x",
							Name:      "base",
						},
						Spec: networkingv1.NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							Ingress: []networkingv1.NetworkPolicyIngressRule{
								{
									Ports: []networkingv1.NetworkPolicyPort{
										{
											Protocol: &udp,
											Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 80},
										},
									},
								},
							},
							PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
						},
					},
				},
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 99,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
		{
			name:                   "ANP allow some over NetPol",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"y/a", "x/a", 80, v1.ProtocolTCP},
				{"y/a", "x/a", 81, v1.ProtocolTCP},
				{"y/a", "x/a", 81, v1.ProtocolUDP},
				{"y/b", "x/a", 80, v1.ProtocolTCP},
				{"y/b", "x/a", 81, v1.ProtocolTCP},
				{"y/b", "x/a", 81, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				netpols: []*networkingv1.NetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "x",
							Name:      "base",
						},
						Spec: networkingv1.NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							Ingress: []networkingv1.NetworkPolicyIngressRule{
								{
									Ports: []networkingv1.NetworkPolicyPort{
										{
											Protocol: &udp,
											Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 80},
										},
									},
								},
							},
							PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
						},
					},
				},
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 99,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{
													MatchLabels: map[string]string{"ns": "x"},
												},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolTCP,
												Start:    80,
												End:      81,
											},
										},
										{
											PortRange: &v1alpha1.PortRange{
												Protocol: v1.ProtocolUDP,
												Start:    80,
												End:      81,
											},
										},
									}),
								},
							},
						},
					},
				},
			},
		},
	}

	runConnectivityTests(t, tests...)
}

func TestBANPWithNetPolV1(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "BANP deny all after NetPol",
			defaultIngressBehavior: probe.ConnectivityBlocked,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				netpols: []*networkingv1.NetworkPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "x",
							Name:      "base",
						},
						Spec: networkingv1.NetworkPolicySpec{
							PodSelector: metav1.LabelSelector{
								MatchLabels: map[string]string{"pod": "a"},
							},
							Ingress: []networkingv1.NetworkPolicyIngressRule{
								{
									From: []networkingv1.NetworkPolicyPeer{
										{
											PodSelector: &metav1.LabelSelector{
												MatchLabels: map[string]string{"pod": "b"},
											},
										},
									},
									Ports: []networkingv1.NetworkPolicyPort{
										{
											Protocol: &udp,
											Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 80},
										},
									},
								},
							},
							PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
						},
					},
				},
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Namespaces: &metav1.LabelSelector{},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Namespaces: &v1alpha1.NamespacedPeer{
											NamespaceSelector: &metav1.LabelSelector{},
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

	runConnectivityTests(t, tests...)
}

func TestANPWithBANP(t *testing.T) {
	tests := []connectivityTest{
		{
			name:                   "BANP deny all after ANP",
			defaultIngressBehavior: probe.ConnectivityBlocked,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/a", "x/a", 80, v1.ProtocolUDP},
				{"x/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												SameLabels: []string{"ns"},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Protocol: v1.ProtocolUDP,
												Port:     80,
											},
										},
									}),
								},
							},
						},
					},
				},
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Namespaces: &metav1.LabelSelector{},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Namespaces: &v1alpha1.NamespacedPeer{
											NamespaceSelector: &metav1.LabelSelector{},
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
			name:                   "ANP pass some and allow rest over BANP",
			defaultIngressBehavior: probe.ConnectivityAllowed,
			defaultEgressBehavior:  probe.ConnectivityAllowed,
			nonDefaultIngress: []flow{
				{"x/a", "x/a", 80, v1.ProtocolUDP},
				{"x/b", "x/a", 80, v1.ProtocolUDP},
			},
			args: args{
				resources: getResources(t, []string{"x", "y"}, []string{"a", "b"}, []int{80, 81}, []v1.Protocol{v1.ProtocolTCP, v1.ProtocolUDP}),
				anps: []*v1alpha1.AdminNetworkPolicy{
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 100,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Pods: &v1alpha1.NamespacedPodSubject{
									NamespaceSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"ns": "x"},
									},
									PodSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"pod": "a"},
									},
								},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionPass,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												SameLabels: []string{"ns"},
											},
										},
									},
									Ports: &([]v1alpha1.AdminNetworkPolicyPort{
										{
											PortNumber: &v1alpha1.Port{
												Protocol: v1.ProtocolUDP,
												Port:     80,
											},
										},
									}),
								},
							},
						},
					},
					{
						Spec: v1alpha1.AdminNetworkPolicySpec{
							Priority: 101,
							Subject: v1alpha1.AdminNetworkPolicySubject{
								Namespaces: &metav1.LabelSelector{},
							},
							Ingress: []v1alpha1.AdminNetworkPolicyIngressRule{
								{
									Action: v1alpha1.AdminNetworkPolicyRuleActionAllow,
									From: []v1alpha1.AdminNetworkPolicyPeer{
										{
											Namespaces: &v1alpha1.NamespacedPeer{
												NamespaceSelector: &metav1.LabelSelector{},
											},
										},
									},
								},
							},
						},
					},
				},
				banp: &v1alpha1.BaselineAdminNetworkPolicy{
					Spec: v1alpha1.BaselineAdminNetworkPolicySpec{
						Subject: v1alpha1.AdminNetworkPolicySubject{
							Namespaces: &metav1.LabelSelector{},
						},
						Ingress: []v1alpha1.BaselineAdminNetworkPolicyIngressRule{
							{
								Action: v1alpha1.BaselineAdminNetworkPolicyRuleActionDeny,
								From: []v1alpha1.AdminNetworkPolicyPeer{
									{
										Namespaces: &v1alpha1.NamespacedPeer{
											NamespaceSelector: &metav1.LabelSelector{},
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

	runConnectivityTests(t, tests...)
}

func runConnectivityTests(t *testing.T, tests ...connectivityTest) {
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.args.resources, "resources must be set")
			require.True(t, tt.defaultIngressBehavior == probe.ConnectivityAllowed || tt.defaultIngressBehavior == probe.ConnectivityBlocked, "default ingress behavior must be set to allowed or blocked")
			require.True(t, tt.defaultEgressBehavior == probe.ConnectivityAllowed || tt.defaultEgressBehavior == probe.ConnectivityBlocked, "default egress behavior must be set to allowed or blocked")

			table := probe.NewTableWithDefaultConnectivity(tt.args.resources, tt.defaultIngressBehavior, tt.defaultEgressBehavior)

			for _, f := range tt.nonDefaultIngress {
				if tt.defaultIngressBehavior == probe.ConnectivityBlocked {
					table.SetIngress(probe.ConnectivityAllowed, f.from, f.to, f.port, f.proto)
				} else {
					table.SetIngress(probe.ConnectivityBlocked, f.from, f.to, f.port, f.proto)
				}
			}

			for _, f := range tt.nonDefaultEgress {
				if tt.defaultEgressBehavior == probe.ConnectivityBlocked {
					table.SetEgress(probe.ConnectivityAllowed, f.from, f.to, f.port, f.proto)
				} else {
					table.SetEgress(probe.ConnectivityBlocked, f.from, f.to, f.port, f.proto)
				}
			}

			parsedPolicy := matcher.BuildV1AndV2NetPols(false, tt.args.netpols, tt.args.anps, tt.args.banp)
			jobBuilder := &probe.JobBuilder{TimeoutSeconds: 3}
			simRunner := probe.NewSimulatedRunner(parsedPolicy, jobBuilder)
			simTable := simRunner.RunProbeForConfig(generator.ProbeAllAvailable, tt.args.resources)

			expected := table.RenderIngress()
			actual := simTable.RenderIngress()
			isEqual := assert.Equal(t, expected, actual)
			if isEqual {
				t.Logf("validated ingress:\n%s\n", expected)
			} else {
				t.Logf("expected ingress:\n%s\n", expected)
				t.Logf("actual ingress:\n%s\n", actual)
			}

			expected = table.RenderEgress()
			actual = simTable.RenderEgress()
			isEqual = assert.Equal(t, expected, actual)
			if isEqual {
				t.Logf("validated egress:\n%s\n", expected)
			} else {
				t.Logf("expected egress:\n%s\n", expected)
				t.Logf("actual egress:\n%s\n", actual)
			}
		})
	}
}

func getResources(t *testing.T, namespaces, podNames []string, ports []int, protocols []v1.Protocol) *probe.Resources {
	kubernetes := kube.NewMockKubernetes(1.0)
	resources, err := probe.NewDefaultResources(kubernetes, namespaces, podNames, ports, protocols, []string{}, 5, false, "registry.k8s.io")
	require.Nil(t, err, "failed to create resources")
	return resources
}
