package echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestApp struct{}

func (t TestApp) Add2Numbers(c Context) error {
	type numbersT struct {
		A int64
		B int64
	}
	var numbers numbersT
	if err := c.Bind(&numbers); err != nil {
		return err
	}
	c.String(200, fmt.Sprintf("%d", numbers.A+numbers.B))
	return nil
}

func (t TestApp) Multiply2Numbers(c Context) error {
	type numbersT struct {
		A int64
		B int64
	}
	var numbers numbersT
	if err := c.Bind(&numbers); err != nil {
		return err
	}
	c.String(200, fmt.Sprintf("%d", numbers.A*numbers.B))
	return nil
}

func (t TestApp) Route(mp Mountpoint) {
	mp.GET("/add2numbers", t.Add2Numbers)
	mp.POST("/multiply2numbers", t.Multiply2Numbers)
}

func TestMountpointWithEcho(t *testing.T) {
	// Test with Echo instance
	e := New()
	app := TestApp{}
	app.Route(e)

	// Test GET route with Echo
	jsonBody := `{"A": 2, "B": 3}`
	req := httptest.NewRequest(http.MethodGet, "/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "5", rec.Body.String())
}

func TestMountpointWithGroup(t *testing.T) {
	e := New()
	g := e.Group("/v1")
	app := TestApp{}
	app.Route(g)

	// Test GET route with group prefix
	jsonBody := `{"A": 10, "B": 20}`
	req := httptest.NewRequest(http.MethodGet, "/v1/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "30", rec.Body.String())

	// Test POST route with group prefix
	jsonBody = `{"A": 5, "B": 5}`
	req = httptest.NewRequest(http.MethodPost, "/v1/multiply2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "25", rec.Body.String())
}

func TestNestedMountpoint(t *testing.T) {
	e := New()
	g1 := e.Group("/v1")
	g2 := g1.Group("/math")
	app := TestApp{}
	app.Route(g2)

	// Test GET route with nested group prefixes
	jsonBody := `{"A": 100, "B": 50}`
	req := httptest.NewRequest(http.MethodGet, "/v1/math/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "150", rec.Body.String())

	// Test POST route with nested group prefixes
	jsonBody = `{"A": 25, "B": 4}`
	req = httptest.NewRequest(http.MethodPost, "/v1/math/multiply2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "100", rec.Body.String())
}

func TestMultipleMountpoints(t *testing.T) {
	e := New()

	// Create two different groups
	v1 := e.Group("/v1")
	v2 := e.Group("/v2")

	// Route the same app to both groups
	app := TestApp{}
	app.Route(v1)
	app.Route(v2)

	// Test v1 endpoint
	jsonBody := `{"A": 10, "B": 10}`
	req := httptest.NewRequest(http.MethodGet, "/v1/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "20", rec.Body.String())

	// Test v2 endpoint
	req = httptest.NewRequest(http.MethodGet, "/v2/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "20", rec.Body.String())
}

func TestMountpointWithMiddleware(t *testing.T) {
	e := New()

	// Create middleware that adds a header to the response
	headerMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			c.Response().Header().Set("X-Test-Header", "middleware-applied")
			return next(c)
		}
	}

	// Create a group with middleware
	g := e.Group("/api", headerMiddleware)

	// Route the app to the group
	app := TestApp{}
	app.Route(g)

	// Test endpoint with middleware
	jsonBody := `{"A": 7, "B": 8}`
	req := httptest.NewRequest(http.MethodGet, "/api/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Check status and body
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "15", rec.Body.String())

	// Check that middleware was applied
	assert.Equal(t, "middleware-applied", rec.Header().Get("X-Test-Header"))

	// Test with middleware applied to the mountpoint directly
	e = New()
	g = e.Group("/api")
	g.Use(headerMiddleware)
	app.Route(g)

	req = httptest.NewRequest(http.MethodGet, "/api/add2numbers", strings.NewReader(jsonBody))
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "15", rec.Body.String())
	assert.Equal(t, "middleware-applied", rec.Header().Get("X-Test-Header"))
}
