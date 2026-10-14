package echo_test

// Regression tests for https://github.com/labstack/echo/issues/3169
// "BindHeaders panics on nil or empty header value slices".
//
// bindData indexes inputValue[0] on the scalar-map paths and on the tagged-struct
// paths without checking the length. A header whose value slice is nil or empty
// therefore panicked with "index out of range [0] with length 0" instead of
// binding nothing. An empty value slice is now treated like an absent key.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestBindHeadersEmptyValueSlice(t *testing.T) {
	// nil and []string{} are distinct states that both reach the same panic.
	// They are not the same as []string{""}, which binds an empty string.
	for name, values := range map[string][]string{
		"nil":   nil,
		"empty": {},
	} {
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header["X-Empty"] = values

			var dst struct {
				Value string `header:"X-Empty"`
			}
			if err := echo.BindHeaders(e.NewContext(req, httptest.NewRecorder()), &dst); err != nil {
				t.Fatalf("BindHeaders returned an error: %v", err)
			}
			if dst.Value != "" {
				t.Errorf("expected an empty value, got %q", dst.Value)
			}
		})
	}
}

func TestBindHeadersEmptyValueSliceIntoMap(t *testing.T) {
	// The scalar-map destination path reads v[0] as well.
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header["X-Empty"] = nil

	dst := map[string]string{}
	if err := echo.BindHeaders(e.NewContext(req, httptest.NewRecorder()), &dst); err != nil {
		t.Fatalf("BindHeaders returned an error: %v", err)
	}
	if _, ok := dst["X-Empty"]; ok {
		t.Errorf("expected no entry for an empty header value, got %v", dst)
	}
}

func TestBindHeadersPresentValueStillBinds(t *testing.T) {
	// The guard must not swallow a header that actually has a value.
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header["X-Token"] = []string{"abc123"}

	var dst struct {
		Token string `header:"X-Token"`
	}
	if err := echo.BindHeaders(e.NewContext(req, httptest.NewRecorder()), &dst); err != nil {
		t.Fatalf("BindHeaders returned an error: %v", err)
	}
	if dst.Token != "abc123" {
		t.Errorf("expected abc123, got %q", dst.Token)
	}
}

func TestBindHeadersEmptyStringValueStillBinds(t *testing.T) {
	// []string{""} is a present, empty value - distinct from an empty slice and
	// must keep binding, so the guard keys on length and not on emptiness.
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header["X-Empty"] = []string{""}

	var dst struct {
		Value string `header:"X-Empty"`
	}
	if err := echo.BindHeaders(e.NewContext(req, httptest.NewRecorder()), &dst); err != nil {
		t.Fatalf("BindHeaders returned an error: %v", err)
	}
	if dst.Value != "" {
		t.Errorf("expected an empty value, got %q", dst.Value)
	}
}
