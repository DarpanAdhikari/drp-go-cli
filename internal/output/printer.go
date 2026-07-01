// Package output provides consistent, symbol-prefixed CLI output used by
// every command instead of ad-hoc fmt calls. Colour is emitted only when
// stdout is a TTY; the --no-color flag suppresses it globally.
package output

import (
	"fmt"
	"os"
)

// colour codes
const (
	reset  = "\033[0m"
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

// noColor is set to true when --no-color is passed or NO_COLOR env is set.
// Commands call SetNoColor(true) after parsing flags.
var noColor bool

// SetNoColor disables ANSI colour output globally.
func SetNoColor(v bool) { noColor = v }

// isTTY reports whether stdout looks like an interactive terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func colorise(code, s string) string {
	if noColor || !isTTY() {
		return s
	}
	return code + s + reset
}

// Success prints a green ✓-prefixed success line to stdout.
func Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorise(green, "✓ "+msg))
}

// Fail prints a red ✗-prefixed failure line to stderr.
func Fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorise(red, "✗ "+msg))
}

// Info prints a cyan-prefixed informational line to stdout.
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorise(cyan, "  "+msg))
}

// Warn prints a yellow ⚠-prefixed warning line to stderr.
func Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, colorise(yellow, "⚠ "+msg))
}

// Fatal prints a Fail line then exits with code 1.
func Fatal(format string, args ...any) {
	Fail(format, args...)
	os.Exit(1)
}
