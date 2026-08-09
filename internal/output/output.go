package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func Table(w io.Writer, header []string, rows [][]string, maxWidth int) error {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len([]rune(cell)) > widths[i] {
				widths[i] = len([]rune(cell))
			}
		}
	}
	total := len(header) - 1
	for _, w := range widths {
		total += w + 2
	}
	if maxWidth > 0 {
		for total > maxWidth {
			biggest := 0
			for i := range widths {
				if widths[i] > widths[biggest] {
					biggest = i
				}
			}
			if widths[biggest] <= 3 {
				break
			}
			widths[biggest]--
			total--
		}
	}
	line := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, c := range cells {
			parts[i] = fmt.Sprintf("%-*s", widths[i], Truncate(c, widths[i]))
		}
		return strings.Join(parts, "  ")
	}
	if _, err := fmt.Fprintln(w, line(header)); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, line(row)); err != nil {
			return err
		}
	}
	return nil
}

func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func Truncate(s string, width int) string {
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 3 {
		return string(r[:width])
	}
	return string(r[:width-3]) + "..."
}
