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

package crdtest

import (
	"embed"
	"path"
	"path/filepath"

	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//go:embed testdata
	testData embed.FS
)

const testDataDir = "testdata"

// channel is an API channel the CRDs are published in, paired with the
// testdata it must accept. Since the experimental CRDs are a superset of the
// standard ones, testdata for a channel is also exercised against every
// channel extending it. That way a field escaping its
// "<network-policy-api:experimental>" gate shows up as a standard-channel
// failure, rather than silently being pruned.
type channel struct {
	// name of the channel, used as the subtest name.
	name string
	// crdDir holds the CRDs of the channel, relative to this package.
	crdDir string
	// fixtures names the testdata subdirectories to apply, in order.
	fixtures []string
}

var (
	standardChannel = channel{
		name:     "standard",
		crdDir:   filepath.Join("..", "..", "config", "crd", "standard"),
		fixtures: []string{"standard"},
	}

	experimentalChannel = channel{
		name:     "experimental",
		crdDir:   filepath.Join("..", "..", "config", "crd", "experimental"),
		fixtures: []string{"standard", "experimental"},
	}

	channels = []channel{standardChannel, experimentalChannel}
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

func TestCRDs(t *testing.T) {
	// A single environment was configured explicitly with -crdDir or
	// -kubeConfig, so the channel it serves is unknown: apply every fixture
	// set and leave it to the caller to point at CRDs that support them.
	if globals.k8sClient != nil {
		testChannel(t, globals.k8sClient, allFixtures())
		return
	}

	for _, ch := range channels {
		t.Run(ch.name, func(t *testing.T) {
			k8sClient, stop, err := startEnv(ch.crdDir)
			if err != nil {
				t.Fatalf("startEnv(%s) = %v, want nil", ch.crdDir, err)
			}
			t.Cleanup(func() {
				if err := stop(); err != nil {
					t.Errorf("stop() = %v, want nil", err)
				}
			})
			testChannel(t, k8sClient, ch.fixtures)
		})
	}
}

// testChannel applies each fixture set against an API server that already has
// the CRDs of the channel under test installed.
func testChannel(t *testing.T, k8sClient client.Client, fixtures []string) {
	for _, f := range fixtures {
		t.Run(f, func(t *testing.T) {
			t.Run("valid", func(t *testing.T) { testValid(t, k8sClient, f) })
			t.Run("invalid", func(t *testing.T) { testInvalid(t, k8sClient, f) })
		})
	}
}

func testValid(t *testing.T, k8sClient client.Client, fixtures string) {
	dir := path.Join(testDataDir, fixtures, "valid")

	entries, err := testData.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() = %v, want nil", err)
	}
	for _, e := range entries {
		filePath := path.Join(dir, e.Name())
		obj, text, err := loadYAML(testData, filePath)
		if err != nil {
			t.Fatalf("loadYAML(%s) = %v, want nil\nYAML was:\n%s", filePath, err, text)
		}
		t.Run(obj.GetName(), func(t *testing.T) {
			if err := k8sClient.Create(context.Background(), obj); err != nil {
				t.Fatalf("Create() = %v, want nil\nYAML was:\n%s", err, text)
			}
			if err := k8sClient.Delete(context.Background(), obj); err != nil {
				t.Fatalf("Delete() = %v, want nil\nYAML was:\n%s", err, text)
			}
		})
	}
}

func testInvalid(t *testing.T, k8sClient client.Client, fixtures string) {
	dir := path.Join(testDataDir, fixtures, "invalid")

	entries, err := testData.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() = %v, want nil", err)
	}
	for _, e := range entries {
		filePath := path.Join(dir, e.Name())
		obj, text, err := loadYAML(testData, filePath)
		if err != nil {
			t.Fatalf("loadYAML(%s) = %v, want nil\nYAML was:\n%s", filePath, err, text)
		}
		t.Run(obj.GetName(), func(t *testing.T) {
			err := k8sClient.Create(context.Background(), obj)
			t.Logf("Create() = %v", err)
			if err == nil {
				t.Fatalf("Create() = nil, want error\nYAML was:\n%s", text)
			}
		})
	}
}

// allFixtures returns every fixture set exactly once, in channel order.
func allFixtures() []string {
	var (
		all  []string
		seen = make(map[string]bool)
	)
	for _, ch := range channels {
		for _, f := range ch.fixtures {
			if seen[f] {
				continue
			}
			seen[f] = true
			all = append(all, f)
		}
	}
	return all
}
