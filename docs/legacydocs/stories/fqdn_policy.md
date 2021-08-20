# Cluster Scoped Policy

## Owners

- Gobind
- (others please feel free to add)

## User Story

As a Kubernetes user, I want to allow/deny egress to specific FQDNs using Network Policy. 
For instance, I would like to use Network Policy to express the following constraint:

{
  egress:
  - to:
    - FQDN:
        url: www.my-trusted-company.com
}

This policy would only permit the pods selected by this policy to send packets to IPs
that belong to www.my-trusted-company.com. It implicitly will deny any packets to other
websites from the selected pods.

You can image this would be handy even for services implemented inside the cluster as
well as other services outside the cluster but not necessarily on the internet.

Many existing products already offer this functionality today:
1. [Cilium FQDN based network policy](https://docs.cilium.io/en/v1.8/policy/language/#dns-based)
2. [Calico FQDN based network policy](https://docs.projectcalico.org/security/calico-enterprise/egress-access-controls)
3. [OpenShift egress firewall with FQDN](https://docs.openshift.com/container-platform/4.3/networking/openshift_sdn/configuring-egress-firewall.html#domain-name-server-resolution_configuring-an-egress-firewall)

(related: see https://github.com/kubernetes/kubernetes/issues/50453)

## SIG Network Proposal

https://docs.google.com/document/d/1Htcy4UXKZytUe-lWJIIEJZzoa3MtCMr-Ms_KONaXirM/edit#
