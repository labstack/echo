# Echo v5 Public API Changes

**Comparison between `master` (v4.15.0) and `v5` (v5.0.0-alpha) branches**

Generated: 2026-01-01

---

## Executive Summary (by authors)

Echo `v5` is maintenance release with **major breaking changes** 
- `Context` is now struct instead of interface and we can add method to it in the future in minor versions.
- Adds new `Router` interface for possible new routing implementations.
- Drops old logging interface and uses moderm `log/slog` instead.
- Rearranges alot of methods/function signatures to make them more consistent.

## Executive Summary (by LLMs)

Echo v5 represents a **major breaking release** with significant architectural changes focused on:
- **Updated generic helpers** to take `*Context` and rename form helpers to `FormValue*`
- **Simplified API surface** by moving Context from interface to concrete struct
- **Modern Go patterns** including slog.Logger integration
- **Enhanced routing** with explicit RouteInfo and Routes types
- **Better error handling** with simplified HTTPError
- **New test helpers** via the `echotest` package

### Change Statistics

- **Major Breaking Changes**: 15+
- **New Functions Added**: 30+
- **Type Signature Changes**: 20+
- **Removed APIs**: 10+
- **New Packages Added**: 1 (`echotest`)
- **Version Change**: `4.15.0` → `5.0.0-alpha`

---

## Critical Breaking Changes

### 1. **Context: Interface → Concrete Struct**

**v4 (master):**
```go
type Context interface {
    Request() *http.Request
    // ... many methods
}

// Handler signature
func handler(c echo.Context) error
```

**v5:**
```go
type Context struct {
    // Has unexported fields
}

// Handler signature - NOW USES POINTER!
func handler(c *echo.Context) error
```

**Impact:** 🔴 **CRITICAL BREAKING CHANGE**
- ALL handlers must change from `echo.Context` to `*echo.Context`
- Context is now a concrete struct, not an interface
- This affects every single handler function in user code

**Migration:**
```go
// Before (v4)
func MyHandler(c echo.Context) error {
    return c.JSON(200, map[string]string{"hello": "world"})
}

// After (v5)
func MyHandler(c *echo.Context) error {
    return c.JSON(200, map[string]string{"hello": "world"})
}
```

---

### 2. **Logger: Custom Interface → slog.Logger**

**v4:**
```go
type Echo struct {
    Logger Logger  // Custom interface with Print, Debug, Info, etc.
}

type Logger interface {
    Output() io.Writer
    SetOutput(w io.Writer)
    Prefix() string
    // ... many custom methods
}

// Context returns Logger interface
func (c Context) Logger() Logger
```

**v5:**
```go
type Echo struct {
    Logger *slog.Logger  // Standard library structured logger
}

// Context returns slog.Logger
func (c *Context) Logger() *slog.Logger
func (c *Context) SetLogger(logger *slog.Logger)
```

**Impact:** 🔴 **BREAKING CHANGE**
- Must use Go's standard `log/slog` package
- Logger interface completely removed
- All logging code needs updating

---

### 3. **Router: From Router to DefaultRouter**

**v4:**
```go
type Router struct { ... }

func NewRouter(e *Echo) *Router
func (e *Echo) Router() *Router
```

**v5:**
```go
type DefaultRouter struct { ... }

func NewRouter(config RouterConfig) *DefaultRouter
func (e *Echo) Router() Router  // Returns interface
```

**Changes:**
- New `Router` interface introduced
- `DefaultRouter` is the concrete implementation
- `NewRouter()` now takes `RouterConfig` instead of `*Echo`
- Added `NewConcurrentRouter(r Router) Router` for thread-safe routing

---

### 4. **Route Return Types Changed**

**v4:**
```go
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Any(path string, h HandlerFunc, m ...MiddlewareFunc) []*Route
func (e *Echo) Routes() []*Route
```

**v5:**
```go
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo
func (e *Echo) Any(path string, h HandlerFunc, m ...MiddlewareFunc) RouteInfo
func (e *Echo) Match(...) Routes  // Returns Routes type
func (e *Echo) Router() Router  // Returns interface
```

**New Types:**
```go
type RouteInfo struct {
    Name       string
    Method     string
    Path       string
    Parameters []string
}

type Routes []RouteInfo  // Collection with helper methods
```

**Impact:** 🔴 **BREAKING CHANGE**
- Route registration methods return `RouteInfo` instead of `*Route`
- New `Routes` collection type with filtering methods
- `Route` struct still exists but used differently

---

### 5. **Response Type Changed**

**v4:**
```go
func (c Context) Response() *Response
type Response struct {
    Writer http.ResponseWriter
    Status int
    Size   int64
    Committed bool
}
func NewResponse(w http.ResponseWriter, e *Echo) *Response
```

**v5:**
```go
func (c *Context) Response() http.ResponseWriter
type Response struct {
    http.ResponseWriter  // Embedded
    Status    int
    Size      int64
    Committed bool
}
func NewResponse(w http.ResponseWriter, logger *slog.Logger) *Response
func UnwrapResponse(rw http.ResponseWriter) (*Response, error)
```

**Changes:**
- Context.Response() returns `http.ResponseWriter` instead of `*Response`
- Response now embeds `http.ResponseWriter`
- NewResponse takes `*slog.Logger` instead of `*Echo`
- New `UnwrapResponse()` helper function

---

### 6. **HTTPError Simplified**

**v4:**
```go
type HTTPError struct {
    Internal error
    Message  interface{}  // Can be any type
    Code     int
}

func NewHTTPError(code int, message ...interface{}) *HTTPError
```

**v5:**
```go
type HTTPError struct {
    Code    int
    Message string  // Now string only
    // Has unexported fields (Internal moved)
}

func NewHTTPError(code int, message string) *HTTPError
func (he HTTPError) Wrap(err error) error  // New method
func (he *HTTPError) StatusCode() int      // Implements HTTPStatusCoder
```

**Changes:**
- `Message` field changed from `interface{}` to `string`
- `NewHTTPError()` now takes `string` instead of `...interface{}`
- Added `HTTPStatusCoder` interface and `StatusCode()` method
- Added `Wrap(err error)` method for error wrapping

---

### 7. **HTTPErrorHandler Signature Changed**

**v4:**
```go
type HTTPErrorHandler func(err error, c Context)

func (e *Echo) DefaultHTTPErrorHandler(err error, c Context)
```

**v5:**
```go
type HTTPErrorHandler func(c *Context, err error)  // Parameters swapped!

func DefaultHTTPErrorHandler(exposeError bool) HTTPErrorHandler  // Now a factory
```

**Impact:** 🔴 **BREAKING CHANGE**
- Parameter order reversed: `(c *Context, err error)` instead of `(err error, c Context)`
- DefaultHTTPErrorHandler is now a factory function that returns HTTPErrorHandler
- Takes `exposeError` bool to control error message exposure

---

## Notable API Changes in v5

### 1. **Generic Parameter Extraction Functions (Updated Signatures)**

These helpers keep the same generic API but now accept `*Context`, and the
form helpers are renamed from `FormParam*` to `FormValue*`:

```go
// Query Parameters
func QueryParam[T any](c *Context, key string, opts ...any) (T, error)
func QueryParamOr[T any](c *Context, key string, defaultValue T, opts ...any) (T, error)
func QueryParams[T any](c *Context, key string, opts ...any) ([]T, error)
func QueryParamsOr[T any](c *Context, key string, defaultValue []T, opts ...any) ([]T, error)

// Path Parameters
func PathParam[T any](c *Context, paramName string, opts ...any) (T, error)
func PathParamOr[T any](c *Context, paramName string, defaultValue T, opts ...any) (T, error)

// Form Values
func FormValue[T any](c *Context, key string, opts ...any) (T, error)
func FormValueOr[T any](c *Context, key string, defaultValue T, opts ...any) (T, error)
func FormValues[T any](c *Context, key string, opts ...any) ([]T, error)
func FormValuesOr[T any](c *Context, key string, defaultValue []T, opts ...any) ([]T, error)

// Generic Parsing
func ParseValue[T any](value string, opts ...any) (T, error)
func ParseValueOr[T any](value string, defaultValue T, opts ...any) (T, error)
func ParseValues[T any](values []string, opts ...any) ([]T, error)
func ParseValuesOr[T any](values []string, defaultValue []T, opts ...any) ([]T, error)
```

`FormParam*` was renamed to `FormValue*`; the rest keep names but now take `*Context`.

**Supported Types:**
- bool, string
- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64
- time.Time, time.Duration
- BindUnmarshaler, encoding.TextUnmarshaler, json.Unmarshaler

**Example Usage:**
```go
// v5 - Type-safe parameter binding
id, err := echo.PathParam[int](c, "id")
page, err := echo.QueryParamOr[int](c, "page", 1)
tags, err := echo.QueryParams[string](c, "tags")
```

---

### 2. **Context Store Helpers Now Use `*Context`**

```go
// Type-safe context value retrieval
func ContextGet[T any](c *Context, key string) (T, error)
func ContextGetOr[T any](c *Context, key string, defaultValue T) (T, error)

// Error types
var ErrNonExistentKey = errors.New("non existent key")
var ErrInvalidKeyType = errors.New("invalid key type")
```

These helpers existed in v4 with `Context` and now accept `*Context`.

**Example:**
```go
// v5
user, err := echo.ContextGet[*User](c, "user")
count, err := echo.ContextGetOr[int](c, "count", 0)
```

---

### 3. **PathValues Type**

New structured path parameter handling:

```go
type PathValue struct {
    Name  string
    Value string
}

type PathValues []PathValue

func (p PathValues) Get(name string) (string, bool)
func (p PathValues) GetOr(name string, defaultValue string) string

// Context methods
func (c *Context) PathValues() PathValues
func (c *Context) SetPathValues(pathValues PathValues)
```

---

### 4. **Time Parsing Options**

```go
type TimeLayout string

const (
    TimeLayoutUnixTime      = TimeLayout("UnixTime")
    TimeLayoutUnixTimeMilli = TimeLayout("UnixTimeMilli")
    TimeLayoutUnixTimeNano  = TimeLayout("UnixTimeNano")
)

type TimeOpts struct {
    Layout          TimeLayout
    ParseInLocation *time.Location
    ToInLocation    *time.Location
}
```

---

### 5. **StartConfig for Server Configuration**

```go
type StartConfig struct {
    Address         string
    HideBanner      bool
    HidePort        bool
    CertFilesystem  fs.FS
    TLSConfig       *tls.Config
    ListenerNetwork string
    ListenerAddrFunc func(addr net.Addr)
    GracefulTimeout  time.Duration
    OnShutdownError  func(err error)
    BeforeServeFunc  func(s *http.Server) error
}

func (sc StartConfig) Start(ctx context.Context, h http.Handler) error
func (sc StartConfig) StartTLS(ctx context.Context, h http.Handler, certFile, keyFile any) error
```

**Example:**
```go
// v5 - More control over server startup
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

sc := echo.StartConfig{
    Address:         ":8080",
    GracefulTimeout: 10 * time.Second,
}
if err := sc.Start(ctx, e); err != nil {
    log.Fatal(err)
}
```

---

### 6. **Echo Config and Constructors**

```go
type Config struct {
    // Configuration for Echo (logger, binder, renderer, etc.)
}

func NewWithConfig(config Config) *Echo
```

This adds a configuration struct for creating an `Echo` instance without
mutating fields after `New()`.

---

### 7. **Enhanced Routing Features**

```go
// New route methods
func (e *Echo) AddRoute(route Route) (RouteInfo, error)
func (e *Echo) Middlewares() []MiddlewareFunc
func (e *Echo) PreMiddlewares() []MiddlewareFunc
type AddRouteError struct{ ... }

// Routes collection with filters
type Routes []RouteInfo

func (r Routes) Clone() Routes
func (r Routes) FilterByMethod(method string) (Routes, error)
func (r Routes) FilterByName(name string) (Routes, error)
func (r Routes) FilterByPath(path string) (Routes, error)
func (r Routes) FindByMethodPath(method string, path string) (RouteInfo, error)
func (r Routes) Reverse(routeName string, pathValues ...any) (string, error)

// RouteInfo operations
func (r RouteInfo) Clone() RouteInfo
func (r RouteInfo) Reverse(pathValues ...any) string
```

---

### 8. **Middleware Configuration Interface**

```go
type MiddlewareConfigurator interface {
    ToMiddleware() (MiddlewareFunc, error)
}
```

Allows middleware configs to be converted to middleware without panicking.

---

### 9. **New Context Methods**

```go
// v5 additions
func (c *Context) FileFS(file string, filesystem fs.FS) error
func (c *Context) FormValueOr(name, defaultValue string) string
func (c *Context) InitializeRoute(ri *RouteInfo, pathValues *PathValues)
func (c *Context) ParamOr(name, defaultValue string) string
func (c *Context) QueryParamOr(name, defaultValue string) string
func (c *Context) RouteInfo() RouteInfo
```

---

### 10. **Virtual Host Support**

```go
func NewVirtualHostHandler(vhosts map[string]*Echo) *Echo
```

Creates an Echo instance that routes requests to different Echo instances based on host.

---

### 11. **New Binder Functions**

```go
func BindBody(c *Context, target any) error
func BindHeaders(c *Context, target any) error
func BindPathValues(c *Context, target any) error  // Renamed from BindPathParams
func BindQueryParams(c *Context, target any) error
```

Top-level binding functions that work with `*Context`.

---

### 12. **New echotest Package**

```go
package echotest // import "github.com/labstack/echo/v5/echotest"

func LoadBytes(t *testing.T, name string, opts ...loadBytesOpts) []byte
func TrimNewlineEnd(bytes []byte) []byte
type ContextConfig struct{ ... }
type MultipartForm struct{ ... }
type MultipartFormFile struct{ ... }
```

Helpers for loading fixtures and constructing test contexts.

---

## Removed APIs in v5

### Constants

```go
// v4 - Removed in v5
const CONNECT = http.MethodConnect  // Use http.MethodConnect directly
```

**Reason:** Deprecated in v4, use stdlib `http.Method*` constants instead.

---

### Constants Added in v5

```go
// v5 additions
const (
    NotFoundRouteName = "echo_route_not_found_name"
)
```

---

### Error Variable Changes

**v4 exports:**
```go
ErrBadRequest
ErrInvalidKeyType
ErrNonExistentKey
```

**v5 exports:**
```go
ErrBadRequest  // Now backed by unexported httpError type
ErrValidatorNotRegistered  // New
ErrInvalidKeyType
ErrNonExistentKey
```

**Reason:** v5 centralizes on `NewHTTPError(code, message)` rather than a broad set
of predefined HTTP error variables.

---

### Functions Removed

```go
// v4 - Removed in v5
func GetPath(r *http.Request) string  // Use r.URL.Path or r.URL.RawPath
```

### Variables Removed

```go
// v4 - Removed in v5
var MethodNotAllowedHandler = func(c Context) error { ... }
var NotFoundHandler = func(c Context) error { ... }
```

### Functions Renamed

```go
// v4
func FormParam[T any](c Context, key string, opts ...any) (T, error)
func FormParamOr[T any](c Context, key string, defaultValue T, opts ...any) (T, error)
func FormParams[T any](c Context, key string, opts ...any) ([]T, error)
func FormParamsOr[T any](c Context, key string, defaultValue []T, opts ...any) ([]T, error)

// v5
func FormValue[T any](c *Context, key string, opts ...any) (T, error)
func FormValueOr[T any](c *Context, key string, defaultValue T, opts ...any) (T, error)
func FormValues[T any](c *Context, key string, opts ...any) ([]T, error)
func FormValuesOr[T any](c *Context, key string, defaultValue []T, opts ...any) ([]T, error)
```

---

### Type Methods Removed/Changed

**Echo struct changes:**
```go
// v4 fields removed in v5
type Echo struct {
    StdLogger        *stdLog.Logger  // Removed
    Server           *http.Server    // Removed (use StartConfig)
    TLSServer        *http.Server    // Removed (use StartConfig)
    Listener         net.Listener    // Removed (use StartConfig)
    TLSListener      net.Listener    // Removed (use StartConfig)
    AutoTLSManager   autocert.Manager // Removed
    ListenerNetwork  string          // Removed
    OnAddRouteHandler func(...)      // Changed to OnAddRoute
    DisableHTTP2      bool           // Removed (use StartConfig)
    Debug             bool           // Removed
    HideBanner        bool           // Removed (use StartConfig)
    HidePort          bool           // Removed (use StartConfig)
}

// v5 Echo struct (simplified)
type Echo struct {
    Binder           Binder
    Filesystem       fs.FS  // NEW
    Renderer         Renderer
    Validator        Validator
    JSONSerializer   JSONSerializer
    IPExtractor      IPExtractor
    OnAddRoute       func(route Route) error  // Simplified
    HTTPErrorHandler HTTPErrorHandler
    Logger           *slog.Logger  // Changed from Logger interface
}
```

---

**Context interface → struct:**
```go
// v4
type Context interface {
    // Had: SetResponse(*Response)
    Response() *Response

    // Had: ParamNames(), SetParamNames(), ParamValues(), SetParamValues()
    // These are removed in v5 (use PathValues() instead)
}

// v5
type Context struct {
    // Concrete struct with unexported fields
}

func (c *Context) Response() http.ResponseWriter  // Changed return type
func (c *Context) PathValues() PathValues         // Replaces ParamNames/Values
```

---

**Types removed:**
```go
// v4
type Map map[string]interface{}
```

**Group changes:**
```go
// v4
func (g *Group) File(path, file string)  // No return value
func (g *Group) Static(pathPrefix, fsRoot string)  // No return value
func (g *Group) StaticFS(pathPrefix string, filesystem fs.FS)  // No return value

// v5
func (g *Group) File(path, file string, middleware ...MiddlewareFunc) RouteInfo
func (g *Group) Static(pathPrefix, fsRoot string, middleware ...MiddlewareFunc) RouteInfo
func (g *Group) StaticFS(pathPrefix string, filesystem fs.FS, middleware ...MiddlewareFunc) RouteInfo
```

Now return `RouteInfo` and accept middleware.

---

### Value Binder Factory Name Changes

```go
// v4
func PathParamsBinder(c Context) *ValueBinder
func QueryParamsBinder(c Context) *ValueBinder
func FormFieldBinder(c Context) *ValueBinder

// v5
func PathValuesBinder(c *Context) *ValueBinder  // Renamed
func QueryParamsBinder(c *Context) *ValueBinder
func FormFieldBinder(c *Context) *ValueBinder
```

---

## Type Signature Changes

### Binder Interface

```go
// v4
type Binder interface {
    Bind(i interface{}, c Context) error
}

// v5
type Binder interface {
    Bind(c *Context, target any) error  // Parameters swapped!
}
```

---

### DefaultBinder Methods

```go
// v4
func (b *DefaultBinder) Bind(i interface{}, c Context) error
func (b *DefaultBinder) BindBody(c Context, i interface{}) error
func (b *DefaultBinder) BindPathParams(c Context, i interface{}) error

// v5
func (b *DefaultBinder) Bind(c *Context, target any) error  // Swapped params
// BindBody, BindPathParams, etc. are now top-level functions
```

---

### JSONSerializer Interface

```go
// v4
type JSONSerializer interface {
    Serialize(c Context, i interface{}, indent string) error
    Deserialize(c Context, i interface{}) error
}

// v5
type JSONSerializer interface {
    Serialize(c *Context, target any, indent string) error
    Deserialize(c *Context, target any) error
}
```

---

### Renderer Interface

```go
// v4
type Renderer interface {
    Render(io.Writer, string, interface{}, Context) error
}

// v5
type Renderer interface {
    Render(c *Context, w io.Writer, templateName string, data any) error
}
```

Parameters reordered with Context first.

---

### NewBindingError

```go
// v4
func NewBindingError(sourceParam string, values []string, message interface{}, internalError error) error

// v5
func NewBindingError(sourceParam string, values []string, message string, err error) error
```

Message parameter changed from `interface{}` to `string`.

---

### HandlerName

```go
// v5 only
func HandlerName(h HandlerFunc) string
```

New utility function to get handler function name.

---

## Middleware Package Changes

### Signature and Type Updates

```go
// CORS now accepts optional allow-origins
func CORS(allowOrigins ...string) echo.MiddlewareFunc

// BodyLimit now accepts bytes
func BodyLimit(limitBytes int64) echo.MiddlewareFunc

// DefaultSkipper now uses *echo.Context
func DefaultSkipper(c *echo.Context) bool

// Trailing slash configs renamed/split
func AddTrailingSlashWithConfig(config AddTrailingSlashConfig) echo.MiddlewareFunc
func RemoveTrailingSlashWithConfig(config RemoveTrailingSlashConfig) echo.MiddlewareFunc
type AddTrailingSlashConfig struct{ ... }
type RemoveTrailingSlashConfig struct{ ... }

// Auth + extractor signatures now use *echo.Context and add ExtractorSource
type BasicAuthValidator func(c *echo.Context, user string, password string) (bool, error)
type Extractor func(c *echo.Context) (string, error)
type ExtractorSource string
type KeyAuthValidator func(c *echo.Context, key string, source ExtractorSource) (bool, error)
type KeyAuthErrorHandler func(c *echo.Context, err error) error

// BodyDump handler now includes err
type BodyDumpHandler func(c *echo.Context, reqBody []byte, resBody []byte, err error)

// ValuesExtractor now returns extractor source and CreateExtractors takes a limit
type ValuesExtractor func(c *echo.Context) ([]string, ExtractorSource, error)
func CreateExtractors(lookups string, limit uint) ([]ValuesExtractor, error)
type ValueExtractorError struct{ ... }

// New constants
const KB = 1024

// Rate limiter store now takes a float64 limit
func NewRateLimiterMemoryStore(rateLimit float64) (store *RateLimiterMemoryStore)
```

### Added Middleware Exports

```go
var ErrInvalidKey = echo.NewHTTPError(http.StatusUnauthorized, "invalid key")
var ErrKeyMissing = echo.NewHTTPError(http.StatusUnauthorized, "missing key")
var RedirectHTTPSConfig = RedirectConfig{ ... }
var RedirectHTTPSWWWConfig = RedirectConfig{ ... }
var RedirectNonHTTPSWWWConfig = RedirectConfig{ ... }
var RedirectNonWWWConfig = RedirectConfig{ ... }
var RedirectWWWConfig = RedirectConfig{ ... }
```

### Removed/Consolidated Middleware Exports

```go
// Removed in v5
func Logger() echo.MiddlewareFunc
func LoggerWithConfig(config LoggerConfig) echo.MiddlewareFunc
func Timeout() echo.MiddlewareFunc
func TimeoutWithConfig(config TimeoutConfig) echo.MiddlewareFunc
type ErrKeyAuthMissing struct{ ... }
type CSRFErrorHandler func(err error, c echo.Context) error
type LoggerConfig struct{ ... }
type LogErrorFunc func(c echo.Context, err error, stack []byte) error
type TargetProvider interface{ ... }
type TrailingSlashConfig struct{ ... }
type TimeoutConfig struct{ ... }
```

Also removed defaults: `DefaultBasicAuthConfig`, `DefaultBodyDumpConfig`, `DefaultBodyLimitConfig`,
`DefaultCORSConfig`, `DefaultDecompressConfig`, `DefaultGzipConfig`, `DefaultLoggerConfig`,
`DefaultRedirectConfig`, `DefaultRequestIDConfig`, `DefaultRewriteConfig`, `DefaultTimeoutConfig`,
`DefaultTrailingSlashConfig`.

---

## Router Interface Changes

### v4 Router (Concrete Struct)

```go
type Router struct { ... }

func NewRouter(e *Echo) *Router
func (r *Router) Add(method, path string, h HandlerFunc)
func (r *Router) Find(method, path string, c Context)
func (r *Router) Reverse(name string, params ...interface{}) string
func (r *Router) Routes() []*Route
```

### v5 Router (Interface + DefaultRouter)

```go
type Router interface {
    Add(routable Route) (RouteInfo, error)
    Remove(method string, path string) error
    Routes() Routes
    Route(c *Context) HandlerFunc
}

type DefaultRouter struct { ... }

func NewRouter(config RouterConfig) *DefaultRouter
func NewConcurrentRouter(r Router) Router  // NEW

type RouterConfig struct {
    NotFoundHandler           HandlerFunc
    MethodNotAllowedHandler   HandlerFunc
    OptionsMethodHandler      HandlerFunc
    AllowOverwritingRoute     bool
    UnescapePathParamValues   bool
    UseEscapedPathForMatching bool
}
```

**Key Changes:**
- Router is now an interface
- DefaultRouter is the concrete implementation
- Add() returns `(RouteInfo, error)` instead of being void
- New `Remove()` method
- New `Route()` method replaces `Find()`
- Configuration through `RouterConfig`

---

## Echo Instance Method Changes

### Route Registration

```go
// v4
func (e *Echo) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route

// v5
func (e *Echo) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) RouteInfo
func (e *Echo) AddRoute(route Route) (RouteInfo, error)  // NEW
```

### Static File Serving

```go
// v4
func (e *Echo) Static(pathPrefix, fsRoot string) *Route
func (e *Echo) StaticFS(pathPrefix string, filesystem fs.FS) *Route
func (e *Echo) File(path, file string, m ...MiddlewareFunc) *Route
func (e *Echo) FileFS(path, file string, filesystem fs.FS, m ...MiddlewareFunc) *Route

// v5
func (e *Echo) Static(pathPrefix, fsRoot string, middleware ...MiddlewareFunc) RouteInfo
func (e *Echo) StaticFS(pathPrefix string, filesystem fs.FS, middleware ...MiddlewareFunc) RouteInfo
func (e *Echo) File(path, file string, middleware ...MiddlewareFunc) RouteInfo
func (e *Echo) FileFS(path, file string, filesystem fs.FS, m ...MiddlewareFunc) RouteInfo
```

Return type changed from `*Route` to `RouteInfo`.

### Server Management

```go
// v4
func (e *Echo) Start(address string) error
func (e *Echo) StartTLS(address string, certFile, keyFile interface{}) error
func (e *Echo) StartAutoTLS(address string) error
func (e *Echo) StartH2CServer(address string, h2s *http2.Server) error
func (e *Echo) StartServer(s *http.Server) error
func (e *Echo) Shutdown(ctx context.Context) error
func (e *Echo) Close() error
func (e *Echo) ListenerAddr() net.Addr
func (e *Echo) TLSListenerAddr() net.Addr
func (e *Echo) DefaultHTTPErrorHandler(err error, c Context)

// v5
func (e *Echo) Start(address string) error  // Simplified
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request)

// Removed: StartTLS, StartAutoTLS, StartH2CServer, StartServer
// Use StartConfig instead for advanced server configuration
// Removed: Shutdown, Close, ListenerAddr, TLSListenerAddr
// Removed: DefaultHTTPErrorHandler (now a top-level factory function)
```

**v5 provides** `StartConfig` type for all advanced server configuration.

### Router Access

```go
// v4
func (e *Echo) Router() *Router
func (e *Echo) Routers() map[string]*Router  // For multi-host
func (e *Echo) Routes() []*Route
func (e *Echo) Reverse(name string, params ...interface{}) string
func (e *Echo) URI(handler HandlerFunc, params ...interface{}) string
func (e *Echo) URL(h HandlerFunc, params ...interface{}) string
func (e *Echo) Host(name string, m ...MiddlewareFunc) *Group

// v5
func (e *Echo) Router() Router  // Returns interface
// Removed: Routers(), Reverse(), URI(), URL(), Host()
// Use router.Routes() and Routes.Reverse() instead
```

---

## NewContext Changes

```go
// v4
func (e *Echo) NewContext(r *http.Request, w http.ResponseWriter) Context
func NewResponse(w http.ResponseWriter, e *Echo) *Response

// v5
func (e *Echo) NewContext(r *http.Request, w http.ResponseWriter) *Context
func NewContext(r *http.Request, w http.ResponseWriter, opts ...any) *Context  // Standalone
func NewResponse(w http.ResponseWriter, logger *slog.Logger) *Response
```

---

## Migration Guide Summary

If you are using Linux you can migrate easier parts like that:
```bash
find . -type f -name "*.go" -exec sed -i 's/ echo.Context/ *echo.Context/g' {} +
find . -type f -name "*.go" -exec sed -i 's/echo\/v4/echo\/v5/g' {} +
```
or in your favorite IDE

Replace all:
1. ` echo.Context` -> ` *echo.Context`
2. `echo/v4` -> `echo/v5`


### 1. Update All Handler Signatures

```go
// Before
func MyHandler(c echo.Context) error { ... }

// After
func MyHandler(c *echo.Context) error { ... }
```

### 2. Update Logger Usage

```go
// Before
e.Logger.Info("Server started")
c.Logger().Error("Something went wrong")

// After
e.Logger.Info("Server started")
c.Logger().Error("Something went wrong")  // Same API, different logger
```

### 3. Use Type-Safe Parameter Extraction

```go
// Before
idStr := c.Param("id")
id, err := strconv.Atoi(idStr)

// After
id, err := echo.PathParam[int](c, "id")
```

### 4. Update Error Handler

```go
// Before
e.HTTPErrorHandler = func(err error, c echo.Context) {
    // handle error
}

// After
e.HTTPErrorHandler = func(c *echo.Context, err error) {  // Swapped!
    // handle error
}

// Or use factory
e.HTTPErrorHandler = echo.DefaultHTTPErrorHandler(true)  // exposeError=true
```

### 5. Update Server Startup

```go
// Before
e.Start(":8080")
e.StartTLS(":443", "cert.pem", "key.pem")

// After
// Simple
e.Start(":8080")

// Advanced with graceful shutdown
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
sc := echo.StartConfig{Address: ":8080"}
sc.Start(ctx, e)
```

### 6. Update Route Info Access

```go
// Before
routes := e.Routes()
for _, r := range routes {
    fmt.Println(r.Method, r.Path)
}

// After
routes := e.Router().Routes()
for _, r := range routes {
    fmt.Println(r.Method, r.Path)
}
```

### 7. Update HTTPError Creation

```go
// Before
return echo.NewHTTPError(400, "invalid request", someDetail)

// After
return echo.NewHTTPError(400, "invalid request")
```

### 8. Update Custom Binder

```go
// Before
type MyBinder struct{}
func (b *MyBinder) Bind(i interface{}, c echo.Context) error { ... }

// After
type MyBinder struct{}
func (b *MyBinder) Bind(c *echo.Context, target any) error { ... }  // Swapped!
```

### 9. Path Parameters

```go
// Before
names := c.ParamNames()
values := c.ParamValues()

// After
pathValues := c.PathValues()
for _, pv := range pathValues {
    fmt.Println(pv.Name, pv.Value)
}
```

### 10. Response Access

```go
// Before
resp := c.Response()
resp.Header().Set("X-Custom", "value")

// After
c.Response().Header().Set("X-Custom", "value")  // Returns http.ResponseWriter

// To get *echo.Response
resp, err := echo.UnwrapResponse(c.Response())
```

### Go Version Requirements

- **v4**: Go 1.24.0 (per `go.mod`)
- **v5**: Go 1.25.0 (per `go.mod`)

---

**Generated by comparing `go doc` output from master (v4.15.0) and v5 (v5.0.0-alpha) branches**
