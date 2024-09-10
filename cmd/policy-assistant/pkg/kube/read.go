package kube

import (
	"context"
	"github.com/mattfenwick/collections/pkg/builtin"
	"github.com/mattfenwick/collections/pkg/file"
	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/mattfenwick/cyclonus/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"os"
	"path/filepath"
	v1alpha12 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
	"sync"
)

// ReadNetworkPoliciesFromPath walks the folder and try to parse each file in
// one of the supported types in the following manner:
// 1. NetworkPolicyList
// 2. NetworkPolicy
// 3. BaselineAdminNetworkPolicy
// 4. AdminNetworkPolicyList
// 5. AdminNetworkPolicy
func ReadNetworkPoliciesFromPath(policyPath string) ([]*networkingv1.NetworkPolicy, []*v1alpha12.AdminNetworkPolicy, *v1alpha12.BaselineAdminNetworkPolicy, error) {
	var netPolicies []*networkingv1.NetworkPolicy
	var adminNetworkPolicies []*v1alpha12.AdminNetworkPolicy
	var baselineAdminNetworkPolicy *v1alpha12.BaselineAdminNetworkPolicy

	err := filepath.Walk(policyPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "unable to walk path %s", path)
		}
		if info.IsDir() {
			logrus.Tracef("not opening dir %s", path)
			return nil
		}
		logrus.Debugf("walking path %s", path)
		bytes, err := file.Read(path)
		if err != nil {
			return err
		}

		// TODO try parsing plain yaml list (that is: not a NetworkPolicyList)
		// policies, err := utils.ParseYaml[[]*networkingv1.NetworkPolicy](bytes)

		// TODO try parsing multiple policies separated by '---' lines
		// policies, err := yaml.ParseMany[networkingv1.NetworkPolicy](bytes)
		// if err == nil {
		// 	logrus.Debugf("parsed %d policies from %s", len(policies), path)
		// 	netPolicies = append(netPolicies, refNetpolList(policies)...)
		// 	return nil
		// }
		// logrus.Errorf("unable to parse multiple policies separated by '---' lines: %+v", err)

		// try parsing a NetworkPolicyList
		policyList, err := utils.ParseYamlStrict[networkingv1.NetworkPolicyList](bytes)
		if err == nil {
			netPolicies = append(netPolicies, refList(policyList.Items)...)
			return nil
		}
		logrus.Debugf("unable to parse list of network policies: %+v", err)

		policy, err := utils.ParseYamlStrict[networkingv1.NetworkPolicy](bytes)
		if err == nil {
			netPolicies = append(netPolicies, policy)
			return nil
		}
		logrus.Debugf("unable to parse network policy: %+v", err)

		banp, err := utils.ParseYamlStrict[v1alpha12.BaselineAdminNetworkPolicy](bytes)
		if err == nil {
			if baselineAdminNetworkPolicy != nil {
				return errors.New("baseline admin network policy already exists")
			}
			baselineAdminNetworkPolicy = banp
			return nil
		}
		logrus.Debugf("unable to base admin network policies: %+v", err)

		anpList, err := utils.ParseYamlStrict[v1alpha12.AdminNetworkPolicyList](bytes)
		if err == nil {
			adminNetworkPolicies = append(adminNetworkPolicies, refList(anpList.Items)...)
			return nil
		}
		logrus.Debugf("unable to parse list of admin network policies: %+v", err)

		anp, err := utils.ParseYamlStrict[v1alpha12.AdminNetworkPolicy](bytes)
		if err == nil {
			adminNetworkPolicies = append(adminNetworkPolicies, anp)
			return nil
		}
		logrus.Debugf("unable to single admin network policies: %+v", err)

		if len(netPolicies) == 0 && len(adminNetworkPolicies) == 0 && baselineAdminNetworkPolicy == nil {
			return errors.WithMessagef(err, "unable to parse any policies from yaml at %s", path)
		}

		return nil
	})
	if err != nil {
		return nil, nil, nil, err
		//return nil, errors.Wrapf(err, "unable to walk filesystem from %s", policyPath)
	}
	if len(netPolicies) > 0 {
		for _, p := range netPolicies {
			if len(p.Spec.PolicyTypes) == 0 {
				return nil, nil, nil, errors.Errorf("missing spec.policyTypes from network policy %s/%s", p.Namespace, p.Name)
			}
		}
	}
	return netPolicies, adminNetworkPolicies, baselineAdminNetworkPolicy, nil
}

func refList[T any](refs []T) []*T {
	return slice.Map(builtin.Reference[T], refs)
}

func ReadNetworkPoliciesFromKube(ctx context.Context, kubeClient IKubernetes, namespaces []string, includeANPs, includeBANPs bool) ([]*networkingv1.NetworkPolicy, []*v1alpha12.AdminNetworkPolicy, *v1alpha12.BaselineAdminNetworkPolicy, error, error, error) {
	var netpols []networkingv1.NetworkPolicy
	var anps []v1alpha12.AdminNetworkPolicy
	var banp *v1alpha12.BaselineAdminNetworkPolicy
	var netErr, anpErr, banpErr error

	var wg sync.WaitGroup
	wg.Add(3)

	go func(w *sync.WaitGroup) {
		defer w.Done()
		netpols, netErr = GetNetworkPoliciesInNamespaces(ctx, kubeClient, namespaces)
		return
	}(&wg)

	go func(w *sync.WaitGroup) {
		defer w.Done()
		if !includeANPs {
			return
		}
		anps, anpErr = GetAdminNetworkPolicies(ctx, kubeClient)
		return
	}(&wg)

	go func(w *sync.WaitGroup) {
		defer w.Done()
		if !includeBANPs {
			return
		}
		banp, banpErr = GetBaselineAdminNetworkPolicy(ctx, kubeClient)
		return
	}(&wg)

	wg.Wait()

	return refList(netpols), refList(anps), banp, netErr, anpErr, banpErr
}
