package connectivity

import (
	"testing"

	"github.com/mattfenwick/cyclonus/pkg/cli"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/stretchr/testify/require"
)

func TestProbe(t *testing.T) {
	t.Run("probe works", func(t *testing.T) {
		npv1, anp, banp, err := kube.ReadNetworkPoliciesFromPath("../../examples/demos/kubecon-eu-2024/policies/")
		require.Nil(t, err)

		policies := matcher.BuildV1AndV2NetPols(false, npv1, anp, banp)

		cli.ProbeSyntheticConnectivity(policies, "../../examples/demos/kubecon-eu-2024/demo-probe.json", nil, nil)

		cli.RunAnalyzeCommand(&cli.AnalyzeArgs{
			PolicyPath: "../../examples/demos/kubecon-eu-2024/policies/",
			ProbePath:  "../../examples/demos/kubecon-eu-2024/demo-probe.json",
		})
	})
}
