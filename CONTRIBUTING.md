# Contributing Guidelines

Welcome to Kubernetes. We are excited about the prospect of you joining our [community](https://git.k8s.io/community)! The Kubernetes community abides by the CNCF [code of conduct](code-of-conduct.md). Here is an excerpt:

_As contributors and maintainers of this project, and in the interest of fostering an open and welcoming community, we pledge to respect all people who contribute through reporting issues, posting feature requests, updating documentation, submitting pull requests or patches, and other activities._

## Getting Started

We have full documentation on how to get started contributing here:

<!---
If your repo has certain guidelines for contribution, put them here ahead of the general k8s resources
-->

- [Contributor License Agreement](https://git.k8s.io/community/CLA.md) Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests
- [Kubernetes Contributor Guide](https://git.k8s.io/community/contributors/guide) - Main contributor documentation, or you can just jump directly to the [contributing section](https://git.k8s.io/community/contributors/guide#contributing)
- [Contributor Cheat Sheet](https://git.k8s.io/community/contributors/guide/contributor-cheatsheet) - Common resources for existing developers

### Installing the Admin Network Policy CRD

1) Clone the repo: `git clone https://github.com/kubernetes-sigs/network-policy-api.git`
2) Run `cd network-policy-api` && `make install`

## Developing the Website

The site documentation is written in Markdown and compiled with
[mkdocs](https://www.mkdocs.org/).

### Setting up local development

1.  Install mkdocs and required plugins:

    ```
    pip install mkdocs mkdocs-material mkdocs-awesome-pages-plugin mkdocs-macros-plugin mike
    ```

2.  To build the docs, run `make docs`
3.  To deploy the docs locally, run `make local-docs`

## Mentorship

- [Mentoring Initiatives](https://git.k8s.io/community/mentoring) - We have a diverse set of mentorship programs available that are always looking for volunteers!

## Contact Information

- [Slack](https://kubernetes.slack.com/messages/sig-network)
- [Mailing List](https://groups.google.com/forum/#!forum/kubernetes-sig-network)
