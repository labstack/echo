package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

// Mock opentracing.Span
type mockSpan struct {
	tracer   opentracing.Tracer
	tags     map[string]interface{}
	opName   string
	finished bool
}

func createSpan(tracer opentracing.Tracer) *mockSpan {
	return &mockSpan{
		tracer: tracer,
		tags:   make(map[string]interface{}),
	}
}

func (sp *mockSpan) isFinished() bool {
	return sp.finished
}

func (sp *mockSpan) getOpName() string {
	return sp.opName
}

func (sp *mockSpan) getTag(key string) interface{} {
	return sp.tags[key]
}

func (sp *mockSpan) Finish() {
	sp.finished = true
}
func (sp *mockSpan) FinishWithOptions(opts opentracing.FinishOptions) {
}
func (sp *mockSpan) Context() opentracing.SpanContext {
	return nil
}
func (sp *mockSpan) SetOperationName(operationName string) opentracing.Span {
	sp.opName = operationName
	return sp
}
func (sp *mockSpan) SetTag(key string, value interface{}) opentracing.Span {
	sp.tags[key] = value
	return sp
}
func (sp *mockSpan) LogFields(fields ...log.Field) {
}
func (sp *mockSpan) LogKV(alternatingKeyValues ...interface{}) {
}
func (sp *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	return sp
}
func (sp *mockSpan) BaggageItem(restrictedKey string) string {
	return ""
}
func (sp *mockSpan) Tracer() opentracing.Tracer {
	return sp.tracer
}
func (sp *mockSpan) LogEvent(event string) {
}
func (sp *mockSpan) LogEventWithPayload(event string, payload interface{}) {
}
func (sp *mockSpan) Log(data opentracing.LogData) {
}

// Mock opentracing.Tracer
type mockTracer struct {
	span mockSpan
}

func (tr *mockTracer) currentSpan() *mockSpan {
	return &tr.span
}

func (tr *mockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return &tr.span
}

func (tr *mockTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}

func (tr *mockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, nil
}

func createMockTracer() *mockTracer {
	tracer := mockTracer{}
	span := createSpan(&tracer)
	tracer.span = *span
	return &tracer
}

func TestTrace(t *testing.T) {
	tracer := createMockTracer()

	e := echo.New()
	e.Use(TraceWithConfig(TraceConfig{
		tracer:        tracer,
		componentName: "EchoTracer",
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, true, tracer.currentSpan().isFinished())
	assert.Equal(t, "GET", tracer.currentSpan().getTag("http.method"))
	assert.Equal(t, "/", tracer.currentSpan().getTag("http.url"))
	assert.Equal(t, "EchoTracer", tracer.currentSpan().getTag("component"))
	assert.Equal(t, uint16(200), tracer.currentSpan().getTag("http.status_code"))
	assert.Equal(t, true, tracer.currentSpan().getTag("error"))
}
