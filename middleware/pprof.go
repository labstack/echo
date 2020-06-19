package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http/pprof"
	"strings"
)

// EnablePProf adds several routes from package `net/http/pprof` to *echo.Echo object.
func EnablePProf(e *echo.Echo, apiPrefix string, middlewareFunc echo.MiddlewareFunc) {
	if !strings.HasPrefix(apiPrefix, "/") {
		apiPrefix = "/" + apiPrefix
	}
	if strings.HasSuffix(apiPrefix, "/") {
		apiPrefix = apiPrefix[:len(apiPrefix)-1]
	}
	pprofAPIGrp := e.Group(apiPrefix + "/pprof")

	// add middleware... maybe for authentication
	if middlewareFunc != nil {
		pprofAPIGrp.Use(middlewareFunc)
	}

	registerPProfAPI(pprofAPIGrp)
}

func registerPProfAPI(pprofAPIGrp *echo.Group) {
	pprofAPIGrp.GET("/", pprofIndexHandler())
	pprofAPIGrp.GET("/heap", pprofHeapHandler())
	pprofAPIGrp.GET("/block", pprofBlockHandler())
	pprofAPIGrp.GET("/trace", pprofTraceHandler())
	pprofAPIGrp.GET("/mutex", pprofMutexHandler())
	pprofAPIGrp.GET("/symbol", pprofSymbolHandler())
	pprofAPIGrp.POST("/symbol", pprofSymbolHandler())
	pprofAPIGrp.GET("/cmdline", pprofCmdlineHandler())
	pprofAPIGrp.GET("/profile", pprofProfileHandler())
	pprofAPIGrp.GET("/goroutine", pprofGoroutineHandler())
	pprofAPIGrp.GET("/threadcreate", pprofThreadCreateHandler())
}

// pprofIndexHandler will pass the call from /debug/pprof to pprof.
func pprofIndexHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Index(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofHeapHandler will pass the call from /debug/pprof/heap to pprof.
func pprofHeapHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("heap").ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

// pprofGoroutineHandler will pass the call from /debug/pprof/goroutine to pprof.
func pprofGoroutineHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("goroutine").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofBlockHandler will pass the call from /debug/pprof/block to pprof.
func pprofBlockHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("block").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofThreadCreateHandler will pass the call from /debug/pprof/threadcreate to pprof.
func pprofThreadCreateHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("threadcreate").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofCmdlineHandler will pass the call from /debug/pprof/cmdline to pprof.
func pprofCmdlineHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Cmdline(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofProfileHandler will pass the call from /debug/pprof/profile to pprof.
func pprofProfileHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Profile(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofSymbolHandler will pass the call from /debug/pprof/symbol to pprof.
func pprofSymbolHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Symbol(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofTraceHandler will pass the call from /debug/pprof/trace to pprof.
func pprofTraceHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Trace(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// pprofMutexHandler will pass the call from /debug/pprof/mutex to pprof.
func pprofMutexHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("mutex").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}
