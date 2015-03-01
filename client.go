package bolt

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
)

type (
	Client struct {
		host       string
		port       string
		transport  transport
		socket     io.ReadWriteCloser
		reader     *bufio.Reader
		writer     *bufio.Writer
		httpClient HttpClient
		pool       sync.Pool
		connected  bool
	}
	HttpClient interface {
		Do(*http.Request) (*http.Response, error)
	}
)

func NewClient(opts ...func(*Client)) (c *Client) {
	c = &Client{
		host:       "localhost",
		port:       "80",
		transport:  TrnspHTTP,
		httpClient: &http.Client{},
	}
	c.pool.New = func() interface{} {
		return &Context{
			Request: &http.Request{
				URL: new(url.URL),
			},
			// Response: new(http.Response),
			// Socket: &Socket{
			// Header: SocketHeader{},
			// },
			client: true,
		}
	}
	// Set options
	for _, o := range opts {
		o(c)
	}
	return
}

func Transport(t transport) func(*Client) {
	return func(c *Client) {
		c.transport = t
	}
}

func Host(h string) func(*Client) {
	return func(c *Client) {
		c.host = h
	}
}

func Port(p string) func(*Client) {
	return func(c *Client) {
		c.port = p
	}
}

func (c *Client) Open(cfg *Config) (err error) {
	switch c.transport {
	case TrnspWS:
	case TrnspTCP:
		c.socket, err = net.Dial("tcp", net.JoinHostPort(c.host, c.port))
		if err != nil {
			return fmt.Errorf("bolt: %v", err)
		}
	default:
		return errors.New("bolt: transport not supported")
	}

	// Request
	c.writer = bufio.NewWriter(c.socket)
	c.reader = bufio.NewReader(c.socket)
	c.writer.WriteByte(byte(CmdINIT)) // Command
	var cid uint32 = 98
	if err = binary.Write(c.writer, binary.BigEndian, cid); err != nil { // Correlation ID
		log.Println(err)
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("bolt: %v", err)
	}
	if err = binary.Write(c.writer, binary.BigEndian, uint16(len(b))); err != nil { // Config length
	}
	c.writer.Write(b) // Config
	c.writer.Flush()

	for {
		var cid uint32
		binary.Read(c.reader, binary.BigEndian, &cid) // Correlation ID
		if err != nil {
			return fmt.Errorf("bolt: %v", err)
		}
		println(cid)
		break
	}

	// Response
	var n uint16
	err = binary.Read(c.reader, binary.BigEndian, &n) // Status code
	if err != nil {
		return fmt.Errorf("bolt: %v", err)
	}
	if n != 200 {
		return fmt.Errorf("bolt: status=%d", n)
	}

	return
}

func (c *Client) Auth() {
}

func (c *Client) Connect(path string) {
}

func (c *Client) Delete() {
}

func (c *Client) Get(path string, hdr Header, hl HandlerFunc) error {
	return c.Request("GET", path, hdr, hl)
}

func (c *Client) Request(method, path string, hd Header, hl HandlerFunc) (err error) {
	ctx := c.pool.Get().(*Context)
	ctx.Transport = c.transport
	switch c.transport {
	case TrnspHTTP:
		ctx.Request.Method = method
		// ctx.Request.Header = hd
		ctx.Request.URL.Scheme = "http"
		ctx.Request.URL.Host = net.JoinHostPort(c.host, c.port)
		ctx.Request.URL.Path = path
		ctx.Response, err = c.httpClient.Do(ctx.Request)
		if err != nil {
			return
		}
		// ctx.Header = ctx.Response.Header
		hl(ctx)
	case TrnspWS, TrnspTCP:
		c.writer.WriteByte(byte(CmdHTTP))   // Command
		c.writer.WriteString(method + "\n") // Method
		c.writer.WriteString(path + "\n")   // Path
		if method == "POST" || method == "PUT" || method == "PATCH" {
			// binary.Write(c.writer, binary.BigEndian, uint16(len(b))) // Header length
			// wt.Write(b)                                              // Header
		}
		c.writer.Flush()
		hl(ctx)
	}
	return
}
