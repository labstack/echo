package db

import "time"

type (
	// Request defines the data to be logged by logger middleware.
	Request struct {
		// ID string `json:"id,omitempty"` (Request ID - Not implemented)
		Time         *time.Time        `json:"time,omitempty"` // http://stackoverflow.com/questions/32643815/golang-json-omitempty-with-time-time-field
		RemoteIP     string            `json:"remote_ip,omitempty"`
		URI          string            `json:"uri,omitempty"`
		Host         string            `json:"host,omitempty"`
		Method       string            `json:"method,omitempty"`
		Path         string            `json:"path,omitempty"`
		Referer      string            `json:"referer,omitempty"`
		UserAgent    string            `json:"user_agent,omitempty"`
		Status       int               `json:"status,omitempty"`
		Latency      time.Duration     `json:"latency,omitempty"`
		LatencyHuman string            `json:"latency_human,omitempty"`
		BytesIn      int64             `json:"bytes_in"`  // Allow 0 value
		BytesOut     int64             `json:"bytes_out"` // As aboves
		Header       map[string]string `json:"header,omitempty"`
		Form         map[string]string `json:"form,omitempty"`
		Query        map[string]string `json:"query,omitempty"`
	}
)
