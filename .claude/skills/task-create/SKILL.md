---
name: task-create
description: 抽象的なタスク依頼から docs/tasks/ にタスク詳細ドキュメントを生成し、backend/issues/open/ と frontend/issues/open/ にAIが実装可能な粒度のイシューを自動作成する。「タスク分解」「イシュー作成」「チケット作成」時に使用。
---

# Task Decompose — タスク分解・イシュー自動生成

抽象的なタスク依頼を受け取り、コードベースを調査した上で:
1. `docs/tasks/open/TASK-XXX.md` にタスク詳細ドキュメントを生成
2. `backend/issues/open/BE-XXX-*.md` にバックエンドイシューを生成
3. `frontend/issues/open/FE-XXX-*.md` にフロントエンドイシューを生成

## 起動トリガー

ユーザーが以下のようなタスク依頼を行った場合に起動:
- `/task-decompose <タスク依頼文>`
- 「タスク分解して」「イシュー作って」「チケット切って」

---

## Phase 1: 採番（自動）

### 1.1 TASK番号の決定

```bash
# docs/tasks/ 内の最大番号 + 1（open/ と closed/ 両方から）
ls docs/tasks/{open,closed}/TASK-*.md 2>/dev/null | sort -t- -k2 -n | tail -1
```

形式: `TASK-XXX`（3桁ゼロ埋め）

### 1.2 BE/FE番号の決定

```bash
# backend: open/ と closed/ 両方から最大番号を取得
ls backend/issues/{open,closed}/BE-*.md 2>/dev/null | grep -oE 'BE-[0-9]+' | sort -t- -k2 -n | tail -1

# frontend: open/ と closed/ 両方から最大番号を取得
ls frontend/issues/{open,closed}/FE-*.md 2>/dev/null | grep -oE 'FE-[0-9]+' | sort -t- -k2 -n | tail -1
```

形式: `BE-XXX`, `FE-XXX`（3桁ゼロ埋め）

---

## Phase 2: コードベース調査（必須）

タスク依頼文から影響範囲を特定するため、以下を調査する。**推測で書かない。必ずコードを読む。**

### 2.1 DB スキーマ調査

```bash
# 関連テーブル・カラムの確認
grep -n "関連キーワード" backend/migrations/001_init.sql
```

### 2.2 Backend モデル・API 調査

```bash
# Go モデル
grep -rn "関連キーワード" backend/internal/model/
# ハンドラ（エンドポイント）
grep -rn "関連キーワード" backend/internal/handler/
# サービス
grep -rn "関連キーワード" backend/internal/service/
# リポジトリ
grep -rn "関連キーワード" backend/internal/repository/
# ルーティング
grep -n "関連パス" backend/cmd/api/main.go
```

### 2.3 Frontend コンポーネント・API 調査

```bash
# 関連 feature
ls frontend/src/features/関連feature/
# API hooks
grep -rn "関連キーワード" frontend/src/features/*/api/
# コンポーネント
grep -rn "関連キーワード" frontend/src/features/*/components/
grep -rn "関連キーワード" frontend/src/features/*/routes/
# 型定義
grep -n "関連キーワード" frontend/src/types/generated/models.ts
```

### 2.4 調査結果の整理

調査結果を以下のカテゴリに分類:
- **DB変更が必要**: テーブル追加・カラム追加・マイグレーション
- **Backendモデル変更**: model/*.go の修正 → `make codegen` 必要
- **Backend API変更**: 新規エンドポイント or 既存修正
- **Frontend API変更**: api/ の追加・修正
- **Frontend UI変更**: components/ or routes/ の修正
- **変更不要**: 既に実装済みの部分

### 2.5 類似機能の特定

タスクと似た既存実装を特定し、参照実装として記録する。
- 同じパターンの feature はどれか（例: CRUD なら `owners/` を参照）
- 同じ UI パターンの画面はどれか
- 同じ DB パターンのテーブルはどれか

---

## Phase 2.5: 仕様確認（推測禁止・質問必須）

**このフェーズは Phase 2 の調査完了後、ドキュメント生成（Phase 3）の前に必ず実行する。**

### 目的

タスク依頼文とコードベース調査から判断できない曖昧な点を洗い出し、**推測せずにユーザーに質問する**。

### 2.5.1 曖昧さチェックリスト

以下の観点で、タスク依頼文に不足・曖昧さがないか検証する:

| カテゴリ | チェック項目 |
|---------|------------|
| **UI/UX 仕様** | レイアウト・配置が明確か？ Figma があるか？ 既存画面を参考にするか？ |
| **ビジネスルール** | 権限・条件分岐・バリデーションルールが明確か？ |
| **エッジケース** | 空データ・上限値・エラー時の挙動が定義されているか？ |
| **データフロー** | どこからデータを取得し、どの形式で保存するか明確か？ |
| **既存機能への影響** | 変更が他の画面・API に副作用を与えないか？ |
| **優先度・スコープ** | must-have と nice-to-have の線引きが明確か？ |
| **命名・用語** | UI ラベル・API パラメータ名が確定しているか？ |

### 2.5.2 質問の提示

曖昧な点が **1つでもあれば**、以下のフォーマットでユーザーに質問する:

```
## 仕様確認（回答をお願いします）

調査の結果、以下の点が不明です。イシュー生成前に確認させてください。

### 必須（これがないと実装方針が決まらない）
1. [質問内容]
2. [質問内容]

### 任意（デフォルト値で進めてもよい場合はスキップ可）
3. [質問内容] — デフォルト: [推定値]
4. [質問内容] — デフォルト: [推定値]
```

**ルール:**
- 「必須」の質問がある場合は **ユーザーの回答を待ってから Phase 3 に進む**
- 「任意」の質問のみの場合は、デフォルト値を提示し、ユーザーが「それでOK」と言えば進める
- **曖昧な点が一切なければ、このフェーズはスキップして Phase 3 に進む**（無理に質問を作らない）

### 2.5.3 回答の記録

ユーザーの回答は TASK ドキュメントの「仕様確認ログ」セクションに記録する（Phase 3 で使用）。

---

## Phase 3: タスク詳細ドキュメント生成

### 出力先: `docs/tasks/open/TASK-XXX-kebab-case-title.md`

### テンプレート

```markdown
# TASK-XXX: タスクタイトル

**作成日**: YYYY-MM-DD
**ステータス**: Open
**依頼元**: （タスク依頼の原文をそのまま引用）

---

## 概要

タスクの目的と背景を1-3行で説明。

## 依頼内容（原文）

> ここにユーザーのタスク依頼をそのまま引用

## 仕様確認ログ

Phase 2.5 でユーザーに質問し、得られた回答を記録する。
曖昧な点がなかった場合は「確認事項なし」と記載。

| # | 質問 | 回答 |
|---|------|------|
| 1 | 質問内容 | ユーザーの回答 |
| 2 | 質問内容 | ユーザーの回答 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | 説明 | BE/FE/DB | BE-XXX / FE-XXX | - | [ ] |
| 2 | 説明 | BE/FE/DB | BE-XXX / FE-XXX | #1 | [ ] |
| 3 | 説明 | BE/FE/DB | BE-XXX / FE-XXX | #1 | [ ] |

## 受入条件（Acceptance Criteria）

ユーザーが「完了」と判断するための具体的・検証可能な条件。
「〜が動く」ではなく「〜の画面で〜を入力し、〜が表示される」レベルで記述する。

- [ ] AC-1: 具体的なシナリオ（Given/When/Then 形式推奨）
- [ ] AC-2: 具体的なシナリオ
- [ ] AC-3: エッジケース（空データ、エラー時、上限値）

## 技術的判断

実装方針で複数の選択肢がある場合、選んだ方針とその理由を記録する。
判断が不要な単純タスクでは「特になし」と記載。

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 例: 状態管理 | useTransition | プロジェクト標準 | useState + try/finally |

## 影響範囲

### DB
- テーブル: `xxx` — 変更内容

### Backend
- `backend/internal/model/xxx.go` — 変更内容
- `backend/internal/handler/xxx_handler.go` — 変更内容
- `backend/internal/service/xxx_service.go` — 変更内容
- `backend/internal/repository/xxx_repository.go` — 変更内容

### Frontend
- `frontend/src/features/xxx/` — 変更内容
- `frontend/src/types/generated/models.ts` — codegen で自動更新

## 参照実装

このタスクと類似するパターンの既存実装。実装時に参考にすべきファイル。

- `features/owners/` — （どのパターンを参照するか具体的に記載）

## リスク・懸念事項

既知のリスクや注意点。なければ「特になし」と記載。

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 例: 既存画面のレイアウト崩れ | 中 | 変更後に全リスト画面を目視確認 |

## 未解決事項

Phase 2.5 で解消できなかった、または実装中に判明する可能性がある事項。
なければ「なし」と記載。実装中に判明した場合は随時追記する。

- [ ] 未解決事項1
- [ ] 未解決事項2

## 実装順序

1. DB マイグレーション（必要な場合）
2. Backend モデル → `make codegen`
3. Backend API（handler → service → repository）
4. Frontend API hooks
5. Frontend UI

## 関連イシュー

- BE-XXX: [タイトル](../../backend/issues/open/BE-XXX-*.md)
- FE-XXX: [タイトル](../../frontend/issues/open/FE-XXX-*.md)
```

---

## Phase 4: Backend イシュー生成

### 出力先: `backend/issues/open/BE-XXX-kebab-case-title.md`

### テンプレート

```markdown
# BE-XXX: イシュータイトル

**Status**: Open
**Priority**: High / Medium / Low
**Affects**: 影響する機能・コンポーネント
**Date Created**: YYYY-MM-DD
**Related**: TASK-XXX, FE-XXX（関連イシュー）

## Summary

1-2行で問題・実装内容を説明。

## 現状のコード

**実際のコードを読んで** 現在の実装を記載（推測禁止）。

```go
// backend/internal/model/xxx.go:行番号
// 現在のコード（関連部分のみ抜粋）
```

## 必要な変更

### 1. DB マイグレーション（該当する場合）

```sql
-- backend/migrations/001_init.sql に追記
ALTER TABLE xxx ADD COLUMN yyy TYPE;
```

### 2. Model 変更

```go
// backend/internal/model/xxx.go
// Before → After のコード差分
```

### 3. Repository 変更

```go
// backend/internal/repository/xxx_repository.go
// 追加・修正するメソッド
```

### 4. Service 変更

```go
// backend/internal/service/xxx_service.go
// 追加・修正するメソッド
```

### 5. Handler 変更

```go
// backend/internal/handler/xxx_handler.go
// 追加・修正するメソッド
```

### 6. Request/Response 変更（該当する場合）

```go
// backend/internal/handler/xxx_request.go
// backend/internal/handler/xxx_response.go
```

## API レスポンス形式（該当する場合）

```json
{
  "data": { ... }
}
```

## フロントエンド影響

- `make codegen` で `models.ts` が更新される
- FE-XXX で対応が必要

## 完了条件

- [ ] DB マイグレーション適用
- [ ] モデル変更 + `make codegen`
- [ ] 3層（handler → service → repository）実装
- [ ] 既存テストが通る
- [ ] API レスポンスが期待通り
```

---

## Phase 5: Frontend イシュー生成

### 出力先: `frontend/issues/open/FE-XXX-kebab-case-title.md`

### テンプレート

```markdown
# FE-XXX: イシュータイトル

**Status**: Open
**Priority**: High / Medium / Low
**Affects**: 影響する機能・コンポーネント
**Date Created**: YYYY-MM-DD
**Related**: TASK-XXX, BE-XXX（関連イシュー）

## Summary

1-2行で問題・実装内容を説明。

## 現状のコード

**実際のコードを読んで** 現在の実装を記載（推測禁止）。

```typescript
// frontend/src/features/xxx/yyy.tsx:行番号
// 現在のコード（関連部分のみ抜粋）
```

## 必要な変更

### 1. 型定義（該当する場合）

```typescript
// frontend/src/features/xxx/api/types.ts
// models.ts からの導出型を追加・修正
```

### 2. API hooks（該当する場合）

```typescript
// frontend/src/features/xxx/api/get-xxx.ts or create-xxx.ts
// 追加・修正する API 関数・hook
```

### 3. コンポーネント変更

```typescript
// frontend/src/features/xxx/components/XxxComponent.tsx
// or frontend/src/features/xxx/routes/XxxPage.tsx
// 追加・修正する UI（Before → After の差分）
```

### 4. hooks 変更（該当する場合）

```typescript
// frontend/src/features/xxx/hooks/useXxxForm.ts
// フォーム状態・バリデーション等の変更
```

## UI 操作フロー

1. ユーザーが「〜」画面を開く
2. 「〜」ボタンをクリック
3. 〜が表示される
4. ...

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）

## 依存関係

- BE-XXX が先に完了している必要がある（API エンドポイントが必要）
- `make codegen` で `models.ts` が更新されている必要がある

## 完了条件

- [ ] 型エラーなし（`ppnpm build` パス）
- [ ] ESLint エラーなし（`ppnpm lint` パス）
- [ ] UI が期待通りに動作
- [ ] 既存機能に影響なし
```

---

## Phase 6: ユーザーへの報告

全ファイル生成後、以下のサマリーを出力:

```
## タスク分解完了

### タスクドキュメント
- docs/tasks/TASK-XXX-title.md

### Backend イシュー (N件)
- BE-XXX: タイトル — 概要
- BE-YYY: タイトル — 概要

### Frontend イシュー (N件)
- FE-XXX: タイトル — 概要
- FE-YYY: タイトル — 概要

### 実装順序
1. BE-XXX → BE-YYY（DB + API）
2. FE-XXX → FE-YYY（UI）

### 依存関係
- FE-XXX は BE-XXX 完了後に着手可能
```

---

## 粒度の基準: 「AIが実装可能」

各イシューは以下の条件を満たす粒度で切る:

1. **単一責務**: 1イシュー = 1つの明確な変更目的
2. **自己完結**: イシュー内の情報だけで実装を開始できる（外部コンテキスト不要）
3. **コード参照付き**: 変更対象のファイルパス・行番号・現在のコードを記載
4. **Before/After明確**: 何を何に変えるかが具体的
5. **完了条件が検証可能**: 「〜が動く」ではなく「〜のレスポンスに〜が含まれる」
6. **1-2時間で完了**: 人間なら1-2時間、AIなら1セッションで完了できる規模

### 分割の判断基準

| 分割する | 分割しない |
|---------|-----------|
| DB変更 + API実装 → 別イシュー | 同一ファイル内の小修正 |
| 新規エンドポイント追加 | 既存エンドポイントの軽微修正 |
| 新規コンポーネント作成 | 既存コンポーネントのprops追加 |
| feature間で独立した変更 | 同一featureの関連変更 |

### BE イシューの粒度

- **DB + Model 変更**: 1イシュー（`make codegen` まで含む）
- **新規 API エンドポイント**: 1イシュー（handler + service + repository の3層セット）
- **既存 API 修正**: 影響範囲が明確なら1イシュー

### FE イシューの粒度

- **新規 API hook + 型定義**: 1イシュー
- **UI コンポーネント変更**: 画面単位で1イシュー
- **新規共有コンポーネント**: 1イシュー

---

## 禁止事項

- **推測でコードを書かない**: 必ずコードベースを読んでから記載
- **仕様を推測で埋めない**: 不明な点は Phase 2.5 でユーザーに質問する。回答を得てから生成に進む
- **存在しないファイルを参照しない**: Glob/Grep で実在を確認
- **イシューに「調査が必要」と書かない**: 調査は Phase 2 で完了させる
- **1イシューに複数の独立した変更を詰め込まない**
- **UI の見た目を推測で記載しない**: Figma デザインがない場合は「UI仕様は別途確認」と明記
- **受入条件を曖昧にしない**: 「〜が動く」ではなく「〜で〜を入力し〜が表示される」レベルで記述する
- **質問を省略しない**: 曖昧な点があるのに質問せず進めることは禁止。ただし曖昧な点がない場合に無理に質問を作ることも禁止
