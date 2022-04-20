# KnFlow

This project contains both:
- a translator from [Serverless Workflow](https://serverlessworkflow.io)
programs to [Knative](https://knative.dev) Service, Broker and Triggers.
- a [minimal runtime](./functions/sw/README.md) 

> Disclaimer: this is just a small prototype to help identifying 
> what's missing in Knative Eventing to support Serverless Workflow 
> implementations built on top of Knative. 

## Basic Usage

To convert a Serverless Workflow program to Knative you can use the following 
command:

```shell
go run ./cmd/sw2kn <workflow.yaml>
```

This command generates a list of Knative objects as YAML. 

You can directly deploy the generated objects by running this command:

```shell
go run ./cmd/sw2kn <workflow.yaml> | kubectl apply -f -
```

You can then create a new workflow instance for invoking the deployed 
Knative service:

```shell
curl $(kubectl get ksvc helloworld -ojsonpath="{@.status.url}") 
workflow instance helloworld-xvlbzgba created.
```

Look at the Knative service logs to observe incoming events and state transitions.

Where `helloworld` corresponds to the workflow id.

## Documentation

[Translation Rules](https://hackmd.io/Ahvkxj2yRu262O-DaJUkNQ)

