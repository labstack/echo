package standard

import (
	"bufio"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/trafficstars/echo/engine/test"
)

func TestRequest(t *testing.T) {
	httpReq, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(test.MultipartRequest)))
	url, _ := url.Parse("http://github.com/trafficstars/echo")
	httpReq.URL = url
	httpReq.RemoteAddr = "127.0.0.1"
	req := NewRequest(httpReq, log.New("echo"))
	test.RequestTest(t, req)
	nr, _ := http.NewRequest("GET", "/", nil)
	req.reset(nr, nil, nil)
	assert.Equal(t, "", req.Host())
}
