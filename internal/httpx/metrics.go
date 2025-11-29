package httpx

import (
	"context"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

// SetupPrometheusExporter creates a Prometheus exporter and returns the meter provider and exporter
func SetupPrometheusExporter() (*metric.MeterProvider, *prometheus.Exporter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, nil, err
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	return provider, exporter, nil
}

// Shutdown gracefully shuts down the meter provider
func Shutdown(ctx context.Context, provider *metric.MeterProvider) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown(ctx)
}
