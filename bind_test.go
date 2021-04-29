package echo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	bindTestStruct struct {
		I           int
		PtrI        *int
		I8          int8
		PtrI8       *int8
		I16         int16
		PtrI16      *int16
		I32         int32
		PtrI32      *int32
		I64         int64
		PtrI64      *int64
		UI          uint
		PtrUI       *uint
		UI8         uint8
		PtrUI8      *uint8
		UI16        uint16
		PtrUI16     *uint16
		UI32        uint32
		PtrUI32     *uint32
		UI64        uint64
		PtrUI64     *uint64
		B           bool
		PtrB        *bool
		F32         float32
		PtrF32      *float32
		F64         float64
		PtrF64      *float64
		S           string
		PtrS        *string
		cantSet     string
		DoesntExist string
		GoT         time.Time
		GoTptr      *time.Time
		T           Timestamp
		Tptr        *Timestamp
		SA          StringArray
	}
	bindTestStructWithTags struct {
		I           int      `json:"I" form:"I"`
		PtrI        *int     `json:"PtrI" form:"PtrI"`
		I8          int8     `json:"I8" form:"I8"`
		PtrI8       *int8    `json:"PtrI8" form:"PtrI8"`
		I16         int16    `json:"I16" form:"I16"`
		PtrI16      *int16   `json:"PtrI16" form:"PtrI16"`
		I32         int32    `json:"I32" form:"I32"`
		PtrI32      *int32   `json:"PtrI32" form:"PtrI32"`
		I64         int64    `json:"I64" form:"I64"`
		PtrI64      *int64   `json:"PtrI64" form:"PtrI64"`
		UI          uint     `json:"UI" form:"UI"`
		PtrUI       *uint    `json:"PtrUI" form:"PtrUI"`
		UI8         uint8    `json:"UI8" form:"UI8"`
		PtrUI8      *uint8   `json:"PtrUI8" form:"PtrUI8"`
		UI16        uint16   `json:"UI16" form:"UI16"`
		PtrUI16     *uint16  `json:"PtrUI16" form:"PtrUI16"`
		UI32        uint32   `json:"UI32" form:"UI32"`
		PtrUI32     *uint32  `json:"PtrUI32" form:"PtrUI32"`
		UI64        uint64   `json:"UI64" form:"UI64"`
		PtrUI64     *uint64  `json:"PtrUI64" form:"PtrUI64"`
		B           bool     `json:"B" form:"B"`
		PtrB        *bool    `json:"PtrB" form:"PtrB"`
		F32         float32  `json:"F32" form:"F32"`
		PtrF32      *float32 `json:"PtrF32" form:"PtrF32"`
		F64         float64  `json:"F64" form:"F64"`
		PtrF64      *float64 `json:"PtrF64" form:"PtrF64"`
		S           string   `json:"S" form:"S"`
		PtrS        *string  `json:"PtrS" form:"PtrS"`
		cantSet     string
		DoesntExist string      `json:"DoesntExist" form:"DoesntExist"`
		GoT         time.Time   `json:"GoT" form:"GoT"`
		GoTptr      *time.Time  `json:"GoTptr" form:"GoTptr"`
		T           Timestamp   `json:"T" form:"T"`
		Tptr        *Timestamp  `json:"Tptr" form:"Tptr"`
		SA          StringArray `json:"SA" form:"SA"`
	}
	Timestamp   time.Time
	TA          []Timestamp
	StringArray []string
	Struct      struct {
		Foo string
	}
	Bar struct {
		Baz int `json:"baz" query:"baz"`
	}
)

func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(time.RFC3339, src)
	*t = Timestamp(ts)
	return err
}

func (a *StringArray) UnmarshalParam(src string) error {
	*a = StringArray(strings.Split(src, ","))
	return nil
}

func (s *Struct) UnmarshalParam(src string) error {
	*s = Struct{
		Foo: src,
	}
	return nil
}

func (t bindTestStruct) GetCantSet() string {
	return t.cantSet
}

var values = map[string][]string{
	"I":       {"0"},
	"PtrI":    {"0"},
	"I8":      {"8"},
	"PtrI8":   {"8"},
	"I16":     {"16"},
	"PtrI16":  {"16"},
	"I32":     {"32"},
	"PtrI32":  {"32"},
	"I64":     {"64"},
	"PtrI64":  {"64"},
	"UI":      {"0"},
	"PtrUI":   {"0"},
	"UI8":     {"8"},
	"PtrUI8":  {"8"},
	"UI16":    {"16"},
	"PtrUI16": {"16"},
	"UI32":    {"32"},
	"PtrUI32": {"32"},
	"UI64":    {"64"},
	"PtrUI64": {"64"},
	"B":       {"true"},
	"PtrB":    {"true"},
	"F32":     {"32.5"},
	"PtrF32":  {"32.5"},
	"F64":     {"64.5"},
	"PtrF64":  {"64.5"},
	"S":       {"test"},
	"PtrS":    {"test"},
	"cantSet": {"test"},
	"T":       {"2016-12-06T19:09:05+01:00"},
	"Tptr":    {"2016-12-06T19:09:05+01:00"},
	"GoT":     {"2016-12-06T19:09:05+01:00"},
	"GoTptr":  {"2016-12-06T19:09:05+01:00"},
	"ST":      {"bar"},
}

func TestToMultipleFields(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?id=1&ID=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	type Root struct {
		ID     int64 `query:"id"`
		Child2 struct {
			ID int64
		}
		Child1 struct {
			ID int64 `query:"id"`
		}
	}

	u := new(Root)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), u.ID)        // perfectly reasonable
		assert.Equal(t, int64(1), u.Child1.ID) // untagged struct containing tagged field gets filled (by tag)
		assert.Equal(t, int64(0), u.Child2.ID) // untagged struct containing untagged field should not be bind
	}
}

func TestBindJSON(t *testing.T) {
	assert := assert.New(t)
	testBindOkay(assert, strings.NewReader(userJSON), nil, MIMEApplicationJSON)
	testBindOkay(assert, strings.NewReader(userJSON), dummyQuery, MIMEApplicationJSON)
	testBindArrayOkay(assert, strings.NewReader(usersJSON), nil, MIMEApplicationJSON)
	testBindArrayOkay(assert, strings.NewReader(usersJSON), dummyQuery, MIMEApplicationJSON)
	testBindError(assert, strings.NewReader(invalidContent), MIMEApplicationJSON, &json.SyntaxError{})
	testBindError(assert, strings.NewReader(userJSONInvalidType), MIMEApplicationJSON, &json.UnmarshalTypeError{})
}

func TestBindXML(t *testing.T) {
	assert := assert.New(t)

	testBindOkay(assert, strings.NewReader(userXML), nil, MIMEApplicationXML)
	testBindOkay(assert, strings.NewReader(userXML), dummyQuery, MIMEApplicationXML)
	testBindArrayOkay(assert, strings.NewReader(userXML), nil, MIMEApplicationXML)
	testBindArrayOkay(assert, strings.NewReader(userXML), dummyQuery, MIMEApplicationXML)
	testBindError(assert, strings.NewReader(invalidContent), MIMEApplicationXML, errors.New(""))
	testBindError(assert, strings.NewReader(userXMLConvertNumberError), MIMEApplicationXML, &strconv.NumError{})
	testBindError(assert, strings.NewReader(userXMLUnsupportedTypeError), MIMEApplicationXML, &xml.SyntaxError{})
	testBindOkay(assert, strings.NewReader(userXML), nil, MIMETextXML)
	testBindOkay(assert, strings.NewReader(userXML), dummyQuery, MIMETextXML)
	testBindError(assert, strings.NewReader(invalidContent), MIMETextXML, errors.New(""))
	testBindError(assert, strings.NewReader(userXMLConvertNumberError), MIMETextXML, &strconv.NumError{})
	testBindError(assert, strings.NewReader(userXMLUnsupportedTypeError), MIMETextXML, &xml.SyntaxError{})
}

func TestBindForm(t *testing.T) {
	assert := assert.New(t)

	testBindOkay(assert, strings.NewReader(userForm), nil, MIMEApplicationForm)
	testBindOkay(assert, strings.NewReader(userForm), dummyQuery, MIMEApplicationForm)
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userForm))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	req.Header.Set(HeaderContentType, MIMEApplicationForm)
	err := c.Bind(&[]struct{ Field string }{})
	assert.Error(err)
}

func TestBindQueryParams(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?id=1&name=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}
}

func TestBindQueryParamsCaseInsensitive(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?ID=1&NAME=Jon+Snow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}
}

func TestBindQueryParamsCaseSensitivePrioritized(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?id=1&ID=2&NAME=Jon+Snow&name=Jon+Doe", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Doe", u.Name)
	}
}

func TestBindUnmarshalParam(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?ts=2016-12-06T19:09:05Z&sa=one,two,three&ta=2016-12-06T19:09:05Z&ta=2016-12-06T19:09:05Z&ST=baz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		T         Timestamp   `query:"ts"`
		TA        []Timestamp `query:"ta"`
		SA        StringArray `query:"sa"`
		ST        Struct
		StWithTag struct {
			Foo string `query:"st"`
		}
	}{}
	err := c.Bind(&result)
	ts := Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC))

	assert := assert.New(t)
	if assert.NoError(err) {
		//		assert.Equal( Timestamp(reflect.TypeOf(&Timestamp{}), time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), result.T)
		assert.Equal(ts, result.T)
		assert.Equal(StringArray([]string{"one", "two", "three"}), result.SA)
		assert.Equal([]Timestamp{ts, ts}, result.TA)
		assert.Equal(Struct{""}, result.ST)       // child struct does not have a field with matching tag
		assert.Equal("baz", result.StWithTag.Foo) // child struct has field with matching tag
	}
}

func TestBindUnmarshalText(t *testing.T) {
	e := New()
	req := httptest.NewRequest(GET, "/?ts=2016-12-06T19:09:05Z&sa=one,two,three&ta=2016-12-06T19:09:05Z&ta=2016-12-06T19:09:05Z&ST=baz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		T  time.Time   `query:"ts"`
		TA []time.Time `query:"ta"`
		SA StringArray `query:"sa"`
		ST Struct
	}{}
	err := c.Bind(&result)
	ts := time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)
	if assert.NoError(t, err) {
		//		assert.Equal(t, Timestamp(reflect.TypeOf(&Timestamp{}), time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), result.T)
		assert.Equal(t, ts, result.T)
		assert.Equal(t, StringArray([]string{"one", "two", "three"}), result.SA)
		assert.Equal(t, []time.Time{ts, ts}, result.TA)
		assert.Equal(t, Struct{""}, result.ST) // field in child struct does not have tag
	}
}

func TestBindUnmarshalParamPtr(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?ts=2016-12-06T19:09:05Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		Tptr *Timestamp `query:"ts"`
	}{}
	err := c.Bind(&result)
	if assert.NoError(t, err) {
		assert.Equal(t, Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), *result.Tptr)
	}
}

func TestBindUnmarshalParamAnonymousFieldPtr(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?baz=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		*Bar
	}{&Bar{}}
	err := c.Bind(&result)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, result.Baz)
	}
}

func TestBindUnmarshalParamAnonymousFieldPtrNil(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?baz=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		*Bar
	}{}
	err := c.Bind(&result)
	if assert.NoError(t, err) {
		assert.Nil(t, result.Bar)
	}
}

func TestBindUnmarshalParamAnonymousFieldPtrCustomTag(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, `/?bar={"baz":100}&baz=1`, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		*Bar `json:"bar" query:"bar"`
	}{&Bar{}}
	err := c.Bind(&result)
	assert.Contains(t, err.Error(), "query/param/form tags are not allowed with anonymous struct field")
}

func TestBindUnmarshalTextPtr(t *testing.T) {
	e := New()
	req := httptest.NewRequest(GET, "/?ts=2016-12-06T19:09:05Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	result := struct {
		Tptr *time.Time `query:"ts"`
	}{}
	err := c.Bind(&result)
	if assert.NoError(t, err) {
		assert.Equal(t, time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC), *result.Tptr)
	}
}

func TestBindMultipartForm(t *testing.T) {
	bodyBuffer := new(bytes.Buffer)
	mw := multipart.NewWriter(bodyBuffer)
	mw.WriteField("id", "1")
	mw.WriteField("name", "Jon Snow")
	mw.Close()
	body := bodyBuffer.Bytes()

	assert := assert.New(t)
	testBindOkay(assert, bytes.NewReader(body), nil, mw.FormDataContentType())
	testBindOkay(assert, bytes.NewReader(body), dummyQuery, mw.FormDataContentType())
}

func TestBindUnsupportedMediaType(t *testing.T) {
	assert := assert.New(t)
	testBindError(assert, strings.NewReader(invalidContent), MIMEApplicationJSON, &json.SyntaxError{})
}

func TestBindbindData(t *testing.T) {
	a := assert.New(t)
	ts := new(bindTestStruct)
	b := new(DefaultBinder)
	err := b.bindData(ts, values, "form")
	a.NoError(err)

	a.Equal(0, ts.I)
	a.Equal(int8(0), ts.I8)
	a.Equal(int16(0), ts.I16)
	a.Equal(int32(0), ts.I32)
	a.Equal(int64(0), ts.I64)
	a.Equal(uint(0), ts.UI)
	a.Equal(uint8(0), ts.UI8)
	a.Equal(uint16(0), ts.UI16)
	a.Equal(uint32(0), ts.UI32)
	a.Equal(uint64(0), ts.UI64)
	a.Equal(false, ts.B)
	a.Equal(float32(0), ts.F32)
	a.Equal(float64(0), ts.F64)
	a.Equal("", ts.S)
	a.Equal("", ts.cantSet)
}

func TestBindParam(t *testing.T) {
	e := New()
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users/:id/:name")
	c.SetParamNames("id", "name")
	c.SetParamValues("1", "Jon Snow")

	u := new(user)
	err := c.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}

	// Second test for the absence of a param
	c2 := e.NewContext(req, rec)
	c2.SetPath("/users/:id")
	c2.SetParamNames("id")
	c2.SetParamValues("1")

	u = new(user)
	err = c2.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "", u.Name)
	}

	// Bind something with param and post data payload
	body := bytes.NewBufferString(`{ "name": "Jon Snow" }`)
	e2 := New()
	req2 := httptest.NewRequest(POST, "/", body)
	req2.Header.Set(HeaderContentType, MIMEApplicationJSON)

	rec2 := httptest.NewRecorder()

	c3 := e2.NewContext(req2, rec2)
	c3.SetPath("/users/:id")
	c3.SetParamNames("id")
	c3.SetParamValues("1")

	u = new(user)
	err = c3.Bind(u)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "Jon Snow", u.Name)
	}
}

func TestBindUnmarshalTypeError(t *testing.T) {
	body := bytes.NewBufferString(`{ "id": "text" }`)
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	u := new(user)

	err := c.Bind(u)

	he := &HTTPError{Code: http.StatusBadRequest, Message: "Unmarshal type error: expected=int, got=string, field=id, offset=14", Internal: err.(*HTTPError).Internal}

	assert.Equal(t, he, err)
}

func TestBindSetWithProperType(t *testing.T) {
	assert := assert.New(t)
	ts := new(bindTestStruct)
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
		assert.NoError(err)
	}
	assertBindTestStruct(assert, ts)

	type foo struct {
		Bar bytes.Buffer
	}
	v := &foo{}
	typ = reflect.TypeOf(v).Elem()
	val = reflect.ValueOf(v).Elem()
	assert.Error(setWithProperType(typ.Field(0).Type.Kind(), "5", val.Field(0)))
}

func TestBindSetFields(t *testing.T) {
	assert := assert.New(t)

	ts := new(bindTestStruct)
	val := reflect.ValueOf(ts).Elem()
	// Int
	if assert.NoError(setIntField("5", 0, val.FieldByName("I"))) {
		assert.Equal(5, ts.I)
	}
	if assert.NoError(setIntField("", 0, val.FieldByName("I"))) {
		assert.Equal(0, ts.I)
	}

	// Uint
	if assert.NoError(setUintField("10", 0, val.FieldByName("UI"))) {
		assert.Equal(uint(10), ts.UI)
	}
	if assert.NoError(setUintField("", 0, val.FieldByName("UI"))) {
		assert.Equal(uint(0), ts.UI)
	}

	// Float
	if assert.NoError(setFloatField("15.5", 0, val.FieldByName("F32"))) {
		assert.Equal(float32(15.5), ts.F32)
	}
	if assert.NoError(setFloatField("", 0, val.FieldByName("F32"))) {
		assert.Equal(float32(0.0), ts.F32)
	}

	// Bool
	if assert.NoError(setBoolField("true", val.FieldByName("B"))) {
		assert.Equal(true, ts.B)
	}
	if assert.NoError(setBoolField("", val.FieldByName("B"))) {
		assert.Equal(false, ts.B)
	}

	ok, err := unmarshalFieldNonPtr("2016-12-06T19:09:05Z", val.FieldByName("T"))
	if assert.NoError(err) {
		assert.Equal(ok, true)
		assert.Equal(Timestamp(time.Date(2016, 12, 6, 19, 9, 5, 0, time.UTC)), ts.T)
	}
}

func BenchmarkBindbindDataWithTags(b *testing.B) {
	b.ReportAllocs()
	assert := assert.New(b)
	ts := new(bindTestStructWithTags)
	binder := new(DefaultBinder)
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = binder.bindData(ts, values, "form")
	}
	assert.NoError(err)
	assertBindTestStruct(assert, (*bindTestStruct)(ts))
}

func assertBindTestStruct(a *assert.Assertions, ts *bindTestStruct) {
	a.Equal(0, ts.I)
	a.Equal(int8(8), ts.I8)
	a.Equal(int16(16), ts.I16)
	a.Equal(int32(32), ts.I32)
	a.Equal(int64(64), ts.I64)
	a.Equal(uint(0), ts.UI)
	a.Equal(uint8(8), ts.UI8)
	a.Equal(uint16(16), ts.UI16)
	a.Equal(uint32(32), ts.UI32)
	a.Equal(uint64(64), ts.UI64)
	a.Equal(true, ts.B)
	a.Equal(float32(32.5), ts.F32)
	a.Equal(float64(64.5), ts.F64)
	a.Equal("test", ts.S)
	a.Equal("", ts.GetCantSet())
}

func testBindOkay(assert *assert.Assertions, r io.Reader, query url.Values, ctype string) {
	e := New()
	path := "/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	req := httptest.NewRequest(http.MethodPost, path, r)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := c.Bind(u)
	if assert.NoError(err) {
		assert.Equal(1, u.ID)
		assert.Equal("Jon Snow", u.Name)
	}
}

func testBindArrayOkay(assert *assert.Assertions, r io.Reader, query url.Values, ctype string) {
	e := New()
	path := "/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	req := httptest.NewRequest(http.MethodPost, path, r)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := []user{}
	err := c.Bind(&u)
	if assert.NoError(err) {
		assert.Equal(1, len(u))
		assert.Equal(1, u[0].ID)
		assert.Equal("Jon Snow", u[0].Name)
	}
}

func testBindError(assert *assert.Assertions, r io.Reader, ctype string, expectedInternal error) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", r)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	req.Header.Set(HeaderContentType, ctype)
	u := new(user)
	err := c.Bind(u)

	switch {
	case strings.HasPrefix(ctype, MIMEApplicationJSON), strings.HasPrefix(ctype, MIMEApplicationXML), strings.HasPrefix(ctype, MIMETextXML),
		strings.HasPrefix(ctype, MIMEApplicationForm), strings.HasPrefix(ctype, MIMEMultipartForm):
		if assert.IsType(new(HTTPError), err) {
			assert.Equal(http.StatusBadRequest, err.(*HTTPError).Code)
			assert.IsType(expectedInternal, err.(*HTTPError).Internal)
		}
	default:
		if assert.IsType(new(HTTPError), err) {
			assert.Equal(ErrUnsupportedMediaType, err)
			assert.IsType(expectedInternal, err.(*HTTPError).Internal)
		}
	}
}

func TestDefaultBinder_BindToStructFromMixedSources(t *testing.T) {
	// tests to check binding behaviour when multiple sources path params, query params and request body are in use
	// binding is done in steps and one source could overwrite previous source binded data
	// these tests are to document this behaviour and detect further possible regressions when bind implementation is changed

	type Opts struct {
		ID   int    `json:"id" form:"id" query:"id"`
		Node string `json:"node" form:"node" query:"node" param:"node"`
		Lang string
	}

	var testCases = []struct {
		name             string
		givenURL         string
		givenContent     io.Reader
		givenMethod      string
		whenBindTarget   interface{}
		whenNoPathParams bool
		expect           interface{}
		expectError      string
	}{
		{
			name:         "ok, POST bind to struct with: path param + query param + body",
			givenMethod:  http.MethodPost,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1}`),
			expect:       &Opts{ID: 1, Node: "node_from_path"}, // query params are not used, node is filled from path
		},
		{
			name:         "ok, PUT bind to struct with: path param + query param + body",
			givenMethod:  http.MethodPut,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1}`),
			expect:       &Opts{ID: 1, Node: "node_from_path"}, // query params are not used
		},
		{
			name:         "ok, GET bind to struct with: path param + query param + body",
			givenMethod:  http.MethodGet,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1}`),
			expect:       &Opts{ID: 1, Node: "xxx"}, // query overwrites previous path value
		},
		{
			name:         "ok, GET bind to struct with: path param + query param + body",
			givenMethod:  http.MethodGet,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:       &Opts{ID: 1, Node: "zzz"}, // body is binded last and overwrites previous (path,query) values
		},
		{
			name:         "ok, DELETE bind to struct with: path param + query param + body",
			givenMethod:  http.MethodDelete,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:       &Opts{ID: 1, Node: "zzz"}, // for DELETE body is binded after query params
		},
		{
			name:         "ok, POST bind to struct with: path param + body",
			givenMethod:  http.MethodPost,
			givenURL:     "/api/real_node/endpoint",
			givenContent: strings.NewReader(`{"id": 1}`),
			expect:       &Opts{ID: 1, Node: "node_from_path"},
		},
		{
			name:         "ok, POST bind to struct with path + query + body = body has priority",
			givenMethod:  http.MethodPost,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:       &Opts{ID: 1, Node: "zzz"}, // field value from content has higher priority
		},
		{
			name:         "nok, POST body bind failure",
			givenMethod:  http.MethodPost,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`{`),
			expect:       &Opts{ID: 0, Node: "node_from_path"}, // query binding has already modified bind target
			expectError:  "code=400, message=unexpected EOF, internal=unexpected EOF",
		},
		{
			name:         "nok, GET with body bind failure when types are not convertible",
			givenMethod:  http.MethodGet,
			givenURL:     "/api/real_node/endpoint?id=nope",
			givenContent: strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:       &Opts{ID: 0, Node: "node_from_path"}, // path params binding has already modified bind target
			expectError:  "code=400, message=strconv.ParseInt: parsing \"nope\": invalid syntax, internal=strconv.ParseInt: parsing \"nope\": invalid syntax",
		},
		{
			name:         "nok, GET body bind failure - trying to bind json array to struct",
			givenMethod:  http.MethodGet,
			givenURL:     "/api/real_node/endpoint?node=xxx",
			givenContent: strings.NewReader(`[{"id": 1}]`),
			expect:       &Opts{ID: 0, Node: "xxx"}, // query binding has already modified bind target
			expectError:  "code=400, message=Unmarshal type error: expected=echo.Opts, got=array, field=, offset=1, internal=json: cannot unmarshal array into Go value of type echo.Opts",
		},
		{ // query param is ignored as we do not know where exactly to bind it in slice
			name:             "ok, GET bind to struct slice, ignore query param",
			givenMethod:      http.MethodGet,
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenContent:     strings.NewReader(`[{"id": 1}]`),
			whenNoPathParams: true,
			whenBindTarget:   &[]Opts{},
			expect: &[]Opts{
				{ID: 1, Node: ""},
			},
		},
		{ // binding query params interferes with body. b.BindBody() should be used to bind only body to slice
			name:             "ok, POST binding to slice should not be affected query params types",
			givenMethod:      http.MethodPost,
			givenURL:         "/api/real_node/endpoint?id=nope&node=xxx",
			givenContent:     strings.NewReader(`[{"id": 1}]`),
			whenNoPathParams: true,
			whenBindTarget:   &[]Opts{},
			expect:           &[]Opts{{ID: 1}},
			expectError:      "",
		},
		{ // path param is ignored as we do not know where exactly to bind it in slice
			name:           "ok, GET bind to struct slice, ignore path param",
			givenMethod:    http.MethodGet,
			givenURL:       "/api/real_node/endpoint?node=xxx",
			givenContent:   strings.NewReader(`[{"id": 1}]`),
			whenBindTarget: &[]Opts{},
			expect: &[]Opts{
				{ID: 1, Node: ""},
			},
		},
		{
			name:             "ok, GET body bind json array to slice",
			givenMethod:      http.MethodGet,
			givenURL:         "/api/real_node/endpoint",
			givenContent:     strings.NewReader(`[{"id": 1}]`),
			whenNoPathParams: true,
			whenBindTarget:   &[]Opts{},
			expect:           &[]Opts{{ID: 1, Node: ""}},
			expectError:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			// assume route we are testing is "/api/:node/endpoint?some_query_params=here"
			req := httptest.NewRequest(tc.givenMethod, tc.givenURL, tc.givenContent)
			req.Header.Set(HeaderContentType, MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if !tc.whenNoPathParams {
				c.SetParamNames("node")
				c.SetParamValues("node_from_path")
			}

			var bindTarget interface{}
			if tc.whenBindTarget != nil {
				bindTarget = tc.whenBindTarget
			} else {
				bindTarget = &Opts{}
			}
			b := new(DefaultBinder)

			err := b.Bind(bindTarget, c)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expect, bindTarget)
		})
	}
}

func TestDefaultBinder_BindBody(t *testing.T) {
	// tests to check binding behaviour when multiple sources path params, query params and request body are in use
	// generally when binding from request body - URL and path params are ignored - unless form is being binded.
	// these tests are to document this behaviour and detect further possible regressions when bind implementation is changed

	type Node struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id"`
		Node string `json:"node" xml:"node" form:"node" query:"node" param:"node"`
	}
	type Nodes struct {
		Nodes []Node `xml:"node" form:"node"`
	}

	var testCases = []struct {
		name             string
		givenURL         string
		givenContent     io.Reader
		givenMethod      string
		givenContentType string
		whenNoPathParams bool
		whenBindTarget   interface{}
		expect           interface{}
		expectError      string
	}{
		{
			name:             "ok, JSON POST bind to struct with: path + query + empty field in body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`{"id": 1}`),
			expect:           &Node{ID: 1, Node: ""}, // path params or query params should not interfere with body
		},
		{
			name:             "ok, JSON POST bind to struct with: path + query + body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:           &Node{ID: 1, Node: "zzz"}, // field value from content has higher priority
		},
		{
			name:             "ok, JSON POST body bind json array to slice (has matching path/query params)",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`[{"id": 1}]`),
			whenNoPathParams: true,
			whenBindTarget:   &[]Node{},
			expect:           &[]Node{{ID: 1, Node: ""}},
			expectError:      "",
		},
		{ // rare case as GET is not usually used to send request body
			name:             "ok, JSON GET bind to struct with: path + query + empty field in body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodGet,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`{"id": 1}`),
			expect:           &Node{ID: 1, Node: ""}, // path params or query params should not interfere with body
		},
		{ // rare case as GET is not usually used to send request body
			name:             "ok, JSON GET bind to struct with: path + query + body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodGet,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`{"id": 1, "node": "zzz"}`),
			expect:           &Node{ID: 1, Node: "zzz"}, // field value from content has higher priority
		},
		{
			name:             "nok, JSON POST body bind failure",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationJSON,
			givenContent:     strings.NewReader(`{`),
			expect:           &Node{ID: 0, Node: ""},
			expectError:      "code=400, message=unexpected EOF, internal=unexpected EOF",
		},
		{
			name:             "ok, XML POST bind to struct with: path + query + empty body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationXML,
			givenContent:     strings.NewReader(`<node><id>1</id><node>yyy</node></node>`),
			expect:           &Node{ID: 1, Node: "yyy"},
		},
		{
			name:             "ok, XML POST bind array to slice with: path + query + body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationXML,
			givenContent:     strings.NewReader(`<nodes><node><id>1</id><node>yyy</node></node></nodes>`),
			whenBindTarget:   &Nodes{},
			expect:           &Nodes{Nodes: []Node{{ID: 1, Node: "yyy"}}},
		},
		{
			name:             "nok, XML POST bind failure",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationXML,
			givenContent:     strings.NewReader(`<node><`),
			expect:           &Node{ID: 0, Node: ""},
			expectError:      "code=400, message=Syntax error: line=1, error=XML syntax error on line 1: unexpected EOF, internal=XML syntax error on line 1: unexpected EOF",
		},
		{
			name:             "ok, FORM POST bind to struct with: path + query + body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationForm,
			givenContent:     strings.NewReader(`id=1&node=yyy`),
			expect:           &Node{ID: 1, Node: "yyy"},
		},
		{
			// NB: form values are taken from BOTH body and query for POST/PUT/PATCH by standard library implementation
			// See: https://golang.org/pkg/net/http/#Request.ParseForm
			name:             "ok, FORM POST bind to struct with: path + query + empty field in body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMEApplicationForm,
			givenContent:     strings.NewReader(`id=1`),
			expect:           &Node{ID: 1, Node: "xxx"},
		},
		{
			// NB: form values are taken from query by standard library implementation
			// See: https://golang.org/pkg/net/http/#Request.ParseForm
			name:             "ok, FORM GET bind to struct with: path + query + empty field in body",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodGet,
			givenContentType: MIMEApplicationForm,
			givenContent:     strings.NewReader(`id=1`),
			expect:           &Node{ID: 0, Node: "xxx"}, // 'xxx' is taken from URL and body is not used with GET by implementation
		},
		{
			name:             "nok, unsupported content type",
			givenURL:         "/api/real_node/endpoint?node=xxx",
			givenMethod:      http.MethodPost,
			givenContentType: MIMETextPlain,
			givenContent:     strings.NewReader(`<html></html>`),
			expect:           &Node{ID: 0, Node: ""},
			expectError:      "code=415, message=Unsupported Media Type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			// assume route we are testing is "/api/:node/endpoint?some_query_params=here"
			req := httptest.NewRequest(tc.givenMethod, tc.givenURL, tc.givenContent)
			switch tc.givenContentType {
			case MIMEApplicationXML:
				req.Header.Set(HeaderContentType, MIMEApplicationXML)
			case MIMEApplicationForm:
				req.Header.Set(HeaderContentType, MIMEApplicationForm)
			case MIMEApplicationJSON:
				req.Header.Set(HeaderContentType, MIMEApplicationJSON)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if !tc.whenNoPathParams {
				c.SetParamNames("node")
				c.SetParamValues("real_node")
			}

			var bindTarget interface{}
			if tc.whenBindTarget != nil {
				bindTarget = tc.whenBindTarget
			} else {
				bindTarget = &Node{}
			}
			b := new(DefaultBinder)

			err := b.BindBody(c, bindTarget)
			if tc.expectError != "" {
				assert.EqualError(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expect, bindTarget)
		})
	}
}
