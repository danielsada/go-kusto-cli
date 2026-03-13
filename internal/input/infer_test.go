package input

import "testing"

func TestInferParams_ClusterAndDatabase(t *testing.T) {
	queries := []string{
		`cluster('https://mycluster.westus.kusto.windows.net').database('MyDB').StormEvents | count`,
	}
	got := InferParams(queries)
	if got.Cluster != "https://mycluster.westus.kusto.windows.net" {
		t.Errorf("cluster: got %q", got.Cluster)
	}
	if got.Database != "MyDB" {
		t.Errorf("database: got %q", got.Database)
	}
}

func TestInferParams_BareClusterName(t *testing.T) {
	queries := []string{`cluster('mycluster').database('TestDB').Table`}
	got := InferParams(queries)
	if got.Cluster != "https://mycluster.kusto.windows.net" {
		t.Errorf("cluster: got %q, want https://mycluster.kusto.windows.net", got.Cluster)
	}
	if got.Database != "TestDB" {
		t.Errorf("database: got %q", got.Database)
	}
}

func TestInferParams_DoubleQuotes(t *testing.T) {
	queries := []string{`cluster("mycluster").database("MyDB").Table | take 10`}
	got := InferParams(queries)
	if got.Cluster != "https://mycluster.kusto.windows.net" {
		t.Errorf("cluster: got %q", got.Cluster)
	}
	if got.Database != "MyDB" {
		t.Errorf("database: got %q", got.Database)
	}
}

func TestInferParams_DatabaseOnly(t *testing.T) {
	queries := []string{`database('Logs').Events | where Level == "Error"`}
	got := InferParams(queries)
	if got.Cluster != "" {
		t.Errorf("expected empty cluster, got %q", got.Cluster)
	}
	if got.Database != "Logs" {
		t.Errorf("database: got %q", got.Database)
	}
}

func TestInferParams_ClusterOnly(t *testing.T) {
	queries := []string{`cluster('https://prod.eastus.kusto.windows.net').database('').Table`}
	got := InferParams(queries)
	if got.Cluster != "https://prod.eastus.kusto.windows.net" {
		t.Errorf("cluster: got %q", got.Cluster)
	}
	// database('') should not match since it's empty
}

func TestInferParams_NoMatch(t *testing.T) {
	queries := []string{`StormEvents | count`}
	got := InferParams(queries)
	if got.Cluster != "" || got.Database != "" {
		t.Errorf("expected empty params, got cluster=%q database=%q", got.Cluster, got.Database)
	}
}

func TestInferParams_MultipleQueries(t *testing.T) {
	queries := []string{
		`StormEvents | count`,
		`cluster('clusterA').database('DB1').Table | take 5`,
		`cluster('clusterB').database('DB2').Table | take 5`,
	}
	got := InferParams(queries)
	if got.Cluster != "https://clusterA.kusto.windows.net" {
		t.Errorf("expected first cluster match, got %q", got.Cluster)
	}
	if got.Database != "DB1" {
		t.Errorf("expected first database match, got %q", got.Database)
	}
}

func TestInferParams_WhitespaceInParens(t *testing.T) {
	queries := []string{`cluster( 'mycluster' ).database( "MyDB" ).Table`}
	got := InferParams(queries)
	if got.Cluster != "https://mycluster.kusto.windows.net" {
		t.Errorf("cluster: got %q", got.Cluster)
	}
	if got.Database != "MyDB" {
		t.Errorf("database: got %q", got.Database)
	}
}

func TestInferParams_EmptyQueries(t *testing.T) {
	got := InferParams(nil)
	if got.Cluster != "" || got.Database != "" {
		t.Errorf("expected empty params for nil input")
	}
}

func TestNormalizeClusterURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://mycluster.westus.kusto.windows.net", "https://mycluster.westus.kusto.windows.net"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"mycluster", "https://mycluster.kusto.windows.net"},
		{"prod-cluster", "https://prod-cluster.kusto.windows.net"},
	}
	for _, tc := range tests {
		got := normalizeClusterURL(tc.input)
		if got != tc.want {
			t.Errorf("normalizeClusterURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
