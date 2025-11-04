/*
Copyright 2025 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package crdtest

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"sigs.k8s.io/network-policy-api/apis/v1alpha2"
)

var (
	globals struct {
		k8sClient client.Client
	}

	watchDir    = flag.String("watch", "", "directory to continuously watch for yaml files to test the CRDs")
	crdDir      = flag.String("crdDir", "", "directory to load the CRDs from. If unset, will load from the default location")
	kubeConfig  = flag.String("kubeConfig", "", "KUBE_CONFIG for the cluster to use. If unset, will use internal testenv")
	prettyWatch = flag.Bool("prettyWatch", true, "pretty-print output to stdout if in -watch mode")
)

const usageText = `Tests the current CRDs against example valid and invalid CRDs. There
are two ways to use this:

# Test cases

Each test case takes its name from the name of the resource being
instantiated.

 - pkg/crdtest/testdata/{valid, invalid}

# Running

Run as a standard Go-test:

  $ go test ./pkg/crdtest

Run in -watch mode for development:

  $ go test -c -o crdtest ./pkg/crdtest # or "make crdtest"
  $ ./crdtest -crdDir config/crd/standard -watch pkg/crdtest/testdata/valid

will apply all yamls in pkg/crdtest/valid folder and continue watching
the changes to the *.yaml files in the folder. When a file is
changed, it will load it into an API server and print the result.

`

func runTestOnFile(path string) {
	klog.Infof("Testing file: %s", path)
	b, err := os.ReadFile(path)
	if err != nil {
		klog.Errorf("Failed to read file %s: %v", path, err)
		return
	}
	obj := &unstructured.Unstructured{}
	dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(b)), 4096)
	if err := dec.Decode(&obj); err != nil {
		if *prettyWatch {
			fmt.Printf("❌ %s: %v\n", path, err)
		}
		klog.Errorf("Failed to decode yaml from %s: %v", path, err)
		return
	}

	if *prettyWatch {
		fmt.Println("---")
		fmt.Println(string(b))
	}

	if obj.GetNamespace() == "" {
		klog.Infof("Applying %s", obj.GetName())
	} else {
		klog.Infof("Applying %s/%s", obj.GetNamespace(), obj.GetName())
	}

	if err := globals.k8sClient.Patch(context.Background(), obj, client.Apply, client.FieldOwner("crdtest")); err != nil {
		if *prettyWatch {
			fmt.Printf("❌ %s: %v\n", path, err)
		}
		klog.Errorf("Failed to apply object from %s: %v", path, err)
	} else {
		if *prettyWatch {
			fmt.Printf("✅ %s\n", path)
		}
		klog.Infof("Successfully applied object from %s", path)
		// Clean up the object.
		if err := globals.k8sClient.Delete(context.Background(), obj); err != nil {
			klog.Errorf("Failed to delete %s", obj.GetName())
		}
	}
	fmt.Println() // Add a newline to make it easier to read the output.
}

func watchAndTest(watchDir string) {
	klog.Info("Watching for changes... (press CTRL-c to exit)")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		todoMap := make(map[string]fsnotify.Event)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				flags := fsnotify.Create | fsnotify.Write
				if (event.Op&flags != 0) && !ignorePath(event.Name) {
					todoMap[event.Name] = event
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Errorf("Watcher error: %v", err)

			// We debounce the FSNotify events to avoid
			// printing out duplicate test results.
			case <-ticker.C:
				for _, ev := range todoMap {
					runTestOnFile(ev.Name)
				}
				// Clear map.
				todoMap = make(map[string]fsnotify.Event)
			}
		}
	}()

	err = watcher.Add(watchDir)
	if err != nil {
		klog.Fatalf("Failed to add dir to watcher: %v", err)
	}

	// Run tests on existing files
	files, err := os.ReadDir(watchDir)
	if err != nil {
		klog.Fatalf("Failed to read dir %s: %v", watchDir, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if ignorePath(file.Name()) {
			continue
		}
		runTestOnFile(filepath.Join(watchDir, file.Name()))
	}

	klog.Infof("Watching for changes in %s...", watchDir)
	select {} // Block forever
}

func ignorePath(path string) bool {
	// Ignore dotfiles (editor temporary files).
	if filepath.Base(path)[0] == '.' {
		return true
	}
	ext := filepath.Ext(path)
	if !(ext == ".yaml" || ext == ".yml") {
		return true
	}
	return false
}

func TestMain(m *testing.M) {
	klog.InitFlags(nil)

	oldUsage := flag.Usage
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
		if oldUsage != nil {
			oldUsage()
		}
		flag.PrintDefaults()
	}

	flag.Parse()

	klog.Info("Running crdtest")

	scheme := runtime.NewScheme()

	var (
		restConfig *rest.Config
		testEnv    *envtest.Environment
		err        error
	)

	corev1.AddToScheme(scheme)
	klog.Info("Added corev1 to scheme")

	v1alpha2.Install(scheme)
	klog.Info("Added v1alpha2 to scheme")

	if *kubeConfig != "" {
		klog.Infof("Using kubeConfig=%q", *kubeConfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to get restConfig from BuildConfigFromFlags: %v", err))
		}
	} else {
		klog.Info("Using testenv")

		// The version used here MUST reflect the available versions at
		// controller-runtime repo:
		//   https://raw.githubusercontent.com/kubernetes-sigs/controller-tools/HEAD/envtest-releases.yaml
		// If the envvar is not passed, the latest GA will be used
		k8sVersion := os.Getenv("K8S_VERSION")

		var paths []string
		if *crdDir != "" {
			paths = []string{*crdDir}
		} else {
			paths = []string{
				filepath.Join("..", "..", "config", "crd", "standard"),
			}
		}

		klog.Infof("Paths to CRDs: %v", paths)

		testEnv = &envtest.Environment{
			Scheme:                      scheme,
			ErrorIfCRDPathMissing:       true,
			DownloadBinaryAssets:        true,
			DownloadBinaryAssetsVersion: k8sVersion,
			CRDInstallOptions: envtest.CRDInstallOptions{
				Paths:           paths,
				CleanUpAfterUse: true,
			},
		}

		startTs := time.Now()
		restConfig, err = testEnv.Start()
		if err != nil {
			panic(fmt.Sprintf("Error initializing test environment: %v (took %v)", err, time.Since(startTs)))
		}
		klog.Infof("testEnv.Start() took %v", time.Since(startTs))
	}

	globals.k8sClient, err = client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		panic(fmt.Sprintf("Failed to get restConfig from BuildConfigFromFlags: %v", err))
	}

	if *watchDir != "" {
		watchAndTest(*watchDir)
		if testEnv != nil {
			if err := testEnv.Stop(); err != nil {
				panic(fmt.Sprintf("error stopping test environment: %v", err))
			}
		}
		os.Exit(0)
	}

	rc := m.Run()
	if testEnv != nil {
		if err := testEnv.Stop(); err != nil {
			panic(fmt.Sprintf("error stopping test environment: %v", err))
		}
	}

	os.Exit(rc)
}
