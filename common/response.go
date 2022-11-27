package common

import (
	"fmt"

	"github.com/nitrictech/go-sdk/faas"
)

// Updates context with error information
func HttpResponse(hc *faas.HttpContext, message string, status int) *faas.HttpContext {
	fmt.Println(message)
	hc.Response.Body = []byte(message)
	hc.Response.Status = status

	return hc
}
