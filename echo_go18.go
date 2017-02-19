// +build go1.8

package echo

import (
	c "context"
)

// Close immediately stops server
// equivalent of func (*http.Server) Close() error
func (e *Echo) Close() error {
	if err := e.TLSServer.Close(); err != nil {
		return err
	}
	return e.Server.Close()
}

// Shutdown stops server gracefully
// equivalent of func (*http.Server) Shutdown(ctx context.Context) error
func (e *Echo) Shutdown(ctx c.Context) error {
	if err := e.TLSServer.Shutdown(ctx); err != nil {
		return err
	}
	return e.Server.Shutdown(ctx)
}
