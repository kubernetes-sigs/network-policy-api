package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *TestCaseGenerator) SpecialServiceCases(nodeIPs []string) []*TestCase {
	cases := loadBalancerClusterCases(nodeIPs)
	cases = append(cases, nodePortClusterCases(nodeIPs)...)
	cases = append(cases, loadBalancerLocalCases()...)
	cases = append(cases, nodePortLocalCases(nodeIPs)...)
	return cases
}

// loadBalancerClusterCases should have the exact same tests as nodePortClusterCases, except for the service type.
func loadBalancerClusterCases(nodeIPs []string) []*TestCase {
	cfg := &ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster}
	baseTags := []string{TagLoadBalancer, TagExternalTrafficPolicyCluster}

	return []*TestCase{
		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: deny all ingress",
			tags(baseTags, TagIngress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow ingress from pods",
			tags(baseTags, TagIngress, TagAllNamespaces, TagCNIBringsSourcePodInfoToOtherNode),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow ingress from pods and nodes",
			tags(baseTags, TagIngress, TagAllNamespaces, TagNoCNISourcePodInfoToOtherNode, TagIPBlockNoExcept),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods-nodes", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: allNamespaceAndNodePeers(nodeIPs),
					},
				},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: deny all egress",
			tags(baseTags, TagEgress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow egress to pods",
			tags(baseTags, TagEgress, TagAllNamespaces, TagCNIBringsSourcePodInfoToOtherNode),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow egress to pods and nodes",
			tags(baseTags, TagEgress, TagAllNamespaces, TagNoCNISourcePodInfoToOtherNode, TagIPBlockNoExcept),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods-nodes", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: allNamespaceAndNodePeers(nodeIPs),
					},
				},
			})...,
		),
	}
}

// nodePortClusterCases should have the exact same tests as loadBalancerClusterCases, except for the service type.
func nodePortClusterCases(nodeIPs []string) []*TestCase {
	cfg := &ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: ToSourcePodNode}
	baseTags := []string{TagNodePort, TagExternalTrafficPolicyCluster, TagToSourcePodNode}

	return []*TestCase{
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all ingress (to source pod's node)",
			tags(baseTags, TagIngress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow ingress from pods (to source pod's node)",
			tags(baseTags, TagIngress, TagAllNamespaces, TagCNIBringsSourcePodInfoToOtherNode),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow ingress from pods and nodes (to source pod's node)",
			tags(baseTags, TagIngress, TagAllNamespaces, TagNoCNISourcePodInfoToOtherNode, TagIPBlockNoExcept),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: allNamespaceAndNodePeers(nodeIPs),
					},
				},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all egress (to source pod's node)",
			tags(baseTags, TagEgress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow egress to pods (to source pod's node)",
			tags(baseTags, TagEgress, TagAllNamespaces),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),
	}
}

// loadBalancerLocalCases should have the exact same tests as nodePortLocalCases, except for the service type
func loadBalancerLocalCases() []*TestCase {
	cfg := &ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerLocal}
	baseTags := []string{TagLoadBalancer, TagExternalTrafficPolicyLocal}

	return []*TestCase{
		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: deny all ingress",
			tags(baseTags, TagIngress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: allow ingress from pods",
			tags(baseTags, TagIngress, TagAllNamespaces),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: deny all egress",
			tags(baseTags, TagEgress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: allow egress to pods",
			tags(baseTags, TagEgress, TagAllNamespaces),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),
	}
}

// nodePortLocalCases should have the exact same tests as loadBalancerLocalCases, except for the service type
func nodePortLocalCases(nodeIPs []string) []*TestCase {
	cfg := &ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortLocal, DestinationNode: ToDestinationPodNode}
	baseTags := []string{TagNodePort, TagExternalTrafficPolicyLocal, TagToDestinationPodNode}

	return []*TestCase{
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: deny all ingress (to destination pod's node)",
			tags(baseTags, TagIngress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow ingress from pods (to destination pod's node)",
			tags(baseTags, TagIngress, TagAllNamespaces),
			cfg,
			createPolicyInAllNamespaces("allow-ingress-from-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress: []NetworkPolicyIngressRule{
					{
						From: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: deny all egress (to destination pod's node)",
			tags(baseTags, TagEgress, TagDenyAll),
			cfg,
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow egress to pods (to destination pod's node)",
			tags(baseTags, TagEgress, TagAllNamespaces, TagCNIBringsSourcePodInfoToOtherNode),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: []NetworkPolicyPeer{
							{NamespaceSelector: &metav1.LabelSelector{}},
						},
					},
				},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow egress to pods and nodes (to destination pod's node)",
			tags(baseTags, TagEgress, TagAllNamespaces, TagNoCNISourcePodInfoToOtherNode, TagIPBlockNoExcept),
			cfg,
			createPolicyInAllNamespaces("allow-egress-to-pods", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress: []NetworkPolicyEgressRule{
					{
						To: allNamespaceAndNodePeers(nodeIPs),
					},
				},
			})...,
		),
	}
}

func createPolicyInAllNamespaces(policyName string, spec NetworkPolicySpec) []*Action {
	return []*Action{
		CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyName,
				Namespace: "x",
			},
			Spec: spec,
		}),
		CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyName,
				Namespace: "y",
			},
			Spec: spec,
		}),
		CreatePolicy(&NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyName,
				Namespace: "z",
			},
			Spec: spec,
		}),
	}
}

func allNamespaceAndNodePeers(ips []string) []NetworkPolicyPeer {
	peers := make([]NetworkPolicyPeer, 1+len(ips))
	peers[0] = NetworkPolicyPeer{NamespaceSelector: &metav1.LabelSelector{}}
	for i, ip := range ips {
		peers[1+i] = NetworkPolicyPeer{IPBlock: &IPBlock{CIDR: ip + "/32"}}
	}
	return peers
}
