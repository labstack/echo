package test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/labstack/echo/engine"
	"github.com/stretchr/testify/assert"
)

const MultipartRequest = `POST /labstack/echo HTTP/1.1
Host: github.com
Connection: close
User-Agent: Mozilla/5.0 (Macintosh; U; Intel Mac OS X; de-de) AppleWebKit/523.10.3 (KHTML, like Gecko) Version/3.0.4 Safari/523.10
Content-Type: multipart/form-data; boundary=Asrf456BGe4h
Content-Length: 261
Accept-Encoding: gzip
Accept-Charset: ISO-8859-1,UTF-8;q=0.7,*;q=0.7
Cache-Control: no-cache
Accept-Language: de,en;q=0.7,en-us;q=0.3
Referer: https://github.com/
Cookie: session=securetoken; user=123
X-Real-IP: 192.168.1.1

--Asrf456BGe4h
Content-Disposition: form-data; name="foo"

bar
--Asrf456BGe4h
Content-Disposition: form-data; name="baz"

bat
--Asrf456BGe4h
Content-Disposition: form-data; name="note"; filename="note.txt"
Content-Type: text/plain

Hello world!
--Asrf456BGe4h--
`

func RequestTest(t *testing.T, request engine.Request) {
	assert.Equal(t, "github.com", request.Host())
	request.SetHost("labstack.com")
	assert.Equal(t, "labstack.com", request.Host())
	request.SetURI("/labstack/echo?token=54321")
	assert.Equal(t, "/labstack/echo?token=54321", request.URI())
	assert.Equal(t, "/labstack/echo", request.URL().Path())
	assert.Equal(t, "https://github.com/", request.Referer())
	assert.Equal(t, "192.168.1.1", request.Header().Get("X-Real-IP"))
	assert.Equal(t, "http", request.Scheme())
	assert.Equal(t, "Mozilla/5.0 (Macintosh; U; Intel Mac OS X; de-de) AppleWebKit/523.10.3 (KHTML, like Gecko) Version/3.0.4 Safari/523.10", request.UserAgent())
	assert.Equal(t, "127.0.0.1", request.RemoteAddress())
	assert.Equal(t, "192.168.1.1", request.RealIP())
	assert.Equal(t, "POST", request.Method())
	assert.Equal(t, int64(261), request.ContentLength())
	assert.Equal(t, "bar", request.FormValue("foo"))

	if fHeader, err := request.FormFile("note"); assert.NoError(t, err) {
		if file, err := fHeader.Open(); assert.NoError(t, err) {
			text, _ := ioutil.ReadAll(file)
			assert.Equal(t, "Hello world!", string(text))
		}
	}

	assert.Equal(t, map[string][]string{"baz": []string{"bat"}, "foo": []string{"bar"}}, request.FormParams())

	if form, err := request.MultipartForm(); assert.NoError(t, err) {
		_, ok := form.File["note"]
		assert.True(t, ok)
	}

	request.SetMethod("PUT")
	assert.Equal(t, "PUT", request.Method())

	request.SetBody(strings.NewReader("Hello"))
	if body, err := ioutil.ReadAll(request.Body()); assert.NoError(t, err) {
		assert.Equal(t, "Hello", string(body))
	}

	if cookie, err := request.Cookie("session"); assert.NoError(t, err) {
		assert.Equal(t, "session", cookie.Name())
		assert.Equal(t, "securetoken", cookie.Value())
	}

	_, err := request.Cookie("foo")
	assert.Error(t, err)

	// Cookies
	cs := request.Cookies()
	if assert.Len(t, cs, 2) {
		assert.Equal(t, "session", cs[0].Name())
		assert.Equal(t, "securetoken", cs[0].Value())
		assert.Equal(t, "user", cs[1].Name())
		assert.Equal(t, "123", cs[1].Value())
	}
}
