# Implementations

This document tracks downstream(actual and planned) implementations of
Network Policy API resources and provides status and resource references for them.

Implementors of Network Policy API resources are encouraged to update this document with status information about their
implementations, the versions they cover, and documentation to help users get started.

## ClusterNetworkPolicy

Updated: 21-April-2026

- [Kube-network-policies](https://github.com/kubernetes-sigs/kube-network-policies)
- [Kube-OVN CNI](https://github.com/kubeovn/kube-ovn) (Experimental support)
- [Calico CNI](https://github.com/projectcalico/calico) (Standard support in Calico v3.32)

## AdminNetworkPolicy and BaseLineAdminNetworkPolicy (v1alpha1)

Updated: 21-April-2026

- [Kube-network-policies](https://github.com/kubernetes-sigs/kube-network-policies/tree/v0.8.1)
- [OVN-Kubernetes CNI](https://github.com/ovn-org/ovn-kubernetes/) (Has implemented standard fields of the API + Nodes/Networks in Experimental)
- [Antrea CNI](https://github.com/antrea-io/antrea/) (Has implemented standard fields of the API)
- [Kube-OVN CNI](https://github.com/kubeovn/kube-ovn) (Has implemented standard fields of the API)
- [Calico CNI](https://github.com/projectcalico/calico) (Has implemented standard fields of the API + Networks peer in experimental in Calico v3.29, v3.30, and v3.31)
- [Cilium CNI](https://github.com/cilium/cilium/issues/23380) (tracking issue)
