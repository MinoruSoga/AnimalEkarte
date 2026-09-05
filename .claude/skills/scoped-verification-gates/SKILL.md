---
name: scoped-verification-gates
description: commit/push 前のスコープ限定ローカル検証ゲート。全体 lint/test は自動実行禁止のため、変更範囲に絞った build/vet/test/gofmt を自分で実行してから完了報告する。「局所検証はユーザー手動」で済ませそうな時に使用。
---

# Scoped Verification Gates

このプロジェクトには「ローカル検証をスキップして『ユーザー手動で実行してください』で完了報告してしまう」chronicな失敗パターンがある。全体 lint/test の自動実行が禁止されているのは事実だが、それは**スコープ限定の検証まで省略してよい理由にはならない**。このskillは、変更範囲に絞った検証を自分で実行し、その結果を報告するための手順である。

## いつ発動するか

- commit / push 前
- backend / frontend の lint・test・gofmt をローカルで確認したい時
- 「動作確認は完了しました」と報告しようとしている時（報告前に必ず参照する）

## 手順

1. **frontend 実行可否の判定**: 現在のモデルが Haiku かを確認する。Haiku限定で `pnpm` コマンドの自動実行が許可されている。Opus/Sonnetはユーザーに手動実行を依頼する（出典: memory feedback_pnpm_haiku）
2. **backend 検証対象の確定**: 変更した既存package/fileを指定し、以降の全コマンドで同じ値を使う。次は `medicalrecord` を変更した場合の例。別domainでは実在する対象へ置き換え、`test -d` / `test -f` で存在を確認してから実行する:
   ```bash
   VERIFY_GO_PACKAGE='./internal/medicalrecord/...'
   VERIFY_GO_FILE='internal/medicalrecord/clinical_plan_service.go'
   test -d backend/internal/medicalrecord
   test -f "backend/$VERIFY_GO_FILE"
   ```
3. **backend lint（スコープ限定）**:
   ```bash
   docker run --rm --tmpfs /root/.cache \
     -v "$PWD/backend:/app" \
     -v ekarte-go-mod-cache:/go/pkg/mod \
     -w /app \
     golangci/golangci-lint:v2.11.4 \
     golangci-lint run "$VERIFY_GO_PACKAGE" --max-same-issues 0 --max-issues-per-linter 0
   ```
   backend イメージには golangci-lint を入れないため、Makefile と同じ公式 pin イメージを使う。キャッシュは毎回フレッシュ化する（stale cacheは偽の0件を返す。出典: ops_golangci_lint_stale_cache_false_zero）。
4. **backend build/vet/test（スコープ限定）**:
   ```bash
   docker compose exec backend go build "$VERIFY_GO_PACKAGE"
   docker compose exec backend go vet "$VERIFY_GO_PACKAGE"
   docker compose exec backend go test "$VERIFY_GO_PACKAGE"
   ```
5. **gofmt**: 秒単位で完了するためスキップ理由が無い。対象ファイルを明示して必ず実行する:
   ```bash
   docker compose exec backend gofmt -l "$VERIFY_GO_FILE"
   ```
6. **DB依存テストは fresh-DB gate**: warm DBでの成功はCIのfresh DB失敗をマスクしうる。DB stateに関わる変更は一度 `DROP DATABASE ekarte_db_test` 相当のリセット後に走らせる（ユーザー承認の上で実施可否を確認する）
7. **codegen影響の確認**: modelやAPIスキーマを変更した場合は `make codegen` 実行後 `git diff` で差分をユーザーに確認依頼する（`make codegen` 自体は自動実行禁止コマンド）
8. **frontend（スコープ限定）**:
   ```bash
   docker compose exec frontend npx vitest run src/features/xxx
   ```
   `pnpm test:run --` は罠——全件実行になる（出典: feedback_frontend_verify_harness_gotchas）

## 検証の罠（実績由来）

- **lint の「全件直した」判定は cap 解除フラグ付きで行う**: `--max-same-issues 0 --max-issues-per-linter 0` を明示付与する。cap が有効な設定では件数が隠蔽される（実 11 件が 10 件表示され、10 件直しても 11 件目が後出しした実例。現行 backend/.golangci.yml は解除済みだがフラグ明示はドリフト耐性がある）。（出典: memory ops_golangci_lint_cap_and_reconcile_20260630）
- **post-edit-typecheck-ts.js の exit 2 には false-positive がある**: 出力が `DB_USER variable is not set` 等の docker compose env 警告のみで `error TS` 行が無ければ実エラーではない。`grep -cE 'error TS'` の有無で判定する。（出典: memory feedback_frontend_verify_harness_gotchas）
- **PostToolUse formatter hook が日本語コメントを行折返しで破壊しコンパイルを壊すことがある**。編集後は touched package を `go build ./internal/<pkg>/...` で必ず確認する。（出典: memory cross_tenant_write_audit_20260629）
- **frontend の import path 改名・移行は `src` `liff/src` `line-reserve/src` の3アプリ全域を対象にする**: 3アプリは `tsconfig.json` の同一 `include`（`src`, `liff/src`, `line-reserve/src`）で同一 `@/` alias を共有する。grep・機械置換を `frontend/src` 限定で行うと他2アプリの死にimportを見逃す（実例: `@/utils/` → `@/lib/` 置換が `line-reserve/src/pages/TrimmingOptionSelectPage.tsx` 等2件を残し、独立レビューの tsc で発見）。（出典: memory fe7_utils_lib_migration_scope_trap_20260718）
- **`tsc --noEmit`（`pnpm type-check`）は test ファイルを検証しない**: `frontend/tsconfig.json` の `exclude` が `src/**/*.test.tsx` 等（`liff/src`・`line-reserve/src` も同様のパターンで除外、`*.spec.*` は `src` のみ）を除外するため、tsc は本番コードのみを type/module 解決チェックする。import 改名の importer にテストファイルが含まれる場合、tsc が exit 0 でも死にimportを検出できない。`docker compose exec frontend npx vitest run <該当パス>` で module 解決を実証する（dead import なら vitest が fail-fast。実例: `@/contexts/auth-context` → `@/hooks/auth-context` の18 importer中大半がテストファイルで、advisor 指摘後 vitest で17件を実証）。（出典: memory fe7_utils_lib_migration_scope_trap_20260718）
- **検証ツール別の盲点を PASS 判定前に照合する**: tsc=テストファイル除外・ESLint（`no-restricted-imports`等）=module解決しない・grep=列挙した dir のみが対象。「検証面 ⊇ 変更面」（変更したファイル集合が検証コマンドのスコープに完全に含まれるか）を確認してから完了報告する。（出典: memory fe7_utils_lib_migration_scope_trap_20260718）

## 良い例・悪い例

✅ 良い例（スコープ限定・実行結果を報告）:
```
backend/internal/medicalrecord/clinical_plan_service.go を変更したので検証した:
$ docker compose exec backend go build ./internal/medicalrecord/... → OK
$ docker compose exec backend go vet ./internal/medicalrecord/...  → OK
$ docker compose exec backend go test ./internal/medicalrecord/... → PASS (12 tests)
$ docker compose exec backend gofmt -l internal/medicalrecord/clinical_plan_service.go → (差分なし)
```

❌ 悪い例(1) — 全体コマンドに逃げる:
```
docker compose exec backend go test ./...
```
（禁止コマンドを自動実行しようとしている。スコープを絞った代替コマンドを使うべき）

❌ 悪い例(2) — 「ユーザー手動で」に逃げる（このskillが解決対象とする失敗パターンそのもの）:
```
変更が完了しました。動作確認はユーザーの方で go test を実行して確認してください。
```
（スコープ限定の検証は自動実行が許可されている。実行せず手動依頼だけで済ませるのは chronic な回避パターン）

## 完了条件

変更パッケージの build / vet / scoped test / gofmt の**実行結果**を報告する。「ユーザー手動実行してください」で済ませることは禁止——それ自体がこのskillの解決対象となる chronic な失敗パターンである。

## 出典

memory: `feedback_claudecode_local_verify_skipping` / `ops_backend_scoped_lint_entrypoint_override` / `feedback_frontend_verify_harness_gotchas` / `feedback_pnpm_haiku` / `ops_golangci_lint_stale_cache_false_zero` / `fe7_utils_lib_migration_scope_trap_20260718`
