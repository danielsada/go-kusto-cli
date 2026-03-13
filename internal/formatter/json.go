package formatter

import (
	"encoding/json"
	"io"

	"github.com/danielsada/go-kusto-cli/internal/client"
)

// JSON formats a DataTable as a JSON array of objects.
type JSON struct{}

// Format writes a DataTable as a JSON array of objects.
func (JSON) Format(w io.Writer, table *client.DataTable) error {
	if len(table.Columns) == 0 {
		_, err := w.Write([]byte("[]\n"))
		return err
	}

	rows := make([]map[string]interface{}, 0, len(table.Rows))
	for _, row := range table.Rows {
		obj := make(map[string]interface{}, len(table.Columns))
		for i, col := range table.Columns {
			if i < len(row) {
				obj[ColumnName(col)] = row[i]
			}
		}
		rows = append(rows, obj)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
