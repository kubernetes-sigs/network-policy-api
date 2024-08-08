# Dev Guide

## Project management

We are using the GitHub issues and project dashboard to manage the list of TODOs
for this project:

* [Open issues][gh-issues]
* [Project dashboard][gh-dashboard]

Issues labeled `good first issue` and `help wanted` are especially good for a
first contribution.

We use [priority labels][prio-labels] to help indicate the timing importance of
resolving an issue, or whether an issue needs more support from its creator or
the community to be prioritized.

[gh-issues]: https://github.com/kubernetes-sigs/network-policy-api/issues
[gh-dashboard]: https://github.com/orgs/kubernetes-sigs/projects/32
[prio-labels]: https://github.com/kubernetes-sigs/network-policy-api/labels?q=priority

## Prerequisites

Before you start developing with Network Policy API, we'd recommend having the
following prerequisites installed:

* [Go](https://golang.org/doc/install): Main programing language for this project.
* [Python3](https://www.python.org/downloads/): To build this documentation page.
* [Kind](https://kubernetes.io/docs/tasks/tools/#kind): To run Kubenetes on the local machine, and it also has a dependency on Docker or Podman.
* [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl): Kubenetes command line tool.


### Building, testing and deploying

Clone the repo:

```
mkdir -p $GOPATH/src/sigs.k8s.io
cd $GOPATH/src/sigs.k8s.io
git clone https://github.com/kubernetes-sigs/network-policy-api.git
cd network-policy-api
```

This project works with Go modules; you can chose to setup your environment
outside $GOPATH as well.

### Building the code

The project uses `make` to drive the build. `make` will run code generators, and
run static analysis against the code and generate Kubernetes CRDs. You can kick
off an overall build from the top-level makefile:

```shell
make install
```

### Submitting a Pull Request

Network Policy API follows a similar pull request process as
[Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md).
Merging a pull request requires the following steps to be completed before the
pull request will be merged automatically.

- [Sign the CLA](https://git.k8s.io/community/CLA.md) (prerequisite)
- [Open a pull request](https://help.github.com/articles/about-pull-requests/)
- Pass [verification](#verify) tests
- Get all necessary approvals from reviewers and code owners

### Verify

Make sure you run the static analysis over the repo before submitting your
changes. The [Prow presubmit][prow-setup] will not let your change merge if
verification fails.

```shell
make verify
```

[prow-setup]: https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/network-policy-api

### Documentation

The site documentation is written in Markdown and compiled with
[mkdocs](https://www.mkdocs.org/). Each PR will automatically include a
[Netlify](https://netlify.com/) deploy preview. When new code merges, it will
automatically be deployed with Netlify to
[network-policy-api.sigs.k8s.io](https://network-policy-api.sigs.k8s.io). If you want to
manually preview docs changes locally, you can install mkdocs and run:

```shell
 make docs
```

You might want to install [Python Vitural environment](https://docs.python.org/3/library/venv.html) to avoid conflicts.

Install the required plugins.

```shell
pip install mkdocs mkdocs-material mkdocs-awesome-pages-plugin mkdocs-macros-plugin
```

Once the build is complete, there would be a new folder called `site` generated, and you can deploy the website locally at port 8000 and run:

```shell
  make local-docs
```
