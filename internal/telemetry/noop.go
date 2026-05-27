//go:build no_telemetry

package telemetry

import "context"

var (
	InstallID = ""
	IsEnabled = false
	Version   = ""
)

func Init(ctx context.Context)   {}
func Shutdown(ctx context.Context) {}

func RecordInstallation()                                         {}
func RecordCLIInvocation(command string)                          {}
func RecordMCPToolCall(tool string)                               {}
func RecordFileAnalyzed(language, extension string)               {}
func RecordQualityScore(language string, score int)               {}
func RecordSmellsDetected(smellType, language string, count int)  {}
func RecordSecretsDetected(category, severity, fileExt string, count int) {}
func RecordDuration(operation, language string, duration float64) {}
func RecordError(errorType string)                                {}
