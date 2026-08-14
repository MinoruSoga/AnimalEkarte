# Phase 1 dry-run — 2026-08-07

Plan: `CorpVault/.../10_pcg_loop_automation_plan.md`

## Command

```bash
python3 /Users/minoru/Vaults/CorpVault/50_Projects/ノア動物病院電子カルテ/bug-md-2agent-loop/run_loop.py \
  --repo /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte \
  --dry-run --mode auto --bug-id BUG-001
```

## Result

- ledger auto: `STATUS.md` (bug.md is move stub)
- PICK BUG-001 [IMPLEMENTED_UNVERIFIED]
- PACKET `reports/bug-md-2agent-loop/packets/BUG-001/{agent2_prompt.md,bug_section.md,unit.json}`
- graph saver on draft task-spec: **exit 1** (schema/paths) — expected Phase 1; `--mode auto` falls back to hermes on `--execute`
- OPEN count: 0 → default queue empty without `--bug-id` / `--include-iu`

## Next (Phase 1b)

- Bind real write/check paths so saver accepts one BUG
- Or run production path: `--mode hermes --execute --bug-id ...`
