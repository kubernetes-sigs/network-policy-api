# Network Policy API User Stories

This directory should be the entry point for any new object or feature proposed 
to the sig network-policy-api subgroup 

It is meant to house all user stories that have been reviewed and agreed upon by
the sig for any new k8's networking security object. The model should follow a 
standard github workflow, with a directory representing proposed new objects 
which contains individual user stories that can be reviewed and commented on. 

All work proposed and merged here should follow the [style guidelines](https://github.com/kubernetes/community/blob/master/contributors/guide/style-guide.md) 
agreed upon by the Kubernetes community

## Directory Structure 

Each new Object should have it's own directory, following directory naming 
conventions seen in [`kubernetes/kubernetes`.](https://github.com/kubernetes/kubernetes) Specifically the directory name 
should be **lowercase** and **descriptive**. The directory structure should 
roughly represent weather or not the new object will augment the existing [network policy API](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#networkpolicy-v1-networking-k8s-io) (`network-policy-api/docs/userstories/v1/<newobject>/`) or create a 
new set of APIs (`network-policy-api/docs/userstories/v2/<newobject>/`). 
If there is some doubt as to which bucket the new object will fall into, the new 
directory should be made outside of either subdirectory (`network-policy-api/docs/userstories/<newobject>/`) and a final 
determination can be made at a later date.

## User Story Structure

To submit a new user story please 
open a PR roughly following the [template](#template) shown below. Each story 
should be in it's own markdown file titled with a descriptive 3 word **lowercase**
title:  `<word1>_<word2>_<word3>.md, ...`

The following template should be used as a suggested guideline, please feel free
to deviate and augment as needed. 

### Template

```markdown

### Summary 

A user story should typically have a summary structured this way:

1. **As a** [user concerned by the story]
2. **I want** [goal of the story]
3. **so that** [reason for the story]

The “so that” part is optional if more details are provided in the description. 

### Description 

The user story should have a reason to exist: what do I need as the user 
described in the summary? This part details any detail that could not be passed 
by the summary.

### Acceptance Criteria

1. [If I do A.]
2. [B should happen.]

[
Also, here are a few points that need to be addressed:

1. Constraint 1;
2. Constraint 2;
3. Constraint 3.
]

### Resources:

* Relavant previous discussions on the story if there are any 
  [i.e link to new story PR]
* Mockups: [Here goes a URL to or the name of the mockup(s) in inVision];


### Notes

[Some complementary notes if necessary:]

* > Here goes a quote from an email
* Here goes whatever useful information can exist…
```

## Wrapping up

Once all or most of the user stories are agreed upon and merged an `INDEX.md` 
file should be created in the `<newobject>` directory. It will serve to organize 
all the user stories, link to individual PR discussions, and provide any 
overarching assumptions for the new object and it's stories. 

### Resources:

* [Style-guides and template for a user story](https://github.com/AlphaFounders/style-guide/blob/master/agile-user-story.md)

[1]: https://github.com/AlphaFounders/style-guide