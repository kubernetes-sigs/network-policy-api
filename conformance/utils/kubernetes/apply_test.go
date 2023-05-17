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

package kubernetes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	_ "sigs.k8s.io/network-policy-api/conformance/utils/flags"
)

func TestPrepareResources(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected []unstructured.Unstructured
		applier  Applier
	}{{
		name:    "empty namespace labels",
		applier: Applier{},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
		}},
	}, {
		name: "simple namespace labels",
		applier: Applier{
			NamespaceLabels: map[string]string{
				"test": "false",
			},
		},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
					"labels": map[string]interface{}{
						"test": "false",
					},
				},
			},
		}},
	}, {
		name: "overwrite namespace labels",
		applier: Applier{
			NamespaceLabels: map[string]string{
				"test": "true",
			},
		},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
  labels:
    test: 'false'
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
					"labels": map[string]interface{}{
						"test": "true",
					},
				},
			},
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(tc.given), 4096)

			resources, err := tc.applier.prepareResources(t, decoder)

			require.NoError(t, err, "unexpected error preparing resources")
			require.EqualValues(t, tc.expected, resources)
		})
	}
}
