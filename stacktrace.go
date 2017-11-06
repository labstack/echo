package echo

import (
	"runtime"
	"strings"
)

// TraceContextKey is the key sets to context
const TraceContextKey = "echo_context_traceback"

type (
	// Tracer interface allows custom traceback
	// function to be implemented
	Tracer interface {
		Trace(c Context)
	}

	// DefaultTracer offers basic traceback
	DefaultTracer struct {
		Full bool
		Size int
	}
)

// NewDefaultTracer creates a new DefaultTracer with
// traceback size sets to 2MB and Full trace to false
func NewDefaultTracer() *DefaultTracer {
	t := &DefaultTracer{
		Full: false,
		Size: 1024 * 2,
	}
	return t
}

// Trace run the stack trace
func (t *DefaultTracer) Trace(c Context) {
	rawStack := make([]byte, t.Size)
	runtime.Stack(rawStack, t.Full)

	stack := t.Format(rawStack)
	c.Set(TraceContextKey, stack)
}

// Format formats the traceback result for easier processing
func (t *DefaultTracer) Format(b []byte) []string {
	stack := strings.Split(string(b), "\n\t")

	for i, v := range stack {
		v = strings.Trim(v, "\n")
		v = strings.Replace(v, "\n", ": ", -1)
		stack[i] = v
	}

	return stack
}
