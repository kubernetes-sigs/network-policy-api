### Summary

I want to allow traffic from one set pod(s) to another set of pod(s) 
in one intent, so that I do not have to link two intents (ingress and egress) 
to understand allowed traffic between two sets of pods.

### Description

NP v1 allowed traffic is defined in a pair of Egress and Ingress intent, 
that forces namespace admin to create two policies, one with source pods 
as pod selector and another with destination pods as pod selector and 
each policy defines one side of the traffic.

This approach has following limitations:
1. More work of namespace admin. Admin has to create two NPs one ingress and other for egress.
2. To understand allowed traffic, admin has to look into multiple policies 
and correlate Egress and Ingress.
3. Possibility of misconfiguration, if only, one half of the traffic is configured.

In NP v2, we can solve this problem by defining a single intent with both 
sets of pods (from and to) in single intent solving all above mentioned problems.

### Acceptance Criteria

If I configure single intent with "from" and "to", All pods in "from" will be 
able to initiate traffic to all pods in "to" as per the intent (ports, protocols).

