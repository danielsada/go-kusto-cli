package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_ExecuteFlag(t *testing.T) {
	result, err := Resolve("StormEvents | count", "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if result.Source != SourceFlag {
		t.Errorf("expected source flag, got %s", result.Source)
	}
	if len(result.Queries) != 1 || result.Queries[0] != "StormEvents | count" {
		t.Errorf("unexpected queries: %v", result.Queries)
	}
}

func TestResolve_ScriptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.kql")
	content := "StormEvents | count\n;\nStormEvents | take 10\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test script: %v", err)
	}

	result, err := Resolve("", path)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if result.Source != SourceScript {
		t.Errorf("expected source script, got %s", result.Source)
	}
	if len(result.Queries) != 2 {
		t.Fatalf("expected 2 queries, got %d: %v", len(result.Queries), result.Queries)
	}
	if result.Queries[0] != "StormEvents | count" {
		t.Errorf("query 0: %q", result.Queries[0])
	}
	if result.Queries[1] != "StormEvents | take 10" {
		t.Errorf("query 1: %q", result.Queries[1])
	}
}

func TestResolve_FlagPriority(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.kql")
	if err := os.WriteFile(path, []byte("ignored query"), 0o644); err != nil {
		t.Fatalf("writing test script: %v", err)
	}

	result, err := Resolve("from flag", path)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if result.Source != SourceFlag {
		t.Errorf("expected flag to take priority, got source %s", result.Source)
	}
	if result.Queries[0] != "from flag" {
		t.Errorf("expected flag query, got %q", result.Queries[0])
	}
}

func TestResolve_MissingScriptFile(t *testing.T) {
	_, err := Resolve("", "/nonexistent/file.kql")
	if err == nil {
		t.Error("expected error for missing script file, got nil")
	}
}

func TestSplitQueries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "semicolon delimited",
			input:    "query1\n;\nquery2",
			expected: []string{"query1", "query2"},
		},
		{
			name:     "blank line delimited",
			input:    "query1\n\nquery2",
			expected: []string{"query1", "query2"},
		},
		{
			name:     "multiline query",
			input:    "StormEvents\n| where State == 'TEXAS'\n| count\n;\nother",
			expected: []string{"StormEvents\n| where State == 'TEXAS'\n| count", "other"},
		},
		{
			name:     "trailing semicolon",
			input:    "query1\n;",
			expected: []string{"query1"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitQueries(tc.input)
			if len(got) != len(tc.expected) {
				t.Fatalf("splitQueries: got %d queries %v, want %d %v", len(got), got, len(tc.expected), tc.expected)
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("query %d: got %q, want %q", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestSplitQueries_WhitespaceOnly(t *testing.T) {
	got := splitQueries("   \n\n   \n")
	if len(got) != 0 {
		t.Errorf("expected 0 queries for whitespace-only input, got %d: %v", len(got), got)
	}
}

func TestSplitQueries_OnlySemicolons(t *testing.T) {
	got := splitQueries(";\n;\n;")
	if len(got) != 0 {
		t.Errorf("expected 0 queries for semicolons-only input, got %d: %v", len(got), got)
	}
}

func TestSplitQueries_WindowsLineEndings(t *testing.T) {
	got := splitQueries("query1\r\n;\r\nquery2\r\n")
	if len(got) != 2 {
		t.Fatalf("expected 2 queries, got %d: %v", len(got), got)
	}
	if got[0] != "query1" {
		t.Errorf("query 0: got %q, want %q", got[0], "query1")
	}
	if got[1] != "query2" {
		t.Errorf("query 1: got %q, want %q", got[1], "query2")
	}
}

func TestResolve_EmptyScriptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.kql")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("writing empty script: %v", err)
	}
	result, err := Resolve("", path)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(result.Queries) != 0 {
		t.Errorf("expected 0 queries for empty script, got %d", len(result.Queries))
	}
}
