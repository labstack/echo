[![Latest release](https://img.shields.io/github/v/release/labstack/echo?style=flat-square&label=release&color=00afd1)](https://github.com/labstack/echo/releases)
[![Last commit](https://img.shields.io/github/last-commit/labstack/echo/master?style=flat-square)](https://github.com/labstack/echo/commits/master)
[![Go Reference](https://pkg.go.dev/badge/github.com/labstack/echo/v5.svg)](https://pkg.go.dev/github.com/labstack/echo/v5)
[![CI](https://img.shields.io/github/actions/workflow/status/labstack/echo/ci.yml?style=flat-square&label=ci)](https://github.com/labstack/echo/actions/workflows/ci.yml)
[![Codecov](https://img.shields.io/codecov/c/github/labstack/echo.svg?style=flat-square)](https://codecov.io/gh/labstack/echo)
[![OpenSSF Scorecard](https://img.shields.io/ossf-scorecard/github.com/labstack/echo?style=flat-square&label=openssf%20scorecard)](https://scorecard.dev/viewer/?uri=github.com/labstack/echo)
[![Sponsors](https://img.shields.io/github/sponsors/labstack?style=flat-square&color=db61a2)](https://github.com/sponsors/labstack)
[![Discussions](https://img.shields.io/badge/community-discussions-00afd1.svg?style=flat-square)](https://github.com/labstack/echo/discussions)
[![X](https://img.shields.io/badge/x-@labstack-000000.svg?style=flat-square)](https://x.com/labstack)
[![License](https://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://github.com/labstack/echo/blob/master/LICENSE)

## Echo

High performance, extensible, minimalist Go web framework.

Echo is built on Go's standard `net/http` and interoperates with it through `echo.WrapHandler` and `echo.WrapMiddleware`. It adds the parts the standard library leaves to you: a fast radix-tree router, request binding with pluggable validation, a deep middleware ecosystem, and centralized error handling. Echo is actively maintained, and `v5` is the current release line.

- [Website](https://echo.labstack.com)
- [Quick start](https://echo.labstack.com/guide/quickstart/)
- [Middleware](https://echo.labstack.com/middleware/)
- [API reference](https://pkg.go.dev/github.com/labstack/echo/v5)
- Help and questions: [GitHub Discussions](https://github.com/labstack/echo/discussions)

### Features

- Fast radix-tree [router](https://echo.labstack.com/guide/routing/) that prioritizes routes smartly
- Route groups, with middleware at the root, group or route level
- [Data binding](https://echo.labstack.com/guide/binding/) for JSON, XML, form, query and path parameters, with pluggable validation
- Helpers for JSON, XML, HTML, file, stream and other responses
- Centralized [error handling](https://echo.labstack.com/guide/error-handling/)
- Structured request logging with `log/slog`
- [Template rendering](https://echo.labstack.com/guide/templates/) with any template engine
- TLS, [HTTP/2](https://echo.labstack.com/cookbook/http2/), [automatic TLS](https://echo.labstack.com/cookbook/auto-tls/) with Let's Encrypt, and [graceful shutdown](https://echo.labstack.com/cookbook/graceful-shutdown/)

## Used by

Echo runs in production across the Go ecosystem, including:

| Project | What runs on Echo |
| --- | --- |
| [Bluesky indigo](https://github.com/bluesky-social/indigo) | AT Protocol Relay and network services |
| [Temporal UI](https://github.com/temporalio/ui) | Temporal Web UI server |
| [NVIDIA Cloud Functions](https://github.com/NVIDIA/nvcf) | LLM API gateway |
| [Azure Container Networking](https://github.com/Azure/azure-container-networking) | Container Networking Service REST API |
| [osbuild-composer](https://github.com/osbuild/osbuild-composer) | Red Hat Image Builder cloud API |
| [Docker Agent](https://github.com/docker/docker-agent) | Agent API, chat and A2A servers |
| [YugabyteDB](https://github.com/yugabyte/yugabyte-db) | yugabyted UI API server |
| [PostHog](https://github.com/PostHog/posthog) | Livestream service |
| [Bytebase](https://github.com/bytebase/bytebase) | Backend server and API |
| [Hatchet](https://github.com/hatchet-dev/hatchet) | API server |
| [CasaOS](https://github.com/IceWhaleTech/CasaOS) | API server |
| [Hanko](https://github.com/teamhanko/hanko) | Authentication backend |

See [ADOPTERS.md](./ADOPTERS.md) for the full list. Using Echo in an open source project? [Add it](./ADOPTERS.md#add-your-project).

## Sponsors

| | Sponsor |
| --- | --- |
| <a href="https://blacksmith.sh"><img src="https://github.com/useblacksmith.png?size=56" height="28" alt="Blacksmith logo"></a> | [Blacksmith](https://blacksmith.sh): faster, drop-in GitHub Actions runners |
| <a href="https://encore.dev"><img src="https://user-images.githubusercontent.com/78424526/214602214-52e0483a-b5fc-4d4c-b03e-0b7b23e012df.svg" height="28" alt="Encore logo"></a> | [Encore](https://encore.dev): the platform for building Go-based cloud backends |
| <a href="https://github.com/gravitycarbon"><img src="https://github.com/gravitycarbon.png?size=56" height="28" alt="Gravity Carbon logo"></a> | [Gravity Carbon](https://github.com/gravitycarbon) |

Echo is independent and community funded. If your company depends on Echo, please consider [sponsoring it](https://github.com/sponsors/labstack).

## Getting started

### Supported versions

- `v5` is the current major version, released on 2026-01-18. See [API_CHANGES_V5.md](./API_CHANGES_V5.md) for the API changes from `v4` and upgrade notes.
- `v4` receives security updates and bug fixes until 2026-12-31.

See [ROADMAP.md](./ROADMAP.md) for where Echo is heading and the version support policy.

### Installation

```sh
go get github.com/labstack/echo/v5
```

Echo supports the three most recent Go [releases](https://go.dev/doc/devel/release) and may work with older ones.

### Example

```go
package main

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLogger()) // log requests with log/slog
	e.Use(middleware.Recover())       // recover from panics and return an error

	// Routes
	e.GET("/", hello)

	// Start server
	if err := e.Start(":8080"); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
```

More in the [guide](https://echo.labstack.com/guide/quickstart/) and the [cookbook](https://echo.labstack.com/cookbook/hello-world/).

## Middleware

### Official

Maintained by the Echo team.

| Repository | Description |
| --- | --- |
| [labstack/echo-jwt](https://github.com/labstack/echo-jwt) | [JWT](https://github.com/golang-jwt/jwt) authentication |
| [labstack/echo-contrib](https://github.com/labstack/echo-contrib) | [Casbin](https://github.com/casbin/casbin), [gorilla/sessions](https://github.com/gorilla/sessions) and [pprof](https://pkg.go.dev/net/http/pprof) |
| [labstack/echo-otel](https://github.com/labstack/echo-otel) | [OpenTelemetry](https://opentelemetry.io/) tracing and metrics |
| [labstack/echo-prometheus](https://github.com/labstack/echo-prometheus) | [Prometheus](https://github.com/prometheus/client_golang/) metrics |

### Third-party

These projects are maintained by their authors, not the Echo team. Review them before use, and check which Echo version (`v4` or `v5`) each supports.

| Repository | Description |
| --- | --- |
| [oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) | [OpenAPI](https://swagger.io/specification/) client and server code generator |
| [swaggo/echo-swagger](https://github.com/swaggo/echo-swagger) | Serves [Swagger](https://swagger.io/) 2.0 API documentation |
| [ziflex/lecho](https://github.com/ziflex/lecho) | [Zerolog](https://github.com/rs/zerolog) logger for Echo |
| [brpaz/echozap](https://github.com/brpaz/echozap) | Uber [Zap](https://github.com/uber-go/zap) logging middleware |
| [samber/slog-echo](https://github.com/samber/slog-echo) | [log/slog](https://pkg.go.dev/log/slog) logging middleware |
| [darkweak/souin](https://github.com/darkweak/souin/tree/master/plugins/echo) | HTTP cache middleware based on [Souin](https://github.com/darkweak/souin), with distributed storage support |
| [mikestefanello/pagoda](https://github.com/mikestefanello/pagoda) | Full-stack web development starter kit built on Echo |
| [go-woo/protoc-gen-echo](https://github.com/go-woo/protoc-gen-echo) | Generates Echo server code from Protocol Buffers |

To add your library, send a pull request.

## Contributing

Use issues for everything.

- For a small change, send a pull request.
- For a bigger change, open an issue to discuss it first.
- A pull request should include tests, documentation and, where it helps, an example.

You can also contribute by reporting issues, suggesting features and improving the documentation.

## Security

Report vulnerabilities privately through [GitHub Security Advisories](https://github.com/labstack/echo/security/advisories/new), not in public issues. See [SECURITY.md](./SECURITY.md) for supported versions.

## Credits

- [Vishal Rana](https://github.com/vishr) (Author and Maintainer)
- [Nitin Rana](https://github.com/nr17) (Consultant)
- [Martti T.](https://github.com/aldas) (Maintainer Emeritus)
- [Roland Lammel](https://github.com/lammel) (Maintainer Emeritus)
- [Pablo Andres Fuente](https://github.com/pafuent) (Maintainer Emeritus)
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](./LICENSE)
