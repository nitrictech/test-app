
entities to test
================
functions
collections
queues
schedules
topics
buckets - TODO
secrets - TODO

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

> topic

post to /send/ {messageType=topic}
> should send to topic and cause worker function to run and receive message
> all is recorded in the history so get history and assert message sent and received

test 3
------

> queue

post to /actions/ {messageType=queue}
> should send to queue and scheduled worker should receive message
> all is recorded in the history so get history and assert message sent and received


How to run
==========

Clone this repo and cd into it.

Then run the app, with any of these commands:
- nitric run
- nitric deployment apply -t <aws,azure,gcp>

If you deployed, then get the api url output from the above command:

```
$ export BASE_URL=<from above>
$ make test
```