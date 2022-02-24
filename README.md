
entities to test
================
functions
collections
queues
schedules
topics
buckets
secrets

test 0
------

All resources are deployed to the destination

test 1
------

> api, functions, collection

list store/
post some items to store
get an item
list store/
delete an item

test 2
------

> schedules, functions

there is a schedule so go and read the history and look for evidence of them running


test 3
------

> topic

post to /send/ {messageType=topic}
> should send to topic and cause worker function to run and receive message
> all is recorded in the history so get history and assert message sent and received


test 4
------

> queue, bucket

post to /actions/ {messageType=queue}
> should send to queue and scheduled worker should receive message and update bucket
> all is recorded in the history so get history and assert message sent and received
