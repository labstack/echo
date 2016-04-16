package echo

import (
	"mime/multipart"

	"github.com/labstack/echo/engine"
)

// RequestReader implements helpers around reading requests.
type RequestReader struct {
	Req engine.Request
}

// Request gets the underlying engine.Request
func (r RequestReader) Request() engine.Request {
	return r.Req
}

// QueryParam returns the query param for the provided name. It is an alias
// for `engine.URL#QueryParam()`.
func (r RequestReader) QueryParam(name string) string {
	return r.Req.URL().QueryParam(name)
}

// QueryParams returns the query parameters as map. It is an alias for `engine.URL#QueryParams()`.
func (r RequestReader) QueryParams() map[string][]string {
	return r.Req.URL().QueryParams()
}

// FormValue returns the form field value for the provided name. It is an
// alias for `engine.Request#FormValue()`.
func (r RequestReader) FormValue(name string) string {
	return r.Req.FormValue(name)
}

// FormParams returns the form parameters as map. It is an alias for `engine.Request#FormParams()`.
func (r RequestReader) FormParams() map[string][]string {
	return r.Req.FormParams()
}

// FormFile returns the multipart form file for the provided name. It is an
// alias for `engine.Request#FormFile()`.
func (r RequestReader) FormFile(name string) (*multipart.FileHeader, error) {
	return r.Req.FormFile(name)
}

// MultipartForm returns the multipart form. It is an alias for `engine.Request#MultipartForm()`.
func (r RequestReader) MultipartForm() (*multipart.Form, error) {
	return r.Req.MultipartForm()
}
