# Agent2 execution packet — BUG-001

You are **dev-engineer** (Agent2). Specialty only.

## Workspace
- cwd: /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
- Read first: AGENTS.md, .claude/CLAUDE.md, nearest CLAUDE.md
- Case vault: /Users/minoru/Vaults/CorpVault/50_Projects/ノア動物病院電子カルテ
- Parent Linear: BRT-4
- Unit: **BUG-001 only**

## Input
Bug section saved at:
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/bug-md-2agent-loop/packets/BUG-001/bug_section.md

## Hard rules
- claim: before edit `git branch --list 'claim/BUG-001'` then claim branch per repo convention
- Prefer isolated git worktree if concurrent risk
- Do NOT: migrate/seed apply, force-push, reset --hard, touch foreign WIP seed CSVs unless this BUG owns them
- Do NOT mark VERIFIED_FIXED
- Do NOT paste secrets/PHI into files that leave the repo
- Scoped verification only (no full-project banned commands per CLAUDE.md)
- Empty diff is failure

## Do
1. Reproduce from source (read code paths cited in bug section)
2. Minimal fix for BUG-001 only
3. Add/adjust tests if feasible
4. Run scoped tests; paste output
5. Write handoff: /Users/minoru/Vaults/CorpVault/55_Handoff/BRT-8_BUG-001_engineer.md
6. Update bug.md section 対応状況 → IMPLEMENTED_UNVERIFIED with evidence — never VERIFIED_FIXED
7. Report FACTS vs inference

## Out
Write result JSON to:
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/bug-md-2agent-loop/packets/BUG-001/result.json
with keys: status (COMPLETE|BLOCKED|FAILED), summary, files_changed, test_commands, test_exit, notes
