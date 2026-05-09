package checker

import (
	"context"
	"log/slog"
)

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

// Pipeline runs checkers sequentially, respecting gates:
// Apply failure → skip Build; Build failure → skip StaticAnalysis.
type Pipeline struct {
	procs    int
	checkers []Checker
}

func NewPipeline(procs int, checkers ...Checker) *Pipeline {
	return &Pipeline{
		procs:    procs,
		checkers: checkers,
	}
}

// Run executes the full pipeline and returns all results.
func (p *Pipeline) Run(ctx context.Context, patchPath, workDir string) []Result {
	var results []Result
	var applyBranch string

	for _, c := range p.checkers {
		slog.Debug("running checker", "checker", c.Name(), "patch", patchPath)
		var branch string
		switch c.Name() {
		case "ApplyCheck", "BuildCheck", "Smatch", "Coccicheck", "Cppcheck":
			branch = applyBranch
		}
		r := c.Check(ctx, patchPath, branch, workDir)
		results = append(results, r)

		if c.Name() == "ApplyCheck" {
			applyBranch = r.Detail // Pass takes the branch name, for example "mainline" or "linux-next"
		}
		if !r.Passed {
			switch c.Name() {
			case "ApplyCheck", "CheckPatchPl", "BuildCheck":
				slog.Warn("pipeline gate failed, stopping", "checker", c.Name())
				return results
			}
		}
	}
	return results
}

func stringPtr(s string) string { return s }
