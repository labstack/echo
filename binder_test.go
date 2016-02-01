package echo

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

type (
	customer struct {
		ID   string `json:"id" xml:"id" form:"id"`
		Name string `json:"name" xml:"name" form:"name"`
	}
)

const (
	customerJSON     = `{"id":"1","name":"Joe"}`
	customerXML      = `<customer><id>1</id><name>Joe</name></customer>`
	customerForm     = `id=1&name=Joe`
	incorrectContent = "this is incorrect content"
)

func TestMaxMemory(t *testing.T) {
	b := new(binder)
	b.SetMaxMemory(20)
	assert.Equal(t, int64(20), b.MaxMemory())
}

func TestJSONBinding(t *testing.T) {
	r, _ := http.NewRequest(POST, "/", strings.NewReader(customerJSON))
	testBindOk(t, r, ApplicationJSON)
	r, _ = http.NewRequest(POST, "/", strings.NewReader(incorrectContent))
	testBindError(t, r, ApplicationJSON)
}

func TestXMLBinding(t *testing.T) {
	r, _ := http.NewRequest(POST, "/", strings.NewReader(customerXML))
	testBindOk(t, r, ApplicationXML)
	r, _ = http.NewRequest(POST, "/", strings.NewReader(incorrectContent))
	testBindError(t, r, ApplicationXML)
}

func TestFormBinding(t *testing.T) {
	r, _ := http.NewRequest(POST, "/", strings.NewReader(customerForm))
	testBindOk(t, r, ApplicationForm)
}

func TestMultipartFormBinding(t *testing.T) {
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	mw.WriteField("id", "1")
	mw.WriteField("name", "Joe")
	mw.Close()
	r, _ := http.NewRequest(POST, "/", body)
	testBindOk(t, r, mw.FormDataContentType())
	r, _ = http.NewRequest(POST, "/", strings.NewReader(incorrectContent))
	testBindError(t, r, mw.FormDataContentType())
}

func TestUnsupportedMediaTypeBinding(t *testing.T) {
	r, _ := http.NewRequest(POST, "/", strings.NewReader(customerJSON))

	// Unsupported
	testBindError(t, r, "")
}

func testBindOk(t *testing.T, r *http.Request, ct string) {
	r.Header.Set(ContentType, ct)
	u := new(customer)
	err := new(binder).Bind(r, u)
	if assert.NoError(t, err) {
		assert.Equal(t, "1", u.ID)
		assert.Equal(t, "Joe", u.Name)
	}
}

func testBindError(t *testing.T, r *http.Request, ct string) {
	r.Header.Set(ContentType, ct)
	u := new(customer)
	err := new(binder).Bind(r, u)

	switch {
	case strings.HasPrefix(ct, ApplicationJSON), strings.HasPrefix(ct, ApplicationXML), strings.HasPrefix(ct, ApplicationForm), strings.HasPrefix(ct, MultipartForm):
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).code)
		}
	default:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, ErrUnsupportedMediaType, err)
		}

	}
}
