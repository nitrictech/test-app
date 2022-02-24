// [START snippet]

package main

import (
	"encoding/json"

	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	history documents.CollectionRef
)

// Updates context with error information
func httpError(ctx *faas.HttpContext, message string, status int) {
	ctx.Response.Body = []byte(message)
	ctx.Response.Status = status
}

func handler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
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

func main() {
	var err error
	history, err = resources.NewCollection("history", resources.CollectionReading)
	if err != nil {
		panic(err)
	}

	mainApi := resources.NewApi("nitric-testr")
	mainApi.Get("/history/", handler)

	err = resources.Run()
	if err != nil {
		panic(err)
	}
}

// [END snippet]
