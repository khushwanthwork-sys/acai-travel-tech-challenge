package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat"
	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/httpx"
	"github.com/acai-travel/tech-challenge/internal/mongox"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twitchtv/twirp"

	"go.opentelemetry.io/otel"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Initialize OpenTelemetry
	meterProvider, _, err := httpx.SetupPrometheusExporter()
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := httpx.Shutdown(shutdownCtx, meterProvider); err != nil {
			slog.Error("Failed to shutdown OpenTelemetry", "error", err)
		}
	}()

	otel.SetMeterProvider(meterProvider)
	slog.Info("OpenTelemetry initialized", "metrics_endpoint", "/metrics")

	// Initialize MongoDB
	mongo := mongox.MustConnect()

	// Initialize components
	repo := model.New(mongo)
	assist := assistant.New()
	server := chat.NewServer(repo, assist)

	// Initialize telemetry
	telemetry, err := httpx.NewTelemetry()
	if err != nil {
		slog.Error("Failed to initialize telemetry", "error", err)
		os.Exit(1)
	}

	// Configure handler
	handler := mux.NewRouter()
	handler.Use(
		telemetry.Middleware,
		httpx.Logger(),
		httpx.Recovery(),
	)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hi, my name is Clippy!")
	})

	// Use standard Prometheus HTTP handler
	handler.Handle("/metrics", promhttp.Handler())

	handler.PathPrefix("/twirp/").Handler(
		pb.NewChatServiceServer(server, twirp.WithServerJSONSkipDefaults(true)),
	)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("Starting the server", "port", 8080)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}
