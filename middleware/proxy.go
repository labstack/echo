package middleware

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

type (
	// ProxyConfig defines the config for Proxy middleware.
	ProxyConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Load balancing technique.
		// Optional. Default value "random".
		// Possible values:
		// - "random"
		// - "round-robin"
		Balance string `json:"balance"`

		// Upstream target URLs
		// Required.
		Targets []*ProxyTarget `json:"targets"`

		balancer proxyBalancer
	}

	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		Name string `json:"name,omitempty"`
		URL  string `json:"url"`
		url  *url.URL
	}

	proxyRandom struct {
		targets []*ProxyTarget
		random  *rand.Rand
	}

	proxyRoundRobin struct {
		targets []*ProxyTarget
		i       int32
	}

	proxyBalancer interface {
		Next() *ProxyTarget
		Length() int
	}
)

func proxyHTTP(u *url.URL, c echo.Context) http.Handler {
	return httputil.NewSingleHostReverseProxy(u)
}

func proxyWS(u *url.URL, c echo.Context) http.Handler {
	return websocket.Handler(func(in *websocket.Conn) {
		defer in.Close()

		r := in.Request()
		t := "ws://" + u.Host + r.RequestURI
		out, err := websocket.Dial(t, "", r.Header.Get("Origin"))
		if err != nil {
			c.Logger().Errorf("ws proxy error, target=%s, err=%v", t, err)
			return
		}
		defer out.Close()

		errc := make(chan error, 2)
		cp := func(w io.Writer, r io.Reader) {
			_, err := io.Copy(w, r)
			errc <- err
		}

		go cp(in, out)
		go cp(out, in)
		err = <-errc
		if err != nil && err != io.EOF {
			c.Logger().Errorf("ws proxy error, url=%s, err=%v", r.URL, err)
		}
	})
}

func (r *proxyRandom) Next() *ProxyTarget {
	return r.targets[r.random.Intn(len(r.targets))]
}

func (r *proxyRandom) Length() int {
	return len(r.targets)
}

func (r *proxyRoundRobin) Next() *ProxyTarget {
	r.i = r.i % int32(len(r.targets))
	atomic.AddInt32(&r.i, 1)
	return r.targets[r.i]
}

func (r *proxyRoundRobin) Length() int {
	return len(r.targets)
}

// Proxy returns an HTTP/WebSocket reverse proxy middleware.
func Proxy(config ProxyConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if config.Targets == nil || len(config.Targets) == 0 {
		panic("echo: proxy middleware requires targets")
	}

	// Initialize
	for _, t := range config.Targets {
		u, err := url.Parse(t.URL)
		if err != nil {
			panic("echo: proxy target url parsing failed" + err.Error())
		}
		t.url = u
	}

	// Balancer
	switch config.Balance {
	case "round-robin":
		config.balancer = &proxyRoundRobin{
			targets: config.Targets,
			i:       -1,
		}
	default: // random
		config.balancer = &proxyRandom{
			targets: config.Targets,
			random:  rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			t := config.balancer.Next().url

			// Tell upstream that the incoming request is HTTPS
			if c.IsTLS() {
				req.Header.Set(echo.HeaderXForwardedProto, "https")
			}

			// Proxy
			if req.Header.Get(echo.HeaderUpgrade) == "websocket" {
				proxyWS(t, c).ServeHTTP(res, req)
			} else {
				proxyHTTP(t, c).ServeHTTP(res, req)
			}

			return
		}
	}
}
