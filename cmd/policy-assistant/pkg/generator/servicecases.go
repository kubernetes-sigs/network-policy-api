package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *TestCaseGenerator) ServiceCases(nodeIPs []string) []*TestCase {
	cases := loadBalancerClusterCases(nodeIPs)
	cases = append(cases, nodePortClusterCases(nodeIPs)...)
	cases = append(cases, loadBalancerLocalCases()...)
	cases = append(cases, nodePortLocalCases()...)
	return cases
}

// loadBalancerClusterCases should have the exact same tests as nodePortClusterCases, except for the service type.
func loadBalancerClusterCases(nodeIPs []string) []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: deny all ingress",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagIngress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow ingress from pods",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagIngress, TagPodsByLabel, TagCNIBringsSourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
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
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagIngress, TagPodsByLabel, TagNoCNISourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
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
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagEgress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Cluster: allow egress to pods",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagEgress, TagPodsByLabel, TagCNIBringsSourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
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
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyCluster, TagEgress, TagPodsByLabel, TagNoCNISourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerCluster},
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
	cases := nodePortClusterCasesDestinationPodNode(nodeIPs)
	cases = append(cases, nodePortClusterCasesNotDestinationPodNode(nodeIPs)...)
	return cases
}

func nodePortClusterCasesDestinationPodNode(nodeIPs []string) []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all ingress (to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagDestinationPodNode, TagIngress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: DestinationPodNode},
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow ingress from pods (to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagDestinationPodNode, TagIngress, TagPodsByLabel),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: DestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all egress (to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagDestinationPodNode, TagEgress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: DestinationPodNode},
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		// FIXME Discrepancy found:40 wrong, 9 ignored, 32 correct
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow egress to pods (to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagDestinationPodNode, TagEgress, TagPodsByLabel, TagCNIBringsSourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: DestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow egress to pods and nodes (to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagDestinationPodNode, TagEgress, TagPodsByLabel, TagNoCNISourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: DestinationPodNode},
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

// NOTE: "not destination pod node" cases have an extra test case and distinction between ingress "from pods" versus "from pods and nodes"
func nodePortClusterCasesNotDestinationPodNode(nodeIPs []string) []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all ingress (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagIngress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		// FIXME Discrepancy found:72 wrong, 9 ignored, 0 correct
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow ingress from pods (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagIngress, TagPodsByLabel, TagCNIBringsSourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow ingress from pods and nodes (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagIngress, TagPodsByLabel, TagNoCNISourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: deny all egress (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagEgress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		// FIXME Discrepancy found:32 wrong, 9 ignored, 40 correct
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow egress to pods (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagEgress, TagPodsByLabel, TagCNIBringsSourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Cluster: allow egress to pods and nodes (NOT to destination pod's node)",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyCluster, TagNotDestinationPodNode, TagEgress, TagPodsByLabel, TagNoCNISourcePodInfoToOtherNode),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortCluster, DestinationNode: NotDestinationPodNode},
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

// loadBalancerLocalCases should have the exact same tests as nodePortLocalCases, except for the service type
func loadBalancerLocalCases() []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: deny all ingress",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyLocal, TagIngress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerLocal},
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: allow ingress from pods",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyLocal, TagIngress, TagPodsByLabel),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerLocal},
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
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyLocal, TagEgress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerLocal},
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		NewSingleStepTestCase("LoadBalancer with externalTrafficPolicy=Local: allow egress to pods",
			NewStringSet(TagLoadBalancer, TagExternalTrafficPolicyLocal, TagEgress, TagPodsByLabel),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: LoadBalancerLocal},
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
func nodePortLocalCases() []*TestCase {
	return []*TestCase{
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: deny all ingress",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyLocal, TagIngress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortLocal, DestinationNode: DestinationPodNode},
			createPolicyInAllNamespaces("deny-ingress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeIngress},
				Ingress:     []NetworkPolicyIngressRule{},
			})...,
		),

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow ingress from pods",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyLocal, TagIngress, TagPodsByLabel),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortLocal, DestinationNode: DestinationPodNode},
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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: deny all egress",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyLocal, TagEgress, TagDenyAll),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortLocal, DestinationNode: DestinationPodNode},
			createPolicyInAllNamespaces("deny-egress", NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []PolicyType{PolicyTypeEgress},
				Egress:      []NetworkPolicyEgressRule{},
			})...,
		),

		// FIXME Discrepancy found:40 wrong, 9 ignored, 32 correct
		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow egress to pods",
			NewStringSet(TagNodePort, TagExternalTrafficPolicyLocal, TagEgress, TagPodsByLabel),
			&ProbeConfig{AllAvailable: true, Mode: ProbeModeServiceName, Service: NodePortLocal, DestinationNode: DestinationPodNode},
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
