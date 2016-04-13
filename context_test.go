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
	rec := test.NewResponseRecorder()
	c := NewContext(rq, rec, e)

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	// ParamNames
	c.Object().pnames = []string{"uid", "fid"}
	assert.EqualValues(t, []string{"uid", "fid"}, c.ParamNames())

	// Param by id
	c.Object().pnames = []string{"id"}
	c.Object().pvalues = []string{"1"}
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
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, MIMEApplicationJSON)

	// XML
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(userXML))
	testBindOk(t, c, MIMEApplicationXML)
	c.Object().request = test.NewRequest(POST, "/", strings.NewReader(invalidContent))
	testBindError(t, c, MIMEApplicationXML)

	// Unsupported
	testBindError(t, c, "")

	//--------
	// Render
	//--------

	tpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.Object().echo.SetRenderer(tpl)
	err := c.Render(http.StatusOK, "hello", "Joe")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, "Hello, Joe!", rec.Body.String())
	}

	c.Object().echo.renderer = nil
	err = c.Render(http.StatusOK, "hello", "Joe")
	assert.Error(t, err)

	// JSON
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.JSON(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}

	// JSON (error)
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// JSONP
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}

	// XML
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.XML(http.StatusOK, user{"1", "Joe"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// XML (error)
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.XML(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// String
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, MIMETextPlainCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Status())
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// Attachment
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	file, err := os.Open("_fixture/images/walle.png")
	if assert.NoError(t, err) {
		err = c.Attachment(file, "walle.png")
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Status())
			assert.Equal(t, "attachment; filename=walle.png", rec.Header().Get(HeaderContentDisposition))
			assert.Equal(t, 219885, rec.Body.Len())
		}
	}

	// NoContent
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Status())

	// Redirect
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(t, http.StatusMovedPermanently, rec.Status())
	assert.Equal(t, "http://labstack.github.io/echo", rec.Header().Get(HeaderLocation))

	// Error
	rec = test.NewResponseRecorder()
	c = NewContext(rq, rec, e)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, rec.Status())

	// Reset
	c.Object().Reset(rq, test.NewResponseRecorder())
}

func TestContextPath(t *testing.T) {
	e := New()
	r := e.Router()

	r.Add(GET, "/users/:id", nil, e)
	c := NewContext(nil, nil, e)
	r.Find(GET, "/users/1", c)
	assert.Equal(t, "/users/:id", c.Path())

	r.Add(GET, "/users/:uid/files/:fid", nil, e)
	c = NewContext(nil, nil, e)
	r.Find(GET, "/users/1/files/1", c)
	assert.Equal(t, "/users/:uid/files/:fid", c.Path())
}

func TestContextQueryParam(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "joe")
	q.Set("email", "joe@labstack.com")
	rq := test.NewRequest(GET, "/?"+q.Encode(), nil)
	c := NewContext(rq, nil, New())
	assert.Equal(t, "joe", c.QueryParam("name"))
	assert.Equal(t, "joe@labstack.com", c.QueryParam("email"))
}

func TestContextFormValue(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "joe")
	f.Set("email", "joe@labstack.com")

	rq := test.NewRequest(POST, "/", strings.NewReader(f.Encode()))
	rq.Header().Add(HeaderContentType, MIMEApplicationForm)

	c := NewContext(rq, nil, New())
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
	c := NewContext(rq, rc, e)

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
			c = NewContext(rq, rc, e)
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
	c := NewContext(nil, nil, e)
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
