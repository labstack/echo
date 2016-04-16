package echo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/engine"
)

// ResponseWriter can write bodies to the underlying response
type ResponseWriter struct {
	Res engine.Response
}

// Response returns the inner engine response
func (r ResponseWriter) Response() engine.Response {
	return r.Res
}

// HTML writes html out to the response with proper Content-Type.
func (r ResponseWriter) HTML(code int, html string) (err error) {
	r.Res.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	r.Res.WriteHeader(code)
	_, err = r.Res.Write([]byte(html))
	return
}

// String writes plain text to the response with proper Content-Type.
func (r ResponseWriter) String(code int, s string) (err error) {
	r.Res.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	r.Res.WriteHeader(code)
	_, err = r.Res.Write([]byte(s))
	return
}

// JSON marshals the object passed in and renders a code as well.
func (r ResponseWriter) JSON(code int, i interface{}) error {
	var byt []byte
	var err error

	byt, err = json.Marshal(i)
	if err != nil {
		return err
	}

	return r.JSONBlob(code, byt)
}

// JSONBlob writes out a JSON blob directly to the response and sets
// the correct Content-Type.
func (r ResponseWriter) JSONBlob(code int, blob []byte) (err error) {
	r.Res.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	r.Res.WriteHeader(code)
	_, err = r.Res.Write(blob)
	return err
}

// JSONP sends a JSONP response with status code. It uses `callback` to construct
// the JSONP payload.
func (r ResponseWriter) JSONP(code int, callback string, i interface{}) error {
	byt, err := json.Marshal(i)
	if err != nil {
		return err
	}
	r.Res.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	r.Res.WriteHeader(code)

	// TODO(aarondl): Is this better than what was there before?
	_, err = fmt.Fprintf(r.Res, "%s(%s);", callback, byt)
	return err
}

// XML encode i and send it to the response with a proper Content-Type.
func (r ResponseWriter) XML(code int, i interface{}) error {
	byt, err := xml.Marshal(i)
	if err != nil {
		return err
	}
	return r.XMLBlob(code, byt)
}

// XMLBlob writes a blob of XML to the response with proper Content-Type.
func (r ResponseWriter) XMLBlob(code int, blob []byte) (err error) {
	r.Res.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	r.Res.WriteHeader(code)
	if _, err = r.Res.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = r.Res.Write(blob)
	return
}

// File sends a response with the content of the file.
func (r ResponseWriter) File(request engine.Request, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return ErrNotFound
	}
	defer f.Close()

	fi, err := f.Stat()
	// TODO(aarondl): Validate why there was no error checking here
	if err != nil {
		return err
	}

	if fi.IsDir() {
		file = filepath.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return ErrNotFound
		}
		defer f.Close() //TODO(aarondl): Validate this close
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}

	return ServeContent(request, r.Res, f, fi.Name(), fi.ModTime())
}

// Attachment sends a response from `io.ReaderSeeker` as attachment, prompting
// client to save the file.
func (r ResponseWriter) Attachment(seeker io.ReadSeeker, name string) (err error) {
	r.Res.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	r.Res.Header().Set(HeaderContentDisposition, "attachment; filename="+name)
	r.Res.WriteHeader(http.StatusOK)
	// TODO(aarondl): Why is this not using ServeContent?
	_, err = io.Copy(r.Res, seeker)
	return
}

// NoContent renders out an http status code.
func (r ResponseWriter) NoContent(code int) error {
	r.Res.WriteHeader(code)
	return nil
}

// Redirect the client to a different URL. The code must be a valid redirect
// code in the range of 300 to 307 or an error is returned.
func (r ResponseWriter) Redirect(code int, url string) error {
	if code < http.StatusMultipleChoices || code > http.StatusTemporaryRedirect {
		return ErrInvalidRedirectCode
	}
	r.Res.Header().Set(HeaderLocation, url)
	r.Res.WriteHeader(code)
	return nil
}

// ServeContent intelligently serves content (like a file) to a client.
// It handles caching as well as setting headers.
//
// TODO: It does not currently handle resuming a file like this because
// it doesn't listen to the headers. It should probably use http.ServeContent
func ServeContent(rq engine.Request, rs engine.Response, content io.ReadSeeker, name string, modtime time.Time) error {
	if t, err := time.Parse(http.TimeFormat, rq.Header().Get(HeaderIfModifiedSince)); err == nil && modtime.Before(t.Add(1*time.Second)) {
		rs.Header().Del(HeaderContentType)
		rs.Header().Del(HeaderContentLength)
		rs.WriteHeader(http.StatusNotModified)
	}

	rs.Header().Set(HeaderContentType, ContentTypeByExtension(name))
	rs.Header().Set(HeaderLastModified, modtime.UTC().Format(http.TimeFormat))
	rs.WriteHeader(http.StatusOK)
	_, err := io.Copy(rs, content)
	return err
}
