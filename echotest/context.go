// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echotest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

// ContextConfig is configuration for creating echo.Context for testing purposes.
type ContextConfig struct {
	// Request will be used instead of default `httptest.NewRequest(http.MethodGet, "/", nil)`
	Request *http.Request

	// Response will be used instead of default `httptest.NewRecorder()`
	Response *httptest.ResponseRecorder

	// QueryValues wil be set as Request.URL.RawQuery value
	QueryValues url.Values

	// Headers wil be set as Request.Header value
	Headers http.Header

	// PathValues initializes context.PathValues with given value.
	PathValues echo.PathValues

	// RouteInfo initializes context.RouteInfo() with given value
	RouteInfo *echo.RouteInfo

	// FormValues creates form-urlencoded form out of given values. If there is no
	// `content-type` header it will be set to `application/x-www-form-urlencoded`
	// In case Request was not set the Request.Method is set to `POST`
	//
	// FormValues, MultipartForm and JSONBody are mutually exclusive.
	FormValues url.Values

	// MultipartForm creates multipart form out of given value. If there is no
	// `content-type` header it will be set to `multipart/form-data`
	// In case Request was not set the Request.Method is set to `POST`
	//
	// FormValues, MultipartForm and JSONBody are mutually exclusive.
	MultipartForm *MultipartForm

	// JSONBody creates JSON body out of given bytes. If there is no
	// `content-type` header it will be set to `application/json`
	// In case Request was not set the Request.Method is set to `POST`
	//
	// FormValues, MultipartForm and JSONBody are mutually exclusive.
	JSONBody []byte
}

// MultipartForm is used to create multipart form out of given value
type MultipartForm struct {
	Fields map[string]string
	Files  []MultipartFormFile
}

// MultipartFormFile is used to create file in multipart form out of given value
type MultipartFormFile struct {
	Fieldname string
	Filename  string
	Content   []byte
}

// ToContext converts ContextConfig to echo.Context
func (conf ContextConfig) ToContext(t *testing.T) *echo.Context {
	c, _ := conf.ToContextRecorder(t)
	return c
}

// ToContextRecorder converts ContextConfig to echo.Context and httptest.ResponseRecorder
func (conf ContextConfig) ToContextRecorder(t *testing.T) (*echo.Context, *httptest.ResponseRecorder) {
	if conf.Response == nil {
		conf.Response = httptest.NewRecorder()
	}
	isDefaultRequest := false
	if conf.Request == nil {
		isDefaultRequest = true
		conf.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	}

	if len(conf.QueryValues) > 0 {
		conf.Request.URL.RawQuery = conf.QueryValues.Encode()
	}
	if len(conf.Headers) > 0 {
		conf.Request.Header = conf.Headers
	}
	if len(conf.FormValues) > 0 {
		body := strings.NewReader(url.Values(conf.FormValues).Encode())
		conf.Request.Body = io.NopCloser(body)
		conf.Request.ContentLength = int64(body.Len())

		if conf.Request.Header.Get(echo.HeaderContentType) == "" {
			conf.Request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		}
		if isDefaultRequest {
			conf.Request.Method = http.MethodPost
		}
	} else if conf.MultipartForm != nil {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		for field, value := range conf.MultipartForm.Fields {
			if err := mw.WriteField(field, value); err != nil {
				t.Fatal(err)
			}
		}
		for _, file := range conf.MultipartForm.Files {
			fw, err := mw.CreateFormFile(file.Fieldname, file.Filename)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = fw.Write(file.Content); err != nil {
				t.Fatal(err)
			}
		}
		if err := mw.Close(); err != nil {
			t.Fatal(err)
		}

		conf.Request.Body = io.NopCloser(&body)
		conf.Request.ContentLength = int64(body.Len())
		if conf.Request.Header.Get(echo.HeaderContentType) == "" {
			conf.Request.Header.Set(echo.HeaderContentType, mw.FormDataContentType())
		}
		if isDefaultRequest {
			conf.Request.Method = http.MethodPost
		}
	} else if conf.JSONBody != nil {
		body := bytes.NewReader(conf.JSONBody)
		conf.Request.Body = io.NopCloser(body)
		conf.Request.ContentLength = int64(body.Len())

		if conf.Request.Header.Get(echo.HeaderContentType) == "" {
			conf.Request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		}
		if isDefaultRequest {
			conf.Request.Method = http.MethodPost
		}
	}

	ec := echo.NewContext(conf.Request, conf.Response, echo.New())
	if conf.RouteInfo == nil {
		conf.RouteInfo = &echo.RouteInfo{
			Name:       "",
			Method:     conf.Request.Method,
			Path:       "/test",
			Parameters: []string{},
		}
		for _, p := range conf.PathValues {
			conf.RouteInfo.Parameters = append(conf.RouteInfo.Parameters, p.Name)
		}
	}
	ec.InitializeRoute(conf.RouteInfo, &conf.PathValues)
	return ec, conf.Response
}

// ServeWithHandler serves ContextConfig with given handler and returns httptest.ResponseRecorder for response checking
func (conf ContextConfig) ServeWithHandler(t *testing.T, handler echo.HandlerFunc, opts ...any) *httptest.ResponseRecorder {
	c, rec := conf.ToContextRecorder(t)

	errHandler := echo.DefaultHTTPErrorHandler(false)
	for _, opt := range opts {
		switch o := opt.(type) {
		case echo.HTTPErrorHandler:
			errHandler = o
		}
	}

	err := handler(c)
	if err != nil {
		errHandler(c, err)
	}
	return rec
}
