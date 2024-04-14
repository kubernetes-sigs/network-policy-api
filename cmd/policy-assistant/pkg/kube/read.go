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
	var adminNetPolicies []*v1alpha12.AdminNetworkPolicy
	var baseAdminNetPolicies *v1alpha12.BaselineAdminNetworkPolicy

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
			baseAdminNetPolicies = banp
			return nil
		}
		logrus.Debugf("unable to base admin network policies: %+v", err)

		anpList, err := utils.ParseYamlStrict[v1alpha12.AdminNetworkPolicyList](bytes)
		if err == nil {
			adminNetPolicies = append(adminNetPolicies, refList(anpList.Items)...)
			return nil
		}
		logrus.Debugf("unable to parse list of admin network policies: %+v", err)

		anp, err := utils.ParseYamlStrict[v1alpha12.AdminNetworkPolicy](bytes)
		if err == nil {
			adminNetPolicies = append(adminNetPolicies, anp)
			return nil
		}
		logrus.Debugf("unable to single admin network policies: %+v", err)

		if len(netPolicies) == 0 && len(adminNetPolicies) == 0 && baseAdminNetPolicies == nil {
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
	return netPolicies, adminNetPolicies, baseAdminNetPolicies, nil
}

func ReadNetworkPoliciesFromKube(ctx context.Context, kubeClient IKubernetes, namespaces []string) ([]*networkingv1.NetworkPolicy, []*v1alpha12.AdminNetworkPolicy, *v1alpha12.BaselineAdminNetworkPolicy, error, error, error) {
	var netpols []*networkingv1.NetworkPolicy
	var anps []*v1alpha12.AdminNetworkPolicy
	var banp *v1alpha12.BaselineAdminNetworkPolicy
	var neterr, anperr, banperr error

	var netPolsCn = make(chan apiResponse[[]networkingv1.NetworkPolicy], 1)
	go func(ch chan apiResponse[[]networkingv1.NetworkPolicy]) {
		ch <- readNetworkPolicies(ctx, kubeClient, namespaces)
	}(netPolsCn)

	var anpsCh = make(chan apiResponse[[]v1alpha12.AdminNetworkPolicy], 1)
	go func(ch chan apiResponse[[]v1alpha12.AdminNetworkPolicy]) {
		ch <- readAdminNetworkPolicies(ctx, kubeClient)
	}(anpsCh)

	var banpsCh = make(chan apiResponse[v1alpha12.BaselineAdminNetworkPolicy], 1)
	go func(ch chan apiResponse[v1alpha12.BaselineAdminNetworkPolicy]) {
		ch <- readBaseAdminNetworkPolicies(ctx, kubeClient)
	}(banpsCh)

	for i := 0; i <= 2; i++ {
		select {
		case result := <-netPolsCn:
			r, err := result.response()
			netpols = refList(r)
			neterr = err
		case result := <-anpsCh:
			r, err := result.response()
			anps = refList(r)
			anperr = err
		case result := <-banpsCh:
			r, err := result.response()
			if err == nil {
				banp = &r
			}
			banperr = err
		}
	}

	return netpols, anps, banp, neterr, anperr, banperr
}

func refList[T any](refs []T) []*T {
	return slice.Map(builtin.Reference[T], refs)
}

type apiResponse[T []networkingv1.NetworkPolicy | []v1alpha12.AdminNetworkPolicy | v1alpha12.BaselineAdminNetworkPolicy] struct {
	data  T
	error error
}

func (t apiResponse[T]) response() (T, error) {
	return t.data, t.error
}

func readNetworkPolicies(ctx context.Context, kubeClient IKubernetes, namespaces []string) apiResponse[[]networkingv1.NetworkPolicy] {
	result, err := GetNetworkPoliciesInNamespaces(ctx, kubeClient, namespaces)
	if err != nil {
		return apiResponse[[]networkingv1.NetworkPolicy]{nil, err}
	}
	return apiResponse[[]networkingv1.NetworkPolicy]{result, err}
}

func readAdminNetworkPolicies(ctx context.Context, kubeClient IKubernetes) apiResponse[[]v1alpha12.AdminNetworkPolicy] {
	result, err := GetAdminNetworkPoliciesInNamespaces(ctx, kubeClient)
	if err != nil {
		return apiResponse[[]v1alpha12.AdminNetworkPolicy]{nil, err}
	}
	return apiResponse[[]v1alpha12.AdminNetworkPolicy]{result, err}
}

func readBaseAdminNetworkPolicies(ctx context.Context, kubeClient IKubernetes) apiResponse[v1alpha12.BaselineAdminNetworkPolicy] {
	result, err := GetBaseAdminNetworkPoliciesInNamespaces(ctx, kubeClient)
	if err != nil {
		return apiResponse[v1alpha12.BaselineAdminNetworkPolicy]{result, err}
	}
	return apiResponse[v1alpha12.BaselineAdminNetworkPolicy]{result, err}
}
