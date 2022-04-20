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
    "github.com/cloudevents/sdk-go/v2/binding"
    "github.com/cloudevents/sdk-go/v2/event"
    cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

type States map[string]State

var (
    states     States
    sw         SW
    operations map[string]func(state State, data *event.Event) *event.Event
    brokerURL  string
)

func init() {
    err := json.Unmarshal([]byte(os.Getenv("STATES")), &sw)
    if err != nil {
        log.Fatalf("Invalid serverless workflow program: %v", err)
    }

    // Index by state name
    states = make(States)
    for _, state := range sw.States {
        states[MakeK8sName(state.Name)] = state
    }

    // Register operations
    operations = make(map[string]func(state State, data *event.Event) *event.Event)
    operations["inject"] = Inject

    brokerURL = os.Getenv("BROKER")
}

// Handle an HTTP Request.
func Handle(ctx context.Context, res http.ResponseWriter, req *http.Request) {
    if !req.URL.Query().Has("state") {
        createNewInstance(ctx, res, req)
        return
    }

    message := cehttp.NewMessageFromHttpRequest(req)
    defer message.Finish(nil)

    event, err := binding.ToEvent(ctx, message)
    if err != nil {
        log.Printf("HTTP body is not a valid CloudEvents: %v", err)
        http.Error(res, "HTTP body is not a valid CloudEvents", http.StatusBadRequest)
        return
    }

    stateName := req.URL.Query().Get("state")
    state, ok := states[stateName]
    if !ok {
        http.Error(res, "State not found", http.StatusNotFound)
        return
    }

    log.Printf("entering %s state\n", stateName)
    log.Printf("  data: %q\n", string(event.Data()))

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

    log.Printf("exiting %s state.\n", stateName)
    log.Printf("  data: %q\n", string(event.Data()))

    res.Write(bytes)
}

func createNewInstance(ctx context.Context, res http.ResponseWriter, req *http.Request) {
    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(res, "Invalid HTTP body", http.StatusBadRequest)
        return
    }
    req.Body.Close()

    var data interface{}
    data = map[string]string{}
    if len(body) > 0 {
        initialdata, err := json.Marshal(body)
        if err != nil {
            log.Printf("HTTP body is not valid JSON: %v", err)
            http.Error(res, "HTTP body is not valid JSON", http.StatusBadRequest)
            return
        }
        data = initialdata
    }

    knflowinstanceid := sw.ID + "-" + RandomString()

    event := cloudevents.NewEvent()

    event.SetID("0")
    event.SetSource("sw/" + sw.ID)
    event.SetType("sw.start")
    event.SetData(cloudevents.ApplicationJSON, data)

    event.SetExtension("knflowinstanceid", knflowinstanceid)
    event.SetExtension("knstatename", sw.Start)

    log.Println(event.String())

    // Publish event to broker
    c, err := cloudevents.NewClientHTTP()
    if err != nil {
        log.Fatalf("failed to create client, %v", err)
    }

    ctx = cloudevents.ContextWithTarget(ctx, brokerURL)

    // Send that Event.
    if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
        log.Printf("failed to send, %v", result)
        http.Error(res, "failed to start workflow", http.StatusInternalServerError)
        return
    }
    log.Printf("workflow instance %s created.\n", knflowinstanceid)
    response := fmt.Sprintf("workflow instance %s created.", knflowinstanceid)
    res.Header().Set("Content-Type", "text/plain")
    res.Write([]byte(response))
    return
}

func Inject(state State, event *event.Event) *event.Event {
    // todo: merge
    event.SetData("application/json", state.Data)
    return event
}
