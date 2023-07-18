package kubernetes

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/network-policy-api/conformance/utils/config"
)

var (
	numStatufulSetReplicas int32 = 2
)

// PokeServer is a utility function that checks if the connection from the provided clientPod in clientNamespace towards the targetHost:targetPort
// using the provided protocol can be established or not and returns the result based on if the expectation is shouldConnect or !shouldConnect
func PokeServer(t *testing.T, clientNamespace, clientPod, protocol, targetHost string, targetPort int32, timeout time.Duration, shouldConnect bool) bool {
	t.Helper()
	cmd := []string{"exec", clientPod, "--"} // command is to be run inside a pod
	timeoutArg := fmt.Sprintf("--timeout=%v", timeout)
	protocolArg := fmt.Sprintf("--protocol=%s", protocol)
	ipPortArg := net.JoinHostPort(targetHost, fmt.Sprintf("%d", targetPort))

	// we leverage the dial connect from agnhost, that is already supporting multiple protocols
	connectCommand := strings.Split(fmt.Sprintf("/agnhost connect %s %s %s",
		timeoutArg,
		protocolArg,
		ipPortArg), " ")

	cmd = append(cmd, connectCommand...)
	var res string
	var err error
	res, err = e2ekubectl.RunKubectl(clientNamespace, cmd...)
	// TODO(tssurya): Improve the error matching to be more specific (https://pkg.go.dev/k8s.io/kubernetes/test/images/agnhost#readme-connect)
	// A connection error looks like this and we need to improve the parsing to be more accurate:
	// Command stdout:
	// stderr:
	// TIMEOUT
	// command terminated with exit code 1
	// error:
	// exit status 1
	// TODO(tssurya): See if we need to add a wait&retry mechanism to test connectivity
	// See https://github.com/kubernetes-sigs/network-policy-api/issues/108 for details.
	if shouldConnect && (err != nil || len(res) > 0) {
		framework.Logf("FAILED Command was %s", connectCommand)
		framework.Logf("FAILED Response was %v, expected connection to succeed from %s to %s, "+
			"but instead it miserably failed: %v", res, clientPod, targetHost, err.Error())
		return false
	} else if !shouldConnect {
		if err == nil && len(res) == 0 {
			framework.Logf("FAILED Command was %s", connectCommand)
			framework.Logf("FAILED Response was %v, expected connection to fail from %s to %s, "+
				"but instead it successfully connected", res, clientPod, targetHost)
			return false
		} else if strings.Contains(err.Error(), "TIMEOUT") {
			framework.Logf("error contained 'TIMEOUT', as expected: %s", err.Error())
			return true
		} else {
			framework.Logf("error didn't contain 'TIMEOUT', as expected: %s", err.Error())
			return false
		}
	}
	return true
}

// NamespacesMustBeReady waits until all Pods are marked Ready. This will
// cause the test to halt if the specified timeout is exceeded.
func NamespacesMustBeReady(t *testing.T, c client.Client, timeoutConfig config.TimeoutConfig, namespaces []string, statefulSetNames []string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.NamespacesMustBeReady)
	defer cancel()

	waitErr := wait.PollImmediate(1*time.Second, timeoutConfig.NamespacesMustBeReady, func() (bool, error) {
		for i, ns := range namespaces {
			statefulSet := &appsv1.StatefulSet{}
			statefulSetKey := types.NamespacedName{
				Namespace: ns,
				Name:      statefulSetNames[i],
			}
			if err := c.Get(ctx, statefulSetKey, statefulSet); err != nil {
				t.Errorf("Error retrieving StatefulSet %s from namespace %s: %v", statefulSetNames[i], ns, err)
			}
			if statefulSet.Status.ReadyReplicas != numStatufulSetReplicas {
				t.Logf("StatefulSet replicas in namespace %s not rolled out yet. %d/%d replicas are available.", ns, statefulSet.Status.ReadyReplicas, numStatufulSetReplicas)
				return false, nil
			}
		}
		t.Logf("Namespaces and Pods in %s namespaces are ready", strings.Join(namespaces, ", "))
		return true, nil
	})
	require.NoErrorf(t, waitErr, "error waiting for %s namespaces to be ready", strings.Join(namespaces, ", "))
}
