package kube

import (
	"context"
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

type IKubernetes interface {
	CreateNamespace(kubeNamespace *v1.Namespace) (*v1.Namespace, error)
	GetNamespace(namespace string) (*v1.Namespace, error)
	SetNamespaceLabels(namespace string, labels map[string]string) (*v1.Namespace, error)
	DeleteNamespace(namespace string) error
	GetAllNamespaces() (*v1.NamespaceList, error)

	CreateNetworkPolicy(kubePolicy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error)
	GetNetworkPoliciesInNamespace(ctx context.Context, namespace string) ([]networkingv1.NetworkPolicy, error)
	UpdateNetworkPolicy(kubePolicy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error)
	DeleteNetworkPolicy(namespace string, name string) error
	DeleteAllNetworkPoliciesInNamespace(namespace string) error

	CreateService(kubeService *v1.Service) (*v1.Service, error)
	GetService(namespace string, name string) (*v1.Service, error)
	DeleteService(namespace string, name string) error
	GetServicesInNamespace(namespace string) ([]v1.Service, error)

	GetAdminNetworkPolicies(ctx context.Context) ([]v1alpha1.AdminNetworkPolicy, error)
	CreateAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.AdminNetworkPolicy) (*v1alpha1.AdminNetworkPolicy, error)
	UpdateAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.AdminNetworkPolicy) (*v1alpha1.AdminNetworkPolicy, error)
	DeleteAdminNetworkPolicy(ctx context.Context, name string) error

	GetBaselineAdminNetworkPolicy(ctx context.Context) (*v1alpha1.BaselineAdminNetworkPolicy, error)
	CreateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.BaselineAdminNetworkPolicy) (*v1alpha1.BaselineAdminNetworkPolicy, error)
	UpdateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.BaselineAdminNetworkPolicy) (*v1alpha1.BaselineAdminNetworkPolicy, error)
	DeleteBaselineAdminNetworkPolicy(ctx context.Context, name string) error

	CreatePod(kubePod *v1.Pod) (*v1.Pod, error)
	GetPod(namespace string, pod string) (*v1.Pod, error)
	DeletePod(namespace string, pod string) error
	SetPodLabels(namespace string, pod string, labels map[string]string) (*v1.Pod, error)
	GetPodsInNamespace(namespace string) ([]v1.Pod, error)

	ExecuteRemoteCommand(namespace string, pod string, container string, command []string) (string, string, error, error)
}

func GetNetworkPoliciesInNamespaces(ctx context.Context, kubernetes IKubernetes, namespaces []string) ([]networkingv1.NetworkPolicy, error) {
	var allNetpols []networkingv1.NetworkPolicy
	for _, ns := range namespaces {
		netpols, err := kubernetes.GetNetworkPoliciesInNamespace(ctx, ns)
		if err != nil {
			return nil, err
		}
		allNetpols = append(allNetpols, netpols...)
	}
	return allNetpols, nil
}

func DeleteAllNetworkPoliciesInNamespaces(kubernetes IKubernetes, namespaces []string) error {
	for _, ns := range namespaces {
		err := kubernetes.DeleteAllNetworkPoliciesInNamespace(ns)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetPodsInNamespaces(kubernetes IKubernetes, namespaces []string) ([]v1.Pod, error) {
	var allPods []v1.Pod
	for _, ns := range namespaces {
		pods, err := kubernetes.GetPodsInNamespace(ns)
		if err != nil {
			return nil, err
		}
		allPods = append(allPods, pods...)
	}
	return allPods, nil
}

func GetServicesInNamespaces(kubernetes IKubernetes, namespaces []string) ([]v1.Service, error) {
	var allServices []v1.Service
	for _, ns := range namespaces {
		svcs, err := kubernetes.GetServicesInNamespace(ns)
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, svcs...)
	}
	return allServices, nil
}

func GetAdminNetworkPolicies(ctx context.Context, kubernetes IKubernetes) ([]v1alpha1.AdminNetworkPolicy, error) {
	return kubernetes.GetAdminNetworkPolicies(ctx)
}

func GetBaselineAdminNetworkPolicy(ctx context.Context, kubernetes IKubernetes) (*v1alpha1.BaselineAdminNetworkPolicy, error) {
	return kubernetes.GetBaselineAdminNetworkPolicy(ctx)
}

type MockNamespace struct {
	NamespaceObject *v1.Namespace
	Netpols         map[string]*networkingv1.NetworkPolicy
	Pods            map[string]*v1.Pod
	Services        map[string]*v1.Service
}

type MockKubernetes struct {
	AdminNetworkPolicies        []v1alpha1.AdminNetworkPolicy
	AdminNetworkPolicyError     error
	BaselineNetworkPolicy       *v1alpha1.BaselineAdminNetworkPolicy
	BaseAdminNetworkPolicyError error
	Namespaces                  map[string]*MockNamespace
	NetworkPolicyError          error
	passRate                    float64
	podID                       int
}

func NewMockKubernetes(passRate float64) *MockKubernetes {
	return &MockKubernetes{
		Namespaces: map[string]*MockNamespace{},
		passRate:   passRate,
		podID:      1,
	}
}

func (m *MockKubernetes) getNamespaceObject(namespace string) (*MockNamespace, error) {
	if ns, ok := m.Namespaces[namespace]; ok {
		return ns, nil
	}
	return nil, errors.Errorf("namespace %s not found", namespace)
}

func (m *MockKubernetes) GetNamespace(namespace string) (*v1.Namespace, error) {
	if ns, ok := m.Namespaces[namespace]; ok {
		labels := map[string]string{}
		for k, v := range ns.NamespaceObject.Labels {
			labels[k] = v
		}
		labels[DefaultNamespaceLabel] = namespace
		return &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ns.NamespaceObject.Name,
				Labels: labels,
			},
		}, nil
	}
	return nil, errors.Errorf("namespace %s not found", namespace)
}

func (m *MockKubernetes) SetNamespaceLabels(namespace string, labels map[string]string) (*v1.Namespace, error) {
	ns, err := m.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	ns.Labels = labels
	return ns, nil
}

func (m *MockKubernetes) DeleteNamespace(ns string) error {
	if _, ok := m.Namespaces[ns]; !ok {
		return errors.Errorf("namespace %s not found", ns)
	}
	delete(m.Namespaces, ns)
	return nil
}

func (m *MockKubernetes) GetAllNamespaces() (*v1.NamespaceList, error) {
	var namespaces []v1.Namespace
	for name := range m.Namespaces {
		ns, err := m.GetNamespace(name)
		utils.DoOrDie(err)
		namespaces = append(namespaces, *ns)
	}
	return &v1.NamespaceList{
		Items: namespaces,
	}, nil
}

func (m *MockKubernetes) CreateNamespace(ns *v1.Namespace) (*v1.Namespace, error) {
	if _, ok := m.Namespaces[ns.Name]; ok {
		return nil, errors.Errorf("namespace %s already present", ns.Name)
	}
	m.Namespaces[ns.Name] = &MockNamespace{
		NamespaceObject: ns,
		Netpols:         map[string]*networkingv1.NetworkPolicy{},
		Pods:            map[string]*v1.Pod{},
		Services:        map[string]*v1.Service{},
	}
	return ns, nil
}

func (m *MockKubernetes) DeleteAllNetworkPoliciesInNamespace(ns string) error {
	nsObject, err := m.getNamespaceObject(ns)
	if err != nil {
		return err
	}
	nsObject.Netpols = map[string]*networkingv1.NetworkPolicy{}
	return nil
}

func (m *MockKubernetes) DeleteNetworkPolicy(ns string, name string) error {
	nsObject, err := m.getNamespaceObject(ns)
	if err != nil {
		return err
	}
	if _, ok := nsObject.Netpols[name]; !ok {
		return errors.Errorf("network policy %s/%s not found", ns, name)
	}
	delete(nsObject.Netpols, name)
	return nil
}

func (m *MockKubernetes) GetNetworkPoliciesInNamespace(ctx context.Context, namespace string) ([]networkingv1.NetworkPolicy, error) {
	if m.NetworkPolicyError != nil {
		return nil, m.NetworkPolicyError
	}

	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return nil, err
	}
	var netpols []networkingv1.NetworkPolicy
	for _, netpol := range nsObject.Netpols {
		netpols = append(netpols, *netpol)
	}
	return netpols, nil
}

func (m *MockKubernetes) UpdateNetworkPolicy(policy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	nsObject, err := m.getNamespaceObject(policy.Namespace)
	if err != nil {
		return nil, err
	}
	if _, ok := nsObject.Netpols[policy.Name]; !ok {
		return nil, errors.Errorf("network policy %s/%s not found", policy.Namespace, policy.Name)
	}
	nsObject.Netpols[policy.Name] = policy
	return policy, nil
}

func (m *MockKubernetes) CreateNetworkPolicy(policy *networkingv1.NetworkPolicy) (*networkingv1.NetworkPolicy, error) {
	nsObject, err := m.getNamespaceObject(policy.Namespace)
	if err != nil {
		return nil, err
	}
	if _, ok := nsObject.Netpols[policy.Name]; ok {
		return nil, errors.Errorf("network policy %s/%s already present", policy.Namespace, policy.Name)
	}
	nsObject.Netpols[policy.Name] = policy
	return policy, nil
}

func (m *MockKubernetes) GetService(namespace string, name string) (*v1.Service, error) {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return nil, err
	}
	if svc, ok := nsObject.Services[name]; ok {
		return svc, nil
	}
	return nil, errors.Errorf("service %s/%s not found", namespace, name)
}

func (m *MockKubernetes) CreateService(svc *v1.Service) (*v1.Service, error) {
	nsObject, err := m.getNamespaceObject(svc.Namespace)
	if err != nil {
		return nil, err
	}
	if _, ok := nsObject.Services[svc.Name]; ok {
		return nil, errors.Errorf("service %s/%s already present", svc.Namespace, svc.Name)
	}
	nsObject.Services[svc.Name] = svc
	return svc, nil
}

func (m *MockKubernetes) DeleteService(namespace string, name string) error {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return err
	}
	if _, ok := nsObject.Services[name]; !ok {
		return errors.Errorf("service %s/%s not found", namespace, name)
	}
	delete(nsObject.Services, name)
	return nil
}

func (m *MockKubernetes) GetServicesInNamespace(namespace string) ([]v1.Service, error) {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return nil, err
	}
	var services []v1.Service
	for _, svc := range nsObject.Services {
		services = append(services, *svc)
	}
	return services, nil
}

func (m *MockKubernetes) GetPodsInNamespace(namespace string) ([]v1.Pod, error) {
	var pods []v1.Pod
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return nil, err
	}
	for _, pod := range nsObject.Pods {
		pods = append(pods, *pod)
	}
	return pods, nil
}

func (m *MockKubernetes) GetPod(namespace string, podName string) (*v1.Pod, error) {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return nil, err
	}
	if pod, ok := nsObject.Pods[podName]; ok {
		return pod, nil
	}
	return nil, errors.Errorf("pod %s/%s not found", namespace, podName)
}

func (m *MockKubernetes) SetPodLabels(namespace string, podName string, labels map[string]string) (*v1.Pod, error) {
	pod, err := m.GetPod(namespace, podName)
	if err != nil {
		return nil, err
	}
	pod.Labels = labels
	return pod, nil
}

func (m *MockKubernetes) CreatePod(pod *v1.Pod) (*v1.Pod, error) {
	nsObject, err := m.getNamespaceObject(pod.Namespace)
	if err != nil {
		return nil, err
	}
	if _, ok := nsObject.Pods[pod.Name]; ok {
		return nil, errors.Errorf("pod %s/%s already exists", pod.Namespace, pod.Name)
	}
	if m.podID >= 255 {
		panic(errors.Errorf("unable to handle more than 254 pods in mock"))
	}
	pod.Status.Phase = v1.PodRunning
	pod.Status.PodIP = fmt.Sprintf("192.168.1.%d", m.podID)
	m.podID++
	nsObject.Pods[pod.Name] = pod
	return pod, nil
}

func (m *MockKubernetes) DeletePod(namespace string, podName string) error {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return err
	}
	if _, ok := nsObject.Pods[podName]; !ok {
		return errors.Errorf("pod %s/%s not found", namespace, podName)
	}
	delete(nsObject.Pods, podName)
	return nil
}

func (m *MockKubernetes) ExecuteRemoteCommand(namespace string, pod string, container string, command []string) (string, string, error, error) {
	nsObject, err := m.getNamespaceObject(namespace)
	if err != nil {
		return "", "", nil, err
	}
	podObject, ok := nsObject.Pods[pod]
	if !ok {
		return "", "", nil, errors.Errorf("pod %s/%s not found", namespace, pod)
	}
	var containerObject *v1.Container
	for _, cont := range podObject.Spec.Containers {
		if cont.Name == container {
			containerObject = &cont
			break
		}
	}
	if containerObject == nil {
		return "", "", nil, errors.Errorf("container %s/%s/%s not found", namespace, pod, container)
	}

	// TODO could look at netpols, pods, etc. to determine if this resolves?

	if rand.Float64() > m.passRate {
		return "", "", errors.Errorf("mock call randomly failed"), nil
	}
	return "", "", nil, nil
}

func (m *MockKubernetes) GetAdminNetworkPolicies(ctx context.Context) ([]v1alpha1.AdminNetworkPolicy, error) {
	return m.AdminNetworkPolicies, m.AdminNetworkPolicyError
}

func (k *MockKubernetes) CreateAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.AdminNetworkPolicy) (*v1alpha1.AdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *MockKubernetes) UpdateAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.AdminNetworkPolicy) (*v1alpha1.AdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *MockKubernetes) DeleteAdminNetworkPolicy(ctx context.Context, name string) error {
	//TODO: implement
	return ErrNotImplemented
}

func (m *MockKubernetes) GetBaselineAdminNetworkPolicy(ctx context.Context) (*v1alpha1.BaselineAdminNetworkPolicy, error) {
	return m.BaselineNetworkPolicy, m.BaseAdminNetworkPolicyError
}

func (k *MockKubernetes) CreateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.BaselineAdminNetworkPolicy) (*v1alpha1.BaselineAdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *MockKubernetes) UpdateBaselineAdminNetworkPolicy(ctx context.Context, policy *v1alpha1.BaselineAdminNetworkPolicy) (*v1alpha1.BaselineAdminNetworkPolicy, error) {
	return nil, ErrNotImplemented
}

func (k *MockKubernetes) DeleteBaselineAdminNetworkPolicy(ctx context.Context, name string) error {
	//TODO: implement
	return ErrNotImplemented
}
