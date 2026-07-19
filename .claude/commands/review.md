---
description: ステージング済み変更のコードレビュー（専門エージェント連携）
argument-hint: [--go | --fe | --db | --security | blank for auto]
---

# コードレビュー

変更ファイルを分析し、適切な専門エージェントでレビューを実行します。

**入力**: $ARGUMENTS

## Step 1: 変更ファイルの分類

```bash
git diff --name-only HEAD
git diff --staged --name-only
```

変更ファイルを以下に分類:
- `.go` ファイル → Go レビュー対象
- `.ts` / `.tsx` ファイル → TypeScript レビュー対象
- `.sql` / migration ファイル → DB レビュー対象
- セキュリティ関連変更（auth, permission, middleware）→ セキュリティレビュー対象

## Step 2: 専門エージェント選択

`$ARGUMENTS` がある場合はフラグに従う。なければ Step 1 の分類で自動決定:

| 変更内容 | 使用エージェント |
|---------|--------------|
| `.go` ファイル変更 | `go-reviewer` |
| `.ts` / `.tsx` 変更 | `typescript-reviewer` |
| `.tsx`/`.jsx` のフックロジック変更 | `react-reviewer`（typescript-reviewerと並列。フック正確性・a11y担当） |
| DB スキーマ・migration | `database-reviewer` |
| auth / permission / security | `security-analyst` |
| 動物患者記録・clinic_id 変更 | `healthcare-reviewer` |
| 全体俯瞰・複合変更 | `reviewer` (軽量) |

複数カテゴリに跨る場合は並列でエージェントを起動する。

## Step 3: Go/Gin backend確認（Go変更時）

`.go` ファイルに変更がある場合、`go-reviewer` は `.claude/refs/go-gin-backend-review.md` に従い、package API、Context、binding/validation、authn/authz/ownership、error、database/security、server lifecycle、tests を確認する。固定layer構成は判定基準にしない。

## Step 4: React 19 パターン確認（TS変更時）

`.ts` / `.tsx` に変更がある場合、`typescript-reviewer` が以下を確認:
- Feature Indexing (deep import 禁止)
- useActionState + SubmitButton パターン
- デザイントークン準拠 (C, STYLE 定数)
- `any` 型使用禁止

## 出力形式

```markdown
## コードレビュー結果

### 🔴 CRITICAL（マージブロック）
- ファイル:行 — 問題 + 修正例

### 🟠 HIGH（対応必須）
- ファイル:行 — 問題

### 🟡 MEDIUM（推奨対応）
- ファイル:行 — 改善提案

### 承認ステータス
[Approve / Warning / Block]
```
