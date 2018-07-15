package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "./Models"
    "github.com/gorilla/mux"
    "github.com/cskr/pubsub"
)

const timeoutTime int = 30
var passMessagesInternally = true
var messageCount int64
var pubSub = pubsub.New(0);

func PostMessage(resp http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    topic := params["topic"]
    var message Models.Message
    if err := json.NewDecoder(req.Body).Decode(&message); err != nil {
        fmt.Println(err)
    }
    messageCount++
    data := Models.Message{Message: message.Message, Event: "msg", Id: messageCount}
    pubSub.TryPub(data, topic)

    resp.WriteHeader(http.StatusNoContent)
}

func GetMessages(resp http.ResponseWriter, req *http.Request) {
    params := mux.Vars(req)
    topic := params["topic"]
    resp.Header().Set("Content-Type", "text/event-stream")
    resp.Header().Set("Cache-Control", "no-cache")
    resp.Header().Set("Connection", "keep-alive")
    resp.Header().Set("Access-Control-Allow-Origin", "*")

    flusher, ok := resp.(http.Flusher)

    if !ok {
        http.Error(resp, "Streaming unsupported!", http.StatusInternalServerError)
        return
    }

    internalMessages := pubSub.Sub(topic)

    timeout := make(chan bool)

    go func() {
        for {
            select {
            case internalMessage := <-internalMessages:
                msg := internalMessage.(Models.Message)
                fmt.Fprintf(resp, "id: %d\n", msg.Id)
                fmt.Fprintf(resp, "event: %s\n", msg.Event)
                fmt.Fprintf(resp, "data: %s\n\n", msg.Message)
                flusher.Flush()
            case <-time.After(time.Duration(timeoutTime) * time.Second):
                fmt.Fprintf(resp, "event: %s\n", "timeout")
                fmt.Fprintf(resp, "data: %d sec\n\n", timeoutTime)
                flusher.Flush()
                timeout <- true
            }
        }
    }()

    <-timeout

    resp.WriteHeader(http.StatusOK)
}

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
        panic(fmt.Sprintf("%s: %s", msg, err))
    }
}

func main() {
    handleRoutes()
}

func handleRoutes() {
    router := mux.NewRouter()
    router.HandleFunc("/infocenter/{topic}", GetMessages).Methods("GET")
    router.HandleFunc("/infocenter/{topic}", PostMessage).Methods("POST")
    if err := http.ListenAndServe(":3002", router); err != nil {
        log.Fatal(err)
    }
}
