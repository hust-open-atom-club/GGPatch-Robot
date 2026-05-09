package checker

import (
	"context"
	"fmt"
)

type contextKey struct{}

// WithChangedPaths stores the changed file paths in ctx for static checkers.
func WithChangedPaths(ctx context.Context, paths []string) context.Context {
	return context.WithValue(ctx, contextKey{}, paths)
}

func changedPathsFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(contextKey{}).([]string)
	return v, ok
}

// ParseChangedPaths extracts the changed file list from patch text.
// Lines like "diff --git a/fs/ext4/inode.c b/fs/ext4/inode.c"
func ParseChangedPaths(patchBody string) []string {
	var paths []string
	for _, line := range splitLines(patchBody) {
		if len(line) > 13 && line[:13] == "diff --git a/" {
			sub := line[13:]
			idx := 0
			for i, c := range sub {
				if c == ' ' {
					idx = i
					break
				}
			}
			if idx > 0 {
				paths = append(paths, sub[:idx])
			}
		}
	}
	return paths
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// BuildReport assembles the final email body from all checker results.
func BuildReport(changedPaths []string, logMsg, applyBranch string, results []Result) string {
	var s string
	s += "--- Changed Paths ---\n"
	for _, p := range changedPaths {
		s += p + "\n"
	}
	if logMsg != "" {
		s += "\n--- Log Message ---\n" + logMsg + "\n"
	}
	s += "\n--- Test Result ---\n"
	for _, r := range results {
		s += r.Detail
	}
	return s
}

// FormatResult is a no-op error wrapper for consistent display.
func FormatResult(r Result) string {
	return fmt.Sprintf("[%s] passed=%v\n%s", r.Name, r.Passed, r.Detail)
}
