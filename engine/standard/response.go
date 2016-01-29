package standard

import "net/http"
import "github.com/labstack/echo/engine"

type (
	Response struct {
		response  http.ResponseWriter
		header    engine.Header
		status    int
		size      int64
		committed bool
	}
)

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		response: w,
		header:   &Header{w.Header()},
	}
}

func (r *Response) Header() engine.Header {
	return r.header
}

func (r *Response) WriteHeader(code int) {
	if r.committed {
		// r.echo.Logger().Warn("response already committed")
		return
	}
	r.status = code
	r.response.WriteHeader(code)
	r.committed = true
}

func (r *Response) Write(b []byte) (n int, err error) {
	n, err = r.response.Write(b)
	r.size += int64(n)
	return
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Size() int64 {
	return r.size
}

func (r *Response) Committed() bool {
	return r.committed
}
