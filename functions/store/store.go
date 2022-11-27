package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/test-app/common"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func postHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	ctx, span := otel.Tracer("functions/store").Start(hc.Request.Context(), hc.Request.Path())

	span.SetAttributes(
		semconv.CodeFunctionKey.String("postHandler"),
		semconv.HTTPMethodKey.String(hc.Request.Method()),
		semconv.HTTPTargetKey.String(hc.Request.Path()),
	)

	defer span.End()

	store := &common.Store{}
	if err := json.Unmarshal(hc.Request.Data(), store); err != nil {
		return next(common.HttpResponse(hc, "error decoding json body", 400))
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
		return next(common.HttpResponse(hc, "error decoding store document", 400))
	}

	if err := storeCol.Doc(store.ID).Set(ctx, storeMap); err != nil {
		return next(common.HttpResponse(hc, "error writing store document", 400))
	}

	//span.SetAttributes(attribute.String("store.id", store.ID))
	return next(common.HttpResponse(hc, fmt.Sprintf("Created store with ID: %s", store.ID), 200))
}

func listHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	query := storeCol.Query()
	results, err := query.Fetch(hc.Request.Context())
	if err != nil {
		return next(common.HttpResponse(hc, "error querying collection: "+err.Error(), 500))
	}

	docs := make([]map[string]interface{}, 0)

	for _, doc := range results.Documents {
		// handle documents
		docs = append(docs, doc.Content())
	}

	b, err := json.Marshal(docs)
	if err != nil {
		return next(common.HttpResponse(hc, err.Error(), 400))
	}

	hc.Response.Body = b
	hc.Response.Headers["Content-Type"] = []string{"application/json"}

	return next(hc)
}

func getHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params := hc.Request.PathParams()
	if params == nil {
		return next(common.HttpResponse(hc, "error retrieving path params", 400))
	}

	id := params["id"]

	doc, err := storeCol.Doc(id).Get(hc.Request.Context())
	if err != nil {
		return next(common.HttpResponse(hc, "error retrieving document "+id, 404))
	}

	b, err := json.Marshal(doc.Content())
	if err != nil {
		return next(common.HttpResponse(hc, err.Error(), 400))
	}

	hc.Response.Headers["Content-Type"] = []string{"application/json"}
	hc.Response.Body = b

	return next(hc)
}

func putHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params := hc.Request.PathParams()
	if params == nil {
		return next(common.HttpResponse(hc, "error retrieving path params", 400))
	}

	id := params["id"]

	_, err := storeCol.Doc(id).Get(hc.Request.Context())
	if err != nil {
		return next(common.HttpResponse(hc, "Error retrieving document "+id, 404))
	}

	store := &common.Store{}
	if err := json.Unmarshal(hc.Request.Data(), store); err != nil {
		return next(common.HttpResponse(hc, "error decoding json body", 400))
	}

	// Convert the document to a map[string]interface{}
	// for storage, future iterations of the go-sdk may include direct interface{} storage as well
	storeMap := make(map[string]interface{})
	err = mapstructure.Decode(store, &storeMap)
	if err != nil {
		return next(common.HttpResponse(hc, "error decoding store document", 400))
	}

	if err := storeCol.Doc(id).Set(hc.Request.Context(), storeMap); err != nil {
		return next(common.HttpResponse(hc, "error writing store document", 400))
	}

	return next(common.HttpResponse(hc, fmt.Sprintf("Updated store with ID: %s", id), 200))
}

func deleteHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params := hc.Request.PathParams()
	if params == nil {
		return next(common.HttpResponse(hc, "error retrieving path params", 400))
	}

	id := params["id"]
	err := storeCol.Doc(id).Delete(hc.Request.Context())
	if err != nil {
		return next(common.HttpResponse(hc, "error deleting document "+id, 400))
	}

	hc.Response.Status = 204

	return next(hc)
}
