# ggpatch-robot

Automated Linux kernel patch testing bot for Google Groups. Monitors a Google Group via IMAP, runs a pipeline of static analysis checks against `mainline` and `linux-next` kernel trees, and replies with results via SMTP.

## Prerequisites

- Go 1.23+
- `git`, `make`, `cppcheck`
- A subscribed email address for the target Google Group

## Quick Start

```bash
cp config.example.json config.json
# Fill in email credentials, whitelists in config.json
make
./ggpatch-robot -config config.json
```

## Configuration

`config.json` fields:

| Field | Description | Default |
| --- | --- | --- |
| `username` | Email account for IMAP/SMTP (e.g., `robot@126.com`) | — |
| `password` | Email password | — |
| `procs` | Parallel jobs for kernel compilation | `20` |
| `interval` | Minutes between inbox checks | `20` |
| `whiteLists` | Allowed recipient address patterns | `[]` |
| `mailingList` | CC address for reply emails | `kernel_testing_robot@googlegroups.com` |

Supported email domains: `126.com`, `hust.edu.cn`.

## Checker Pipeline

1. **CheckPatchPl** — `scripts/checkpatch.pl` on the patch
2. **ApplyCheck** — `git apply --check` on `linux-next`, fallback to `mainline`
3. **BuildCheck** — `make -jN` incremental build
4. **Coccicheck** — before/after with coccinelle
5. **Cppcheck** — before/after with cppcheck

Patches with a `Reviewed-by:` tag or non-whitelisted recipients are skipped. The pipeline stops at the first gate failure.

## Architecture

```text
cmd/ggpatch-robot/main.go
internal/
  config/       # config.json parsing & validation
  mail/         # IMAP receive, SMTP send, patch extraction
  kernel/       # git clone/pull/build/apply/revert
  checker/      # Checker interface, pipeline, each checker tool
  engine/       # main loop orchestration
```

## License

Apache-2.0
