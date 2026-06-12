---
description: Run a harness health gate with threshold-based blocker tracking.
argument-hint: "[--scope repo|hooks|skills|commands|agents] [--threshold 75] [--format text|json] [--root path]"
---

# Harness Health Command

Run harness health checks and convert audit output into an actionable gate.

## Purpose

- Use existing deterministic logic from `scripts/harness-audit.js`.
- Return score/threshold status in one place.
- Turn top `harness-audit` actions into explicit blockers when below threshold.

## Usage

`/harness-health [--scope repo|hooks|skills|commands|agents] [--threshold 75] [--format text|json] [--root path]`

## Recommended Pattern

- Daily: `/harness-health --scope repo --threshold 75`
- PR gate: `/harness-health --scope repo --threshold 85 --format json`
- Weekly: `/harness-health --scope commands --threshold 80`

## Deterministic Engine

This command always runs:

```bash
node scripts/harness-audit.js <scope> --format json --root <root>
```

and adds only the following interpretation layer:

- `status = BLOCKED` if score is below threshold.
- `status = OK` if score is above or equal to threshold.
- Blocker list with `owner` and `due` when blocked.

Do not invent additional checks.

## Output Contracts

### Text mode

- Show scope/root/score/threshold/category count.
- Show top 3 actions.
- If blocked, show blocker lines with `owner/due/action`.

### JSON mode

- Emit strict JSON schema used by automation.

## Arguments

- `--scope` optional: `repo`(default), `hooks`, `skills`, `commands`, `agents`
- `--threshold` optional numeric threshold (0-100)
- `--format` optional `text` or `json`
- `--root` optional repo root
- `--owner` optional blocker owner (default: `unassigned`)
- `--due-days` optional number of days for blocker due date (default: `3`)
