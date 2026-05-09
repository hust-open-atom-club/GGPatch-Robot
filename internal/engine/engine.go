package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ggpatch-robot/internal/checker"
	"ggpatch-robot/internal/config"
	"ggpatch-robot/internal/kernel"
	"ggpatch-robot/internal/mail"
)

type receiver interface {
	Receive(ctx context.Context) ([]mail.RawEmail, error)
}

type sender interface {
	Send(ctx context.Context, to []string, subject, body string) error
}

type checkerInterface interface {
	Name() string
	Check(ctx context.Context, patchPath, branch, workDir string) checker.Result
}

type Engine struct {
	cfg      *config.MailInfo
	receiver receiver
	sender   sender
	mgr      *kernel.Manager
	checkers []checkerInterface

	checkPatchPl *checker.CheckPatchPl
	applyCheck   *checker.ApplyCheck
	buildCheck   *checker.BuildCheck
}

func New(
	cfg *config.MailInfo,
	recv receiver,
	send sender,
	mgr *kernel.Manager,
	smatch *checker.Smatch,
	cocci *checker.Coccicheck,
	cppcheck *checker.Cppcheck,
) *Engine {
	return &Engine{
		cfg:          cfg,
		receiver:     recv,
		sender:       send,
		mgr:          mgr,
		checkPatchPl: &checker.CheckPatchPl{},
		applyCheck:   checker.NewApplyCheck(mgr),
		buildCheck:   checker.NewBuildCheck(mgr),
		checkers:     []checkerInterface{smatch, cocci, cppcheck},
	}
}

func (e *Engine) Run(ctx context.Context) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	if err := e.init(ctx, workDir); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	slog.Info("engine started", "interval", e.cfg.Interval)
	for {
		emails, err := e.receiver.Receive(ctx)
		if err != nil {
			slog.Warn("receive error", "error", err)
			time.Sleep(time.Duration(e.cfg.Interval) * time.Minute)
			continue
		}
		if len(emails) == 0 {
			slog.Debug("no new patches")
			time.Sleep(time.Duration(e.cfg.Interval) * time.Minute)
			continue
		}

		e.update(ctx, workDir)

		for _, raw := range emails {
			slog.Info("processing email", "from", raw.FromAddr, "subject", raw.Subject)

			report, headers, ok := e.processOne(ctx, workDir, raw)
			if !ok {
				continue
			}

			to := []string{headers.FromAddr}
			if e.cfg.MailingList != "" {
				to = append(to, e.cfg.MailingList)
			}
			subject := "Re: " + raw.Subject
			if err := e.sender.Send(ctx, to, subject, report); err != nil {
				slog.Warn("send failed", "error", err)
			}
		}

		time.Sleep(time.Duration(e.cfg.Interval) * time.Minute)
	}
}

func (e *Engine) init(ctx context.Context, workDir string) error {
	slog.Info("initializing")
	for _, d := range []string{"patch", "log"} {
		if err := os.MkdirAll(filepath.Join(workDir, d), 0777); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}
	if err := e.mgr.ClonePulls(ctx, workDir); err != nil {
		return err
	}
	// Initial build of both kernels
	for _, branch := range []string{"mainline", "linux-next"} {
		if err := e.mgr.Build(ctx, filepath.Join(workDir, branch)); err != nil {
			return fmt.Errorf("initial build %s: %w", branch, err)
		}
	}
	slog.Info("initialization done")
	return nil
}

func (e *Engine) update(ctx context.Context, workDir string) {
	slog.Debug("updating repos")
	if err := e.mgr.ClonePulls(ctx, workDir); err != nil {
		slog.Warn("update failed", "error", err)
	}
}

func (e *Engine) isWhitelisted(addr string) bool {
	for _, suffix := range e.cfg.WhiteLists {
		if strings.Contains(addr, suffix) {
			return true
		}
	}
	return false
}

func (e *Engine) processOne(ctx context.Context, workDir string, raw mail.RawEmail) (string, mail.Headers, bool) {
	patchName, patchBody, h, ok := mail.PatchExtract(raw, e.isWhitelisted)
	if !ok {
		slog.Debug("patch extraction skipped", "from", raw.FromAddr)
		return "", mail.Headers{}, false
	}

	changedPaths := checker.ParseChangedPaths(patchBody)
	if len(changedPaths) == 0 {
		slog.Debug("no changed paths in patch")
		return "", mail.Headers{}, false
	}

	// Extract log message (text before Signed-off-by/Fixes)
	logMsg := extractLogMsg(patchBody)

	// Write patch file
	patchFile := filepath.Join(workDir, "patch", patchName)
	if err := os.WriteFile(patchFile, []byte(patchBody), 0644); err != nil {
		slog.Warn("write patch failed", "error", err)
		return "", mail.Headers{}, false
	}

	// Run checker pipeline
	var results []checker.Result

	r := e.checkPatchPl.Check(ctx, patchName, "", workDir)
	results = append(results, r)
	if !r.Passed {
		report := checker.BuildReport(changedPaths, logMsg, "", results)
		return report, h, true
	}

	ar := e.applyCheck.Check(ctx, patchName, "", workDir)
	results = append(results, ar)
	if !ar.Passed {
		report := checker.BuildReport(changedPaths, logMsg, "", results)
		return report, h, true
	}
	applyBranch := ar.Detail // "linux-next" or "mainline"

	br := e.buildCheck.Check(ctx, patchName, applyBranch, workDir)
	results = append(results, br)
	if !br.Passed {
		report := checker.BuildReport(changedPaths, logMsg, applyBranch, results)
		return report, h, true
	}

	ctx = checker.WithChangedPaths(ctx, changedPaths)
	for _, c := range e.checkers {
		cr := c.Check(ctx, patchName, applyBranch, workDir)
		results = append(results, cr)
	}

	report := checker.BuildReport(changedPaths, logMsg, applyBranch, results)
	return report, h, true
}

func extractLogMsg(text string) string {
	end := strings.Index(text, "Fixes:")
	if end == -1 {
		end = strings.Index(text, "Signed-off-by:")
	}
	if end == -1 {
		end = strings.Index(text, "---")
	}
	if end > 0 {
		return text[:end]
	}
	return ""
}
