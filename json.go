// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	"bytes"
	"encoding/json"
	"sync"
)

// DefaultJSONSerializer implements JSON encoding using encoding/json.
type DefaultJSONSerializer struct{}

// jsonBufPool reuses buffers for reading request bodies during JSON
// deserialization, avoiding the per-request decoder and its internal read
// buffer that json.NewDecoder allocates.
var jsonBufPool = sync.Pool{New: newJSONBuf}

func newJSONBuf() any { return new(bytes.Buffer) }

// maxPooledJSONBuf caps the capacity of buffers returned to jsonBufPool so a
// single large request body cannot pin an oversized buffer in the pool.
const maxPooledJSONBuf = 1 << 16 // 64 KiB

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (d DefaultJSONSerializer) Serialize(c *Context, target any, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(target)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
//
// The body is read into a pooled buffer and decoded with json.Unmarshal rather
// than streaming through json.NewDecoder. This avoids allocating a decoder and
// its internal read buffer on every request. json.Unmarshal does not retain a
// reference to the input slice, so the buffer is safe to reuse afterwards.
//
// Note: the full request body is read into memory before decoding. As with any
// JSON parser, decoding untrusted input can allocate large amounts of memory;
// guard such endpoints with middleware.BodyLimit (or http.MaxBytesReader),
// which makes the body read here fail fast once the limit is exceeded.
func (d DefaultJSONSerializer) Deserialize(c *Context, target any) error {
	buf := jsonBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		// Do not return oversized buffers to the pool — they would pin memory.
		if buf.Cap() <= maxPooledJSONBuf {
			jsonBufPool.Put(buf)
		}
	}()
	if _, err := buf.ReadFrom(c.Request().Body); err != nil {
		return ErrBadRequest.Wrap(err)
	}
	if err := json.Unmarshal(buf.Bytes(), target); err != nil {
		return ErrBadRequest.Wrap(err)
	}
	return nil
}
