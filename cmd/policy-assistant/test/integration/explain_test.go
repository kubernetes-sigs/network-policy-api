package connectivity

import (
	"github.com/mattfenwick/cyclonus/examples"
	"github.com/mattfenwick/cyclonus/pkg/kube/netpol"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExplain(t *testing.T) {
	t.Run("prints network policies v1", func(t *testing.T) {
		expected := "+---------+--------------------+----------------------------------------------------------------------------+----------------------+-------------------------+---------------------------+\n" +
			"|  TYPE   |      SUBJECT       |                                SOURCE RULES                                |         PEER         |         ACTION          |       PORT/PROTOCOL       |\n" +
			"+---------+--------------------+----------------------------------------------------------------------------+----------------------+-------------------------+---------------------------+\n" +
			"| Ingress | Namespace:         | [NPv1] default/accidental-and                                              | Namespace:           | NPv1: All peers allowed | all ports, all protocols  |\n" +
			"|         |    default         | [NPv1] default/accidental-or                                               |    default           |                         |                           |\n" +
			"|         | Pod:               |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |    a = b           |                                                                            |    role = client     |                         |                           |\n" +
			"+         +                    +                                                                            +----------------------+                         +                           +\n" +
			"|         |                    |                                                                            | Namespace:           |                         |                           |\n" +
			"|         |                    |                                                                            |    user = alice      |                         |                           |\n" +
			"|         |                    |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |                    |                                                                            |    role = client     |                         |                           |\n" +
			"+         +                    +                                                                            +----------------------+                         +                           +\n" +
			"|         |                    |                                                                            | Namespace:           |                         |                           |\n" +
			"|         |                    |                                                                            |    user = alice      |                         |                           |\n" +
			"|         |                    |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |                    |                                                                            |    all               |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-nothing-to-v2-all-web                                 | no pods, no ips      |                         | no ports, no protocols    |\n" +
			"|         |    default         |                                                                            |                      |                         |                           |\n" +
			"|         | Pod:               |                                                                            |                      |                         |                           |\n" +
			"|         |    all = web       |                                                                            |                      |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-specific-port-from-role-monitoring-to-app-apiserver   | Namespace:           |                         | port 5000 on protocol TCP |\n" +
			"|         |    default         |                                                                            |    default           |                         |                           |\n" +
			"|         | Pod:               |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |    app = apiserver |                                                                            |    role = monitoring |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-from-app-bookstore-to-app-bookstore-role-api          | Namespace:           |                         | all ports, all protocols  |\n" +
			"|         |    default         |                                                                            |    default           |                         |                           |\n" +
			"|         | Pod:               |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |    app = bookstore |                                                                            |    app = bookstore   |                         |                           |\n" +
			"|         |    role = api      |                                                                            |                      |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +                           +\n" +
			"|         | Namespace:         | [NPv1] default/allow-from-multiple-to-app-bookstore-role-db                | Namespace:           |                         |                           |\n" +
			"|         |    default         |                                                                            |    default           |                         |                           |\n" +
			"|         | Pod:               |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |    app = bookstore |                                                                            |    app = bookstore   |                         |                           |\n" +
			"|         |    role = db       |                                                                            |    role = api        |                         |                           |\n" +
			"+         +                    +                                                                            +----------------------+                         +                           +\n" +
			"|         |                    |                                                                            | Namespace:           |                         |                           |\n" +
			"|         |                    |                                                                            |    default           |                         |                           |\n" +
			"|         |                    |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |                    |                                                                            |    app = bookstore   |                         |                           |\n" +
			"|         |                    |                                                                            |    role = search     |                         |                           |\n" +
			"+         +                    +                                                                            +----------------------+                         +                           +\n" +
			"|         |                    |                                                                            | Namespace:           |                         |                           |\n" +
			"|         |                    |                                                                            |    default           |                         |                           |\n" +
			"|         |                    |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |                    |                                                                            |    app = inventory   |                         |                           |\n" +
			"|         |                    |                                                                            |    role = web        |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-nothing                                               | no pods, no ips      |                         | no ports, no protocols    |\n" +
			"|         |    default         |                                                                            |                      |                         |                           |\n" +
			"|         | Pod:               |                                                                            |                      |                         |                           |\n" +
			"|         |    app = foo       |                                                                            |                      |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-all-to-app-web                                        | all pods, all ips    |                         | all ports, all protocols  |\n" +
			"|         |    default         | [NPv1] default/allow-all-to-version2-app-web                               |                      |                         |                           |\n" +
			"|         | Pod:               | [NPv1] default/allow-all-to-version3-app-web                               |                      |                         |                           |\n" +
			"|         |    app = web       | [NPv1] default/allow-all-to-version4-app-web                               |                      |                         |                           |\n" +
			"|         |                    | [NPv1] default/allow-from-anywhere-to-app-web                              |                      |                         |                           |\n" +
			"|         |                    | [NPv1] default/allow-from-namespace-to-app-web                             |                      |                         |                           |\n" +
			"|         |                    | [NPv1] default/allow-from-namespace-with-labels-type-monitoring-to-app-web |                      |                         |                           |\n" +
			"|         |                    | [NPv1] default/allow-nothing-to-app-web                                    |                      |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +                           +\n" +
			"|         | Namespace:         | [NPv1] default/allow-all-within-namespace                                  | Namespace:           |                         |                           |\n" +
			"|         |    default         | [NPv1] default/allow-nothing-to-anything                                   |    default           |                         |                           |\n" +
			"|         | Pod:               |                                                                            | Pod:                 |                         |                           |\n" +
			"|         |    all pods        |                                                                            |    all               |                         |                           |\n" +
			"+---------+--------------------+----------------------------------------------------------------------------+----------------------+-------------------------+---------------------------+\n" +
			"|         |                    |                                                                            |                      |                         |                           |\n" +
			"+---------+--------------------+----------------------------------------------------------------------------+----------------------+-------------------------+---------------------------+\n" +
			"| Egress  | Namespace:         | [NPv1] default/allow-egress-on-port-app-foo                                | all pods, all ips    | NPv1: All peers allowed | port 53 on protocol TCP   |\n" +
			"|         |    default         | [NPv1] default/allow-egress-to-all-namespace-from-app-foo-on-port-53       |                      |                         | port 53 on protocol UDP   |\n" +
			"|         | Pod:               | [NPv1] default/allow-no-egress-from-labels-app-foo                         |                      |                         |                           |\n" +
			"|         |    app = foo       | [NPv1] default/allow-nothing                                               |                      |                         |                           |\n" +
			"+         +--------------------+----------------------------------------------------------------------------+----------------------+                         +---------------------------+\n" +
			"|         | Namespace:         | [NPv1] default/allow-no-egress-from-namespace                              | no pods, no ips      |                         | no ports, no protocols    |\n" +
			"|         |    default         |                                                                            |                      |                         |                           |\n" +
			"|         | Pod:               |                                                                            |                      |                         |                           |\n" +
			"|         |    all pods        |                                                                            |                      |                         |                           |\n" +
			"+---------+--------------------+----------------------------------------------------------------------------+----------------------+-------------------------+---------------------------+\n" +
			""
		policies := matcher.BuildV1AndV2NetPols(true, netpol.AllExamples, nil, nil)
		require.Equal(t, expected, policies.ExplainTable())
	})

	t.Run("prints network ANPs and BANPs", func(t *testing.T) {
		expected := "+---------+------------------------------------------+-----------------------------+------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|  TYPE   |                 SUBJECT                  |        SOURCE RULES         |                                  PEER                                  |                              ACTION                              |       PORT/PROTOCOL        |\n" +
			"+---------+------------------------------------------+-----------------------------+------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"| Ingress | Namespace:                               | [ANP] default/example-anp   | Namespace:                                                             | ANP:                                                             | all ports, all protocols   |\n" +
			"|         |    kubernetes.io/metadata.name Exists [] | [ANP] default/example-anp-2 |    kubernetes.io/metadata.name = network-policy-conformance-hufflepuff |    pri=16 (example-anp-2): Deny                                  |                            |\n" +
			"|         |                                          | [BANP] default/default      | Pod:                                                                   |    pri=20 (example-anp): Deny                                    |                            |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Deny                                                          |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+                            +\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             |                            |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-ravenclaw  |    pri=16 (example-anp-2): Allow (ineffective rules: Deny, Pass) |                            |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Allow (ineffective rules: Deny, Pass)   |                            |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Allow                                                         |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             | port 80 on protocol TCP    |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-slytherin  |    pri=16 (example-anp-2): Deny (ineffective rules: Pass)        | port 53 on protocol UDP    |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Deny (ineffective rules: Pass)          | port 9003 on protocol SCTP |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Deny                                                          |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             | port 80 on protocol TCP    |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-hufflepuff |    pri=16 (example-anp-2): Allow                                 | port 5353 on protocol UDP  |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Allow                                   | port 9003 on protocol SCTP |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Allow                                                         |                            |\n" +
			"+---------+------------------------------------------+-----------------------------+------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|         |                                          |                             |                                                                        |                                                                  |                            |\n" +
			"+---------+------------------------------------------+-----------------------------+------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"| Egress  | Namespace:                               | [ANP] default/example-anp   | Namespace:                                                             | BANP:                                                            | all ports, all protocols   |\n" +
			"|         |    kubernetes.io/metadata.name Exists [] | [ANP] default/example-anp-2 |    Not Same labels - Test1                                             |    Deny                                                          |                            |\n" +
			"|         |                                          | [BANP] default/default      | Pod:                                                                   |                                                                  |                            |\n" +
			"|         |                                          |                             |    all                                                                 |                                                                  |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+                            +\n" +
			"|         |                                          |                             | Namespace:                                                             | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |    Same labels - Test                                                  |    Allow                                                         |                            |\n" +
			"|         |                                          |                             | Pod:                                                                   |                                                                  |                            |\n" +
			"|         |                                          |                             |    all                                                                 |                                                                  |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+                            +\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             |                            |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-hufflepuff |    pri=16 (example-anp-2): Deny                                  |                            |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Deny                                    |                            |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Deny                                                          |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+                            +\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             |                            |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-ravenclaw  |    pri=16 (example-anp-2): Allow (ineffective rules: Deny, Pass) |                            |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Allow (ineffective rules: Deny, Pass)   |                            |\n" +
			"|         |                                          |                             |    all                                                                 |                                                                  |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             | port 80 on protocol TCP    |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-slytherin  |    pri=16 (example-anp-2): Deny (ineffective rules: Pass)        | port 53 on protocol UDP    |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Deny (ineffective rules: Pass)          | port 9003 on protocol SCTP |\n" +
			"|         |                                          |                             |    all                                                                 |                                                                  |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+                            +\n" +
			"|         |                                          |                             | Namespace:                                                             | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name Exists []                               |    Deny                                                          |                            |\n" +
			"|         |                                          |                             | Pod:                                                                   |                                                                  |                            |\n" +
			"|         |                                          |                             |    all                                                                 |                                                                  |                            |\n" +
			"+         +                                          +                             +------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			"|         |                                          |                             | Namespace:                                                             | ANP:                                                             | port 8080 on protocol TCP  |\n" +
			"|         |                                          |                             |    kubernetes.io/metadata.name = network-policy-conformance-hufflepuff |    pri=16 (example-anp-2): Allow                                 | port 5353 on protocol UDP  |\n" +
			"|         |                                          |                             | Pod:                                                                   |    pri=20 (example-anp): Allow                                   | port 9003 on protocol SCTP |\n" +
			"|         |                                          |                             |    all                                                                 | BANP:                                                            |                            |\n" +
			"|         |                                          |                             |                                                                        |    Allow                                                         |                            |\n" +
			"+---------+------------------------------------------+-----------------------------+------------------------------------------------------------------------+------------------------------------------------------------------+----------------------------+\n" +
			""
		policies := matcher.BuildV1AndV2NetPols(false, nil, examples.CoreGressRulesCombinedANB, examples.CoreGressRulesCombinedBANB)
		require.Equal(t, expected, policies.ExplainTable())
	})
}
