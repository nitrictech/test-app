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
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	storeCol documents.CollectionRef
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
	if store.ID == "" {
		store.ID = uuid.New().String()
	}

	// Convert the document to a map[string]interface{}
	// for storage, future iterations of the go-sdk may include direct interface{} storage as well
	storeMap := make(map[string]interface{})
	err := mapstructure.Decode(store, &storeMap)
	if err != nil {
		httpError(ctx, "error decoding store document", 400)
		return ctx, nil
	}

	if err := storeCol.Doc(store.ID).Set(storeMap); err != nil {
		httpError(ctx, "error writing store document", 400)
		return ctx, nil
	}

	ctx.Response.Status = 200
	ctx.Response.Body = []byte(fmt.Sprintf("Created store with ID: %s", store.ID))

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

	return next(ctx)
}

func putHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params, ok := ctx.Extras["params"].(map[string]string)
	if !ok || params == nil {
		return nil, fmt.Errorf("error retrieving path params")
	}

	id := params["id"]

	_, err := storeCol.Doc(id).Get()
	if err != nil {
		ctx.Response.Body = []byte("Error retrieving document " + id)
		ctx.Response.Status = 404
	} else {
		store := &common.Store{}
		if err := json.Unmarshal(ctx.Request.Data(), store); err != nil {
			httpError(ctx, "error decoding json body", 400)
			return ctx, nil
		}

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

		ctx.Response.Status = 200
		ctx.Response.Body = []byte(fmt.Sprintf("Updated store with ID: %s", id))
	}

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

	return next(ctx)
}

func main() {
	var err error

	storeCol, err = resources.NewCollection("store", resources.CollectionWriting, resources.CollectionReading, resources.CollectionDeleting)
	if err != nil {
		panic(err)
	}

	mainApi := resources.NewApi("nitric-testr")
	mainApi.Post("/store", postHandler)
	mainApi.Get("/store", listHandler)
	mainApi.Get("/store/:id", common.PathParser("/store/:id"), getHandler)
	mainApi.Put("/store/:id", common.PathParser("/store/:id"), putHandler)
	mainApi.Delete("/store/:id", common.PathParser("/store/:id"), deleteHandler)

	err = resources.Run()
	if err != nil {
		panic(err)
	}
}

// [END snippet]
