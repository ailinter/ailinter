//go:build !no_telemetry

package telemetry

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var mcpClientName string

func clientType() string {
	if ct := os.Getenv("AILINTER_CLIENT_TYPE"); ct != "" {
		return ct
	}
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		return "mcp"
	}
	return "cli"
}

func SetMCPClient(name string) {
	if name != "" {
		mcpClientName = name
	}
}

func RecordInstallation() {
	if !IsEnabled || installations == nil {
		return
	}
	attrs := baseAttrs()
	if FirstRun {
		attrs = append(attrs, attribute.Bool("first_run", true))
	}
	installations.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

func RecordCLIInvocation(command string) {
	RecordCLIInvocationWithFlags(command, nil)
}

func RecordCLIInvocationWithFlags(command string, activeFlags map[string]string) {
	if !IsEnabled || cliInvocations == nil {
		return
	}
	attrs := append(baseAttrs(),
		attribute.String("command", command),
		attribute.String("os", runtime.GOOS),
		attribute.String("arch", runtime.GOARCH),
	)
	for k, v := range activeFlags {
		attrs = append(attrs, attribute.String("cli.flag."+k, v))
	}
	cliInvocations.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

func RecordMCPToolCall(tool string) {
	if !IsEnabled || mcpToolCalls == nil {
		return
	}
	attrs := append(baseAttrs(),
		attribute.String("tool", tool),
	)
	if mcpClientName != "" {
		attrs = append(attrs, attribute.String("mcp.client", mcpClientName))
	}
	mcpToolCalls.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

func RecordFileAnalyzed(language, extension string) {
	if !IsEnabled || filesAnalyzed == nil {
		return
	}
	filesAnalyzed.Add(context.Background(), 1, metric.WithAttributes(append(baseAttrs(),
		attribute.String("language", language),
		attribute.String("extension", extension),
	)...))
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

func RecordDirScan(fileCount int, langCounts map[string]int) {
	if !IsEnabled || cliInvocations == nil {
		return
	}
	attrs := append(baseAttrs(),
		attribute.String("command", "check"),
		attribute.Int("scan.file_count", fileCount),
	)
	for lang, count := range langCounts {
		attrs = append(attrs, attribute.Int("scan.lang."+lang, count))
	}
	cliInvocations.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

// Ensure fmt used.
var _ = fmt.Sprintf
