// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"net/http"
	"strings"
)

// NewVirtualHostHandler creates instance of Echo that routes requests to given virtual hosts
// when hosts in request does not exist in given map the request is served by returned Echo instance.
//
// Host names are matched case-insensitively (RFC 9110 Section 4.2.3), so a request with
// `Host: Example.com` reaches the instance registered as `example.com`. The map is copied,
// later changes to it have no effect.
func NewVirtualHostHandler(vhosts map[string]*Echo) *Echo {
	e := New()
	hosts := make(map[string]*Echo, len(vhosts))
	for host, vhost := range vhosts {
		hosts[strings.ToLower(host)] = vhost
	}
	e.serveHTTPFunc = func(w http.ResponseWriter, r *http.Request) {
		if vh, ok := hosts[strings.ToLower(r.Host)]; ok {
			vh.ServeHTTP(w, r)
			return
		}
		e.serveHTTP(w, r) // unknown host in request
	}
	return e
}
