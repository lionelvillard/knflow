package function

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"

    cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Machine map[string]State

var (
    machine    Machine
    sw         SW
    operations map[string]func(state State, data cloudevents.Event) cloudevents.Event
    brokerURL  string
)

func init() {
    err := json.Unmarshal([]byte(os.Getenv("MACHINE")), &sw)
    if err != nil {
        log.Fatalf("Invalid serverless workflow: %v", err)
    }

    // Index by states
    machine = make(Machine)
    for _, state := range sw.States {
        machine[MakeK8sName(state.Name)] = state
    }

    // Register operations
    operations = make(map[string]func(state State, data cloudevents.Event) cloudevents.Event)
    operations["inject"] = Inject

    brokerURL = os.Getenv("BROKER")
}

// Handle an HTTP Request.
func Handle(ctx context.Context, res http.ResponseWriter, req *http.Request) {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(res, "Invalid HTTP body", http.StatusBadRequest)
        return
    }
    req.Body.Close()

    // REVISIT: accepts non-cloudevents for initial state
    event := cloudevents.NewEvent() // Workflow data
    err = json.Unmarshal(body, &event)
    if err != nil {
        log.Printf("HTTP body is not a valid CloudEvents: %v", err)
        http.Error(res, "HTTP body is not a valid CloudEvents", http.StatusBadRequest)
        return
    }

    if !req.URL.Query().Has("state") {
        // Starting workflow instance
        knflowinstanceid := sw.ID+"-"+RandomString()

        event.SetExtension("knflowinstanceid", knflowinstanceid)
        event.SetExtension("knstatename", sw.Start)

        // Publish event to broker
        c, err := cloudevents.NewClientHTTP()
        if err != nil {
            log.Fatalf("failed to create client, %v", err)
        }

        ctx = cloudevents.ContextWithTarget(ctx, brokerURL)

        event.SetSource("sw")

        // Send that Event.
        if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
            log.Printf("failed to send, %v", result)
            http.Error(res, "failed to start workflow", http.StatusInternalServerError)
            return
        }

        response := fmt.Sprintf("workflow instance %s created.", knflowinstanceid)
        res.Header().Set("Content-Type", "text/plain")
        res.Write([]byte(response))
        return
    }

    stateName := req.URL.Query().Get("state")
    state, ok := machine[stateName]
    if !ok {
        http.Error(res, "State not found", http.StatusNotFound)
        return
    }

    op := operations[state.Type]
    if op == nil {
        log.Printf("Operation not supported: %v", err)
        http.Error(res, "Operation not supported", http.StatusNotImplemented)
        return
    }

    reply := op(state, event)

    bytes, err := json.Marshal(reply)
    if err != nil {
        log.Printf("Internal error: %v", err)
        http.Error(res, "internal error", http.StatusInternalServerError)
        return
    }

    // End state?
    if end, ok := state.End.(bool); ok && end {
        log.Printf("WORKFLOW %s ENDED", event.Extensions()["knflowinstanceid"])
        log.Println(string(reply.Data()))
        res.WriteHeader(200)
        return
    }

    res.Write(bytes)
}

func Inject(state State, event cloudevents.Event) cloudevents.Event {
    // todo: merge
    event.SetData("application/json", state.Data)
    return event
}
