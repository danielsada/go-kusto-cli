package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/danielsada/go-kusto-cli/internal/auth"
	"github.com/danielsada/go-kusto-cli/internal/cli"
	"github.com/danielsada/go-kusto-cli/internal/client"
	"github.com/danielsada/go-kusto-cli/internal/input"
)

// Set via -ldflags at build time by GoReleaser.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	cluster := flag.String("c", "", "Kusto cluster URL (e.g. https://mycluster.region.kusto.windows.net)")
	database := flag.String("d", "", "Database name")
	execute := flag.String("e", "", "KQL query to execute")
	script := flag.String("s", "", "Path to KQL script file")
	format := flag.String("f", "table", "Output format: table, csv, json")
	output := flag.String("o", "", "Output file path (default: stdout)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-kusto-cli %s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: go-kusto-cli -c <cluster> -d <database> [-e <query> | -s <script>] [-f table|csv|json] [-o <file>]\n\n")
		fmt.Fprintf(os.Stderr, "A CLI for querying Azure Data Explorer (Kusto) via the REST API.\n\n")
		fmt.Fprintf(os.Stderr, "Query input (pick one, in priority order):\n")
		fmt.Fprintf(os.Stderr, "  -e  Inline KQL query\n")
		fmt.Fprintf(os.Stderr, "  -s  Path to a script file containing KQL queries (separated by ; or blank lines)\n")
		fmt.Fprintf(os.Stderr, "  stdin  Pipe a query via stdin\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := &cli.App{
		Cluster:  *cluster,
		Database: *database,
		Execute:  *execute,
		Script:   *script,
		Format:   *format,
		Output:   *output,
		Auth:     auth.GetToken,
		ClientNew: func(clusterURL, token string) cli.QueryFunc {
			c := client.New(clusterURL, token)
			return c.Query
		},
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		CreateFile:   func(name string) (io.WriteCloser, error) { return os.Create(name) },
		InputResolve: input.Resolve,
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
