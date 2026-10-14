// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build go1.27

package echo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextValueOK(t *testing.T) {
	c := NewContext(nil, nil)

	c.Set("key", int64(123))

	v, err := c.Value[int64]("key")
	assert.NoError(t, err)
	assert.Equal(t, int64(123), v)
}

func TestContextValueNonExistentKey(t *testing.T) {
	c := NewContext(nil, nil)

	v, err := c.Value[int64]("nope")
	assert.ErrorIs(t, err, ErrNonExistentKey)
	assert.Equal(t, int64(0), v)
}

func TestContextValueInvalidCast(t *testing.T) {
	c := NewContext(nil, nil)

	c.Set("key", int64(123))

	v, err := c.Value[bool]("key")
	assert.ErrorIs(t, err, ErrInvalidKeyType)
	assert.False(t, v)
}

func TestContextValueStoredNilHasInvalidType(t *testing.T) {
	c := NewContext(nil, nil)

	c.Set("key", nil)

	v, err := c.Value[any]("key")
	assert.ErrorIs(t, err, ErrInvalidKeyType)
	assert.Nil(t, v)
}

func TestContextValueOrOK(t *testing.T) {
	c := NewContext(nil, nil)

	c.Set("key", int64(123))

	v, err := c.ValueOr[int64]("key", 999)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), v)
}

func TestContextValueOrNonExistentKey(t *testing.T) {
	c := NewContext(nil, nil)

	v, err := c.ValueOr[int64]("nope", 999)
	assert.NoError(t, err)
	assert.Equal(t, int64(999), v)
}

func TestContextValueOrInvalidCast(t *testing.T) {
	c := NewContext(nil, nil)

	c.Set("key", int64(123))

	v, err := c.ValueOr[float32]("key", float32(999))
	assert.ErrorIs(t, err, ErrInvalidKeyType)
	assert.Equal(t, float32(0), v)
}
