# Git / Worktree Safety (Agent Mandatory)

並行エージェントと AI セッションで未コミット変更が消える事故を防ぐための **必須** ルール。

## 絶対禁止（エージェントは実行しない・提案もしない）

次は hook / permission deny でもブロックされる。ユーザーが明示指示しても、まず代替案を出す。

| 禁止 | 理由 |
|------|------|
| `git reset --hard`（任意の ref 含む） | working tree + index を破棄 |
| `git clean -fd` / `-fdx` / `-df` | untracked を削除 |
| `git checkout -- .` / `git restore .` | 全変更を破棄 |
| `git push --force` / `-f` / `--force-with-lease` | 共有履歴の破壊（lease 付きもエージェント自動実行禁止） |

## 許可される代替

| 目的 | 使う |
|------|------|
| リモートに追随 | `git fetch` + `git merge` / `git pull --ff-only` |
| 特定ファイルだけ戻す | `git restore <path>`（パス限定・必要ならユーザー確認） |
| 作業の一時退避 | `git stash push -u -m "..."`（共有 tree では最終手段） |
| 履歴を触らず同期 | 新しい **worktree** で `origin/main` を checkout |

## 並行タスク = worktree 必須

- **同一 working tree で複数 Grok/Claude/Codex を並行させない。**
- 各タスクは次のいずれかで隔離する:
  1. `git worktree add ../AnimalEkarte-<task> -b <branch> origin/main`
  2. Grok subagent の `isolation: "worktree"`
  3. 専用 clone
- 共有 `main` tree では **1 セッションだけ**が編集してよい。
- 他セッションの未コミット変更がある tree で `checkout` / `switch` / `merge` / `pull` する前に `git status` を確認し、他人の WIP を潰さない。

## コミット運用

- 重要な途中結果は WIP commit してよい（`chore: wip <topic>`）。
- 長いタスクの途中で hard sync が必要なら、まず commit か worktree 分離。

## 事故後

- `git reflog` / `git fsck --lost-found` / エディタ Local History を案内する。
- `reset --hard` で消した変更を「もう一度同じコマンドで直す」ことはしない。
