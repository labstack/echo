// +build !jsoniter

package json

import (
	j "encoding/json"
)

var (
	PKG           = STDJSON
	Marshal       = j.Marshal
	MarshalIndent = j.MarshalIndent
	Unmarshal     = j.Unmarshal
	NewDecoder    = j.NewDecoder
)
