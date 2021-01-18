package standard

import (
	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo/engine/test"
	"net/http"
	"testing"
)

func TestHeader(t *testing.T) {
	header := &Header{http.Header{}}
	test.HeaderTest(t, header)

	header.reset(http.Header{})
	assert.Len(t, header.Keys(), 0)
}
