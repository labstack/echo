package echo

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
)

// Context represents the context of the current HTTP request. It holds request and
// response objects, path, path parameters, data and registered handler.
type Context interface {
	// Request returns `*http.Request`.
	Request() *http.Request

	// SetRequest sets `*http.Request`.
	SetRequest(r *http.Request)

	// SetResponse sets `*Response`.
	SetResponse(r *Response)

	// Response returns `*Response`.
	Response() *Response

	// IsTLS returns true if HTTP connection is TLS otherwise false.
	IsTLS() bool

	// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
	IsWebSocket() bool

	// Scheme returns the HTTP protocol scheme, `http` or `https`.
	Scheme() string

	// RealIP returns the client's network address based on `X-Forwarded-For`
	// or `X-Real-IP` request header.
	// The behavior can be configured using `Echo#IPExtractor`.
	RealIP() string

	// RouteInfo returns current request route information. Method, Path, Name and params if they exist for matched route.
	// In case of 404 (route not found) and 405 (method not allowed) RouteInfo returns generic struct for these cases.
	RouteInfo() RouteInfo

	// Path returns the registered path for the handler.
	Path() string

	// PathParam returns path parameter by name.
	PathParam(name string) string

	// PathParamDefault returns the path parameter or default value for the provided name.
	//
	// Notes for DefaultRouter implementation:
	// Path parameter could be empty for cases like that:
	// * route `/release-:version/bin` and request URL is `/release-/bin`
	// * route `/api/:version/image.jpg` and request URL is `/api//image.jpg`
	// but not when path parameter is last part of route path
	// * route `/download/file.:ext` will not match request `/download/file.`
	PathParamDefault(name string, defaultValue string) string

	// PathParams returns path parameter values.
	PathParams() PathParams

	// SetPathParams sets path parameters for current request.
	SetPathParams(params PathParams)

	// QueryParam returns the query param for the provided name.
	QueryParam(name string) string

	// QueryParamDefault returns the query param or default value for the provided name.
	QueryParamDefault(name, defaultValue string) string

	// QueryParams returns the query parameters as `url.Values`.
	QueryParams() url.Values

	// QueryString returns the URL query string.
	QueryString() string

	// FormValue returns the form field value for the provided name.
	FormValue(name string) string

	// FormValueDefault returns the form field value or default value for the provided name.
	FormValueDefault(name, defaultValue string) string

	// FormValues returns the form field values as `url.Values`.
	FormValues() (url.Values, error)

	// FormFile returns the multipart form file for the provided name.
	FormFile(name string) (*multipart.FileHeader, error)

	// MultipartForm returns the multipart form.
	MultipartForm() (*multipart.Form, error)

	// Cookie returns the named cookie provided in the request.
	Cookie(name string) (*http.Cookie, error)

	// SetCookie adds a `Set-Cookie` header in HTTP response.
	SetCookie(cookie *http.Cookie)

	// Cookies returns the HTTP cookies sent with the request.
	Cookies() []*http.Cookie

	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, val interface{})

	// Bind binds path params, query params and the request body into provided type `i`. The default binder
	// binds body based on Content-Type header.
	Bind(i interface{}) error

	// Validate validates provided `i`. It is usually called after `Context#Bind()`.
	// Validator must be registered using `Echo#Validator`.
	Validate(i interface{}) error

	// Render renders a template with data and sends a text/html response with status
	// code. Renderer must be registered using `Echo.Renderer`.
	Render(code int, name string, data interface{}) error

	// HTML sends an HTTP response with status code.
	HTML(code int, html string) error

	// HTMLBlob sends an HTTP blob response with status code.
	HTMLBlob(code int, b []byte) error

	// String sends a string response with status code.
	String(code int, s string) error

	// JSON sends a JSON response with status code.
	JSON(code int, i interface{}) error

	// JSONPretty sends a pretty-print JSON with status code.
	JSONPretty(code int, i interface{}, indent string) error

	// JSONBlob sends a JSON blob response with status code.
	JSONBlob(code int, b []byte) error

	// JSONP sends a JSONP response with status code. It uses `callback` to construct
	// the JSONP payload.
	JSONP(code int, callback string, i interface{}) error

	// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
	// to construct the JSONP payload.
	JSONPBlob(code int, callback string, b []byte) error

	// XML sends an XML response with status code.
	XML(code int, i interface{}) error

	// XMLPretty sends a pretty-print XML with status code.
	XMLPretty(code int, i interface{}, indent string) error

	// XMLBlob sends an XML blob response with status code.
	XMLBlob(code int, b []byte) error

	// Blob sends a blob response with status code and content type.
	Blob(code int, contentType string, b []byte) error

	// Stream sends a streaming response with status code and content type.
	Stream(code int, contentType string, r io.Reader) error

	// File sends a response with the content of the file.
	File(file string) error

	// FileFS sends a response with the content of the file from given filesystem.
	FileFS(file string, filesystem fs.FS) error

	// Attachment sends a response as attachment, prompting client to save the
	// file.
	Attachment(file string, name string) error

	// Inline sends a response as inline, opening the file in the browser.
	Inline(file string, name string) error

	// NoContent sends a response with no body and a status code.
	NoContent(code int) error

	// Redirect redirects the request to a provided URL with status code.
	Redirect(code int, url string) error

	// Error invokes the registered global HTTP error handler. Generally used by middleware.
	// A side-effect of calling global error handler is that now Response has been committed (sent to the client) and
	// middlewares up in chain can not change Response status code or Response body anymore.
	//
	// Avoid using this method in handlers as no middleware will be able to effectively handle errors after that.
	// Instead of calling this method in handler return your error and let it be handled by middlewares or global error handler.
	Error(err error)

	// Echo returns the `Echo` instance.
	//
	// WARNING: Remember that Echo public fields and methods are coroutine safe ONLY when you are NOT mutating them
	// anywhere in your code after Echo server has started.
	Echo() *Echo
}

// ServableContext is interface that Echo context implementation must implement to be usable in middleware/handlers and
// be able to be routed by Router.
type ServableContext interface {
	Context         // minimal set of methods for middlewares and handler
	RoutableContext // minimal set for routing. These methods should not be accessed in middlewares/handlers

	// Reset resets the context after request completes. It must be called along
	// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
	// See `Echo#ServeHTTP()`
	Reset(r *http.Request, w http.ResponseWriter)
}

const (
	// ContextKeyHeaderAllow is set by Router for getting value for `Allow` header in later stages of handler call chain.
	// Allow header is mandatory for status 405 (method not found) and useful for OPTIONS method requests.
	// It is added to context only when Router does not find matching method handler for request.
	ContextKeyHeaderAllow = "echo_header_allow"
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
	defaultIndent = "  "
)

// DefaultContext is default implementation of Context interface and can be embedded into structs to compose
// new Contexts with extended/modified behaviour.
type DefaultContext struct {
	request  *http.Request
	response *Response

	route RouteInfo
	path  string

	// pathParams holds path/uri parameters determined by Router. Lifecycle is handled by Echo to reduce allocations.
	pathParams *PathParams
	// currentParams hold path parameters set by non-Echo implementation (custom middlewares, handlers) during the lifetime of Request.
	// Lifecycle is not handle by Echo and could have excess allocations per served Request
	currentParams PathParams

	query url.Values
	store Map
	echo  *Echo
	lock  sync.RWMutex
}

// NewDefaultContext creates new instance of DefaultContext.
// Argument pathParamAllocSize must be value that is stored in Echo.contextPathParamAllocSize field and is used
// to preallocate PathParams slice.
func NewDefaultContext(e *Echo, pathParamAllocSize int) *DefaultContext {
	p := make(PathParams, pathParamAllocSize)
	return &DefaultContext{
		pathParams: &p,
		store:      make(Map),
		echo:       e,
	}
}

// Reset resets the context after request completes. It must be called along
// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
// See `Echo#ServeHTTP()`
func (c *DefaultContext) Reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response.reset(w)
	c.query = nil
	c.store = nil

	c.route = nil
	c.path = ""
	// NOTE: Don't reset because it has to have length of c.echo.contextPathParamAllocSize at all times
	*c.pathParams = (*c.pathParams)[:0]
	c.currentParams = nil
}

func (c *DefaultContext) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

// Request returns `*http.Request`.
func (c *DefaultContext) Request() *http.Request {
	return c.request
}

// SetRequest sets `*http.Request`.
func (c *DefaultContext) SetRequest(r *http.Request) {
	c.request = r
}

// Response returns `*Response`.
func (c *DefaultContext) Response() *Response {
	return c.response
}

// SetResponse sets `*Response`.
func (c *DefaultContext) SetResponse(r *Response) {
	c.response = r
}

// IsTLS returns true if HTTP connection is TLS otherwise false.
func (c *DefaultContext) IsTLS() bool {
	return c.request.TLS != nil
}

// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
func (c *DefaultContext) IsWebSocket() bool {
	upgrade := c.request.Header.Get(HeaderUpgrade)
	return strings.EqualFold(upgrade, "websocket")
}

// Scheme returns the HTTP protocol scheme, `http` or `https`.
func (c *DefaultContext) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.IsTLS() {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := c.request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := c.request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := c.request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

// RealIP returns the client's network address based on `X-Forwarded-For`
// or `X-Real-IP` request header.
// The behavior can be configured using `Echo#IPExtractor`.
func (c *DefaultContext) RealIP() string {
	if c.echo != nil && c.echo.IPExtractor != nil {
		return c.echo.IPExtractor(c.request)
	}
	// Fall back to legacy behavior
	if ip := c.request.Header.Get(HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			xffip := strings.TrimSpace(ip[:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return ip
	}
	if ip := c.request.Header.Get(HeaderXRealIP); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.request.RemoteAddr)
	return ra
}

// Path returns the registered path for the handler.
func (c *DefaultContext) Path() string {
	return c.path
}

// SetPath sets the registered path for the handler.
func (c *DefaultContext) SetPath(p string) {
	c.path = p
}

// RouteInfo returns current request route information. Method, Path, Name and params if they exist for matched route.
// In case of 404 (route not found) and 405 (method not allowed) RouteInfo returns generic struct for these cases.
func (c *DefaultContext) RouteInfo() RouteInfo {
	return c.route
}

// SetRouteInfo sets the route info of this request to the context.
func (c *DefaultContext) SetRouteInfo(ri RouteInfo) {
	c.route = ri
}

// RawPathParams returns raw path pathParams value. Allocation of PathParams is handled by Context.
func (c *DefaultContext) RawPathParams() *PathParams {
	return c.pathParams
}

// SetRawPathParams replaces any existing param values with new values for this context lifetime (request).
//
// DO NOT USE!
// Do not set any other value than what you got from RawPathParams as allocation of PathParams is handled by Context.
// If you mess up size of pathParams size your application will panic/crash during routing
func (c *DefaultContext) SetRawPathParams(params *PathParams) {
	c.pathParams = params
}

// PathParam returns path parameter by name.
func (c *DefaultContext) PathParam(name string) string {
	if c.currentParams != nil {
		return c.currentParams.Get(name, "")
	}

	return c.pathParams.Get(name, "")
}

// PathParamDefault returns the path parameter or default value for the provided name.
func (c *DefaultContext) PathParamDefault(name, defaultValue string) string {
	return c.pathParams.Get(name, defaultValue)
}

// PathParams returns path parameter values.
func (c *DefaultContext) PathParams() PathParams {
	if c.currentParams != nil {
		return c.currentParams
	}

	result := make(PathParams, len(*c.pathParams))
	copy(result, *c.pathParams)
	return result
}

// SetPathParams sets path parameters for current request.
func (c *DefaultContext) SetPathParams(params PathParams) {
	c.currentParams = params
}

// QueryParam returns the query param for the provided name.
func (c *DefaultContext) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

// QueryParamDefault returns the query param or default value for the provided name.
// Note: QueryParamDefault does not distinguish if query had no value by that name or value was empty string
// This means URLs `/test?search=` and `/test` would both return `1` for `c.QueryParamDefault("search", "1")`
func (c *DefaultContext) QueryParamDefault(name, defaultValue string) string {
	value := c.QueryParam(name)
	if value == "" {
		value = defaultValue
	}
	return value
}

// QueryParams returns the query parameters as `url.Values`.
func (c *DefaultContext) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

// QueryString returns the URL query string.
func (c *DefaultContext) QueryString() string {
	return c.request.URL.RawQuery
}

// FormValue returns the form field value for the provided name.
func (c *DefaultContext) FormValue(name string) string {
	return c.request.FormValue(name)
}

// FormValueDefault returns the form field value or default value for the provided name.
// Note: FormValueDefault does not distinguish if form had no value by that name or value was empty string
func (c *DefaultContext) FormValueDefault(name, defaultValue string) string {
	value := c.FormValue(name)
	if value == "" {
		value = defaultValue
	}
	return value
}

// FormValues returns the form field values as `url.Values`.
func (c *DefaultContext) FormValues() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

// FormFile returns the multipart form file for the provided name.
func (c *DefaultContext) FormFile(name string) (*multipart.FileHeader, error) {
	f, fh, err := c.request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, nil
}

// MultipartForm returns the multipart form.
func (c *DefaultContext) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

// Cookie returns the named cookie provided in the request.
func (c *DefaultContext) Cookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

// SetCookie adds a `Set-Cookie` header in HTTP response.
func (c *DefaultContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

// Cookies returns the HTTP cookies sent with the request.
func (c *DefaultContext) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

// Get retrieves data from the context.
func (c *DefaultContext) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

// Set saves data in the context.
func (c *DefaultContext) Set(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

// Bind binds path params, query params and the request body into provided type `i`. The default binder
// binds body based on Content-Type header.
func (c *DefaultContext) Bind(i interface{}) error {
	return c.echo.Binder.Bind(c, i)
}

// Validate validates provided `i`. It is usually called after `Context#Bind()`.
// Validator must be registered using `Echo#Validator`.
func (c *DefaultContext) Validate(i interface{}) error {
	if c.echo.Validator == nil {
		return ErrValidatorNotRegistered
	}
	return c.echo.Validator.Validate(i)
}

// Render renders a template with data and sends a text/html response with status
// code. Renderer must be registered using `Echo.Renderer`.
func (c *DefaultContext) Render(code int, name string, data interface{}) (err error) {
	if c.echo.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = c.echo.Renderer.Render(buf, name, data, c); err != nil {
		return
	}
	return c.HTMLBlob(code, buf.Bytes())
}

// HTML sends an HTTP response with status code.
func (c *DefaultContext) HTML(code int, html string) (err error) {
	return c.HTMLBlob(code, []byte(html))
}

// HTMLBlob sends an HTTP blob response with status code.
func (c *DefaultContext) HTMLBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

// String sends a string response with status code.
func (c *DefaultContext) String(code int, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *DefaultContext) jsonPBlob(code int, callback string, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.echo.Debug || pretty {
		indent = defaultIndent
	}
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if err = c.echo.JSONSerializer.Serialize(c, i, indent); err != nil {
		return
	}
	if _, err = c.response.Write([]byte(");")); err != nil {
		return
	}
	return
}

func (c *DefaultContext) json(code int, i interface{}, indent string) error {
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.response.Status = code
	return c.echo.JSONSerializer.Serialize(c, i, indent)
}

// JSON sends a JSON response with status code.
func (c *DefaultContext) JSON(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.echo.Debug || pretty {
		indent = defaultIndent
	}
	return c.json(code, i, indent)
}

// JSONPretty sends a pretty-print JSON with status code.
func (c *DefaultContext) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

// JSONBlob sends a JSON blob response with status code.
func (c *DefaultContext) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (c *DefaultContext) JSONP(code int, callback string, i interface{}) (err error) {
	return c.jsonPBlob(code, callback, i)
}

// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
// to construct the JSONP payload.
func (c *DefaultContext) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(MIMEApplicationJavaScriptCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.response.Write(b); err != nil {
		return
	}
	_, err = c.response.Write([]byte(");"))
	return
}

func (c *DefaultContext) xml(code int, i interface{}, indent string) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	enc := xml.NewEncoder(c.response)
	if indent != "" {
		enc.Indent("", indent)
	}
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	return enc.Encode(i)
}

// XML sends an XML response with status code.
func (c *DefaultContext) XML(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.echo.Debug || pretty {
		indent = defaultIndent
	}
	return c.xml(code, i, indent)
}

// XMLPretty sends a pretty-print XML with status code.
func (c *DefaultContext) XMLPretty(code int, i interface{}, indent string) (err error) {
	return c.xml(code, i, indent)
}

// XMLBlob sends an XML blob response with status code.
func (c *DefaultContext) XMLBlob(code int, b []byte) (err error) {
	c.writeContentType(MIMEApplicationXMLCharsetUTF8)
	c.response.WriteHeader(code)
	if _, err = c.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

// Blob sends a blob response with status code and content type.
func (c *DefaultContext) Blob(code int, contentType string, b []byte) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = c.response.Write(b)
	return
}

// Stream sends a streaming response with status code and content type.
func (c *DefaultContext) Stream(code int, contentType string, r io.Reader) (err error) {
	c.writeContentType(contentType)
	c.response.WriteHeader(code)
	_, err = io.Copy(c.response, r)
	return
}

// File sends a response with the content of the file.
func (c *DefaultContext) File(file string) error {
	return fsFile(c, file, c.echo.Filesystem)
}

// FileFS serves file from given file system.
//
// When dealing with `embed.FS` use `fs := echo.MustSubFS(fs, "rootDirectory") to create sub fs which uses necessary
// prefix for directory path. This is necessary as `//go:embed assets/images` embeds files with paths
// including `assets/images` as their prefix.
func (c *DefaultContext) FileFS(file string, filesystem fs.FS) error {
	return fsFile(c, file, filesystem)
}

func fsFile(c Context, file string, filesystem fs.FS) error {
	f, err := filesystem.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.ToSlash(filepath.Join(file, indexPage)) // ToSlash is necessary for Windows. fs.Open and os.Open are different in that aspect.
		f, err = filesystem.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	ff, ok := f.(io.ReadSeeker)
	if !ok {
		return errors.New("file does not implement io.ReadSeeker")
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), ff)
	return nil
}

// Attachment sends a response as attachment, prompting client to save the file.
func (c *DefaultContext) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

// Inline sends a response as inline, opening the file in the browser.
func (c *DefaultContext) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *DefaultContext) contentDisposition(file, name, dispositionType string) error {
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

// NoContent sends a response with no body and a status code.
func (c *DefaultContext) NoContent(code int) error {
	c.response.WriteHeader(code)
	return nil
}

// Redirect redirects the request to a provided URL with status code.
func (c *DefaultContext) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.response.Header().Set(HeaderLocation, url)
	c.response.WriteHeader(code)
	return nil
}

// Error invokes the registered global HTTP error handler. Generally used by middleware.
// A side-effect of calling global error handler is that now Response has been committed (sent to the client) and
// middlewares up in chain can not change Response status code or Response body anymore.
//
// Avoid using this method in handlers as no middleware will be able to effectively handle errors after that.
// Instead of calling this method in handler return your error and let it be handled by middlewares or global error handler.
func (c *DefaultContext) Error(err error) {
	c.echo.HTTPErrorHandler(c, err)
}

// Echo returns the `Echo` instance.
func (c *DefaultContext) Echo() *Echo {
	return c.echo
}
