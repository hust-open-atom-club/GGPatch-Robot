package checker

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"ggpatch-robot/internal/kernel"
)

func staticCheck(ctx context.Context, name, patchPath, branch, workDir string, mgr *kernel.Manager,
	warnKey, errKey string, pathFn func(string) string, checker string, checkerArgs ...string) Result {

	dir := filepath.Join(workDir, branch)
	if branch == "" {
		return Result{Name: name, Passed: true, Detail: fmt.Sprintf("*** %s SKIP (no branch) ***\n", name)}
	}

	patchFile := filepath.Join(workDir, "patch", patchPath)

	paths, _ := changedPathsFromContext(ctx)

	var details []string
	for _, path := range paths {
		if !strings.Contains(path, ".c") {
			continue
		}

		target := path
		if pathFn != nil {
			target = pathFn(path)
		}

		args := append([]string{}, checkerArgs...)
		args = append(args, target)
		preOut, _ := mgr.RunScript(ctx, dir, checker, args...)

		if err := mgr.Apply(ctx, dir, patchFile); err != nil {
			slog.Warn("static check: apply failed", "checker", name, "error", err)
			break
		}
		postOut, _ := mgr.RunScript(ctx, dir, checker, args...)
		if err := mgr.Revert(ctx, dir, patchFile); err != nil {
			slog.Warn("static check: revert failed", "checker", name, "error", err)
		}

		diff := Logcmp(preOut, postOut, warnKey, errKey)
		if hasDiff(diff) {
			details = append(details, formatDiff(diff))
		}
	}

	if len(details) > 0 {
		return Result{
			Name:   name,
			Passed: false,
			Detail: "*** " + name + " FAILED ***\n" + strings.Join(details, "\n"),
		}
	}
	return Result{
		Name:   name,
		Passed: true,
		Detail: "*** " + name + " PASS ***\n",
	}
}

func hasDiff(d LogDiff) bool {
	return len(d.NewErrors) > 0 || len(d.NewWarnings) > 0 ||
		len(d.UnsolvedErrors) > 0 || len(d.UnsolvedWarnings) > 0
}

func formatDiff(d LogDiff) string {
	var b strings.Builder
	if len(d.UnsolvedWarnings) > 0 {
		b.WriteString("Unsolved warning:\n")
		for _, w := range d.UnsolvedWarnings {
			b.WriteString(w + "\n")
		}
	}
	if len(d.UnsolvedErrors) > 0 {
		b.WriteString("Unsolved error:\n")
		for _, e := range d.UnsolvedErrors {
			b.WriteString(e + "\n")
		}
	}
	if len(d.NewWarnings) > 0 {
		b.WriteString("New warning:\n")
		for _, w := range d.NewWarnings {
			b.WriteString(w + "\n")
		}
	}
	if len(d.NewErrors) > 0 {
		b.WriteString("New error:\n")
		for _, e := range d.NewErrors {
			b.WriteString(e + "\n")
		}
	}
	return b.String()
}
