package kernel

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Manager struct {
	procs int
}

func NewManager(procs int) *Manager {
	if procs == 0 {
		procs = 20
	}
	return &Manager{procs: procs}
}

// ClonePulls ensures mainline, linux-next, and smatch repos exist and are up to date.
func (m *Manager) ClonePulls(ctx context.Context, workDir string) error {
	dirs := []struct {
		name string
		url  string
	}{
		{"mainline", "https://mirrors.hust.college/git/linux.git"},
		{"linux-next", "https://mirrors.hust.college/git/linux-next.git"},
		{"smatch", "git://repo.or.cz/smatch.git"},
	}

	for _, d := range dirs {
		if err := m.cloneOrPull(ctx, workDir, d.name, d.url); err != nil {
			return fmt.Errorf("%s: %w", d.name, err)
		}
	}
	return nil
}

func (m *Manager) cloneOrPull(ctx context.Context, workDir, name, url string) error {
	repoPath := filepath.Join(workDir, name)
	if err := run(ctx, workDir, "ls", "-l", name); err != nil {
		slog.Info("cloning", "repo", name, "url", url)
		cloneURL := url
		if name == "mainline" {
			if err := run(ctx, workDir, "git", "clone", "--depth=1", url); err != nil {
				return fmt.Errorf("clone: %w", err)
			}
			if err := run(ctx, workDir, "mv", "linux", name); err != nil {
				return fmt.Errorf("rename: %w", err)
			}
		} else {
			if err := run(ctx, workDir, "git", "clone", "--depth=1", cloneURL, name); err != nil {
				return fmt.Errorf("clone: %w", err)
			}
		}
		return nil
	}
	slog.Debug("pulling", "repo", name)
	if err := run(ctx, repoPath, "git", "pull"); err != nil {
		slog.Warn("pull failed", "repo", name, "error", err)
	}
	return nil
}

// Build compiles the kernel tree at the given path with allyesconfig.
func (m *Manager) Build(ctx context.Context, dir string) error {
	if err := run(ctx, dir, "make", "allyesconfig"); err != nil {
		return fmt.Errorf("allyesconfig: %w", err)
	}
	if err := run(ctx, dir, "make", "-j"+strconv.Itoa(m.procs)); err != nil {
		return fmt.Errorf("make: %w", err)
	}
	return nil
}

// BuildCheck runs make -jN in the given directory for incremental build check.
func (m *Manager) BuildCheck(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "make", "-j"+strconv.Itoa(m.procs))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// Apply applies a patch to the given branch directory.
func (m *Manager) Apply(ctx context.Context, dir, patchPath string) error {
	return run(ctx, dir, "git", "apply", patchPath)
}

// Revert reverts a previously applied patch.
func (m *Manager) Revert(ctx context.Context, dir, patchPath string) error {
	return run(ctx, dir, "git", "apply", "-R", patchPath)
}

// ApplyCheck tests whether a patch can be applied cleanly.
func (m *Manager) ApplyCheck(ctx context.Context, dir, patchPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "apply", "--check", patchPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// RunScript executes a script in a given directory with arguments.
func (m *Manager) RunScript(ctx context.Context, dir, script string, args ...string) (string, error) {
	cmdArgs := append([]string{script}, args...)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Checkers returns the path to smatch kchecker and coccinelle for a branch.
func (m *Manager) SmatchDir(workDir string) string {
	return filepath.Join(workDir, "smatch")
}

func run(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
}
