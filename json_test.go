package echo

import (
	testify "github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Note this test is deliberately simple as there's not a lot to test.
// Just need to ensure it writes JSONs. The heavy work is done by the context methods.
func TestDefaultJSONEncoder_JSON(t *testing.T) {
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

	enc := new(DefaultJSONEncoder)

	err := enc.JSON(user{1, "Jon Snow"}, "", c)
	if assert.NoError(err) {
		assert.Equal(userJSON+"\n", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec).(*context)
	err = enc.JSON(user{1, "Jon Snow"}, "  ", c)
	if assert.NoError(err) {
		assert.Equal(userJSONPretty+"\n", rec.Body.String())
	}
}
