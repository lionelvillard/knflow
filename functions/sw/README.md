# Serverless Workflow State Machine

This function provides a minimal runtime for Serverless Workflow programs.

## Running the function locally

First build the function:

```shell
kn func build
```

then run it:

```sh
kn func run -b -e STATES=<STATES> -e BROKER=<BROKER_URL>
```

where `MACHINE` is a compiled serverless workflow description and `BROKER`
the URL pointing to a Knative Broker.

## Example

```shell
kn func run -b -e STATES='{"id":"helloworld", "start":"hello", "states":[{"
name":"hello","type":"inject","data":"HELLO", "end":true}]}' -e BROKER="localhost:8081"
```

In another terminal, use this command for making the Knative Eventing Broker accessible from your local machine:

```shell
kubectl port-forward -n knative-eventing mt-broker-ingress-XXX 8081:8080
```

In another terminal, create a new workflow instance by invoking the function with no argument
and a (dummy) CloudEvent:

```shell
curl http://localhost:8080 -d '{"specversion":"1.0", "type":"atype", "id":"1"}'
workflow instance helloworld-xvlbzgba created.
```

Then manually trigger the next state:

```shell
curl http://localhost:8080?state=hello -d '{"specversion":"1.0", "type":"atype", "id":"1", "knflowinstanceid":"helloworld-xvlbzgba"}'
```

In the function terminal, you should see:

```shell
WORKFLOW helloworld-xvlbzgba ENDED
"HELLO"
```



