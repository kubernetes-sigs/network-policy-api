package cli

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/examples"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"strings"
	"time"

	"github.com/mattfenwick/collections/pkg/json"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/probe"
	"github.com/mattfenwick/cyclonus/pkg/generator"

	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ParseMode        = "parse"
	ExplainMode      = "explain"
	QueryTrafficMode = "query-traffic"
	QueryTargetMode  = "query-target"
	ProbeMode        = "probe"
)

var AllModes = []string{
	ParseMode,
	ExplainMode,
	QueryTrafficMode,
	QueryTargetMode,
	ProbeMode,
}

const DefaultTimeout = 3 * time.Minute

type AnalyzeArgs struct {
	AllNamespaces      bool
	Namespaces         []string
	UseExamplePolicies bool
	PolicyPath         string
	Context            string
	SimplifyPolicies   bool

	Modes []string

	// traffic
	TrafficPath string

	// targets
	TargetPodPath string

	// synthetic probe
	ProbePath string

	Timeout time.Duration
}

func SetupAnalyzeCommand() *cobra.Command {
	args := &AnalyzeArgs{}

	command := &cobra.Command{
		Use:   "analyze",
		Short: "analyze network policies",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, as []string) {
			RunAnalyzeCommand(args)
		},
	}

	command.Flags().BoolVar(&args.UseExamplePolicies, "use-example-policies", false, "if true, reads example policies")
	command.Flags().BoolVarP(&args.AllNamespaces, "all-namespaces", "A", false, "reads kube resources from all namespaces; same as kubectl's '--all-namespaces'/'-A' flag")
	command.Flags().StringSliceVarP(&args.Namespaces, "namespace", "n", []string{}, "namespaces to read kube resources from; similar to kubectl's '--namespace'/'-n' flag, except that multiple namespaces may be passed in and is empty if not set explicitly (instead of 'default' as in kubectl)")
	command.Flags().StringVar(&args.PolicyPath, "policy-path", "", "may be a file or a directory; if set, will attempt to read policies from the path")
	command.Flags().StringVar(&args.Context, "context", "", "selects kube context to read policies from; only reads from kube if one or more namespaces or all namespaces are specified")
	command.Flags().BoolVar(&args.SimplifyPolicies, "simplify-policies", true, "if true, reduce policies to simpler form while preserving semantics")

	command.Flags().StringSliceVar(&args.Modes, "mode", []string{ExplainMode}, "analysis modes to run; allowed values are "+strings.Join(AllModes, ","))

	command.Flags().StringVar(&args.TargetPodPath, "target-pod-path", "", "path to json target pod file -- json array of dicts")
	command.Flags().StringVar(&args.TrafficPath, "traffic-path", "", "path to json traffic file, containing of a list of traffic objects")
	command.Flags().StringVar(&args.ProbePath, "probe-path", "", "path to json model file for synthetic probe")
	command.Flags().DurationVar(&args.Timeout, "kube-client-timeout", DefaultTimeout, "kube client timeout")

	return command
}

func RunAnalyzeCommand(args *AnalyzeArgs) {
	// 1. read policies from kube
	var kubePolicies []*networkingv1.NetworkPolicy
	var kubeANPs []*v1alpha1.AdminNetworkPolicy
	var kubeBANP *v1alpha1.BaselineAdminNetworkPolicy
	var kubePods []v1.Pod
	var kubeNamespaces []v1.Namespace
	var netpolErr, anpErr, banpErr error
	if args.AllNamespaces || len(args.Namespaces) > 0 {
		kubeClient, err := kube.NewKubernetesForContext(args.Context)
		utils.DoOrDie(err)

		namespaces := args.Namespaces
		if args.AllNamespaces {
			nsList, err := kubeClient.GetAllNamespaces()
			utils.DoOrDie(err)
			kubeNamespaces = nsList.Items
			namespaces = []string{v1.NamespaceAll}
		}

		includeANPS, includeBANPSs := shouldIncludeANPandBANP(kubeClient.ClientSet)

		ctx, cancel := context.WithTimeout(context.TODO(), args.Timeout)
		defer cancel()

		kubePolicies, kubeANPs, kubeBANP, netpolErr, anpErr, banpErr = kube.ReadNetworkPoliciesFromKube(ctx, kubeClient, namespaces, includeANPS, includeBANPSs)

		if netpolErr != nil {
			logrus.Errorf("unable to read network policies from kube, ns '%s': %+v", namespaces, err)
		}
		if anpErr != nil {
			logrus.Errorf("Unable to fetch admin network policies: %s \n", anpErr)
		}
		if banpErr != nil {
			logrus.Errorf("Unable to fetch base admin network policies: %s \n", banpErr)
		}
	}
	// 2. read policies from file
	if args.PolicyPath != "" {
		policiesFromPath, anpsFromPath, banpFromPath, err := kube.ReadNetworkPoliciesFromPath(args.PolicyPath)
		utils.DoOrDie(err)
		kubePolicies = append(kubePolicies, policiesFromPath...)
		kubeANPs = append(kubeANPs, anpsFromPath...)
		if banpFromPath != nil && kubeBANP != nil {
			logrus.Debugf("More that one banp parsed - setting banp from file")
		}
		kubeBANP = banpFromPath
	}
	// 3. read example policies
	if args.UseExamplePolicies {
		kubePolicies = append(kubePolicies, netpol.AllExamples...)

		kubeANPs = append(kubeANPs, examples.CoreGressRulesCombinedANB...)
		if kubeBANP != nil {
			logrus.Debugf("More that onew banp parsed - setting banp from the examples")
		}
		kubeBANP = examples.CoreGressRulesCombinedBANB
	}

	logrus.Debugf("parsed policies:\n%s", json.MustMarshalToString(kubePolicies))
	policies := matcher.BuildV1AndV2NetPols(args.SimplifyPolicies, kubePolicies, kubeANPs, kubeBANP)

	for _, mode := range args.Modes {
		switch mode {
		case ParseMode:
			fmt.Println("parsed policies:")
			ParsePolicies(kubePolicies)
		case ExplainMode:
			fmt.Println("explained policies:")
			ExplainPolicies(policies)
		case QueryTargetMode:
			pods := make([]*QueryTargetPod, len(kubePods))
			for i, p := range kubePods {
				pods[i] = &QueryTargetPod{
					Namespace: p.Namespace,
					Labels:    p.Labels,
				}
			}
			fmt.Println("query target:")
			QueryTargets(policies, args.TargetPodPath, pods)
		case QueryTrafficMode:
			fmt.Println("query traffic:")
			QueryTraffic(policies, args.TrafficPath)
		case ProbeMode:
			fmt.Println("probe:")
			ProbeSyntheticConnectivity(policies, args.ProbePath, kubePods, kubeNamespaces)
		default:
			panic(errors.Errorf("unrecognized mode %s", mode))
		}
	}
}

func ParsePolicies(kubePolicies []*networkingv1.NetworkPolicy) {
	fmt.Println(kube.NetworkPoliciesToTable(kubePolicies))
}

func ExplainPolicies(explainedPolicies *matcher.Policy) {
	fmt.Printf("%s\n", explainedPolicies.ExplainTable())
}

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

type SyntheticProbeConnectivityConfig struct {
	Resources *probe.Resources
	Probes    []*generator.PortProtocol
}

func ProbeSyntheticConnectivity(explainedPolicies *matcher.Policy, modelPath string, kubePods []v1.Pod, kubeNamespaces []v1.Namespace) {
	if modelPath != "" {
		config, err := json.ParseFile[SyntheticProbeConnectivityConfig](modelPath)
		utils.DoOrDie(err)

		jobBuilder := &probe.JobBuilder{TimeoutSeconds: 10}

		// run probes
		for _, probeConfig := range config.Probes {
			probeResult := probe.NewSimulatedRunner(explainedPolicies, jobBuilder).
				RunProbeForConfig(generator.NewProbeConfig(probeConfig.Port, probeConfig.Protocol, generator.ProbeModeServiceName), config.Resources)

			logrus.Infof("probe on port %s, protocol %s", probeConfig.Port.String(), probeConfig.Protocol)

			fmt.Printf("Ingress:\n%s\n", probeResult.RenderIngress())

			fmt.Printf("Egress:\n%s\n", probeResult.RenderEgress())

			fmt.Printf("Combined:\n%s\n\n\n", probeResult.RenderTable())
		}
	}

	resources := &probe.Resources{
		Namespaces: map[string]map[string]string{},
		Pods:       []*probe.Pod{},
	}

	nsMap := map[string]v1.Namespace{}
	for _, ns := range kubeNamespaces {
		nsMap[ns.Name] = ns
		resources.Namespaces[ns.Name] = ns.Labels
	}

	for _, pod := range kubePods {
		var containers []*probe.Container
		for _, cont := range pod.Spec.Containers {
			if len(cont.Ports) == 0 {
				logrus.Warnf("skipping container %s/%s/%s, no ports available", pod.Namespace, pod.Name, cont.Name)
				continue
			}
			port := cont.Ports[0]
			containers = append(containers, &probe.Container{
				Name:     cont.Name,
				Port:     int(port.ContainerPort),
				Protocol: port.Protocol,
				PortName: port.Name,
			})
		}
		if len(containers) == 0 {
			logrus.Warnf("skipping pod %s/%s, no containers available", pod.Namespace, pod.Name)
			continue
		}
		resources.Pods = append(resources.Pods, &probe.Pod{
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			Labels:     pod.Labels,
			IP:         pod.Status.PodIP,
			Containers: containers,
		})
	}

	simRunner := probe.NewSimulatedRunner(explainedPolicies, &probe.JobBuilder{TimeoutSeconds: 10})
	simulatedProbe := simRunner.RunProbeForConfig(generator.ProbeAllAvailable, resources)
	fmt.Printf("Ingress:\n%s\n", simulatedProbe.RenderIngress())
	fmt.Printf("Egress:\n%s\n", simulatedProbe.RenderEgress())
	fmt.Printf("Combined:\n%s\n\n\n", simulatedProbe.RenderTable())
}

func shouldIncludeANPandBANP(client *kubernetes.Clientset) (bool, bool) {
	var includeANP, includeBANP bool
	_, resources, _, err := client.DiscoveryClient.GroupsAndMaybeResources()
	if err != nil {
		logrus.Errorf("Unable to fetch all registered resources: %s", err)
		return includeANP, includeBANP
	}
	gv := schema.GroupVersion{Group: "policy.networking.k8s.io", Version: "v1alpha1"}

	if groupResources, ok := resources[gv]; ok {
		for _, res := range groupResources.APIResources {
			switch res.Kind {
			case "AdminNetworkPolicy":
				includeANP = true
			case "BaselineAdminNetworkPolicy":
				includeBANP = true
			default:
				continue
			}
		}
	}

	return includeANP, includeBANP
}
