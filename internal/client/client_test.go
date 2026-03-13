package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew_TrailingSlash(t *testing.T) {
	c := New("https://cluster.kusto.windows.net/", "token")
	if c.ClusterURL != "https://cluster.kusto.windows.net" {
		t.Errorf("expected trailing slash removed, got %q", c.ClusterURL)
	}
}

func TestNew_NoTrailingSlash(t *testing.T) {
	c := New("https://cluster.kusto.windows.net", "token")
	if c.ClusterURL != "https://cluster.kusto.windows.net" {
		t.Errorf("expected URL unchanged, got %q", c.ClusterURL)
	}
}

func TestNew_TokenStored(t *testing.T) {
	// Token is now unexported; verify indirectly via a successful query
	// that includes the bearer token in the Authorization header.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer my-token" {
			t.Errorf("expected Bearer my-token, got %q", got)
		}
		_, _ = w.Write([]byte(`[{"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":false}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "my-token")
	_, _ = c.Query(context.Background(), "db", "query")
}

func TestNew_HTTPClientSet(t *testing.T) {
	c := New("https://cluster.kusto.windows.net", "token")
	if c.HTTPClient == nil {
		t.Fatal("expected HTTPClient to be set")
	}
	if c.HTTPClient.Timeout == 0 {
		t.Error("expected HTTPClient timeout to be set")
	}
}

func TestQuery_Success(t *testing.T) {
	resp := `[
		{"FrameType":"DataSetHeader","IsProgressive":false,"Version":"v2.0"},
		{"FrameType":"DataTable","TableId":0,"TableKind":"PrimaryResult","TableName":"Result",
		 "Columns":[{"ColumnName":"Count","ColumnType":"long"}],
		 "Rows":[[42]]},
		{"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":false}
	]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v2/rest/query") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("unexpected Authorization header: %q", got)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Errorf("unexpected Content-Type: %q", got)
		}

		body, _ := io.ReadAll(r.Body)
		var req queryRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("invalid request body: %v", err)
		}
		if req.DB != "testdb" {
			t.Errorf("expected db=testdb, got %q", req.DB)
		}
		if req.CSL != "StormEvents | count" {
			t.Errorf("expected csl=StormEvents | count, got %q", req.CSL)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-token")
	frames, err := c.Query(context.Background(), "testdb", "StormEvents | count")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(frames))
	}
}

func TestQuery_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "bad-token")
	_, err := c.Query(context.Background(), "testdb", "query")
	if err == nil {
		t.Fatal("expected error for HTTP 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain status code, got: %v", err)
	}
}

func TestQuery_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c := New(srv.URL, "token")
	_, err := c.Query(context.Background(), "testdb", "query")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestQuery_ServerError500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`Internal Server Error`))
	}))
	defer srv.Close()

	c := New(srv.URL, "token")
	_, err := c.Query(context.Background(), "testdb", "query")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to contain 500, got: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		maxLen int
	}{
		{"short", "short", 10},
		{"exact", "exact", 5},
		{"this is longer", "this is...", 7},
		{"", "", 5},
	}
	for _, tc := range tests {
		got := truncate(tc.input, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
		}
	}
}
