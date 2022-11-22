package main

import (
	"encoding/json"

	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/test-app/common"
)

func historyGetHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	query := history.Query()
	results, err := query.Fetch(hc.Request.Context())
	if err != nil {
		return common.HttpResponse(hc, "error querying collection: "+err.Error(), 500)
	}

	docs := make([]map[string]interface{}, 0)
	for _, doc := range results.Documents {
		docs = append(docs, doc.Content())
	}

	b, err := json.Marshal(docs)
	if err != nil {
		return common.HttpResponse(hc, err.Error(), 400)
	}

	hc.Response.Body = b
	hc.Response.Headers["Content-Type"] = []string{"application/json"}

	return next(hc)
}

func factDeleteHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params := hc.Request.PathParams()
	if params == nil {
		return common.HttpResponse(hc, "error retrieving path params", 400)
	}

	id := params["id"]

	err := history.Doc(id).Delete(hc.Request.Context())
	if err != nil {
		_, _ = common.HttpResponse(hc, "Error deleting document "+id+" err "+err.Error(), 404)
	} else {
		hc.Response.Status = 204
	}

	return next(hc)
}
