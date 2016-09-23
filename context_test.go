package echo

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"text/template"
	"time"

	"golang.org/x/net/context"

	"strings"

	"net/url"

	"encoding/xml"

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
	e := New()
	req, _ := http.NewRequest(POST, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*echoContext)

	// Echo
	assert.Equal(t, e, c.Echo())

	// Request
	assert.NotNil(t, c.Request())

	// Response
	assert.NotNil(t, c.Response())

	//--------
	// Render
	//--------

	tmpl := &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	c.echo.Renderer = tmpl
	err := c.Render(http.StatusOK, "hello", "Jon Snow")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, Jon Snow!", rec.Body.String())
	}

	c.echo.Renderer = nil
	err = c.Render(http.StatusOK, "hello", "Jon Snow")
	assert.Error(t, err)

	// JSON
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.JSON(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, userJSON, rec.Body.String())
	}

	// JSON (error)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// JSONP
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	callback := "callback"
	err = c.JSONP(http.StatusOK, callback, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationJavaScriptCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, callback+"("+userJSON+");", rec.Body.String())
	}

	// XML
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.XML(http.StatusOK, user{1, "Jon Snow"})
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMEApplicationXMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, xml.Header+userXML, rec.Body.String())
	}

	// XML (error)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.XML(http.StatusOK, make(chan bool))
	assert.Error(t, err)

	// String
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.String(http.StatusOK, "Hello, World!")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextPlainCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// HTML
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	err = c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(t, "Hello, <strong>World!</strong>", rec.Body.String())
	}

	// Stream
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	r := strings.NewReader("response from a stream")
	err = c.Stream(http.StatusOK, "application/octet-stream", r)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/octet-stream", rec.Header().Get(HeaderContentType))
		assert.Equal(t, "response from a stream", rec.Body.String())
	}

	// Attachment
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	file, err := os.Open("_fixture/images/walle.png")
	if assert.NoError(t, err) {
		err = c.Attachment(file, "walle.png")
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "attachment; filename=walle.png", rec.Header().Get(HeaderContentDisposition))
			assert.Equal(t, 219885, rec.Body.Len())
		}
	}

	// Inline
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	file, err = os.Open("_fixture/images/walle.png")
	if assert.NoError(t, err) {
		err = c.Inline(file, "walle.png")
		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "inline; filename=walle.png", rec.Header().Get(HeaderContentDisposition))
			assert.Equal(t, 219885, rec.Body.Len())
		}
	}

	// NoContent
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	c.NoContent(http.StatusOK)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Error
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*echoContext)
	c.Error(errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Reset
	c.Reset(req, httptest.NewRecorder())
}

func TestContextCookie(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/", nil)
	theme := "theme=light"
	user := "user=Jon Snow"
	req.Header.Add(HeaderCookie, theme)
	req.Header.Add(HeaderCookie, user)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*echoContext)

	// Read single
	cookie, err := c.Cookie("theme")
	if assert.NoError(t, err) {
		assert.Equal(t, "theme", cookie.Name)
		assert.Equal(t, "light", cookie.Value)
	}

	// Read multiple
	for _, cookie := range c.Cookies() {
		switch cookie.Name {
		case "theme":
			assert.Equal(t, "light", cookie.Value)
		case "user":
			assert.Equal(t, "Jon Snow", cookie.Value)
		}
	}

	// Write
	cookie = &http.Cookie{
		Name:     "SSID",
		Value:    "Ap4PGTEq",
		Domain:   "labstack.com",
		Path:     "/",
		Expires:  time.Now(),
		Secure:   true,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "SSID")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "Ap4PGTEq")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "labstack.com")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "Secure")
	assert.Contains(t, rec.Header().Get(HeaderSetCookie), "HttpOnly")
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

func TestContextPathParam(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/", nil)
	c := e.NewContext(req, nil)

	// ParamNames
	c.SetParamNames("uid", "fid")
	assert.EqualValues(t, []string{"uid", "fid"}, c.ParamNames())

	// ParamValues
	c.SetParamValues("101", "501")
	assert.EqualValues(t, []string{"101", "501"}, c.ParamValues())

	// P
	assert.Equal(t, "101", c.P(0))

	// Param
	assert.Equal(t, "501", c.Param("fid"))
}

func TestContextFormValue(t *testing.T) {
	f := make(url.Values)
	f.Set("name", "Jon Snow")
	f.Set("email", "jon@labstack.com")

	e := New()
	req, _ := http.NewRequest(POST, "/", strings.NewReader(f.Encode()))
	req.Header.Add(HeaderContentType, MIMEApplicationForm)
	c := e.NewContext(req, nil)

	// FormValue
	assert.Equal(t, "Jon Snow", c.FormValue("name"))
	assert.Equal(t, "jon@labstack.com", c.FormValue("email"))

	// FormParams
	params, err := c.FormParams()
	if assert.NoError(t, err) {
		assert.Equal(t, url.Values{
			"name":  []string{"Jon Snow"},
			"email": []string{"jon@labstack.com"},
		}, params)
	}
}

func TestContextQueryParam(t *testing.T) {
	q := make(url.Values)
	q.Set("name", "Jon Snow")
	q.Set("email", "jon@labstack.com")
	req, _ := http.NewRequest(GET, "/?"+q.Encode(), nil)
	e := New()
	c := e.NewContext(req, nil)

	// QueryParam
	assert.Equal(t, "Jon Snow", c.QueryParam("name"))
	assert.Equal(t, "jon@labstack.com", c.QueryParam("email"))

	// QueryParams
	assert.Equal(t, url.Values{
		"name":  []string{"Jon Snow"},
		"email": []string{"jon@labstack.com"},
	}, c.QueryParams())
}

func TestContextFormFile(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	mr := multipart.NewWriter(buf)
	w, err := mr.CreateFormFile("file", "test")
	if assert.NoError(t, err) {
		w.Write([]byte("test"))
	}
	mr.Close()
	req, _ := http.NewRequest(POST, "/", buf)
	req.Header.Set(HeaderContentType, mr.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	f, err := c.FormFile("file")
	if assert.NoError(t, err) {
		assert.Equal(t, "test", f.Filename)
	}
}

func TestContextMultipartForm(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	mw.WriteField("name", "Jon Snow")
	mw.Close()
	req, _ := http.NewRequest(POST, "/", buf)
	req.Header.Set(HeaderContentType, mw.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	f, err := c.MultipartForm()
	if assert.NoError(t, err) {
		assert.NotNil(t, f)
	}
}

func TestContextRedirect(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.Equal(t, nil, c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://labstack.github.io/echo", rec.Header().Get(HeaderLocation))
	assert.Error(t, c.Redirect(310, "http://labstack.github.io/echo"))
}

func TestContextEmbedded(t *testing.T) {
	var c Context
	c = new(echoContext)
	c.SetStdContext(context.WithValue(c, "key", "val"))
	assert.Equal(t, "val", c.Value("key"))
	now := time.Now()
	ctx, _ := context.WithDeadline(context.Background(), now)
	c.SetStdContext(ctx)
	n, _ := ctx.Deadline()
	assert.Equal(t, now, n)
	assert.Equal(t, context.DeadlineExceeded, c.Err())
	assert.NotNil(t, c.Done())
}

func TestContextStore(t *testing.T) {
	var c Context
	c = new(echoContext)
	c.Set("name", "Jon Snow")
	assert.Equal(t, "Jon Snow", c.Get("name"))
}

func TestContextServeContent(t *testing.T) {
	e := New()
	req, _ := http.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	fs := http.Dir("_fixture/images")
	f, err := fs.Open("walle.png")
	if assert.NoError(t, err) {
		fi, err := f.Stat()
		if assert.NoError(t, err) {
			// Not cached
			if assert.NoError(t, c.ServeContent(f, fi.Name(), fi.ModTime())) {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			// Cached
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)
			req.Header.Set(HeaderIfModifiedSince, fi.ModTime().UTC().Format(http.TimeFormat))
			if assert.NoError(t, c.ServeContent(f, fi.Name(), fi.ModTime())) {
				assert.Equal(t, http.StatusNotModified, rec.Code)
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
