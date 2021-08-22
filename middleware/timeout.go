package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------------------------------------------
// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
// WARNING: Timeout middleware causes more problems than it solves.
// WARNING: This middleware should be first middleware as it messes with request Writer and could cause data race if
// 					it is in other position
//
// Depending on out requirements you could be better of setting timeout to context and
// check its deadline from handler.
//
// For example: create middleware to set timeout to context
// func RequestTimeout(timeout time.Duration) echo.MiddlewareFunc {
//	return func(next echo.HandlerFunc) echo.HandlerFunc {
//		return func(c echo.Context) error {
//			timeoutCtx, cancel := context.WithTimeout(c.Request().Context(), timeout)
//			c.SetRequest(c.Request().WithContext(timeoutCtx))
//			defer cancel()
//			return next(c)
//		}
//	}
//}
//
// Create handler that checks for context deadline and runs actual task in separate coroutine
// Note: separate coroutine may not be even if you do not want to process continue executing and
// just want to stop long-running handler to stop and you are using "context aware" methods (ala db queries with ctx)
// 	e.GET("/", func(c echo.Context) error {
//
//		doneCh := make(chan error)
//		go func(ctx context.Context) {
//			doneCh <- myPossiblyLongRunningBackgroundTaskWithCtx(ctx)
//		}(c.Request().Context())
//
//		select { // wait for task to finish or context to timeout/cancelled
//		case err := <-doneCh:
//			if err != nil {
//				return err
//			}
//			return c.String(http.StatusOK, "OK")
//		case <-c.Request().Context().Done():
//			if c.Request().Context().Err() == context.DeadlineExceeded {
//				return c.String(http.StatusServiceUnavailable, "timeout")
//			}
//			return c.Request().Context().Err()
//		}
//
//	})
//

type (
	// TimeoutConfig defines the config for Timeout middleware.
	TimeoutConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// ErrorMessage is written to response on timeout in addition to http.StatusServiceUnavailable (503) status code
		// It can be used to define a custom timeout error message
		ErrorMessage string

		// OnTimeoutRouteErrorHandler is an error handler that is executed for error that was returned from wrapped route after
		// request timeouted and we already had sent the error code (503) and message response to the client.
		// NB: do not write headers/body inside this handler. The response has already been sent to the client and response writer
		// will not accept anything no more. If you want to know what actual route middleware timeouted use `c.Path()`
		OnTimeoutRouteErrorHandler func(err error, c echo.Context)

		// Timeout configures a timeout for the middleware, defaults to 0 for no timeout
		// NOTE: when difference between timeout duration and handler execution time is almost the same (in range of 100microseconds)
		// the result of timeout does not seem to be reliable - could respond timeout, could respond handler output
		// difference over 500microseconds (0.5millisecond) response seems to be reliable
		Timeout time.Duration
	}
)

var (
	// DefaultTimeoutConfig is the default Timeout middleware config.
	DefaultTimeoutConfig = TimeoutConfig{
		Skipper:      DefaultSkipper,
		Timeout:      0,
		ErrorMessage: "",
	}
)

// Timeout returns a middleware which returns error (503 Service Unavailable error) to client immediately when handler
// call runs for longer than its time limit. NB: timeout does not stop handler execution.
func Timeout() echo.MiddlewareFunc {
	return TimeoutWithConfig(DefaultTimeoutConfig)
}

// TimeoutWithConfig returns a Timeout middleware with config.
// See: `Timeout()`.
func TimeoutWithConfig(config TimeoutConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTimeoutConfig.Skipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.Timeout == 0 {
				return next(c)
			}

			handlerWrapper := echoHandlerFuncWrapper{
				ctx:        c,
				handler:    next,
				errChan:    make(chan error, 1),
				errHandler: config.OnTimeoutRouteErrorHandler,
			}
			handler := http.TimeoutHandler(handlerWrapper, config.Timeout, config.ErrorMessage)
			handler.ServeHTTP(c.Response().Writer, c.Request())

			select {
			case err := <-handlerWrapper.errChan:
				return err
			default:
				return nil
			}
		}
	}
}

type echoHandlerFuncWrapper struct {
	ctx        echo.Context
	handler    echo.HandlerFunc
	errHandler func(err error, c echo.Context)
	errChan    chan error
}

func (t echoHandlerFuncWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// replace echo.Context Request with the one provided by TimeoutHandler to let later middlewares/handler on the chain
	// handle properly it's cancellation
	t.ctx.SetRequest(r)

	// replace writer with TimeoutHandler custom one. This will guarantee that
	// `writes by h to its ResponseWriter will return ErrHandlerTimeout.`
	originalWriter := t.ctx.Response().Writer
	t.ctx.Response().Writer = rw

	// in case of panic we restore original writer and call panic again
	// so it could be handled with global middleware Recover()
	defer func() {
		if err := recover(); err != nil {
			t.ctx.Response().Writer = originalWriter
			panic(err)
		}
	}()

	err := t.handler(t.ctx)
	if ctxErr := r.Context().Err(); ctxErr == context.DeadlineExceeded {
		if err != nil && t.errHandler != nil {
			t.errHandler(err, t.ctx)
		}
		return // on timeout we can not send handler error to client because `http.TimeoutHandler` has already sent headers
	}
	// we restore original writer only for cases we did not timeout. On timeout we have already sent response to client
	// and should not anymore send additional headers/data
	// so on timeout writer stays what http.TimeoutHandler uses and prevents writing headers/body
	if err != nil {
		// Error must be written into Writer created in `http.TimeoutHandler` so to get Response into `commited` state.
		// So call global error handler to write error to the client. This is needed or `http.TimeoutHandler` will send
		// status code by itself and after that our tries to write status code will not work anymore and/or create errors in
		// log about `superfluous response.WriteHeader call from`
		t.ctx.Error(err)
		// we pass error from handler to middlewares up in handler chain to act on it if needed. But this means that
		// global error handler is probably be called twice as `t.ctx.Error` already does that.

		// NB: later call of the global error handler or middlewares will not take any effect, as echo.Response will be
		// already marked as `committed` because we called global error handler above.
		t.ctx.Response().Writer = originalWriter // make sure we restore before we signal original coroutine about the error
		t.errChan <- err
		return
	}
	t.ctx.Response().Writer = originalWriter
}
