package checker

import (
	"context"

	"ggpatch-robot/internal/kernel"
)

type Cppcheck struct {
	mgr *kernel.Manager
}

func NewCppcheck(mgr *kernel.Manager) *Cppcheck {
	return &Cppcheck{mgr: mgr}
}

func (c *Cppcheck) Name() string { return "Cppcheck" }

func (c *Cppcheck) Check(ctx context.Context, patchPath, branch, workDir string) Result {
	return staticCheck(ctx, c.Name(), patchPath, branch, workDir, c.mgr,
		"warn", "error", nil,
		"cppcheck")
}
