package checker

import (
	"context"
	"log/slog"
	"path/filepath"

	"ggpatch-robot/internal/kernel"
)

type BuildCheck struct {
	mgr *kernel.Manager
}

func NewBuildCheck(mgr *kernel.Manager) *BuildCheck {
	return &BuildCheck{mgr: mgr}
}

func (c *BuildCheck) Name() string { return "BuildCheck" }

func (c *BuildCheck) Check(ctx context.Context, _, branch, workDir string) Result {
	dir := filepath.Join(workDir, branch)
	out, err := c.mgr.BuildCheck(ctx, dir)
	if err != nil {
		slog.Warn("build check failed", "branch", branch, "error", err)
		return Result{
			Name:   c.Name(),
			Passed: false,
			Detail: "*** BuildCheck FAILED ***\n" + out + "\n",
		}
	}
	return Result{
		Name:   c.Name(),
		Passed: true,
		Detail: "*** BuildCheck PASS ***\n",
	}
}
