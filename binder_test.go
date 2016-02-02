package echo

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type (
	customer struct {
		ID   int    `json:"id" xml:"id" form:"id"`
		Name string `json:"name" xml:"name" form:"name"`
	}

	testStruct struct {
		I           int
		I8          int8
		I16         int16
		I32         int32
		I64         int64
		UI          uint
		UI8         uint8
		UI16        uint16
		UI32        uint32
		UI64        uint64
		B           bool
		F32         float32
		F64         float64
		S           string
		cantSet     string
		DoesntExist string
	}
)

func (t testStruct) GetCantSet() string {
	return t.cantSet
}

var values = map[string][]string{
	"I":       {"0"},
	"I8":      {"8"},
	"I16":     {"16"},
	"I32":     {"32"},
	"I64":     {"64"},
	"UI":      {"0"},
	"UI8":     {"8"},
	"UI16":    {"16"},
	"UI32":    {"32"},
	"UI64":    {"64"},
	"B":       {"true"},
	"F32":     {"32.5"},
	"F64":     {"64.5"},
	"S":       {"test"},
	"cantSet": {"test"},
}

const (
	customerJSON     = `{"id":1,"name":"Joe"}`
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
	r, _ = http.NewRequest(POST, "/", nil)
	testBindError(t, r, ApplicationForm)
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
	testBindError(t, r, "")
}

func TestBindFormFunc(t *testing.T) {
	r, _ := http.NewRequest(POST, "/", strings.NewReader(customerForm))
	r.Header.Set(ContentType, ApplicationForm)
	b := new(binder)
	c := new(customer)
	if assert.NoError(t, b.bindForm(r, c)) {
		assertCustomer(t, c)
	}
}

func TestBindMultiPartFormFunc(t *testing.T) {
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	mw.WriteField("id", "1")
	mw.WriteField("name", "Joe")
	mw.Close()
	r, _ := http.NewRequest(POST, "/", body)
	r.Header.Set(ContentType, mw.FormDataContentType())
	b := new(binder)
	c := new(customer)
	if assert.NoError(t, b.bindMultiPartForm(r, c)) {
		assertCustomer(t, c)
	}
}

func assertCustomer(t *testing.T, c *customer) {
	assert.Equal(t, 1, c.ID)
	assert.Equal(t, "Joe", c.Name)
}

func TestMapForm(t *testing.T) {
	ts := new(testStruct)
	mapForm(ts, values)
	assertTestStruct(t, ts)
}

func TestSetWithProperType(t *testing.T) {
	ts := new(testStruct)
	typ := reflect.TypeOf(ts).Elem()
	val := reflect.ValueOf(ts).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		if len(values[typeField.Name]) == 0 {
			continue
		}
		val := values[typeField.Name][0]
		err := setWithProperType(typeField.Type.Kind(), val, structField)
		assert.NoError(t, err)
	}
	assertTestStruct(t, ts)

	type foo struct {
		Bar bytes.Buffer
	}
	v := &foo{}
	typ = reflect.TypeOf(v).Elem()
	val = reflect.ValueOf(v).Elem()
	assert.Error(t, setWithProperType(typ.Field(0).Type.Kind(), "5", val.Field(0)))
}

func TestSetFields(t *testing.T) {
	ts := new(testStruct)
	val := reflect.ValueOf(ts).Elem()
	// Int
	if assert.NoError(t, setIntField("5", 0, val.FieldByName("I"))) {
		assert.Equal(t, 5, ts.I)
	}
	if assert.NoError(t, setIntField("", 0, val.FieldByName("I"))) {
		assert.Equal(t, 0, ts.I)
	}

	// Uint
	if assert.NoError(t, setUintField("10", 0, val.FieldByName("UI"))) {
		assert.Equal(t, uint(10), ts.UI)
	}
	if assert.NoError(t, setUintField("", 0, val.FieldByName("UI"))) {
		assert.Equal(t, uint(0), ts.UI)
	}

	// Float
	if assert.NoError(t, setFloatField("15.5", 0, val.FieldByName("F32"))) {
		assert.Equal(t, float32(15.5), ts.F32)
	}
	if assert.NoError(t, setFloatField("", 0, val.FieldByName("F32"))) {
		assert.Equal(t, float32(0.0), ts.F32)
	}

	// Bool
	if assert.NoError(t, setBoolField("true", val.FieldByName("B"))) {
		assert.Equal(t, true, ts.B)
	}
	if assert.NoError(t, setBoolField("", val.FieldByName("B"))) {
		assert.Equal(t, false, ts.B)
	}
}

func assertTestStruct(t *testing.T, ts *testStruct) {
	assert.Equal(t, 0, ts.I)
	assert.Equal(t, int8(8), ts.I8)
	assert.Equal(t, int16(16), ts.I16)
	assert.Equal(t, int32(32), ts.I32)
	assert.Equal(t, int64(64), ts.I64)
	assert.Equal(t, uint(0), ts.UI)
	assert.Equal(t, uint8(8), ts.UI8)
	assert.Equal(t, uint16(16), ts.UI16)
	assert.Equal(t, uint32(32), ts.UI32)
	assert.Equal(t, uint64(64), ts.UI64)
	assert.Equal(t, true, ts.B)
	assert.Equal(t, float32(32.5), ts.F32)
	assert.Equal(t, float64(64.5), ts.F64)
	assert.Equal(t, "test", ts.S)
	assert.Equal(t, "", ts.GetCantSet())
}

func testBindOk(t *testing.T, r *http.Request, ct string) {
	r.Header.Set(ContentType, ct)
	c := new(customer)
	err := new(binder).Bind(r, c)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, c.ID)
		assert.Equal(t, "Joe", c.Name)
	}
}

func testBindError(t *testing.T, r *http.Request, ct string) {
	r.Header.Set(ContentType, ct)
	u := new(customer)
	err := new(binder).Bind(r, u)

	switch {
	case strings.HasPrefix(ct, ApplicationJSON), strings.HasPrefix(ct, ApplicationXML),
		strings.HasPrefix(ct, ApplicationForm), strings.HasPrefix(ct, MultipartForm):
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).code)
		}
	default:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, ErrUnsupportedMediaType, err)
		}

	}
}
