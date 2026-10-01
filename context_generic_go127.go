// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build go1.27

package echo

// Value retrieves a value from the context store or ErrNonExistentKey error if the key is missing.
// Returns ErrInvalidKeyType error if the value is not castable to type T.
func (c *Context) Value[T any](key string) (T, error) {
	return contextValue[T](c, key)
}

// ValueOr retrieves a value from the context store or returns a default value when the key
// is missing. Returns ErrInvalidKeyType error if the value is not castable to type T.
func (c *Context) ValueOr[T any](key string, defaultValue T) (T, error) {
	return contextValueOr(c, key, defaultValue)
}
