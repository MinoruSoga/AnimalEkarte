---
name: implement-issue
description: "Linear Issue を指定して、コード規約準拠の実装 → セルフレビュー → Linear の証跡更新までを行う。`/implement BRT-123` のように使用。実行 SoT は Linear、今期外・見送り判断は docs/work/phase2-deferred.md。"
---

# Implement Issue — タスク実装ワークフロー

Linear の該当 Issue を読み込み、コード規約に準拠した実装 → セルフレビュー → 証跡更新までを実行する。

> **SoT の注意**: 実行タスク、仕様、受け入れ条件、状態、依存関係は Linear Issue が正本。今期外・見送りの索引は [`docs/work/phase2-deferred.md`](../../../docs/work/phase2-deferred.md) であり、実装台帳ではない。削除済み root ledger や旧 HTML を現在の依存関係・完了状態の根拠に使わない。
>
> GitHub Issue（`#NNN`）が Linear Issue から参照される場合は、仕様・受け入れ条件について該当 Issue 本文とコメント（`gh issue view <NNN>`）も確認する。

## 起動トリガー

- `/implement <Linear ID>`（例: `/implement BRT-123`）— Linear で Issue 本文・状態・コメントを確認する
- 引数なしの場合: `docs/work/phase2-deferred.md` は見送り索引であることを示し、実装対象の Linear ID をユーザーに求める
- 旧 `TASK-XXX` / `BUG-XXX` / `BE-XXX` / `FE-XXX` を指定された場合: git 履歴で経緯確認のみ行い、現在の実装対象は対応する Linear Issue を確認してから決める

引数は `$ARGUMENTS` 変数で受け取る。

---

## Phase 1: タスク選択・読み込み

### 1.1 引数解析

- Linear ID → Linear で該当 Issue の本文、コメント、状態、親子関係を確認する。
- repo 内の見送り判断だけを確認する必要がある場合:

```bash
grep -n '<キーワード>' docs/work/phase2-deferred.md
```

### 1.1b claim 確認（着手前必須・AGENTS.md packet claim protocol）

```bash
git branch --list 'claim/<Linear ID>'
```

非空なら**ハードストップ**: claim ブランチ名を挙げて BLOCKED を報告し、編集しない。空なら `git branch claim/<Linear ID>` で取得してから着手する（取得失敗も BLOCKED）。claim ブランチの削除（解放）は main 統合後の USER 専権 — エージェントは自他を問わず削除しない。

### 1.2 タスクセクション読み込み

Linear Issue 本文・コメント・リンク先仕様を読み込み、以下を抽出:
- **問題**: 何が問題か・実装内容の概要
- **根拠**: 対象ファイル・行番号・現状コードの実測情報
- **修正方針**: 採用案・参照実装・具体的なコード変更指示
- **受け入れ条件**: 検証可能な完了条件
- **状態**: 優先度・Linear の依存 Issue・前提条件

### タスクセクション記載の実測再検証（着手前必須）

タスクセクションの行番号・「現状のコード」・残タスク認識は起票時点のスナップショットであり陳腐化前提。着手前に現行コードを grep/Read で突合し、既に解消済みの項目は「陳腐化・実測訂正」として報告する。タスク本文の「Context 要約」と「行番号・Constraints」が食い違ったら後者（実コード）を優先する。

（出典: memory be_refactor_execution_20260702 / closed_issue_reaudit_20260707 / issue_g3_1_phase1_food_lstep_tag_helpers_20260709）

### 1.3 依存関係チェック

- Linear の依存関係と各前提 Issue の状態を確認する。完了状態の正本は Linear であり、ローカル検索の不一致から完了を推定しない。
- `docs/work/phase2-deferred.md` に再開条件がある場合は、Linear Issue の owner と再開条件が満たされるまで着手しない。
- 依存 Issue が未完了の場合、ユーザーに警告する: 「<前提ID> が未完了。先に実装するか？」

```bash
# 見送り判断の確認（ヒットは実装可否ではなく再開条件を意味する）
grep -n '<前提ID>' docs/work/phase2-deferred.md
# 旧番号体系の経緯だけを確認する場合
git log --all --oneline -- 'docs/archive/**' | head
```
---

## Phase 2: コンテキスト収集

### 2.1 対象ファイルの特定

- Linear Issue の「根拠」「修正方針」から、変更対象ファイルパスを抽出し、全て Read で読み込む。

### 2.2 参照実装の確認

**FE イシューの場合:**
- `features/owners/` の対応パターンを確認（ベストプラクティス参照実装）
- 変更内容に応じて、以下のファイルから該当パターンを読む:
  - フォーム系 → `features/owners/routes/OwnerForm.tsx` + `features/owners/hooks/use-owner-form.ts`
  - リスト系 → `features/owners/routes/OwnersList.tsx`
  - API hooks → `features/owners/api/` 内の対応ファイル
  - loader → `features/owners/loaders.ts`

**BE イシューの場合:**
- `.claude/rules/go-gin-backend-guidelines.md` と `backend/CLAUDE.md` を確認
- 同じresourceの既存contractと実装を参照し、packageは凝集性・利用者・依存方向で選ぶ（固定layerをtemplate化しない）

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
| 3 | フォーム送信は `useActionState`（isPending 内蔵）。`useTransition` はリスト再取得・ナビ・削除等の非フォーム操作のみ | フォーム送信 / 非フォーム操作の pending 管理 |
| 4 | `lazy()` + `Suspense` で遅延ロード | 重いモーダル・ダイアログ |
| 5 | feature 外部からの import は barrel（`features/xxx/index.ts`）必須。直接 import が正当なのは feature 内部の相対 import と lazy 動的 import のみ | 全 import |
| 6 | 三項演算子 `? ... : null`（`&&` 禁止） | 条件レンダー |
| 7 | `useState(() => ...)` lazy init | 高コストな初期化 |
| 8 | 静的 JSX はモジュール定数に巻き上げ | Select 選択肢、テーブルヘッダ等 |
| 9 | API 由来リストは `useMemo` でキャッシュ | API データの JSX 変換 |
| 10 | loader 内の独立フェッチは `Promise.all` | 複数 API 呼び出し |

**追加禁止チェック:**
- `any` 型 → `unknown` + 型ガード
- `FC` / `forwardRef` → 関数宣言 + ref as prop
- `useState(false)` + `setIsPending` → `useActionState`（フォーム送信）/ `useTransition`（非フォーム操作）
- 型は `models.ts` から `Omit`/`Partial` で導出（手書き interface 禁止）
- `console.log` → 削除

### 3.2 BE 実装チェックリスト

| # | パターン | 適用条件 |
|---|---------|---------|
| 1 | request-scoped関数は `ctx context.Context` を第1引数で受け、下流へ伝播 | request/DB/外部API |
| 2 | body/query/URI/headerを型付き入力へbindし、error・形式・範囲を検証 | HTTP boundary |
| 3 | authentication / authorization / ownershipを独立して確認 | protected resource |
| 4 | packageを凝集性・利用者・依存方向で選び、固定layerを増やさない | 新規package/API |
| 5 | error chainを `%w` で保持し、unknown errorの内部情報を返さない | error処理 |
| 6 | 同じerrorを重複ログせず、secret・個人情報をlogに含めない | logging |
| 7 | interfaceはconsumer側の必要最小methodだけ。mock目的で先行作成しない | abstraction |
| 8 | HTTPは`httptest`、DB/transaction/tenant境界はriskに応じintegration test | testing |

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

タスクセクションの「受け入れ条件」項目を1つずつ検証。

### 4.3 コード規約チェック

**FE の場合:**
- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] feature 外部への deep import なし（barrel 経由か確認）
- [ ] `&&` 条件レンダーなし（三項演算子を使用）
- [ ] フォーム送信は `useActionState`、非フォーム操作の pending は `useTransition`（`useState(false)` + `setIsPending` なし）
- [ ] 型は `models.ts` から導出
- [ ] `console.log` なし
- [ ] feature 間 import なし

**BE の場合:**
- [ ] request-scoped処理にContextを伝播し、structへ保存していない
- [ ] エラーは `fmt.Errorf("...: %w", err)` でラップ
- [ ] binding/validation/authentication/authorization/ownershipを分離
- [ ] package/interfaceは利用者と凝集性に基づく必要最小限
- [ ] unknown error・secret・個人情報をresponse/logに出していない
- [ ] tenant boundaryとOpenAPI contractをtestしている

### 4.4 Lint・ビルド・テスト実行（Docker 経由・スコープ限定）

**全体 lint / build / test（`pnpm lint` `pnpm build` `pnpm test:run` `golangci-lint run ./...` `go test ./...`）は CLAUDE.md の自動実行禁止コマンド。** 変更スコープに限定した検証を自分で実行し、全体検証はユーザーに手動実行を依頼する。

```bash
# FE の場合 — 変更した feature に限定（`--` 付き pnpm test は全件実行になる罠。npx vitest run <path> が正。出典: memory feedback_frontend_verify_harness_gotchas）
docker compose exec frontend npx vitest run src/features/<対象feature>

# BE の場合 — 変更したパッケージに限定
docker compose exec backend go build ./internal/<対象パッケージ>/...
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

### 5.1 Linear Issue を更新

Linear Issue に実装コミット、変更概要、実行した検証コマンドと結果、残余リスクを記録する。状態遷移はレーンの承認手順に従い、実装者が Done を自己遷移しない。

### 5.2 関連 Issue の確認

同じ親 Issue・Wave・クラスタの関連 Issue がある場合、影響と残余を Linear に記録する。`docs/work/phase2-deferred.md` は見送り判断専用であり、実装結果を書き足さない。

### 5.3 claim 報告（解放は USER 専権）

claim ブランチ（`claim/<Linear ID>`）は**削除しない**。実装コミット完了後にブランチ名をユーザーへ報告し、main 統合後の解放（`git branch -D`）は USER が行う（AGENTS.md packet claim protocol）。

### 5.4 完了報告

以下のフォーマットでユーザーに報告:

```
## 実装完了: <Linear ID>

### 変更ファイル
- `path/to/file1.tsx` — 変更内容
- `path/to/file2.ts` — 変更内容

### レビュー結果
- スコープ限定 vet/test: PASS（実行したコマンドと結果を明記）
- 完了条件: 全項目クリア
- 全体 lint/test: ユーザー手動実行待ち（コマンド提示済み）

### タスク
- <Linear ID> → Linear にコミット・検証結果・残余を記録済み（Done 遷移は承認手順に従う）
- claim/<Linear ID> 保持中 — main 統合後に USER が解放
```

---

## エラーハンドリング

| 状況 | 対応 |
|------|------|
| Linear Issue が見つからない | `docs/work/phase2-deferred.md` の再開条件と git 履歴を補助的に確認し、対応する Linear ID をユーザーに求める |
| 依存イシューが未完了 | 警告表示、ユーザーに続行確認 |
| Docker コンテナ未起動 | `make up` の実行を提案 |
| Lint/Build 失敗 | エラー内容を表示し、Phase 3 に戻って修正 |
| codegen 失敗 | Go モデルのコンパイルエラーを確認 |

---

## 禁止事項

- **タスクに書かれていない変更を勝手に行わない**: スコープはタスクの「修正方針」に限定
- **UI を推測で実装しない**: Figma デザインがない場合、UI 変更はタスクの指示に厳密に従う
- **ローカルで npm/go コマンドを実行しない**: 必ず Docker 経由
- **テストを省略しない**: タスクの受け入れ条件にテストがあれば必ず実行
- **タスクの受け入れ条件を勝手に変更しない**: 条件を満たせない場合はユーザーに報告
