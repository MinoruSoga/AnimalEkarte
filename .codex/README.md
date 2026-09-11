# Codex project settings

Start at [CODEX.md](CODEX.md) and [AGENTS.md](../AGENTS.md). Project rules are maintained in [.claude/CLAUDE.md](../.claude/CLAUDE.md); operation and verification guidance is in [agent-harness.md](../docs/ops/agent-harness.md).

Keep generic configuration in the user scope. Generated agents, commands and Skills must be updated through their canonical `.claude` sources and the repository generator, not edited as independent copies.

Project role generation uses `.claude/codex-agent-manifest.json` (16 domain roles). Generic architect, debugger, formatter and reviewer roles inherit user configuration; domain planner, implementer and test-strategist use `animalekarte-` names. Both mirror generators record file hashes and stop before overwriting modified or unknown outputs.
