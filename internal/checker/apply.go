package checker

import (
	"context"
	"log/slog"
	"path/filepath"

	"ggpatch-robot/internal/kernel"
)

type ApplyCheck struct {
	mgr *kernel.Manager
}

func NewApplyCheck(mgr *kernel.Manager) *ApplyCheck {
	return &ApplyCheck{mgr: mgr}
}

func (c *ApplyCheck) Name() string { return "ApplyCheck" }

func (c *ApplyCheck) Check(ctx context.Context, patchPath, _, workDir string) Result {
	patch := filepath.Join(workDir, "patch", patchPath)
	branches := []string{"linux-next", "mainline"}

	for _, branch := range branches {
		dir := filepath.Join(workDir, branch)
		out, err := c.mgr.ApplyCheck(ctx, dir, patch)
		if err == nil {
			slog.Info("patch applies", "branch", branch)
			return Result{
				Name:   c.Name(),
				Passed: true,
				Detail: branch, // carries the winning branch
			}
		}
		slog.Debug("apply failed on branch", "branch", branch, "output", out)
	}
	return Result{
		Name:   c.Name(),
		Passed: false,
		Detail: "*** ApplyCheck FAILED ***\nPatch does not apply to mainline or linux-next\n",
	}
}
