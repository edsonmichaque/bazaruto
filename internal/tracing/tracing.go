package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer with additional functionality.
type Tracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

// NewTracer creates a new tracer instance.
func NewTracer(serviceName, endpoint string) (*Tracer, error) {
	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // Sample 10% of traces
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer
	tracer := provider.Tracer(serviceName)

	return &Tracer{
		tracer:   tracer,
		provider: provider,
	}, nil
}

// StartSpan starts a new span with the given name and options.
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartSpanWithAttributes starts a new span with attributes.
func (t *Tracer) StartSpanWithAttributes(ctx context.Context, name string, attrs map[string]interface{}, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	spanOpts := make([]trace.SpanStartOption, len(opts))
	copy(spanOpts, opts)

	// Add attributes
	for key, value := range attrs {
		spanOpts = append(spanOpts, trace.WithAttributes(
			attributeFromValue(key, value),
		))
	}

	return t.tracer.Start(ctx, name, spanOpts...)
}

// StartHTTPRequestSpan starts a span for HTTP request handling.
func (t *Tracer) StartHTTPRequestSpan(ctx context.Context, method, path string) (context.Context, trace.Span) {
	attrs := map[string]interface{}{
		"http.method": method,
		"http.url":    path,
		"span.kind":   "server",
	}

	return t.StartSpanWithAttributes(ctx, fmt.Sprintf("%s %s", method, path), attrs)
}

// StartDatabaseSpan starts a span for database operations.
func (t *Tracer) StartDatabaseSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	attrs := map[string]interface{}{
		"db.operation": operation,
		"db.table":     table,
		"span.kind":    "client",
	}

	return t.StartSpanWithAttributes(ctx, fmt.Sprintf("db.%s", operation), attrs)
}

// StartServiceSpan starts a span for service operations.
func (t *Tracer) StartServiceSpan(ctx context.Context, service, operation string) (context.Context, trace.Span) {
	attrs := map[string]interface{}{
		"service.name":   service,
		"service.method": operation,
		"span.kind":      "internal",
	}

	return t.StartSpanWithAttributes(ctx, fmt.Sprintf("%s.%s", service, operation), attrs)
}

// Shutdown gracefully shuts down the tracer provider.
func (t *Tracer) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return t.provider.Shutdown(ctx)
}

// attributeFromValue creates an OpenTelemetry attribute from a key-value pair.
func attributeFromValue(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

// GetTracer returns the underlying OpenTelemetry tracer.
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// GetProvider returns the underlying tracer provider.
func (t *Tracer) GetProvider() *sdktrace.TracerProvider {
	return t.provider
}
