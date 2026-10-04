// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type mapBindingKey string
type mapBindingValue string
type mapBindingValues []string

func TestBindNamedMapTypes(t *testing.T) {
	for _, source := range []string{"query", "bind", "path", "header", "form", "multipart"} {
		for _, tc := range []struct {
			name   string
			target func() any
			want   any
		}{
			{"named_key", func() any { return new(map[mapBindingKey]string) }, map[mapBindingKey]string{"item": "first"}},
			{"named_both", func() any { return new(map[mapBindingKey]mapBindingValue) }, map[mapBindingKey]mapBindingValue{"item": "first"}},
			{"named_key_interface", func() any { return new(map[mapBindingKey]any) }, map[mapBindingKey]any{"item": "first"}},
			{"named_value", func() any { return new(map[string]mapBindingValue) }, map[string]mapBindingValue{"item": "first"}},
			{"named_slice", func() any { return new(map[string]mapBindingValues) }, map[string]mapBindingValues{"item": {"first", "second"}}},
			{"plain_string", func() any { return new(map[string]string) }, map[string]string{"item": "first"}},
			{"plain_slice", func() any { return new(map[string][]string) }, map[string][]string{"item": {"first", "second"}}},
			{"plain_interface", func() any { return new(map[string]any) }, map[string]any{"item": "first"}},
		} {
			t.Run(source+"/"+tc.name, func(t *testing.T) {
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("public binding panicked: %v", p)
					}
				}()
				e := echo.New()
				target := tc.target()
				e.Any("/:item", func(c *echo.Context) error {
					var err error
					switch source {
					case "bind":
						err = c.Bind(target)
					case "query":
						err = echo.BindQueryParams(c, target)
					case "path":
						err = echo.BindPathValues(c, target)
					case "header":
						err = echo.BindHeaders(c, target)
					default:
						err = echo.BindBody(c, target)
					}
					if err != nil {
						return err
					}
					return c.NoContent(http.StatusNoContent)
				})
				req := httptest.NewRequest(http.MethodPost, "/first?item=first&item=second", nil)
				if source == "bind" {
					req.Method = http.MethodGet
				}
				if source == "header" {
					req.Header.Set("Item", "first")
					req.Header.Add("Item", "second")
				}
				if source == "form" {
					req = httptest.NewRequest(http.MethodPost, "/first", strings.NewReader("item=first&item=second"))
					req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
				}
				if source == "multipart" {
					req = httptest.NewRequest(http.MethodPost, "/first", strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"item\"\r\n\r\nfirst\r\n--boundary\r\nContent-Disposition: form-data; name=\"item\"\r\n\r\nsecond\r\n--boundary--\r\n"))
					req.Header.Set(echo.HeaderContentType, "multipart/form-data; boundary=boundary")
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
				if rec.Code != http.StatusNoContent {
					t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
				}
				want := tc.want
				if source == "path" {
					switch tc.name {
					case "named_slice":
						want = map[string]mapBindingValues{"item": {"first"}}
					case "plain_slice":
						want = map[string][]string{"item": {"first"}}
					}
				}
				if source == "header" {
					v := reflect.ValueOf(want)
					value := v.MapIndex(reflect.ValueOf("item").Convert(v.Type().Key()))
					v.SetMapIndex(reflect.ValueOf("item").Convert(v.Type().Key()), reflect.Value{})
					v.SetMapIndex(reflect.ValueOf("Item").Convert(v.Type().Key()), value)
				}
				got := reflect.ValueOf(target).Elem().Interface()
				if !reflect.DeepEqual(want, got) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			})
		}
	}
}
