# Codex project entrypoint

Read [AGENTS.md](../AGENTS.md), then [.claude/CLAUDE.md](../.claude/CLAUDE.md) and only task-relevant references. The project workflow and completion contract are in [agent-harness.md](../docs/ops/agent-harness.md).

Generic model, sandbox, Hook, connector and lifecycle preferences belong in user settings. If the user `agent-task-lifecycle` Skill is available, use it with the project constraints; its absence must not require a new prompt or prevent ordinary authorized work.

Use the accepted task and its acceptance criteria to investigate, implement, review, test and fix findings. Do not require the user to write separate BE/FE prompts or approve each iteration. Preserve clinical isolation and foreign WIP; migration application, claim release and integration into main remain user actions.

Read the relevant specification via [docs/README.md](../docs/README.md), the API contract at [backend/docs/api.yaml](../backend/docs/api.yaml), and Linear for current task state. A missing connector means UNKNOWN external state, not completed work. Draft external updates locally until posting is authorized.

Discover tools and Skills in the current session. Legacy `/harness` or `/implement` examples are not proof that Codex exposes those commands. Do not infer feature maturity or available agents from a different runtime.
