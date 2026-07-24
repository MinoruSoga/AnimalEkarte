---
name: golang-refactoring
description: "Golang refactoring — the safe, at-scale process for restructuring existing Go code: a coverage-adaptive safety net, tool-driven behavior-preserving transforms (gopls Rename/Inline/Extract, `gofmt -r`, `eg`, `gopatch`, `go/analysis` fixers), the Fowler catalog mapped to Go, breaking import cycles, moving types across packages, and small reversible landing units that defer to repository-specific execution rules. Apply when code is hard to maintain, a function/type has grown too large, a code smell needs fixing, adding a feature is blocked by the current structure, or the user asks to clean up, refactor, or improve Go code — also for renaming at scale, extracting functions/interfaces, moving code between packages, splitting packages, or planning a multi-step refactor. Target architecture, verification, and delivery always defer to repository-local rules; external skills may supply only compatible mechanics."
license: MIT
metadata:
  author: samber
  version: "1.0.0"
allowed-tools: Read Edit Write Glob Grep Bash(docker:*) Bash(git:*) Bash(gopls:*) Bash(benchstat:*) LSP mcp__gopls__* Agent AskUserQuestion EnterWorktree ExitWorktree WebFetch WebSearch
---

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-refactoring` skill takes precedence.

> ⚠️ **ANIMALEKARTE PROJECT OVERRIDE (BE9 — community defaults belowより優先)**
>
> - BE9移行は2026-07-24にcode complete（release pending）。境界の正本は[ADR-006](../../../docs/architecture/adr/006-backend-domain-package-boundaries.md)と[boundary map](../../../docs/architecture/be9-2a-boundary-map.md)、完遂経緯はgit履歴とする。旧BE8およびSession A/Bの履歴を実行手順として使わない。
> - PR/push/`gh`等の外部操作は自動実行しない。Session Aはcleanかつquiescentなlocal `main`の唯一writer、Session Bは同じimmutable baseから作る専用branch/worktreeのwriterとする。既存のdirty差分を共有worktreeから変更・破棄せず、central surface、handoff、共有DB lease条件を満たす場合だけ並行化する。
> - bareなGo commandとfull-repository commandは、このskillとreferencesにあるcommunity例も含めて実行しない。[`.claude/CLAUDE.md`](../../CLAUDE.md)とBE9のbatchごとのgateに従い、変更package/fileだけをDocker経由で検証する。full test/lint等が必要ならユーザー手動gateとして提示する。
> - DIは`main.go`だけに限定しない。closure/struct/constructorを使い、`cmd/api`または必要最小限のcomposition packageで型安全に組み立てる。package globalやuntyped context injectionを新設しない。
> - 本skillから再利用するのはtool-driven transform、blast-radius safety net、structural/behavioral分離、依存実測に基づくcycle解消である。genericなstacked-PR/refactoring-branch、固定行数、毎stepの人手承認はAnimalEkarteへ適用しない。

**Persona:** You are a Go refactoring engineer. You never change structure and behavior in the same landing unit — you keep a green test net, prefer behavior-preserving tools over hand-edits, and make each unit coherent, reversible, and reviewable.

**Thinking mode:** Apply the project's extended-reasoning criteria for architecture and large-refactor planning. Map blast radius, dependency order, ownership, and parallel-safety before editing; do not assume a particular harness command such as `ultrathink` exists.

**Orchestration mode:** Use the BE9 session ownership and synchronization barriers for multi-step work. A single-pass automation is appropriate only for one bounded mechanical sweep whose files, symbols, dependencies, and verification do not overlap another active lane.

**Modes:**

- **Plan mode** (mandatory gate before any edit) — use gopls or repository search to map structure and blast radius, build a refactoring inventory, and decide ordering. Ask before execution only when an unresolved choice materially changes scope; once scope is authorized, do not add mid-task approval gates. Use [workflow.md](references/workflow.md) only for inventory and ordering concepts.
- **Execute mode** — follow BE9's Session A/B ownership, integration-tip, worktree, handoff, and shared-DB barriers. Session A writes directly to a clean local `main`; Session B writes only its disjoint domain-local candidate in one separate worktree from the same immutable base. Session A alone writes central/shared surfaces and integrates Session B's re-frozen candidate serially.
- **Simple-sweep mode** — apply one bounded mechanical, behavior-preserving transform with no overlapping writer or dependent step.
- **Review mode** — verify structural/behavioral separation, behavior preservation, ownership, and scoped gate evidence before accepting a landing unit.

**Dependencies:** Use `gopls`, `golangci-lint`, `benchstat`, `deadcode`, `eg`, or `gopatch` only when already available in an approved project/container environment. Do not install host tooling automatically. If a preferred tool is unavailable, use repository search plus the narrowest safe alternative and record the limitation.

# Go Refactoring — Safe Change at Scale

- Refactoring (Fowler) is changing code's internal structure to make it easier to understand or cheaper to modify, **without changing observable behavior**.
- Go tooling can prove several transforms are behavior-preserving _by construction_ — e.g. gopls refuses a Rename rather than risk a broken build.
- That guarantee is silent on anything reflection can reach (struct tags, `text/template` field references) — a safety net still matters.

## The Core Loop

**Understand → Safety net → Small tool-driven step → Verify → Atomic single-category landing unit.** Repeat.

1. **Understand** — map the change's blast radius with gopls (references, call hierarchy, package API) before touching anything.
2. **Safety net** — before touching code with inadequate coverage, add tests first.
   - Gate the strategy on the _blast radius's_ test coverage, not global coverage.
   - Treat writing that test as your own mechanism for checking the change — not a formality left for the reviewer. A green suite you wrote yourself is what actually lets you tell "this is behavior-preserving" from "I hope this is behavior-preserving."
   - See [safety-net.md](references/safety-net.md) for the HIGH/MEDIUM/LOW thresholds and characterization-testing recipes for untested code.
3. **Small tool-driven step** — prefer a mechanical, tool-driven transform over a hand-edit. See [go-tooling.md](references/go-tooling.md) and [catalog.md](references/catalog.md).
4. **Verify** — run Docker-based checks only for changed packages/files and the slice-specific gates named by BE9. Add scoped `-race` for concurrency changes and `benchstat`-backed benchmarks for hot paths. Do not auto-run bare or full `./...` commands.
5. **Atomic single-category landing unit** — the unit is purely structural or purely behavioral, never both.

## Hard Rules

- **Never mix structural and behavioral changes in one landing unit.**
  - A reviewer scrutinizing a rename for correctness and a reviewer scrutinizing a feature for side effects need different postures.
  - Mixing them forces one reviewer to wear both hats at once, and the fast, low-scrutiny review a pure rename deserves gets lost.
- **Split a code move from a code optimization into two sequential landing units, even though both are structural.**
  - They need different verification — the move is proven safe by gopls plus build/test, the optimization needs benchmarks and a closer correctness read.
  - They touch the same code, so run them one after another; use separate worktrees only under the BE9 ownership rules.
  - Do not use a fixed line-count threshold. Split by symbol ownership, dependency, behavior category, independent verification, and revertability.
- **Prefer gopls Rename/Inline over LLM hand-edits.**
  - Both are behavior-preserving by construction — Rename refuses on shadowing, interface-satisfaction breakage, or malformed code rather than silently producing a bad diff; Inline substitutes side-effect-bearing arguments into `var` temporaries rather than duplicating them.
  - A hand-edit across dozens of call sites has no such guarantee and measurably misses cases.
- **When a change recurs across many sites, generate a rewrite tool instead of hand-editing each site.**
  - Escalate `gofmt -r` → `eg` → `gopatch` → a `go/analysis` fixer, in order of increasing power (see [go-tooling.md](references/go-tooling.md)).
  - A generated tool is reviewable, re-runnable, and testable against golden files — dozens of individual hand-edits are none of those things.
- **Use a type alias (`type A = B`) only while real compatibility callers remain.**
  - If fan-in is already zero, remove the old symbol instead of adding a facade. Every temporary alias/delegate needs a concrete deletion condition. See [structural.md](references/structural.md) for mechanics, not a mandatory project policy.
- **Resolve import cycles from measured consumers and ownership.**
  - Prefer a consumer-side minimal interface when a real consumer boundary exists, but also evaluate type ownership, value/function parameters, explicit orchestration, or a genuinely cohesive shared leaf. Do not create speculative `common`/`util`/`interfaces` packages.
- **Resolve expensive design choices before execution.**
  - Cross-package moves, exported-API changes, deletion, versioning, or untested code must be represented in the accepted plan. Once scope is clear, continue without mid-task confirmation unless a safety boundary or material scope expansion appears.
- **Grep for tag and reflection references after any rename.**
  - gopls Rename only guards against _compilation_ breakage — it cannot see a struct tag, a `text/template` field reference, or a `reflect`-driven dispatch that still points at the old name.
  - Renaming a field silently desyncs it from its `json`/`db` tag.
- **Load the project-local `go-security` or `security-checklist` skill whenever a step changes code logic, not just its shape.**
  - A mechanical, tool-verified transform can't introduce a vulnerability, but a behavioral change can.
  - Treat "changes what the code does" as the trigger for a security-and-safety pass, not an afterthought reserved for the final review.
- **Record an immutable baseline and preserve unrelated dirty work.**
  - Capture the exact HEAD, status, relevant file hashes, staged/unstaged patch, and untracked paths required by the BE9 landing matrix before editing.
  - Use the designated path owner, central single writer, and path/symbol allowlist. If a step goes red, restore or reconstruct only that step from the recorded baseline; never reset or revert another user's existing changes.
  - Integrate a unit only after its exact candidate tree passes the required scoped gates.

## When Not to Refactor

Refactoring is an investment that only pays off if a future change is coming to spend it on. Question it — or skip it — when:

- **The code works and nothing planned will touch it again.**
  - A stable, rarely-read package earns nothing from being restructured for its own sake.
  - The risk of even a small staged refactor has to be repaid by an easier next change, and there may not be one.
- **It's critical production code with no tests.** Don't refactor it directly.
  - Add a characterization-test baseline and record it in the accepted plan before touching the critical path; treat that gate as non-negotiable.
- **The deadline is tight.**
  - A staged refactor needs review and verification bandwidth between landing units.
  - Starting one under time pressure either stalls or gets rushed, defeating the separation and verification discipline.
  - Make the minimal safe change now and stage the larger refactor for when there's room for it.
- **There's no clear purpose.**
  - "Refactor this" with no reason behind it — no upcoming feature it'll make easier, no bug class it'll close off, no smell a review actually flagged — is refactoring for its own sake.
  - Confirm the purpose while establishing scope rather than assuming one.

## Risk Stratification

| Risk | Transforms | Safety requirement |
| --- | --- | --- |
| **Low** | gopls Rename, Extract Variable/Constant, Inline Variable, `gofmt -s`, organize imports, local `refactor.rewrite.*` actions | Docker-scoped build/vet/test after the step is enough |
| **Medium** | Extract Function/Method (Extract is best-effort — verify comments/behavior survived), Inline Call across packages, single-parameter add/remove, introducing generics | Add or confirm targeted tests over the blast radius first |
| **High** | Change signature across many callers, moving types/functions across packages, splitting/merging packages, breaking import cycles, exported-API or major-version changes | Full scoped safety net + recorded plan and boundary evidence before landing |

**Diagnose:** 1- gopls refusing a Rename or Inline is a real semantic hazard, not a tool bug — investigate before forcing the change by hand 2- run `go vet`, project lint, test, and `-race` only through Docker and only for changed packages; any new failure blocks the unit 3- use `benchstat` for hot paths and stop on a meaningful delta 4- measure coverage on the touched package/path, while treating the project coverage ratchet as the quality gate 5- if only a prohibited full command can prove the result, give the exact command to the user for manual execution (see [safety-net.md](references/safety-net.md); its bare/full commands are community examples, not AnimalEkarte execution commands)

## Workflow: Plan → Stage → Land

- AnimalEkarte refactors land as the ordered, independently verifiable units defined by active BE9. Session A uses a clean local `main` as lane A and the integration tip; Session B uses one separate branch/worktree from the same immutable base. They may implement two mutually non-conflicting domain-local candidates concurrently. After A lands lane A plus its required central gate sync to `main`, B reconstructs and re-freezes its candidate on the new main tip; A then lands that exact candidate and applies B's central change request separately. Ephemeral read-only fixed-tree reviewer agents ensure neither lane self-approves; they are not additional implementation sessions.
- [workflow.md](references/workflow.md) is community reference material. Read only its inventory and ordering concepts; its mandatory sign-off cadence, refactoring branch, per-change PR, marker, worktree, and bare/full Go command model do not apply here:
  - the planning gate and refactoring inventory
  - the three interacting orderings (structural-before-behavioral, conflict-avoidance, dependency order)
  - when to run steps in parallel versus sequentially
  - how to keep each unit independently reviewable and reversible

## Detailed References

- **[workflow.md](references/workflow.md)** — use only the inventory, dependency ordering, and conflict-analysis concepts; AnimalEkarte's git/session/approval model comes from active BE9 and project autonomy rules.
- **[catalog.md](references/catalog.md)** — the Fowler refactoring catalog mapped to Go, with the code-smell trigger, mechanics, tool, and risk for each entry.
- **[go-tooling.md](references/go-tooling.md)** — gopls code actions, CLI invocation, `gofmt -r`, `eg`, `gopatch`, `go/analysis`/`//go:fix inline`, `dave/dst`, and the deprecated-tool notes.
- **[safety-net.md](references/safety-net.md)** — use the coverage-adaptive and characterization-testing concepts; translate every command to Docker-scoped project verification.
- **[structural.md](references/structural.md)** — use import-cycle and type-alias mechanics conditionally; ADR-006 and active BE9 own target boundaries and facade deletion rules.

## Cross-References

- Use the project naming rules and `naming-analyzer` when deciding what to rename identifiers to; this skill owns how to apply the rename safely at scale.
- Target directory/package layout is owned by the project Go/Gin guideline, ADR-006, and active BE9. External layout skills may contribute compatible move mechanics only.
- External modernization guidance is optional. Do not apply `interface{}`→`any`; AnimalEkarte prohibits introducing `any` in Go and TypeScript.
- Use the project Go/Gin guideline and `coding-standards` for control-flow clarity and function shape.
- Target DI and interface boundaries are owned by the project Go/Gin guideline and ADR-006; external design-pattern guidance is advisory only.
- Use the project-local `golang-testing` skill for test-writing practices that make the safety net trustworthy.
- Use the project-local `go-linting` skill for lint configuration; execute only Docker-scoped lint permitted by the project rules.
- Use the project-local `go-security` or `security-checklist` skill to review any step that changes code logic, not just its shape.

If you encounter a bug or unexpected behavior in `gopls`, prepare a local issue draft with a minimal reproducer. Post it to <https://github.com/golang/go/issues> only after explicit user approval.
