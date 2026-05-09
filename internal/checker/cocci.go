package checker

import (
	"context"
	"strings"

	"ggpatch-robot/internal/kernel"
)

type Coccicheck struct {
	mgr *kernel.Manager
}

func NewCoccicheck(mgr *kernel.Manager) *Coccicheck {
	return &Coccicheck{mgr: mgr}
}

func (c *Coccicheck) Name() string { return "Coccicheck" }

func (c *Coccicheck) Check(ctx context.Context, patchPath, branch, workDir string) Result {
	cToO := func(path string) string {
		end := strings.Index(path, ".c")
		if end == -1 {
			return path
		}
		return path[:end] + ".o"
	}
	return staticCheck(ctx, c.Name(), patchPath, branch, workDir, c.mgr,
		"WARNING", "ERROR", cToO,
		"make", "C=1", "CHECK=scripts/coccicheck")
}
