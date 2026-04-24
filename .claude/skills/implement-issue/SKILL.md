---
name: implement-issue
description: イシュー番号（BE-XXX / FE-XXX）を指定して、コード規約準拠の実装 → セルフレビュー → イシュークローズまでを自動化する。`/implement FE-038` のように使用。
---

# Implement Issue — イシュー実装ワークフロー

イシューファイルを読み込み、コード規約に準拠した実装 → セルフレビュー → クローズ処理までを実行する。

## 起動トリガー

- `/implement BE-XXX` または `/implement FE-XXX`（例: `/implement FE-038`）
- 旧形式番号も対応: `/implement 003`（`open/` 内からファイル名に `003` を含むものを検索）
- 引数なしの場合: `frontend/issues/open/` + `backend/issues/open/` を一覧表示し、ユーザーに選択させる

引数は `$ARGUMENTS` 変数で受け取る。

---

## Phase 1: イシュー選択・読み込み

### 1.1 引数解析

- `BE-XXX` → `backend/issues/open/BE-XXX-*.md` を検索
- `FE-XXX` → `frontend/issues/open/FE-XXX-*.md` を検索
- 旧形式（`003`, `master-002` 等プレフィックスなし）→ 両ディレクトリから `*003*` でパターン検索
- 引数なし → 以下を実行して一覧表示:

```bash
echo "=== Backend Issues (Open) ==="
ls backend/issues/open/*.md 2>/dev/null | sort
echo ""
echo "=== Frontend Issues (Open) ==="
ls frontend/issues/open/*.md 2>/dev/null | sort
```

ユーザーに番号またはファイル名を選択させる。

### 1.2 イシューファイル読み込み

イシューファイルを Read で読み込み、以下を抽出:
- **Summary**: 実装内容の概要
- **親タスク**: 親 TASK へのリンク（`**親タスク**: [TASK-XXX](...)` 形式）
- **Related**: 依存イシュー（`BE-XXX`, `FE-XXX`, `TASK-XXX`）
- **依存関係**: 「依存関係」セクションに記載された前提条件
- **完了条件**: チェックリスト項目
- **必要な変更**: 具体的なコード変更指示

### 1.3 依存関係チェック

- `Related` と「依存関係」セクションに記載された前提イシューが `closed/` に存在するか確認
- FE イシューが BE イシューに依存する場合:
  - `backend/issues/closed/` に対応する BE イシューがあるか確認
  - なければユーザーに警告: 「BE-XXX が未完了。先に実装するか？」

```bash
# 依存イシューのクローズ確認（新旧両形式に対応）
ls backend/issues/closed/BE-XXX-*.md 2>/dev/null
ls backend/issues/closed/*XXX*.md 2>/dev/null
ls frontend/issues/closed/FE-XXX-*.md 2>/dev/null
ls frontend/issues/closed/*XXX*.md 2>/dev/null
```

---

## Phase 2: コンテキスト収集

### 2.1 対象ファイルの特定

イシュー内の「現状のコード」「必要な変更」セクションから、変更対象ファイルパスを抽出し、全て Read で読み込む。

### 2.2 参照実装の確認

**FE イシューの場合:**
- `features/owners/` の対応パターンを確認（ベストプラクティス参照実装）
- 変更内容に応じて、以下のファイルから該当パターンを読む:
  - フォーム系 → `features/owners/routes/OwnerForm.tsx` + `features/owners/hooks/useOwnerForm.ts`
  - リスト系 → `features/owners/routes/OwnersList.tsx`
  - API hooks → `features/owners/api/` 内の対応ファイル
  - loader → `features/owners/loaders.ts`

**BE イシューの場合:**
- `backend/CLAUDE.md` のレイヤードパターンを確認
- 同じドメインの既存実装を参照（例: owners の handler/service/repository）

### 2.3 コーディングルール読み込み

**FE の場合:**
- `frontend/CODING_RULES.md` の Section 12（Vercel React Best Practices）を読み込む

**BE の場合:**
- `backend/CLAUDE.md` の実装パターンを読み込む

---

## Phase 3: 実装（コード規約準拠）

### 3.1 FE 実装チェックリスト（10パターン）

実装時に以下を全て確認・適用する:

| # | パターン | 適用条件 |
|---|---------|---------|
| 1 | `memo()` + `useCallback` でセクション分割 | 大型フォーム・リスト行 |
| 2 | `useDeferredValue` で検索フィルタ遅延 | フィルタ・検索入力 |
| 3 | `useTransition` で pending 管理 | API 書き込み（保存・削除） |
| 4 | `lazy()` + `Suspense` で遅延ロード | 重いモーダル・ダイアログ |
| 5 | 直接ファイル import（barrel 禁止） | 全 import |
| 6 | 三項演算子 `? ... : null`（`&&` 禁止） | 条件レンダー |
| 7 | `useState(() => ...)` lazy init | 高コストな初期化 |
| 8 | 静的 JSX はモジュール定数に巻き上げ | Select 選択肢、テーブルヘッダ等 |
| 9 | API 由来リストは `useMemo` でキャッシュ | API データの JSX 変換 |
| 10 | loader 内の独立フェッチは `Promise.all` | 複数 API 呼び出し |

**追加禁止チェック:**
- `any` 型 → `unknown` + 型ガード
- `FC` / `forwardRef` → 関数宣言 + ref as prop
- `useState(false)` + `setIsPending` → `useTransition`
- 型は `models.ts` から `Omit`/`Partial` で導出（手書き interface 禁止）
- `console.log` → 削除

### 3.2 BE 実装チェックリスト

| # | パターン | 適用条件 |
|---|---------|---------|
| 1 | 全関数に `ctx context.Context` 第一引数 | 全関数 |
| 2 | handler: `*_request.go` → `service.XxxInput` → `toXxxResponse()` | 新規/修正ハンドラ |
| 3 | service: HTTP を知らない（`binding:` タグ禁止、`*gin.Context` 禁止） | service 層 |
| 4 | PATCH: ポインタ型 + `buildXxxUpdateFields()` → `map[string]any` | PATCH エンドポイント |
| 5 | エラー: sentinel → `fmt.Errorf("...: %w", err)` → `RespondError(c, err)` | エラー処理 |
| 6 | slog は service 層のみ（handler・repository には書かない） | ログ |
| 7 | インターフェース最小化（3-5メソッド） | 新規インターフェース |

### 3.3 実装実行

- `implementer` エージェント（Sonnet）を使って実装を並列実行してよい
- DB マイグレーションがある場合: `backend/migrations/001_init.sql` を直接編集（リリース前運用）
- モデル変更がある場合: `make codegen` を実行して `models.ts` を更新

```bash
# モデル変更後の codegen
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && make codegen
```

---

## Phase 4: セルフレビュー

### 4.1 reviewer エージェント起動

`reviewer` エージェント（Haiku）を起動し、以下を検証させる。

### 4.2 完了条件チェック

イシューファイルの「完了条件」チェックリスト項目を1つずつ検証。

### 4.3 コード規約チェック

**FE の場合:**
- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] `&&` 条件レンダーなし（三項演算子を使用）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` なし）
- [ ] 型は `models.ts` から導出
- [ ] `console.log` なし
- [ ] feature 間 import なし

**BE の場合:**
- [ ] 全関数に `ctx context.Context`
- [ ] エラーは `fmt.Errorf("...: %w", err)` でラップ
- [ ] slog は service 層のみ
- [ ] PATCH はポインタ型 + `buildXxxUpdateFields()`
- [ ] service に `*gin.Context` / `binding:` タグなし

### 4.4 Lint・ビルド・テスト実行（Docker 経由）

```bash
# FE の場合（3段階: lint → 型チェック → テスト）
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec frontend ppnpm lint
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec frontend ppnpm build
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec frontend pppnpm test:run

# BE の場合（3段階: lint → vet/build → テスト）
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec backend golangci-lint run ./...
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec backend go vet ./...
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec backend go build ./...
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte && docker compose exec backend go test ./... -v
```

### 4.5 問題があれば Phase 3 に戻る（最大3回）

Lint エラー・型エラー・テスト失敗・規約違反があれば修正して再レビュー。
**3回修正しても解決しない場合はユーザーに報告して判断を仰ぐ。**

---

## Phase 5: クローズ処理

### 5.1 イシューファイル更新

イシューファイルの先頭に YAML frontmatter を更新:

```
**Status**: Closed
```

`closed_at` と実装コミット情報を末尾に追記:

```markdown
## クローズ情報

- **Closed At**: YYYY-MM-DD
- **変更ファイル**: （変更したファイルの一覧）
```

### 5.2 ファイル移動

```bash
# closed/ ディレクトリが存在することを確認
mkdir -p backend/issues/closed frontend/issues/closed

# BE イシューの場合
mv backend/issues/open/BE-XXX-*.md backend/issues/closed/

# FE イシューの場合
mv frontend/issues/open/FE-XXX-*.md frontend/issues/closed/

# 旧形式イシューの場合（BE-/FE- プレフィックスなし）
# mv [backend|frontend]/issues/open/XXX-*.md [backend|frontend]/issues/closed/
```

### 5.3 親 TASK ドキュメント更新（存在する場合）

イシューの `親タスク` フィールドまたは `Related` に `TASK-XXX` がある場合:
1. `docs/tasks/open/TASK-XXX-*.md` を読み込む
2. 「サブタスク分解」テーブルの該当行にチェックを入れる
3. 全サブタスクが完了していれば、TASK 自体もクローズ候補としてユーザーに通知

### 5.4 完了報告

以下のフォーマットでユーザーに報告:

```
## 実装完了: [BE/FE]-XXX

### 変更ファイル
- `path/to/file1.tsx` — 変更内容
- `path/to/file2.ts` — 変更内容

### レビュー結果
- Lint: PASS
- Build: PASS
- 完了条件: 全項目クリア

### イシュー
- [BE/FE]-XXX → closed/ に移動済み
```

---

## エラーハンドリング

| 状況 | 対応 |
|------|------|
| イシューファイルが見つからない | ユーザーに番号の確認を求める |
| 依存イシューが未完了 | 警告表示、ユーザーに続行確認 |
| Docker コンテナ未起動 | `make up` の実行を提案 |
| Lint/Build 失敗 | エラー内容を表示し、Phase 3 に戻って修正 |
| codegen 失敗 | Go モデルのコンパイルエラーを確認 |

---

## 禁止事項

- **イシューに書かれていない変更を勝手に行わない**: スコープはイシューの「必要な変更」に限定
- **UI を推測で実装しない**: Figma デザインがない場合、UI 変更はイシューの指示に厳密に従う
- **ローカルで npm/go コマンドを実行しない**: 必ず Docker 経由
- **テストを省略しない**: イシューの完了条件にテストがあれば必ず実行
- **イシューの完了条件を勝手に変更しない**: 条件を満たせない場合はユーザーに報告
