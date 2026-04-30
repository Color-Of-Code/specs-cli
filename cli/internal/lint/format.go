package lint

import (
	"bytes"
	"os"
	"strings"
	"unicode/utf8"
)

// FormatFile reads a markdown file and returns the formatted content.
// Formatting operations:
//   - Normalize line endings to LF
//   - Remove trailing whitespace from each line
//   - Collapse 3+ consecutive blank lines to 2
//   - Align table pipe columns
//   - Ensure file ends with exactly one newline
func FormatFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Format(data), nil
}

// Format applies markdown formatting rules to the input bytes.
func Format(data []byte) []byte {
	// Normalize CRLF/CR to LF.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))

	lines := strings.Split(string(data), "\n")

	// Trim trailing whitespace from each line.
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Collapse consecutive blank lines (keep at most one blank line).
	lines = collapseBlankLines(lines)

	// Align markdown tables.
	lines = alignTables(lines)

	// Join and ensure exactly one trailing newline.
	result := strings.Join(lines, "\n")
	result = strings.TrimRight(result, "\n") + "\n"

	return []byte(result)
}

// collapseBlankLines ensures no more than two consecutive blank lines exist
// (i.e. at most one blank line between content).
func collapseBlankLines(lines []string) []string {
	var out []string
	blankCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 1 {
				out = append(out, line)
			}
		} else {
			blankCount = 0
			out = append(out, line)
		}
	}
	return out
}

// isTableRow returns true if the line looks like a markdown table row (starts
// with |). It ignores lines inside fenced code blocks.
func isTableRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
}

// isSeparatorRow returns true if it's a table separator row like |---|---|
func isSeparatorRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	inner := strings.Trim(trimmed, "|")
	for _, cell := range strings.Split(inner, "|") {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}
		stripped := strings.Trim(cell, ":-")
		if stripped != "" {
			return false
		}
	}
	return true
}

// alignTables finds consecutive table rows and aligns columns.
func alignTables(lines []string) []string {
	var result []string
	fenced := false
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks.
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			result = append(result, line)
			i++
			continue
		}
		if fenced {
			result = append(result, line)
			i++
			continue
		}

		// Detect start of a table block.
		if isTableRow(line) {
			// Collect all consecutive table rows.
			tableStart := i
			for i < len(lines) && isTableRow(lines[i]) {
				i++
			}
			tableLines := lines[tableStart:i]
			result = append(result, formatTable(tableLines)...)
			continue
		}

		result = append(result, line)
		i++
	}
	return result
}

// formatTable aligns a block of table rows by padding cells.
func formatTable(rows []string) []string {
	if len(rows) == 0 {
		return rows
	}

	// Parse cells.
	table := make([][]string, len(rows))
	maxCols := 0
	for i, row := range rows {
		trimmed := strings.TrimSpace(row)
		// Strip leading and trailing pipes.
		inner := trimmed[1 : len(trimmed)-1]
		cells := strings.Split(inner, "|")
		for j, cell := range cells {
			cells[j] = strings.TrimSpace(cell)
		}
		table[i] = cells
		if len(cells) > maxCols {
			maxCols = len(cells)
		}
	}

	// Compute max display width per column.
	colWidths := make([]int, maxCols)
	for _, cells := range table {
		for j, cell := range cells {
			w := displayWidth(cell)
			if w > colWidths[j] {
				colWidths[j] = w
			}
		}
	}

	// Rebuild rows with aligned columns.
	out := make([]string, len(rows))
	for i, cells := range table {
		var buf strings.Builder
		buf.WriteString("|")
		for j := 0; j < maxCols; j++ {
			cell := ""
			if j < len(cells) {
				cell = cells[j]
			}
			buf.WriteString(" ")
			if isSeparatorRow(rows[i]) {
				buf.WriteString(padSeparator(cell, colWidths[j]))
			} else {
				buf.WriteString(padRight(cell, colWidths[j]))
			}
			buf.WriteString(" |")
		}
		out[i] = buf.String()
	}
	return out
}

// displayWidth returns the display width of a string (rune count, treating
// each rune as width 1 for simplicity in a terminal context).
func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}

// padRight pads s with spaces on the right to reach width w.
func padRight(s string, w int) string {
	need := w - displayWidth(s)
	if need <= 0 {
		return s
	}
	return s + strings.Repeat(" ", need)
}

// padSeparator handles separator cells (like ----, :---, :---:) and pads
// them with dashes to match the target width.
func padSeparator(cell string, w int) string {
	if cell == "" {
		return strings.Repeat("-", w)
	}
	// Detect alignment markers.
	hasLeft := strings.HasPrefix(cell, ":")
	hasRight := strings.HasSuffix(cell, ":")
	need := w
	if hasLeft {
		need--
	}
	if hasRight {
		need--
	}
	if need < 1 {
		need = 1
	}
	var buf strings.Builder
	if hasLeft {
		buf.WriteString(":")
	}
	buf.WriteString(strings.Repeat("-", need))
	if hasRight {
		buf.WriteString(":")
	}
	return buf.String()
}

// FormatFileInPlace formats a file and writes it back if content changed.
// Returns true if the file was modified.
func FormatFileInPlace(path string) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	formatted := Format(original)
	if bytes.Equal(original, formatted) {
		return false, nil
	}
	return true, os.WriteFile(path, formatted, 0o644)
}
