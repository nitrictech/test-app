// [START snippet]

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asalkeld/test-app/common"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/api/events"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/api/secrets"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	history documents.CollectionRef
	queue   queues.Queue
	topic   resources.Topic
	safe    secrets.SecretRef
)

func historyGetHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	query := history.Query()
	results, err := query.Fetch()
	if err != nil {
		return common.HttpResponse(ctx, "error querying collection: "+err.Error(), 500)
	}

	docs := make([]map[string]interface{}, 0)
	for _, doc := range results.Documents {
		docs = append(docs, doc.Content())
	}

	b, err := json.Marshal(docs)
	if err != nil {
		return common.HttpResponse(ctx, err.Error(), 400)
	}

	ctx.Response.Body = b
	ctx.Response.Headers["Content-Type"] = []string{"application/json"}

	return next(ctx)
}

func factDeleteHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params, ok := ctx.Extras["params"].(map[string]string)
	if !ok || params == nil {
		return common.HttpResponse(ctx, "error retrieving path params", 400)
	}

	id := params["id"]

	err := history.Doc(id).Delete()
	if err != nil {
		common.HttpResponse(ctx, "Error deleting document "+id+" err "+err.Error(), 404)
	} else {
		ctx.Response.Status = 204
	}

	return next(ctx)
}

func sendPostHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	m := &common.Message{}
	if err := json.Unmarshal(ctx.Request.Data(), m); err != nil {
		return common.HttpResponse(ctx, "error decoding json body", 400)
	}

	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	mMap := make(map[string]interface{})
	err := mapstructure.Decode(m, &mMap)
	if err != nil {
		return common.HttpResponse(ctx, "error decoding message document", 400)
	}

	switch strings.ToLower(m.MessageType) {
	case "topic":
		_, err = topic.Publish(&events.Event{
			ID:          m.ID,
			PayloadType: m.PayloadType,
			Payload:     mMap,
		})
	case "queue":
		_, err = queue.Send([]*queues.Task{
			{
				ID:          m.ID,
				PayloadType: m.PayloadType,
				Payload:     mMap,
			},
		})
	}
	if err != nil {
		common.HttpResponse(ctx, "error sending:"+err.Error(), 400)
	} else {
		ctx.Response.Status = 200
		ctx.Response.Body = []byte(fmt.Sprintf("Run action : %v", m))
	}

	return next(ctx)
}

func safePostHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	_, err := safe.Put(ctx.Request.Data())
	if err != nil {
		common.HttpResponse(ctx, "error Putting:"+err.Error(), 400)
	} else {
		ctx.Response.Status = 200
	}

	return next(ctx)
}

func safeGetHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	sv, err := safe.Latest().Access()
	if err != nil {
		return common.HttpResponse(ctx, err.Error(), 400)
	}

	ctx.Response.Body = sv.AsBytes()
	ctx.Response.Headers["Content-Type"] = []string{http.DetectContentType(ctx.Response.Body)}

	return next(ctx)
}

func run() error {
	var err error
	history, err = resources.NewCollection("history", resources.CollectionReading, resources.CollectionDeleting)
	if err != nil {
		return err
	}

	safe, err = resources.NewSecret("safe", resources.SecretEverything...)
	if err != nil {
		return err
	}

	queue, err = resources.NewQueue("work", resources.QueueSending)
	if err != nil {
		return err
	}

	topic, err = resources.NewTopic("ping", resources.TopicPublishing)
	if err != nil {
		return err
	}

	mainApi := resources.NewApi("nitric-testr")
	mainApi.Get("/history", historyGetHandler)
	mainApi.Delete("/history/:id", common.PathParser("/history/:id"), factDeleteHandler)

	mainApi.Post("/send", sendPostHandler)

	mainApi.Post("/safe", safePostHandler)
	mainApi.Get("/safe", safeGetHandler)

	return resources.Run()
}

func main() {
	err := run()
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		panic(err)
	}
}
