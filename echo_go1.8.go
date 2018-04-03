// +build go1.8

package echo

import (
	stdContext "context"
	"fmt"
)

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (e *Echo) Close() error {
	cumuledErr := fmt.Errorf("on close:")
	ok := true
	if e.Quic {
		if err := e.QuicServer.Close(); err != nil {
			ok = false
			cumuledErr = fmt.Errorf("\nQUIC: %s", err.Error())
			e.QuicServer.Close()
		}
	}
	if err := e.TLSServer.Close(); err != nil {
		ok = false
		cumuledErr = fmt.Errorf("\nTLS: %s", err.Error())
		return err
	}
	if err := e.Server.Close(); err != nil {
		ok = false
		cumuledErr = fmt.Errorf("\nHTTP: %s", err.Error())
		return err
	}

	if ok {
		return nil
	}
	return cumuledErr
}

// Shutdown stops server the gracefully.
// It internally calls `http.Server#Shutdown()`.
func (e *Echo) Shutdown(ctx stdContext.Context) error {
	cumuledErr := fmt.Errorf("on shutdown:")
	ok := true
	if e.Quic {
		if err := e.QuicServer.Shutdown(ctx); err != nil {
			ok = false
			cumuledErr = fmt.Errorf("\nQUIC: %s", err.Error())
			e.QuicServer.Close()
		}
	}
	if err := e.TLSServer.Shutdown(ctx); err != nil {
		ok = false
		cumuledErr = fmt.Errorf("\nTLS: %s", err.Error())
		return err
	}
	if err := e.Server.Shutdown(ctx); err != nil {
		ok = false
		cumuledErr = fmt.Errorf("\nHTTP: %s", err.Error())
		return err
	}

	if ok {
		return nil
	}
	return cumuledErr
}
