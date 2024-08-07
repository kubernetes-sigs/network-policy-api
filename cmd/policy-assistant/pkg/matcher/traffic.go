package matcher

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
)

type Traffic struct {
	Source      *TrafficPeer
	Destination *TrafficPeer

	ResolvedPort     int
	ResolvedPortName string
	Protocol         v1.Protocol
}

func (t *Traffic) Table() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)

	pp := fmt.Sprintf("%d (%s) on %s", t.ResolvedPort, t.ResolvedPortName, t.Protocol)
	table.SetHeader([]string{"Port/Protocol", "Source/Dest", "Pod IP", "Namespace", "NS Labels", "Pod Labels"})

	source := []string{pp, "source", t.Source.IP}
	if t.Source.Internal != nil {
		i := t.Source.Internal
		source = append(source, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		source = append(source, "", "", "")
	}
	table.Append(source)

	dest := []string{pp, "destination", t.Destination.IP}
	if t.Destination.Internal != nil {
		i := t.Destination.Internal
		dest = append(dest, i.Namespace, labelsToString(i.NamespaceLabels), labelsToString(i.PodLabels))
	} else {
		dest = append(dest, "", "", "")
	}
	table.Append(dest)

	table.Render()
	return tableString.String()
}

func labelsToString(labels map[string]string) string {
	format := func(k string) string { return fmt.Sprintf("%s: %s", k, labels[k]) }
	return strings.Join(slice.Map(format, slice.Sort(maps.Keys(labels))), "\n")
}

func (t *Traffic) PrettyString() string {
	if t == nil || t.Source == nil || t.Destination == nil {
		return "<undefined>"
	}

	src := t.Source.Internal.Workload
	if src == "" {
		if t.Source.Internal == nil {
			return "<undefined>"
		}

		src = fmt.Sprintf("%s/%s", t.Source.Internal.Namespace, labelsToStringSlim(t.Source.Internal.PodLabels))
	}

	dst := t.Destination.Internal.Workload
	if dst == "" {
		if t.Destination.Internal == nil {
			return "<undefined>"
		}

		dst = fmt.Sprintf("%s/%s", t.Destination.Internal.Namespace, labelsToStringSlim(t.Destination.Internal.PodLabels))
	}

	return fmt.Sprintf("%s -> %s:%d (%s)", src, dst, t.ResolvedPort, t.Protocol)
}

func labelsToStringSlim(labels map[string]string) string {
	format := func(k string) string { return fmt.Sprintf("%s=%s", k, labels[k]) }
	list := strings.Join(slice.Map(format, slice.Sort(maps.Keys(labels))), ",")
	return fmt.Sprintf("[%s]", list)
}

type TrafficPeer struct {
	Internal *InternalPeer
	// IP external to cluster
	IP string
}

func (p *TrafficPeer) Namespace() string {
	if p.Internal == nil {
		return ""
	}
	return p.Internal.Namespace
}

func (p *TrafficPeer) IsExternal() bool {
	return p.Internal == nil
}

func (p *TrafficPeer) Translate() TrafficPeer {
	//Translates kubernetes workload types to TrafficPeers.
	var podsNetworking []*PodNetworking
	var podLabels map[string]string
	var namespaceLabels map[string]string
	var workloadOwner string
	var workloadKind string
	var internalPeer InternalPeer
	workloadOwnerExists := false
	workloadMetadata := strings.Split(strings.ToLower(p.Internal.Workload), "/")
	if len(workloadMetadata) != 3 || (workloadMetadata[0] == "" || workloadMetadata[1] == "" || workloadMetadata[2] == "") || (workloadMetadata[1] != "daemonset" && workloadMetadata[1] != "statefulset" && workloadMetadata[1] != "replicaset" && workloadMetadata[1] != "deployment" && workloadMetadata[1] != "pod") {
		logrus.Fatalf("Bad Workload structure: Types supported are pod, replicaset, deployment, daemonset, statefulset, and 3 fields are required with this structure, <namespace>/<workloadType>/<workloadName>")
	}
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	ns, err := kubeClient.GetNamespace(workloadMetadata[0])
	utils.DoOrDie(err)
	kubePods, err := kube.GetPodsInNamespaces(kubeClient, []string{workloadMetadata[0]})
	if err != nil {
		logrus.Fatalf("unable to read pods from kube, ns '%s': %+v", workloadMetadata[0], err)
	}
	for _, pod := range kubePods {
		if workloadMetadata[1] == "deployment" && pod.OwnerReferences != nil && pod.OwnerReferences[0].Kind == "ReplicaSet" {
			kubeReplicaSets, err := kubeClient.GetReplicaSet(workloadMetadata[0], pod.OwnerReferences[0].Name)
			if err != nil {
				logrus.Fatalf("unable to read Replicaset from kube, rs '%s': %+v", pod.OwnerReferences[0].Name, err)
			}
			if kubeReplicaSets.OwnerReferences != nil {
				workloadOwner = kubeReplicaSets.OwnerReferences[0].Name
				workloadKind = "deployment"
			}

		} else if (workloadMetadata[1] == "daemonset" || workloadMetadata[1] == "statefulset" || workloadMetadata[1] == "replicaset") && pod.OwnerReferences != nil {
			workloadOwner = pod.OwnerReferences[0].Name
			workloadKind = pod.OwnerReferences[0].Kind
		} else if workloadMetadata[1] == "pod" {
			workloadOwner = pod.Name
			workloadKind = "pod"
		}
		if strings.ToLower(workloadOwner) == workloadMetadata[2] && strings.ToLower(workloadKind) == workloadMetadata[1] {
			podLabels = pod.Labels
			namespaceLabels = ns.Labels
			podNetworking := PodNetworking{
				IP: pod.Status.PodIP,
			}
			podsNetworking = append(podsNetworking, &podNetworking)
			workloadOwnerExists = true

		}
	}

	if !workloadOwnerExists {
		logrus.Infof("workload not found on the cluster")
		internalPeer = InternalPeer{
			Workload: "",
		}
	} else {
		internalPeer = InternalPeer{
			Workload:        p.Internal.Workload,
			PodLabels:       podLabels,
			NamespaceLabels: namespaceLabels,
			Namespace:       workloadMetadata[0],
			Pods:            podsNetworking,
		}
	}

	TranslatedPeer := TrafficPeer{
		Internal: &internalPeer,
	}
	return TranslatedPeer
}

func DeploymentsToTrafficPeers() []TrafficPeer {
	//Translates all pods associated with deployments to TrafficPeers.
	var deploymentPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeDeployments, err := kubeClient.GetDeploymentsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read deployments from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, deployment := range kubeDeployments {
			tmpInternalPeer := InternalPeer{
				Workload: namespace.Name + "/deployment/" + deployment.Name,
			}
			tmpPeer := TrafficPeer{
				Internal: &tmpInternalPeer,
			}
			tmpPeerTranslated := tmpPeer.Translate()
			if tmpPeerTranslated.Internal.Workload != "" {
				deploymentPeers = append(deploymentPeers, tmpPeerTranslated)
			}

		}

	}

	return deploymentPeers
}

func DaemonSetsToTrafficPeers() []TrafficPeer {
	//Translates all pods associated with daemonSets to TrafficPeers.
	var daemonSetPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeDaemonSets, err := kubeClient.GetDaemonSetsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read daemonSets from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, daemonSet := range kubeDaemonSets {
			tmpInternalPeer := InternalPeer{
				Workload: namespace.Name + "/daemonset/" + daemonSet.Name,
			}
			tmpPeer := TrafficPeer{
				Internal: &tmpInternalPeer,
			}
			tmpPeerTranslated := tmpPeer.Translate()
			if tmpPeerTranslated.Internal.Workload != "" {
				daemonSetPeers = append(daemonSetPeers, tmpPeerTranslated)
			}
		}

	}

	return daemonSetPeers
}

func StatefulSetsToTrafficPeers() []TrafficPeer {
	//Translates all pods associated with statefulSets to TrafficPeers.
	var statefulSetPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeStatefulSets, err := kubeClient.GetStatefulSetsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read statefulSets from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, statefulSet := range kubeStatefulSets {
			tmpInternalPeer := InternalPeer{
				Workload: namespace.Name + "/statefulset/" + statefulSet.Name,
			}
			tmpPeer := TrafficPeer{
				Internal: &tmpInternalPeer,
			}
			tmpPeerTranslated := tmpPeer.Translate()
			if tmpPeerTranslated.Internal.Workload != "" {
				statefulSetPeers = append(statefulSetPeers, tmpPeerTranslated)
			}
		}

	}

	return statefulSetPeers
}

func ReplicaSetsToTrafficPeers() []TrafficPeer {
	//Translates all pods associated with replicaSets that are not associated with deployments to TrafficPeers.
	var replicaSetPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubeReplicaSets, err := kubeClient.GetReplicaSetsInNamespace(namespace.Name)
		if err != nil {
			logrus.Fatalf("unable to read replicaSets from kube, ns '%s': %+v", namespace.Name, err)
		}

		for _, replicaSet := range kubeReplicaSets {
			if replicaSet.OwnerReferences != nil {
				continue
			} else {
				tmpInternalPeer := InternalPeer{
					Workload: namespace.Name + "/replicaset/" + replicaSet.Name,
				}
				tmpPeer := TrafficPeer{
					Internal: &tmpInternalPeer,
				}
				tmpPeerTranslated := tmpPeer.Translate()
				if tmpPeerTranslated.Internal.Workload != "" {
					replicaSetPeers = append(replicaSetPeers, tmpPeerTranslated)
				}

			}
		}

	}

	return replicaSetPeers
}

func PodsToTrafficPeers() []TrafficPeer {
	//Translates all pods that are not associated with other workload types (deployment, replicaSet, daemonSet, statefulSet.) to TrafficPeers.
	var podPeers []TrafficPeer
	kubeClient, err := kube.NewKubernetesForContext("")
	utils.DoOrDie(err)
	kubeNamespaces, err := kubeClient.GetAllNamespaces()
	if err != nil {
		logrus.Fatalf("unable to read namespaces from kube: %+v", err)
	}

	for _, namespace := range kubeNamespaces.Items {
		kubePods, err := kube.GetPodsInNamespaces(kubeClient, []string{namespace.Name})
		if err != nil {
			logrus.Fatalf("unable to read pods from kube, ns '%s': %+v", namespace.Name, err)
		}
		for _, pod := range kubePods {
			if pod.OwnerReferences != nil {
				continue
			} else {
				tmpInternalPeer := InternalPeer{
					Workload: namespace.Name + "/pod/" + pod.Name,
				}
				tmpPeer := TrafficPeer{
					Internal: &tmpInternalPeer,
				}
				tmpPeerTranslated := tmpPeer.Translate()
				if tmpPeerTranslated.Internal.Workload != "" {
					podPeers = append(podPeers, tmpPeerTranslated)
				}
			}
		}

	}

	return podPeers
}

// Internal to cluster
type InternalPeer struct {
	// optional: if set, will override remaining values with information from cluster
	Workload        string
	PodLabels       map[string]string
	NamespaceLabels map[string]string
	Namespace       string
	// optional
	Pods []*PodNetworking
}

type PodNetworking struct {
	IP string
	// don't worry about populating below fields right now
	IsHostNetworking bool
	NodeLabels       []string
}
