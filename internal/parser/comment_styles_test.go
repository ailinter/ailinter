package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/comments/*.txt
var commentFixtures embed.FS

func commentLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := commentFixtures.ReadFile("testdata/comments/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

type commentCase struct {
	lang    string
	fixture string
	want    bool
}

var commentCases = []commentCase{
	{"python", "python_docstring.txt", true},
	{"html", "html_comments.txt", true},
	{"sql", "sql_comments.txt", true},
}

func TestDetectExcessiveCommentsFor(t *testing.T) {
	for _, tc := range commentCases {
		t.Run(tc.fixture, func(t *testing.T) {
			smell := parser.DetectExcessiveCommentsFor(tc.lang, commentLines(t, tc.fixture), 0.3)
			if (smell != nil) != tc.want {
				t.Errorf("excessive_comments present = %v, want %v", smell != nil, tc.want)
			}
		})
	}
}
