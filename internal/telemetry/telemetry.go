//go:build !no_telemetry

package telemetry

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	InstallID string
	IsEnabled bool
	Version   string

	shutdownFn func(context.Context) error
	initOnce   sync.Once
)

var defaultEndpoint = getEnvOrDefault("AILINTER_TELEMETRY_ENDPOINT", "https://telemetry.ailinter.dev")

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}



func Init(ctx context.Context) {
	initOnce.Do(func() {
		IsEnabled = resolveEnabled()
		if !IsEnabled {
			otel.SetMeterProvider(metricnoop.NewMeterProvider())
			return
		}

		InstallID = loadOrCreateInstallID()

		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceName("ailinter"),
				semconv.ServiceVersion(Version),
				semconv.OSName(runtime.GOOS),
				semconv.HostArchKey.String(runtime.GOARCH),
				attribute.String("install.id", InstallID),
			),
		)
		if err != nil {
			IsEnabled = false
			otel.SetMeterProvider(metricnoop.NewMeterProvider())
			return
		}

		exporter, err := newExporter(ctx)
		if err != nil {
			IsEnabled = false
			otel.SetMeterProvider(metricnoop.NewMeterProvider())
			return
		}

		reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(15*time.Second))
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
		otel.SetMeterProvider(provider)
		shutdownFn = provider.Shutdown

		initMetrics(provider.Meter("ailinter"))
	})
}

func Shutdown(ctx context.Context) {
	if shutdownFn != nil {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		shutdownFn(ctx2)
	}
}

func resolveEnabled() bool {
	if v := strings.ToLower(os.Getenv("AILINTER_NO_TELEMETRY")); v == "1" || v == "true" {
		return false
	}
	return true
}

func loadOrCreateInstallID() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(configDir, "ailinter")
	idFile := filepath.Join(dir, "install_id")
	if data, err := os.ReadFile(idFile); err == nil {
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:16])
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return ""
	}
	os.MkdirAll(dir, 0700)
	os.WriteFile(idFile, id, 0600)
	hash := sha256.Sum256(id)
	return hex.EncodeToString(hash[:16])
}

func initMetrics(m metric.Meter) {
	initCLIMetrics(m)
}
