package xhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAppendHandlerC(t *testing.T) {
	init := 0
	h1 := func(next HandlerC) HandlerC {
		init++
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 2)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	c := Chain{}
	c.UseC(h1)
	c.UseC(h2)
	assert.Len(t, c, 2)

	h := c.Handler(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 2, ctx.Value("test"), "second handler should overwrite first handler's context value")
	}))

	h.ServeHTTP(nil, nil)
	h.ServeHTTP(nil, nil)
	assert.Equal(t, 1, init, "handler init called once")
}

func TestAppendHandler(t *testing.T) {
	init := 0
	h1 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next http.Handler) http.Handler {
		init++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Change r and w values
			w = httptest.NewRecorder()
			r = &http.Request{}
			next.ServeHTTP(w, r)
		})
	}
	c := Chain{}
	c.UseC(h1)
	c.Use(h2)
	assert.Len(t, c, 2)

	h := c.Handler(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 1, ctx.Value("test"),
			"the first handler value should be pass through the second (non-aware) one")
		// Test r and w overwrite
		assert.NotNil(t, w)
		assert.NotNil(t, r)
	}))

	h.ServeHTTP(nil, nil)
	h.ServeHTTP(nil, nil)
	// There's no safe way to not initialize non ctx aware handlers on each request :/
	//assert.Equal(t, 1, init, "handler init called once")
}

func TestChainHandlerC(t *testing.T) {
	handlerCalls := 0
	h1 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 2)
			next.ServeHTTPC(ctx, w, r)
		})
	}

	c := Chain{}
	c.UseC(h1)
	c.UseC(h2)
	h := c.HandlerC(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handlerCalls++

		assert.Equal(t, 2, ctx.Value("test"),
			"second handler should overwrite first handler's context value")
		assert.Equal(t, 1, ctx.Value("mainCtx"),
			"the mainCtx value should be pass through")
	}))

	mainCtx := context.WithValue(context.Background(), "mainCtx", 1)
	h.ServeHTTPC(mainCtx, nil, nil)

	assert.Equal(t, 3, handlerCalls, "all handler called once")
}

func TestAdd(t *testing.T) {
	handlerCalls := 0
	h1 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next http.Handler) http.Handler {
		handlerCalls++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Change r and w values
			w = httptest.NewRecorder()
			r = &http.Request{}
			next.ServeHTTP(w, r)
		})
	}
	h3 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 2)
			next.ServeHTTPC(ctx, w, r)
		})
	}

	c := Chain{}
	c.Add(h1, h2, h3)
	h := c.HandlerC(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handlerCalls++

		assert.Equal(t, 2, ctx.Value("test"),
			"third handler should overwrite first handler's context value")
		assert.Equal(t, 1, ctx.Value("mainCtx"),
			"the mainCtx value should be pass through")
	}))

	mainCtx := context.WithValue(context.Background(), "mainCtx", 1)
	h.ServeHTTPC(mainCtx, nil, nil)
	assert.Equal(t, 4, handlerCalls, "all handler called once")
}

func TestWith(t *testing.T) {
	handlerCalls := 0
	h1 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next http.Handler) http.Handler {
		handlerCalls++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Change r and w values
			w = httptest.NewRecorder()
			r = &http.Request{}
			next.ServeHTTP(w, r)
		})
	}
	h3 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			handlerCalls++
			ctx = context.WithValue(ctx, "test", 2)
			next.ServeHTTPC(ctx, w, r)
		})
	}

	c := Chain{}
	c.Add(h1)
	d := c.With(h2, h3)

	h := c.HandlerC(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handlerCalls++

		assert.Equal(t, 1, ctx.Value("test"),
			"third handler should not overwrite the first handler's context value")
		assert.Equal(t, 1, ctx.Value("mainCtx"),
			"the mainCtx value should be pass through")
	}))
	i := d.HandlerC(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		handlerCalls++

		assert.Equal(t, 2, ctx.Value("test"),
			"third handler should overwrite first handler's context value")
		assert.Equal(t, 1, ctx.Value("mainCtx"),
			"the mainCtx value should be pass through")
	}))

	mainCtx := context.WithValue(context.Background(), "mainCtx", 1)
	h.ServeHTTPC(mainCtx, nil, nil)
	assert.Equal(t, 2, handlerCalls, "all handlers called once")
	handlerCalls = 0
	i.ServeHTTPC(mainCtx, nil, nil)
	assert.Equal(t, 4, handlerCalls, "all handler called once")
}
