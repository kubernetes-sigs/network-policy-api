# Implementations

This document tracks downstream(actual and planned) implementations of
Network Policy API resources and provides status and resource references for them.

Implementors of Network Policy API resources are encouraged to update this document with status information about their
implementations, the versions they cover, and documentation to help users get started.

## AdminNetworkPolicy and BaseLineAdminNetworkPolicy

Updated: 10-29-2024

- [Kube-network-policies](https://github.com/kubernetes-sigs/kube-network-policies)
- [OVN-Kubernetes CNI](https://github.com/ovn-org/ovn-kubernetes/) (Has implemented standard fields of the API + Nodes/Networks in Experimental)
- [Antrea CNI](https://github.com/antrea-io/antrea/) (Has implemented standard fields of the API)
- [KubeOVN CNI](https://github.com/kubeovn/kube-ovn) (Has implemented standard fields of the API)
- [Calico CNI](https://github.com/projectcalico/calico/issues/7578) (work in progress)
- [Cilium CNI](https://github.com/cilium/cilium/issues/23380) (tracking issue)
