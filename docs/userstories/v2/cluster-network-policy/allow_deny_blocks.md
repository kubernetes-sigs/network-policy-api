### Summary

As cluster admin, I should be able to accept both allow and deny rules 
for a given cluster policy.

### Description

Cluster admin should have an option to add allow or deny rules. 
User should be able to specify in allow and deny blocks.

### Acceptance Criteria

Create deny rules at cluster level - traffic should be denied
Create allow rules at cluster level - traffic should be allowed, 
if there are no deny rules for this traffic.

### Notes

Typically firewalls have sequence numbers for evaluation of rules 
(could be allow or deny rules), but it would be easier to create two blocks, 
those are deny block rules and allow block rules. 
Deny block rule will have higher precedence.


