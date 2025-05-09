/*
Copyright 2022 The Kubernetes Authors.

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

package suite

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	confv1a1 "sigs.k8s.io/network-policy-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/network-policy-api/conformance/utils/config"
)

type ConformanceProfileTestSuite struct {
	ConformanceTestSuite

	// implementation contains the details of the implementation, such as
	// organization, project, etc.
	implementation confv1a1.Implementation

	// conformanceProfiles is a compiled list of profiles to check
	// conformance against.
	conformanceProfiles sets.Set[ConformanceProfileName]

	// running indicates whether the test suite is currently running
	running bool

	// results stores the pass or fail results of each test that was run by
	// the test suite, organized by the tests unique name.
	results map[string]testResult

	// experimentalSupportedFeatures is a compiled list of named features that were
	// marked as supported, and is used for reporting the test results.
	experimentalSupportedFeatures map[ConformanceProfileName]sets.Set[SupportedFeature]

	// experimentalUnsupportedFeatures is a compiled list of named features that were
	// marked as not supported, and is used for reporting the test results.
	experimentalUnsupportedFeatures map[ConformanceProfileName]sets.Set[SupportedFeature]

	// lock is a mutex to help ensure thread safety of the test suite object.
	lock sync.RWMutex
}

// Options can be used to initialize a ConformanceProfileTestSuite.
type ConformanceProfileOptions struct {
	Options

	Implementation      confv1a1.Implementation
	ConformanceProfiles sets.Set[ConformanceProfileName]
}

// NewConformanceProfileTestSuite is a helper to use for creating a new ConformanceProfileTestSuite.
func NewConformanceProfileTestSuite(s ConformanceProfileOptions) (*ConformanceProfileTestSuite, error) {
	config.SetupTimeoutConfig(&s.TimeoutConfig)

	suite := &ConformanceProfileTestSuite{
		results:                         make(map[string]testResult),
		experimentalUnsupportedFeatures: make(map[ConformanceProfileName]sets.Set[SupportedFeature]),
		experimentalSupportedFeatures:   make(map[ConformanceProfileName]sets.Set[SupportedFeature]),
		conformanceProfiles:             s.ConformanceProfiles,
		implementation:                  s.Implementation,
	}

	// test suite callers are required to provide a conformance profile OR at
	// minimum a list of features which they support.
	if s.SupportedFeatures == nil && s.ConformanceProfiles.Len() == 0 && !s.EnableAllSupportedFeatures {
		return nil, fmt.Errorf("no conformance profile was selected for test run, and no supported features were provided so no tests could be selected")
	}

	// test suite callers can potentially just run all tests by saying they
	// cover all features, if they don't they'll need to have provided a
	// conformance profile or at least some specific features they support.
	if s.EnableAllSupportedFeatures {
		s.SupportedFeatures = AllFeatures
	} else {
		if s.SupportedFeatures == nil {
			s.SupportedFeatures = sets.New[SupportedFeature]()
		}
		// the use of a conformance profile implicitly enables any features of
		// that profile which are supported at a Standard level of support.
		for _, conformanceProfileName := range s.ConformanceProfiles.UnsortedList() {
			conformanceProfile, err := getConformanceProfileForName(conformanceProfileName)
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve conformance profile: %w", err)
			}
			for _, f := range conformanceProfile.ExperimentalFeatures.UnsortedList() {
				if s.SupportedFeatures.Has(f) {
					if suite.experimentalSupportedFeatures[conformanceProfileName] == nil {
						suite.experimentalSupportedFeatures[conformanceProfileName] = sets.New[SupportedFeature]()
					}
					suite.experimentalSupportedFeatures[conformanceProfileName].Insert(f)
				} else {
					if suite.experimentalUnsupportedFeatures[conformanceProfileName] == nil {
						suite.experimentalUnsupportedFeatures[conformanceProfileName] = sets.New[SupportedFeature]()
					}
					suite.experimentalUnsupportedFeatures[conformanceProfileName].Insert(f)
				}
			}
		}
	}
	suite.ConformanceTestSuite = *New(s.Options)
	return suite, nil
}

// -----------------------------------------------------------------------------
// Conformance Test Suite - Public Methods
// -----------------------------------------------------------------------------

// Setup ensures the base resources required for conformance tests are installed
// in the cluster. It also ensures that all relevant resources are ready.
func (suite *ConformanceProfileTestSuite) Setup(t *testing.T) {
	suite.ConformanceTestSuite.Setup(t)
}

// Run runs the provided set of conformance tests.
func (suite *ConformanceProfileTestSuite) Run(t *testing.T, tests []ConformanceTest) error {
	// verify that the test suite isn't already running, don't start a new run
	// until the previous run finishes
	suite.lock.Lock()
	if suite.running {
		suite.lock.Unlock()
		return fmt.Errorf("can't run the test suite multiple times in parallel: the test suite is already running")
	}

	// if the test suite is not currently running, reset reporting and start a
	// new test run.
	suite.running = true
	suite.results = nil
	suite.lock.Unlock()

	// run all tests and collect the test results for conformance reporting
	results := make(map[string]testResult)
	for _, test := range tests {
		succeeded := t.Run(test.ShortName, func(t *testing.T) {
			test.Run(t, &suite.ConformanceTestSuite)
		})
		res := testSucceeded
		if suite.SkipTests.Has(test.ShortName) {
			res = testSkipped
		}
		if !suite.SupportedFeatures.HasAll(test.Features...) {
			res = testNotSupported
		}

		if !succeeded {
			res = testFailed
		}

		results[test.ShortName] = testResult{
			test:   test,
			result: res,
		}
	}

	// now that the tests have completed, mark the test suite as not running
	// and report the test results.
	suite.lock.Lock()
	suite.running = false
	suite.results = results
	suite.lock.Unlock()

	return nil
}

// Report emits a ConformanceReport for the previously completed test run.
// If no run completed prior to running the report, and error is emitted.
func (suite *ConformanceProfileTestSuite) Report() (*confv1a1.ConformanceReport, error) {
	suite.lock.RLock()
	if suite.running {
		suite.lock.RUnlock()
		return nil, fmt.Errorf("can't generate report: the test suite is currently running")
	}
	defer suite.lock.RUnlock()

	profileReports := newReports()
	for _, testResult := range suite.results {
		conformanceProfiles := getConformanceProfilesForTest(testResult.test, suite.conformanceProfiles)
		for _, profile := range conformanceProfiles.UnsortedList() {
			profileReports.addTestResults(*profile, testResult)
		}
	}

	profileReports.compileResults(suite.experimentalSupportedFeatures, suite.experimentalUnsupportedFeatures)

	return &confv1a1.ConformanceReport{
		TypeMeta: v1.TypeMeta{
			APIVersion: "policy.networking.k8s.io/v1alpha1",
			Kind:       "ConformanceReport",
		},
		Date:           time.Now().Format(time.RFC3339),
		Implementation: suite.implementation,
		// TODO: Need to add logic to how we can determine against which API version test was run against
		// Shouldn't this be same as Implementation.Version?
		// Currently pinning it to the version where profiles are going to be introduced in
		// We might need to bump this with every version we release
		NetworkPolicyV2APIVersion: "v0.1.2",
		ProfileReports:            profileReports.list(),
	}, nil
}

// ParseImplementation parses implementation-specific flag arguments and
// creates a *confv1a1.Implementation.
func ParseImplementation(org, project, url, version, contact, addInfo string) (*confv1a1.Implementation, error) {
	if org == "" {
		return nil, errors.New("implementation's organization cannot be empty")
	}
	if project == "" {
		return nil, errors.New("implementation's project cannot be empty")
	}
	if url == "" {
		return nil, errors.New("implementation's url cannot be empty")
	}
	if version == "" {
		return nil, errors.New("implementation's version cannot be empty")
	}
	if addInfo == "" {
		return nil, errors.New("implementation's CI integration used to generate the report cannot be empty")
	}
	contacts := strings.SplitN(contact, ",", -1)
	if len(contacts) == 0 {
		return nil, errors.New("implementation's contact can not be empty")
	}

	// TODO: add data validation https://github.com/kubernetes-sigs/gateway-api/issues/2178
	// This is also relevant to network-policy-api usage of gateway-api's conformance profile feature

	return &confv1a1.Implementation{
		Organization:          org,
		Project:               project,
		URL:                   url,
		Version:               version,
		Contact:               contacts,
		AdditionalInformation: addInfo,
	}, nil
}

// ParseConformanceProfiles parses flag arguments and converts the string to
// sets.Set[ConformanceProfileName].
func ParseConformanceProfiles(p string) sets.Set[ConformanceProfileName] {
	res := sets.Set[ConformanceProfileName]{}
	if p == "" {
		return res
	}

	for _, value := range strings.Split(p, ",") {
		res.Insert(ConformanceProfileName(value))
	}
	return res
}
