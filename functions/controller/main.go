// [START snippet]

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asalkeld/test-app/common"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/api/events"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	history documents.CollectionRef
	queue   queues.Queue
	topic   resources.Topic
)

// Updates context with error information
func httpError(ctx *faas.HttpContext, message string, status int) {
	ctx.Response.Body = []byte(message)
	ctx.Response.Status = status
}

func historyGetHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	query := history.Query()
	results, err := query.Fetch()
	if err != nil {
		return nil, err
	}

	docs := make([]map[string]interface{}, 0)
	for _, doc := range results.Documents {
		docs = append(docs, doc.Content())
	}

	b, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	ctx.Response.Body = b
	ctx.Response.Headers["Content-Type"] = []string{"application/json"}

	return next(ctx)
}

func factDeleteHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params, ok := ctx.Extras["params"].(map[string]string)
	if !ok || params == nil {
		return nil, fmt.Errorf("error retrieving path params")
	}

	id := params["id"]

	err := history.Doc(id).Delete()
	if err != nil {
		ctx.Response.Body = []byte("Error deleting document " + id)
		ctx.Response.Status = 404
	} else {
		ctx.Response.Status = 204
	}

	return next(ctx)
}

func sendPostHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	m := &common.Message{}
	if err := json.Unmarshal(ctx.Request.Data(), m); err != nil {
		httpError(ctx, "error decoding json body", 400)
		return ctx, nil
	}

	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	mMap := make(map[string]interface{})
	err := mapstructure.Decode(m, &mMap)
	if err != nil {
		httpError(ctx, "error decoding message document", 400)
		return ctx, nil
	}

	switch strings.ToLower(m.MessageType) {
	case "topic":
		topic.Publish(&events.Event{
			ID:          m.ID,
			PayloadType: m.PayloadType,
			Payload:     mMap,
		})
	case "queue":
		queue.Send([]*queues.Task{
			{
				ID:          m.ID,
				PayloadType: m.PayloadType,
				Payload:     mMap,
			},
		})
	}

	ctx.Response.Status = 200
	ctx.Response.Body = []byte(fmt.Sprintf("Run action : %v", m))

	return next(ctx)
}

func main() {
	var err error
	history, err = resources.NewCollection("history", resources.CollectionReading)
	if err != nil {
		panic(err)
	}

	queue, err = resources.NewQueue("work", resources.QueueSending)
	if err != nil {
		panic(err)
	}

	topic, err = resources.NewTopic("ping", resources.TopicPublishing)
	if err != nil {
		panic(err)
	}

	mainApi := resources.NewApi("nitric-testr")
	mainApi.Get("/history", historyGetHandler)
	mainApi.Delete("/history/:id", common.PathParser("/history/:id"), factDeleteHandler)

	mainApi.Post("/send", sendPostHandler)

	err = resources.Run()
	if err != nil {
		panic(err)
	}
}

// [END snippet]
