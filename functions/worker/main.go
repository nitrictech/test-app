// [START snippet]

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
)

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
		common.RecordFact(history, ec.Request.Topic(), "received event", string(ec.Request.Data()))
		return next(ec)
	})

	err = resources.NewSchedule("job", "1 minutes", func(ec *faas.EventContext, next faas.EventHandler) (*faas.EventContext, error) {
		fmt.Println("got scheduled event ", string(ec.Request.Data()))
		tasks, err := queue.Receive(10)
		if err != nil {
			fmt.Println(err)
			return nil, err
		} else {
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

				common.RecordFact(history, queue.Name(), "task complete", string(b))
				task.Complete()
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
