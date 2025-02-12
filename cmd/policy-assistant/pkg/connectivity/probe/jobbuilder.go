package probe

import (
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/network-policy-api/policy-assistant/pkg/generator"
)

type JobBuilder struct {
	TimeoutSeconds int
}

func (j *JobBuilder) GetJobsForProbeConfig(resources *Resources, config *generator.ProbeConfig) *Jobs {
	if config.AllAvailable {
		return j.GetJobsAllAvailableServers(resources, config)
	} else if config.PortProtocol != nil {
		return j.GetJobsForNamedPortProtocol(resources, config.PortProtocol.Port, config.PortProtocol.Protocol, config)
	} else {
		panic(errors.Errorf("invalid ProbeConfig %+v", config))
	}
}

func (j *JobBuilder) GetJobsForNamedPortProtocol(resources *Resources, port intstr.IntOrString, protocol v1.Protocol, config *generator.ProbeConfig) *Jobs {
	jobs := &Jobs{}
	for _, podFrom := range resources.Pods {
		for _, podTo := range resources.Pods {
			job := &Job{
				FromKey:             podFrom.PodString().String(),
				FromNamespace:       podFrom.Namespace,
				FromNamespaceLabels: resources.Namespaces[podFrom.Namespace],
				FromPod:             podFrom.Name,
				FromPodLabels:       podFrom.Labels,
				FromContainer:       podFrom.Containers[0].Name,
				FromIP:              podFrom.IP,
				ToKey:               podTo.PodString().String(),
				ToHost:              podTo.Host(config),
				ToNamespace:         podTo.Namespace,
				ToNamespaceLabels:   resources.Namespaces[podTo.Namespace],
				ToPodLabels:         podTo.Labels,
				ToIP:                podTo.IP,
				DestinationPort:     -1,
				ResolvedPort:        -1,
				ResolvedPortName:    "",
				Protocol:            protocol,
				TimeoutSeconds:      j.TimeoutSeconds,
			}

			switch port.Type {
			case intstr.String:
				job.ResolvedPortName = port.StrVal
				// TODO what about protocol?
				portInt, err := podTo.ResolveNamedPort(port.StrVal)
				if err != nil {
					jobs.BadNamedPort = append(jobs.BadNamedPort, job)
					continue
				}
				job.ResolvedPort = portInt
				job.DestinationPort = portInt
			case intstr.Int:
				job.ResolvedPort = int(port.IntVal)
				job.DestinationPort = int(port.IntVal)
				// TODO what about protocol?
				portName, err := podTo.ResolveNumberedPort(int(port.IntVal))
				if err != nil {
					jobs.BadPortProtocol = append(jobs.BadPortProtocol, job)
					continue
				}
				job.ResolvedPortName = portName
			default:
				panic(errors.Errorf("invalid IntOrString value %+v", port))
			}

			if config.Service == generator.NodePortCluster || config.Service == generator.NodePortLocal {
				job.ResolvedPort = podTo.NodePort(config.Service, job.DestinationPort)
			}

			if config.Service == generator.LoadBalancerLocal && podFrom.LocalNodeIP != podTo.LocalNodeIP {
				jobs.Ignored = append(jobs.Ignored, job)
				continue
			}

			jobs.Valid = append(jobs.Valid, job)
		}
	}
	return jobs
}

func (j *JobBuilder) GetJobsAllAvailableServers(resources *Resources, config *generator.ProbeConfig) *Jobs {
	jobs := &Jobs{}
	for _, podFrom := range resources.Pods {
		for _, podTo := range resources.Pods {
			for _, contTo := range podTo.Containers {
				job := &Job{
					FromKey:             podFrom.PodString().String(),
					FromNamespace:       podFrom.Namespace,
					FromNamespaceLabels: resources.Namespaces[podFrom.Namespace],
					FromPod:             podFrom.Name,
					FromPodLabels:       podFrom.Labels,
					FromContainer:       podFrom.Containers[0].Name,
					FromIP:              podFrom.IP,
					ToKey:               podTo.PodString().String(),
					ToHost:              podTo.Host(config),
					ToNamespace:         podTo.Namespace,
					ToNamespaceLabels:   resources.Namespaces[podTo.Namespace],
					ToPodLabels:         podTo.Labels,
					ToContainer:         contTo.Name,
					ToIP:                podTo.IP,
					DestinationPort:     contTo.Port,
					ResolvedPort:        contTo.Port,
					ResolvedPortName:    contTo.PortName,
					Protocol:            contTo.Protocol,
					TimeoutSeconds:      j.TimeoutSeconds,
				}

				if config.Service == generator.NodePortCluster || config.Service == generator.NodePortLocal {
					job.DestinationPort = podTo.NodePort(config.Service, job.DestinationPort)
				}

				if config.Service == generator.LoadBalancerLocal && podFrom.LocalNodeIP != podTo.LocalNodeIP {
					jobs.Ignored = append(jobs.Ignored, job)
					continue
				}

				jobs.Valid = append(jobs.Valid, job)
			}
		}
	}

	return jobs
}
