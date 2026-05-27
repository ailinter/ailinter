//go:build !no_telemetry

package telemetry

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	cliInvocations  metric.Int64Counter
	mcpToolCalls    metric.Int64Counter
	filesAnalyzed   metric.Int64Counter
	qualityScore    metric.Int64Histogram
	smellsDetected  metric.Int64Counter
	secretsDetected metric.Int64Counter
	durationSecs    metric.Float64Histogram
	errorCounter    metric.Int64Counter
	installations   metric.Int64Counter
)

func initCLIMetrics(m metric.Meter) {
	cliInvocations, _ = m.Int64Counter(
		"cli.invocations",
		metric.WithDescription("Number of CLI invocations"),
	)
	mcpToolCalls, _ = m.Int64Counter(
		"mcp.tool_calls",
		metric.WithDescription("Number of MCP tool calls"),
	)
	filesAnalyzed, _ = m.Int64Counter(
		"files.analyzed",
		metric.WithDescription("Number of files analyzed"),
	)
	qualityScore, _ = m.Int64Histogram(
		"quality.score",
		metric.WithDescription("Code quality score distribution"),
		metric.WithExplicitBucketBoundaries(0, 10, 20, 30, 40, 50, 60, 70, 74, 75, 80, 85, 90, 94, 95, 100),
	)
	smellsDetected, _ = m.Int64Counter(
		"smells.detected",
		metric.WithDescription("Number of code smells detected"),
	)
	secretsDetected, _ = m.Int64Counter(
		"secrets.detected",
		metric.WithDescription("Number of secrets detected"),
	)
	durationSecs, _ = m.Float64Histogram(
		"duration.seconds",
		metric.WithDescription("Operation duration in seconds"),
	)
	errorCounter, _ = m.Int64Counter(
		"errors",
		metric.WithDescription("Number of errors encountered"),
	)
	installations, _ = m.Int64Counter(
		"installations",
		metric.WithDescription("New installation count"),
	)
}

func baseAttrs(extra ...attribute.KeyValue) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("version", Version),
	}
	attrs = append(attrs, extra...)
	return attrs
}
