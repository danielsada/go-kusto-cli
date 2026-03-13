package client

import (
	"encoding/json"
	"strings"
	"testing"
)

var sampleResponse = `[
  {"FrameType":"DataSetHeader","IsProgressive":false,"Version":"v2.0"},
  {"FrameType":"DataTable","TableId":0,"TableKind":"PrimaryResult","TableName":"PrimaryResult",
   "Columns":[{"ColumnName":"Name","ColumnType":"string"},{"ColumnName":"Count","ColumnType":"long"}],
   "Rows":[["Alice",42],["Bob",7]]},
  {"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":false}
]`

func TestParseFrames(t *testing.T) {
	frames, err := ParseFrames([]byte(sampleResponse))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	if len(frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(frames))
	}
	if frames[0].FrameType != FrameDataSetHeader {
		t.Errorf("frame 0: expected DataSetHeader, got %s", frames[0].FrameType)
	}
	if frames[1].FrameType != FrameDataTable {
		t.Errorf("frame 1: expected DataTable, got %s", frames[1].FrameType)
	}
	if frames[2].FrameType != FrameDataSetCompletion {
		t.Errorf("frame 2: expected DataSetCompletion, got %s", frames[2].FrameType)
	}
}

func TestExtractPrimaryTable(t *testing.T) {
	frames, err := ParseFrames([]byte(sampleResponse))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}

	dt, err := ExtractPrimaryTable(frames)
	if err != nil {
		t.Fatalf("ExtractPrimaryTable: %v", err)
	}
	if dt.TableKind != "PrimaryResult" {
		t.Errorf("expected PrimaryResult, got %s", dt.TableKind)
	}
	if len(dt.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(dt.Columns))
	}
	if dt.Columns[0].ColumnName != "Name" {
		t.Errorf("column 0: expected Name, got %s", dt.Columns[0].ColumnName)
	}
	if len(dt.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(dt.Rows))
	}
	if dt.Rows[0][0].(string) != "Alice" {
		t.Errorf("row 0 col 0: expected Alice, got %v", dt.Rows[0][0])
	}
}

func TestCheckErrors_NoErrors(t *testing.T) {
	frames, err := ParseFrames([]byte(sampleResponse))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	if err := CheckErrors(frames); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckErrors_WithErrors(t *testing.T) {
	errResp := `[
	  {"FrameType":"DataSetHeader","IsProgressive":false,"Version":"v2.0"},
	  {"FrameType":"DataSetCompletion","HasErrors":true,"Cancelled":false,
	   "OneApiErrors":[{"error":{"code":"BadRequest","message":"Syntax error"}}]}
	]`
	frames, err := ParseFrames([]byte(errResp))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	if err := CheckErrors(frames); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExtractPrimaryTable_NoPrimary(t *testing.T) {
	noData := `[
	  {"FrameType":"DataSetHeader","IsProgressive":false,"Version":"v2.0"},
	  {"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":false}
	]`
	frames, err := ParseFrames([]byte(noData))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	_, err = ExtractPrimaryTable(frames)
	if err == nil {
		t.Error("expected error for missing PrimaryResult, got nil")
	}
}

func TestParseFrames_InvalidJSON(t *testing.T) {
	_, err := ParseFrames([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDataSetCompletion_Unmarshal(t *testing.T) {
	raw := `{"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":true}`
	var dc DataSetCompletion
	if err := json.Unmarshal([]byte(raw), &dc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !dc.Canceled {
		t.Error("expected Canceled to be true")
	}
}

func TestCheckErrors_Canceled(t *testing.T) {
	cancelResp := `[
	  {"FrameType":"DataSetCompletion","HasErrors":false,"Cancelled":true}
	]`
	frames, err := ParseFrames([]byte(cancelResp))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	err = CheckErrors(frames)
	if err == nil {
		t.Error("expected error for canceled query, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected 'canceled' in error, got: %v", err)
	}
}

func TestCheckErrors_HasErrorsEmptyArray(t *testing.T) {
	errResp := `[
	  {"FrameType":"DataSetCompletion","HasErrors":true,"Cancelled":false,"OneApiErrors":[]}
	]`
	frames, err := ParseFrames([]byte(errResp))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	err = CheckErrors(frames)
	if err == nil {
		t.Error("expected error for HasErrors=true with empty array")
	}
}

func TestExtractPrimaryTable_MultipleTables(t *testing.T) {
	multiResp := `[
	  {"FrameType":"DataTable","TableId":0,"TableKind":"QueryStatus","TableName":"Status",
	   "Columns":[{"ColumnName":"Status","ColumnType":"string"}],"Rows":[["Done"]]},
	  {"FrameType":"DataTable","TableId":1,"TableKind":"PrimaryResult","TableName":"Result",
	   "Columns":[{"ColumnName":"Val","ColumnType":"long"}],"Rows":[[99]]}
	]`
	frames, err := ParseFrames([]byte(multiResp))
	if err != nil {
		t.Fatalf("ParseFrames: %v", err)
	}
	dt, err := ExtractPrimaryTable(frames)
	if err != nil {
		t.Fatalf("ExtractPrimaryTable: %v", err)
	}
	if dt.TableKind != "PrimaryResult" {
		t.Errorf("expected PrimaryResult, got %s", dt.TableKind)
	}
	if dt.Rows[0][0].(float64) != 99 {
		t.Errorf("expected 99, got %v", dt.Rows[0][0])
	}
}

func TestParseFrames_UnknownFrameType(t *testing.T) {
	resp := `[{"FrameType":"SomeNewFrameType","data":"hello"}]`
	frames, err := ParseFrames([]byte(resp))
	if err != nil {
		t.Fatalf("ParseFrames should not error on unknown frame type: %v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}
	if frames[0].FrameType != "SomeNewFrameType" {
		t.Errorf("expected SomeNewFrameType, got %s", frames[0].FrameType)
	}
}
