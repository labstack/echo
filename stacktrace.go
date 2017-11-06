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
	}
)

// Trace run the stack trace
func (t *DefaultTracer) Trace(c Context) {
	rawStack := []byte{}
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
