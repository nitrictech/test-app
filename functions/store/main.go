package main

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"

	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/api/secrets"
	"github.com/nitrictech/go-sdk/api/storage"
	"github.com/nitrictech/go-sdk/resources"
)

var (
	mainApi  resources.Api
	storeCol documents.CollectionRef
	history  documents.CollectionRef
	queue    queues.Queue
	topic    resources.Topic
	safe     secrets.SecretRef
	bucky    storage.Bucket
)

func run() error {
	if os.Getenv("OTELCOL_BIN") != "" {
		ctx := context.TODO()
		tp, err := newTraceProvider(ctx)
		if err != nil {
			return err
		}

		otel.SetTracerProvider(tp)
		defer func() {
			tp.ForceFlush(ctx)
			_ = tp.Shutdown(ctx)
		}()
	}

	var err error

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

	bucky, err = resources.NewBucket("bucky", resources.BucketEverything...)
	if err != nil {
		return err
	}

	storeCol, err = resources.NewCollection("store", resources.CollectionWriting, resources.CollectionReading, resources.CollectionDeleting)
	if err != nil {
		return err
	}

	history, err = resources.NewCollection("history", resources.CollectionWriting, resources.CollectionReading, resources.CollectionDeleting)
	if err != nil {
		return err
	}

	mainApi, err = resources.NewApi("nitric-testr")
	if err != nil {
		return err
	}

	mainApi.Get("/history", historyGetHandler)
	mainApi.Delete("/history/:id", factDeleteHandler)

	mainApi.Post("/send", sendPostHandler)

	mainApi.Post("/safe", safePostHandler)
	mainApi.Get("/safe", safeGetHandler)

	mainApi.Post("/file", filePostHandler)
	mainApi.Get("/file", filesGetHandler)
	mainApi.Get("/file/:name", fileGetHandler)

	mainApi.Post("/store", postHandler)
	mainApi.Get("/store", listHandler)
	mainApi.Get("/store/:id", getHandler)
	mainApi.Put("/store/:id", putHandler)
	mainApi.Delete("/store/:id", deleteHandler)

	err = resources.Run()
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
