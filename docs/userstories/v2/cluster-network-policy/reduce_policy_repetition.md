### Summary

I would like to reduce policy repetition due to different environments. 

### Description

In certain cases same policies are repeated for different environments. 
For example testing policy would be repeated for production policy with 
different envionment label. The reason for repeation is - not to 
allow cross communication between testing and production workloads.
 
This may not be for one dimension, 
it could be for multiple dimensions (multiple labels). 
Example cross site or deployment communication is not allowed.

### Acceptance Criteria

Cross communication between different environments should be stopped.

Configure CNP v2 with same policy without repeating for multiple enviroments, 
but there should be any cross communication between multiple enviroments 
(production traffic shouldn't be communicating with testing or developemt traffic)

### Notes

For example an application might have different dimensions like deployment and site.
Deployment dimension has few options like ‘development’, ‘testing’, ‘pre production’, 
and ‘production’. Site dimension has many options like ‘paris’, ‘uk’, ‘sunnyvale’, 
and etc.., All of your application communication should belong to its 
deployment option and there shouldn’t be any cross communication. 
Similar restrictions may be there for the site dimension also. 

Due to the above restrictions, the number of policies for the same application 
needs to be configured by choosing that dimension in the selector and also 
during rule specification.

Same concept is added for NP v2.
