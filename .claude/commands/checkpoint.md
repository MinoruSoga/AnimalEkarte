---
description: ワークフローのチェックポイント作成・一覧・確認
argument-hint: "create <name> | list | verify <name>"
---

# チェックポイント管理

$ARGUMENTS に従ってチェックポイントを操作してください。

## create \<name\>

1. `git status` で未コミット変更を確認
2. 変更があれば `git stash` または一時コミット（ユーザーに確認）
3. チェックポイントを `.claude/checkpoints.log` に追記:

```bash
echo "$(date +%Y-%m-%d-%H:%M) | $NAME | $(git rev-parse --short HEAD) | $(git branch --show-current)" >> .claude/checkpoints.log
```

4. 結果を報告: `CHECKPOINT CREATED: $NAME @ $(git rev-parse --short HEAD)`

## list

`.claude/checkpoints.log` を読んで一覧表示:

```
CHECKPOINTS
===========
DATE              NAME              SHA      BRANCH
2026-04-24-10:00  feature-start     abc1234  main
2026-04-24-12:30  core-done         def5678  main
```

## verify \<name\>

1. ログから指定チェックポイントを取得
2. 現在状態と比較して報告:

```
CHECKPOINT VERIFY: $NAME
========================
Git: $CHECKPOINT_SHA → $(git rev-parse --short HEAD) (N commits ahead)
Tests: [実行して結果を記載]
Files changed since checkpoint: N files
```

## 引数なしの場合

`list` と同じ動作。

## 注意

- Docker 環境での手動 build/test 実行は禁止（ユーザーに実行してもらう）
- チェックポイントログは `.claude/checkpoints.log` に平文で保存
- `git push` や `git stash` は実行前にユーザーへ確認を取ること
