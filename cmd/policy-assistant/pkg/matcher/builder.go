package matcher

import (
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

func BuildNetworkPolicies(simplify bool, netpols []*networkingv1.NetworkPolicy) *Policy {
	return BuildV1AndV2NetPols(simplify, netpols, nil, nil)
}

func BuildV1AndV2NetPols(simplify bool, netpols []*networkingv1.NetworkPolicy, anps []*v1alpha1.AdminNetworkPolicy, banp *v1alpha1.BaselineAdminNetworkPolicy) *Policy {
	np := NewPolicy()
	for _, p := range netpols {
		ingress, egress := BuildTarget(p)
		np.AddTarget(true, ingress)
		np.AddTarget(false, egress)
	}

	priorities := make(map[int32]struct{})
	for _, p := range anps {
		if _, ok := priorities[p.Spec.Priority]; ok {
			panic(errors.Errorf("duplicate priorities are undefined. priority: %d", p.Spec.Priority))
		}
		priorities[p.Spec.Priority] = struct{}{}

		ingress, egress := BuildTargetANP(p)
		np.AddTarget(true, ingress)
		np.AddTarget(false, egress)
	}

	if banp != nil {
		// there can only be one BANP by definition
		ingress, egress := BuildTargetBANP(banp)
		np.AddTarget(true, ingress)
		np.AddTarget(false, egress)
	}

	if simplify {
		np.Simplify()
	}

	return np
}

func getPolicyNamespace(policy *networkingv1.NetworkPolicy) string {
	if policy.Namespace == "" {
		return v1.NamespaceDefault
	}
	return policy.Namespace
}

func BuildTarget(netpol *networkingv1.NetworkPolicy) (*Target, *Target) {
	var ingress *Target
	var egress *Target
	if len(netpol.Spec.PolicyTypes) == 0 {
		panic(errors.Errorf("invalid NetworkPolicy: need at least 1 type"))
	}
	policyNamespace := getPolicyNamespace(netpol)
	for _, pType := range netpol.Spec.PolicyTypes {
		switch pType {
		case networkingv1.PolicyTypeIngress:
			ingress = &Target{
				SubjectMatcher: NewSubjectV1(policyNamespace, netpol.Spec.PodSelector),
				SourceRules:    []NetPolID{netPolID(netpol)},
				Peers:          BuildIngressMatcher(policyNamespace, netpol.Spec.Ingress),
			}
		case networkingv1.PolicyTypeEgress:
			egress = &Target{
				SubjectMatcher: NewSubjectV1(policyNamespace, netpol.Spec.PodSelector),
				SourceRules:    []NetPolID{netPolID(netpol)},
				Peers:          BuildEgressMatcher(policyNamespace, netpol.Spec.Egress),
			}
		}
	}
	return ingress, egress
}

func BuildIngressMatcher(policyNamespace string, ingresses []networkingv1.NetworkPolicyIngressRule) []PeerMatcher {
	var matchers []PeerMatcher
	for _, ingress := range ingresses {
		matchers = append(matchers, BuildPeerMatcher(policyNamespace, ingress.Ports, ingress.From)...)
	}
	return matchers
}

func BuildEgressMatcher(policyNamespace string, egresses []networkingv1.NetworkPolicyEgressRule) []PeerMatcher {
	var matchers []PeerMatcher
	for _, egress := range egresses {
		matchers = append(matchers, BuildPeerMatcher(policyNamespace, egress.Ports, egress.To)...)
	}
	return matchers
}

func BuildPeerMatcher(policyNamespace string, npPorts []networkingv1.NetworkPolicyPort, peers []networkingv1.NetworkPolicyPeer) []PeerMatcher {
	if len(npPorts) == 0 && len(peers) == 0 {
		return []PeerMatcher{AllPeersPorts}
	}
	// 1. build port matcher
	port := BuildPortMatcher(npPorts)
	// 2. build Peers
	if len(peers) == 0 {
		return []PeerMatcher{&PortsForAllPeersMatcher{Port: port}}
	}

	var matchers []PeerMatcher
	for _, from := range peers {
		ip, ns, pod := BuildIPBlockNamespacePodMatcher(policyNamespace, from)
		// invalid netpol guards
		if ip == nil && ns == nil && pod == nil {
			panic(errors.Errorf("invalid NetworkPolicyPeer: all of IPBlock, NamespaceSelector, and PodSelector are nil"))
		}
		if ip != nil && (ns != nil || pod != nil) {
			panic(errors.Errorf("invalid NetworkPolicyPeer: if NamespaceSelector or PodSelector is non-nil, IPBlock must be nil"))
		}
		// process a valid netpol
		if ip != nil {
			ip.Port = port
			matchers = append(matchers, ip)
		} else {
			matchers = append(matchers, &PodPeerMatcher{
				Namespace: ns,
				Pod:       pod,
				Port:      port,
			})
		}
	}
	return matchers
}

func BuildIPBlockNamespacePodMatcher(policyNamespace string, peer networkingv1.NetworkPolicyPeer) (*IPPeerMatcher, NamespaceMatcher, PodMatcher) {
	if peer.IPBlock != nil {
		return &IPPeerMatcher{
			IPBlock: peer.IPBlock,
			Port:    nil, // remember to set this elsewhere!
		}, nil, nil
	}

	podSel := peer.PodSelector
	var podMatcher PodMatcher
	if podSel == nil || kube.IsLabelSelectorEmpty(*podSel) {
		podMatcher = &AllPodMatcher{}
	} else {
		podMatcher = &LabelSelectorPodMatcher{Selector: *podSel}
	}

	nsSel := peer.NamespaceSelector
	var nsMatcher NamespaceMatcher
	if nsSel == nil {
		nsMatcher = &ExactNamespaceMatcher{Namespace: policyNamespace}
	} else if kube.IsLabelSelectorEmpty(*nsSel) {
		nsMatcher = &AllNamespaceMatcher{}
	} else {
		nsMatcher = &LabelSelectorNamespaceMatcher{Selector: *nsSel}
	}

	return nil, nsMatcher, podMatcher
}

func BuildPortMatcher(npPorts []networkingv1.NetworkPolicyPort) PortMatcher {
	if len(npPorts) == 0 {
		return &AllPortMatcher{}
	} else {
		matcher := &SpecificPortMatcher{}
		for _, p := range npPorts {
			singlePort, portRange := BuildSinglePortMatcher(p)
			if singlePort != nil {
				matcher.Ports = append(matcher.Ports, singlePort)
			} else {
				matcher.PortRanges = append(matcher.PortRanges, portRange)
			}
		}
		return matcher
	}
}

func BuildSinglePortMatcher(npPort networkingv1.NetworkPolicyPort) (*PortProtocolMatcher, *PortRangeMatcher) {
	protocol := v1.ProtocolTCP
	if npPort.Protocol != nil {
		protocol = *npPort.Protocol
	}
	if npPort.EndPort == nil {
		return &PortProtocolMatcher{
			Port:     npPort.Port,
			Protocol: protocol,
		}, nil
	}
	// we have a port range: make sure it's valid
	if npPort.Port == nil {
		panic(errors.Errorf("invalid port range: start port is nil"))
	}
	if npPort.Port.Type == intstr.String {
		panic(errors.Errorf("invalid port range: start port is string"))
	}
	if *npPort.EndPort < npPort.Port.IntVal {
		panic(errors.Errorf("invalid port range: end port < start port"))
	}
	return nil, &PortRangeMatcher{
		From:     int(npPort.Port.IntVal),
		To:       int(*npPort.EndPort),
		Protocol: protocol,
	}
}

func BuildTargetANP(anp *v1alpha1.AdminNetworkPolicy) (*Target, *Target) {
	if len(anp.Spec.Ingress) == 0 && len(anp.Spec.Egress) == 0 {
		panic(errors.Errorf("invalid AdminNetworkPolicy: need at least one egress or ingress rule"))
	}

	var ingress *Target
	var egress *Target

	if len(anp.Spec.Ingress) > 0 {
		ingress = &Target{
			SubjectMatcher: NewSubjectAdmin(&anp.Spec.Subject),
			SourceRules:    []NetPolID{netPolID(anp)},
		}

		for _, r := range anp.Spec.Ingress {
			v := AdminActionToVerdict(r.Action)
			matchers := BuildPeerMatcherAdmin(r.From, r.Ports)
			for _, m := range matchers {
				matcherAdmin := NewPeerMatcherANP(m, v, int(anp.Spec.Priority), anp.Name)
				ingress.Peers = append(ingress.Peers, matcherAdmin)
			}
		}
	}

	if len(anp.Spec.Egress) > 0 {
		egress = &Target{
			SubjectMatcher: NewSubjectAdmin(&anp.Spec.Subject),
			SourceRules:    []NetPolID{netPolID(anp)},
		}

		for _, r := range anp.Spec.Egress {
			v := AdminActionToVerdict(r.Action)
			matchers := BuildPeerMatcherAdmin(r.To, r.Ports)
			for _, m := range matchers {
				matcherAdmin := NewPeerMatcherANP(m, v, int(anp.Spec.Priority), anp.Name)
				egress.Peers = append(egress.Peers, matcherAdmin)
			}
		}
	}

	return ingress, egress
}

func BuildTargetBANP(banp *v1alpha1.BaselineAdminNetworkPolicy) (*Target, *Target) {
	if len(banp.Spec.Ingress) == 0 && len(banp.Spec.Egress) == 0 {
		panic(errors.Errorf("invalid BaselineAdminNetworkPolicy: need at least one egress or ingress rule"))
	}

	var ingress *Target
	var egress *Target

	if len(banp.Spec.Ingress) > 0 {
		ingress = &Target{
			SubjectMatcher: NewSubjectAdmin(&banp.Spec.Subject),
			SourceRules:    []NetPolID{netPolID(banp)},
		}

		for _, r := range banp.Spec.Ingress {
			v := BaselineAdminActionToVerdict(r.Action)
			matchers := BuildPeerMatcherAdmin(r.From, r.Ports)
			for _, m := range matchers {
				matcherAdmin := NewPeerMatcherBANP(m, v, banp.Name)
				ingress.Peers = append(ingress.Peers, matcherAdmin)
			}
		}
	}

	if len(banp.Spec.Egress) > 0 {
		egress = &Target{
			SubjectMatcher: NewSubjectAdmin(&banp.Spec.Subject),
			SourceRules:    []NetPolID{netPolID(banp)},
		}

		for _, r := range banp.Spec.Egress {
			v := BaselineAdminActionToVerdict(r.Action)
			matchers := BuildPeerMatcherAdmin(r.To, r.Ports)
			for _, m := range matchers {
				matcherAdmin := NewPeerMatcherBANP(m, v, banp.Name)
				egress.Peers = append(egress.Peers, matcherAdmin)
			}
		}
	}

	return ingress, egress
}

func BuildPeerMatcherAdmin(peers []v1alpha1.AdminNetworkPolicyPeer, ports *[]v1alpha1.AdminNetworkPolicyPort) []*PodPeerMatcher {
	if len(peers) == 0 {
		panic(errors.Errorf("invalid admin to/from field: must have at least one peer"))
	}

	// 1. build port matcher
	var portMatcher PortMatcher
	if ports == nil {
		portMatcher = BuildPortMatcherAdmin(nil)
	} else {
		portMatcher = BuildPortMatcherAdmin(*ports)
	}

	// 2. build Peers
	var peerMatchers []*PodPeerMatcher
	for _, peer := range peers {
		if (peer.Namespaces == nil && peer.Pods == nil) || (peer.Namespaces != nil && peer.Pods != nil) {
			panic(errors.Errorf("invalid admin peer: must have exactly one of Namespaces or Pods"))
		}

		var ns v1alpha1.NamespacedPeer
		var podMatcher PodMatcher
		if peer.Pods != nil {
			ns = peer.Pods.Namespaces

			// TODO account for Tenancy or Pod same/not-same labels when these become features
			podSel := peer.Pods.PodSelector
			if kube.IsLabelSelectorEmpty(podSel) {
				podMatcher = &AllPodMatcher{}
			} else {
				podMatcher = &LabelSelectorPodMatcher{Selector: podSel}
			}
		} else {
			// peer.Namespaces is non-nil
			ns = *peer.Namespaces
			podMatcher = &AllPodMatcher{}
		}

		nonNilCount := 0
		if ns.NamespaceSelector != nil {
			nonNilCount++
		}
		if ns.SameLabels != nil {
			nonNilCount++
		}
		if ns.NotSameLabels != nil {
			nonNilCount++
		}
		if nonNilCount != 1 {
			panic(errors.Errorf("invalid admin peer: must have exactly one of NamespaceSelector, SameLabels, or NotSameLabels"))
		}

		var nsMatcher NamespaceMatcher
		if ns.NamespaceSelector != nil {
			if kube.IsLabelSelectorEmpty(*ns.NamespaceSelector) {
				nsMatcher = &AllNamespaceMatcher{}
			} else {
				nsMatcher = &LabelSelectorNamespaceMatcher{
					Selector: *ns.NamespaceSelector,
				}
			}
		} else if ns.SameLabels != nil {
			nsMatcher = &SameLabelsNamespaceMatcher{
				labels: ns.SameLabels,
			}
		} else {
			// ns.NotSameLabels is non-nil
			nsMatcher = &NotSameLabelsNamespaceMatcher{
				labels: ns.NotSameLabels,
			}
		}

		m := &PodPeerMatcher{
			Namespace: nsMatcher,
			Pod:       podMatcher,
			Port:      portMatcher,
		}
		peerMatchers = append(peerMatchers, m)
	}

	return peerMatchers
}

func BuildPortMatcherAdmin(ports []v1alpha1.AdminNetworkPolicyPort) PortMatcher {
	if len(ports) == 0 {
		return &AllPortMatcher{}
	} else {
		matcher := &SpecificPortMatcher{}
		for _, p := range ports {
			singlePort, portRange := BuildSinglePortMatcherAdmin(p)
			if singlePort != nil {
				matcher.Ports = append(matcher.Ports, singlePort)
			} else {
				matcher.PortRanges = append(matcher.PortRanges, portRange)
			}
		}
		return matcher
	}
}

func BuildSinglePortMatcherAdmin(port v1alpha1.AdminNetworkPolicyPort) (*PortProtocolMatcher, *PortRangeMatcher) {
	nonNilCount := 0
	if port.PortNumber != nil {
		nonNilCount++
	}
	if port.NamedPort != nil {
		nonNilCount++
	}
	if port.PortRange != nil {
		nonNilCount++
	}
	if nonNilCount != 1 {
		panic(errors.Errorf("invalid port: must have exactly one of PortNumber, NamedPort, or PortRange"))
	}

	if port.PortNumber != nil {
		// default is TCP if protocol field is empty
		proto := v1.ProtocolTCP
		if port.PortNumber.Protocol != "" {
			proto = port.PortNumber.Protocol
		}

		m := &PortProtocolMatcher{
			Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: port.PortNumber.Port},
			Protocol: proto,
		}

		return m, nil
	}

	if port.NamedPort != nil {
		// determine protocol from the named port name
		// e.g. "serve-80-tcp" -> TCP
		proto := v1.Protocol("Unknown")
		if endsIn(*port.NamedPort, "-udp") {
			proto = v1.ProtocolUDP
		} else if endsIn(*port.NamedPort, "-tcp") {
			proto = v1.ProtocolTCP
		} else if endsIn(*port.NamedPort, "-sctp") {
			proto = v1.ProtocolSCTP
		}

		m := &PortProtocolMatcher{
			Port:     &intstr.IntOrString{Type: intstr.String, StrVal: *port.NamedPort},
			Protocol: proto,
		}

		return m, nil
	}

	// port.PortRange is non-nil
	// default is TCP if protocol field is empty
	proto := v1.ProtocolTCP
	if port.PortRange.Protocol != "" {
		proto = port.PortRange.Protocol
	}

	if port.PortRange.Start >= port.PortRange.End {
		panic(errors.Errorf("invalid port range: start >= end"))
	}

	return nil, &PortRangeMatcher{
		From:     int(port.PortRange.Start),
		To:       int(port.PortRange.End),
		Protocol: proto,
	}
}

func endsIn(s string, suffix string) bool {
	return len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix
}
