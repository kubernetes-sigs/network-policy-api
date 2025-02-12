package probe

import (
	"time"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/generator"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/kube"
)

type Resources struct {
	Namespaces map[string]map[string]string
	Pods       []*Pod
	//ExternalIPs []string
	ports     []int
	protocols []v1.Protocol
}

func NewDefaultResources(kubernetes kube.IKubernetes, namespaces []string, podNames []string, ports []int, protocols []v1.Protocol,
	podCreationTimeoutSeconds int, batchJobs bool, imageRegistry string,
	services []generator.ServiceKind) (*Resources, error) {
	r := &Resources{
		Namespaces: map[string]map[string]string{},
		ports:      ports,
		protocols:  protocols,
	}

	for _, ns := range namespaces {
		for _, podName := range podNames {
			r.Pods = append(r.Pods, NewDefaultPod(ns, podName, ports, protocols, batchJobs, imageRegistry))
		}
		r.Namespaces[ns] = map[string]string{"ns": ns}
	}

	if err := r.CreateResourcesInKube(kubernetes, services); err != nil {
		return nil, err
	}
	if err := r.waitForPodsReady(kubernetes, podCreationTimeoutSeconds); err != nil {
		return nil, err
	}
	if err := r.getPodIPsFromKube(kubernetes); err != nil {
		return nil, err
	}
	if err := r.getNamespaceLabelsFromKube(kubernetes); err != nil {
		return nil, err
	}

	haveLoadBalancer := false
	haveNodePort := false
	for _, svc := range services {
		if svc == generator.LoadBalancerCluster || svc == generator.LoadBalancerLocal {
			haveLoadBalancer = true
		}
		if svc == generator.NodePortCluster || svc == generator.NodePortLocal {
			haveNodePort = true
		}
	}

	if haveLoadBalancer {
		if err := r.getExternalIPs(kubernetes, podCreationTimeoutSeconds); err != nil {
			return nil, err
		}
	}

	if haveNodePort {
		if err := r.getNodePorts(kubernetes); err != nil {
			return nil, err
		}

		if err := r.setRemoteNodeIPs(); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Resources) waitForPodsReady(kubernetes kube.IKubernetes, timeoutSeconds int) error {
	sleep := 5
	for i := 0; i < timeoutSeconds; i += sleep {
		podList, err := kube.GetPodsInNamespaces(kubernetes, r.NamespacesSlice())
		if err != nil {
			return err
		}

		ready := 0
		for _, pod := range podList {
			if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
				ready++
			}
		}
		if ready == len(r.Pods) {
			return nil
		}

		logrus.Infof("waiting for %d pods to be running and have IP addresses; currently %d are ready", len(r.Pods), ready)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return errors.Errorf("pods not ready")
}

func (r *Resources) getPodIPsFromKube(kubernetes kube.IKubernetes) error {
	podList, err := kube.GetPodsInNamespaces(kubernetes, r.NamespacesSlice())
	if err != nil {
		return err
	}

	for _, kubePod := range podList {
		if kubePod.Status.PodIP == "" {
			return errors.Errorf("no ip found for pod %s/%s", kubePod.Namespace, kubePod.Name)
		}

		pod, err := r.GetPod(kubePod.Namespace, kubePod.Name)
		if err != nil {
			return errors.Errorf("unable to find pod %s/%s in resources", kubePod.Namespace, kubePod.Name)
		}
		pod.IP = kubePod.Status.PodIP
		if pod.IP == "" {
			return errors.Errorf("empty ip for pod %s/%s", kubePod.Namespace, kubePod.Name)
		}
		logrus.Debugf("ip for pod %s/%s: %s", pod.Namespace, pod.Name, pod.IP)

		pod.LocalNodeIP = kubePod.Status.HostIP
		if pod.LocalNodeIP == "" {
			return errors.Errorf("empty node ip for pod %s/%s", pod.Namespace, pod.Name)
		}
		logrus.Debugf("node ip for pod %s/%s: %s", pod.Namespace, pod.Name, pod.LocalNodeIP)

		kubeService, err := kubernetes.GetService(pod.Namespace, pod.ServiceName(generator.ClusterIP))
		if err != nil {
			return errors.Errorf("unable to get service %s/%s", pod.Namespace, pod.ServiceName(generator.ClusterIP))
		}
		pod.ServiceIP = kubeService.Spec.ClusterIP
		if pod.ServiceIP == "" {
			return errors.Errorf("empty service ip for pod %s/%s", pod.Namespace, pod.Name)
		}
		logrus.Debugf("service ip for pod %s/%s: %s", pod.Namespace, pod.Name, pod.ServiceIP)
	}

	return nil
}

func (r *Resources) getExternalIPs(kubernetes kube.IKubernetes, maxRetrySeconds int) error {
	waitBetweenRetries := time.Duration(5) * time.Second
	timeout := time.After(time.Duration(maxRetrySeconds) * time.Second)
	ticker := time.NewTicker(waitBetweenRetries)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return errors.Errorf("timed out waiting for external IPs")
		case <-ticker.C:
			allFound := true
			for _, pod := range r.Pods {
				svcKinds := []generator.ServiceKind{generator.LoadBalancerLocal, generator.LoadBalancerCluster}
				if len(pod.ExternalServiceIPs) == len(svcKinds) {
					continue
				}

				for _, kind := range svcKinds {
					svcName := pod.ServiceName(kind)
					kubeService, err := kubernetes.GetService(pod.Namespace, svcName)
					if err != nil {
						logrus.Errorf("unable to get service. will retry. svc: %s/%s", pod.Namespace, svcName)
						allFound = false
						break
					}

					if len(kubeService.Status.LoadBalancer.Ingress) == 0 {
						allFound = false
						break
					}
					ip := kubeService.Status.LoadBalancer.Ingress[0].IP
					if ip == "" {
						allFound = false
						break
					}

					if pod.ExternalServiceIPs == nil {
						pod.ExternalServiceIPs = make(map[generator.ServiceKind]string, len(svcKinds))
					}
					pod.ExternalServiceIPs[kind] = ip
				}

				if !allFound {
					break
				}
			}

			if allFound {
				return nil
			}
		}
	}
}

func (r *Resources) getNodePorts(kubernetes kube.IKubernetes) error {
	for _, pod := range r.Pods {
		svcKinds := []generator.ServiceKind{generator.NodePortLocal, generator.NodePortCluster}
		for _, kind := range svcKinds {
			svcName := pod.ServiceName(kind)
			kubeService, err := kubernetes.GetService(pod.Namespace, svcName)
			if err != nil {
				return errors.Errorf("unable to get service %s/%s", pod.Namespace, svcName)
			}

			if len(kubeService.Spec.Ports) == 0 {
				return errors.Errorf("no ports found for service %s/%s", pod.Namespace, svcName)
			}

			if pod.NodePorts == nil {
				pod.NodePorts = make(map[generator.ServiceKind]map[int]int, len(svcKinds))
			}

			if pod.NodePorts[kind] == nil {
				pod.NodePorts[kind] = make(map[int]int, len(kubeService.Spec.Ports))
			}

			for _, port := range kubeService.Spec.Ports {
				if port.TargetPort.IntVal == 0 {
					return errors.Errorf("no target port found for service %s/%s", pod.Namespace, svcName)
				}

				if port.NodePort == 0 {
					return errors.Errorf("no node port found for service %s/%s", pod.Namespace, svcName)
				}

				pod.NodePorts[kind][int(port.TargetPort.IntVal)] = int(port.NodePort)
			}
		}
	}
	return nil
}

func (r *Resources) setRemoteNodeIPs() error {
	// set remote node IPs
	nodeIPs := make(map[string]struct{})
	for _, pod := range r.Pods {
		nodeIPs[pod.LocalNodeIP] = struct{}{}
	}

	if len(nodeIPs) == 1 {
		// TODO only check this there are nodeport tests
		return errors.New("all cyclonus pods were scheduled on one node. tests involving remote node IPs will fail. please provision more nodes")
	} else {
		for _, pod := range r.Pods {
			for nodeIP := range nodeIPs {
				if pod.LocalNodeIP != nodeIP {
					pod.RemoteNodeIP = nodeIP
					break
				}
			}
		}
	}

	return nil
}

func (r *Resources) getNamespaceLabelsFromKube(kubernetes kube.IKubernetes) error {
	nsList, err := kubernetes.GetAllNamespaces()
	if err != nil {
		return err
	}

	for _, kubeNs := range nsList.Items {
		for label, value := range kubeNs.Labels {
			if ns, ok := r.Namespaces[kubeNs.Name]; ok {
				ns[label] = value
			}
		}
	}

	return nil
}

func (r *Resources) GetNodeIPs() []string {
	nodeIPs := make(map[string]struct{})
	for _, pod := range r.Pods {
		nodeIPs[pod.LocalNodeIP] = struct{}{}
	}
	var ips []string
	for ip := range nodeIPs {
		ips = append(ips, ip)
	}
	return ips
}

func (r *Resources) GetPod(ns string, name string) (*Pod, error) {
	for _, pod := range r.Pods {
		if pod.Namespace == ns && pod.Name == name {
			return pod, nil
		}
	}
	return nil, errors.Errorf("unable to find pod %s/%s", ns, name)
}

// CreateNamespace returns a new object with a new namespace.  It should not affect the original Resources object.
func (r *Resources) CreateNamespace(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; ok {
		return nil, errors.Errorf("namespace %s already found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		newNamespaces[oldNs] = oldLabels
	}
	newNamespaces[ns] = labels
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       r.Pods,
	}, nil
}

// UpdateNamespaceLabels returns a new object with an updated namespace.  It should not affect the original Resources object.
func (r *Resources) UpdateNamespaceLabels(ns string, labels map[string]string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("namespace %s not found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		newNamespaces[oldNs] = oldLabels
	}
	newNamespaces[ns] = labels
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       r.Pods,
	}, nil
}

// DeleteNamespace returns a new object without the namespace.  It should not affect the original Resources object.
func (r *Resources) DeleteNamespace(ns string) (*Resources, error) {
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("namespace %s not found", ns)
	}
	newNamespaces := map[string]map[string]string{}
	for oldNs, oldLabels := range r.Namespaces {
		if oldNs != ns {
			newNamespaces[oldNs] = oldLabels
		}
	}
	var pods []*Pod
	for _, pod := range r.Pods {
		if pod.Namespace == ns {
			// skip
		} else {
			pods = append(pods, pod)
		}
	}
	return &Resources{
		Namespaces: newNamespaces,
		Pods:       pods,
	}, nil
}

// CreatePod returns a new object with a new pod.  It should not affect the original Resources object.
func (r *Resources) CreatePod(ns string, podName string, labels map[string]string) (*Resources, error) {
	// TODO this needs to be improved
	//   for now, let's assume all pods have the same containers and just copy the containers from the first pod
	if _, ok := r.Namespaces[ns]; !ok {
		return nil, errors.Errorf("can't find namespace %s", ns)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       append(append([]*Pod{}, r.Pods...), NewPod(ns, podName, labels, "TODO", r.Pods[0].Containers)),
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

// SetPodLabels returns a new object with an updated pod.  It should not affect the original Resources object.
func (r *Resources) SetPodLabels(ns string, podName string, labels map[string]string) (*Resources, error) {
	var pods []*Pod
	found := false
	for _, existingPod := range r.Pods {
		if existingPod.Namespace == ns && existingPod.Name == podName {
			found = true
			pods = append(pods, existingPod.SetLabels(labels))
		} else {
			pods = append(pods, existingPod)
		}
	}
	if !found {
		return nil, errors.Errorf("no pod named %s/%s found", ns, podName)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       pods,
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

// DeletePod returns a new object without the deleted pod.  It should not affect the original Resources object.
func (r *Resources) DeletePod(ns string, podName string) (*Resources, error) {
	var newPods []*Pod
	found := false
	for _, pod := range r.Pods {
		if pod.Namespace == ns && pod.Name == podName {
			found = true
		} else {
			newPods = append(newPods, pod)
		}
	}
	if !found {
		return nil, errors.Errorf("pod %s/%s not found", ns, podName)
	}
	return &Resources{
		Namespaces: r.Namespaces,
		Pods:       newPods,
		//ExternalIPs: r.ExternalIPs,
	}, nil
}

func (r *Resources) SortedPodNames() []string {
	return slice.Sort(slice.Map(
		func(p *Pod) string { return p.PodString().String() },
		r.Pods))
}

func (r *Resources) NamespacesSlice() []string {
	return maps.Keys(r.Namespaces)
}

func (r *Resources) CreateResourcesInKube(kubernetes kube.IKubernetes, services []generator.ServiceKind) error {
	for ns, labels := range r.Namespaces {
		_, err := kubernetes.GetNamespace(ns)
		if err != nil {
			_, err := kubernetes.CreateNamespace(KubeNamespace(ns, labels))
			if err != nil {
				return err
			}
		}
	}

	for _, pod := range r.Pods {
		_, err := kubernetes.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			_, err := kubernetes.CreatePod(pod.KubePod())
			if err != nil {
				return err
			}
		}

		for _, kindStr := range services {
			kind := generator.ServiceKind(kindStr)
			svc := pod.KubeService(kind)
			// not sure why we get the service here but not in TestCaseState.CreatePod()
			_, err = kubernetes.GetService(svc.Namespace, svc.Name)
			if err != nil {
				_, err = kubernetes.CreateService(svc)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func KubeNamespace(ns string, labels map[string]string) *v1.Namespace {
	return &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns, Labels: labels}}
}
