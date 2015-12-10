package echo

// PreRequestHandlerFunc is a func which will be called before control for
// the request is passed to the middleware and handler. The purpose is to
// provide a transparent way for a platform to hook into the process such
// as Googo AppEngine to setup it's own content and logger
type PreRequestHandlerFunc func(*Context)

var preRequestHandlerFuncs []PreRequestHandlerFunc

func RegisterPreRequestHandlerFunc(fn PreRequestHandlerFunc) {
	preRequestHandlerFuncs = append(preRequestHandlerFuncs, fn)
}
