package bolt

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type (
	Bolt struct {
		Router                  *router
		handlers                []HandlerFunc
		maxParam                byte
		notFoundHandler         HandlerFunc
		methodNotAllowedHandler HandlerFunc
		tcpListener             *net.TCPListener
		// udpConn                 *net.UDPConn
		pool sync.Pool
	}

	command   byte
	format    byte
	transport byte

	HandlerFunc func(*Context)
)

const (
	CmdINIT command = 1 + iota
	CmdAUTH
	CmdHTTP
	CmdPUB
	CmdSUB
	CmdUSUB
)

const (
	TrnspHTTP transport = 1 + iota
	TrnspWS
	TrnspTCP
)

const (
	FmtJSON format = 1 + iota
	FmtMsgPack
	FmtBinary = 20
)

const (
	MIME_JSON = "application/json"
	MIME_MP   = "application/x-msgpack"

	HdrAccept             = "Accept"
	HdrContentDisposition = "Content-Disposition"
	HdrContentLength      = "Content-Length"
	HdrContentType        = "Content-Type"
)

var MethodMap = map[string]uint8{
	"CONNECT": 1,
	"DELETE":  2,
	"GET":     3,
	"HEAD":    4,
	"OPTIONS": 5,
	"PATCH":   6,
	"POST":    7,
	"PUT":     8,
	"TRACE":   9,
}

func New(opts ...func(*Bolt)) (b *Bolt) {
	b = &Bolt{
		maxParam: 5,
		notFoundHandler: func(c *Context) {
			http.Error(c.Writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		},
		methodNotAllowedHandler: func(c *Context) {
			http.Error(c.Writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		},
	}
	b.Router = NewRouter(b)
	b.pool.New = func() interface{} {
		return &Context{
			Writer: NewResponse(nil),
			Socket: new(Socket),
			params: make(Params, b.maxParam),
			store:  make(store),
			i:      -1,
		}
	}

	// Set options
	for _, o := range opts {
		o(b)
	}

	return
}

func MaxParam(n uint8) func(*Bolt) {
	return func(b *Bolt) {
		b.maxParam = n
	}
}

func NotFoundHandler(h HandlerFunc) func(*Bolt) {
	return func(b *Bolt) {
		b.notFoundHandler = h
	}
}

func MethodNotAllowedHandler(h HandlerFunc) func(*Bolt) {
	return func(b *Bolt) {
		b.methodNotAllowedHandler = h
	}
}

// Use adds middleware(s) in the chain.
func (b *Bolt) Use(h ...HandlerFunc) {
	b.handlers = append(b.handlers, h...)
}

func (b *Bolt) Connect(path string, h ...HandlerFunc) {
	b.Handle("CONNECT", path, h)
}

func (b *Bolt) Delete(path string, h ...HandlerFunc) {
	b.Handle("DELETE", path, h)
}

func (b *Bolt) Get(path string, h ...HandlerFunc) {
	b.Handle("GET", path, h)
}

func (b *Bolt) Head(path string, h ...HandlerFunc) {
	b.Handle("HEAD", path, h)
}

func (b *Bolt) Options(path string, h ...HandlerFunc) {
	b.Handle("OPTIONS", path, h)
}

func (b *Bolt) Patch(path string, h ...HandlerFunc) {
	b.Handle("PATCH", path, h)
}

func (b *Bolt) Post(path string, h ...HandlerFunc) {
	b.Handle("POST", path, h)
}

func (b *Bolt) Put(path string, h ...HandlerFunc) {
	b.Handle("PUT", path, h)
}

func (b *Bolt) Trace(path string, h ...HandlerFunc) {
	b.Handle("TRACE", path, h)
}

func (b *Bolt) Handle(method, path string, h []HandlerFunc) {
	h = append(b.handlers, h...)
	l := len(h)
	b.Router.Add(method, path, func(c *Context) {
		c.handlers = h
		c.l = l
		c.i = -1
		c.Next()
	})
}

func (b *Bolt) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Find and execute handler
	h, c, s := b.Router.Find(r.Method, r.URL.Path)
	if h != nil {
		c.Transport = TrnspHTTP
		c.Writer.ResponseWriter = rw
		c.Request = r
		h(c)
	} else {
		if s == NotFound {
			b.notFoundHandler(c)
		} else if s == NotAllowed {
			b.methodNotAllowedHandler(c)
		}
	}
	b.pool.Put(c)
}

func (b *Bolt) handleSocket(conn io.ReadWriteCloser, tp transport) {
	// TODO: From pool?
	defer conn.Close()
	s := &Socket{
		Transport: tp,
		config:    Config{},
		conn:      conn,
		Reader:    bufio.NewReader(conn),
		Writer:    bufio.NewWriter(conn),
		bolt:      b,
	}
Loop:
	for {
		c, err := s.Reader.ReadByte() // Command
		if err != nil {
			log.Println(err, 222, c)
		}
		cmd := command(c)
		println(cmd)
		switch cmd {
		case CmdINIT:
			s.Init()
		case CmdHTTP:
			s.HTTP()
		default:
			break Loop
		}
	}
}

func (b *Bolt) RunHTTP(addr string) {
	log.Fatal(http.ListenAndServe(addr, b))
}

func (b *Bolt) RunWS(addr string) {
	http.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		b.handleSocket(ws, TrnspWS)
	}))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (b *Bolt) RunTCP(addr string) {
	a, _ := net.ResolveTCPAddr("tcp", addr)
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		log.Fatalf("bolt: %v", err)
	}
	b.tcpListener = l
	go b.serve("tcp")
}

func (b *Bolt) serve(net string) {
	switch net {
	case "ws":
	case "tcp":
		for {
			conn, err := b.tcpListener.Accept()
			if err != nil {
				log.Print(err)
				return
			}
			go b.handleSocket(conn, TrnspTCP)
		}
	default:
		// TODO: handle it!
	}
}
