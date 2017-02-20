// +build go1.8

package echo

import (
	stdContext "context"
)

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (e *Echo) Close() error {
	if err := e.TLSServer.Close(); err != nil {
		return err
	}
	return e.Server.Close()
}

// Shutdown stops server the gracefully.
// It internally calls `http.Server#Shutdown()`.
func (e *Echo) Shutdown(ctx stdContext.Context) error {
	if err := e.TLSServer.Shutdown(ctx); err != nil {
		return err
	}
	return e.Server.Shutdown(ctx)
}
