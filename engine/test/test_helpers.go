package test

import (
	"github.com/labstack/echo/engine"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func HeaderTest(t *testing.T, header engine.Header) {
	h := "X-My-Header"
	v := "value"
	nv := "new value"
	h1 := "X-Another-Header"

	header.Add(h, v)
	assert.Equal(t, v, header.Get(h))

	header.Set(h, nv)
	assert.Equal(t, nv, header.Get(h))

	assert.True(t, header.Contains(h))

	header.Del(h)
	assert.False(t, header.Contains(h))

	header.Add(h, v)
	header.Add(h1, v)

	for _, expected := range []string{h, h1} {
		found := false
		for _, actual := range header.Keys() {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Header %s not found", expected)
		}
	}
}

func URLTest(t *testing.T, url engine.URL) {
	path := "/echo/test"
	url.SetPath(path)
	assert.Equal(t, path, url.Path())
	assert.Equal(t, map[string][]string{"param1": []string{"value1", "value2"}, "param2": []string{"value3"}}, url.QueryParams())
	assert.Equal(t, "value1", url.QueryParam("param1"))
	assert.Equal(t, "param1=value1&param1=value2&param2=value3", url.QueryString())
}

func CookieTest(t *testing.T, coockie engine.Cookie) {
	assert.Equal(t, "github.com", coockie.Domain())
	assert.Equal(t, time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC), coockie.Expires())
	assert.True(t, coockie.HTTPOnly())
	assert.True(t, coockie.Secure())
	assert.Equal(t, "session", coockie.Name())
	assert.Equal(t, "/", coockie.Path())
	assert.Equal(t, "securetoken", coockie.Value())
}
