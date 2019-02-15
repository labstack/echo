package echo

import (
	"bytes"
	stdContext "context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name"`
	}
)

const (
	userJSON                    = `{"id":1,"name":"Jon Snow"}`
	userXML                     = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm                    = `id=1&name=Jon Snow`
	invalidContent              = "invalid content"
	userJSONInvalidType         = `{"id":"1","name":"Jon Snow"}`
	userXMLConvertNumberError   = `<user><id>Number one</id><name>Jon Snow</name></user>`
	userXMLUnsupportedTypeError = `<user><>Number one</><name>Jon Snow</name></user>`
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

const userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`

func TestEcho(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Router
	assert.NotNil(t, e.Router())

	// DefaultHTTPErrorHandler
	e.DefaultHTTPErrorHandler(errors.New("error"), c)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestEchoStatic(t *testing.T) {
	e := New()

	assert := assert.New(t)

	// OK
	e.Static("/images", "_fixture/images")
	c, b := request(http.MethodGet, "/images/walle.png", e)
	assert.Equal(http.StatusOK, c)
	assert.NotEmpty(b)

	// No file
	e.Static("/images", "_fixture/scripts")
	c, _ = request(http.MethodGet, "/images/bolt.png", e)
	assert.Equal(http.StatusNotFound, c)

	// Directory
	e.Static("/images", "_fixture/images")
	c, _ = request(http.MethodGet, "/images", e)
	assert.Equal(http.StatusNotFound, c)

	// Directory with index.html
	e.Static("/", "_fixture")
	c, r := request(http.MethodGet, "/", e)
	assert.Equal(http.StatusOK, c)
	assert.Equal(true, strings.HasPrefix(r, "<!doctype html>"))

	// Sub-directory with index.html
	c, r = request(http.MethodGet, "/folder", e)
	assert.Equal(http.StatusOK, c)
	assert.Equal(true, strings.HasPrefix(r, "<!doctype html>"))
}

func TestEchoFile(t *testing.T) {
	e := New()
	e.File("/walle", "_fixture/images/walle.png")
	c, b := request(http.MethodGet, "/walle", e)
	assert.Equal(t, http.StatusOK, c)
	assert.NotEmpty(t, b)
}

func TestEchoMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	// Route
	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(http.MethodGet, "/", e)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return errors.New("error")
		}
	})
	e.GET("/", NotFoundHandler)
	c, _ := request(http.MethodGet, "/", e)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestEchoHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(http.MethodGet, "/ok", e)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestEchoWrapHandler(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	if assert.NoError(t, h(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestEchoWrapMiddleware(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.Write([]byte("mw"))
			h.ServeHTTP(w, r)
		})
	})
	h := mw(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(c)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}

func TestEchoConnect(t *testing.T) {
	e := New()
	testMethod(t, http.MethodConnect, "/", e)
}

func TestEchoDelete(t *testing.T) {
	e := New()
	testMethod(t, http.MethodDelete, "/", e)
}

func TestEchoGet(t *testing.T) {
	e := New()
	testMethod(t, http.MethodGet, "/", e)
}

func TestEchoHead(t *testing.T) {
	e := New()
	testMethod(t, http.MethodHead, "/", e)
}

func TestEchoOptions(t *testing.T) {
	e := New()
	testMethod(t, http.MethodOptions, "/", e)
}

func TestEchoPatch(t *testing.T) {
	e := New()
	testMethod(t, http.MethodPatch, "/", e)
}

func TestEchoPost(t *testing.T) {
	e := New()
	testMethod(t, http.MethodPost, "/", e)
}

func TestEchoPut(t *testing.T) {
	e := New()
	testMethod(t, http.MethodPut, "/", e)
}

func TestEchoTrace(t *testing.T) {
	e := New()
	testMethod(t, http.MethodTrace, "/", e)
}

func TestEchoAny(t *testing.T) { // JFC
	e := New()
	e.Any("/", func(c Context) error {
		return c.String(http.StatusOK, "Any")
	})
}

func TestEchoMatch(t *testing.T) { // JFC
	e := New()
	e.Match([]string{http.MethodGet, http.MethodPost}, "/", func(c Context) error {
		return c.String(http.StatusOK, "Match")
	})
}

func TestEchoURL(t *testing.T) {
	e := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	e.GET("/static/file", static)
	e.GET("/users/:id", getUser)
	g := e.Group("/group")
	g.GET("/users/:uid/files/:fid", getFile)

	assert := assert.New(t)

	assert.Equal("/static/file", e.URL(static))
	assert.Equal("/users/:id", e.URL(getUser))
	assert.Equal("/users/1", e.URL(getUser, "1"))
	assert.Equal("/group/users/1/files/:fid", e.URL(getFile, "1"))
	assert.Equal("/group/users/1/files/1", e.URL(getFile, "1", "1"))
}

func TestEchoRoutes(t *testing.T) {
	e := New()
	routes := []*Route{
		{http.MethodGet, "/users/:user/events", ""},
		{http.MethodGet, "/users/:user/events/public", ""},
		{http.MethodPost, "/repos/:owner/:repo/git/refs", ""},
		{http.MethodPost, "/repos/:owner/:repo/git/tags", ""},
	}
	for _, r := range routes {
		e.Add(r.Method, r.Path, func(c Context) error {
			return c.String(http.StatusOK, "OK")
		})
	}

	if assert.Equal(t, len(routes), len(e.Routes())) {
		for _, r := range e.Routes() {
			found := false
			for _, rr := range routes {
				if r.Method == rr.Method && r.Path == rr.Path {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Route %s %s not found", r.Method, r.Path)
			}
		}
	}
}

func TestEchoEncodedPath(t *testing.T) {
	e := New()
	e.GET("/:id", func(c Context) error {
		return c.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/with%2Fslash", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEchoGroup(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	e.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))
	h := func(c Context) error {
		return c.NoContent(http.StatusOK)
	}

	//--------
	// Routes
	//--------

	e.GET("/users", h)

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.GET("", h)

	// Nested groups with middleware
	g2 := e.Group("/group2")
	g2.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})
	g3 := g2.Group("/group3")
	g3.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})
	g3.GET("", h)

	request(http.MethodGet, "/users", e)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request(http.MethodGet, "/group1", e)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request(http.MethodGet, "/group2/group3", e)
	assert.Equal(t, "023", buf.String())
}

func TestEchoNotFound(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEchoMethodNotAllowed(t *testing.T) {
	e := New()
	e.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "Echo!")
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestEchoContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(context), c)
	e.ReleaseContext(c)
}

func TestEchoStart(t *testing.T) {
	e := New()
	go func() {
		assert.NoError(t, e.Start(":0"))
	}()
	time.Sleep(200 * time.Millisecond)
}

func TestEchoStartTLS(t *testing.T) {
	e := New()
	go func() {
		err := e.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
		// Prevent the test to fail after closing the servers
		if err != http.ErrServerClosed {
			assert.NoError(t, err)
		}
	}()
	time.Sleep(200 * time.Millisecond)

	e.Close()
}

func TestEchoStartTLSByteString(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)

	testCases := []struct {
		cert        interface{}
		key         interface{}
		expectedErr error
		name        string
	}{
		{
			cert:        "_fixture/certs/cert.pem",
			key:         "_fixture/certs/key.pem",
			expectedErr: nil,
			name:        `ValidCertAndKeyFilePath`,
		},
		{
			cert:        cert,
			key:         key,
			expectedErr: nil,
			name:        `ValidCertAndKeyByteString`,
		},
		{
			cert:        cert,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidKeyType`,
		},
		{
			cert:        0,
			key:         key,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertType`,
		},
		{
			cert:        0,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertAndKeyTypes`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e.HideBanner = true

			go func() {
				err := e.StartTLS(":0", test.cert, test.key)
				if test.expectedErr != nil {
					require.EqualError(t, err, test.expectedErr.Error())
				} else if err != http.ErrServerClosed { // Prevent the test to fail after closing the servers
					require.NoError(t, err)
				}
			}()
			time.Sleep(200 * time.Millisecond)

			require.NoError(t, e.Close())
		})
	}
}

func TestEchoStartAutoTLS(t *testing.T) {
	e := New()
	errChan := make(chan error, 0)

	go func() {
		errChan <- e.StartAutoTLS(":0")
	}()
	time.Sleep(200 * time.Millisecond)

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	default:
		assert.NoError(t, e.Close())
	}
}

func testMethod(t *testing.T, method, path string, e *Echo) {
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(func(c Context) error {
		return c.String(http.StatusOK, method)
	})
	i := interface{}(e)
	reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{p, h})
	_, body := request(method, path, e)
	assert.Equal(t, method, body)
}

func request(method, path string, e *Echo) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestHTTPError(t *testing.T) {
	err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
		"code": 12,
	})
	assert.Equal(t, "code=400, message=map[code:12]", err.Error())
}

func TestEchoClose(t *testing.T) {
	e := New()
	errCh := make(chan error)

	go func() {
		errCh <- e.Start(":0")
	}()

	time.Sleep(200 * time.Millisecond)

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, e.Close())

	err := <-errCh
	assert.Equal(t, err.Error(), "http: Server closed")
}

func TestEchoShutdown(t *testing.T) {
	e := New()
	errCh := make(chan error)

	go func() {
		errCh <- e.Start(":0")
	}()

	time.Sleep(200 * time.Millisecond)

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 10*time.Second)
	defer cancel()
	assert.NoError(t, e.Shutdown(ctx))

	err := <-errCh
	assert.Equal(t, err.Error(), "http: Server closed")
}
