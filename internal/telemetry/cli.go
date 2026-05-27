//go:build !no_telemetry

package telemetry

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func RecordInstallation() {
	if !IsEnabled || installations == nil {
		return
	}
	installations.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("version", Version),
	))
}

func RecordCLIInvocation(command string) {
	if !IsEnabled || cliInvocations == nil {
		return
	}
	cliInvocations.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("command", command),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
		attribute.String("version", Version),
	))
}

func RecordMCPToolCall(tool string) {
	if !IsEnabled || mcpToolCalls == nil {
		return
	}
	mcpToolCalls.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("tool", tool),
		attribute.String("version", Version),
	))
}

func RecordFileAnalyzed(language, extension string) {
	if !IsEnabled || filesAnalyzed == nil {
		return
	}
	filesAnalyzed.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("language", language),
		attribute.String("extension", extension),
		attribute.String("version", Version),
	))
}

func RecordQualityScore(language string, score int) {
	if !IsEnabled || qualityScore == nil {
		return
	}
	qualityScore.Record(context.Background(), int64(score), metric.WithAttributes(
		attribute.String("language", language),
	))
}

func RecordSmellsDetected(smellType, language string, count int) {
	if !IsEnabled || smellsDetected == nil {
		return
	}
	smellsDetected.Add(context.Background(), int64(count), metric.WithAttributes(
		attribute.String("smell_type", smellType),
		attribute.String("language", language),
	))
}

func RecordSecretsDetected(category, severity, fileExt string, count int) {
	if !IsEnabled || secretsDetected == nil {
		return
	}
	secretsDetected.Add(context.Background(), int64(count), metric.WithAttributes(
		attribute.String("category", category),
		attribute.String("severity", severity),
		attribute.String("file_ext", fileExt),
	))
}

func RecordDuration(operation, language string, duration float64) {
	if !IsEnabled || durationSecs == nil {
		return
	}
	durationSecs.Record(context.Background(), duration, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("language", language),
	))
}

func RecordError(errorType string) {
	if !IsEnabled || errorCounter == nil {
		return
	}
	errorCounter.Add(context.Background(), 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
	))
}
