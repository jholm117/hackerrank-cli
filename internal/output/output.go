package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type TableWriter struct {
	w       *tabwriter.Writer
	headers []string
	rows    [][]string
}

func NewTableWriter(w io.Writer) *TableWriter {
	return &TableWriter{
		w: tabwriter.NewWriter(w, 0, 4, 2, ' ', 0),
	}
}

func (t *TableWriter) SetHeader(headers []string) {
	t.headers = headers
}

func (t *TableWriter) Append(row []string) {
	t.rows = append(t.rows, row)
}

func (t *TableWriter) Render() {
	if len(t.headers) > 0 {
		fmt.Fprintln(t.w, strings.Join(t.headers, "\t"))
	}
	for _, row := range t.rows {
		fmt.Fprintln(t.w, strings.Join(row, "\t"))
	}
	t.w.Flush()
}

func WriteJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
