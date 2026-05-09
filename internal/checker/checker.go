package checker

import "context"

// Result is the outcome of a single checker.
type Result struct {
	Name   string
	Passed bool
	Detail string
}

// Checker is the interface each static analysis tool implements.
type Checker interface {
	Name() string
	Check(ctx context.Context, patchPath, branch, workDir string) Result
}
