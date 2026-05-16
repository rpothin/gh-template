package ui

import (
	"fmt"
	"os"
)

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// Success prints a green checkmark message to stderr.
func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, colorGreen+"  ✓ "+colorReset+format+"\n", args...)
}

// Warning prints a yellow warning message to stderr.
func Warning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, colorYellow+"  ⚠ "+colorReset+format+"\n", args...)
}

// Error prints a red error message to stderr.
func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, colorRed+"  ✗ "+colorReset+format+"\n", args...)
}

// Info prints a cyan informational message to stderr.
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, colorCyan+"  ℹ "+colorReset+format+"\n", args...)
}

// Header prints a bold section header to stderr.
func Header(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n"+colorBold+format+colorReset+"\n", args...)
}

// AuditMatch prints a green ✓ for a field that matches the config.
func AuditMatch(field string, value interface{}) {
	fmt.Fprintf(os.Stderr, colorGreen+"  ✓ "+colorReset+"%s: %v\n", field, value)
}

// AuditDrift prints a red ✗ for a field that diverges from the config.
func AuditDrift(field string, want, got interface{}) {
	fmt.Fprintf(os.Stderr, colorRed+"  ✗ "+colorReset+"%s: want %v, got %v\n", field, want, got)
}

// AuditExtra prints a yellow ⚠ for values present in live repo but absent in config.
func AuditExtra(field string, value interface{}) {
	fmt.Fprintf(os.Stderr, colorYellow+"  ⚠ "+colorReset+"%s: %v (not in config)\n", field, value)
}

// SummaryLine prints a summary line to stderr.
func SummaryLine(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n"+format+"\n", args...)
}
