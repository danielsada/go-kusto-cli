package formatter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/danielsada/go-kusto-cli/internal/client"
)

var testTable = &client.DataTable{
	TableKind: "PrimaryResult",
	Columns: []client.Column{
		{ColumnName: "Name", ColumnType: "string"},
		{ColumnName: "Count", ColumnType: "long"},
	},
	Rows: [][]interface{}{
		{"Alice", float64(42)},
		{"Bob", float64(7)},
	},
}

func TestTableFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := (Table{}).Format(&buf, testTable); err != nil {
		t.Fatalf("Table.Format: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Count") {
		t.Error("table output missing headers")
	}
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "42") {
		t.Error("table output missing data")
	}
	if !strings.Contains(out, "---") {
		t.Error("table output missing separator")
	}
}

func TestCSVFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := (CSV{}).Format(&buf, testTable); err != nil {
		t.Fatalf("CSV.Format: %v", err)
	}
	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parsing CSV output: %v", err)
	}
	if len(records) != 3 { // header + 2 data rows
		t.Fatalf("expected 3 CSV records, got %d", len(records))
	}
	if records[0][0] != "Name" || records[0][1] != "Count" {
		t.Errorf("CSV headers: got %v", records[0])
	}
	if records[1][0] != "Alice" || records[1][1] != "42" {
		t.Errorf("CSV row 1: got %v", records[1])
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := (JSON{}).Format(&buf, testTable); err != nil {
		t.Fatalf("JSON.Format: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("parsing JSON output: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 JSON objects, got %d", len(result))
	}
	if result[0]["Name"] != "Alice" {
		t.Errorf("JSON row 0 Name: got %v", result[0]["Name"])
	}
	if result[0]["Count"] != float64(42) {
		t.Errorf("JSON row 0 Count: got %v", result[0]["Count"])
	}
}

func TestEmptyTable(t *testing.T) {
	empty := &client.DataTable{Columns: []client.Column{}}
	var buf bytes.Buffer
	for _, f := range []Formatter{Table{}, CSV{}, JSON{}} {
		buf.Reset()
		if err := f.Format(&buf, empty); err != nil {
			t.Errorf("Format on empty table: %v", err)
		}
	}
}

func TestCellString(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{false, "false"},
	}
	for _, tc := range tests {
		got := CellString(tc.input)
		if got != tc.expected {
			t.Errorf("CellString(%v): got %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestColumnName_Fallback(t *testing.T) {
	col := client.Column{ColumnName: "", DataType: "Int64"}
	got := ColumnName(col)
	if got != "Int64" {
		t.Errorf("expected DataType fallback, got %q", got)
	}
}

func TestColumnName_PreferColumnName(t *testing.T) {
	col := client.Column{ColumnName: "MyCol", DataType: "string"}
	got := ColumnName(col)
	if got != "MyCol" {
		t.Errorf("expected ColumnName, got %q", got)
	}
}

func TestCSV_SpecialCharacters(t *testing.T) {
	tbl := &client.DataTable{
		Columns: []client.Column{{ColumnName: "Value", ColumnType: "string"}},
		Rows: [][]interface{}{
			{"hello, world"},
			{"has \"quotes\""},
			{"line\nbreak"},
		},
	}
	var buf bytes.Buffer
	if err := (CSV{}).Format(&buf, tbl); err != nil {
		t.Fatalf("CSV.Format: %v", err)
	}
	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parsing CSV: %v", err)
	}
	if len(records) != 4 { // header + 3 rows
		t.Fatalf("expected 4 records, got %d", len(records))
	}
	if records[1][0] != "hello, world" {
		t.Errorf("expected escaped comma, got %q", records[1][0])
	}
}

func TestTable_PartialRows(t *testing.T) {
	tbl := &client.DataTable{
		Columns: []client.Column{
			{ColumnName: "A", ColumnType: "string"},
			{ColumnName: "B", ColumnType: "string"},
			{ColumnName: "C", ColumnType: "string"},
		},
		Rows: [][]interface{}{
			{"only-one"},
		},
	}
	var buf bytes.Buffer
	if err := (Table{}).Format(&buf, tbl); err != nil {
		t.Fatalf("Table.Format: %v", err)
	}
	if !strings.Contains(buf.String(), "only-one") {
		t.Error("expected partial row data in output")
	}
}

func TestJSON_NilValues(t *testing.T) {
	tbl := &client.DataTable{
		Columns: []client.Column{
			{ColumnName: "Name", ColumnType: "string"},
			{ColumnName: "Value", ColumnType: "long"},
		},
		Rows: [][]interface{}{
			{"test", nil},
		},
	}
	var buf bytes.Buffer
	if err := (JSON{}).Format(&buf, tbl); err != nil {
		t.Fatalf("JSON.Format: %v", err)
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("parsing JSON: %v", err)
	}
	if result[0]["Value"] != nil {
		t.Errorf("expected nil value, got %v", result[0]["Value"])
	}
}

func TestTable_SingleColumn(t *testing.T) {
	tbl := &client.DataTable{
		Columns: []client.Column{{ColumnName: "X", ColumnType: "long"}},
		Rows:    [][]interface{}{{float64(1)}, {float64(2)}},
	}
	var buf bytes.Buffer
	if err := (Table{}).Format(&buf, tbl); err != nil {
		t.Fatalf("Table.Format: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 { // header + separator + 2 rows
		t.Errorf("expected 4 lines, got %d: %v", len(lines), lines)
	}
}
