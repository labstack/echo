package echo

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
)

type (
	binderTestStruct struct {
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

func (t binderTestStruct) GetCantSet() string {
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

func TestBinderJSON(t *testing.T) {
	testBinderOkay(t, strings.NewReader(userJSON), MIMEApplicationJSON)
	testBinderError(t, strings.NewReader(invalidContent), MIMEApplicationJSON)
}

func TestBinderXML(t *testing.T) {
	testBinderOkay(t, strings.NewReader(userXML), MIMEApplicationXML)
	testBinderError(t, strings.NewReader(invalidContent), MIMEApplicationXML)
}

func TestBinderForm(t *testing.T) {
	testBinderOkay(t, strings.NewReader(userForm), MIMEApplicationForm)
	testBinderError(t, nil, MIMEApplicationForm)
	e := New()
	req := test.NewRequest(POST, "/", strings.NewReader(userForm))
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	req.Header().Set(HeaderContentType, MIMEApplicationForm)
	var obj = make([]struct{ Field string }, 0)
	err := c.Bind(&obj)
	assert.Error(t, err)
}

func TestBinderQueryParams(t *testing.T) {
	e := New()
	req := test.NewRequest(GET, "/?id=1&name=Jon Snow", nil)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}
}

func TestBinderMultipartForm(t *testing.T) {
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	mw.WriteField("id", "1")
	mw.WriteField("name", "Jon Snow")
	mw.Close()
	testBinderOkay(t, body, mw.FormDataContentType())
}

func TestBinderUnsupportedMediaType(t *testing.T) {
	testBinderError(t, strings.NewReader(invalidContent), MIMEApplicationJSON)
}

// func assertCustomer(t *testing.T, c *user) {
// 	assert.Equal(t, 1, c.ID)
// 	assert.Equal(t, "Joe", c.Name)
// }

func TestBinderbindForm(t *testing.T) {
	ts := new(binderTestStruct)
	b := new(binder)
	b.bindData(ts, values)
	assertBinderTestStruct(t, ts)
}

func TestBinderSetWithProperType(t *testing.T) {
	ts := new(binderTestStruct)
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
	assertBinderTestStruct(t, ts)

	type foo struct {
		Bar bytes.Buffer
	}
	v := &foo{}
	typ = reflect.TypeOf(v).Elem()
	val = reflect.ValueOf(v).Elem()
	assert.Error(t, setWithProperType(typ.Field(0).Type.Kind(), "5", val.Field(0)))
}

func TestBinderSetFields(t *testing.T) {
	ts := new(binderTestStruct)
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

func assertBinderTestStruct(t *testing.T, ts *binderTestStruct) {
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

func testBinderOkay(t *testing.T, r io.Reader, ctype string) {
	e := New()
	req := test.NewRequest(POST, "/", r)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	req.Header().Set(HeaderContentType, ctype)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}
}

func testBinderError(t *testing.T, r io.Reader, ctype string) {
	e := New()
	req := test.NewRequest(POST, "/", r)
	rec := test.NewResponseRecorder()
	c := e.NewContext(req, rec)
	req.Header().Set(HeaderContentType, ctype)
	u := new(user)
	err := c.Bind(u)

	switch {
	case strings.HasPrefix(ctype, MIMEApplicationJSON), strings.HasPrefix(ctype, MIMEApplicationXML),
		strings.HasPrefix(ctype, MIMEApplicationForm), strings.HasPrefix(ctype, MIMEMultipartForm):
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, http.StatusBadRequest, err.(*HTTPError).Code)
		}
	default:
		if assert.IsType(t, new(HTTPError), err) {
			assert.Equal(t, ErrUnsupportedMediaType, err)
		}
	}
}
