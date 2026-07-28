---
description: GitHub PR を専門エージェントでレビューし gh pr review で投稿
argument-hint: "<pr-number | pr-url> [--focus=security|go|fe|db]"
---

# PR レビュー

**入力**: $ARGUMENTS

## Step 1: PR 情報取得

```bash
gh pr view <NUMBER> --json number,title,body,author,baseRefName,headRefName,changedFiles,additions,deletions
gh pr diff <NUMBER>
```

PR 番号が未指定の場合は現在ブランチの PR を使用:
```bash
gh pr view --json number,title,body
```

## Step 2: 変更ファイル分類

`gh pr diff --name-only` の出力から:
- `.go` → Go レビュー対象
- `.ts` / `.tsx` → TypeScript レビュー対象
- `.sql` / migrations → DB レビュー対象
- `auth/`, `permission/`, `middleware/` → セキュリティレビュー対象
- 動物患者記録・`clinic_id` 関連 → healthcare-reviewer 対象

## Step 3: 専門エージェント並列起動

`--focus` フラグがある場合は対象エージェントのみ。なければ変更内容に応じて自動選択:

| 変更内容 | 使用エージェント |
|---------|--------------|
| `.go` ファイル | `go-reviewer` |
| `.ts` / `.tsx` | `typescript-reviewer` |
| `.tsx`/`.jsx` フックロジック | `react-reviewer`（typescript-reviewerと並列） |
| DB / migration | `database-reviewer` |
| auth / security | `security-analyst` |
| 患者記録・clinic_id | `healthcare-reviewer` |
| デッドコード・重複 | `code-simplifier` |
| 暗黙的エラー隠蔽 | `silent-failure-hunter` |

各エージェントに渡す情報:
- PR diff の該当ファイル部分
- 変更前後のフルファイルコンテキスト
- PR タイトル・説明（実装意図の確認）

## Step 4: 結果統合

- 重複する指摘を除外
- 深刻度でソート: CRITICAL → HIGH → MEDIUM → LOW
- 各指摘にファイル:行番号を付与

## Step 5: GitHub へ投稿

```bash
# CRITICAL/HIGH がある場合
gh pr review <NUMBER> --request-changes --body "<summary>"

# MEDIUM/LOW のみの場合
gh pr review <NUMBER> --comment --body "<summary>"

# 問題なし
gh pr review <NUMBER> --approve --body "<summary>"
```

## 判定基準

| 状態 | 決定 |
|------|------|
| CRITICAL 指摘あり（セキュリティ・データ消失リスク）| **BLOCK** |
| HIGH 指摘あり（バグ・ロジックエラー）| **REQUEST CHANGES** |
| MEDIUM/LOW のみ | **COMMENT** |
| CRITICAL/HIGHなし、Go/Gin reviewおよびapplication invariantに適合 | **APPROVE** |

## 出力形式

```markdown
PR #<NUMBER>: <TITLE>
決定: APPROVE / REQUEST CHANGES / BLOCK

CRITICAL: N件 | HIGH: N件 | MEDIUM: N件 | LOW: N件

### CRITICAL
- `path/to/file:行` — 問題内容

### HIGH
- `path/to/file:行` — 問題内容

### MEDIUM / LOW
（省略）
```

## 注意

- `gh` CLI が未認証の場合はローカル結果のみ報告（投稿スキップ）
- Draft PR は **COMMENT** のみ（Approve/Block 禁止）
- 変更ファイル > 50 の場合: ソースファイルを優先し、設定・ドキュメントは後回し
