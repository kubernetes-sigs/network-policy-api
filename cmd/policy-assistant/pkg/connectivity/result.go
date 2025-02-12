package connectivity

import (
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/connectivity/probe"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/generator"
)

type Result struct {
	// TODO should resources be captured per-step for tests that modify them?
	InitialResources *probe.Resources
	TestCase         *generator.TestCase
	Steps            []*StepResult
	Err              error
}

func (r *Result) Features() map[string][]string {
	return r.TestCase.GetFeatures()
}

func (r *Result) Passed(ignoreLoopback bool) bool {
	for _, step := range r.Steps {
		if !step.Passed(ignoreLoopback) {
			return false
		}
	}
	return true
}
