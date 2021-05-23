package echo

import "encoding/json"

// DefaultJSONEncoder implements JSON encoding using encoding/json.
type DefaultJSONEncoder struct{}

// JSON converts an interface into a json and writes it to the response.
func (d DefaultJSONEncoder) JSON(i interface{}, indent string, c Context) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}
