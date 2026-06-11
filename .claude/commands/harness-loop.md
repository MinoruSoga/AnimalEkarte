---
description: Run iterative refinement cycles using /plan → /tdd → /code-review → /harness-health.
argument-hint: "[--scope repo|hooks|skills|commands|agents] [--threshold 75] [--max-iterations 5] [--root path]"
---

# Harness Refinement Loop

Run controlled iterative harness improvement cycles and generate per-iteration operator prompts.

## Usage

`/harness-loop [--scope repo|hooks|skills|commands|agents] [--threshold 75] [--max-iterations 5] [--root path]`

## How this works

- Uses deterministic harness scoring (`scripts/harness-audit.js`) through `scripts/harness-health.js`.
- Repeats checks until one of these is met:
  - score reaches the threshold (`OK`)
  - max iterations reached
  - top 3 failing actions stop changing (stall)
- For each iteration, outputs next operator runbook:
  1) `/plan`
  2) `/tdd`
  3) implement high-impact actions
  4) `/code-review`
  5) rerun this loop for the next iteration

## Recommended usage

- PR gate: ` /harness-loop --scope repo --threshold 85 --max-iterations 5`
- Daily stabilization: ` /harness-loop --scope repo --threshold 75 --max-iterations 3`

## Output

- Text mode: prints each iteration summary and next actions.
- JSON mode: machine-readable run history and generated prompts.

## Arguments

- `--scope` optional: `repo`(default), `hooks`, `skills`, `commands`, `agents`
- `--threshold` optional integer (0-100)
- `--max-iterations` optional (default: 5)
- `--root` optional project root
- `--format` optional `text|json`
- `--owner` optional blocker owner (default: `unassigned`)
- `--due-days` optional blocker due days (default: 2)
