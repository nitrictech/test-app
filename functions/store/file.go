package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nitrictech/go-sdk/faas"
	"github.com/nitrictech/test-app/common"
)

type fileRef struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func filePostHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	fmt.Println("filePost data:", string(hc.Request.Data()))

	f := &fileRef{}
	err := json.Unmarshal(hc.Request.Data(), &f)
	if err != nil {
		return next(common.HttpResponse(hc, "error un-marshalling:"+err.Error(), http.StatusBadRequest))
	}

	f.URL, err = bucky.File(f.Name).UploadUrl(hc.Request.Context(), int(time.Hour.Seconds()))
	if err != nil {
		return next(common.HttpResponse(hc, "error Putting:"+err.Error(), 400))
	}

	b, err := json.Marshal(f)
	if err != nil {
		return next(common.HttpResponse(hc, "error Marshalling:"+err.Error(), 400))
	}

	hc.Response.Status = 200
	hc.Response.Body = b

	return next(hc)
}

func fileGetHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	params := hc.Request.PathParams()
	if params == nil {
		return next(common.HttpResponse(hc, "error retrieving path params", 400))
	}

	name := params["name"]
	fmt.Println("fileGet name:", name)

	var err error

	f := &fileRef{
		Name: name,
	}

	f.URL, err = bucky.File(f.Name).DownloadUrl(hc.Request.Context(), int(time.Hour.Seconds()))
	if err != nil {
		return next(common.HttpResponse(hc, "error Getting:"+err.Error(), 400))
	}

	b, err := json.Marshal(f)
	if err != nil {
		return next(common.HttpResponse(hc, "error Marshalling:"+err.Error(), 400))
	}

	hc.Response.Status = 200
	hc.Response.Body = b

	return next(hc)
}

func filesGetHandler(hc *faas.HttpContext, next faas.HttpHandler) (*faas.HttpContext, error) {
	files, err := bucky.Files(hc.Request.Context())
	if err != nil {
		return next(common.HttpResponse(hc, err.Error(), 400))
	}

	frs := []*fileRef{}
	for _, f := range files {
		dURL, err := f.DownloadUrl(hc.Request.Context(), int(time.Hour.Seconds()))
		if err == nil {
			frs = append(frs, &fileRef{Name: f.Name(), URL: dURL})
		}
	}

	b, _ := json.Marshal(frs)

	hc.Response.Body = b
	hc.Response.Headers["Content-Type"] = []string{http.DetectContentType(hc.Response.Body)}
	fmt.Println("filesGet data:", frs)

	return next(hc)

}
