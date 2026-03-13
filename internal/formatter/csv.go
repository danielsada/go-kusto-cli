package formatter

import (
	"encoding/csv"
	"io"

	"github.com/danielsada/go-kusto-cli/internal/client"
)

// CSV formats a DataTable as CSV.
type CSV struct{}

// Format writes a DataTable as CSV.
func (CSV) Format(w io.Writer, table *client.DataTable) error {
	if len(table.Columns) == 0 {
		return nil
	}

	cw := csv.NewWriter(w)

	// Header row
	headers := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		headers[i] = ColumnName(col)
	}
	if err := cw.Write(headers); err != nil {
		return err
	}

	// Data rows
	for _, row := range table.Rows {
		record := make([]string, len(table.Columns))
		for i := range table.Columns {
			if i < len(row) {
				record[i] = CellString(row[i])
			}
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
