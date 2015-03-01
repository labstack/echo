package bolt

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

var u = user{
	Id:   "1",
	Name: "Joe",
}

func startTCPServer() (b *Bolt, addr string) {
	var wg sync.WaitGroup
	b = New()
	a, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.RunTCP(a.String())
	}()
	wg.Wait()
	addr = b.tcpListener.Addr().String()
	return
}

func connectTCPServer(addr string) (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", addr, time.Second)
	return
}

func TestSocketInit(t *testing.T) {
	b, addr := startTCPServer()
	conn, err := connectTCPServer(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer b.tcpListener.Close()

	// Request
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(CmdINIT)) // Command
	cfg := &Config{
		Format: FmtJSON,
	}
	bt, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	binary.Write(buf, binary.BigEndian, uint16(len(bt))) // Config length
	buf.Write(bt)                                        // Config
	buf.WriteTo(conn)

	// Response
	var n uint16
	err = binary.Read(conn, binary.BigEndian, &n) // Status code
	if err != nil {
		t.Fatal(err)
	}
	if n != 200 {
		t.Errorf("status code should be 200, found %d", n)
	}
}

func TestSocketHTTP(t *testing.T) {
	b, addr := startTCPServer()
	conn, err := connectTCPServer(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	defer b.tcpListener.Close()

	// GET
	b.Get("/users", func(c *Context) {
		c.Render(200, FmtJSON, u)
	})
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(CmdHTTP)) // Command
	buf.WriteString("GET\n")     // Method
	buf.WriteString("/users\n")  // Path
	buf.WriteTo(conn)
	var n uint16
	err = binary.Read(conn, binary.BigEndian, &n) // Status code
	if err != nil {
		t.Fatal(err)
	}
	if n != 200 {
		t.Errorf("status code should be 200, found %d", n)
	}
	verifyUser(conn, t)

	// POST
	b.Post("/users", func(c *Context) {
		c.Bind(c.Socket.config.Format, &user{})
		c.Render(201, FmtJSON, u)
	})
	buf.Reset()
	buf.WriteByte(byte(CmdHTTP)) // Command
	buf.WriteString("POST\n")    // Method
	buf.WriteString("/users\n")  // Path
	bt, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	binary.Write(buf, binary.BigEndian, int64(len(bt))) // Body length
	buf.Write(bt)                                       // Body
	buf.WriteTo(conn)
	err = binary.Read(conn, binary.BigEndian, &n) // Status code
	if err != nil {
		t.Fatal(err)
	}
	if n != 201 {
		t.Errorf("status code should be 201, found %d", n)
	}
	verifyUser(conn, t)
}

func verifyUser(rd io.Reader, t *testing.T) {
	var l int64
	err := binary.Read(rd, binary.BigEndian, &l) // Body length
	if err != nil {
		t.Fatal(err)
	}
	bd := io.LimitReader(rd, l) // Body
	u2 := new(user)
	dec := json.NewDecoder(bd)
	err = dec.Decode(u2)
	if err != nil {
		t.Fatal(err)
	}
	if u2.Id != u.Id {
		t.Error("user id should be %s, found %s", u.Id, u2.Id)
	}
	if u2.Name != u.Name {
		t.Error("user name should be %s, found %s", u.Name, u2.Name)
	}
}
