// +build jsoniter

package json

import jsoniter "github.com/json-iterator/go"
import encodingJson "encoding/json"

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
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
	UnmarshalTypeError = encodingJson.UnmarshalTypeError
	// SyntaxError is exported by echo/json package.
	SyntaxError = encodingJson.SyntaxError
	// RawMessage is exported by echo/json package.
	RawMessage = encodingJson.RawMessage
)
