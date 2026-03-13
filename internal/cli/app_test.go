package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/danielsada/go-kusto-cli/internal/client"
	"github.com/danielsada/go-kusto-cli/internal/input"
)

var sampleFrames = []client.Frame{
	{FrameType: client.FrameDataSetHeader},
	{FrameType: client.FrameDataTable, Raw: []byte(`{
		"FrameType":"DataTable","TableId":0,"TableKind":"PrimaryResult","TableName":"Result",
		"Columns":[{"ColumnName":"Name","ColumnType":"string"},{"ColumnName":"Count","ColumnType":"long"}],
		"Rows":[["Alice",42]]
	}`)},
	{FrameType: client.FrameDataSetCompletion, Raw: []byte(`{
		"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":false
	}`)},
}

type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }

func mockApp(stdout, stderr *bytes.Buffer) *App {
	return &App{
		Cluster:  "https://test.kusto.windows.net",
		Database: "testdb",
		Execute:  "TestQuery | count",
		Format:   "table",
		Auth:     func() (string, error) { return "mock-token", nil },
		ClientNew: func(_, _ string) QueryFunc {
			return func(_ context.Context, _, _ string) ([]client.Frame, error) {
				return sampleFrames, nil
			}
		},
		Stdout:       stdout,
		Stderr:       stderr,
		CreateFile:   func(name string) (io.WriteCloser, error) { return nopCloser{stdout}, nil },
		InputResolve: input.Resolve,
	}
}

func TestApp_Run_Success(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "Alice") {
		t.Errorf("expected Alice in output, got: %s", stdout.String())
	}
}

func TestApp_Run_MissingCluster(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Cluster = ""
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for missing cluster")
	}
	if !strings.Contains(err.Error(), "-c (cluster)") {
		t.Errorf("expected cluster error, got: %v", err)
	}
}

func TestApp_Run_MissingDatabase(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Database = ""
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for missing database")
	}
	if !strings.Contains(err.Error(), "-d (database)") {
		t.Errorf("expected database error, got: %v", err)
	}
}

func TestApp_Run_InvalidFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Format = "xml"
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("expected format in error, got: %v", err)
	}
}

func TestApp_Run_AuthFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Auth = func() (string, error) { return "", fmt.Errorf("not logged in") }
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestApp_Run_QueryFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.ClientNew = func(_, _ string) QueryFunc {
		return func(_ context.Context, _, _ string) ([]client.Frame, error) {
			return nil, fmt.Errorf("connection refused")
		}
	}
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for query failure")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected connection error, got: %v", err)
	}
}

func TestApp_Run_CSVFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Format = "csv"
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "Alice") {
		t.Errorf("expected Alice in CSV output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Name,Count") || !strings.Contains(stdout.String(), "Name") {
		t.Errorf("expected CSV headers, got: %s", stdout.String())
	}
}

func TestApp_Run_JSONFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Format = "json"
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "Alice") {
		t.Errorf("expected Alice in JSON output, got: %s", stdout.String())
	}
}

func TestApp_Run_OutputFile(t *testing.T) {
	var fileBuf bytes.Buffer
	var stderr bytes.Buffer
	app := mockApp(&bytes.Buffer{}, &stderr)
	app.Output = "/tmp/test-output.csv"
	app.Format = "csv"
	app.CreateFile = func(_ string) (io.WriteCloser, error) {
		return nopCloser{&fileBuf}, nil
	}
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(fileBuf.String(), "Alice") {
		t.Errorf("expected output in file buffer, got: %s", fileBuf.String())
	}
}

func TestApp_Run_OutputFileError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Output = "/nonexistent/path/file.csv"
	app.CreateFile = func(_ string) (io.WriteCloser, error) {
		return nil, fmt.Errorf("permission denied")
	}
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for file creation failure")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission error, got: %v", err)
	}
}

func TestApp_Run_MultipleQueries(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Execute = ""
	app.InputResolve = func(_, _ string) (*input.Result, error) {
		return &input.Result{
			Source:  input.SourceScript,
			Queries: []string{"query1", "query2"},
		}, nil
	}
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(stderr.String(), "Query 1/2") {
		t.Errorf("expected query progress on stderr, got: %s", stderr.String())
	}
}

func TestApp_Run_InputResolveError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Execute = ""
	app.InputResolve = func(_, _ string) (*input.Result, error) {
		return nil, fmt.Errorf("no query provided")
	}
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for input resolution failure")
	}
	if !strings.Contains(err.Error(), "no query provided") {
		t.Errorf("expected input error, got: %v", err)
	}
}

func TestApp_Run_CheckErrorsFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.ClientNew = func(_, _ string) QueryFunc {
		return func(_ context.Context, _, _ string) ([]client.Frame, error) {
			frames := []client.Frame{
				{FrameType: client.FrameDataSetCompletion, Raw: []byte(`{
					"FrameType":"DataSetCompletion","HasErrors":true,"Cancelled":false,
					"OneApiErrors":[{"error":{"code":"BadRequest","message":"Syntax error"}}]
				}`)},
			}
			return frames, nil
		}
	}
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("expected error from CheckErrors")
	}
	if !strings.Contains(err.Error(), "Syntax error") {
		t.Errorf("expected syntax error, got: %v", err)
	}
}

func TestApp_Run_InferClusterAndDatabase(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Cluster = ""
	app.Database = ""
	app.Execute = `cluster('https://inferred.westus.kusto.windows.net').database('InferredDB').Events | count`

	// Capture which cluster/db the client was created with.
	var gotCluster string
	app.ClientNew = func(clusterURL, _ string) QueryFunc {
		gotCluster = clusterURL
		return func(_ context.Context, _, _ string) ([]client.Frame, error) {
			return sampleFrames, nil
		}
	}

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotCluster != "https://inferred.westus.kusto.windows.net" {
		t.Errorf("expected inferred cluster, got %q", gotCluster)
	}
}

func TestApp_Run_InferBareClusterName(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Cluster = ""
	app.Database = ""
	app.Execute = `cluster('myprod').database('Logs').Events | take 10`

	var gotCluster string
	app.ClientNew = func(clusterURL, _ string) QueryFunc {
		gotCluster = clusterURL
		return func(_ context.Context, _, _ string) ([]client.Frame, error) {
			return sampleFrames, nil
		}
	}

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotCluster != "https://myprod.kusto.windows.net" {
		t.Errorf("expected normalized cluster URL, got %q", gotCluster)
	}
}

func TestApp_Run_FlagsOverrideInference(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Cluster = "https://explicit.kusto.windows.net"
	app.Database = "ExplicitDB"
	app.Execute = `cluster('inferred').database('InferredDB').Events | count`

	var gotCluster, gotDB string
	app.ClientNew = func(clusterURL, _ string) QueryFunc {
		gotCluster = clusterURL
		return func(_ context.Context, database, _ string) ([]client.Frame, error) {
			gotDB = database
			return sampleFrames, nil
		}
	}

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotCluster != "https://explicit.kusto.windows.net" {
		t.Errorf("expected explicit cluster to win, got %q", gotCluster)
	}
	if gotDB != "ExplicitDB" {
		t.Errorf("expected explicit database to win, got %q", gotDB)
	}
}

func TestApp_Run_InferDatabaseOnly(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := mockApp(&stdout, &stderr)
	app.Cluster = "https://mycluster.kusto.windows.net"
	app.Database = ""
	app.Execute = `database('InferredDB').Events | count`

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
