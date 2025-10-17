/*
Copyright 2025 The Kubernetes Authors.
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
	"embed"
	_ "embed"
	"path"

	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	//go:embed valid/*.yaml
	validCases embed.FS
	//go:embed invalid/*.yaml
	invalidCases embed.FS
)

// loadYAML from the given embedded test case.
func loadYAML(fs embed.FS, name string) (*unstructured.Unstructured, string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return nil, "", err
	}
	obj := &unstructured.Unstructured{}
	dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(b)), 4096)
	if err := dec.Decode(&obj); err != nil {
		return nil, "", err
	}
	return obj, string(b), nil
}

func TestValid(t *testing.T) {
	const validDir = "valid"

	entries, err := validCases.ReadDir(validDir)
	if err != nil {
		t.Fatalf("ReadDir() = %v, want nil", err)
	}
	for _, e := range entries {
		obj, text, err := loadYAML(validCases, path.Join(validDir, e.Name()))
		if err != nil {
			t.Fatalf("loadYAML(%s) = %v, want nil\nYAML was:\n%s", e.Name(), err, text)
		}
		t.Run(obj.GetName(), func(t *testing.T) {
			if err := globals.k8sClient.Create(context.Background(), obj); err != nil {
				t.Fatalf("Create() = %v, want nil\nYAML was:\n%s", err, text)
			}
			if err := globals.k8sClient.Delete(context.Background(), obj); err != nil {
				t.Fatalf("Delete() = %v, want nil\nYAML was:\n%s", err, text)
			}
		})
	}
}

func TestInvalid(t *testing.T) {
	const validDir = "invalid"

	entries, err := invalidCases.ReadDir(validDir)
	if err != nil {
		t.Fatalf("ReadDir() = %v, want nil", err)
	}
	for _, e := range entries {
		obj, text, err := loadYAML(invalidCases, path.Join(validDir, e.Name()))

		if err != nil {
			t.Fatalf("loadYAML(%s) = %v, want nil\nYAML was:\n%s", e.Name(), err, text)
		}
		t.Run(obj.GetName(), func(t *testing.T) {
			err := globals.k8sClient.Create(context.Background(), obj)
			t.Logf("Create() = %v", err)
			if err == nil {
				t.Fatalf("Create() = nil, want error\nYAML was:\n%s", text)
			}
		})
	}
}
