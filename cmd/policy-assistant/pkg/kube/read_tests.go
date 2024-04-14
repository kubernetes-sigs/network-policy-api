package kube

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunReadNetworkPolicyTests() {
	Describe("ReadNetworkPolicies", func() {
		It("Should read a single policy from a single file", func() {
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../networkpolicies/features/portrange1.yaml")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(1))
		})
		It("Should read a list of policies from a single file", func() {
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../networkpolicies/yaml-syntax/yaml-list.yaml")
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
			policies, _, _, err := ReadNetworkPoliciesFromPath("../../networkpolicies/simple-example")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(7))

			policies, _, _, err = ReadNetworkPoliciesFromPath("../../networkpolicies/")
			Expect(err).To(BeNil())
			Expect(len(policies)).To(Equal(14))
		})

		It("Should read multiple admin network policies", func() {
			_, anps, _, err := ReadNetworkPoliciesFromPath("../../anps/")
			Expect(err).To(BeNil())
			Expect(len(anps)).To(Equal(3))
		})

		It("Should read a base admin network policy", func() {
			_, _, banp, err := ReadNetworkPoliciesFromPath("../../banp/")
			Expect(err).To(BeNil())
			Expect(banp).ToNot(BeNil())
		})

		// TODO test to show what happens for duplicate names
	})
}
