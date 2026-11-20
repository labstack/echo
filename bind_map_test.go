// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindQueryParamsSupportedMapValueTypes(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?a=1&a=2&b=3", nil)

	t.Run("map[string]string", func(t *testing.T) {
		c := e.NewContext(req, httptest.NewRecorder())
		dest := map[string]string{}
		require.NoError(t, BindQueryParams(c, &dest))
		assert.Equal(t, map[string]string{"a": "1", "b": "3"}, dest)
	})

	t.Run("map[string][]string", func(t *testing.T) {
		c := e.NewContext(req, httptest.NewRecorder())
		dest := map[string][]string{}
		require.NoError(t, BindQueryParams(c, &dest))
		assert.Equal(t, map[string][]string{"a": {"1", "2"}, "b": {"3"}}, dest)
	})

	t.Run("map[string]any", func(t *testing.T) {
		c := e.NewContext(req, httptest.NewRecorder())
		dest := map[string]any{}
		require.NoError(t, BindQueryParams(c, &dest))
		assert.Equal(t, map[string]any{"a": "1", "b": "3"}, dest)
	})
}

func TestBindQueryParamsUnsupportedMapValueTypes(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/?a=1&a=2", nil)

	tests := []struct {
		name     string
		target   any
		expected any
	}{
		{
			name:     "map[string]int",
			target:   &map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "map[string][]int",
			target:   &map[string][]int{},
			expected: map[string][]int{},
		},
		{
			name:     "map[string]bool",
			target:   &map[string]bool{},
			expected: map[string]bool{},
		},
		{
			name: "nil map[string]int",
			target: func() *map[string]int {
				var dest map[string]int
				return &dest
			}(),
			expected: map[string]int(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := e.NewContext(req, httptest.NewRecorder())

			err := BindQueryParams(c, tt.target)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, StatusCode(err))
			assert.Equal(t, tt.expected, reflect.ValueOf(tt.target).Elem().Interface())
		})
	}
}

func TestBindHeadersUnsupportedMapValueTypes(t *testing.T) {
	e := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("A", "1")
	c := e.NewContext(req, httptest.NewRecorder())

	var dest map[string]int
	err := BindHeaders(c, &dest)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	assert.Nil(t, dest)
}
