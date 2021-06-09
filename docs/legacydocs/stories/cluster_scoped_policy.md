# Cluster Scoped Policy

## Owners

- Gobind
- Abhishek Raut
- Yang

## User Story

A a platform operator or cluster administrator, I need to create policies that
impact pods across multiple (or all) namespaces. As an example, these
"cluster scoped" policies could allow access to and from platform shared
services. I would like to specify the binding of policy to pods by composing both
pod and namespace selectors (and other selectors).

* Clarify that selectors could span namespaces.
* Ability to set defaults for clusters (specific use case)

Separate user story for ordering policies (actions supported as well as
delegation)

Decide "scope" of story

