package echo

import (
	"encoding/json"
	"encoding/xml"
)

type (
	// Encoder declares object to []byte encoding
	Encoder interface {
		Encode(interface{}) ([]byte, error)
	}

	jsonEncoder struct{}

	xmlEncoder struct{}
)

func (e *jsonEncoder) Encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (e *xmlEncoder) Encode(i interface{}) ([]byte, error) {
	return xml.Marshal(i)
}
