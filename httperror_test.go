package echo

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestHTTPError(t *testing.T) {
	t.Run("non-internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})

		assert.Equal(t, "code=400, message=map[code:12]", err.Error())
	})
	t.Run("internal", func(t *testing.T) {
		err := NewHTTPError(http.StatusBadRequest, map[string]interface{}{
			"code": 12,
		})
		err = err.WithInternal(errors.New("internal error"))
		assert.Equal(t, "code=400, message=map[code:12], internal=internal error", err.Error())
	})
}

func TestNewHTTPErrorWithInternal(t *testing.T) {
	he := NewHTTPErrorWithInternal(http.StatusBadRequest, errors.New("test"), "test message")
	assert.Equal(t, "code=400, message=test message, internal=test", he.Error())
}

func TestNewHTTPErrorWithInternal_noCustomMessage(t *testing.T) {
	he := NewHTTPErrorWithInternal(http.StatusBadRequest, errors.New("test"))
	assert.Equal(t, "code=400, message=Bad Request, internal=test", he.Error())
}

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
		err = err.WithInternal(errors.New("internal error"))
		assert.Equal(t, "internal error", errors.Unwrap(err).Error())
	})
}
