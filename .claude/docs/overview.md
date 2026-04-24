# Claude Code Framework Overview

AnimalEkarte プロジェクトの Claude Code 設定概要。

## Quick Links

| Category | Location | Purpose |
|----------|----------|---------|
| **Configuration** | `.claude/settings.json` | permissions, hooks, env |
| **Rules** | `.claude/rules/` | 1 auto-loaded rule file (claude-code-usage.md) |
| **Hooks** | `.claude/hooks/` | 14 Node.js/shell hooks (ECC準拠 JSON protocol) |
| **Memory** | `~/.claude/projects/.../memory/` | persistent knowledge files |
| **Workflows** | `.claude/docs/workflows/` | Development workflow guides |
| **Patterns** | `.claude/docs/patterns/` | Pattern implementation references |

## Architecture

```
┌─────────────────────────────────────┐
│  Claude Code Session                │
├─────────────────────────────────────┤
│  Agent Team (17 agents)             │
│  - architect (Opus)                 │
│  - planner (Opus)                   │
│  - security-analyst (Opus)          │
│  - implementer (Sonnet)             │
│  - go-expert (Sonnet)               │
│  - go-reviewer (Sonnet)             │
│  - typescript-reviewer (Sonnet)     │
│  - refactor-cleaner (Sonnet)        │
│  - database-reviewer (Sonnet)       │
│  - performance-optimizer (Sonnet)   │
│  - tdd-guide (Sonnet)               │
│  - test-strategist (Sonnet)         │
│  - silent-failure-hunter (Sonnet)   │
│  - debugger (Haiku)                 │
│  - formatter (Haiku)                │
│  - researcher (Haiku)               │
│  - reviewer (Haiku)                 │
├─────────────────────────────────────┤
│  Rules (1 file, auto-loaded)        │
│  - claude-code-usage.md             │
│  Refs (14 files, on-demand)         │
│  go-language, typescript-react,     │
│  gin-architecture-compliance,       │
│  database-design, testing, api,     │
│  naming-conventions, security, ...  │
├─────────────────────────────────────┤
│  Hooks (14 scripts, ECC protocol)   │
│  PreToolUse:                        │
│    - block-dangerous (exit 2)       │
│    - block-no-verify (exit 2)       │
│    - commit-quality (exit 2)        │
│    - large-file-block (exit 2)      │
│    - config-protection (warn)       │
│    - git-push-reminder (warn)       │
│  PostToolUse:                       │
│    - console-warn (warn)            │
│    - file-size-warn (warn)          │
│    - format-go (gofmt)              │
│    - typecheck-ts (tsc)             │
│  PreCompact / Stop / SessionStart:  │
│    - save-state, save-progress      │
│    - desktop-notify, session-init   │
├─────────────────────────────────────┤
│  Memory (persistent knowledge)      │
│  - project-architecture.md          │
│  - coding-standards.md              │
│  - backend-patterns.md              │
│  - frontend-patterns.md             │
│  - db-schema-reference.md           │
│  - security-checklist.md            │
└─────────────────────────────────────┘
```

## Key Files

### Configuration
- `settings.json`: permissions (deny/allow/ask), hooks, env
- `CLAUDE.md`: Project instructions (overrides defaults)

### Hooks (Node.js, JSON protocol)
- `hooks/pre-bash-block-dangerous.js`: Block rm -rf /, dd, mkfs (exit 2)
- `hooks/pre-bash-block-no-verify.js`: Block --no-verify / --no-gpg-sign (exit 2)
- `hooks/pre-bash-commit-quality.js`: Commit quality gate — secrets, commit msg (exit 2)
- `hooks/pre-bash-git-push-reminder.js`: Warn before git push (stderr)
- `hooks/pre-write-large-file-block.js`: Block 800+ line files (exit 2)
- `hooks/pre-edit-config-protection.js`: Warn on linter/formatter config edits
- `hooks/post-edit-console-warn.js`: Detect console.log (stderr warning)
- `hooks/post-edit-file-size-warn.js`: Warn when file exceeds 500/800 lines
- `hooks/post-edit-format-go.js`: Auto gofmt via Docker (non-blocking)
- `hooks/post-edit-typecheck-ts.js`: Auto tsc --noEmit via Docker (non-blocking)
- `hooks/pre-compact-save-state.js`: Save git state before compaction
- `hooks/stop-save-progress.js`: Save session progress on stop
- `hooks/stop-desktop-notify.js`: macOS notification when Claude finishes
- `hooks/session-init.sh`: Session start log

## Development Workflow

1. **Session Start**: Rules auto-loaded, hooks active
2. **Coding**: Hooks auto-check (dangerous commands, console.log, file size)
3. **Go edits**: Auto-formatted via gofmt
4. **Before push**: Hook warns to review changes

## Next Steps

See `.claude/docs/workflows/` for specific development workflows.
See `.claude/docs/patterns/` for pattern implementation references.
