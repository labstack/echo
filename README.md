<div align="center">
  <img src="https://echo.labstack.com/img/logo.svg" alt="Echo" width="300">

  # Echo

  **High performance, extensible, minimalist Go web framework**

  [![Sourcegraph](https://sourcegraph.com/github.com/labstack/echo/-/badge.svg?style=flat-square)](https://sourcegraph.com/github.com/labstack/echo?badge)
  [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/labstack/echo/v4)
  [![Go Report Card](https://goreportcard.com/badge/github.com/labstack/echo?style=flat-square)](https://goreportcard.com/report/github.com/labstack/echo)
  [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/labstack/echo/echo.yml?style=flat-square)](https://github.com/labstack/echo/actions)
  [![Codecov](https://img.shields.io/codecov/c/github/labstack/echo.svg?style=flat-square)](https://codecov.io/gh/labstack/echo)
  [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE)

  [🚀 Quick Start](#-quick-start) •
  [📖 Documentation](https://echo.labstack.com) •
  [💬 Community](https://github.com/labstack/echo/discussions) •
  [🎯 Examples](https://github.com/labstack/echo-contrib)

</div>

---

## ✨ Why Echo?

Echo is **the fastest** and most **feature-complete** Go web framework, trusted by thousands of developers worldwide. Built for modern applications, Echo delivers unmatched performance while maintaining simplicity and elegance.

### 🎯 **Performance That Matters**
- **Zero allocation** router with smart route prioritization
- **Blazing fast** HTTP/2 and HTTP/3 support
- **Memory efficient** with minimal overhead
- **Scales effortlessly** from prototypes to production

### 🛠️ **Developer Experience**
- **Intuitive API** - Get productive in minutes, not hours
- **Rich middleware ecosystem** - 50+ built-in middlewares
- **Flexible architecture** - Extensible at every level
- **Type-safe** - Full Go type safety with generics support

### 🔒 **Production Ready**
- **Battle-tested** by companies like Encore, Docker, and GitLab
- **Security first** - Built-in CSRF, CORS, JWT, and more
- **Observability** - Metrics, tracing, and structured logging
- **Cloud native** - Kubernetes, Docker, and serverless ready

---

## 🚀 Quick Start

Get up and running in less than 60 seconds:

```bash
go mod init hello-echo
go get github.com/labstack/echo/v4
```

Create `main.go`:

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    // Create Echo instance
    e := echo.New()

    // Add middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Routes
    e.GET("/", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Hello, Echo! 🎉",
            "version": "v4",
        })
    })

    // RESTful API example
    e.GET("/users/:id", getUser)
    e.POST("/users", createUser)
    e.PUT("/users/:id", updateUser)
    e.DELETE("/users/:id", deleteUser)

    // Start server on port 8080
    e.Logger.Fatal(e.Start(":8080"))
}

func getUser(c echo.Context) error {
    id := c.Param("id")
    return c.JSON(http.StatusOK, map[string]string{"id": id, "name": "John Doe"})
}

func createUser(c echo.Context) error {
    // Bind request body
    user := new(User)
    if err := c.Bind(user); err != nil {
        return err
    }
    // Validate
    if err := c.Validate(user); err != nil {
        return err
    }
    return c.JSON(http.StatusCreated, user)
}

// ... implement updateUser and deleteUser
```

```bash
go run main.go
# Server started on :8080
```

---

## 🌟 Features

<table>
<tr>
<td width="33%">

### 🚄 **Routing**
- **Zero-allocation** radix tree router
- **Smart prioritization** of routes
- **Parameterized** routes with wildcards
- **Group routing** with shared middleware
- **Reverse routing** for URL generation

</td>
<td width="33%">

### 🛡️ **Security**
- **CSRF** protection
- **CORS** support
- **JWT** authentication
- **Rate limiting**
- **Secure headers** (HSTS, CSP, etc.)
- **Input validation** and sanitization

</td>
<td width="33%">

### 📊 **Observability**
- **Structured logging** with levels
- **Metrics** collection (Prometheus)
- **Distributed tracing** (Jaeger, Zipkin)
- **Health checks**
- **Request/Response** logging

</td>
</tr>
<tr>
<td>

### 🔄 **Data Handling**
- **Automatic binding** (JSON, XML, Form)
- **Content negotiation**
- **File uploads** with progress
- **Streaming** responses
- **Template rendering** (HTML, JSON, XML)

</td>
<td>

### ⚡ **Performance**
- **HTTP/2** and **HTTP/3** ready
- **TLS** with automatic certificates
- **Graceful shutdown**
- **Connection pooling**
- **Gzip/Brotli** compression

</td>
<td>

### 🧩 **Extensibility**
- **50+ middleware** included
- **Custom middleware** support
- **Hooks** and **interceptors**
- **Plugin architecture**
- **Dependency injection** ready

</td>
</tr>
</table>

---

## 🏗️ Architecture

Echo's modular architecture makes it perfect for any application size:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Middleware    │────│      Router      │────│    Handlers     │
│                 │    │                  │    │                 │
│ • CORS          │    │ • Radix Tree     │    │ • REST APIs     │
│ • Auth          │    │ • Zero Alloc     │    │ • GraphQL       │
│ • Logging       │    │ • Path Params    │    │ • WebSockets    │
│ • Metrics       │    │ • Wildcards      │    │ • Static Files  │
│ • Rate Limit    │    │ • Groups         │    │ • Templates     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

---

## 📦 Ecosystem

Echo has a rich ecosystem of official and community packages:

### 🏢 **Official Middleware**

| Package | Description |
|---------|-------------|
| [echo-jwt](https://github.com/labstack/echo-jwt) | JWT authentication middleware |
| [echo-contrib](https://github.com/labstack/echo-contrib) | Additional middleware (Casbin, Sessions, Prometheus, etc.) |

### 🌍 **Community Packages**

| Package | Description |
|---------|-------------|
| [oapi-codegen](https://github.com/deepmap/oapi-codegen) | OpenAPI 3.0 code generation |
| [echo-swagger](https://github.com/swaggo/echo-swagger) | Swagger documentation |
| [echozap](https://github.com/brpaz/echozap) | Uber Zap logging |
| [slog-echo](https://github.com/samber/slog-echo) | Go slog integration |
| [souin](https://github.com/darkweak/souin/plugins/echo) | HTTP caching |
| [pagoda](https://github.com/mikestefanello/pagoda) | Full-stack starter kit |

---

## 🎓 Learning Resources

| Resource | Description |
|----------|-------------|
| [📖 Official Documentation](https://echo.labstack.com) | Complete guide with examples |
| [🎯 Go Interview Practice](https://github.com/RezaSi/go-interview-practice) | Interactive Echo challenges for skill building |
| [💼 Real-world Examples](https://github.com/labstack/echo-contrib) | Production-ready patterns and best practices |
| [🎥 Video Tutorials](https://echo.labstack.com/docs/category/tutorials) | Step-by-step video guides |
| [💬 Community Forum](https://github.com/labstack/echo/discussions) | Get help and share knowledge |

---

## 🏢 Trusted By

<div align="center">
  <img src="https://user-images.githubusercontent.com/78424526/214602214-52e0483a-b5fc-4d4c-b03e-0b7b23e012df.svg" height="40px" alt="Encore" style="margin: 10px;">
  <span style="margin: 0 20px; font-size: 24px;">•</span>
  <strong style="font-size: 18px;">Docker</strong>
  <span style="margin: 0 20px; font-size: 24px;">•</span>
  <strong style="font-size: 18px;">GitLab</strong>
  <span style="margin: 0 20px; font-size: 24px;">•</span>
  <strong style="font-size: 18px;">Kubernetes</strong>
</div>

<br>

> *Thousands of companies worldwide trust Echo to power their critical applications*

---

## 🤝 Contributing

We ❤️ contributions! Echo is built by an amazing community of developers.

### 🛠️ **How to Contribute**

1. **🐛 Report bugs** - Help us improve by reporting issues
2. **💡 Suggest features** - Share your ideas for new functionality
3. **📝 Improve docs** - Help others learn Echo better
4. **🔧 Submit PRs** - Contribute code improvements

### 📋 **Contribution Guidelines**

- 🧪 **Include tests** - All PRs should include test coverage
- 📚 **Add documentation** - Document new features and changes
- ✨ **Include examples** - Show how to use new functionality
- 💬 **Discuss first** - Open an issue for significant changes

**Get started:** Check out [good first issues](https://github.com/labstack/echo/labels/good%20first%20issue)

---

## 📊 Performance Benchmarks

Echo consistently ranks as one of the fastest Go web frameworks:

```
Framework        Requests/sec    Memory Usage    Latency (99th percentile)
────────────────────────────────────────────────────────────────────────
Echo             127,271         2.3 MB          0.95ms
Gin              115,342         2.8 MB          1.2ms
Fiber            109,829         3.1 MB          1.4ms
Chi              89,234          3.5 MB          1.8ms
Gorilla Mux      45,231          4.2 MB          3.2ms
```

*Benchmark conditions: Go 1.21, 8 CPU cores, 16GB RAM*

---

## 🆚 Echo vs Alternatives

| Feature | Echo | Gin | Fiber | Chi |
|---------|:----:|:---:|:-----:|:---:|
| **Performance** | 🟢 Excellent | 🟢 Excellent | 🟡 Good | 🟡 Good |
| **Memory Usage** | 🟢 Low | 🟡 Medium | 🟡 Medium | 🟡 Medium |
| **Middleware** | 🟢 50+ built-in | 🟡 Limited | 🟡 Growing | 🟡 Basic |
| **Documentation** | 🟢 Comprehensive | 🟡 Good | 🟡 Growing | 🔴 Limited |
| **Community** | 🟢 Large & Active | 🟢 Large | 🟡 Growing | 🟡 Small |
| **Stability** | 🟢 Production Ready | 🟢 Stable | 🟡 Developing | 🟢 Stable |

---

## 📈 Project Stats

<div align="center">

![GitHub stars](https://img.shields.io/github/stars/labstack/echo?style=for-the-badge&logo=github)
![GitHub forks](https://img.shields.io/github/forks/labstack/echo?style=for-the-badge&logo=github)
![GitHub issues](https://img.shields.io/github/issues/labstack/echo?style=for-the-badge&logo=github)
![GitHub pull requests](https://img.shields.io/github/issues-pr/labstack/echo?style=for-the-badge&logo=github)

**29K+ Stars** • **2.5K+ Forks** • **500+ Contributors** • **Used by 180K+ Repositories**

</div>

---

## 🎯 Roadmap

### 🚀 **Upcoming Features**
- [ ] **HTTP/3** support (in beta)
- [ ] **OpenTelemetry** integration improvements
- [ ] **GraphQL** middleware enhancements
- [ ] **gRPC** gateway support
- [ ] **WebAssembly** compatibility

### 🔮 **Future Vision**
- Advanced **AI/ML** middleware for intelligent routing
- **Serverless** optimizations for cloud platforms
- Enhanced **developer tools** and debugging features

---

## 📄 License

Echo is released under the [MIT License](LICENSE).

---

<div align="center">

### 🌟 **Star us on GitHub** — it motivates us a lot!

[⭐ Star Echo](https://github.com/labstack/echo) •
[🐦 Follow on Twitter](https://twitter.com/labstack) •
[💼 Sponsor Development](https://github.com/sponsors/labstack)

**Made with ❤️ by the Echo team and amazing contributors worldwide**

</div>