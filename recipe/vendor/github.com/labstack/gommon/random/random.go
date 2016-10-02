package random

import (
	"math/rand"
	"time"
)

type (
	Random struct {
		charset Charset
	}

	Charset string
)

const (
	Alphanumeric Charset = Alphabetic + Numeric
	Alphabetic   Charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numeric      Charset = "0123456789"
	Hex          Charset = Numeric + "abcdef"
)

var (
	global = New()
)

func New() *Random {
	rand.Seed(time.Now().UnixNano())
	return &Random{
		charset: Alphanumeric,
	}
}

func (r *Random) SetCharset(c Charset) {
	r.charset = c
}

func (r *Random) String(length uint8) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = r.charset[rand.Int63()%int64(len(r.charset))]
	}
	return string(b)
}

func SetCharset(c Charset) {
	global.SetCharset(c)
}

func String(length uint8) string {
	return global.String(length)
}
