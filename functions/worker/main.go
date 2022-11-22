package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/api/documents"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/go-sdk/resources"
	"github.com/nitrictech/test-app/common"
	"google.golang.org/grpc/status"
)

// 4
var (
	history documents.CollectionRef
	queue   queues.Queue
	topic   resources.Topic
)

func main() {
	var err error
	history, err = resources.NewCollection("history", resources.CollectionWriting)
	if err != nil {
		panic(err)
	}

	queue, err = resources.NewQueue("work", resources.QueueReceving)
	if err != nil {
		panic(err)
	}

	topic, err = resources.NewTopic("ping")
	if err != nil {
		panic(err)
	}

	topic.Subscribe(func(ec *faas.EventContext, next faas.EventHandler) (*faas.EventContext, error) {
		fmt.Printf("received on %s mesg %s", ec.Request.Topic(), string(ec.Request.Data()))
		common.RecordFact(ec.Request.Context(), history, ec.Request.Topic(), "received event", string(ec.Request.Data()))
		return next(ec)
	})

	err = resources.NewSchedule("five-min-schedule", "5 minutes", func(ec *faas.EventContext, next faas.EventHandler) (*faas.EventContext, error) {
		fmt.Println("scheduled job")

		tasks, err := queue.Receive(ec.Request.Context(), 10)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		fmt.Printf("got (%d) tasks\n", len(tasks))

		for _, task := range tasks {
			msg := &common.Message{}
			err := mapstructure.Decode(task.Task().Payload, msg)
			if err != nil {
				fmt.Println(err)
				continue
			}

			b, err := json.Marshal(msg)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if err = task.Complete(ec.Request.Context()); err != nil {
				fmt.Println(err)

				if s, ok := status.FromError(err); ok {
					for _, item := range s.Details() {
						fmt.Printf("%v", item)
					}

					fmt.Printf("%v\n", s)
				} else {
					fmt.Printf("err type %T\n", err)
				}
			} else {
				common.RecordFact(ec.Request.Context(), history, queue.Name(), "task complete", string(b))
			}
		}

		return next(ec)
	})
	if err != nil {
		panic(err)
	}

	err = resources.Run()
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		panic(err)
	}
}

// [END snippet]
