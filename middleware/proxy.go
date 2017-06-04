package middleware

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/labstack/echo"
)

type (
	// ProxyConfig defines the config for Proxy middleware.
	ProxyConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Balance defines a load balancing technique.
		// Required.
		// Possible values:
		// - ProxyRandom
		// - ProxyRoundRobin
		Balancer ProxyBalancer
	}

	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		URL *url.URL
	}

	RandomBalancer struct {
		Targets []*ProxyTarget
		random  *rand.Rand
	}

	RoundRobinBalancer struct {
		Targets []*ProxyTarget
		i       uint32
	}

	ProxyBalancer interface {
		Next() *ProxyTarget
	}
)

func proxyHTTP(t *ProxyTarget) http.Handler {
	return httputil.NewSingleHostReverseProxy(t.URL)
}

func proxyRaw(t *ProxyTarget, c echo.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := w.(http.Hijacker)
		if !ok {
			c.Error(errors.New("proxy raw, not a hijacker"))
			return
		}

		in, _, err := h.Hijack()
		if err != nil {
			c.Error(fmt.Errorf("proxy raw hijack error=%v, url=%s", r.URL, err))
			return
		}
		defer in.Close()

		out, err := net.Dial("tcp", t.URL.Host)
		if err != nil {
			he := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("proxy raw dial error=%v, url=%s", r.URL, err))
			c.Error(he)
			return
		}
		defer out.Close()

		err = r.Write(out)
		if err != nil {
			he := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("proxy raw request copy error=%v, url=%s", r.URL, err))
			c.Error(he)
			return
		}

		errc := make(chan error, 2)
		cp := func(dst io.Writer, src io.Reader) {
			_, err := io.Copy(dst, src)
			errc <- err
		}

		go cp(out, in)
		go cp(in, out)
		err = <-errc
		if err != nil && err != io.EOF {
			c.Logger().Errorf("proxy raw error=%v, url=%s", r.URL, err)
		}
	})
}

func (r *RandomBalancer) Next() *ProxyTarget {
	if r.random == nil {
		r.random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	}
	return r.Targets[r.random.Intn(len(r.Targets))]
}

func (r *RoundRobinBalancer) Next() *ProxyTarget {
	r.i = r.i % uint32(len(r.Targets))
	t := r.Targets[r.i]
	atomic.AddUint32(&r.i, 1)
	return t
}

// Proxy returns an HTTP/WebSocket reverse proxy middleware.
func Proxy(config ProxyConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if config.Balancer == nil {
		panic("echo: proxy middleware requires balancer")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			t := config.Balancer.Next()

			// Proxy
			upgrade := req.Header.Get(echo.HeaderUpgrade)
			accept := req.Header.Get(echo.HeaderAccept)

			switch {
			case upgrade == "websocket" || upgrade == "Websocket":
				proxyRaw(t, c).ServeHTTP(res, req)
			case accept == "text/event-stream":
			default:
				proxyHTTP(t).ServeHTTP(res, req)
			}

			return
		}
	}
}
