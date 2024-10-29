package cli

import (
	"fmt"

	"github.com/mattfenwick/collections/pkg/json"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// case ParseMode:
//
//	fmt.Println("parsed policies:")
//	ParsePolicies(kubePolicies)
func ParsePolicies(kubePolicies []*networkingv1.NetworkPolicy) {
	fmt.Println(kube.NetworkPoliciesToTable(kubePolicies))
}

// case QueryTargetMode:
// 	pods := make([]*QueryTargetPod, len(kubePods))
// 	for i, p := range kubePods {
// 		pods[i] = &QueryTargetPod{
// 			Namespace: p.Namespace,
// 			Labels:    p.Labels,
// 		}
// 	}
// 	fmt.Println("query target:")
// 	QueryTargets(policies, args.TargetPodPath, pods)
// case QueryTrafficMode:
// 	fmt.Println("query traffic:")
// 	QueryTraffic(policies, args.TrafficPath)

// QueryTargetPod matches targets; targets exist in only a single namespace and can't be matched by namespace
//
//	label, therefore we match by exact namespace and by pod labels.
type QueryTargetPod struct {
	Namespace string
	Labels    map[string]string
}

func QueryTargets(explainedPolicies *matcher.Policy, podPath string, pods []*QueryTargetPod) {
	if podPath != "" {
		podsFromFile, err := json.ParseFile[[]*QueryTargetPod](podPath)
		utils.DoOrDie(err)
		pods = append(pods, *podsFromFile...)
	}

	for _, pod := range pods {
		fmt.Printf("pod in ns %s with labels %+v:\n\n", pod.Namespace, pod.Labels)

		targets, combinedRules := QueryTargetHelper(explainedPolicies, pod)

		fmt.Printf("Matching targets:\n%s\n", targets.ExplainTable())
		fmt.Printf("Combined rules:\n%s\n\n\n", combinedRules.ExplainTable())
	}
}

func QueryTargetHelper(policies *matcher.Policy, pod *QueryTargetPod) (*matcher.Policy, *matcher.Policy) {
	podInfo := &matcher.InternalPeer{
		Namespace: pod.Namespace,
		PodLabels: pod.Labels,
	}
	ingressTargets := policies.TargetsApplyingToPod(true, podInfo)
	combinedIngressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, ingressTargets)

	egressTargets := policies.TargetsApplyingToPod(false, podInfo)
	combinedEgressTarget := matcher.CombineTargetsIgnoringPrimaryKey(pod.Namespace, metav1.LabelSelector{MatchLabels: pod.Labels}, egressTargets)

	var combinedIngresses []*matcher.Target
	if combinedIngressTarget != nil {
		combinedIngresses = []*matcher.Target{combinedIngressTarget}
	}
	var combinedEgresses []*matcher.Target
	if combinedEgressTarget != nil {
		combinedEgresses = []*matcher.Target{combinedEgressTarget}
	}

	return matcher.NewPolicyWithTargets(ingressTargets, egressTargets), matcher.NewPolicyWithTargets(combinedIngresses, combinedEgresses)
}

func QueryTraffic(explainedPolicies *matcher.Policy, trafficPath string) {
	if trafficPath == "" {
		logrus.Fatalf("%+v", errors.Errorf("path to traffic file required for QueryTraffic command"))
	}
	allTraffics, err := json.ParseFile[[]*matcher.Traffic](trafficPath)
	utils.DoOrDie(err)

	for _, traffic := range *allTraffics {
		fmt.Printf("Traffic:\n%s\n", traffic.Table())

		result := explainedPolicies.IsTrafficAllowed(traffic)
		fmt.Printf("Is traffic allowed?\n%s\n\n\n", result.Table())
	}
}
