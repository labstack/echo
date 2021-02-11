// +build !jsoniter

package json

import "encoding/json"

var (
	// Marshal is exported by echo/json package.
	Marshal = json.Marshal
	// Unmarshal is exported by echo/json package.
	Unmarshal = json.Unmarshal
	// NewDecoder is exported by ehco/json package.
	NewDecoder = json.NewDecoder
	// NewEncoder is exported by echo/json package.
	NewEncoder = json.NewEncoder
)

type(
	// UnmarshalTypeError is exported by echo/json package.
	UnmarshalTypeError = json.UnmarshalTypeError
	// SyntaxError is exported by echo/json package.
	SyntaxError = json.SyntaxError
	// RawMessage is exported by echo/json package.
	RawMessage = json.RawMessage
)
