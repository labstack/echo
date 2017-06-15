package middleware

import (
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

// TODO: Handle TLS proxy

type (
	// ProxyConfig defines the config for Proxy middleware.
	ProxyConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Balancer defines a load balancing technique.
		// Required.
		// Possible values:
		// - RandomBalancer
		// - RoundRobinBalancer
		Balancer ProxyBalancer
	}

	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		URL *url.URL
	}

	// RandomBalancer implements a random load balancing technique.
	RandomBalancer struct {
		Targets []*ProxyTarget
		random  *rand.Rand
	}

	// RoundRobinBalancer implements a round-robin load balancing technique.
	RoundRobinBalancer struct {
		Targets []*ProxyTarget
		i       uint32
	}

	// ProxyBalancer defines an interface to implement a load balancing technique.
	ProxyBalancer interface {
		Next() *ProxyTarget
	}
)

var (
	// DefaultProxyConfig is the default Proxy middleware config.
	DefaultProxyConfig = ProxyConfig{
		Skipper: DefaultSkipper,
	}
)

func proxyHTTP(t *ProxyTarget) http.Handler {
	return httputil.NewSingleHostReverseProxy(t.URL)
}

func proxyRaw(t *ProxyTarget, c echo.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		in, _, err := c.Response().Hijack()
		if err != nil {
			c.Error(fmt.Errorf("proxy raw, hijack error=%v, url=%s", t.URL, err))
			return
		}
		defer in.Close()

		out, err := net.Dial("tcp", t.URL.Host)
		if err != nil {
			he := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("proxy raw, dial error=%v, url=%s", t.URL, err))
			c.Error(he)
			return
		}
		defer out.Close()

		// Write header
		err = r.Write(out)
		if err != nil {
			he := echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("proxy raw, request header copy error=%v, url=%s", t.URL, err))
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
			c.Logger().Errorf("proxy raw, copy body error=%v, url=%s", t.URL, err)
		}
	})
}

// Next randomly returns an upstream target.
func (r *RandomBalancer) Next() *ProxyTarget {
	if r.random == nil {
		r.random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	}
	return r.Targets[r.random.Intn(len(r.Targets))]
}

// Next returns an upstream target using round-robin technique.
func (r *RoundRobinBalancer) Next() *ProxyTarget {
	r.i = r.i % uint32(len(r.Targets))
	t := r.Targets[r.i]
	atomic.AddUint32(&r.i, 1)
	return t
}

// Proxy returns a Proxy middleware.
//
// Proxy middleware forwards the request to upstream server using a configured load balancing technique.
func Proxy(balancer ProxyBalancer) echo.MiddlewareFunc {
	c := DefaultProxyConfig
	c.Balancer = balancer
	return ProxyWithConfig(c)
}

// ProxyWithConfig returns a Proxy middleware with config.
// See: `Proxy()`
func ProxyWithConfig(config ProxyConfig) echo.MiddlewareFunc {
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
			tgt := config.Balancer.Next()

			// Fix header
			if req.Header.Get(echo.HeaderXRealIP) == "" {
				req.Header.Set(echo.HeaderXRealIP, c.RealIP())
			}
			if req.Header.Get(echo.HeaderXForwardedProto) == "" {
				req.Header.Set(echo.HeaderXForwardedProto, c.Scheme())
			}
			if c.IsWebSocket() && req.Header.Get(echo.HeaderXForwardedFor) == "" { // For HTTP, it is automatically set by Go HTTP reverse proxy.
				req.Header.Set(echo.HeaderXForwardedFor, c.RealIP())
			}

			// Proxy
			switch {
			case c.IsWebSocket():
				proxyRaw(tgt, c).ServeHTTP(res, req)
			case req.Header.Get(echo.HeaderAccept) == "text/event-stream":
			default:
				proxyHTTP(tgt).ServeHTTP(res, req)
			}

			return
		}
	}
}
