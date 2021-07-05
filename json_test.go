package echo

import (
	testify "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Encode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*context)

	assert := testify.New(t)

	// Echo
	assert.Equal(e, c.Echo())

	// Request
	assert.NotNil(c.Request())

	// Response
	assert.NotNil(c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	err := enc.Serialize(c, user{1, "Jon Snow"}, "")
	if assert.NoError(err) {
		assert.Equal(userJSON+"\n", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*context)
	err = enc.Serialize(c, user{1, "Jon Snow"}, "  ")
	if assert.NoError(err) {
		assert.Equal(userJSONPretty+"\n", rec.Body.String())
	}
}

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONCodec_Decode(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec).(*context)

	assert := testify.New(t)

	// Echo
	assert.Equal(e, c.Echo())

	// Request
	assert.NotNil(c.Request())

	// Response
	assert.NotNil(c.Response())

	//--------
	// Default JSON encoder
	//--------

	enc := new(DefaultJSONSerializer)

	var u = user{}
	err := enc.Deserialize(c, &u)
	if assert.NoError(err) {
		assert.Equal(u, user{ID: 1, Name: "Jon Snow"})
	}

	var userUnmarshalSyntaxError = user{}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(invalidContent))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*context)
	err = enc.Deserialize(c, &userUnmarshalSyntaxError)
	assert.IsType(&HTTPError{}, err)
	assert.EqualError(err, "code=400, message=Syntax error: offset=1, error=invalid character 'i' looking for beginning of value, internal=invalid character 'i' looking for beginning of value")

	var userUnmarshalTypeError = struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(userJSON))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*context)
	err = enc.Deserialize(c, &userUnmarshalTypeError)
	assert.IsType(&HTTPError{}, err)
	assert.EqualError(err, "code=400, message=Unmarshal type error: expected=string, got=number, field=id, offset=7, internal=json: cannot unmarshal number into Go struct field .id of type string")

}
