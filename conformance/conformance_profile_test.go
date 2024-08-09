/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conformance_test

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/network-policy-api/apis/v1alpha1"
	confv1a1 "sigs.k8s.io/network-policy-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/network-policy-api/conformance/tests"
	"sigs.k8s.io/network-policy-api/conformance/utils/flags"
	"sigs.k8s.io/network-policy-api/conformance/utils/suite"
)

var (
	cfg                 *rest.Config
	c                   client.Client
	clientSet           kubernetes.Interface
	supportedFeatures   sets.Set[suite.SupportedFeature]
	exemptFeatures      sets.Set[suite.SupportedFeature]
	implementation      *confv1a1.Implementation
	conformanceProfiles sets.Set[suite.ConformanceProfileName]
)

// TestConformanceProfiles runs conformance tests and generates a profile report.
// If profiles are not passed while running the tests, we fall back to standard conformance tests
func TestConformanceProfiles(t *testing.T) {
	var err error
	cfg, err = config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}
	c, err = client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		t.Fatalf("error building Kube config for client-go: %v", err)
	}
	clientSet, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		t.Fatalf("error when creating Kubernetes ClientSet: %v", err)
	}

	v1alpha1.Install(c.Scheme())

	// standard conformance flags
	supportedFeatures = suite.ParseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures = suite.ParseSupportedFeatures(*flags.ExemptFeatures)

	// conformance profile flags
	conformanceProfiles = suite.ParseConformanceProfiles(*flags.ConformanceProfiles)

	if conformanceProfiles.Len() > 0 {
		// if some conformance profiles have been set, run the experimental conformance suite...
		implementation, err = suite.ParseImplementation(
			*flags.ImplementationOrganization,
			*flags.ImplementationProject,
			*flags.ImplementationURL,
			*flags.ImplementationVersion,
			*flags.ImplementationContact,
			*flags.ImplementationAdditionalInformation,
		)
		if err != nil {
			t.Fatalf("Error parsing implementation's details: %v", err)
		}
		testConformance(t)
	} else {
		// ...otherwise run the standard conformance suite.
		t.Run("standard conformance tests", TestConformance)
	}
}

func testConformance(t *testing.T) {
	t.Logf("Running conformance profile tests with cleanup: %t\n debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]\n conformance profiles: [%v]",
		*flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures, *flags.ConformanceProfiles)

	cSuite, err := suite.NewConformanceProfileTestSuite(
		suite.ConformanceProfileOptions{
			Options: suite.Options{
				Client:                     c,
				ClientSet:                  clientSet,
				KubeConfig:                 *cfg,
				Debug:                      *flags.ShowDebug,
				CleanupBaseResources:       *flags.CleanupBaseResources,
				SupportedFeatures:          supportedFeatures,
				ExemptFeatures:             exemptFeatures,
				EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
			},
			Implementation:      *implementation,
			ConformanceProfiles: conformanceProfiles,
		})
	if err != nil {
		t.Fatalf("error creating experimental conformance test suite: %v", err)
	}

	cSuite.Setup(t)
	cSuite.Run(t, tests.ConformanceTests)
	report, err := cSuite.Report()
	if err != nil {
		t.Fatalf("error generating conformance profile report: %v", err)
	}
	writeReport(t.Logf, *report, *flags.ReportOutput)
}

func writeReport(logf func(string, ...any), report confv1a1.ConformanceReport, output string) error {
	rawReport, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	if output != "" {
		if err = os.WriteFile(output, rawReport, 0600); err != nil {
			return err
		}
	}
	logf("Conformance report:\n%s", string(rawReport))

	return nil
}
