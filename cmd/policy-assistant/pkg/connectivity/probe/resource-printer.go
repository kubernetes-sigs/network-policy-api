package probe

import (
	"fmt"
	"strings"

	"github.com/mattfenwick/collections/pkg/slice"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/generator"
)

func (r *Resources) RenderTable() string {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetHeader([]string{"Namespace", "NS Labels", "Pod", "Pod Labels", "IPs", "Containers/Ports"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	nsToPod := map[string][]*Pod{}
	for _, pod := range r.Pods {
		ns := pod.Namespace
		if _, ok := nsToPod[ns]; !ok {
			nsToPod[ns] = []*Pod{}
		}
		if _, ok := r.Namespaces[ns]; !ok {
			panic(errors.Errorf("cannot handle pod %s/%s: namespace not found", ns, pod.Name))
		}
		nsToPod[ns] = append(nsToPod[ns], pod)
	}

	namespaces := slice.Sort(maps.Keys(nsToPod))
	for _, ns := range namespaces {
		labels := r.Namespaces[ns]
		nsLabelLines := labelsToLines(labels)
		for _, pod := range nsToPod[ns] {
			podLabelLines := labelsToLines(pod.Labels)
			for _, cont := range pod.Containers {
				ips := []string{
					fmt.Sprintf("pod: %s", pod.IP),
					fmt.Sprintf("service: %s", pod.ServiceIP),
				}
				if pod.ExternalServiceIPs != nil {
					for _, kind := range []generator.ServiceKind{generator.LoadBalancerCluster, generator.LoadBalancerLocal} {
						if ip, ok := pod.ExternalServiceIPs[kind]; ok {
							ips = append(ips, fmt.Sprintf("%s: %s", kind, ip))
						}
					}
				}
				ips = append(ips, fmt.Sprintf("node: %s", pod.LocalNodeIP))

				ports := []string{
					fmt.Sprintf("%s, port %s: %d on %s", cont.Name, cont.PortName, cont.Port, cont.Protocol),
				}
				for _, kind := range []generator.ServiceKind{generator.NodePortCluster, generator.NodePortLocal} {
					nodePort := pod.NodePort(kind, cont.Port)
					if nodePort != 0 {
						ports = append(ports, fmt.Sprintf("%s: %d", kind, nodePort))
					}
				}

				table.Append([]string{
					ns,
					nsLabelLines,
					pod.Name,
					podLabelLines,
					strings.Join(ips, "\n"),
					strings.Join(ports, "\n"),
				})
			}
		}
	}

	table.Render()
	return tableString.String()
}

func labelsToLines(labels map[string]string) string {
	keys := slice.Sort(maps.Keys(labels))
	var lines []string
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, labels[key]))
	}
	return strings.Join(lines, "\n")
}
