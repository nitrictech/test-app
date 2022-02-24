// [START snippet]

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/asalkeld/test-app/common"
	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/api/events"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	storeCol documents.CollectionRef
	history  documents.CollectionRef
	queue    queues.Queue
	topic    resources.Topic
)

// Updates context with error information
func httpError(ctx *faas.HttpContext, message string, status int) {
	ctx.Response.Body = []byte(message)
	ctx.Response.Status = status
}

func postHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	store := &common.Store{}
	if err := json.Unmarshal(ctx.Request.Data(), store); err != nil {
		httpError(ctx, "error decoding json body", 400)
		return ctx, nil
	}

	// get the current time and set the store time
	orderTime := time.Now()
	store.DateStored = orderTime.Format(time.RFC3339)

	// set the ID of the store
	id := uuid.New().String()
	store.ID = id

	// Convert the document to a map[string]interface{}
	// for storage, future iterations of the go-sdk may include direct interface{} storage as well
	storeMap := make(map[string]interface{})
	err := mapstructure.Decode(store, &storeMap)
	if err != nil {
		httpError(ctx, "error decoding store document", 400)
		return ctx, nil
	}

	if err := storeCol.Doc(id).Set(storeMap); err != nil {
		httpError(ctx, "error writing store document", 400)
		return ctx, nil
	}

	common.RecordFact(history, "store", "create", fmt.Sprint(storeMap))

	ctx.Response.Status = 200
	ctx.Response.Body = []byte(fmt.Sprintf("Created store with ID: %s", id))

	topic.Publish(&events.Event{ID: id, PayloadType: "create", Payload: storeMap})

	return next(ctx)
}

func listHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	query := storeCol.Query()
	results, err := query.Fetch()
	if err != nil {
		return nil, err
	}

	docs := make([]map[string]interface{}, 0)

	for _, doc := range results.Documents {
		// handle documents
		docs = append(docs, doc.Content())
	}

	b, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	ctx.Response.Body = b
	ctx.Response.Headers["Content-Type"] = []string{"application/json"}

	common.RecordFact(history, "store", "list", fmt.Sprint(len(docs)))

	return next(ctx)
}

func getHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params, ok := ctx.Extras["params"].(map[string]string)
	if !ok || params == nil {
		return nil, fmt.Errorf("error retrieving path params")
	}

	id := params["id"]

	doc, err := storeCol.Doc(id).Get()
	if err != nil {
		ctx.Response.Body = []byte("Error retrieving document " + id)
		ctx.Response.Status = 404
	} else {
		b, err := json.Marshal(doc.Content())
		if err != nil {
			return nil, err
		}

		ctx.Response.Headers["Content-Type"] = []string{"application/json"}
		ctx.Response.Body = b
	}

	common.RecordFact(history, "store", "list", fmt.Sprint(doc))

	return next(ctx)
}

func deleteHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params, ok := ctx.Extras["params"].(map[string]string)
	if !ok || params == nil {
		return nil, fmt.Errorf("error retrieving path params")
	}

	id := params["id"]

	err := storeCol.Doc(id).Delete()
	if err != nil {
		ctx.Response.Body = []byte("Error deleting document " + id)
		ctx.Response.Status = 404
	} else {
		ctx.Response.Status = 204
	}

	common.RecordFact(history, "store", "delete", id)
	queue.Send([]*queues.Task{
		{
			ID:          id,
			PayloadType: "delete",
		},
	})

	return next(ctx)
}

func main() {
	var err error

	storeCol, err = resources.NewCollection("store", resources.CollectionWriting, resources.CollectionReading, resources.CollectionDeleting)
	if err != nil {
		panic(err)
	}

	queue, err = resources.NewQueue("work", resources.QueueSending)
	if err != nil {
		panic(err)
	}

	topic, err = resources.NewTopic("ping")
	if err != nil {
		panic(err)
	}

	history, err = resources.NewCollection("history", resources.CollectionWriting)
	if err != nil {
		panic(err)
	}

	mainApi := resources.NewApi("nitric-testr")
	mainApi.Post("/store/", postHandler)
	mainApi.Get("/store/", listHandler)
	mainApi.Get("/store/:id", common.PathParser("/store/:id"), getHandler)
	mainApi.Delete("/store/:id", common.PathParser("/store/:id"), deleteHandler)

	err = resources.Run()
	if err != nil {
		panic(err)
	}
}

// [END snippet]
