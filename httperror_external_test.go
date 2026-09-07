// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

// run tests as external package to get real feel for API
package echo_test

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v5"
	"net/http"
	"net/http/httptest"
)

func ExampleDefaultHTTPErrorHandler() {
	e := echo.New()
	e.GET("/api/endpoint", func(c *echo.Context) error {
		return &apiError{
			Code: http.StatusBadRequest,
			Body: map[string]any{"message": "custom error"},
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/endpoint?err=1", nil)
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)

	fmt.Printf("%d %s", resp.Code, resp.Body.String())

	// Output: 400 {"error":{"message":"custom error"}}
}

type apiError struct {
	Code int
	Body any
}

func (e *apiError) StatusCode() int {
	return e.Code
}

func (e *apiError) MarshalJSON() ([]byte, error) {
	type body struct {
		Error any `json:"error"`
	}
	return json.Marshal(body{Error: e.Body})
}

func (e *apiError) Error() string {
	return http.StatusText(e.Code)
}
