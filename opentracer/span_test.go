package opentracer

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/signalfx/signalfx-go-tracing/ddtrace/tracer"
	"github.com/signalfx/signalfx-go-tracing/tracing"
	"github.com/signalfx/signalfx-go-tracing/zipkinserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_span_LogFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	ot := New(tracing.WithEndpointURL(zipkin.URL()))
	opentracing.SetGlobalTracer(ot)

	span := opentracing.GlobalTracer().StartSpan("test-span")
	span.LogFields(log.Int("int", 5), log.String("str", "value"))
	span.LogFields(log.Bool("bool", true))
	span.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 1)

	annotations := spans[0].Annotations
	require.Len(annotations, 2)

	assert.Equal(`{"int":5,"str":"value"}`, *annotations[0].Value)
	assert.Equal(`{"bool":true}`, *annotations[1].Value)
}

func TestMixedSpans(t *testing.T) {
	assert := assert.New(t)

	zipkin := zipkinserver.Start()
	defer zipkin.Stop()

	ot := New(tracing.WithEndpointURL(zipkin.URL()))
	opentracing.SetGlobalTracer(ot)

	span0, _ := tracer.StartSpanFromContext(context.Background(), "span-1")
	span1 := opentracing.StartSpan("span-2", opentracing.ChildOf(span0.Context()))

	span1.Finish()
	span0.Finish()

	tracer.ForceFlush()
	spans := zipkin.WaitForSpans(t, 2)

	assert.Equal(spans[0].TraceID, spans[1].TraceID)
	assert.Equal(spans[0].ID, *spans[1].ParentID)
}