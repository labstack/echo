package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/standard"
)

func request(c echo.Context) error {
	req := c.Request().(*standard.Request).Request
	format := "<pre><strong>Request Information</strong>\n\n<code>Protocol: %s\nHost: %s\nRemote Address: %s\nMethod: %s\nPath: %s\n</code></pre>"
	return c.HTML(http.StatusOK, fmt.Sprintf(format, req.Proto, req.Host, req.RemoteAddr, req.Method, req.URL.Path))
}

func stream(c echo.Context) error {
	res := c.Response().(*standard.Response).ResponseWriter
	gone := res.(http.CloseNotifier).CloseNotify()
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	res.WriteHeader(http.StatusOK)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Fprint(res, "<pre><strong>Clock Stream</strong>\n\n<code>")
	for {
		fmt.Fprintf(res, "%v\n", time.Now())
		res.(http.Flusher).Flush()
		select {
		case <-ticker.C:
		case <-gone:
			break
		}
	}
}

func main() {
	e := echo.New()
	e.GET("/request", request)
	e.GET("/stream", stream)
	e.Run(standard.WithConfig(engine.Config{
		Address:     ":1323",
		TLSCertFile: "cert.pem",
		TLSKeyFile:  "key.pem",
	}))
}
