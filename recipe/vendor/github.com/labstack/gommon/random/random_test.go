package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	assert.Len(t, String(32), 32)
	r := New()
	r.SetCharset(Numeric)
	assert.Len(t, r.String(8), 8)
}
