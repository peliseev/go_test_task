# Queue in Go with REST interface
## Technical Specification (TS)

Implement a queue broker as a web service. The service should handle 2 methods:

1. PUT /queue?v=message

Put the message "message" into the queue with the name "queue" (the queue name can be any), for example:

* curl -XPUT http://127.0.0.1/pet?v=cat
* curl -XPUT http://127.0.0.1/pet?v=dog
* curl -XPUT http://127.0.0.1/role?v=manager
* curl -XPUT http://127.0.0.1/role?v=executive

In response {empty body + status 200 (ok)}

If the v parameter is missing - empty body + status 400 (bad request)

2. GET /queue

Take (in a FIFO manner) a message from the queue named "queue" and return it in the body of the http request, for example (the result that should be obtained with the above put's):

* curl http://127.0.0.1/pet => cat
* curl http://127.0.0.1/pet => dog
* curl http://127.0.0.1/pet => {empty body + status 404 (not found)}
* curl http://127.0.0.1/pet => {empty body + status 404 (not found)}
* curl http://127.0.0.1/role => manager
* curl http://127.0.0.1/role => executive
* curl http://127.0.0.1/role => {empty body + status 404 (not found)}

When making GET requests, make it possible to set a timeout argument:

* curl http://127.0.0.1/pet?timeout=N

If there is no ready message in the queue, the recipient must wait either until the message arrives or until the timeout expires (N - number of seconds). In case the message does not appear - return the code 404. Receivers must receive messages in the same order as the requests were received from them, if 2 receivers are waiting for messages (using timeout), the first message should be received by the one who requested it first.

The port on which the service will listen should be specified in the command line arguments.