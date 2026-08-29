package dutree

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteText renders the tree as indented, human-readable lines with sizes
// rounded to the nearest sensible unit, largest children first.
func WriteText(w io.Writer, e *Entry) error {
	return writeText(w, e, 0)
}

func writeText(w io.Writer, e *Entry, depth int) error {
	if e == nil {
		return nil
	}
	indent := strings.Repeat("  ", depth)
	if _, err := fmt.Fprintf(w, "%s%-8s %s\n", indent, HumanSize(e.Size), e.Name); err != nil {
		return err
	}
	for _, c := range e.Children {
		if err := writeText(w, c, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON renders the tree as a single JSON document. It carries the same
// data as WriteText, just structured for another program to consume.
func WriteJSON(w io.Writer, e *Entry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(e)
}

// HumanSize formats a byte count the way `du -h` does: the smallest unit
// that keeps the number under 1024, one decimal place once we're past
// plain bytes.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := "KMGTPE"
	return fmt.Sprintf("%.1f%ciB", float64(bytes)/float64(div), units[exp])
}
