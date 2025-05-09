/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	gentype "k8s.io/client-go/gentype"
	v1alpha1 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
	apisv1alpha1 "sigs.k8s.io/network-policy-api/pkg/client/applyconfiguration/apis/v1alpha1"
	typedapisv1alpha1 "sigs.k8s.io/network-policy-api/pkg/client/clientset/versioned/typed/apis/v1alpha1"
)

// fakeAdminNetworkPolicies implements AdminNetworkPolicyInterface
type fakeAdminNetworkPolicies struct {
	*gentype.FakeClientWithListAndApply[*v1alpha1.AdminNetworkPolicy, *v1alpha1.AdminNetworkPolicyList, *apisv1alpha1.AdminNetworkPolicyApplyConfiguration]
	Fake *FakePolicyV1alpha1
}

func newFakeAdminNetworkPolicies(fake *FakePolicyV1alpha1) typedapisv1alpha1.AdminNetworkPolicyInterface {
	return &fakeAdminNetworkPolicies{
		gentype.NewFakeClientWithListAndApply[*v1alpha1.AdminNetworkPolicy, *v1alpha1.AdminNetworkPolicyList, *apisv1alpha1.AdminNetworkPolicyApplyConfiguration](
			fake.Fake,
			"",
			v1alpha1.SchemeGroupVersion.WithResource("adminnetworkpolicies"),
			v1alpha1.SchemeGroupVersion.WithKind("AdminNetworkPolicy"),
			func() *v1alpha1.AdminNetworkPolicy { return &v1alpha1.AdminNetworkPolicy{} },
			func() *v1alpha1.AdminNetworkPolicyList { return &v1alpha1.AdminNetworkPolicyList{} },
			func(dst, src *v1alpha1.AdminNetworkPolicyList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.AdminNetworkPolicyList) []*v1alpha1.AdminNetworkPolicy {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.AdminNetworkPolicyList, items []*v1alpha1.AdminNetworkPolicy) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
