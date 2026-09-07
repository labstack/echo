package echotest_test

import (
	"strings"
	"testing"

	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"
)

const testJSONContent = `{
  "field": "value"
}`

func TestLoadBytesOK(t *testing.T) {
	data := echotest.LoadBytes(t, "testdata/test.json")
	assert.Equal(t, []byte(testJSONContent+"\n"), data)
}

func TestLoadBytes_custom(t *testing.T) {
	data := echotest.LoadBytes(t, "testdata/test.json", func(bytes []byte) []byte {
		return []byte(strings.ToUpper(string(bytes)))
	})
	assert.Equal(t, []byte(strings.ToUpper(testJSONContent)+"\n"), data)
}
