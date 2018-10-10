// +build jsoniter

package json

import (
	"github.com/json-iterator/go"
)

var (
	PKG           = JSONITER
	jnt           = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal       = jnt.Marshal
	MarshalIndent = jnt.MarshalIndent
	Unmarshal     = jnt.Unmarshal
	NewDecoder    = jnt.NewDecoder
)
