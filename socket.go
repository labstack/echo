package bolt

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
)

type (
	Socket struct {
		Transport   transport
		Header      Header
		Body        io.ReadCloser
		config      Config
		conn        io.ReadWriteCloser
		Reader      *bufio.Reader
		Writer      *bufio.Writer
		bolt        *Bolt
		initialized bool
	}
	Config struct {
		Format format `json:"format,omitempty"`
	}
	Header map[string]string
)

func (s *Socket) Init() {
	// Request
	var cid uint32 // Correlation ID
	err := binary.Read(s.Reader, binary.BigEndian, &cid)
	if err != nil {
		log.Println(err)
	}

	var l uint16 // Config length
	err = binary.Read(s.Reader, binary.BigEndian, &l)
	if err != nil {
		log.Println(err)
	}

	rd := io.LimitReader(s.Reader, int64(l)) // Config
	dec := json.NewDecoder(rd)
	if err = dec.Decode(&s.config); err != nil {
		log.Println(err)
	}

	// Response
	if err = binary.Write(s.Writer, binary.BigEndian, cid); err != nil { // Correlation ID
		log.Println(err)
	}
	if err = binary.Write(s.Writer, binary.BigEndian, uint16(200)); err != nil { // Status code
		log.Println(err)
	}
	s.Writer.Flush()
	s.initialized = true
}

func (s *Socket) Auth() {
}

func (s *Socket) HTTP() {
	// Method
	m, err := s.Reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	m = m[:len(m)-1]

	// Path
	p, err := s.Reader.ReadString('\n')
	if err != nil {
		log.Println(err)
	}
	p = p[:len(p)-1]

	if m == "POST" || m == "PUT" || m == "PATCH" {
		var l int64
		err = binary.Read(s.Reader, binary.BigEndian, &l) // Body length
		if err != nil {
			log.Println(err)
		}
		s.Body = ioutil.NopCloser(io.LimitReader(s.Reader, l)) // Body
	}

	h, c, st := s.bolt.Router.Find(m, p)
	c.Socket = s
	c.Transport = s.Transport
	if h != nil {
		h(c)
	} else {
		if st == NotFound {
			s.bolt.notFoundHandler(c)
		} else if st == NotAllowed {
			s.bolt.methodNotAllowedHandler(c)
		}
	}
	s.bolt.pool.Put(c)
}
