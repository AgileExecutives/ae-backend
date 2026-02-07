
# Client Remaining Units feature
please investigate in the code and create and implementation plan

a client - if his therapy is payed by cost-provider - has a number of approved units. a unit usually has 60min depending on the cost provider (add a field minutes_per_unit default 60 to cost_provider with default 60).

the client has a number of already conducted / scheduled minutes of sessions plus a number of extra efforts that are billable.

keep the maximum number of units the client can schedule in the client object.

the remaining units are calculated from the remaining minutes of the sessions and billable extra efforts.

the client module can subscribe to the session created and updated and deleted events as well to the extra effort created updated and deleted events and then call a service that counts the remaining minutes.

remaining units - those are the number of sessions we can schedule.


# Invoice Number Configured on and Unique in Organization 
