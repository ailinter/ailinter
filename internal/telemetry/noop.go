//go:build no_telemetry

package telemetry

import "context"

var (
	InstallID = ""
	IsEnabled = false
	Version   = ""
	FirstRun  = false
)

func Init(ctx context.Context)     {}
func Shutdown(ctx context.Context) {}

func RecordInstallation()                                                  {}
func RecordCLIInvocation(command string)                                   {}
func RecordCLIInvocationWithFlags(command string, flags map[string]string) {}
func RecordMCPToolCall(tool string)                                        {}
func RecordFileAnalyzed(language, extension string)                        {}
func RecordQualityScore(language string, score int)                        {}
func RecordSmellsDetected(smellType, language string, count int)           {}
func RecordSecretsDetected(category, severity, fileExt string, count int)  {}
func RecordDuration(operation, language string, duration float64)          {}
func RecordError(errorType string)                                         {}
func RecordDirScan(fileCount int, langCounts map[string]int)               {}
func SetMCPClient(name string)                                             {}
