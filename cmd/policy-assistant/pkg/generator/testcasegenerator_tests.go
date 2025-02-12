package generator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RunTestCaseGeneratorTests() {
	Describe("TestCaseGenerator", func() {
		It("Overall number of test cases", func() {
			gen := NewTestCaseGenerator(true, []string{"x", "y", "z"}, []string{}, []string{})

			Expect(len(gen.PeersTestCases("1.2.3.4"))).To(Equal(112))
			Expect(len(gen.ActionTestCases())).To(Equal(6))
			Expect(len(gen.RulesTestCases())).To(Equal(4))
			Expect(len(gen.UpstreamE2ETestCases())).To(Equal(13))
			Expect(len(gen.TargetTestCases())).To(Equal(6))
			Expect(len(gen.ExampleTestCases())).To(Equal(1))
			Expect(len(gen.PortProtocolTestCases())).To(Equal(70))
			Expect(len(gen.ConflictTestCases())).To(Equal(16))
			Expect(len(gen.NamespaceTestCases())).To(Equal(2))
			Expect(len(gen.SpecialServiceCases([]string{"10.224.0.5", "10.224.0.6"}))).To(Equal(20))

			Expect(len(gen.GenerateTestCases("1.2.3.4", []string{"10.224.0.5", "10.224.0.6"}))).To(Equal(250))
		})
	})
}
