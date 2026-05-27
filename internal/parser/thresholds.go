package parser

import "strings"

// Thresholds holds the configurable thresholds for all detectors.
type Thresholds struct {
	NestingWarning           int
	NestingAlert             int
	FileLOCWarning           int
	FileLOCAlert             int
	FileLOCCritical          int
	FuncLOCWarning           int
	FuncLOCAlert             int
	FuncCCWarning            int
	FuncCCAlert              int
	FuncCountWarning         int
	FuncCountAlert           int
	BumpyRoadBumpsWarning    int
	BumpyRoadNestingDepth    int
	ComplexCondBranchesWarn  int
	ComplexCondBranchesAlert int
	MaxArgumentsWarn         int
	MaxArgumentsAlert        int
	BrainClassMinFunc        int
	ParagraphMaxConsecutive  int
	LazyMinLines             int
	DupMinLines              int
	DupMinSimilarity         float64
	CohesionIsolationPct     float64
	CommentRatioWarning      float64
	GlobalDataWarning        int
	LongScopeVarLines        int
	LongSwitchWarn           int
	LongSwitchAlert          int
}

// DefaultThresholds returns per-language default thresholds.
func DefaultThresholds(lang string) Thresholds {
	switch strings.ToLower(lang) {
	case "python":
		return Thresholds{
			NestingWarning: 4, NestingAlert: 5,
			FileLOCWarning: 600, FileLOCAlert: 2000, FileLOCCritical: 4000,
			FuncLOCWarning: 70, FuncLOCAlert: 300,
			FuncCCWarning: 9, FuncCCAlert: 20,
			FuncCountWarning: 25, FuncCountAlert: 50,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 20, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 10, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	case "javascript", "typescript":
		return Thresholds{
			NestingWarning: 3, NestingAlert: 5,
			FileLOCWarning: 700, FileLOCAlert: 2000, FileLOCCritical: 4000,
			FuncLOCWarning: 60, FuncLOCAlert: 200,
			FuncCCWarning: 9, FuncCCAlert: 20,
			FuncCountWarning: 25, FuncCountAlert: 50,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 20, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 10, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	case "ruby":
		return Thresholds{
			NestingWarning: 3, NestingAlert: 5,
			FileLOCWarning: 600, FileLOCAlert: 1500, FileLOCCritical: 3000,
			FuncLOCWarning: 60, FuncLOCAlert: 200,
			FuncCCWarning: 7, FuncCCAlert: 15,
			FuncCountWarning: 20, FuncCountAlert: 40,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 15, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 8, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	case "swift":
		return Thresholds{
			NestingWarning: 4, NestingAlert: 5,
			FileLOCWarning: 800, FileLOCAlert: 2000, FileLOCCritical: 4000,
			FuncLOCWarning: 80, FuncLOCAlert: 300,
			FuncCCWarning: 9, FuncCCAlert: 20,
			FuncCountWarning: 25, FuncCountAlert: 50,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 20, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 10, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	case "kotlin":
		return Thresholds{
			NestingWarning: 4, NestingAlert: 5,
			FileLOCWarning: 800, FileLOCAlert: 2000, FileLOCCritical: 4000,
			FuncLOCWarning: 70, FuncLOCAlert: 300,
			FuncCCWarning: 9, FuncCCAlert: 20,
			FuncCountWarning: 25, FuncCountAlert: 50,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 20, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 10, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	default: // Go, C++, Java, Rust, C#, Swift, Kotlin
		return Thresholds{
			NestingWarning: 4, NestingAlert: 5,
			FileLOCWarning: 1000, FileLOCAlert: 2000, FileLOCCritical: 4000,
			FuncLOCWarning: 80, FuncLOCAlert: 300,
			FuncCCWarning: 9, FuncCCAlert: 20,
			FuncCountWarning: 25, FuncCountAlert: 50,
			BumpyRoadBumpsWarning: 2, BumpyRoadNestingDepth: 2,
			ComplexCondBranchesWarn: 2, ComplexCondBranchesAlert: 10,
			MaxArgumentsWarn: 4, MaxArgumentsAlert: 7,
			BrainClassMinFunc: 20, ParagraphMaxConsecutive: 20, LazyMinLines: 3,
			DupMinLines: 10, DupMinSimilarity: 0.75, CohesionIsolationPct: 0.5,
			CommentRatioWarning: 0.3, GlobalDataWarning: 5, LongScopeVarLines: 50,
			LongSwitchWarn: 10, LongSwitchAlert: 20,
		}
	}
}
