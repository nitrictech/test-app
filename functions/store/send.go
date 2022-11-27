package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/nitrictech/go-sdk/api/events"
	"github.com/nitrictech/go-sdk/api/queues"
	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/test-app/common"
)

func sendPostHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	fmt.Println("sendPostHandler")
	m := &common.Message{}
	if err := json.Unmarshal(hc.Request.Data(), m); err != nil {
		return next(common.HttpResponse(hc, "error decoding json body", 400))
	}

	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	mMap := make(map[string]interface{})

	err := mapstructure.Decode(m, &mMap)
	if err != nil {
		return next(common.HttpResponse(hc, "error decoding message document", 400))
	}

	switch strings.ToLower(m.MessageType) {
	case "topic":
		_, err = topic.Publish(hc.Request.Context(),
			&events.Event{
				ID:          m.ID,
				PayloadType: m.PayloadType,
				Payload:     mMap,
			}, events.WithDelay(time.Duration(m.Delay)*time.Second))
	case "queue":
		_, err = queue.Send(hc.Request.Context(), []*queues.Task{
			{
				ID:          m.ID,
				PayloadType: m.PayloadType,
				Payload:     mMap,
			},
		})
	default:
		err = fmt.Errorf("unknown message type %s", m.MessageType)
	}
	if err != nil {
		return next(common.HttpResponse(hc, "error sending:"+err.Error(), 400))
	}

	fmt.Printf("sent message id %s", m.ID)
	hc.Response.Status = 200
	hc.Response.Body = []byte(fmt.Sprintf("Run action : %v", m))

	return next(hc)
}
