package lint

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Violation represents a single style issue found in a file.
type Violation struct {
	File    string
	Line    int
	Rule    string
	Message string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: [%s] %s", v.File, v.Line, v.Rule, v.Message)
}

// CheckFileStyle runs all enabled style rules on a single markdown file and
// returns any violations found.
func CheckFileStyle(path string, cfg *StyleRules) []Violation {
	data, err := os.ReadFile(path)
	if err != nil {
		return []Violation{{File: path, Line: 1, Rule: "io", Message: err.Error()}}
	}

	var violations []Violation

	// Line-based checks (these don't need a parsed AST).
	violations = append(violations, checkLineRules(path, data, cfg)...)

	// AST-based checks.
	violations = append(violations, checkASTRules(path, data, cfg)...)

	return violations
}

// checkLineRules performs line-by-line checks that are simpler than full AST parsing.
func checkLineRules(path string, data []byte, cfg *StyleRules) []Violation {
	var violations []Violation
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0
	prevBlank := false
	fenced := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Track fenced code blocks.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			if fenced {
				prevBlank = false
			}
			continue
		}

		if fenced {
			prevBlank = false
			continue
		}

		// No trailing whitespace.
		if cfg.NoTrailingWhitespace && line != strings.TrimRight(line, " \t") {
			violations = append(violations, Violation{
				File: path, Line: lineNum, Rule: "no_trailing_whitespace",
				Message: "trailing whitespace",
			})
		}

		// No consecutive blank lines.
		isBlank := trimmed == ""
		if cfg.NoConsecutiveBlankLines && isBlank && prevBlank {
			violations = append(violations, Violation{
				File: path, Line: lineNum, Rule: "no_consecutive_blank_lines",
				Message: "consecutive blank lines",
			})
		}
		prevBlank = isBlank

		// Line length.
		limit := cfg.LineLengthLimit()
		if limit > 0 && len(line) > limit {
			// Skip lines inside tables or with long URLs.
			if !strings.HasPrefix(trimmed, "|") && !strings.Contains(line, "](http") {
				violations = append(violations, Violation{
					File: path, Line: lineNum, Rule: "line_length",
					Message: fmt.Sprintf("line length %d exceeds %d", len(line), limit),
				})
			}
		}

		// List marker consistency.
		if cfg.ListMarker != "" {
			if cfg.ListMarker == "dash" && (strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ")) {
				violations = append(violations, Violation{
					File: path, Line: lineNum, Rule: "list_marker",
					Message: "expected dash (-) list marker",
				})
			} else if cfg.ListMarker == "asterisk" && (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "+ ")) {
				violations = append(violations, Violation{
					File: path, Line: lineNum, Rule: "list_marker",
					Message: "expected asterisk (*) list marker",
				})
			}
		}
	}

	return violations
}

// checkASTRules uses goldmark to parse the document and check AST-level rules.
func checkASTRules(path string, data []byte, cfg *StyleRules) []Violation {
	var violations []Violation
	lines := bytes.Split(data, []byte("\n"))

	md := goldmark.New()
	reader := text.NewReader(data)
	doc := md.Parser().Parse(reader)

	firstHeading := true

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			line := lineOf(node, lines)

			// Heading style.
			if cfg.HeadingStyle == "atx" && line > 0 && line <= len(lines) {
				raw := string(lines[line-1])
				if !strings.HasPrefix(strings.TrimSpace(raw), "#") {
					violations = append(violations, Violation{
						File: path, Line: line, Rule: "heading_style",
						Message: "expected ATX-style heading (# ...)",
					})
				}
			}

			// First heading must be h1.
			if cfg.FirstHeadingH1 && firstHeading && node.Level != 1 {
				violations = append(violations, Violation{
					File: path, Line: line, Rule: "first_heading_h1",
					Message: fmt.Sprintf("first heading should be h1, got h%d", node.Level),
				})
			}
			firstHeading = false

			// Blank line before heading.
			if cfg.BlankLinesAroundHeadings && line > 1 {
				prev := string(lines[line-2])
				if strings.TrimSpace(prev) != "" {
					violations = append(violations, Violation{
						File: path, Line: line, Rule: "blank_lines_around_headings",
						Message: "expected blank line before heading",
					})
				}
			}

		case *ast.FencedCodeBlock:
			line := lineOf(node, lines)

			// Fenced code language.
			if cfg.FencedCodeLanguage {
				lang := node.Language(data)
				if len(lang) == 0 {
					violations = append(violations, Violation{
						File: path, Line: line, Rule: "fenced_code_language",
						Message: "fenced code block should have a language specified",
					})
				}
			}

			// Blank line before fence.
			if cfg.BlankLinesAroundFences && line > 1 {
				prev := string(lines[line-2])
				if strings.TrimSpace(prev) != "" {
					violations = append(violations, Violation{
						File: path, Line: line, Rule: "blank_lines_around_fences",
						Message: "expected blank line before fenced code block",
					})
				}
			}
		}

		return ast.WalkContinue, nil
	})

	// Inline HTML check (if disallowed).
	if !cfg.InlineHTML {
		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			switch node := n.(type) {
			case *ast.RawHTML:
				line := lineOf(node, lines)
				violations = append(violations, Violation{
					File: path, Line: line, Rule: "inline_html",
					Message: "inline HTML is not allowed",
				})
			case *ast.HTMLBlock:
				line := lineOf(node, lines)
				violations = append(violations, Violation{
					File: path, Line: line, Rule: "inline_html",
					Message: "HTML block is not allowed",
				})
			}
			return ast.WalkContinue, nil
		})
	}

	return violations
}

// lineOf extracts the 1-based line number from an AST node.
func lineOf(node ast.Node, lines [][]byte) int {
	if node.Lines().Len() > 0 {
		seg := node.Lines().At(0)
		return countLines(lines, seg.Start)
	}
	if node.HasChildren() {
		return lineOf(node.FirstChild(), lines)
	}
	return 1
}

// countLines returns the 1-based line number for a byte offset.
func countLines(lines [][]byte, offset int) int {
	pos := 0
	for i, line := range lines {
		// +1 accounts for the \n stripped by bytes.Split
		end := pos + len(line) + 1
		if offset < end {
			return i + 1
		}
		pos = end
	}
	return len(lines)
}
