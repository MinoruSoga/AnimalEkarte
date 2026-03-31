# Claude Code Framework Overview

AnimalEkarte プロジェクトの Claude Code 設定概要。

## Quick Links

| Category | Location | Purpose |
|----------|----------|---------|
| **Configuration** | `.claude/settings.json` | permissions, hooks, env |
| **Rules** | `.claude/rules/` | 8 enforced rules (自動ロード) |
| **Hooks** | `.claude/hooks/` | 6 Node.js/shell hooks (ECC準拠 JSON protocol) |
| **Memory** | `~/.claude/projects/.../memory/` | persistent knowledge files |
| **Workflows** | `.claude/docs/workflows/` | Development workflow guides |
| **Patterns** | `.claude/docs/patterns/` | Pattern implementation references |

## Architecture

```
┌─────────────────────────────────────┐
│  Claude Code Session                │
├─────────────────────────────────────┤
│  Agent Team (10 agents)             │
│  - architect (Opus)                 │
│  - implementer (Sonnet)             │
│  - reviewer (Haiku)                 │
│  - debugger (Haiku)                 │
│  - researcher (Haiku)               │
│  - formatter (Haiku)                │
│  - go-expert (Sonnet)               │
│  - performance-optimizer (Sonnet)   │
│  - security-analyst (Opus)          │
│  - test-strategist (Sonnet)         │
├─────────────────────────────────────┤
│  Rules (8 files, auto-loaded)       │
│  go-language, typescript-react,     │
│  database-design, docker-rules,     │
│  git-workflow, performance-rules,   │
│  accessibility-rules, error-handling│
├─────────────────────────────────────┤
│  Hooks (ECC JSON stdin/stdout)      │
│  PreToolUse:                        │
│    - block-dangerous (Bash, exit 2) │
│    - git-push-reminder (Bash, warn) │
│    - large-file-block (Write, exit 2)│
│  PostToolUse:                       │
│    - console-warn (Edit, warn)      │
│    - format-go (Edit, gofmt)        │
│  SessionStart:                      │
│    - session-init (project detect)  │
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
- `hooks/pre-bash-git-push-reminder.js`: Warn before git push (stderr)
- `hooks/pre-write-large-file-block.js`: Block 800+ line files (exit 2)
- `hooks/post-edit-console-warn.js`: Detect console.log (stderr warning)
- `hooks/post-edit-format-go.js`: Auto gofmt via Docker
- `hooks/session-init.sh`: Session start log

## Development Workflow

1. **Session Start**: Rules auto-loaded, hooks active
2. **Coding**: Hooks auto-check (dangerous commands, console.log, file size)
3. **Go edits**: Auto-formatted via gofmt
4. **Before push**: Hook warns to review changes

## Next Steps

See `.claude/docs/workflows/` for specific development workflows.
See `.claude/docs/patterns/` for pattern implementation references.
