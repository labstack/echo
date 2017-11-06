package echo

import "testing"

func TestDefaultTracer(t *testing.T) {
	tr := NewDefaultTracer()

	if tr.Full == true {
		t.Errorf("default traceback setting Full has wrong value, got %t", tr.Full)
	}

	if tr.Size == 0 {
		t.Errorf("default traceback setting Size has wrong value, got %d", tr.Size)
	}
}

func TestDefaultTracerTrace(t *testing.T) {
	c := &context{}
	tr := &DefaultTracer{Size: 1024 * 2}

	tr.Trace(c)

	tracebacks := c.Get(TraceContextKey)

	switch v := tracebacks.(type) {
	case []string:
		if x := len(v); x == 0 {
			t.Errorf("expecting traceback length greator than 0, got %d", x)
		}
	default:
		t.Errorf("expect traceback to be StringArray, got %V instead", v)
	}
}
