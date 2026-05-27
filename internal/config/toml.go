package config

import (
	"os"
	"path/filepath"

	"github.com/ailinter/ailinter/internal/parser"
	"github.com/pelletier/go-toml/v2"
)

// UserConfig mirrors .ailinter.toml structure.
type UserConfig struct {
	Extends string       `toml:"extends"`
	Rules   RulesSection `toml:"rules"`
}

// RulesSection holds per-detector overrides.
type RulesSection struct {
	DeepNesting          *RuleInt      `toml:"deep_nesting"`
	BrainMethod          *RuleLoc      `toml:"brain_method"`
	FileBloat            *RuleLoc      `toml:"file_bloat"`
	ComplexConditional   *RuleBranches `toml:"complex_conditional"`
	CyclomaticComplexity *RuleInt      `toml:"cyclomatic_complexity"`
	BumpyRoad            *RuleBump     `toml:"bumpy_road"`
	LongParameterList    *RuleInt      `toml:"long_parameter_list"`
	LazyElement          *RuleLine     `toml:"lazy_element"`
	ParagraphOfCode      *RuleLine     `toml:"paragraph_of_code"`
	MessageChains        *RuleLine     `toml:"message_chains"`
	PrimitiveObsession   *RuleLine     `toml:"primitive_obsession"`
	ExcessiveComments    *RuleLine     `toml:"excessive_comments"`
	GlobalData           *RuleLine     `toml:"global_data"`
	LongScopeVariable    *RuleLine     `toml:"long_scope_variable"`
	DuplicatedCode       *RuleFloat    `toml:"duplicated_code"`
	LongSwitch           *RuleInt      `toml:"long_switch"`
}

// RuleInt handles thresholds with warning/alert levels.
type RuleInt struct {
	Weight  float64 `toml:"weight"`
	Warning int     `toml:"warning"`
	Alert   int     `toml:"alert"`
}

// RuleLoc handles thresholds with LOC-based warning_lines/alert_lines.
type RuleLoc struct {
	Weight       float64 `toml:"weight"`
	WarningLines int     `toml:"warning_lines"`
	AlertLines   int     `toml:"alert_lines"`
}

// RuleBump handles Bumpy Road thresholds.
type RuleBump struct {
	Weight       float64 `toml:"weight"`
	BumpsWarning int     `toml:"bumps_warning"`
	BumpDepth    int     `toml:"bump_depth"`
}

// RuleBranches handles complex_conditional thresholds.
type RuleBranches struct {
	Weight          float64 `toml:"weight"`
	BranchesWarning int     `toml:"branches_warning"`
	BranchesAlert   int     `toml:"branches_alert"`
}

// RuleLine handles thresholds with a single line-count threshold.
type RuleLine struct {
	Warning int     `toml:"warning"`
	Min     int     `toml:"min_lines"`
	Max     int     `toml:"max_consecutive"`
	Ratio   float64 `toml:"ratio"`
}

// RuleFloat handles thresholds with float/similarity values.
type RuleFloat struct {
	MinLines      int     `toml:"min_lines"`
	MinSimilarity float64 `toml:"min_similarity"`
}

// LoadProjectThresholds walks from dir upward to find .ailinter.toml,
// loads the defaults for the language, then merges user overrides.
func LoadProjectThresholds(dir, lang string) parser.Thresholds {
	t := parser.DefaultThresholds(lang)

	cfgPath := findConfig(dir)
	if cfgPath == "" {
		return t
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return t
	}

	var uc UserConfig
	if err := toml.Unmarshal(data, &uc); err != nil {
		return t
	}

	mergeOverrides(&t, &uc)
	return t
}

func findConfig(dir string) string {
	for {
		path := filepath.Join(dir, ".ailinter.toml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func LoadProjectConfigFile(cwd string) *ProjectConfig {
	cfgPath := findConfig(cwd)
	if cfgPath == "" {
		return nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}

	var uc UserConfig
	if err := toml.Unmarshal(data, &uc); err != nil {
		return nil
	}

	pc := &ProjectConfig{
		Path:    cfgPath,
		Extends: uc.Extends,
	}
	if uc.Rules.DeepNesting != nil {
		w := uc.Rules.DeepNesting.Warning
		pc.Rules.DeepNesting = &w
	}
	if uc.Rules.BrainMethod != nil {
		w := uc.Rules.BrainMethod.WarningLines
		pc.Rules.BrainMethod = &w
	}
	if uc.Rules.FileBloat != nil {
		w := uc.Rules.FileBloat.WarningLines
		pc.Rules.FileBloat = &w
	}
	if uc.Rules.CyclomaticComplexity != nil {
		w := uc.Rules.CyclomaticComplexity.Warning
		pc.Rules.CyclomaticComplexity = &w
	}
	if uc.Rules.BumpyRoad != nil {
		w := uc.Rules.BumpyRoad.BumpsWarning
		pc.Rules.BumpyRoad = &w
	}
	if uc.Rules.LongParameterList != nil {
		w := uc.Rules.LongParameterList.Warning
		pc.Rules.LongParameterList = &w
	}
	if uc.Rules.ComplexConditional != nil {
		w := uc.Rules.ComplexConditional.BranchesWarning
		pc.Rules.ComplexConditional = &w
	}
	if uc.Rules.LongSwitch != nil {
		w := uc.Rules.LongSwitch.Warning
		pc.Rules.LongSwitch = &w
	}
	if uc.Rules.ExcessiveComments != nil {
		r := uc.Rules.ExcessiveComments.Ratio
		pc.Rules.ExcessiveComments = &r
	}
	return pc
}

func mergeOverrides(t *parser.Thresholds, uc *UserConfig) {
	if uc.Rules.DeepNesting != nil {
		r := uc.Rules.DeepNesting
		if r.Warning > 0 {
			t.NestingWarning = r.Warning
		}
		if r.Alert > 0 {
			t.NestingAlert = r.Alert
		}
	}
	if uc.Rules.BrainMethod != nil {
		r := uc.Rules.BrainMethod
		if r.WarningLines > 0 {
			t.FuncLOCWarning = r.WarningLines
		}
		if r.AlertLines > 0 {
			t.FuncLOCAlert = r.AlertLines
		}
	}
	if uc.Rules.FileBloat != nil {
		r := uc.Rules.FileBloat
		if r.WarningLines > 0 {
			t.FileLOCWarning = r.WarningLines
		}
		if r.AlertLines > 0 {
			t.FileLOCAlert = r.AlertLines
		}
	}
	if uc.Rules.CyclomaticComplexity != nil {
		r := uc.Rules.CyclomaticComplexity
		if r.Warning > 0 {
			t.FuncCCWarning = r.Warning
		}
		if r.Alert > 0 {
			t.FuncCCAlert = r.Alert
		}
	}
	if uc.Rules.BumpyRoad != nil {
		r := uc.Rules.BumpyRoad
		if r.BumpsWarning > 0 {
			t.BumpyRoadBumpsWarning = r.BumpsWarning
		}
	}
	if uc.Rules.LongParameterList != nil {
		r := uc.Rules.LongParameterList
		if r.Warning > 0 {
			t.MaxArgumentsWarn = r.Warning
		}
		if r.Alert > 0 {
			t.MaxArgumentsAlert = r.Alert
		}
	}
	if uc.Rules.ComplexConditional != nil {
		r := uc.Rules.ComplexConditional
		if r.BranchesWarning > 0 {
			t.ComplexCondBranchesWarn = r.BranchesWarning
		}
		if r.BranchesAlert > 0 {
			t.ComplexCondBranchesAlert = r.BranchesAlert
		}
	}
	if uc.Rules.LazyElement != nil {
		r := uc.Rules.LazyElement
		if r.Min > 0 {
			t.LazyMinLines = r.Min
		}
	}
	if uc.Rules.ParagraphOfCode != nil {
		r := uc.Rules.ParagraphOfCode
		if r.Max > 0 {
			t.ParagraphMaxConsecutive = r.Max
		}
	}
	if uc.Rules.ExcessiveComments != nil {
		if uc.Rules.ExcessiveComments.Ratio > 0 {
			t.CommentRatioWarning = uc.Rules.ExcessiveComments.Ratio
		}
	}
	if uc.Rules.GlobalData != nil {
		if uc.Rules.GlobalData.Warning > 0 {
			t.GlobalDataWarning = uc.Rules.GlobalData.Warning
		}
	}
	if uc.Rules.LongScopeVariable != nil {
		if uc.Rules.LongScopeVariable.Min > 0 {
			t.LongScopeVarLines = uc.Rules.LongScopeVariable.Min
		}
	}
	if uc.Rules.LongSwitch != nil {
		r := uc.Rules.LongSwitch
		if r.Warning > 0 {
			t.LongSwitchWarn = r.Warning
		}
		if r.Alert > 0 {
			t.LongSwitchAlert = r.Alert
		}
	}
	if uc.Rules.DuplicatedCode != nil {
		r := uc.Rules.DuplicatedCode
		if r.MinLines > 0 {
			t.DupMinLines = r.MinLines
		}
		if r.MinSimilarity > 0 {
			t.DupMinSimilarity = r.MinSimilarity
		}
	}
}
