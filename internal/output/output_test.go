package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	w := NewTableWriter(&buf)
	w.SetHeader([]string{"ID", "NAME", "STATE"})
	w.Append([]string{"123", "My Test", "active"})
	w.Append([]string{"456", "Other Test", "draft"})
	w.Render()

	out := buf.String()
	if !strings.Contains(out, "123") {
		t.Errorf("table missing ID 123:\n%s", out)
	}
	if !strings.Contains(out, "My Test") {
		t.Errorf("table missing name:\n%s", out)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"id": "123"}
	if err := WriteJSON(&buf, data); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"id"`) {
		t.Errorf("json missing id field:\n%s", out)
	}
}
