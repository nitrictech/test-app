package main

import (
	"fmt"
	"net/http"

	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/test-app/common"
)

func safePostHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	fmt.Println("safePost data:", string(hc.Request.Data()))
	_, err := safe.Put(hc.Request.Context(), hc.Request.Data())
	if err != nil {
		return next(common.HttpResponse(hc, "error Putting:"+err.Error(), 400))
	}

	hc.Response.Status = 200

	return next(hc)
}

func safeGetHandler(ctx *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	sv, err := safe.Latest().Access()
	if err != nil {
		return next(common.HttpResponse(ctx, err.Error(), 400))
	}
	ctx.Response.Body = sv.AsBytes()
	ctx.Response.Headers["Content-Type"] = []string{http.DetectContentType(ctx.Response.Body)}
	fmt.Println("safeGet data:", sv.AsString())

	return next(ctx)
}
