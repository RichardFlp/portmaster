package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RichardFlp/portmaster/internal/output"
	"github.com/RichardFlp/portmaster/internal/ports"
)

type frame struct {
	width     int
	height    int
	title     string
	listeners []ports.Listener
	selected  int
	detail    string
	status    string
	keys      string
	changes   []string
}

func render(f frame) string {
	var b strings.Builder
	b.WriteString("\x1b[H")
	if f.width < 30 || f.height < 8 {
		msg := "Terminal too small (minimum 30x8)."
		pad := (f.width - len(msg)) / 2
		if pad < 0 {
			pad = 0
		}
		b.WriteString("\x1b[J")
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(msg)
		return b.String()
	}
	if f.detail != "" {
		return renderDetail(&b, f)
	}
	return renderTable(&b, f)
}

func borderLine(width int) string {
	return "+" + strings.Repeat("-", width-2) + "+"
}

func renderTable(b *strings.Builder, f frame) string {
	headers := []string{"PORT", "PROTOCOL", "PID", "PROCESS", "ADDRESS"}
	var rows [][]string
	for _, l := range f.listeners {
		pid := strconv.Itoa(l.PID)
		if l.PID == 0 {
			pid = "-"
		}
		proc := l.Process
		if proc == "" {
			proc = "-"
		}
		rows = append(rows, []string{
			strconv.Itoa(l.Port),
			strings.ToUpper(l.Protocol),
			pid,
			proc,
			l.Address,
		})
	}
	widths := fitColumns(headers, rows, f.width)
	dataRows := f.height - 8
	if dataRows < 0 {
		dataRows = 0
	}
	if dataRows > len(rows) {
		dataRows = len(rows)
	}
	start := 0
	if f.selected >= dataRows {
		start = f.selected - dataRows + 1
	}
	b.WriteString(borderLine(f.width))
	b.WriteString("\n")
	b.WriteString(titleLine(f.width, f.title, len(f.listeners)))
	b.WriteString("\n")
	b.WriteString(tableBorder(widths))
	b.WriteString("\n")
	b.WriteString(tableRow(widths, headers))
	b.WriteString("\n")
	b.WriteString(tableBorder(widths))
	b.WriteString("\n")
	for i := start; i < start+dataRows; i++ {
		line := tableRow(widths, rows[i])
		if i == f.selected {
			line = "\x1b[7m" + line + "\x1b[0m"
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(tableBorder(widths))
	b.WriteString("\n")
	status := f.status
	if status == "" && len(f.changes) > 0 {
		status = strings.Join(f.changes, "  ")
	}
	b.WriteString(fullLine(f.width, status))
	b.WriteString("\n")
	b.WriteString(fullLine(f.width, f.keys))
	return b.String()
}

func renderDetail(b *strings.Builder, f frame) string {
	lines := strings.Split(f.detail, "\n")
	content := f.height - 6
	if len(lines) > content {
		lines = lines[:content]
	}
	b.WriteString(borderLine(f.width))
	b.WriteString("\n")
	b.WriteString(titleLine(f.width, f.title, len(f.listeners)))
	b.WriteString("\n")
	b.WriteString(borderLine(f.width))
	b.WriteString("\n")
	for _, line := range lines {
		b.WriteString(fullLine(f.width, line))
		b.WriteString("\n")
	}
	for i := len(lines); i < content; i++ {
		b.WriteString(fullLine(f.width, ""))
		b.WriteString("\n")
	}
	b.WriteString(borderLine(f.width))
	b.WriteString("\n")
	b.WriteString(fullLine(f.width, f.keys))
	return b.String()
}

func titleLine(width int, title string, count int) string {
	inner := width - 2
	left := " " + title
	right := fmt.Sprintf("%d ports ", count)
	padding := inner - len(left) - len(right)
	if padding < 1 {
		padding = 1
	}
	return "|" + left + strings.Repeat(" ", padding) + right + "|"
}

func fullLine(width int, text string) string {
	inner := width - 2
	max := inner - 2
	runes := []rune(text)
	if len(runes) > max {
		runes = []rune(output.Truncate(text, max))
	}
	padding := inner - 2 - len(runes)
	if padding < 0 {
		padding = 0
	}
	return "| " + string(runes) + strings.Repeat(" ", padding) + " |"
}

func tableBorder(widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("-", w+2)
	}
	return "+" + strings.Join(parts, "+") + "+"
}

func tableRow(widths []int, cells []string) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = fmt.Sprintf(" %-*s ", widths[i], output.Truncate(c, widths[i]))
	}
	return "|" + strings.Join(parts, "|") + "|"
}

func fitColumns(headers []string, rows [][]string, width int) []int {
	n := len(headers)
	widths := make([]int, n)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, c := range row {
			if i < n && len([]rune(c)) > widths[i] {
				widths[i] = len([]rune(c))
			}
		}
	}
	available := width - 3*n - 1
	total := 0
	for _, w := range widths {
		total += w
	}
	for total > available {
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
	for total < available {
		widths[n-1]++
		total++
	}
	return widths
}
