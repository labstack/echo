package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
)

type (
	ResponseRecorder struct {
		engine.Response
		Body *bytes.Buffer
	}
)

func NewRequest(method, url string, body io.Reader) engine.Request {
	r, _ := http.NewRequest(method, url, body)
	return standard.NewRequest(r)
}

func NewResponseRecorder() *ResponseRecorder {
	r := httptest.NewRecorder()
	return &ResponseRecorder{
		Response: standard.NewResponse(r),
		Body:     r.Body,
	}
}
