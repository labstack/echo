package bytes

import (
	"fmt"
	"regexp"
	"strconv"
)

type (
	Bytes struct {
	}
)

const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
)

var (
	pattern = regexp.MustCompile(`(?i)^(-?\d+)([KMGTP]B?|B)$`)
	global  = New()
)

// New creates a Bytes instance.
func New() *Bytes {
	return &Bytes{}
}

// Format formats bytes integer to human readable string.
// For example, 31323 bytes will return 30.59KB.
func (*Bytes) Format(b int64) string {
	multiple := ""
	value := float64(b)

	switch {
	case b < KB:
		return strconv.FormatInt(b, 10) + "B"
	case b < MB:
		value /= KB
		multiple = "KB"
	case b < MB:
		value /= KB
		multiple = "KB"
	case b < GB:
		value /= MB
		multiple = "MB"
	case b < TB:
		value /= GB
		multiple = "GB"
	case b < PB:
		value /= TB
		multiple = "TB"
	case b < EB:
		value /= PB
		multiple = "PB"
	}

	return fmt.Sprintf("%.02f%s", value, multiple)
}

// Parse parses human readable bytes string to bytes integer.
// For example, 6GB (6G is also valid) will return 6442450944.
func (*Bytes) Parse(value string) (i int64, err error) {
	parts := pattern.FindStringSubmatch(value)
	if len(parts) < 3 {
		return 0, fmt.Errorf("error parsing value=%s", value)
	}
	bytesString := parts[1]
	multiple := parts[2]
	bytes, err := strconv.ParseInt(bytesString, 10, 64)
	if err != nil {
		return
	}

	switch multiple {
	case "B":
		return bytes * B, nil
	case "K", "KB":
		return bytes * KB, nil
	case "M", "MB":
		return bytes * MB, nil
	case "G", "GB":
		return bytes * GB, nil
	case "T", "TB":
		return bytes * TB, nil
	case "P", "PB":
		return bytes * PB, nil
	}

	return
}

// Format wraps global Bytes's Format function.
func Format(b int64) string {
	return global.Format(b)
}

// Parse wraps global Bytes's Parse function.
func Parse(val string) (int64, error) {
	return global.Parse(val)
}
