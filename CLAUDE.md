# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
make                    # builds ggpatch-robot binary
make test               # runs all tests
./ggpatch-robot -config config.json
```

Module: `ggpatch-robot`, Go 1.23+. Dependencies: `go-imap`, `go-imap-id`, `go-message`.

## Architecture

ggpatch-robot is an automated Linux kernel patch testing bot. It polls a Google Group via IMAP for `[PATCH*` emails, runs a pipeline of static analysis checks against mainline and linux-next kernel trees, and replies with results via SMTP.

### Package Layout

```text
cmd/ggpatch-robot/main.go          # entry point, wires dependencies
internal/
  config/       # parse & validate config.json, defaults
  mail/         # IMAP receiver, SMTP sender, patch extraction from email body
  kernel/       # clone/pull/build/apply/revert for kernel + smatch repos
  checker/      # Checker interface, pipeline orchestration, each checker, Logcmp
  engine/       # main loop: init → poll → process → check → send → sleep
```

### Data Flow

1. `config.Load()` parses config.json, resolves email server settings from domain
2. `Engine.Run()` initializes: creates `patch/`/`log/` dirs, clones & builds mainline, linux-next, smatch
3. Main loop: IMAP `Receive()` → `update()` repos → for each `[PATCH*` email:
   - `PatchExtract()` filters whitelist/Reviewed-by, extracts patch body + changed paths
   - Checker pipeline: `CheckPatchPl` → `ApplyCheck` (linux-next, fallback mainline) → `BuildCheck` → `StaticAnalysis` (Smatch, Coccicheck, Cppcheck)
   - SMTP `Send()` reply with report
4. Static checkers do before/after diff: run tool, apply patch, run again, revert, `Logcmp()`

### Key Conventions

- `context.Context` first param on all I/O and exec methods
- Interfaces defined at call site (small, 1-2 methods): `receiver`, `sender`, `Checker`
- `log/slog` structured logging; error wrapping with `%w`; no global state
- The old code in `KTBot/` is preserved untouched; the rewrite lives at the repo root
