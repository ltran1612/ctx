package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	Bold   = color.New(color.Bold)
	Green  = color.New(color.FgGreen)
	Yellow = color.New(color.FgYellow)
	Cyan   = color.New(color.FgCyan)
	Dim    = color.New(color.Faint)
	Red    = color.New(color.FgRed)
)

// Table prints rows in a tab-aligned table with a bold header row.
// headers and each row must have the same number of columns.
func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	Bold.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// Println prints a line to stdout.
func Println(a ...any) {
	fmt.Fprintln(os.Stdout, a...)
}

// Printf prints formatted text to stdout.
func Printf(format string, a ...any) {
	fmt.Fprintf(os.Stdout, format, a...)
}

// Errorf prints an error message to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
}

// Warnf prints a warning to stderr.
func Warnf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "warning: "+format+"\n", a...)
}

// Success prints a green success message.
func Success(format string, a ...any) {
	Green.Fprintf(os.Stdout, format+"\n", a...)
}

// Info prints a dim informational message.
func Info(format string, a ...any) {
	Dim.Fprintf(os.Stdout, format+"\n", a...)
}

// FprintMarkdown writes markdown to w with minimal terminal formatting:
// ## headings are bold+cyan, <!-- comments --> are dimmed.
func FprintMarkdown(w io.Writer, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "## "):
			Cyan.Fprintln(w, line)
		case strings.HasPrefix(trimmed, "# "):
			Bold.Fprintln(w, line)
		case strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->"):
			Dim.Fprintln(w, line)
		default:
			fmt.Fprintln(w, line)
		}
	}
}
