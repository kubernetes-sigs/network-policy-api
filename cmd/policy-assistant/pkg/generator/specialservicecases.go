package generator

import (
	. "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test cases for traffic to LoadBalancer and NodePort services (for both externalTrafficPolicy values).
//
// For LoadBalancer tests, tests send traffic from each Pod to each LoadBalancer service's external IP.
// The source Pod's node is expected to "intercept" this traffic and redirect to the one backend Pod (which may be on another node).
// In the case of etp=Local, the backend Pod must be on the source node. Therefore, tests ignore inter-node probes.
// In the case of etp=Cluster, if the backend Pod is on another node, the CNI will SNAT traffic to the node IP.
// For some NetworkPolicy implementations, the source Pod info is lost in this case; to allow traffic requires a policy that allows traffic from the source node.
//
// For NodePort tests, tests send traffic from each Pod to each NodePort service's node port.
// In the case of etp=Local, the tests send traffic to the destination Pod's node.
// In the case of etp=Cluster, the tests send traffic to the source Pod's node (since this etp supports redirecting traffic to another node).
//
// Example runs for two NetworkPolicy implementations:
// Cilium:
//   policy-assistant generate --pod-creation-timeout-seconds 600 --server-protocol TCP,UDP --ignore-loopback --include special-services --exclude ip-block-no-except
//
// Azure NPM:
//   policy-assistant generate --pod-creation-timeout-seconds 600 --server-protocol TCP,UDP --ignore-loopback --include special-services --exclude cni-brings-source-pod-info-to-other-node

// TODO: it would be good in the future to add unit tests for these special services. Some relevant code has already been added in the mock kubernetes.

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

		NewSingleStepTestCase("NodePort with externalTrafficPolicy=Local: allow egress to pods and nodes (to destination pod's node)",
			tags(baseTags, TagEgress, TagAllNamespaces, TagIPBlockNoExcept),
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
