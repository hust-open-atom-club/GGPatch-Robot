package checker

import (
	"context"
	"log/slog"
	"os/exec"
	"path/filepath"
)

type CheckPatchPl struct{}

func (c *CheckPatchPl) Name() string { return "CheckPatchPl" }

func (c *CheckPatchPl) Check(ctx context.Context, patchPath, _, workDir string) Result {
	script := filepath.Join(workDir, "mainline", "scripts", "checkpatch.pl")
	patch := filepath.Join(workDir, "patch", patchPath)

	cmd := exec.CommandContext(ctx, script, patch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("checkpatch.pl failed", "error", err)
		return Result{
			Name:   c.Name(),
			Passed: false,
			Detail: string(out),
		}
	}
	return Result{
		Name:   c.Name(),
		Passed: true,
		Detail: "*** CheckPatch PASS ***\n",
	}
}
