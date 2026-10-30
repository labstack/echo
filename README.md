<p align="center">
  <a href="https://echo.labstack.com">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/labstack/echo/master/.github/assets/echo-logo-dark.svg">
      <img alt="Echo" src="https://raw.githubusercontent.com/labstack/echo/master/.github/assets/echo-logo-light.svg" height="72">
    </picture>
  </a>
</p>

<p align="center"><strong>High performance, extensible, minimalist Go web framework.</strong><br>Echo v4 maintenance line</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/labstack/echo/v4"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/labstack/echo/v4.svg"></a>
  <a href="https://github.com/labstack/echo/actions/workflows/echo.yml?query=branch%3Av4"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/labstack/echo/echo.yml?branch=v4&style=flat-square&label=ci"></a>
  <a href="https://codecov.io/gh/labstack/echo/tree/v4"><img alt="Codecov" src="https://img.shields.io/codecov/c/github/labstack/echo/v4.svg?style=flat-square"></a>
  <a href="https://github.com/labstack/echo/discussions"><img alt="Discussions" src="https://img.shields.io/badge/community-discussions-00afd1.svg?style=flat-square"></a>
  <a href="https://x.com/labstack"><img alt="X" src="https://img.shields.io/badge/x-@labstack-000000.svg?style=flat-square"></a>
  <a href="https://github.com/labstack/echo/blob/v4/LICENSE"><img alt="License" src="https://img.shields.io/badge/license-mit-blue.svg?style=flat-square"></a>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/labstack/echo/v4">v4 API reference</a> ·
  <a href="https://github.com/labstack/echo/blob/master/API_CHANGES_V5.md">Upgrade to v5</a> ·
  <a href="https://echo.labstack.com">Website (v5)</a> ·
  <a href="https://github.com/labstack/echo/discussions">Discussions</a> ·
  <a href="https://echo.labstack.com/pro#support">Commercial support</a>
</p>

## About

This branch holds Echo v4, the maintenance line. Echo v5 is the current release; start new projects on [v5](https://github.com/labstack/echo).

### Support

Echo v4 receives bug fixes until 2026-12-31 and free security fixes until 2027-12-31, on its latest release only. It gets no new features. See [SECURITY.md](https://github.com/labstack/echo/blob/master/SECURITY.md) for the full policy.

Echo v4 requires Go 1.26 or later (since v4.16.1), to use patched `golang.org/x/` dependencies.

[Echo Support](https://echo.labstack.com/pro#support) is a paid annual plan from LabStack for v4 and v5 applications. It includes a named contact, email support with a two-business-day response target, and an annual maintenance-status letter.

## Upgrade notes

### Client IP security migration since v4.16.3

> **Important:** Behind a proxy, set `Echo#IPExtractor` before you upgrade to v4.16.3. Otherwise all clients behind the proxy share one IP address for rate limiting, access controls and logging.

Echo v4.16.3 changes client IP defaults. Without `Echo#IPExtractor`, `Context.RealIP()` uses the connection peer and ignores `X-Forwarded-For` and `X-Real-IP`. Proxy middleware always replaces upstream `X-Real-IP` with this result. This fixes [CVE-2026-118651](https://github.com/labstack/echo/security/advisories/GHSA-99jh-6h7p-pp36) and [CVE-2026-118653](https://github.com/labstack/echo/security/advisories/GHSA-246p-cpwv-v3jq). Applications that already set `Echo#IPExtractor` keep their behavior. A custom rate-limit identifier that does not call `RealIP()` is also unaffected.

Behind a proxy, configure `Echo#IPExtractor` before starting the server. Otherwise, clients behind the same proxy share its IP address for rate limiting, access controls and logging. Use `echo.ExtractIPDirect()` when there is no proxy. For proxies that set `X-Forwarded-For`, turn off the broad default trust ranges and trust only your proxy addresses:

```go
// After e := echo.New(); imports "log" and "net".
_, proxyRange, err := net.ParseCIDR("10.0.0.10/32") // replace with your proxy's address or CIDR
if err != nil {
	log.Fatal(err)
}
e.IPExtractor = echo.ExtractIPFromXFFHeader(
	echo.TrustLoopback(false),
	echo.TrustLinkLocal(false),
	echo.TrustPrivateNet(false),
	echo.TrustIPRange(proxyRange),
)
```

If your proxy uses `X-Real-IP`, use `echo.ExtractIPFromRealIPHeader` with the same trust options. Configure the outermost proxy to replace client-supplied identity headers.

### Security defaults since v4.16.1

See the [v4.16.1 release notes](https://github.com/labstack/echo/releases/tag/v4.16.1) for the advisories.

`BodyDump` captures at most 5 MiB of the request and 5 MiB of the response by default. The application and client still receive the complete bodies. Set `BodyDumpConfig.MaxRequestBytes` and `MaxResponseBytes` to change the capture limits; `-1` removes a limit.

`Decompress` limits decompressed request bodies to 100 MiB by default. Reading beyond the limit returns HTTP 413, including through JSON and XML binding. Set `DecompressConfig.MaxDecompressedSize` to change the limit; `-1` removes it. Apply any smaller application body limit after decompression.

The proxy middleware validates WebSocket upgrade responses, closes both WebSocket connections when either side disconnects, and removes incoming `Forwarded` headers. Backends should trust only the forwarding headers the proxy sets and enforce their own authorization.

### Request scheme since v4.16.0

`Context.Scheme()` uses `X-Forwarded-Proto` and related headers only when the request comes directly from a loopback, link-local or private network address, or a unix socket. If a proxy with a public address terminates TLS and forwards plain HTTP to Echo, `HTTPSRedirect` keeps redirecting until you set `Echo#SchemeExtractor` to trust that proxy. See the [v4.16.0 changelog](https://github.com/labstack/echo/blob/v4/CHANGELOG.md#v4160---2026-09-27) for this and the other fixes in that release.

## Features

- Router with static, parameter and wildcard routes; static routes match first
- RESTful APIs with route groups
- Middleware at the root, group or route level
- Data binding for JSON, XML, form, query and path parameters
- Helpers for common HTTP responses
- Centralized HTTP error handling
- Template rendering with any template engine
- Customizable logging
- Automatic TLS through Let's Encrypt
- HTTP/2, including h2c

## Sponsors

| | Sponsor |
| --- | --- |
| <a href="https://blacksmith.sh"><img src="https://github.com/useblacksmith.png?size=56" height="28" alt="Blacksmith logo"></a> | [Blacksmith](https://blacksmith.sh): faster, drop-in GitHub Actions runners |
| <a href="https://encore.dev"><img src="https://user-images.githubusercontent.com/78424526/214602214-52e0483a-b5fc-4d4c-b03e-0b7b23e012df.svg" height="28" alt="Encore logo"></a> | [Encore](https://encore.dev): the platform for building Go-based cloud backends |

Echo is independent and community funded. If your company depends on Echo, please consider [sponsoring it](https://github.com/sponsors/labstack).

## Getting started

### Installation

```sh
go get github.com/labstack/echo/v4
```

### Example

```go
package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.RequestLogger()) // log requests with slog
	e.Use(middleware.Recover())       // turn panics into errors

	e.GET("/", hello)

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
```

Run it and send a request:

```console
$ go run .
$ curl http://localhost:8080/
Hello, World!
```

## Official middleware

The Echo team maintains these repositories.

| Repository | Description |
| --- | --- |
| [github.com/labstack/echo-jwt](https://github.com/labstack/echo-jwt) | [JWT](https://github.com/golang-jwt/jwt) middleware |
| [github.com/labstack/echo-contrib](https://github.com/labstack/echo-contrib) | [casbin](https://github.com/casbin/casbin), [gorilla/sessions](https://github.com/gorilla/sessions), [jaegertracing](https://github.com/uber/jaeger-client-go), [prometheus](https://github.com/prometheus/client_golang/), [pprof](https://pkg.go.dev/net/http/pprof) and [zipkin](https://github.com/openzipkin/zipkin-go) middleware |
| [github.com/labstack/echo-otel](https://github.com/labstack/echo-otel/tree/v4) | [OpenTelemetry](https://opentelemetry.io/) tracing and metrics middleware (`github.com/labstack/echo-otel/v4`) |

## Third-party middleware

<details>
<summary>Third-party middleware list</summary>

The Echo team does not review or maintain these projects. Check their quality and security, and that each one supports v4, before you use them.

| Repository | Description |
| --- | --- |
| [oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) | Generates server and client code from [OpenAPI](https://swagger.io/specification/) specifications |
| [github.com/swaggo/echo-swagger](https://github.com/swaggo/echo-swagger) | Serves [Swagger](https://swagger.io/) 2.0 API documentation |
| [github.com/ziflex/lecho](https://github.com/ziflex/lecho) | [Zerolog](https://github.com/rs/zerolog) adapter for the Echo logger interface |
| [github.com/brpaz/echozap](https://github.com/brpaz/echozap) | Uber's [Zap](https://github.com/uber-go/zap) adapter for the Echo logger interface |
| [github.com/samber/slog-echo](https://github.com/samber/slog-echo) | [slog](https://pkg.go.dev/log/slog) request logging middleware |
| [github.com/darkweak/souin/plugins/echo](https://github.com/darkweak/souin/tree/master/plugins/echo) | HTTP cache based on [Souin](https://github.com/darkweak/souin), with distributed and local storage backends |
| [github.com/mikestefanello/pagoda](https://github.com/mikestefanello/pagoda) | Full-stack web starter kit built with Echo |
| [github.com/go-woo/protoc-gen-echo](https://github.com/go-woo/protoc-gen-echo) | Generates Echo server code from Protocol Buffers |

</details>

To add your library, send a pull request.

## Contribute

v4 takes bug and security fixes only (see [Support](#support)). Send new features to v5 on the [master branch](https://github.com/labstack/echo).

- Open an issue before a large change. Send small fixes directly as a pull request.
- Include a test, documentation and, where it helps, an example.

Report security issues privately as [SECURITY.md](https://github.com/labstack/echo/blob/master/SECURITY.md) describes, not in public issues.

## Credits

- [Vishal Rana](https://github.com/vishr) (Author and Maintainer)
- [Nitin Rana](https://github.com/nr17) (Consultant)
- [Martti T.](https://github.com/aldas) (Maintainer Emeritus)
- [Roland Lammel](https://github.com/lammel) (Maintainer Emeritus)
- [Pablo Andres Fuente](https://github.com/pafuent) (Maintainer Emeritus)
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](https://github.com/labstack/echo/blob/v4/LICENSE)
