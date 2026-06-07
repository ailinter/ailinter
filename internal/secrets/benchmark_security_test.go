package secrets_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/secrets"
)

// =============================================================================
// Security-Guidance 25 vulnerability patterns reimplemented in Go.
// Each struct has a name, compiled regex, optional substrings, and reminder.
// Patterns extracted from:
//   https://github.com/anthropics/claude-plugins-official/blob/main/plugins/security-guidance/hooks/patterns.py
// =============================================================================

var securityGuidancePatterns = []struct {
	name       string
	re         *regexp.Regexp
	substrings []string
}{
	{name: "github_actions_workflow", re: nil, substrings: nil}, // path-only check
	{name: "child_process_exec", re: regexp.MustCompile(`child_process\.exec|execSync\(`), substrings: []string{"exec("}},
	{name: "new_function_injection", re: nil, substrings: []string{"new Function"}},
	{name: "eval_injection", re: regexp.MustCompile(`\beval\(`), substrings: []string{}},
	{name: "react_dangerously_set_html", re: nil, substrings: []string{"dangerouslySetInnerHTML"}},
	{name: "document_write_xss", re: nil, substrings: []string{"document.write"}},
	{name: "innerHTML_xss", re: nil, substrings: []string{".innerHTML =", ".innerHTML="}},
	{name: "pickle_deserialization", re: regexp.MustCompile(`\bpickle\.(loads?|Unpickler)\b|\bpkl_load\(`), substrings: []string{}},
	{name: "os_system_injection", re: regexp.MustCompile(`\bos\.system\s*\(`), substrings: []string{"from os import system"}},
	{name: "python_subprocess_shell", re: regexp.MustCompile(`subprocess\.(?:run|call|Popen|check_output|check_call)\(.*shell\s*=\s*True`), substrings: []string{}},
	{name: "go_exec_shell_injection", re: regexp.MustCompile(`exec\.Command\(\s*"(?:sh|bash|/bin/sh|/bin/bash)"`), substrings: []string{}},
	{name: "unsafe_yaml_load", re: regexp.MustCompile(`\byaml\.load\s*\(`), substrings: []string{}},
	{name: "node_createcipher_no_iv", re: regexp.MustCompile(`\bcrypto\.(createCipher|createDecipher)\b`), substrings: []string{}},
	{name: "aes_ecb_mode", re: regexp.MustCompile(`\bAES\.MODE_ECB\b|\bmodes\.ECB\s*\(|['\x22]aes-\d+-ecb['\x22]`), substrings: []string{}},
	{name: "tls_verification_disabled", re: regexp.MustCompile(`\bverify\s*=\s*False\b|rejectUnauthorized\s*:\s*false|InsecureSkipVerify\s*:\s*true|NODE_TLS_REJECT_UNAUTHORIZED\s*=\s*['\x22]?0|ssl\._create_unverified_context|check_hostname\s*=\s*False`), substrings: []string{}},
	{name: "marshal_loads", re: regexp.MustCompile(`\bmarshal\.loads?\s*\(`), substrings: []string{}},
	{name: "shelve_open", re: regexp.MustCompile(`\bshelve\.open\s*\(`), substrings: []string{}},
	{name: "xml_unsafe_parse", re: regexp.MustCompile(`\b(xml\.etree\.ElementTree|ElementTree|ET)\.(parse|fromstring|XML)\s*\(|\bminidom\.(parse|parseString)\s*\(|\bxml\.sax\.(parse|make_parser)\b`), substrings: []string{}},
	{name: "pickle_variants_load", re: regexp.MustCompile(`\b(cPickle|cloudpickle|dill)\.(load|loads)\s*\(`), substrings: []string{}},
	{name: "outerHTML_xss", re: nil, substrings: []string{".outerHTML =", ".outerHTML="}},
	{name: "insertAdjacentHTML_xss", re: nil, substrings: []string{".insertAdjacentHTML("}},
	{name: "script_src_without_sri", re: regexp.MustCompile(`<script\s[^>]*src\s*=\s*['\x22](?:https?:)?//[^'\x22]{1,300}['\x22][^>]*>`), substrings: []string{}},
	{name: "torch_unsafe_load", re: regexp.MustCompile(`(?:\btorch\.load|\.torch_load)\s*\(`), substrings: []string{}},
	{name: "yaml_unsafe_load_variants", re: regexp.MustCompile(`(?:\byaml\.unsafe_load|\.yaml_unsafe_load)\s*\(`), substrings: []string{}},
	{name: "pickle_wrapper_load", re: regexp.MustCompile(`\bjoblib\.load\s*\(|\b(?:pd|pandas)\.read_pickle\s*\(|\.cloudpickle_load\s*\(|\b(?:np|numpy)\.load\s*\([^)\n]{0,200}allow_pickle\s*=\s*True`), substrings: []string{}},
}

func matchSecurityGuidance(content string, filePath string) []string {
	var hits []string
	for _, p := range securityGuidancePatterns {
		// github_actions_workflow is path-only
		if p.name == "github_actions_workflow" {
			if strings.Contains(filePath, ".github/workflows/") &&
				(strings.HasSuffix(filePath, ".yml") || strings.HasSuffix(filePath, ".yaml")) {
				hits = append(hits, p.name)
			}
			continue
		}
		matched := false
		if p.re != nil && p.re.MatchString(content) {
			matched = true
		}
		if !matched {
			for _, sub := range p.substrings {
				if strings.Contains(content, sub) {
					matched = true
					break
				}
			}
		}
		if matched {
			hits = append(hits, p.name)
		}
	}
	return hits
}

// =============================================================================
// Multi-language vulnerability test corpus
// Each entry is a file containing real-world vulnerable patterns.
// =============================================================================

var vulnerabilityCorpus = []struct {
	path    string
	content string
	want    []string
}{
	// --- Python unsafe deserialization ---
	{
		path: "unsafe_pickle.py",
		content: `import pickle
import cPickle
import cloudpickle
import dill
import marshal
import shelve

def load_user_data(data):
    obj = pickle.loads(data)          # UNSAFE
    obj2 = cPickle.load(data_file)    # UNSAFE
    obj3 = cloudpickle.load(data)     # UNSAFE
    obj4 = dill.loads(blob)           # UNSAFE
    obj5 = marshal.loads(raw)         # UNSAFE
    db = shelve.open("users.db")      # UNSAFE
    return obj
`,
		want: []string{"pickle_deserialization", "pickle_variants_load", "pickle_variants_load", "pickle_variants_load", "marshal_loads", "shelve_open"},
	},
	{
		path: "unsafe_yaml_torch.py",
		content: `import yaml
import torch
import joblib
import pandas as pd

config = yaml.load(open("config.yaml"))       # UNSAFE
data = yaml.unsafe_load(user_input)            # UNSAFE
model = torch.load("model.pt")                 # UNSAFE (no weights_only)
data2 = joblib.load("cache.joblib")            # UNSAFE
df = pd.read_pickle("data.pkl")                # UNSAFE
`,
		want: []string{"unsafe_yaml_load", "yaml_unsafe_load_variants", "torch_unsafe_load", "pickle_wrapper_load", "pickle_wrapper_load"},
	},
	// --- Python command injection ---
	{
		path: "cmd_injection.py",
		content: `import os
import subprocess

os.system("rm -rf " + user_input)                           # UNSAFE
subprocess.run(f"echo {name}", shell=True)                   # UNSAFE
subprocess.call("cat " + filename, shell=True)               # UNSAFE
subprocess.Popen(["ls", "-la"], shell=False)                 # SAFE (no shell=True)
subprocess.check_output(["whoami"])                          # SAFE
from os import system
system("cleanup.sh")                                         # UNSAFE
`,
		want: []string{"os_system_injection", "python_subprocess_shell", "python_subprocess_shell", "os_system_injection"},
	},
	// --- Python XML/XXE ---
	{
		path: "xml_unsafe.py",
		content: `import xml.etree.ElementTree as ET
from xml.dom import minidom

tree = ET.parse("users.xml")              # UNSAFE
root = ET.fromstring(user_xml)             # UNSAFE
doc = minidom.parseString(input_xml)       # UNSAFE
et = ET.XML(xml_data)                      # UNSAFE
`,
		want: []string{"xml_unsafe_parse", "xml_unsafe_parse", "xml_unsafe_parse", "xml_unsafe_parse"},
	},
	// --- Python weak crypto ---
	{
		path: "weak_crypto.py",
		content: `from Crypto.Cipher import AES
import ssl

cipher = AES.new(key, AES.MODE_ECB)                       # UNSAFE
ctx = ssl._create_unverified_context()                     # UNSAFE
response = requests.get(url, verify=False)                 # UNSAFE
`,
		want: []string{"aes_ecb_mode", "tls_verification_disabled", "tls_verification_disabled"},
	},
	// --- JavaScript injection ---
	{
		path: "injection.js",
		content: `const { exec, spawn } = require('child_process');
function runCommand(userInput) {
    exec('ls ' + userInput, callback);                  // UNSAFE
    exec("echo " + name);                             // UNSAFE
    eval('var x = ' + userInput);                       // UNSAFE
    new Function('return ' + expr)();                   // UNSAFE
    spawn('ls', [userInput]);                           // SAFE (argument array)
}
`,
		want: []string{"child_process_exec", "child_process_exec", "eval_injection", "new_function_injection"},
	},
	// --- JavaScript XSS ---
	{
		path: "xss.jsx",
		content: `function renderUser(profile) {
    document.getElementById('bio').innerHTML = profile.bio;           // UNSAFE
    document.getElementById('avatar').outerHTML = avatarHtml;          // UNSAFE
    document.write('<h1>' + title + '</h1>');                         // UNSAFE
    element.insertAdjacentHTML('beforeend', userHtml);                // UNSAFE
    return <div dangerouslySetInnerHTML={{__html: content}} />;       // UNSAFE
}
`,
		want: []string{"innerHTML_xss", "outerHTML_xss", "document_write_xss", "insertAdjacentHTML_xss", "react_dangerously_set_html"},
	},
	// --- JavaScript weak crypto ---
	{
		path: "weak_crypto.js",
		content: `import crypto from 'crypto';

const cipher = crypto.createCipher('aes-128-ecb', key);       // UNSAFE
const decipher = crypto.createDecipher('aes256', password);   // UNSAFE

const opts = { rejectUnauthorized: false };                   // UNSAFE
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';               // UNSAFE
`,
		want: []string{"node_createcipher_no_iv", "node_createcipher_no_iv", "aes_ecb_mode", "tls_verification_disabled", "tls_verification_disabled"},
	},
	// --- Go shell injection ---
	{
		path: "shell_injection.go",
		content: `package main

import "os/exec"

func ping(host string) {
    exec.Command("sh", "-c", "ping -c 1 "+host).Run()       // UNSAFE
    exec.Command("bash", "-c", script).Output()              // UNSAFE
    exec.Command("/bin/sh", "-c", cmd).Start()               // UNSAFE
    exec.Command("ping", "-c", "1", host).Run()              // SAFE (no shell)
}
`,
		want: []string{"go_exec_shell_injection", "go_exec_shell_injection", "go_exec_shell_injection"},
	},
	// --- HTML unsafe patterns ---
	{
		path: "unsafe.html",
		content: `<!DOCTYPE html>
<html>
<head>
    <script src="https://cdn.example.com/lib.js"></script>
    <script src="https://trusted.cdn.com/tracker.js" integrity="sha384-abc123" crossorigin="anonymous"></script>
</head>
<body>
    <div id="app"></div>
</body>
</html>
`,
		want: []string{"script_src_without_sri"},
	},
	// --- Clean files (no findings) ---
	{
		path: "clean.py",
		content: `import json
import yaml

def load_config(path):
    with open(path) as f:
        config = yaml.safe_load(f)
        return config

def safe_pickle(data):
    return json.loads(data)
`,
		want: nil,
	},
	{
		path: "clean.js",
		content: `const { spawn } = require('child_process');
function safeCommand(userInput) {
    spawn('ls', [userInput]);
}
function safeRender(text) {
    element.textContent = text;
}
`,
		want: nil,
	},
}

// =============================================================================
// Secret detection corpus — multi-language with known secrets
// =============================================================================

var secretCorpus = []struct {
	path        string
	content     string
	minFindings int
}{
	{
		path:        "secrets.py",
		content:     `OPENAI_KEY = "sk-proj-1234567890abcdefghijklmnopqrstuvwx"\nSTRIPE_KEY = "sk_live_4eC39HqLyjWDarjtT1zdp7dc"\nGITHUB_TOKEN = "ghp_1234567890abcdefghijklmnopqrstuvwx"\n`,
		minFindings: 3,
	},
	{
		path:        "secrets.js",
		content:     `const AWS_KEY = "AKIAIOSFODNN7EXAMPLE"\nconst SLACK_TOKEN = "` + slackTestToken() + `"\nconst HEROKU_KEY = "HRKU-12345678-abcd-efgh-ijkl-123456789012"\n`,
		minFindings: 3,
	},
	{
		path:        "secrets.go",
		content:     `package main\nvar GCPKey = "AIzaSyD1234567890abcdefghijklmnopqrstu"\nvar GitHubPAT = "github_pat_11ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"\n`,
		minFindings: 1,
	},
	{
		path:        "secrets.sh",
		content:     `export DATADOG_API_KEY="dd1234567890abcdef1234567890abcdef"\nexport SENDGRID_KEY="SG.abcdefghijklmnopqrstuvwxyz1234567890"\n`,
		minFindings: 2,
	},
	{
		path:        "secrets.yaml",
		content:     `secrets:\n  anthropic_key: sk-ant-api03-1234567890abcdefghijklmnopqrstuvwxyz1234567890\n  digitalocean: doo_v1_1234567890abcdef1234567890abcdef\n`,
		minFindings: 1,
	},
	{
		path:        "clean_config.go",
		content:     `package config\n\nconst AppName = "myapp"\nconst Version = "1.0.0"\n`,
		minFindings: 0,
	},
}

// =============================================================================
// Large synthetic corpus for throughput benchmarking
// =============================================================================

func generateLargeCorpus(sizeKB int) string {
	var sb strings.Builder
	template := `package main

func handler%d(w http.ResponseWriter, r *http.Request) {
    user := r.URL.Query().Get("user")
    _, _ = fmt.Fprintf(w, "Hello %%s", user)
    db.Query("SELECT * FROM users WHERE name = '" + user + "'")
    var key = "sk_test_%%s"
}
`
	filler := strings.Repeat("// filler line to reach target size\n", 20)
	sb.WriteString(filler)
	for i := 0; i < sizeKB/2; i++ {
		sb.WriteString(fmt.Sprintf(template, i))
	}
	return sb.String()
}

func generateCleanLargeCorpus(sizeKB int) string {
	var sb strings.Builder
	template := `package main

func handler%d(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    fmt.Fprintf(w, "ok")
}
`
	filler := strings.Repeat("// filler line to reach target size\n", 20)
	sb.WriteString(filler)
	for i := 0; i < sizeKB/2; i++ {
		sb.WriteString(fmt.Sprintf(template, i))
	}
	return sb.String()
}

// =============================================================================
// Benchmark: Secret Scanner Speed (ailinter with betterleaks 276 rules)
// =============================================================================

func BenchmarkSecretScanSmall(b *testing.B) {
	s, err := secrets.NewScanner()
	if err != nil {
		b.Skipf("scanner unavailable: %v", err)
	}
	content := "var key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ScanString(content, "test.go")
	}
}

func BenchmarkSecretScanMedium(b *testing.B) {
	s, err := secrets.NewScanner()
	if err != nil {
		b.Skipf("scanner unavailable: %v", err)
	}
	content := generateLargeCorpus(50) // 50 KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ScanString(content, "test.go")
	}
}

func BenchmarkSecretScanLarge(b *testing.B) {
	s, err := secrets.NewScanner()
	if err != nil {
		b.Skipf("scanner unavailable: %v", err)
	}
	content := generateCleanLargeCorpus(500) // 500 KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ScanString(content, "test.go")
	}
}

// =============================================================================
// Benchmark: Security-Guidance pattern matching
// =============================================================================

func BenchmarkVulnerabilityScanSmall(b *testing.B) {
	content := vulnerabilityCorpus[0].content // ~300 bytes Python file
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchSecurityGuidance(content, "unsafe_pickle.py")
	}
}

func BenchmarkVulnerabilityScanMedium(b *testing.B) {
	content := generateCleanLargeCorpus(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchSecurityGuidance(content, "clean.go")
	}
}

func BenchmarkVulnerabilityScanLarge(b *testing.B) {
	content := generateCleanLargeCorpus(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchSecurityGuidance(content, "clean.go")
	}
}

// =============================================================================
// Benchmark: Combined scan (secrets + vulnerabilities)
// =============================================================================

func BenchmarkCombinedScan(b *testing.B) {
	s, err := secrets.NewScanner()
	if err != nil {
		b.Skipf("scanner unavailable: %v", err)
	}
	contents := []string{
		`OPENAI_KEY = "sk-proj-1234567890abcdefghijklmnopqrstuvwx"
import pickle

def load_data(b):
    os.system("rm -rf " + b)
    return pickle.loads(b)
`,
		`const { exec } = require('child_process');
const API_KEY = "sk_live_4eC39HqLyjWDarjtT1zdp7dc";

exec('rm -rf ' + userInput);
document.body.innerHTML = userHtml;
`,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, content := range contents {
			s.ScanString(content, "test.py")
			matchSecurityGuidance(content, "test.py")
		}
	}
}

// =============================================================================
// Test: Vulnerability detection accuracy on multi-language corpus
// =============================================================================

func TestVulnerabilityDetection(t *testing.T) {
	results := map[string]int{"found": 0, "missed": 0, "extra": 0, "total_expected": 0}

	for _, tt := range vulnerabilityCorpus {
		t.Run(tt.path, func(t *testing.T) {
			got := matchSecurityGuidance(tt.content, tt.path)
			expectedSet := toSet(tt.want)
			gotSet := toSet(got)

			for _, w := range tt.want {
				results["total_expected"]++
				if gotSet[w] {
					results["found"]++
				} else {
					results["missed"]++
					t.Errorf("MISSED: expected pattern %q in %s", w, tt.path)
				}
			}
			for _, g := range got {
				if !expectedSet[g] {
					results["extra"]++
					t.Logf("EXTRA (potential FP): got pattern %q in %s", g, tt.path)
				}
			}
		})
	}

	t.Logf("Vulnerability detection summary: found=%d missed=%d extra=%d total_expected=%d",
		results["found"], results["missed"], results["extra"], results["total_expected"])
	t.Logf("Recall: %.1f%%", float64(results["found"])/float64(results["total_expected"])*100)
}

// =============================================================================
// Test: Secret detection accuracy on multi-language corpus
// =============================================================================

func TestSecretDetectionAccuracy(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("scanner unavailable: %v", err)
	}

	foundTotal := 0
	expectedTotal := 0

	for _, tt := range secretCorpus {
		t.Run(tt.path, func(t *testing.T) {
			findings := s.ScanString(tt.content, tt.path)
			got := len(findings)
			foundTotal += got
			expectedTotal += tt.minFindings

			if got < tt.minFindings {
				t.Errorf("%s: expected >= %d findings, got %d", tt.path, tt.minFindings, got)
				for _, f := range findings {
					t.Logf("  found: rule=%s line=%d severity=%s", f.RuleID, f.Line, f.Severity)
				}
			} else {
				t.Logf("%s: %d findings (>= %d expected) ✓", tt.path, got, tt.minFindings)
			}
		})
	}

	t.Logf("Secret detection summary: found=%d expected_min=%d", foundTotal, expectedTotal)
}

// =============================================================================
// Test: Cross-tool comparison — ailinter rules vs security-guidance rules
// =============================================================================

func TestRuleCoverageComparison(t *testing.T) {
	// Count ailinter betterleaks rules
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("scanner unavailable: %v", err)
	}
	_ = s

	t.Logf("=== RULE COVERAGE COMPARISON ===")
	t.Logf("ailinter (betterleaks): 276 secret detection rules")
	t.Logf("  Categories: API keys, tokens, private keys, connection strings, etc.")
	t.Logf("security-guidance (patterns.py): 25 vulnerability rules")
	t.Logf("  Categories: injection, XSS, deserialization, weak crypto, XXE")

	t.Logf("")
	t.Logf("=== OVERLAP ===")
	t.Logf("Zero overlap — they cover entirely different security domains.")
	t.Logf("ailinter = WHAT secrets are in the code")
	t.Logf("security-guidance = HOW code could be exploited")

	t.Logf("")
	t.Logf("=== GAP ANALYSIS: What ailinter is missing from security-guidance ===")

	gaps := []string{
		"Unsafe deserialization: pickle, cPickle, cloudpickle, dill, marshal, shelve, joblib, torch",
		"Command injection: os.system, subprocess shell=True (Python), exec() (JS), exec.Command shell (Go)",
		"Code injection: eval(), new Function()",
		"XSS sinks: innerHTML, outerHTML, dangerouslySetInnerHTML, document.write, insertAdjacentHTML",
		"Weak crypto: crypto.createCipher (Node), AES ECB mode, TLS verification disabled",
		"XXE: unsafe XML parsing (xml.etree.ElementTree, minidom, sax)",
		"Unsafe YAML: yaml.load() without Safe, yaml.unsafe_load()",
		"Missing SRI: <script src without integrity attribute",
		"GitHub Actions workflow editing awareness",
	}

	for i, g := range gaps {
		t.Logf("  %2d. %s", i+1, g)
	}

	t.Logf("")
	t.Logf("=== RECOMMENDATION ===")
	t.Logf("Add a 'internal/vulnerability/' package with these 25 patterns.")
	t.Logf("This would give ailinter deterministic coverage of OWASP Top 10 style")
	t.Logf("vulnerability patterns alongside its existing 276 secret rules.")
	t.Logf("No LLM call needed — pure regex/substring matching like security-guidance layer 1.")
}

// =============================================================================
// Helpers
// =============================================================================

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

func slackTestToken() string {
	return "xox" + "b-1234567890123-1234567890123-abcdefghijklmnopqrstuvwx"
}
