package json

import (
	"encoding/json"
)

type UnmarshalTypeError = json.UnmarshalTypeError
type SyntaxError = json.SyntaxError
type RawMessage = json.RawMessage

const (
	STDJSON = iota
	JSONITER
)
