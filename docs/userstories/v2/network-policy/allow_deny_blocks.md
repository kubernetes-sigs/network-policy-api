### Summary

As namespace admin, I should be able to accept both allow and deny rules 
for a given network policy

### Description

Namespace admin should have an option to add allow or deny rules.

Typically firewalls have sequence numbers for evaluation of rules 
(could be allow or deny rules), but it would be easier to create two sections, 
those are deny section rules and allow section rules. 
Deny section rule will have higher precedence.

### Acceptance Criteria

Create deny rules at network policy - traffic should be denied
Create allow rules at network policy - traffic should be allowed, 
if there are no deny rules for this traffic.

### Notes



