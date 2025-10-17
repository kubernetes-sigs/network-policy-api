# Network Policy API Working Group
 👋 Welcome to the Network Policy API Project - we are happy to have you here! Before you get started here are some useful links:

- [Bi-Weekly Meeting Agenda](https://docs.google.com/document/d/1AtWQy2fNa4qXRag9cCp5_HsefD7bxKe3ea2RPn8jnSs/edit#heading=h.ajvcztp6cza)
- [NetworkPolicy v1 Docs](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [API reference spec](/reference/spec/)

## What is the Network Policy API Project?

The network policy API subgroup is a part of [SIG Network](https://github.com/kubernetes/community/tree/master/sig-network),
formed to address further work involving Kubernetes network security beyond the core NetworkPolicy v1 resource.
The [`network-policy-api`](https://github.com/kubernetes-sigs/network-policy-api/) repository contains APIs which are
created and maintained by this subgroup.

## APIs

We refer to "APIs" in the plural here because this project will house
multiple K8s CRD resources geared towards different users and use cases.

### Active (APIs undergoing active development)

- [ClusterNetworkPolicy](api-overview.md#the-clusternetworkpolicy-resource)

### Previous (APIs that have evolved to the next stage and are no longer actively developed)

- [AdminNetworkPolicy](https://network-policy-api.sigs.k8s.io/reference/spec/#policy.networking.k8s.io%2fv1alpha1.AdminNetworkPolicy) and 
[BaselineAdminNetworkPolicy](https://network-policy-api.sigs.k8s.io/reference/spec/#policy.networking.k8s.io%2fv1alpha1.BaselineAdminNetworkPolicy)

### Future (Possible APIs to be created in the future)

- DeveloperNetworkPolicy

## Outreach

- [Kubecon NA 2022 Contributors Summit](https://youtu.be/00nVssi2oPA)
- [Kubecon NA 2022 SIG Network Deep Dive](https://www.youtube.com/watch?v=qn9bM5Cwvg0&t=752s)
- [Kubecon EU 2023 SIG Network Deep Dive](https://www.youtube.com/watch?v=0uPEFcWn-_o)
- [Kubecon EU 2025 SIG Network Intro and Updates](https://www.youtube.com/watch?v=lBOdQHNNgEU)

## Community, discussion, contribution, and support
Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-network-policy-api)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-network)

The Network Policy API Meeting happens bi-weekly on Tuesday at 9am Pacific
Time (16:00 UTC):

* [Zoom link](https://zoom.us/j/96264742248)
* [Meeting Agenda](https://docs.google.com/document/d/1AtWQy2fNa4qXRag9cCp5_HsefD7bxKe3ea2RPn8jnSs/edit#heading=h.ajvcztp6cza)

To get started contributing please take time to read over the [contributing guidelines](https://github.com/kubernetes-sigs/network-policy-api/blob/master/CONTRIBUTING.md) as well as the [developer guide](https://github.com/kubernetes/community/blob/master/contributors/devel/README.md). You can then take a look at the open issues labelled 'good-first-issue' [here](https://github.com/kubernetes-sigs/network-policy-api/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

### Code of conduct
Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](https://github.com/kubernetes/community/blob/master/code-of-conduct.md).
