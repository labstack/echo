package middleware

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/semaphore"
)

type (
	// QueueConfig defines the configuration for the queue
	QueueConfig struct {
		Skipper    Skipper
		BeforeFunc BeforeFunc

		QueueSize int
		Workers   int

		QueueTimeout  time.Duration
		WorkerTimeout time.Duration
	}
)

// errors
var (
	// ErrQueueTimeout is thrown when the request waits more than QueueTimeout to be queued
	ErrQueueTimeout = echo.NewHTTPError(http.StatusTooManyRequests, "queue limit reached")
	// ErrWorkerTimeout is thrown when the request waits more than WorkerTimeout for a worker to start processing it
	ErrWorkerTimeout = echo.NewHTTPError(http.StatusRequestTimeout, "request took too long to process")
)

// DefaultQueueConfig defines default values for QueueConfig
var DefaultQueueConfig = QueueConfig{
	Skipper:       DefaultSkipper,
	QueueSize:     100,
	Workers:       runtime.GOMAXPROCS(0),
	QueueTimeout:  30 * time.Second,
	WorkerTimeout: 10 * time.Second,
}

/*
Queue returns a queue middleware

	e := echo.New()

	e.GET("/queue", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, middleware.Queue())

*/
func Queue() echo.MiddlewareFunc {
	config := DefaultQueueConfig

	return QueueWithConfig(config)
}

/*
QueueWithConfig returns a queue middleware

	e := echo.New()

	config := middleware.QueueConfig{
		Skipper: DefaultSkipper,
		QueueSize:     2,
		Workers:       1,
		QueueTimeout:  20 * time.Second,
		WorkerTimeout: 10 * time.Second,
	}

	e.GET("/queue", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, middleware.QueueWithConfig(config))

*/
func QueueWithConfig(config QueueConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultQueueConfig.Skipper
	}

	queueSemaphore := semaphore.NewWeighted(int64(config.QueueSize))
	workersSemaphore := semaphore.NewWeighted(int64(config.Workers))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			ctxQueue, queueCancel := context.WithTimeout(c.Request().Context(), config.QueueTimeout)

			if err := queueSemaphore.Acquire(ctxQueue, 1); err != nil {
				queueCancel()

				if errors.Is(err, context.DeadlineExceeded) {
					return ErrQueueTimeout
				}

				return err
			}

			ctxWorker, workerCancel := context.WithTimeout(ctxQueue, config.WorkerTimeout)

			defer func() {
				workerCancel()
				queueCancel()
			}()

			if err := workersSemaphore.Acquire(ctxWorker, 1); err != nil {
				queueSemaphore.Release(1)

				if errors.Is(err, context.DeadlineExceeded) {
					return ErrWorkerTimeout
				}

				return err
			}

			defer func() {
				workersSemaphore.Release(1)
				queueSemaphore.Release(1)
			}()

			return next(c)
		}
	}
}
