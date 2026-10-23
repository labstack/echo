// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build go1.27

package echo

// ParsePathParam returns the path parameter by name parsed as type T.
// It returns ErrNonExistentKey if the parameter does not exist.
// See PathParam for supported types, options, and parsing behavior.
func (c *Context) ParsePathParam[T any](paramName string, opts ...any) (T, error) {
	return PathParam[T](c, paramName, opts...)
}

// ParsePathParamOr returns the path parameter by name parsed as type T, or defaultValue if the
// parameter does not exist or is empty.
// See PathParamOr for supported types, options, and parsing behavior.
func (c *Context) ParsePathParamOr[T any](paramName string, defaultValue T, opts ...any) (T, error) {
	return PathParamOr[T](c, paramName, defaultValue, opts...)
}

// ParseQueryParam returns the first query parameter value for key parsed as type T.
// It returns ErrNonExistentKey if the parameter does not exist.
// See QueryParam for supported types, options, and parsing behavior.
func (c *Context) ParseQueryParam[T any](key string, opts ...any) (T, error) {
	return QueryParam[T](c, key, opts...)
}

// ParseQueryParamOr returns the first query parameter value for key parsed as type T, or
// defaultValue if the parameter does not exist or is empty.
// See QueryParamOr for supported types, options, and parsing behavior.
func (c *Context) ParseQueryParamOr[T any](key string, defaultValue T, opts ...any) (T, error) {
	return QueryParamOr[T](c, key, defaultValue, opts...)
}

// ParseQueryParams returns all query parameter values for key parsed as a slice of T.
// It returns ErrNonExistentKey if the parameter does not exist.
// See QueryParams for supported types, options, and parsing behavior.
func (c *Context) ParseQueryParams[T any](key string, opts ...any) ([]T, error) {
	return QueryParams[T](c, key, opts...)
}

// ParseQueryParamsOr returns all query parameter values for key parsed as a slice of T, or
// defaultValue if the parameter does not exist.
// See QueryParamsOr for supported types, options, and parsing behavior.
func (c *Context) ParseQueryParamsOr[T any](key string, defaultValue []T, opts ...any) ([]T, error) {
	return QueryParamsOr[T](c, key, defaultValue, opts...)
}

// ParseFormValue returns the first form field value for key parsed as type T.
// It returns ErrNonExistentKey if the field does not exist.
// See FormValue for supported types, options, and parsing behavior.
func (c *Context) ParseFormValue[T any](key string, opts ...any) (T, error) {
	return FormValue[T](c, key, opts...)
}

// ParseFormValueOr returns the first form field value for key parsed as type T, or defaultValue if
// the field does not exist or is empty.
// See FormValueOr for supported types, options, and parsing behavior.
func (c *Context) ParseFormValueOr[T any](key string, defaultValue T, opts ...any) (T, error) {
	return FormValueOr[T](c, key, defaultValue, opts...)
}

// ParseFormValues returns all form field values for key parsed as a slice of T.
// It returns ErrNonExistentKey if the field does not exist.
// See FormValues for supported types, options, and parsing behavior.
func (c *Context) ParseFormValues[T any](key string, opts ...any) ([]T, error) {
	return FormValues[T](c, key, opts...)
}

// ParseFormValuesOr returns all form field values for key parsed as a slice of T, or defaultValue if
// the field does not exist.
// See FormValuesOr for supported types, options, and parsing behavior.
func (c *Context) ParseFormValuesOr[T any](key string, defaultValue []T, opts ...any) ([]T, error) {
	return FormValuesOr[T](c, key, defaultValue, opts...)
}
