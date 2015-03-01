package bolt

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

type (
	httpClient struct{}
	user       struct {
		Id   string
		Name string
	}
)

func mockTcpServer(method, path string) (b *Bolt) {
	b = New()
	b.Handle(method, path, nil)
	b.RunTCP(":9999")
	return
}

func mockTcpReq(status uint16, hd Header, body []byte) (c *Client) {
	c = NewClient()
	c.transport = TrnspTCP
	// c.socket = new(bytes.Buffer)
	c.socket = os.Stdout
	// l := int64(len(body))
	// hd.Set("Content-Length", strconv.FormatInt(l, 10))
	c.socket.Write(body)
	return
}

func (c *httpClient) Do(req *http.Request) (res *http.Response, err error) {
	res = &http.Response{}
	if req.Method == "GET" {
		res.StatusCode = 200
		u := &user{"1", "Joe"}
		d, _ := json.Marshal(u)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(d))
		res.Header = make(http.Header)
		res.Header.Add("Accept", MIME_JSON)
	}
	return
}

func TestClientPost(t *testing.T) {
}

func TestClientHttpGet(t *testing.T) {
	c := NewClient()
	c.httpClient = &httpClient{}
	hd := make(Header)
	// hd.Set("Accept", MIME_JSON)
	if err := c.Get("/users/1", hd, func(ctx *Context) {
		u := new(user)
		ctx.Decode(u)
		if u.Id != "1" {
			t.Error()
		}
	}); err != nil {
		t.Error(err)
	}
}

func TestTCPClient(t *testing.T) {
	b, addr := startTCPServer()
	defer b.tcpListener.Close()

	// Open
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	c := NewClient(Transport(TrnspTCP), Host(host), Port(port))
	go func() {
		err = c.Open(&Config{
			Format: FmtJSON,
		})
		if err != nil {
			t.Fatal(err)
		}
	}()
	time.Sleep(32 * time.Millisecond)

	// Get
	// b.Get("/users", func(c *Context) {
	// c.Render(200, FmtJSON, u)
	// })
	// err = c.Get("/users", nil, func(c *Context) {
	// })
	// time.Sleep(32 * time.Millisecond)
	// if err != nil {
	// t.Fatal(err)
	// }
}
