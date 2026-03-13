package client

import (
	"encoding/json"
	"fmt"
)

// FrameType identifies the kind of v2 response frame.
type FrameType string

// Frame types returned by the Kusto REST API v2.
const (
	FrameDataSetHeader     FrameType = "DataSetHeader"
	FrameDataTable         FrameType = "DataTable"
	FrameDataSetCompletion FrameType = "DataSetCompletion"
	FrameTableHeader       FrameType = "TableHeader"
	FrameTableFragment     FrameType = "TableFragment"
	FrameTableCompletion   FrameType = "TableCompletion"
	FrameTableProgress     FrameType = "TableProgress"
)

// Frame is a single frame in the v2 response.
type Frame struct {
	FrameType FrameType       `json:"FrameType"`
	Raw       json.RawMessage `json:"-"`
}

// DataSetHeader is the first frame in a v2 response.
type DataSetHeader struct {
	Version       string    `json:"Version"`
	FrameType     FrameType `json:"FrameType"`
	IsProgressive bool      `json:"IsProgressive"`
}

// Column describes a column in a DataTable.
type Column struct {
	ColumnName string `json:"ColumnName"`
	ColumnType string `json:"ColumnType"`
	DataType   string `json:"DataType"`
}

// DataTable contains tabular data in the v2 response.
type DataTable struct {
	FrameType FrameType       `json:"FrameType"`
	TableKind string          `json:"TableKind"`
	TableName string          `json:"TableName"`
	Columns   []Column        `json:"Columns"`
	Rows      [][]interface{} `json:"Rows"`
	TableID   int             `json:"TableId"`
}

// DataSetCompletion is the last frame in a v2 response.
type DataSetCompletion struct {
	FrameType    FrameType `json:"FrameType"`
	OneAPIErrors []struct {
		ErrorMessage struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	} `json:"OneApiErrors"`
	HasErrors bool `json:"HasErrors"`
	Canceled  bool `json:"Cancelled"` //nolint:misspell // Kusto API uses British spelling
}

// ParseFrames parses the v2 JSON response into typed frames.
func ParseFrames(data []byte) ([]Frame, error) {
	var rawFrames []json.RawMessage
	if err := json.Unmarshal(data, &rawFrames); err != nil {
		return nil, fmt.Errorf("unmarshaling frames array: %w", err)
	}

	frames := make([]Frame, len(rawFrames))
	for i, raw := range rawFrames {
		var f struct {
			FrameType FrameType `json:"FrameType"`
		}
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, fmt.Errorf("reading frame type at index %d: %w", i, err)
		}
		frames[i] = Frame{
			FrameType: f.FrameType,
			Raw:       raw,
		}
	}
	return frames, nil
}

// ExtractPrimaryTable finds the first PrimaryResult DataTable from frames.
func ExtractPrimaryTable(frames []Frame) (*DataTable, error) {
	for _, f := range frames {
		if f.FrameType != FrameDataTable {
			continue
		}
		var dt DataTable
		if err := json.Unmarshal(f.Raw, &dt); err != nil {
			return nil, fmt.Errorf("unmarshaling DataTable: %w", err)
		}
		if dt.TableKind == "PrimaryResult" {
			return &dt, nil
		}
	}
	return nil, fmt.Errorf("no PrimaryResult table found in response")
}

// CheckErrors inspects frames for errors in DataSetCompletion.
func CheckErrors(frames []Frame) error {
	for _, f := range frames {
		if f.FrameType != FrameDataSetCompletion {
			continue
		}
		var dc DataSetCompletion
		if err := json.Unmarshal(f.Raw, &dc); err != nil {
			return fmt.Errorf("unmarshaling DataSetCompletion: %w", err)
		}
		if dc.HasErrors {
			if len(dc.OneAPIErrors) > 0 {
				return fmt.Errorf("query error: %s", dc.OneAPIErrors[0].ErrorMessage.Message)
			}
			return fmt.Errorf("query completed with errors")
		}
		if dc.Canceled {
			return fmt.Errorf("query was canceled")
		}
	}
	return nil
}
