package echotest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testJSONContent = `{
  "field": "value"
}`

func TestLoadBytesOK(t *testing.T) {
	data := LoadBytes(t, "testdata/test.json")
	assert.Equal(t, []byte(testJSONContent+"\n"), data)
}

func TestLoadBytesOK_TrimNewlineEnd(t *testing.T) {
	data := LoadBytes(t, "testdata/test.json", TrimNewlineEnd)
	assert.Equal(t, []byte(testJSONContent), data)
}
