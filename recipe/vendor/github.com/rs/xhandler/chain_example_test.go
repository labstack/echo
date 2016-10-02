package xhandler_test

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
)

func ExampleChain() {
	c := xhandler.Chain{}
	// Append a context-aware middleware handler
	c.UseC(xhandler.CloseHandler)

	// Mix it with a non-context-aware middleware handler
	c.Use(cors.Default().Handler)

	// Another context-aware middleware handler
	c.UseC(xhandler.TimeoutHandler(2 * time.Second))

	mux := http.NewServeMux()

	// Use c.Handler to terminate the chain with your final handler
	mux.Handle("/", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))

	// You can reuse the same chain for other handlers
	mux.Handle("/api", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the API!")
	})))
}

func ExampleAddChain() {
	c := xhandler.Chain{}

	close := xhandler.CloseHandler
	cors := cors.Default().Handler
	timeout := xhandler.TimeoutHandler(2 * time.Second)
	auth := func(next xhandler.HandlerC) xhandler.HandlerC {
		return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if v := ctx.Value("Authorization"); v == nil {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTPC(ctx, w, r)
		})
	}

	c.Add(close, cors, timeout)

	mux := http.NewServeMux()

	// Use c.Handler to terminate the chain with your final handler
	mux.Handle("/", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))

	// Create a new chain from an existing one, and add route-specific middleware to it
	protected := c.With(auth)

	mux.Handle("/admin", protected.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "protected endpoint!")
	})))
}

func ExampleIf() {
	c := xhandler.Chain{}

	// Add a timeout handler only if the URL path matches a prefix
	c.UseC(xhandler.If(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return strings.HasPrefix(r.URL.Path, "/with-timeout/")
		},
		xhandler.TimeoutHandler(2*time.Second),
	))

	http.Handle("/", c.Handler(xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Welcome to the home page!")
	})))
}
