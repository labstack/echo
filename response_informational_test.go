// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_Informational(t *testing.T) {
	created := func(c *echo.Context) error {
		return c.String(http.StatusCreated, "created")
	}
	implicit := func(c *echo.Context) error {
		_, err := c.Response().Write([]byte("created"))
		return err
	}
	handlerError := func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "unavailable")
	}
	jsonResponse := func(c *echo.Context) error {
		return c.JSON(http.StatusAccepted, "created")
	}

	for _, protocol := range []string{"HTTP/1.1", "HTTP/2.0"} {
		t.Run(protocol, func(t *testing.T) {
			for _, tc := range []struct {
				name              string
				method            string
				handler           echo.HandlerFunc
				middleware        echo.MiddlewareFunc
				presetStatus      int
				hintsInSerializer bool
				wantStatus        int
				wantBody          string
				wantEncoding      string
			}{
				{name: "explicit status", handler: created, wantStatus: http.StatusCreated, wantBody: "created"},
				{name: "implicit status", handler: implicit, wantStatus: http.StatusOK, wantBody: "created"},
				{name: "preset status", handler: implicit, presetStatus: http.StatusAccepted, wantStatus: http.StatusAccepted, wantBody: "created"},
				{name: "JSON", handler: jsonResponse, wantStatus: http.StatusAccepted, wantBody: "\"created\"\n"},
				{name: "handler error", handler: handlerError, wantStatus: http.StatusServiceUnavailable, wantBody: "{\"message\":\"unavailable\"}\n"},
				{
					name: "no content",
					handler: func(c *echo.Context) error {
						return c.NoContent(http.StatusNoContent)
					},
					wantStatus: http.StatusNoContent,
				},
				{name: "HEAD", method: http.MethodHead, handler: created, wantStatus: http.StatusCreated},
				{name: "HEAD error", method: http.MethodHead, handler: handlerError, wantStatus: http.StatusServiceUnavailable},
				{name: "gzip", handler: created, middleware: middleware.Gzip(), wantStatus: http.StatusCreated, wantBody: "created", wantEncoding: "gzip"},
				{
					name:       "below gzip threshold",
					handler:    created,
					middleware: middleware.GzipWithConfig(middleware.GzipConfig{MinLength: 100}),
					wantStatus: http.StatusCreated,
					wantBody:   "created",
				},
				{name: "gzip error", handler: handlerError, middleware: middleware.Gzip(), wantStatus: http.StatusServiceUnavailable, wantBody: "{\"message\":\"unavailable\"}\n"},
				{name: "JSON serializer", handler: jsonResponse, hintsInSerializer: true, wantStatus: http.StatusAccepted, wantBody: "\"created\"\n"},
				{
					name: "JSON serializer error",
					handler: func(c *echo.Context) error {
						return c.JSON(http.StatusCreated, make(chan int))
					},
					hintsInSerializer: true,
					wantStatus:        http.StatusInternalServerError,
					wantBody:          "{\"message\":\"Internal Server Error\"}\n",
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					e := echo.NewWithConfig(echo.Config{
						Router: echo.NewRouter(echo.RouterConfig{AutoHandleHEAD: true}),
					})
					if tc.middleware != nil {
						e.Use(tc.middleware)
					}
					var logs bytes.Buffer
					e.Logger = slog.New(slog.NewTextHandler(&logs, nil))

					var beforeStatuses []int
					var afterSizes []int64
					firstHintReceived := make(chan struct{})
					codes := []int{http.StatusContinue, http.StatusProcessing, http.StatusEarlyHints, http.StatusEarlyHints, 199}
					writeHints := func(c *echo.Context) {
						res, err := echo.UnwrapResponse(c.Response())
						if !assert.NoError(t, err) {
							return
						}
						status := res.Status
						for i, code := range codes {
							c.Response().Header().Set("Link", "</style"+strconv.Itoa(i)+".css>; rel=preload; as=style")
							c.Response().WriteHeader(code)
							if i == 0 {
								select {
								case <-firstHintReceived:
								case <-c.Request().Context().Done():
									t.Error("informational response was not sent before the final response")
									return
								}
							}
							assert.False(t, res.Committed)
							assert.Equal(t, status, res.Status)
							assert.Zero(t, res.Size)
							assert.Empty(t, beforeStatuses)
							assert.Empty(t, afterSizes)
						}
						c.Response().Header().Del("Link")
						c.Response().Header().Set("X-Final", "value")
					}
					if tc.hintsInSerializer {
						hintsWritten := false
						e.JSONSerializer = informationalJSONSerializer{writeHints: func(c *echo.Context) {
							if !hintsWritten {
								hintsWritten = true
								writeHints(c)
							}
						}}
					}
					e.GET("/", func(c *echo.Context) error {
						res, err := echo.UnwrapResponse(c.Response())
						if err != nil {
							return err
						}
						if tc.presetStatus != 0 {
							res.Status = tc.presetStatus
						}
						res.Before(func() {
							beforeStatuses = append(beforeStatuses, res.Status)
							c.Response().Header().Set("X-Before", "final")
						})
						res.After(func() {
							assert.True(t, res.Committed)
							assert.Equal(t, tc.wantStatus, res.Status)
							afterSizes = append(afterSizes, res.Size)
						})
						if !tc.hintsInSerializer {
							writeHints(c)
						}
						return tc.handler(c)
					})

					done := make(chan struct{})
					server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						defer close(done)
						e.ServeHTTP(w, r)
					}))
					if protocol == "HTTP/2.0" {
						server.EnableHTTP2 = true
						server.StartTLS()
					} else {
						server.Start()
					}
					t.Cleanup(server.Close)
					client := server.Client()
					client.Timeout = 5 * time.Second

					method := tc.method
					if method == "" {
						method = http.MethodGet
					}
					req, err := http.NewRequest(method, server.URL, nil)
					require.NoError(t, err)
					req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
					var gotCodes []int
					var gotHeaders []http.Header
					req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
						Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
							gotCodes = append(gotCodes, code)
							gotHeaders = append(gotHeaders, http.Header(header).Clone())
							if len(gotCodes) == 1 {
								close(firstHintReceived)
							}
							return nil
						},
					}))
					resp, err := client.Do(req)
					require.NoError(t, err)
					defer resp.Body.Close()
					wireBody, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					body := wireBody
					if resp.Header.Get(echo.HeaderContentEncoding) == "gzip" {
						zr, err := gzip.NewReader(bytes.NewReader(wireBody))
						require.NoError(t, err)
						defer zr.Close()
						body, err = io.ReadAll(zr)
						require.NoError(t, err)
					}
					require.NoError(t, resp.Body.Close())
					<-done

					assert.Equal(t, protocol, resp.Proto)
					assert.Equal(t, codes, gotCodes)
					for i, header := range gotHeaders {
						assert.Equal(t, "</style"+strconv.Itoa(i)+".css>; rel=preload; as=style", header.Get("Link"))
						assert.Empty(t, header.Get("X-Final"))
						assert.Empty(t, header.Get("X-Before"))
						assert.Empty(t, header.Get(echo.HeaderContentEncoding))
					}
					assert.Equal(t, tc.wantStatus, resp.StatusCode)
					assert.Equal(t, tc.wantBody, string(body))
					assert.Equal(t, tc.wantEncoding, resp.Header.Get(echo.HeaderContentEncoding))
					assert.Empty(t, resp.Header.Get("Link"))
					assert.Equal(t, "value", resp.Header.Get("X-Final"))
					assert.Equal(t, "final", resp.Header.Get("X-Before"))
					assert.Equal(t, []int{tc.wantStatus}, beforeStatuses)
					assert.Empty(t, logs.String())
					if tc.name == "HEAD" {
						assert.Equal(t, int64(len("created")), resp.ContentLength)
						assert.Equal(t, []int64{int64(len("created"))}, afterSizes)
					} else if len(wireBody) != 0 {
						require.NotEmpty(t, afterSizes)
						assert.Equal(t, int64(len(wireBody)), afterSizes[len(afterSizes)-1])
					}
				})
			}
		})
	}
}

type informationalJSONSerializer struct {
	echo.DefaultJSONSerializer
	writeHints func(*echo.Context)
}

func (s informationalJSONSerializer) Serialize(c *echo.Context, target any, indent string) error {
	s.writeHints(c)
	return s.DefaultJSONSerializer.Serialize(c, target, indent)
}

func TestResponse_FinalStatusRemainsCommitted(t *testing.T) {
	for _, code := range []int{http.StatusSwitchingProtocols, http.StatusCreated, http.StatusNoContent, http.StatusNotModified, http.StatusBadRequest, http.StatusInternalServerError} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			e := echo.New()
			var logs bytes.Buffer
			e.Logger = slog.New(slog.NewTextHandler(&logs, nil))
			e.GET("/", func(c *echo.Context) error {
				res, err := echo.UnwrapResponse(c.Response())
				if err != nil {
					return err
				}
				beforeCalls := 0
				res.Before(func() { beforeCalls++ })
				c.Response().WriteHeader(code)
				c.Response().WriteHeader(http.StatusEarlyHints)
				c.Response().WriteHeader(http.StatusAccepted)
				assert.True(t, res.Committed)
				assert.Equal(t, code, res.Status)
				assert.Equal(t, 1, beforeCalls)
				return nil
			})
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, code, rec.Code)
			assert.Empty(t, rec.Body.String())
			assert.Equal(t, 2, strings.Count(logs.String(), "echo: response already written to client"))
		})
	}
}

func TestResponse_InformationalAfterFinal(t *testing.T) {
	for _, mode := range []string{"net/http", "Echo", "gzip", "buffered gzip", "buffered JSON"} {
		t.Run(mode, func(t *testing.T) {
			for _, tc := range []struct {
				name           string
				explicitStatus bool
				bodyFirst      bool
				headerOnly     bool
			}{
				{name: "explicit status", explicitStatus: true},
				{name: "header only", explicitStatus: true, headerOnly: true},
				{name: "implicit status", bodyFirst: true},
				{name: "explicit status and body", explicitStatus: true, bodyFirst: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					const body = "\"created\"\n"
					writeResponse := func(w http.ResponseWriter) error {
						w.Header().Set("X-Response", "retained")
						if tc.explicitStatus {
							w.WriteHeader(http.StatusCreated)
						}
						if tc.bodyFirst {
							if _, err := io.WriteString(w, body); err != nil {
								return err
							}
						}
						w.WriteHeader(http.StatusEarlyHints)
						if !tc.headerOnly && !tc.bodyFirst {
							_, err := io.WriteString(w, body)
							return err
						}
						return nil
					}

					var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						assert.NoError(t, writeResponse(w))
					})
					if mode != "net/http" {
						e := echo.New()
						switch mode {
						case "gzip":
							e.Use(middleware.Gzip())
						case "buffered gzip", "buffered JSON":
							e.Use(middleware.GzipWithConfig(middleware.GzipConfig{MinLength: 100}))
						}
						if mode == "buffered JSON" {
							e.JSONSerializer = responseWriterJSONSerializer{writeResponse: writeResponse}
						}
						e.GET("/", func(c *echo.Context) error {
							if mode == "buffered JSON" {
								return c.JSON(http.StatusOK, nil)
							}
							return writeResponse(c.Response())
						})
						handler = e
					}

					server := httptest.NewServer(handler)
					t.Cleanup(server.Close)
					client := server.Client()
					client.Timeout = 5 * time.Second
					req, err := http.NewRequest(http.MethodGet, server.URL, nil)
					require.NoError(t, err)
					req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
					var gotCodes []int
					req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
						Got1xxResponse: func(code int, _ textproto.MIMEHeader) error {
							gotCodes = append(gotCodes, code)
							return nil
						},
					}))
					resp, err := client.Do(req)
					require.NoError(t, err)
					defer resp.Body.Close()
					var reader io.Reader = resp.Body
					if resp.Header.Get(echo.HeaderContentEncoding) == "gzip" {
						zr, err := gzip.NewReader(resp.Body)
						require.NoError(t, err)
						defer zr.Close()
						reader = zr
					}
					gotBody, err := io.ReadAll(reader)
					require.NoError(t, err)
					require.NoError(t, resp.Body.Close())

					assert.Empty(t, gotCodes)
					wantStatus := http.StatusOK
					if tc.explicitStatus {
						wantStatus = http.StatusCreated
					}
					assert.Equal(t, wantStatus, resp.StatusCode)
					assert.Equal(t, "retained", resp.Header.Get("X-Response"))
					if tc.headerOnly {
						assert.Empty(t, gotBody)
					} else {
						assert.Equal(t, body, string(gotBody))
					}
				})
			}
		})
	}
}

type responseWriterJSONSerializer struct {
	echo.DefaultJSONSerializer
	writeResponse func(http.ResponseWriter) error
}

func (s responseWriterJSONSerializer) Serialize(c *echo.Context, _ any, _ string) error {
	return s.writeResponse(c.Response())
}
