package echo

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
