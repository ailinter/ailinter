package parser

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// HotspotEntry represents a file with its commit frequency.
type HotspotEntry struct {
	FilePath     string
	CommitCount  int
	Authors      int
	Lines        int
	QualityScore float64
	Priority     float64 // composite score: commits * (11 - qualityScore)
}

// GitHotspotResult holds the full hotspot analysis for a repo.
type GitHotspotResult struct {
	Entries  []HotspotEntry
	RepoPath string
	Error    string
}

// AnalyzeGitHotspots scans the git log to find frequently-changed files.
// Returns files ranked by composite priority (high commits + low health).
func AnalyzeGitHotspots(repoPath string, maxDepth int) GitHotspotResult {
	result := GitHotspotResult{RepoPath: repoPath}
	if maxDepth <= 0 {
		maxDepth = 500
	}

	// git log --numstat --format="%H %an" -n <maxDepth>
	cmd := exec.Command("git", "log", "--numstat", "--format=%H %an", "-n", strconv.Itoa(maxDepth))
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Sprintf("git log failed: %v", err)
		return result
	}

	entries := parseGitLog(string(out))
	result.Entries = rankHotspots(entries)
	return result
}

func parseGitLog(output string) []HotspotEntry {
	fileMap := make(map[string]*HotspotEntry)
	lines := strings.Split(output, "\n")
	var currentAuthors map[string]bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		currentAuthors = processCommitLine(line, currentAuthors)
		processNumstatLine(line, fileMap, currentAuthors)
	}

	result := make([]HotspotEntry, 0, len(fileMap))
	for _, e := range fileMap {
		result = append(result, *e)
	}
	return result
}

func processCommitLine(line string, currentAuthors map[string]bool) map[string]bool {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return currentAuthors
	}
	if strings.HasPrefix(line, "commit ") || (len(parts[0]) >= 40 && len(parts) != 3) {
		if len(parts[0]) >= 40 && looksLikeHash(parts[0]) && len(parts) == 2 {
			m := make(map[string]bool)
			m[parts[1]] = true
			return m
		}
		return nil
	}
	return currentAuthors
}

func processNumstatLine(line string, fileMap map[string]*HotspotEntry, currentAuthors map[string]bool) {
	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		return
	}
	filePath := parts[2]
	if !isSourceFileExt(filePath) {
		return
	}
	e, ok := fileMap[filePath]
	if !ok {
		e = &HotspotEntry{FilePath: filePath}
		fileMap[filePath] = e
	}
	e.CommitCount++
	if currentAuthors != nil && e.Authors == 0 {
		e.Authors = len(currentAuthors)
	}
}

func looksLikeHash(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func rankHotspots(entries []HotspotEntry) []HotspotEntry {
	for i, e := range entries {
		entries[i].Priority = float64(e.CommitCount) * (11.0 - e.QualityScore)
	}
	return entries
}

// DetectHotspotSmell identifies files that are both frequently changed AND unhealthy.
func DetectHotspotSmell(entry HotspotEntry) *Smell {
	if entry.CommitCount < 5 || entry.QualityScore > 7.0 {
		return nil
	}
	sev := "warning"
	if entry.CommitCount >= 20 {
		sev = "alert"
	}
	if entry.CommitCount >= 50 && entry.QualityScore < 4.0 {
		sev = "critical"
	}
	return &Smell{
		Name:     "hotspot",
		Severity: sev,
		Message:  fmt.Sprintf("churn: %s — %d commits, health: %.1f, priority: %.1f", entry.FilePath, entry.CommitCount, entry.QualityScore, entry.Priority),
		AIPrompt: fmt.Sprintf("HOTSPOT: '%s' is changed frequently (%d commits) and has low quality (%.1f). HIGHEST PRIORITY for refactoring.", entry.FilePath, entry.CommitCount, entry.QualityScore),
	}
}

// Internal caching for git results to avoid repeated git calls.
var (
	hotspotCache   = make(map[string]GitHotspotResult)
	hotspotCacheMu sync.Mutex
)

// GetCachedHotspots returns cached hotspot analysis or runs a new one.
func GetCachedHotspots(repoPath string, maxDepth int) GitHotspotResult {
	hotspotCacheMu.Lock()
	defer hotspotCacheMu.Unlock()

	key := fmt.Sprintf("%s:%d", repoPath, maxDepth)
	if cached, ok := hotspotCache[key]; ok {
		return cached
	}
	result := AnalyzeGitHotspots(repoPath, maxDepth)
	hotspotCache[key] = result
	return result
}

func isSourceFileExt(path string) bool {
	ext := ""
	idx := strings.LastIndex(path, ".")
	if idx != -1 {
		ext = path[idx:]
	}
	switch ext {
	case ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".java", ".rs", ".rb",
		".c", ".cpp", ".cc", ".cxx", ".h", ".hpp",
		".cs", ".swift", ".kt", ".scala", ".php", ".pl",
		".sh", ".bash", ".tf", ".yaml", ".yml", ".toml",
		".json", ".xml", ".html", ".css", ".sql":
		return true
	}
	return false
}
