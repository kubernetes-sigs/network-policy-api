/*
Copyright 2021 The Kubernetes Authors.
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

package main

import (
	"fmt"
	"log"
	"os"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/yaml"
)

const (
	bundleVersionAnnotation = "policy.networking.k8s.io/bundle-version"

	// These values must be updated during the release process
	bundleVersion = "v0.1.0"
	approvalLink  = "https://github.com/kubernetes-sigs/network-policy-api/pull/93"
)

// This generation code is largely copied from
// github.com/kubernetes-sigs/controller-tools/blob/ab52f76cc7d167925b2d5942f24bf22e30f49a02/pkg/crd/gen.go
func main() {
	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/network-policy-api/apis/v1alpha1",
	)
	if err != nil {
		log.Fatalf("failed to load package roots: %s", err)
	}

	generator := &crd.Generator{}

	parser := &crd.Parser{
		Collector: &markers.Collector{Registry: &markers.Registry{}},
		Checker: &loader.TypeChecker{
			NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
		},
	}

	err = generator.RegisterMarkers(parser.Collector.Registry)
	if err != nil {
		log.Fatalf("failed to register markers: %s", err)
	}

	crd.AddKnownTypes(parser)
	for _, r := range roots {
		parser.NeedPackage(r)
	}

	metav1Pkg := crd.FindMetav1(roots)
	if metav1Pkg == nil {
		log.Fatalf("no objects in the roots, since nothing imported metav1")
	}

	kubeKinds := crd.FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		log.Fatalf("no objects in the roots")
	}

	for _, groupKind := range kubeKinds {
		log.Printf("generating CRD for %v\n", groupKind)

		parser.NeedCRDFor(groupKind, nil)
		crdRaw := parser.CustomResourceDefinitions[groupKind]

		// Inline version of "addAttribution(&crdRaw)" ...
		if crdRaw.ObjectMeta.Annotations == nil {
			crdRaw.ObjectMeta.Annotations = map[string]string{}
		}
		crdRaw.ObjectMeta.Annotations[bundleVersionAnnotation] = bundleVersion
		crdRaw.ObjectMeta.Annotations[apiext.KubeAPIApprovedAnnotation] = approvalLink

		// Prevent the top level metadata for the CRD to be generated regardless of the intention in the arguments
		crd.FixTopLevelMetadata(crdRaw)

		conv, err := crd.AsVersion(crdRaw, apiext.SchemeGroupVersion)
		if err != nil {
			log.Fatalf("failed to convert CRD: %s", err)
		}

		out, err := yaml.Marshal(conv)
		if err != nil {
			log.Fatalf("failed to marshal CRD: %s", err)
		}

		fileName := fmt.Sprintf("config/crd/%s_%s.yaml", crdRaw.Spec.Group, crdRaw.Spec.Names.Plural)
		err = os.WriteFile(fileName, out, 0o600)
		if err != nil {
			log.Fatalf("failed to write CRD: %s", err)
		}
	}
}
