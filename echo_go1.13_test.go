// +build go1.13

package echo

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_Unwrap(t *testing.T) {
	t.Run("non-internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})

		assert.Nil(t, errors.Unwrap(err))
	})
	t.Run("internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})
		err.SetInternal(errors.New("internal error"))
		assert.Equal(t, "internal error", errors.Unwrap(err).Error())
	})
}
