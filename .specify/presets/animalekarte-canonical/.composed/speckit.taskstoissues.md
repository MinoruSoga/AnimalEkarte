---
description: Convert tasks to issues only with explicit external-write authorization.
scripts:
  sh: scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks
  ps: scripts/powershell/check-prerequisites.ps1 -Json -RequireTasks -IncludeTasks
  py: scripts/python/check_prerequisites.py --json --require-tasks --include-tasks
---


## AnimalEkarte プロジェクト適用（ローカルpreset）

これは公式Spec Kitの拡張点を使うプロジェクト固有の指示。業務仕様の正本は docs/spec/ のドメイン・lifecycle・workflow、文書規約は docs/spec/authoring-guide.md。最初に AGENTS.md と関連する既存規約を読む。既存の臨床安全・医院分離・Git安全・Docker限定検証規則を保持する。

spec.md は変更の依頼・範囲・正本への参照索引であり、業務ルールの第二正本にしない。plan/tasks は変更単位の計画で、正本FR-ID・GWTへリンクする。spec.mdのリンク先を読むことが前提。要件IDは安定したドメイン接頭辞を使い、決定状態と実装・実機検証状態を分ける。未知の臨床・会計・権限・入力・優先度を推測で補完しない。未決事項の件数上限を設けず、質問だけ小分けにする。

featureディレクトリは明示する。通常は SPECIFY_FEATURE_DIRECTORY=docs/work/spec-kit/<意味を表す変更名> とする。未指定なら対象正本と変更範囲を確認してから確定し、既存ファイルを上書きしない。今回の導入では実予約のfeatureは作らない。current mainを保持し、Git拡張・自動branch・commit・push・stash・reset・force initを使わない。

スキルの呼出だけで機能実装・外部送信・デプロイ・フルCIを許可されたと解釈しない。品質チェック済みは要件の記述品質を示し、実機PASSやPlane Doneを意味しない。constitution.md は既存正本への参照口であり、新しい独立した規範を作らない。元の未決事項を解消できないときは保持して、依存する実装・確定を止める。




## Deprecation Notice

You **MUST display the following concise warning first**, as the first step of this command and before the User Input and Pre-Execution Checks sections below:

> ⚠️ `__SPECKIT_COMMAND_TASKSTOISSUES__` is deprecated and will be removed in a future minor release. Use the bundled `github` extension instead: run `specify extension add github`, then invoke `__SPECKIT_COMMAND_GITHUB_TASKSTOISSUES__`.

Continue with the existing workflow unchanged. Do not install or enable the extension automatically, do not stop, and do not run the replacement command.

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Pre-Execution Checks

**Check for extension hooks (before tasks-to-issues conversion)**:
- Check if `.specify/extensions.yml` exists in the project root.
- If it exists, read it and look for entries under the `hooks.before_taskstoissues` key
- If the YAML cannot be parsed or is invalid, do not skip silently: tell the user that `.specify/extensions.yml` could not be read (include the parser error) and that no hooks were checked, including any mandatory (`optional: false`) hooks registered there, then continue normally
- Filter out hooks where `enabled` is explicitly `false`. Treat hooks without an `enabled` field as enabled by default.
- For each remaining hook, do **not** attempt to interpret or evaluate hook `condition` expressions:
  - If the hook has no `condition` field, or it is null/empty, treat the hook as executable
  - If the hook defines a non-empty `condition`, skip the hook and leave condition evaluation to the HookExecutor implementation
- For each executable hook, output the following based on its `optional` flag:
  - **Optional hook** (`optional: true`):
    ```
    ## Extension Hooks

    **Optional Pre-Hook**: {extension}
    Command: `/{command}`
    Description: {description}

    Prompt: {prompt}
    To execute: `/{command}`
    ```
  - **Mandatory hook** (`optional: false`):
    ```
    ## Extension Hooks

    **Automatic Pre-Hook**: {extension}
    Executing: `/{command}`
    EXECUTE_COMMAND: {command}

    Wait for the result of the hook command before proceeding to the Outline.
    ```
    After emitting the block above you MUST actually invoke the hook and wait for it to finish before continuing. Run it the same way you would run the command yourself in this agent/session (the invocation may differ from the literal `{command}` id shown above, e.g. a skills-mode agent runs it as `/skill:speckit-...` or `$speckit-...`). Emitting the block alone does not run the hook.
- If no hooks are registered or `.specify/extensions.yml` does not exist, skip silently

## Outline

1. Run `{SCRIPT}` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute. For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").
1. **IF EXISTS**: Load `/memory/constitution.md` for project principles and governance constraints.
1. From the executed script, extract the path to **tasks**.
1. Get the Git remote by running:

```bash
git config --get remote.origin.url
```

> [!CAUTION]
> ONLY PROCEED TO NEXT STEPS IF THE REMOTE IS A GITHUB URL

1. **Fetch existing issues for deduplication**: Before creating anything, build the set of task IDs you are about to process from `tasks.md` (each is a `T` followed by **at least** three digits, e.g. `T001` — `__SPECKIT_COMMAND_CONVERGE__` assigns new IDs with `T{M+1:03d}`, which is a floor rather than a cap, so once a file has more than 999 tasks the IDs are four digits or longer). Then use the GitHub MCP server's `list_issues` tool to look for issues that already cover those IDs. Do not pass a `state` value, since omitting it makes the tool return both open and closed issues. Request `perPage: 100` to keep the number of calls down, and since the tool uses cursor-based pagination, request pages with the `after` parameter (using the `endCursor` from the previous response). For each issue title, match it against the task ID pattern `\bT\d{3,}\b` (the `{3,}` accepts four-digit and longer IDs — with `\d{3}` a title containing `T1000` would not match at all, because the trailing `\b` cannot fall between two digits, so that task would be silently neither deduplicated nor created; word boundaries still stop a token like `ST001` from matching, and force the whole digit run to be consumed so `T100` can never match inside `T1000`; this also recognises titles written as `T001 ...`, `T001: ...` or `[T001] ...`) and, when it matches one of your task IDs, mark that ID as already having an issue. Stop paginating as soon as every task ID has been matched, or when there are no more pages, so you do not keep fetching the whole repository's issue history once all task IDs are accounted for. This bounds the number of calls on repos with large issue histories and still prevents duplicates when the command is re-run after `tasks.md` is regenerated or the skill is re-invoked.
1. For each task in the list, use the GitHub MCP server to create a new issue in the repository that is representative of the Git remote. Task lines in `tasks.md` start with a markdown checkbox, so first strip the leading `- [ ]` (and any `[P]` / `[US#]` markers) to recover the task ID and its description. Create the issue with a single canonical title of the form `T001: <description>`, with the ID written once followed by the task description (for example, the line `- [ ] T001 Create project structure` becomes the title `T001: Create project structure`).
   - **Skip** any task whose ID is already present in the set of existing issues from the previous step, and report it (for example, `T001 already has an issue, skipping`).
   - Only create issues for tasks that do not yet have a matching issue.

> [!CAUTION]
> UNDER NO CIRCUMSTANCES EVER CREATE ISSUES IN REPOSITORIES THAT DO NOT MATCH THE REMOTE URL

## Post-Execution Checks

**Check for extension hooks (after tasks-to-issues conversion)**:
Check if `.specify/extensions.yml` exists in the project root.
- If it exists, read it and look for entries under the `hooks.after_taskstoissues` key
- If the YAML cannot be parsed or is invalid, do not skip silently: tell the user that `.specify/extensions.yml` could not be read (include the parser error) and that no hooks were checked, including any mandatory (`optional: false`) hooks registered there, then continue normally
- Filter out hooks where `enabled` is explicitly `false`. Treat hooks without an `enabled` field as enabled by default.
- For each remaining hook, do **not** attempt to interpret or evaluate hook `condition` expressions:
  - If the hook has no `condition` field, or it is null/empty, treat the hook as executable
  - If the hook defines a non-empty `condition`, skip the hook and leave condition evaluation to the HookExecutor implementation
- For each executable hook, output the following based on its `optional` flag:
  - **Optional hook** (`optional: true`):
    ```
    ## Extension Hooks

    **Optional Hook**: {extension}
    Command: `/{command}`
    Description: {description}

    Prompt: {prompt}
    To execute: `/{command}`
    ```
  - **Mandatory hook** (`optional: false`):
    ```
    ## Extension Hooks

    **Automatic Hook**: {extension}
    Executing: `/{command}`
    EXECUTE_COMMAND: {command}
    ```
    After emitting the block above you MUST actually invoke the hook and wait for it to finish before continuing. Run it the same way you would run the command yourself in this agent/session (the invocation may differ from the literal `{command}` id shown above, e.g. a skills-mode agent runs it as `/skill:speckit-...` or `$speckit-...`). Emitting the block alone does not run the hook.
- If no hooks are registered or `.specify/extensions.yml` does not exist, skip silently
