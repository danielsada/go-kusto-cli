package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/danielsada/go-kusto-cli/internal/client"
	"github.com/danielsada/go-kusto-cli/internal/formatter"
	"github.com/danielsada/go-kusto-cli/internal/input"
)

// QueryFunc executes a KQL query and returns v2 response frames.
type QueryFunc func(ctx context.Context, database, query string) ([]client.Frame, error)

// AuthFunc obtains an authentication token.
type AuthFunc func() (string, error)

// App holds the configuration and dependencies for a CLI run.
type App struct {
	Auth         AuthFunc
	ClientNew    func(clusterURL, token string) QueryFunc
	Stdout       io.Writer
	Stderr       io.Writer
	CreateFile   func(name string) (io.WriteCloser, error)
	InputResolve func(executeFlag, scriptFlag string) (*input.Result, error)
	Cluster      string
	Database     string
	Execute      string
	Script       string
	Format       string
	Output       string
}

// Run executes the CLI logic. Returns an error instead of calling os.Exit.
// The provided context controls cancellation of in-flight queries.
func (a *App) Run(ctx context.Context) error {
	resolved, err := a.InputResolve(a.Execute, a.Script)
	if err != nil {
		return fmt.Errorf("resolving input: %w", err)
	}

	// Infer cluster/database from KQL if not provided via flags.
	if a.Cluster == "" || a.Database == "" {
		inferred := input.InferParams(resolved.Queries)
		if a.Cluster == "" {
			a.Cluster = inferred.Cluster
		}
		if a.Database == "" {
			a.Database = inferred.Database
		}
	}

	if a.Cluster == "" || a.Database == "" {
		return fmt.Errorf("-c (cluster) and -d (database) are required (or use cluster()/database() in your query)")
	}

	token, err := a.Auth()
	if err != nil {
		return fmt.Errorf("authentication: %w", err)
	}

	fmtr, err := resolveFormatter(a.Format)
	if err != nil {
		return err
	}

	w := a.Stdout
	if a.Output != "" {
		f, createErr := a.CreateFile(a.Output)
		if createErr != nil {
			return fmt.Errorf("creating output file: %w", createErr)
		}
		defer func() { _ = f.Close() }()
		w = f
	}

	queryFn := a.ClientNew(a.Cluster, token)

	for i, query := range resolved.Queries {
		if len(resolved.Queries) > 1 {
			_, _ = fmt.Fprintf(a.Stderr, "--- Query %d/%d ---\n", i+1, len(resolved.Queries))
		}

		frames, queryErr := queryFn(ctx, a.Database, query)
		if queryErr != nil {
			return fmt.Errorf("query execution: %w", queryErr)
		}

		if checkErr := client.CheckErrors(frames); checkErr != nil {
			return checkErr
		}

		table, extractErr := client.ExtractPrimaryTable(frames)
		if extractErr != nil {
			return extractErr
		}

		if fmtErr := fmtr.Format(w, table); fmtErr != nil {
			return fmt.Errorf("writing output: %w", fmtErr)
		}

		if len(resolved.Queries) > 1 && i < len(resolved.Queries)-1 {
			if _, fpErr := fmt.Fprintln(w); fpErr != nil {
				return fmt.Errorf("writing separator: %w", fpErr)
			}
		}
	}
	return nil
}

func resolveFormatter(format string) (formatter.Formatter, error) {
	switch format {
	case "table":
		return formatter.Table{}, nil
	case "csv":
		return formatter.CSV{}, nil
	case "json":
		return formatter.JSON{}, nil
	default:
		return nil, fmt.Errorf("unknown format %q (use table, csv, or json)", format)
	}
}
