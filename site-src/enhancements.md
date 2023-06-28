# Network Policy Enhancement Proposal (NPEP)

Network Policy Enhancement Proposals (NPEP) serve a similar purpose to the [KEP][kep]
process for the main Kubernetes project:

1. Ensure that changes to the API follow a known process and discussion
   in the OSS community.
2. Make changes and proposals discoverable (current and future).
3. Document design ideas, tradeoffs, decisions that were made for
   historical reference.

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
Every NPEP has a `Status` field, that is updated with the progress of NPEP.
Current statuses are:

* "Provisional"
* "Implementable"
* "Standard"

they are described in more details further in this section.
To start NPEP process, create a PR adding `npep-<issue number>.md` file in the 
[npep folder](https://github.com/kubernetes-sigs/network-policy-api/tree/master/npep)
using the [template NPEP](https://github.com/kubernetes-sigs/network-policy-api/blob/master/npep/npep-95.md) as a starting point. 

### 4. Provisional: Agree on the Goals
Although it can be tempting to start writing out all the details of your
proposal, it's important to first ensure we all agree on the goals. The first
version of your NPEP should have "Provisional" status and leave out any implementation details, 
focusing primarily on "Goals" and "Non-Goals".

### 5. Implementable: Document Implementation Details
Now that everyone agrees on the goals, it is time to start writing out your
proposed implementation details. These implementation details should be very
thorough, including the proposed API spec, and covering any relevant edge cases.
Note that it may be helpful to use a shared doc for part of this phase to enable
faster iteration on potential designs.

It is likely that throughout this process, you will discuss a variety of
alternatives. Be sure to document all of these in the NPEP, and why we decided
against them. At this stage, the NPEP should be targeting the "Implementable" stage.

### 6. Standard: Make API changes
With the NPEP marked as "Implementable", it is time to actually make the proposed changes in our API. 
In some cases, these changes will be documentation
only, but in most cases, some API changes will also be required.

When updating NPEP from "Implementable" to "Standard", status update may be the 
only required change, but also feel free to link any API, docs, etc. changes that
were made as an implementation for a given NPEP.

Before changes are released they MUST be documented. NPEPs that have not been
both implemented and documented before a release cut off will be excluded from
the release.


### 7. Close out the NPEP issue
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