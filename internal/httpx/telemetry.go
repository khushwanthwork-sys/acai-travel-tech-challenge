package httpx

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry holds OpenTelemetry instrumentation
type Telemetry struct {
	tracer          trace.Tracer
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter
	activeRequests  metric.Int64UpDownCounter
}

// NewTelemetry creates a new Telemetry instance with metrics and tracing
func NewTelemetry() (*Telemetry, error) {
	meter := otel.Meter("acai-travel-chat-service")

	requestCounter, err := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create duration histogram: %w", err)
	}

	errorCounter, err := meter.Int64Counter(
		"http.server.errors",
		metric.WithDescription("Total number of HTTP errors (4xx and 5xx)"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create error counter: %w", err)
	}

	activeRequests, err := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active requests counter: %w", err)
	}

	return &Telemetry{
		tracer:          otel.Tracer("acai-travel-chat-service"),
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		errorCounter:    errorCounter,
		activeRequests:  activeRequests,
	}, nil
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Middleware returns HTTP middleware that instruments requests with metrics and tracing
func (t *Telemetry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		// Start a span for distributed tracing
		ctx, span := t.tracer.Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
			),
		)
		defer span.End()

		// Increment active requests
		t.activeRequests.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			),
		)

		// Wrap response writer to capture status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler with traced context
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Calculate request duration
		duration := time.Since(start).Seconds()

		// Common attributes for metrics
		attrs := []attribute.KeyValue{
			attribute.String("http.method", r.Method),
			attribute.String("http.route", r.URL.Path),
			attribute.Int("http.status_code", rw.statusCode),
		}

		// Record request counter
		t.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

		// Record request duration
		t.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

		// Record errors (4xx and 5xx status codes)
		if rw.statusCode >= 400 {
			errorAttrs := append(attrs,
				attribute.String("http.status_class", fmt.Sprintf("%dxx", rw.statusCode/100)),
			)
			t.errorCounter.Add(ctx, 1, metric.WithAttributes(errorAttrs...))

			// Mark span as error
			span.SetAttributes(attribute.Bool("error", true))
		}

		// Decrement active requests
		t.activeRequests.Add(ctx, -1,
			metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			),
		)

		// Add final span attributes
		span.SetAttributes(
			attribute.Int("http.status_code", rw.statusCode),
			attribute.Float64("http.duration_seconds", duration),
		)
	})
}
