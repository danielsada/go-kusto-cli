package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/danielsada/go-kusto-cli/internal/client"
)

// Table formats a DataTable as an ASCII table.
type Table struct{}

// Format writes a DataTable as an ASCII table.
func (Table) Format(w io.Writer, table *client.DataTable) error {
	if len(table.Columns) == 0 {
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	// Header
	headers := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		headers[i] = ColumnName(col)
	}
	if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
		return err
	}

	// Separator
	seps := make([]string, len(table.Columns))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", len(h))
	}
	if _, err := fmt.Fprintln(tw, strings.Join(seps, "\t")); err != nil {
		return err
	}

	// Rows
	for _, row := range table.Rows {
		cells := make([]string, len(table.Columns))
		for i := range table.Columns {
			if i < len(row) {
				cells[i] = CellString(row[i])
			}
		}
		if _, err := fmt.Fprintln(tw, strings.Join(cells, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}
