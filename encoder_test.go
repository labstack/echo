package echo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJson(t *testing.T) {
	enc := new(jsonEncoder)
	if encoded, err := enc.Encode(&user{1, "Jon Snow"}); assert.NoError(t, err) {
		assert.Equal(t, userJSON, string(encoded))
	}
}

func TestXml(t *testing.T) {
	enc := new(xmlEncoder)
	if encoded, err := enc.Encode(&user{1, "Jon Snow"}); assert.NoError(t, err) {
		assert.Equal(t, userXML, string(encoded))
	}
}

