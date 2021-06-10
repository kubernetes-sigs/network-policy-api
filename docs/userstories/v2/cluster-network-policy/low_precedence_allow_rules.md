### Summary

As cluster admin, I would be able to allow certain traffic at lower precedence.

### Description

Cluster admin can allow certain traffic at lower precedence, 
whereas namespace admin can override with deny rule NP v2.

### Acceptance Criteria

CNP v2 post section allows traffic, and NP v2 deny the same traffic - Finally traffic should be denied.

CNP v2 post section allows traffic, and NP v2 also allow the same traffic - Finally traffic should be allowed.

CNP v2 post section allows traffic, and NP v2 doesn't have any rules for this traffic - Finally traffic should be allowed.

