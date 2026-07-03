---
name: implement-issue
description: docs/tasks/open/ のタスクID（FEAT-XXX / PERF-XXX / BUG-XXX / SEED-XXX 等）を指定して、コード規約準拠の実装 → セルフレビュー → タスククローズまでを自動化する。`/implement PERF-FOLLOWUP-01` のように使用。旧 BE-XXX / FE-XXX イシュー体系は docs/archive/ に移設済み（新規実装には使わない）。
---

# Implement Issue — タスク実装ワークフロー

タスクファイルを読み込み、コード規約に準拠した実装 → セルフレビュー → クローズ処理までを実行する。

> **パス正本の注意**: 旧 `backend/issues/` / `frontend/issues/` 体系は**廃止済み**。
> closed のみ `docs/archive/backend-issues/` / `docs/archive/frontend-issues/` に残存する。
> 現行のタスク管理は `docs/tasks/{open,closed,pending}/`（.gitignore 済み — タスクファイルの commit 提案はしない）。

## 起動トリガー

- `/implement <タスクID>`（例: `/implement PERF-FOLLOWUP-01`）— `docs/tasks/open/` からファイル名前方一致で検索
- 引数なしの場合: `docs/tasks/open/` を一覧表示し、ユーザーに選択させる
- 旧 `BE-XXX` / `FE-XXX` 番号を指定された場合: `docs/archive/*/closed/` を参照して経緯確認のみ行い、新規実装には使わない

引数は `$ARGUMENTS` 変数で受け取る。

---

## Phase 1: タスク選択・読み込み

### 1.1 引数解析

- タスクID（`FEAT-XXX`, `PERF-XXX`, `BUG-XXX`, `SEED-XXX` 等）→ `docs/tasks/open/<ID>*.md` を前方一致検索
- 引数なし → 以下を実行して一覧表示:

```bash
echo "=== Open Tasks ==="
ls docs/tasks/open/*.md 2>/dev/null | sort
echo ""
echo "=== Pending Tasks (着手保留) ==="
ls docs/tasks/pending/*.md 2>/dev/null | sort
```

ユーザーに番号またはファイル名を選択させる。

### 1.2 タスクファイル読み込み

タスクファイルを Read で読み込み、以下を抽出:
- **Summary**: 実装内容の概要
- **親タスク**: 親 TASK へのリンク（`**親タスク**: [TASK-XXX](...)` 形式）
- **Related**: 依存イシュー（`BE-XXX`, `FE-XXX`, `TASK-XXX`）
- **依存関係**: 「依存関係」セクションに記載された前提条件
- **完了条件**: チェックリスト項目
- **必要な変更**: 具体的なコード変更指示

### 1.3 依存関係チェック

- `Related` と「依存関係」セクションに記載された前提タスクが `docs/tasks/closed/` に存在するか確認
- 前提タスクが未完了（`open/` または `pending/` に残存）の場合:
  - ユーザーに警告: 「<前提ID> が未完了。先に実装するか？」

```bash
# 依存タスクのクローズ確認
ls docs/tasks/closed/*XXX*.md 2>/dev/null
# 旧イシュー体系（BE-XXX / FE-XXX）はアーカイブを確認
ls docs/archive/backend-issues/closed/*XXX*.md 2>/dev/null
ls docs/archive/frontend-issues/closed/*XXX*.md 2>/dev/null
```

---

## Phase 2: コンテキスト収集

### 2.1 対象ファイルの特定

イシュー内の「現状のコード」「必要な変更」セクションから、変更対象ファイルパスを抽出し、全て Read で読み込む。

### 2.2 参照実装の確認

**FE イシューの場合:**
- `features/owners/` の対応パターンを確認（ベストプラクティス参照実装）
- 変更内容に応じて、以下のファイルから該当パターンを読む:
  - フォーム系 → `features/owners/routes/OwnerForm.tsx` + `features/owners/hooks/use-owner-form.ts`
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
- DB マイグレーションがある場合: **適用済み migration の編集は禁止**（checksum mismatch → STG db_reset が必要になる。出典: memory ops_applied_migration_edit_requires_db_reset）。`backend/migrations/` の最終番号 +1 で新規ファイルを追加し、`migration-seed-safety` スキルのチェックリストに従う
- モデル変更がある場合: `make codegen` を実行して `models.ts` を更新（※ CLAUDE.md の自動実行禁止コマンド。ユーザーに実行を依頼する）

```bash
# モデル変更後の codegen
make codegen
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

### 4.4 Lint・ビルド・テスト実行（Docker 経由・スコープ限定）

**全体 lint / build / test（`pnpm lint` `pnpm build` `pnpm test:run` `golangci-lint run ./...` `go test ./...`）は CLAUDE.md の自動実行禁止コマンド。** 変更スコープに限定した検証を自分で実行し、全体検証はユーザーに手動実行を依頼する。

```bash
# FE の場合 — 変更した feature に限定（`--` 付き pnpm test は全件実行になる罠。npx vitest run <path> が正。出典: memory feedback_frontend_verify_harness_gotchas）
docker compose exec frontend npx vitest run src/features/<対象feature>

# BE の場合 — 変更したパッケージに限定
docker compose exec backend go build ./...
docker compose exec backend go vet ./internal/<対象パッケージ>/...
docker compose exec backend go test ./internal/<対象パッケージ>/...
```

全体検証が必要な場合の依頼文例:

```
変更完了。全体検証は以下を手動実行してください:
$ docker compose exec backend go test ./...
$ docker compose exec frontend pnpm lint && docker compose exec frontend pnpm test:run
```

### 4.5 問題があれば Phase 3 に戻る（最大3回）

Lint エラー・型エラー・テスト失敗・規約違反があれば修正して再レビュー。
**3回修正しても解決しない場合はユーザーに報告して判断を仰ぐ。**

---

## Phase 5: クローズ処理

### 5.1 タスクファイル更新

タスクファイルの先頭の Status を更新:

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
# タスクファイルを closed へ移動（docs/tasks は .gitignore 済み — commit 提案はしない）
mv docs/tasks/open/<タスクID>*.md docs/tasks/closed/
```

### 5.3 親 TASK ドキュメント更新（存在する場合）

イシューの `親タスク` フィールドまたは `Related` に `TASK-XXX` がある場合:
1. `docs/tasks/open/TASK-XXX-*.md` を読み込む
2. 「サブタスク分解」テーブルの該当行にチェックを入れる
3. 全サブタスクが完了していれば、TASK 自体もクローズ候補としてユーザーに通知

### 5.4 完了報告

以下のフォーマットでユーザーに報告:

```
## 実装完了: <タスクID>

### 変更ファイル
- `path/to/file1.tsx` — 変更内容
- `path/to/file2.ts` — 変更内容

### レビュー結果
- スコープ限定 vet/test: PASS（実行したコマンドと結果を明記）
- 完了条件: 全項目クリア
- 全体 lint/test: ユーザー手動実行待ち（コマンド提示済み）

### タスク
- <タスクID> → docs/tasks/closed/ に移動済み
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
