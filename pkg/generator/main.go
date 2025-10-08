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
	"regexp"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/yaml"
)

const (
	bundleVersionAnnotation = "policy.networking.k8s.io/bundle-version"
	channelAnnotation       = "policy.networking.k8s.io/channel"

	// These values must be updated during the release process
	bundleVersion = "v0.1.7"
	approvalLink  = "https://github.com/kubernetes-sigs/network-policy-api/pull/300"
)

var standardKinds = map[string]bool{
	"ClusterNetworkPolicy": true,
}

// This generation code is largely copied from
// github.com/kubernetes-sigs/controller-tools/blob/ab52f76cc7d167925b2d5942f24bf22e30f49a02/pkg/crd/gen.go
func main() {
	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/network-policy-api/apis/v1alpha2",
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

	channels := []string{"standard", "experimental"}
	for _, channel := range channels {
		for _, groupKind := range kubeKinds {
			if channel == "standard" && !standardKinds[groupKind.Kind] {
				continue
			}

			log.Printf("generating %s CRD for %v\n", channel, groupKind)

			parser.NeedCRDFor(groupKind, nil)
			crdRaw := parser.CustomResourceDefinitions[groupKind]

			// Inline version of "addAttribution(&crdRaw)" ...
			if crdRaw.ObjectMeta.Annotations == nil {
				crdRaw.ObjectMeta.Annotations = map[string]string{}
			}
			crdRaw.ObjectMeta.Annotations[bundleVersionAnnotation] = bundleVersion
			crdRaw.ObjectMeta.Annotations[channelAnnotation] = channel
			crdRaw.ObjectMeta.Annotations[apiext.KubeAPIApprovedAnnotation] = approvalLink

			// Prevent the top level metadata for the CRD to be generated regardless of the intention in the arguments
			crd.FixTopLevelMetadata(crdRaw)

			channelCrd := crdRaw.DeepCopy()
			for _, version := range channelCrd.Spec.Versions {
				version.Schema.OpenAPIV3Schema.Properties = channelTweaks(channel, version.Schema.OpenAPIV3Schema.Properties)
				if _, ok := version.Schema.OpenAPIV3Schema.Properties["namedPort"]; ok {
					fmt.Println("Changes not persisted")
				}
			}

			conv, err := crd.AsVersion(*channelCrd, apiext.SchemeGroupVersion)
			if err != nil {
				log.Fatalf("failed to convert CRD: %s", err)
			}

			out, err := yaml.Marshal(conv)
			if err != nil {
				log.Fatalf("failed to marshal CRD: %s", err)
			}

			fileName := fmt.Sprintf("config/crd/%s/%s_%s.yaml", channel, crdRaw.Spec.Group, crdRaw.Spec.Names.Plural)
			err = os.WriteFile(fileName, out, 0o600)
			if err != nil {
				log.Fatalf("failed to write CRD: %s", err)
			}
		}
	}
}

func channelTweaks(channel string, props map[string]apiext.JSONSchemaProps) map[string]apiext.JSONSchemaProps {
	for name := range props {
		jsonProps := props[name]
		if channel == "standard" {
			if strings.Contains(jsonProps.Description, "<network-policy-api:experimental>") {
				log.Printf("Deleting experimental field from standard channel %s\n", name)
				delete(props, name)
				continue
			}

			if strings.Contains(jsonProps.Description, "<network-policy-api:experimental:validation>") {
				log.Printf("Deleting experimental validation for standard field %s %+v\n", name, jsonProps.Items.Schema.XValidations)
				jsonProps.Items.Schema.XValidations = nil
			}

			if jsonProps.Items != nil {
				if jsonProps.Items.Schema != nil {
					if strings.Contains(jsonProps.Items.Schema.Description, "<network-policy-api:experimental:validation>") {
						log.Printf("Deleting experimental validation for standard type %s %+v\n", name, jsonProps.Items.Schema.XValidations)
						jsonProps.Items.Schema.XValidations = nil
					}
				}
			}
		}

		if channel == "experimental" && strings.Contains(jsonProps.Description, "<network-policy-api:experimental:validation:") {
			validationRe := regexp.MustCompile(`<network-policy-api:experimental:validation:Enum=([A-Za-z;]*)>`)
			match := validationRe.FindStringSubmatch(jsonProps.Description)
			if len(match) != 2 {
				log.Fatalf("Invalid network-policy-api:experimental:validation tag for %s", name)
			}
			jsonProps.Enum = []apiext.JSON{}
			for _, val := range strings.Split(match[1], ";") {
				jsonProps.Enum = append(jsonProps.Enum, apiext.JSON{Raw: []byte("\"" + val + "\"")})
			}
		}

		experimentalRe := regexp.MustCompile(`<network-policy-api:experimental:.*>`)
		jsonProps.Description = experimentalRe.ReplaceAllLiteralString(jsonProps.Description, "")

		if len(jsonProps.Properties) > 0 {
			jsonProps.Properties = channelTweaks(channel, jsonProps.Properties)
		} else if jsonProps.Items != nil && jsonProps.Items.Schema != nil {
			jsonProps.Items.Schema.Properties = channelTweaks(channel, jsonProps.Items.Schema.Properties)
		}
		props[name] = jsonProps
	}
	return props
}
