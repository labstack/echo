// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build !go1.20

package echo

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// TODO: remove when Go 1.23 is released and we do not support 1.19 anymore
func responseControllerFlush(rw http.ResponseWriter) error {
	for {
		switch t := rw.(type) {
		case interface{ FlushError() error }:
			return t.FlushError()
		case http.Flusher:
			t.Flush()
			return nil
		case interface{ Unwrap() http.ResponseWriter }:
			rw = t.Unwrap()
		default:
			return fmt.Errorf("%w", http.ErrNotSupported)
		}
	}
}

// TODO: remove when Go 1.23 is released and we do not support 1.19 anymore
func responseControllerHijack(rw http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	for {
		switch t := rw.(type) {
		case http.Hijacker:
			return t.Hijack()
		case interface{ Unwrap() http.ResponseWriter }:
			rw = t.Unwrap()
		default:
			return nil, nil, fmt.Errorf("%w", http.ErrNotSupported)
		}
	}
}
