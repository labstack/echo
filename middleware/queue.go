package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/semaphore"
)

type (
	// QueueConfig defines the configuration for the queue
	QueueConfig struct {
		Skipper    Skipper
		BeforeFunc BeforeFunc

		QueueSize int64
		Workers   int64

		QueueTimeout  time.Duration
		WorkerTimeout time.Duration
	}
)

// errors
var (
	// ErrQueueFull denotes an error raised when queue limit is reached
	ErrQueueFull = echo.NewHTTPError(http.StatusTooManyRequests, "queue limit reached")
	// ErrQueueTimeout denotes an error raised when context times out
	ErrQueueTimeout = echo.NewHTTPError(http.StatusRequestTimeout, "request took too long to process")
)

// DefaultQueueConfig defines default values for QueueConfig
var DefaultQueueConfig = QueueConfig{
	Skipper:       DefaultSkipper,
	QueueSize:     100,
	Workers:       8,
	QueueTimeout:  30 * time.Second,
	WorkerTimeout: 10 * time.Second,
}

/*
Queue returns a queue middleware

	e := echo.New()

	e.GET("/queue", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}, Queue())

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

	queueSemaphore := semaphore.NewWeighted(config.QueueSize)
	workersSemaphore := semaphore.NewWeighted(config.Workers)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			if config.BeforeFunc != nil {
				config.BeforeFunc(c)
			}

			ctxQueue, _ := context.WithTimeout(context.Background(), config.QueueTimeout)

			if err := queueSemaphore.Acquire(ctxQueue, 1); err != nil {
				c.Error(ErrQueueFull)
				return nil
			}

			ctxWorker, _ := context.WithTimeout(ctxQueue, config.WorkerTimeout)

			if err := workersSemaphore.Acquire(ctxWorker, 1); err != nil {
				queueSemaphore.Release(1)
				c.Error(ErrQueueTimeout)
				return nil
			}

			err := next(c)

			workersSemaphore.Release(1)
			queueSemaphore.Release(1)

			return err
		}
	}
}
