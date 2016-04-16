package echo

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"text/template"

	"strings"

	"net/url"

	"encoding/xml"

	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

type (
	Template struct {
		templates *template.Template
	}
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func TestContext(t *testing.T) {
	userJSON := `{"id":"1","name":"Joe"}`
	userXML := `<user><id>1</id><name>Joe</name></user>`
	invalidContent := "invalid content"

	e := New()
	rq := test.NewRequest(POST, "/", strings.NewReader(userJSON))
	rc := test.NewResponseRecorder()
	c := e.NewContext(rq, rc).(*context)

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	// ParamNames
	c.pnames = []string{"uid", "fid"}
	assert.EqualValues(t, []string{"uid", "fid"}, c.ParamNames())

	// Param by id
	c.pnames = []string{"id"}
	c.pvalues = []string{"1"}
	assert.Equal(t, "1", c.P(0))

	// Param by name
	assert.Equal(t, "1", c.Param("id"))

	// Store
	c.Set("user", "Joe")
	assert.Equal(t, "Joe", c.Get("user"))

	//------
	// Bind
	//------

	// JSON
	testBindOk(t, c, MIMEApplicationJSON)
	c.request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, MIMEApplicationJSON)

	// XML
	c.request = test.NewRequest(POST, "/", strings.NewReader(userXML))
	testBindOk(t, c, MIMEApplicationXML)
	c.request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, MIMEApplicationXML)

	// Unsupported
	testBindError(t, c, "")

	//--------
	// Render
	//--------

	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.echo.SetRenderer(tpl)
	err := c.Render(http.StatusOK, "hello", "Joe")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, "Hello, Joe!", rc.Body.String())
	}

	c.echo.renderer = nil
	err = c.Render(http.StatusOK, "hello", "Joe")
	assert.Error(t, err)

	// JSON
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rc.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON, rc.Body.String())
	}

	// JSON (error)
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// JSONP
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rc.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+");", rc.Body.String())
	}

	// XML
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.XML(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rc.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rc.Body.String())
	}

	// XML (error)
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.XML(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// String
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, MIMETextPlainCharsetUTF8, rc.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, World!", rc.Body.String())
	}

	// HTML
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rc.Status())
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rc.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rc.Body.String())
	}

	// Attachment
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	file, err := os.Open("_fixture/images/walle.png")
	if assert.NoError(t, err) {
		err = c.Attachment(file, "walle.png")
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rc.Status())
			assert.Equal(t, "attachment; filename=walle.png", rc.Header().Get(HeaderContentDisposition))
			assert.Equal(t, 219885, rc.Body.Len())
		}
	}

	// NoContent
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rc.Status())

	// Redirect
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(t, http.StatusMovedPermanently, rc.Status())
	assert.Equal(t, "http://labstack.github.io/echo", rc.Header().Get(HeaderLocation))

	// Error
	rc = test.NewResponseRecorder()
	c = e.NewContext(rq, rc).(*context)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, rc.Status())

	// Reset
	c.Reset(rq, test.NewResponseRecorder())
}

func TestContextPath(t *testing.T) {
	e := New()
	r := e.Router()

	r.Add(GET, "/users/:id", nil, e)
	c := e.NewContext(nil, nil)
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "/users/:id", c.Path())

	r.Add(GET, "/users/:uid/files/:fid", nil, e)
	c = e.NewContext(nil, nil)
	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "/users/:uid/files/:fid", c.Path())
}

func TestContextQueryParam(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "joe")
	q.Set("email", "joe@labstack.com")
	rq := test.NewRequest(GET, "/?"+q.Encode(), nil)
	e := New()
	c := e.NewContext(rq, nil)
	assert.Equal(t, "joe", c.QueryParam("name"))
	assert.Equal(t, "joe@labstack.com", c.QueryParam("email"))
}

func TestContextFormValue(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "joe")
	f.Set("email", "joe@labstack.com")

	e := New()
	rq := test.NewRequest(POST, "/", strings.NewReader(f.Encode()))
	rq.Header().Add(HeaderContentType, MIMEApplicationForm)

	c := e.NewContext(rq, nil)
	assert.Equal(t, "joe", c.FormValue("name"))
	assert.Equal(t, "joe@labstack.com", c.FormValue("email"))
}

func TestContextNetContext(t *testing.T) {
	// c := new(context)
	// c.Context = xcontext.WithValue(nil, "key", "val")
	// assert.Equal(t, "val", c.Value("key"))
}

func TestContextServeContent(t *testing.T) {
	e := New()
	rq := test.NewRequest(GET, "/", nil)
	rc := test.NewResponseRecorder()
	c := e.NewContext(rq, rc)

	fs := http.Dir("_fixture/images")
	f, err := fs.Open("walle.png")
	if assert.NoError(t, err) {
		fi, err := f.Stat()
		if assert.NoError(t, err) {
			// Not cached
			if assert.NoError(t, c.ServeContent(f, fi.Name(), fi.ModTime())) {
				assert.Equal(t, http.StatusOK, rc.Status())
			}

			// Cached
			rc = test.NewResponseRecorder()
			c = e.NewContext(rq, rc)
			rq.Header().Set(HeaderIfModifiedSince, fi.ModTime().UTC().Format(http.TimeFormat))
			if assert.NoError(t, c.ServeContent(f, fi.Name(), fi.ModTime())) {
				assert.Equal(t, http.StatusNotModified, rc.Status())
			}
		}
	}
}

func TestContextHandler(t *testing.T) {
	e := New()
	r := e.Router()
	b := new(bytes.Buffer)

	r.Add(GET, "/handler", func(Context) error {
		_, err := b.Write([]byte("handler"))
		return err
	}, e)
	c := e.NewContext(nil, nil)
	r.Find(GET, "/handler", c)
	c.Handler()(c)
	assert.Equal(t, "handler", b.String())
}

func testBindOk(t *testing.T, c Context, ct string) {
	c.Request().Header().Set(HeaderContentType, ct)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, "1", u.ID)
		assert.Equal(t, "Joe", u.Name)
	}
}

func testBindError(t *testing.T, c Context, ct string) {
	c.Request().Header().Set(HeaderContentType, ct)
	u := new(user)
	err := c.Bind(u)

	switch ct {
	case MIMEApplicationJSON, MIMEApplicationXML:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
		}
	default:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, ErrUnsupportedMediaType, err)
		}

	}
}
