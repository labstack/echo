// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

//go:build go1.20

package middleware

import (
	"bufio"
	"net"
	"net/http"
)

func responseControllerFlush(rw http.ResponseWriter) error {
	return http.NewResponseController(rw).Flush()
}

func responseControllerHijack(rw http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(rw).Hijack()
}
