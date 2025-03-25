package probe

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/generator"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/kube"
)

const (
	agnhostImage = "e2e-test-images/agnhost:2.43"
	// FIXME use a real image repository
	policyAssistantWorkerImage = "docker.io/policy-assistant-worker:latest"
)

func NewPod(ns string, name string, labels map[string]string, ip string, containers []*Container) *Pod {
	return &Pod{
		Namespace:  ns,
		Name:       name,
		Labels:     labels,
		IP:         ip,
		Containers: containers,
	}
}

func NewDefaultPod(ns string, name string, ports []int, protocols []v1.Protocol, batchJobs bool, imageRegistry string) *Pod {
	var containers []*Container
	for _, port := range ports {
		for _, protocol := range protocols {
			containers = append(containers, NewDefaultContainer(port, protocol, batchJobs, imageRegistry))
		}
	}
	return &Pod{
		Namespace:  ns,
		Name:       name,
		Labels:     map[string]string{"pod": name},
		IP:         "TODO",
		Containers: containers,
	}
}

type Pod struct {
	Namespace          string
	Name               string
	Labels             map[string]string
	ServiceIP          string
	IP                 string
	Containers         []*Container
	ExternalServiceIPs map[generator.ServiceKind]string
	// nodePorts maps service kind to target port to node port. Assumes that different protocols on the same target port share the same node port.
	nodePorts map[generator.ServiceKind]map[int]int
	// LocalNodeIP should be set to the IP of the node that the pod is running on
	LocalNodeIP string
	// RemoteNodeIP should be set to the IP of any remote node which another cyclonus pod is running on
	// This value is currently unused but might be used for future tests.
	RemoteNodeIP string
	// TODO populate in future if needed for AdminNetPol node selector
	NodeLabels map[string]string
}

func (p *Pod) Host(config *generator.ProbeConfig, srcPod *Pod) string {
	switch config.Service {
	case generator.NodePortLocal, generator.NodePortCluster:
		switch config.DestinationNode {
		case generator.ToSourcePodNode:
			return srcPod.LocalNodeIP
		case generator.ToDestinationPodNode:
			return p.LocalNodeIP
		default:
			panic(errors.Errorf("invalid node port mode %s", config.DestinationNode))
		}
	case generator.LoadBalancerLocal, generator.LoadBalancerCluster:
		return p.ExternalServiceIPs[config.Service]
	case generator.ClusterIP:
		switch config.Mode {
		case generator.ProbeModeServiceName:
			return kube.QualifiedServiceAddress(p.ServiceName(config.Service), p.Namespace)
		case generator.ProbeModePodIP:
			return p.IP
		case generator.ProbeModeServiceIP:
			return p.ServiceIP
		default:
			panic(errors.Errorf("invalid mode %s", config.Mode))
		}
	default:
		panic(errors.Errorf("invalid service kind %s", config.Service))
	}
}

func (p *Pod) IsEqualToKubePod(kubePod v1.Pod) (string, bool) {
	kubeConts := kubePod.Spec.Containers
	if len(kubeConts) != len(p.Containers) {
		return fmt.Sprintf("have %d containers, expected %d", len(p.Containers), len(kubeConts)), false
	}
	for i, kubeCont := range kubeConts {
		cont := p.Containers[i]
		if len(kubeCont.Ports) != 1 {
			return fmt.Sprintf("container %d: expected 1 port, found %d", i, len(kubeCont.Ports)), false
		}
		if int(kubeCont.Ports[0].ContainerPort) != cont.Port {
			return fmt.Sprintf("container %d: expected port %d, found %d", i, cont.Port, kubeCont.Ports[0].ContainerPort), false
		}
		if kubeCont.Ports[0].Protocol != cont.Protocol {
			return fmt.Sprintf("container %d: expected protocol %s, found %s", i, cont.Protocol, kubeCont.Ports[0].Protocol), false
		}
	}

	return "", true
}

func (p *Pod) ServiceName(kind generator.ServiceKind) string {
	switch kind {
	case generator.ClusterIP:
		return fmt.Sprintf("s-%s-%s", p.Namespace, p.Name)
	case generator.NodePortLocal:
		return fmt.Sprintf("s-%s-%s-nodeport-local", p.Namespace, p.Name)
	case generator.LoadBalancerLocal:
		return fmt.Sprintf("s-%s-%s-loadbalancer-local", p.Namespace, p.Name)
	case generator.NodePortCluster:
		return fmt.Sprintf("s-%s-%s-nodeport-cluster", p.Namespace, p.Name)
	case generator.LoadBalancerCluster:
		return fmt.Sprintf("s-%s-%s-loadbalancer-cluster", p.Namespace, p.Name)
	default:
		panic(errors.Errorf("invalid service kind %s", kind))
	}
}

func (p *Pod) KubePod() *v1.Pod {
	zero := int64(0)
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Labels:    p.Labels,
			Namespace: p.Namespace,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: &zero,
			Containers:                    p.KubeContainers(),
		},
	}
}

func (p *Pod) KubeService(kind generator.ServiceKind) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.ServiceName(kind),
			Namespace: p.Namespace,
		},
		Spec: v1.ServiceSpec{
			Ports:    slice.Map(func(cont *Container) v1.ServicePort { return cont.KubeServicePort() }, p.Containers),
			Selector: p.Labels,
		},
	}

	switch kind {
	case generator.ClusterIP:
		svc.Spec.Type = v1.ServiceTypeClusterIP
	case generator.NodePortLocal:
		svc.Spec.Type = v1.ServiceTypeNodePort
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
	case generator.NodePortCluster:
		svc.Spec.Type = v1.ServiceTypeNodePort
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
	case generator.LoadBalancerLocal:
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		// disable node port allocation
		svc.Spec.AllocateLoadBalancerNodePorts = new(bool)
		// optimization to not use public IPs on azure
		svc.ObjectMeta.Annotations = map[string]string{
			"service.beta.kubernetes.io/azure-load-balancer-internal": "true",
		}
	case generator.LoadBalancerCluster:
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeCluster
		// node port allocation is required
		// optimization to not use public IPs on azure
		svc.ObjectMeta.Annotations = map[string]string{
			"service.beta.kubernetes.io/azure-load-balancer-internal": "true",
		}
	}

	return svc
}

func (p *Pod) KubeContainers() []v1.Container {
	return slice.Map(func(cont *Container) v1.Container { return cont.KubeContainer() }, p.Containers)
}

func (p *Pod) ResolveNamedPort(port string) (int, error) {
	for _, c := range p.Containers {
		if c.PortName == port {
			return c.Port, nil
		}
	}
	return 0, errors.Errorf("unable to resolve named port %s on pod %s/%s", port, p.Namespace, p.Name)
}

func (p *Pod) ResolveNumberedPort(port int) (string, error) {
	for _, c := range p.Containers {
		if c.Port == port {
			return c.PortName, nil
		}
	}
	return "", errors.Errorf("unable to resolve numbered port %d on pod %s/%s", port, p.Namespace, p.Name)
}

func (p *Pod) IsServingPortProtocol(port int, protocol v1.Protocol) bool {
	for _, cont := range p.Containers {
		if cont.Port == port && cont.Protocol == protocol {
			return true
		}
	}
	return false
}

func (p *Pod) SetLabels(labels map[string]string) *Pod {
	return &Pod{
		Namespace:  p.Namespace,
		Name:       p.Name,
		Labels:     labels,
		IP:         p.IP,
		Containers: p.Containers,
	}
}

func (p *Pod) PodString() PodString {
	return NewPodString(p.Namespace, p.Name)
}

func (p *Pod) NodePort(svc generator.ServiceKind, port int) int {
	if len(p.nodePorts) == 0 || len(p.nodePorts[svc]) == 0 {
		return 0
	}
	return p.nodePorts[svc][port]
}

type Container struct {
	Name          string
	Port          int
	Protocol      v1.Protocol
	PortName      string
	BatchJobs     bool
	ImageRegistry string
}

func NewDefaultContainer(port int, protocol v1.Protocol, batchJobs bool, imageRegistry string) *Container {
	return &Container{
		Name:          fmt.Sprintf("cont-%d-%s", port, strings.ToLower(string(protocol))),
		Port:          port,
		Protocol:      protocol,
		PortName:      fmt.Sprintf("serve-%d-%s", port, strings.ToLower(string(protocol))),
		BatchJobs:     batchJobs,
		ImageRegistry: imageRegistry,
	}
}

func (c *Container) KubeServicePort() v1.ServicePort {
	return v1.ServicePort{
		Name:     fmt.Sprintf("service-port-%s-%d", strings.ToLower(string(c.Protocol)), c.Port),
		Protocol: c.Protocol,
		Port:     int32(c.Port),
	}
}

func (c *Container) Image() string {
	if c.BatchJobs {
		return policyAssistantWorkerImage
	}
	return c.ImageRegistry + "/" + agnhostImage
}

func (c *Container) KubeContainer() v1.Container {
	var cmd []string
	var env []v1.EnvVar

	switch c.Protocol {
	case v1.ProtocolTCP:
		cmd = []string{"/agnhost", "serve-hostname", "--tcp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolUDP:
		cmd = []string{"/agnhost", "serve-hostname", "--udp", "--http=false", "--port", fmt.Sprintf("%d", c.Port)}
	case v1.ProtocolSCTP:
		//cmd = []string{"/agnhost", "netexec", "--sctp-port", fmt.Sprintf("%d", c.Port)}
		env = append(env, v1.EnvVar{
			Name:  fmt.Sprintf("SERVE_SCTP_PORT_%d", c.Port),
			Value: "foo",
		})
		cmd = []string{"/agnhost", "porter"}
	default:
		panic(errors.Errorf("invalid protocol %s", c.Protocol))
	}
	return v1.Container{
		Name:            c.Name,
		ImagePullPolicy: v1.PullIfNotPresent,
		Image:           c.Image(),
		Command:         cmd,
		Env:             env,
		SecurityContext: &v1.SecurityContext{},
		Ports: []v1.ContainerPort{
			{
				ContainerPort: int32(c.Port),
				Name:          c.PortName,
				Protocol:      c.Protocol,
			},
		},
	}
}
