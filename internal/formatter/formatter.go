package formatter

import (
	"fmt"
	"io"

	"github.com/danielsada/go-kusto-cli/internal/client"
)

// Formatter writes a DataTable to the given writer.
type Formatter interface {
	Format(w io.Writer, table *client.DataTable) error
}

// ColumnName returns the display name for a column, preferring ColumnName over DataType.
func ColumnName(c client.Column) string {
	if c.ColumnName != "" {
		return c.ColumnName
	}
	return c.DataType
}

// CellString converts a cell value to a string for display.
func CellString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}
