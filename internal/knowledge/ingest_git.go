package knowledge

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitCache stores cached git history for fast reboots.
type GitCache struct {
	LastCommitHash string        `json:"last_commit_hash"`
	Commits        []CommitEntry `json:"commits"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// CommitEntry represents a single git commit in the knowledge graph.
type CommitEntry struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Files     []string  `json:"files"`
}

// bugKeywords are words that indicate a bug-fixing commit.
var bugKeywords = []string{
	"fix", "bug", "regression", "hotfix", "incident",
	"patch", "workaround", "crash", "panic", "leak",
	"vuln", "vulnerability", "cve", "security",
	"defect", "error", "issue", "rollback",
}

const maxCommits = 500
const gitLogFormat = "%H|%ai|%s"

// IngestGitHistory ingests git history into the graph, using a cache for performance.
func IngestGitHistory(graph *Graph, repoPath string) error {
	Logf("ingesting git history from %s", repoPath)

	cache, err := loadGitCache(graph.KnowledgeDir)
	if err != nil {
		Logf("warning: could not load git cache: %v, rebuilding", err)
		cache = &GitCache{}
	}

	// Get current HEAD
	headHash, err := runGit(repoPath, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	headHash = strings.TrimSpace(headHash)

	// If HEAD matches cache and we have commits, load from cache
	if headHash == cache.LastCommitHash && len(cache.Commits) > 0 {
		Logf("git cache hit: %d commits loaded from cache", len(cache.Commits))
		for _, entry := range cache.Commits {
			addCommitToGraph(graph, entry)
		}
		return nil
	}

	// Determine the time to fetch from
	sinceTime := ""
	if cache.LastCommitHash != "" && len(cache.Commits) > 0 {
		// Use the timestamp of the last cached commit
		lastTime := cache.Commits[len(cache.Commits)-1].Timestamp
		sinceTime = "--since=" + lastTime.Format(time.RFC3339)
	}

	// Run git log to get commits and their files
	logArgs := []string{"--all", fmt.Sprintf("--format=%s", gitLogFormat), "--name-only"}
	if sinceTime != "" {
		logArgs = append(logArgs, sinceTime)
	}

	cmd := exec.Command("git", append([]string{"-C", repoPath, "log"}, logArgs...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git log: %s: %w", strings.TrimSpace(string(output)), err)
	}
	logOutput := string(output)

	entries := parseGitLog(logOutput)

	// Merge with existing cached commits (prepend new ones)
	var allEntries []CommitEntry
	existingHashes := make(map[string]bool)
	for _, e := range cache.Commits {
		existingHashes[e.Hash] = true
	}

	for _, e := range entries {
		if !existingHashes[e.Hash] {
			allEntries = append(allEntries, e)
			existingHashes[e.Hash] = true
		}
	}
	allEntries = append(allEntries, cache.Commits...)

	// Limit to maxCommits
	if len(allEntries) > maxCommits {
		allEntries = allEntries[:maxCommits]
	}

	// Add to graph
	for _, entry := range allEntries {
		addCommitToGraph(graph, entry)
	}

	// Save cache
	cache.LastCommitHash = headHash
	cache.Commits = allEntries
	cache.UpdatedAt = time.Now()

	if err := saveGitCache(graph.KnowledgeDir, cache); err != nil {
		Logf("warning: could not save git cache: %v", err)
	}

	Logf("ingested %d commits from git history", len(allEntries))
	return nil
}

// addCommitToGraph adds a single commit entry to the knowledge graph.
func addCommitToGraph(graph *Graph, entry CommitEntry) {
	commitID := slug("commit", entry.Hash[:12])
	if _, exists := graph.GetNode(commitID); exists {
		return // already ingested
	}

	shortMsg := truncate(entry.Message, 72)
	graph.AddNode(Node{
		ID:    commitID,
		Type:  NodeCommit,
		Label: shortMsg,
		Properties: map[string]interface{}{
			"hash":      entry.Hash,
			"timestamp": entry.Timestamp.Format(time.RFC3339),
			"message":   entry.Message,
			"files":     entry.Files,
		},
	})

	// CHANGED edges to affected files
	for _, f := range entry.Files {
		fileNodeID := slug("file", f)
		if _, exists := graph.GetNode(fileNodeID); exists {
			graph.AddEdge(commitID, fileNodeID, EdgeChanged, nil)
		}
	}

	// Bug detection
	if isBugCommit(entry.Message) {
		bugID := slug("bug", entry.Hash[:12])
		if _, exists := graph.GetNode(bugID); !exists {
			severity := "low"
			msg := strings.ToLower(entry.Message)
			if strings.Contains(msg, "critical") || strings.Contains(msg, "security") || strings.Contains(msg, "cve") {
				severity = "critical"
			} else if strings.Contains(msg, "regression") || strings.Contains(msg, "crash") || strings.Contains(msg, "panic") {
				severity = "high"
			} else if strings.Contains(msg, "fix") || strings.Contains(msg, "bug") {
				severity = "medium"
			}

			graph.AddNode(Node{
				ID:    bugID,
				Type:  NodeBug,
				Label: fmt.Sprintf("Bug: %s", shortMsg),
				Properties: map[string]interface{}{
					"commit":   entry.Hash,
					"severity": severity,
					"message":  entry.Message,
				},
			})

			// Bug → Commit
			graph.AddEdge(bugID, commitID, EdgeCausedBy, nil)
		}
	}
}

// isBugCommit checks if a commit message indicates a bug fix.
func isBugCommit(msg string) bool {
	msg = strings.ToLower(msg)
	for _, kw := range bugKeywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

// parseGitLog parses the output of `git log --all --format=... --name-only`.
func parseGitLog(output string) []CommitEntry {
	var entries []CommitEntry
	var current *CommitEntry

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		// Check if this line starts a new commit (format: HASH|TIMESTAMP|MESSAGE)
		if strings.Contains(line, "|") && strings.Count(line, "|") >= 2 {
			if current != nil {
				entries = append(entries, *current)
			}

			parts := strings.SplitN(line, "|", 3)
			if len(parts) < 3 {
				current = nil
				continue
			}

			ts, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
			if err != nil {
				// Try alternative format
				ts, err = time.Parse("2006-01-02T15:04:05-07:00", parts[1])
				if err != nil {
					ts = time.Now()
				}
			}

			current = &CommitEntry{
				Hash:      strings.TrimSpace(parts[0]),
				Timestamp: ts,
				Message:   strings.TrimSpace(parts[2]),
			}
		} else if current != nil {
			// This is a file path line
			filePath := strings.TrimSpace(line)
			if filePath != "" {
				current.Files = append(current.Files, filePath)
			}
		}
	}

	// Don't forget the last one
	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}

// runGit runs a git command and returns the output.
func runGit(repoPath, arg string, extraArgs ...string) (string, error) {
	args := append([]string{"-C", repoPath, arg}, extraArgs...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", arg, strings.TrimSpace(string(output)), err)
	}
	return string(output), nil
}

// loadGitCache loads the git cache from the given knowledge directory.
func loadGitCache(knowledgeDir string) (*GitCache, error) {
	path := filepath.Join(knowledgeDir, "git-cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return nil, fmt.Errorf("empty cache file")
	}
	var cache GitCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("unmarshal cache: %w", err)
	}
	return &cache, nil
}

// saveGitCache saves the git cache to the given knowledge directory.
func saveGitCache(knowledgeDir string, cache *GitCache) error {
	path := filepath.Join(knowledgeDir, "git-cache.json")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
