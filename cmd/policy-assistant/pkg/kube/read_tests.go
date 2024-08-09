package kube

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunReadNetworkPolicyTests() {
	Describe("ReadNetworkPolicies", func() {
		It("Should read a single policy from a single file", func() {
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../test/example-policies/networkpolicies/features/portrange1.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(1))
		})
		It("Should read a list of policies from a single file", func() {
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../test/example-policies/networkpolicies/yaml-syntax/yaml-list.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(3))
		})

		// TODO test case to read multiple policies from plain yaml list

		// TODO
		// It("Should read multiple policies separated by '---' lines from a single file", func() {
		// 	policies, err := ReadNetworkPoliciesFromPath("../../networkpolicies/yaml-syntax/triple-dash-separated.yaml")
		// 	Expect(err).To(BeNil())
		// 	Expect(len(policies)).To(Equal(3))
		// })

		It("Should read multiple policies from all files in a directory", func() {
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../test/example-policies/networkpolicies/simple-example")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(7))

			policies, _, _, err = ReadNetworkPoliciesFromPath("../../test/example-policies/networkpolicies/")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(14))
		})

		It("Should read multiple admin network policies", func() {
			_, anps, _, err := ReadNetworkPoliciesFromPath("../../test/example-policies/anps/")
			Expect(err).To(BeNil())
			Expect(len(anps)).To(Equal(3))
		})

		It("Should read a base admin network policy", func() {
			_, _, banp, err := ReadNetworkPoliciesFromPath("../../test/example-policies/banp/")
			Expect(err).To(BeNil())
			Expect(banp).ToNot(BeNil())
		})

		It("Should parse multiple types from folder", func() {
			policiies, anps, bapn, err := ReadNetworkPoliciesFromPath("../../test/example-policies/")
			Expect(err).To(BeNil())
			Expect(len(policiies)).To(Equal(14))
			Expect(len(anps)).To(Equal(3))
			Expect(bapn).ToNot(BeNil())
		})

		// TODO test to show what happens for duplicate names
	})
}
