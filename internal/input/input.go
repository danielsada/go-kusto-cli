package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Source represents where the query input came from.
type Source string

// Input source constants.
const (
	SourceFlag   Source = "flag"
	SourceScript Source = "script"
	SourceStdin  Source = "stdin"
)

// Result holds the resolved queries and their source.
type Result struct {
	Source  Source
	Queries []string
}

// Resolve determines the query input based on priority: flag > script > stdin.
// Returns a list of queries to execute sequentially.
func Resolve(executeFlag string, scriptFlag string) (*Result, error) {
	if executeFlag != "" {
		return &Result{
			Queries: []string{executeFlag},
			Source:  SourceFlag,
		}, nil
	}

	if scriptFlag != "" {
		queries, err := readScriptFile(scriptFlag)
		if err != nil {
			return nil, fmt.Errorf("reading script file: %w", err)
		}
		return &Result{
			Queries: queries,
			Source:  SourceScript,
		}, nil
	}

	query, err := readStdin()
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	if query == "" {
		return nil, fmt.Errorf("no query provided: use -e, -s, or pipe a query via stdin")
	}
	return &Result{
		Queries: []string{query},
		Source:  SourceStdin,
	}, nil
}

// readScriptFile reads a KQL script file and splits it into individual queries.
// Queries are delimited by semicolons on their own line or blank-line separators.
func readScriptFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return splitQueries(string(data)), nil
}

// readStdin reads all available input from stdin.
func readStdin() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if (info.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}

	reader := bufio.NewReader(os.Stdin)
	var sb strings.Builder
	_, err = io.Copy(&sb, reader)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(sb.String()), nil
}

// splitQueries splits a script into individual queries.
// Queries are separated by lines containing only ";" or by empty lines.
func splitQueries(script string) []string {
	var queries []string
	var current strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(script))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == ";" || trimmed == "" {
			q := strings.TrimSpace(current.String())
			if q != "" {
				queries = append(queries, q)
			}
			current.Reset()
			continue
		}

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)
	}

	q := strings.TrimSpace(current.String())
	if q != "" {
		queries = append(queries, q)
	}

	return queries
}
