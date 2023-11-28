Adapted with :blue_heart: from the [gateway api project GEP documentation](https://gateway-api.sigs.k8s.io/)

# Network Policy Enhancement Proposal (NPEP)

Network Policy Enhancement Proposals (NPEP) serve a similar purpose to the [KEP][kep]
process for the main Kubernetes project:

1. Ensure that changes to the API follow a known process and discussion
   in the OSS community.
2. Make changes and proposals discoverable (current and future).
3. Document design ideas, tradeoffs, decisions that were made for
   historical reference.

## NPEP Status

Every NPEP has a `Status` field, that is updated with the progress of NPEP.
Current statuses are:

* **Provisional:** The goals described by this NPEP have consensus but
  implementation details have not been agreed to yet.
* **Implementable:** The goals and implementation details described by this NPEP
  have consensus but have not been fully implemented yet.
* **Experimental:** This NPEP has been implemented and is part of the
  "Experimental" release channel. Breaking changes are still possible, up to
  and including complete removal and moving to `Rejected`.
* **Standard:** This NPEP has been implemented and is part of the
  "Standard" release channel. It should be quite stable.
* **Deferred:** We do not currently have bandwidth to handle this NPEP, it
  may be revisited in the future.
* **Rejected:** This proposal was considered by the community but ultimately
  rejected.
* **Replaced:** This proposal was considered by the community but ultimately
  replaced by a newer proposal.
* **Withdrawn:** This proposal was considered by the community but ultimately
  withdrawn by the author.

## Process

### 1. Discuss with the community

Before creating a NPEP, share your high level idea with the community. Check
[Community, discussion, contribution, and support](/#community-discussion-contribution-and-support)
for the suitable communication option.

### 2. Create an Issue

[Create a NPEP issue](https://github.com/kubernetes-sigs/network-policy-api/issues/new?assignees=&labels=kind%2Fenhancement&projects=&template=enhancement-proposal.md&title=%5BENHANCEMENT%5D)
in the repo describing your change.
At this point, you should copy the outcome of any other conversations or documents
into this Issue.

### 3. Create a first PR for your NPEP

NPEP process is supposed to be iterative, adding more details with every iteration.
To start NPEP process, create a PR adding `npep-<issue number>.md` file in the
[npep folder](https://github.com/kubernetes-sigs/network-policy-api/tree/master/npep)
using the [template NPEP](https://github.com/kubernetes-sigs/network-policy-api/blob/master/npep/npep-95.md) as a
starting point. Make sure to also add your new NPEP to the website, this can be done within
the `index.md` file at the root of the repo.

### 4. Provisional: Agree on the Goals and applicable User-Stories

Although it can be tempting to start writing out all the details of your
proposal, it's important to first ensure we all agree on the goals. The first
version of your NPEP should have "Provisional" status and leave out any implementation details,
focusing primarily on Goals, Non-Goals and User Stories.

### 5. Implementable: Document Implementation Details

Now that everyone agrees on the goals, it is time to start writing out your
proposed implementation details. These implementation details should be very
thorough, including the proposed API spec, and covering any relevant edge cases.
Note that it may be helpful to use a shared doc for part of this phase to enable
faster iteration on potential designs.

It is likely that throughout this process, you will discuss a variety of
alternatives. Be sure to document all of these in the NPEP, and why we decided
against them. At this stage, the NPEP should be targeting the "Implementable" stage.

### 6. Experimental: Mark API changes as "Experimental"

With the NPEP marked as "Implementable", it is time to actually make the proposed changes in our API.
In some cases, these changes will be documentation
only, but in most cases, some API changes will also be required.

It is important that every new feature of the API is marked as "Experimental" when it is introduced. Within the API, we
use <network-policy-api:experimental> tags to denote experimental fields. Within Golang packages (conformance tests,
CLIs, e.t.c.) we use the experimental Golang build tag to denote experimental functionality.

Some other requirements must be met before marking a NPEP Experimental:

* the graduation criteria to reach Standard MUST be filled out
a proposed probationary period (see next section) must be included in the NPEP and approved by maintainers.

* When updating NPEP from "Implementable" to "Experimental", status update may be the
only required change, but also feel free to link any API, docs, etc. changes that
were made as an implementation for a given NPEP.

* Before changes are released they MUST be documented. NPEPs that have not been
both implemented and documented before a release cut off will be excluded from
the release.

#### Probationary Period

Any NPEP in the `Experimental` phase is automatically under a "probationary
period" where it will come up for re-assessment if its graduation criteria are
not met within a given time period. NPEP that wish to move into `Experimental`
status MUST document a proposed period (6 months is the suggested default) that
MUST be approved by maintainers. Maintainers MAY select an alternative time
duration for a probationary period if deemed appropriate, and will document
their reasoning.

> **Rationale**: This probationary period exists to avoid NPEP getting "stale"
> and to provide guidance to implementations about how relevant features should
> be used, given that they are not guaranteed to become supported.

At the end of a probationary period if the NPEP has not been able to resolve
its graduation criteria it will move to "Rejected" status. In extenuating
circumstances an extension of that period may be accepted by approval from
maintainers. NPEP which are `Rejected` in this way are removed from the
experimental CRDs and more or less put on hold. NPEP may be allowed to move back
into `Experimental` status from `Rejected` for another probationary period if a
new strategy for achieving their graduation criteria can be established. Any
such plan to take a NPEP "off the shelf" must be reviewed and accepted by the
maintainers.

> **Warning**: It is extremely important** that projects which implement
> `Experimental` features clearly document that these features may be removed in
> future releases.

### 7. Graduate the NPEP to "Standard"

Once this feature has met the [graduation criteria](/versioning/#graduation-criteria), it is
time to graduate it to the "Standard" channel of the API. Depending on the feature, this may include
any of the following:

1. Graduating the resource to beta
2. Graduating fields to "standard" by removing `<network-policy-api:experimental>` tags
3. Graduating a concept to "standard" by updating documentation

### 8. Close out the NPEP issue

The NPEP issue should only be closed when the work is "done" (whatever
that means for that NPEP).

## Out of scope

What is out of scope: see [text from KEP][kep-when-to-use]. Examples:

* Bug fixes
* Small changes (API validation, documentation, fixups). It is always
  possible that the reviewers will determine a "small" change ends up
  requiring a NPEP.

[kep]: https://github.com/kubernetes/enhancements
[kep-when-to-use]: https://github.com/kubernetes/enhancements/tree/master/keps#do-i-have-to-use-the-kep-process
