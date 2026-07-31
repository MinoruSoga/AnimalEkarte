# Full Gate Triage — 2026-08-01

計測のみ。製品コード・テスト・設定は変更していない。

## 1. 実行条件（再現条件）

| 項目 | 値 |
|------|-----|
| Claim | `claim/FULLGATE-TRIAGE-20260801`（取得済み。削除は USER 専権） |
| ゲート直前 HEAD | `c61bcdcfb` — `docs: add next-window USER runbook after seed static green` |
| ゲート直後 HEAD | `fb91f5451` — `test(scenarios): elevate V03/V04 residual after batch3 browser measure` |
| HEAD 差分 | **あり** — ゲート実行中に他セッションが main を前進させた |
| ゲート直前 porcelain | `M bug.md` / `M docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md` / `M docs/ops/testing/scenarios/V04-settings-master-forms.md` / `M todo.md` / `?? reports/2026-08-01-task-010-batch3.md` |
| ゲート直後 porcelain | `M bug.md`（シナリオ docs / todo / batch3 レポートはゲート中にコミットまたは整理された） |
| レポート作成時 porcelain | `M bug.md` + `?? reports/2026-08-01-full-gate-triage.md`（本ファイル） |
| 保持中 claim 数 | 35（`BUG-001`..`BUG-032` + `TASK-020` + `TASK-021` + 本 claim） |
| E2E 資格情報 | `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` は未注入（存在 probe EXIT:1）。backend `go test` には無関係 |

**信頼性注記**: backend 全パッケージ走査中に HEAD と porcelain が変化した。失敗 5 件はすべて inventory / 固定カウント系の静的アサートであり、並行 commit 内容（シナリオ docs）とは独立に再現する見込みが高い。それでも「この木のスナップショットに対する 1 回の実測」として扱うこと。

### 実行コマンド

```bash
# backend（-p 1 必須: 共有 test DB 干渉防止）
docker compose exec -T backend go test -p 1 ./... > /tmp/gates-go.log 2>&1; echo "EXIT:$?" | tee -a /tmp/gates-go.log

# frontend
docker compose exec -T frontend pnpm type-check > /tmp/gates-tsc.log 2>&1; echo "EXIT:$?" | tee -a /tmp/gates-tsc.log
```

生ログ: `/tmp/gates-go.log`（161 行）、`/tmp/gates-tsc.log`（8 行）。

---

## 2. ゲート結果サマリ

### backend `go test -p 1 ./...`

| 指標 | 値 |
|------|-----|
| 終了コード | **1** |
| `ok` パッケージ | **38** |
| `FAIL` パッケージ | **4** |
| `?`（no test files） | **4** |
| 失敗テスト | **5** |

**FAIL パッケージ（ログより逐語・タブ欠落は docker 出力の整形）**:

```
FAIL  github.com/animal-ekarte/backend/cmd/api
FAIL  github.com/animal-ekarte/backend/cmd/migrate
FAIL  github.com/animal-ekarte/backend/internal/lintscan
FAIL  github.com/animal-ekarte/backend/internal/model
```

**失敗テスト（`--- FAIL` 全件）**:

```
--- FAIL: TestRouteCompositionSmoke_TargetGraphRegistersEverySurface
--- FAIL: TestTopLevelDDLMigrationInventoryIsSingleInit
--- FAIL: TestERDTableCount_MatchesSchema
--- FAIL: TestSchemaDrift
--- FAIL: TestAllModelsExhaustive
```

**ドメインパッケージは緑（プロンプトの「既知赤」期待とは不一致。実測が正）**:

- `ok  github.com/animal-ekarte/backend/internal/billing`
- `ok  github.com/animal-ekarte/backend/internal/lstep`
- `ok  github.com/animal-ekarte/backend/internal/reservation`
- `ok  github.com/animal-ekarte/backend/internal/medicalrecord`
- `ok  github.com/animal-ekarte/backend/internal/owner`
- `ok  github.com/animal-ekarte/backend/internal/staff`

### frontend `pnpm type-check`

| 指標 | 値 |
|------|-----|
| 終了コード | **0** |
| `tsc --noEmit` 診断 | **0 件** |

**検査対象外（緑 ≠ 全域型健全）** — `frontend/tsconfig.json`:

- include: `src`, `liff/src`, `line-reserve/src`
- exclude: `**/*.test.ts(x)`, `**/*.spec.ts(x)`（各アプリ配下）
- E2E (`frontend/e2e/`) は別 tsconfig。本ゲート対象外。
- `src/types/generated/models.ts` は GORM 由来生成物。handler response DTO 欠落は本ゲートでは捕まらない。

生出力:

```
> animal-ekarte-frontend@0.1.0 type-check /app
> tsc --noEmit

EXIT:0
```

（compose の `DB_*` unset warning 3 行のみ。型診断ではない。）

---

## 3. 失敗全件の分類表

| # | パッケージ | テスト名 | 担当レーン | 分類 | 根拠（file:line / 出力） | 次に動く人 |
|---|------------|----------|------------|------|--------------------------|------------|
| 1 | `cmd/api` | `TestRouteCompositionSmoke_TargetGraphRegistersEverySurface` | **レーン無し** | **テスト側の問題** | 固定ピン `require.Len(t, routes, 496)` at `backend/cmd/api/route_composition_smoke_test.go:40-42`（2026-07-30 計測）。実測 497。製品側は `POST /api/v1/estimates/:id/successors` が `internal/billing/routes.go` に正規登録済み（billing の routes snapshot は green）。ログ: `should have 496 item(s), but has 497` | composition-gate 残件の unclaimed owner。ピンを 497 に更新（ルート削除ではない） |
| 2 | `cmd/migrate` | `TestTopLevelDDLMigrationInventoryIsSingleInit` | **レーン無し** | **テスト側の問題** | テストは post-consolidation で直下 DDL を `001_init.sql` のみ要求（`sql_migrations_integration_test.go:115-135`）。disk 実体は `001`+`002_estimate_successor_and_numbering.sql`+`003_cash_register_close_append_only.sql`。runner は全 `*.sql` を昇順適用（`main.go` runSQLMigrations）。`migrations/CLAUDE.md` も本数固定を否定 | migrate 在庫ゲート修正の unclaimed owner。002/003 を消さない |
| 3 | `internal/lintscan` | `TestERDTableCount_MatchesSchema` | **レーン無し** | **テスト側の問題** | schema 側は `001_init.sql` のみ読取 → 115。ERD マーカーは `TABLE_COUNT=116`（`docs/architecture/erd.md`）。差分は `003` の `cash_register_close_adjustments`。ログ: `schema=115 declared=116` at `erd_table_count_drift_test.go:187-192` | lintscan / ERD ゲート担当。全 `migrations/*.sql` 合算へ更新 |
| 4 | `internal/model` | `TestSchemaDrift` | **レーン無し** | **テスト側の問題** | ログ: `[CashRegisterClose] ... "deleted_at" がGoモデルにフィールドとして定義されていない`（`schema_drift_test.go:502` 付近）。Go は意図的に `DeletedAt` 無し（`cash_register_close.go:7-11` append-only）。`003` は DB 列を残し app soft-delete 禁止。allowlist は型不一致のみで未マップ列スロット無し | billing/W-013 残件 + model drift メンテ。`DeletedAt` を戻すのは製品方針違反 |
| 5 | `internal/model` | `TestAllModelsExhaustive` | **レーン無し** | **並行作業の産物** | ログ: `allModels() に未登録のモデル ... CashRegisterCloseAdjustment`（`schema_drift_test.go:598`）。型と `TableName()` は `cash_register_close_adjustment.go` に存在。billing の adjustment 経路も実装済み。ゲートは正しく検出、inventory 追記漏れ | W-013 / billing モデル追加者。`allModels()` に 1 行追加 |

### 分類分布

| 分類 | 件数 |
|------|------|
| 製品欠陥 | **0** |
| テスト側の問題 | **4**（#1–#4） |
| 並行作業の産物 | **1**（#5） |
| 環境依存 | **0** |

### クラスタ所見

- 5 件すべて **ゲート在庫の未追随**（route 本数ピン / single-init 固定 / ERD カウントソース / soft-delete 列マッピング / allModels 登録）。
- 実行時の業務ロジック赤（reservation multi-clinic、billing LTV、lstep checkup sync 等の「既知失敗」）は **今回のフルスイートでは再現しなかった**。
- live claim（BUG-001..032 / TASK-020 / TASK-021）はいずれも本 5 件の一次オーナーではない。

---

## 4. 失敗ごとの詳細（ログ引用）

### 4.1 cmd/api — route count 496→497

```
--- FAIL: TestRouteCompositionSmoke_TargetGraphRegistersEverySurface (0.00s)
    route_composition_smoke_test.go:42:
        Error: "map[...]" should have 496 item(s), but has 497
FAIL
FAIL  github.com/animal-ekarte/backend/cmd/api
```

ソースピン:

```go
// Measured 2026-07-30 red-sweep: 496 unique Method+Path pairs.
require.Len(t, routes, 496)
```

### 4.2 cmd/migrate — multi-file DDL inventory

```
--- FAIL: TestTopLevelDDLMigrationInventoryIsSingleInit (0.00s)
    sql_migrations_integration_test.go:135: top-level DDL files = [001_init.sql 002_estimate_successor_and_numbering.sql 003_cash_register_close_append_only.sql], want exactly [001_init.sql]
FAIL
FAIL  github.com/animal-ekarte/backend/cmd/migrate
```

### 4.3 internal/lintscan — ERD 115 vs 116

```
--- FAIL: TestERDTableCount_MatchesSchema (0.00s)
    erd_table_count_drift_test.go:187: ERD table count: schema=115 declared=116
    erd_table_count_drift_test.go:192: ERD table count drift: schema defines 115 distinct table(s), marker declares 116
FAIL
FAIL  github.com/animal-ekarte/backend/internal/lintscan
```

### 4.4 internal/model — SchemaDrift + AllModelsExhaustive

```
--- FAIL: TestSchemaDrift (0.31s)
    ...
    schema_drift_test.go:502: スキーマ差分を検出 (1件):
        [CashRegisterClose] テーブル "cash_register_closes" のカラム "deleted_at" がGoモデルにフィールドとして定義されていない
--- FAIL: TestAllModelsExhaustive (0.02s)
    schema_drift_test.go:598: allModels() に未登録のモデル (1件、TableName()実装あり): CashRegisterCloseAdjustment
        新規モデル追加時は allModels() への追記が必須です。
FAIL
FAIL  github.com/animal-ekarte/backend/internal/model
```

（allowlist ログの `AuditLog.ip_address` / nullability 注意 86 件は fail 対象外の情報出力。）

---

## 5. 直したくなったが直さなかった項目

| 項目 | なぜ直さないか |
|------|----------------|
| route ピン 496→497 | 本課題は計測のみ。レーン無しだが 34 本 claim 保持中の tree への介入は二重実装リスク |
| migrate single-init テスト更新 | 同上。002/003 は正規増分；テスト側修正は別 packet |
| ERD カウントソースを全 SQL 合算 | lintscan 修正は別タスク |
| SchemaDrift allowlist / deleted_at | 製品方針に関わる。W-013 オーナー判断が必要 |
| allModels() に Adjustment 追加 | 1 行だが「直さない」制約と claim 衝突回避を優先 |
| プロンプトが触れた「既知の reservation/billing/lstep 赤」 | 今回 ok。無い失敗を作らない |

---

## 6. 受け入れ検証メモ（本レポート作成側）

| 項目 | 結果 |
|------|------|
| FAIL パッケージ列挙 vs ログ | 4 = 4 |
| `--- FAIL` テスト列挙 vs ログ | 5 = 5 |
| 各失敗にレーン・分類・根拠・次の人 | 全 5 件あり |
| コード変更 | 本レポートのみ（他パス未編集） |
| 空白汚染 | レポート作成後に `git diff --no-index --check /dev/null reports/2026-08-01-full-gate-triage.md` で確認 |

---

## 7. 推奨フォローアップ（発注用・本課題外）

優先はコスト最小の在庫追随。いずれも **新規 unclaimed packet** が妥当（既存 BUG/TASK claim と衝突しない）。

1. **GATE-INV-ROUTE** — `route_composition_smoke_test.go` の 496→497（successors を明示 assert 推奨）
2. **GATE-INV-MIGRATE** — `TestTopLevelDDLMigrationInventoryIsSingleInit` を multi-file 正規在庫に合わせる（同ファイル disposable integration の `want 1` も同時）
3. **GATE-INV-ERD** — lintscan が全 `migrations/*.sql` を数える／hardcode を 116 に
4. **GATE-INV-MODEL-W013** — (a) `allModels()` に `CashRegisterCloseAdjustment` (b) `deleted_at` 意図的未マップの allowlist または将来 DROP 方針

ドメイン業務の赤は今回フルスイートでは観測されず。BUG スイープは製品バグ用 claim のまま、ゲート在庫は別レーンで切るのが衝突回避になる。

---

## 8. Orchestration evidence

- **Mode**: subagent fan-out（native Workflow はゲート長時間 docker 実行とレポート単一 writer のため、調査・分類の並列に subagent を使用。ユーザ明示の workflow-style 要件を満たす）
- **Investigate**:
  - `019fb8fa-ea98-7062-a8e8-4abcb27cfdf9` explore — claim→domain map / prior reports / tsconfig exclude — completed — integrated into §§1–2
  - `019fb8fa-ea98-7062-a8e8-4ac29408e5e0` explore — testdb AutoMigrate / env traps / -p 1 — completed — integrated into 実行条件
- **Gates** (parent shell, long-running):
  - backend `go test -p 1 ./...` → EXIT:1, log `/tmp/gates-go.log`
  - frontend `pnpm type-check` → EXIT:0, log `/tmp/gates-tsc.log`
- **Classify** (parallel, read-only):
  - `019fb8fe-bec9-79c2-920b-d8b902daa3ad` — cmd/api route fail — completed — row #1
  - `019fb8fe-beca-7650-aa7c-3831ee40adcb` — cmd/migrate DDL fail — completed — row #2
  - `019fb8fe-beca-7650-aa7c-384d995aca54` — lintscan ERD fail — completed — row #3
  - `019fb8fe-becb-7221-9a27-9436a060ca60` — model SchemaDrift + AllModels — completed — rows #4–#5
- **Writer-owned path**: `reports/2026-08-01-full-gate-triage.md` のみ（parent が単一 writer）
- **Join**: 全 subagent completed。cancel なし。

---

## Completion Report

- Run status: **COMPLETE**

### Checklist Results

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|----------------|-------------------|-----------------|--------|--------------------|----------|
| backend 全パッケージ結果 | go test 全結果取得 | EXIT:1、ok 38 / FAIL 4 / ? 4 | PASS | `docker compose exec -T backend go test -p 1 ./... > /tmp/gates-go.log` | EXIT:1; FAIL pkgs cmd/api,cmd/migrate,internal/lintscan,internal/model |
| frontend 型検査 | type-check 結果取得 | EXIT:0、診断 0 | PASS | `docker compose exec -T frontend pnpm type-check > /tmp/gates-tsc.log` | `tsc --noEmit` + EXIT:0; 0 error TS |
| 失敗漏れなし | FAIL pkg/test 全列挙 | 4 pkg / 5 tests を §3 表に記載 | PASS | `grep` 相当のログ突合 | --- FAIL 5 本 = 表 5 行 |
| 全失敗に担当レーン | claim 対応 or レーン無し | 全 5 件「レーン無し」 | PASS | `git branch --list 'claim/*'` + ソース対応 | live claims は BUG/TASK のみ；本 5 件に非対応 |
| 全失敗に分類と根拠 | 4 分類のいずれか + file:line | §3 表 | PASS | ソース読取（分類 subagent + 親検証） | e.g. route_composition_smoke_test.go:42; sql_migrations_integration_test.go:135 |
| 再現条件記録 | 前後 HEAD/porcelain | §1 に記載。前後差あり | PASS | 実行前後 `git log -1` / `status --porcelain` | pre c61bcdcfb → post fb91f5451 |
| コード未変更 | レポート以外 0 | 本ファイルのみ追加 | PASS | porcelain vs baseline | 新規 `reports/2026-08-01-full-gate-triage.md` のみが本課題産出 |
| 空白汚染なし | check クリーン | 作成後検証 | PASS | `git diff --no-index --check /dev/null <report>` | 診断行なし（作成直後に実行） |
| workflow-style orchestration | fan-out + join | 6 subagents + 2 gates | PASS | Deliverables §8 | all joined; mode=subagent fan-out |

### Run Summary

- Changed files: `reports/2026-08-01-full-gate-triage.md`（新規）のみ（本課題）
- Failure Signature log: none（受け入れ項目の FAIL なし。ゲート赤は想定内）
- Staged plan ledger: not applicable
- Risk Tier: Local write | Safety boundary events: none

### 追加 Deliverables 項目

- Root-cause summary: フルスイート赤は **在庫ゲート 4 パッケージ / 5 テスト** のみ。製品ドメインパッケージは green。主因は W-013/TASK-012 後の inventory テスト未追随。
- Minimal patch plan: §7 の 4 packet（本課題では未実装）。
- Harness: construction — 計測レポート 1 本。Execution loop: sequential — ゲート並列 → 分類並列 → レポート単一パス。Stop: checklist 全 PASS。
- De-Sloppify: ログに無い「既知赤」推測を削除し、実測の billing/lstep/reservation ok を明記。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js .../fast-full-gate-triage-20260801.md` → **PASS EXIT:0**
- Assumptions 逸脱: なし（subagent 利用可、docker exec 成功）。プロンプトの「reservation/billing/lstep 既知失敗」記述は **実測と不一致**（実測優先で ok と記録）。
- Claim release: USER が `git branch -D claim/FULLGATE-TRIAGE-20260801`（統合後）。

---

## 受領検証時の訂正（2026-08-01）

5 件の失敗を独立に再現し、分類の根拠を追跡した。**「製品欠陥 0 件」という結論は維持する**が、
2 件の分類を訂正し、1 件に根拠を補う。

### 訂正 1 — `cmd/migrate` 「top-level DDL は 001 のみ」はテスト側の問題ではない

`sql_migrations_integration_test.go:116,135` は 2026-07-17 の migration 完全統合決定
（`001_init.sql` 単一ファイル・upgrade path 002-011 削除）を機械化したものである。
現在の top-level には `002_estimate_successor_and_numbering.sql` と
`003_cash_register_close_append_only.sql` が存在し、この規則に違反している。

テストは古くなっていない。**規則が破られたか、規則が黙って撤回されたかのどちらかである。**
「テストが stale」と分類すると、記録された決定より後発のコードを無条件に優先することになる。
処置は次のいずれかで、どちらも人の決定を要する。

1. 002 / 003 を 001 へ統合し、既存 DB は DB_RESET とする（元の決定を維持）
2. 単一ファイル方針を撤回した決定を記録し、そのうえでテストを退役させる

### 訂正 2 — `CashRegisterClose.deleted_at` 未マップは意図的設計の副作用

モデルから `DeletedAt` が落ちているのは欠落ではない。
`003_cash_register_close_append_only.sql` が soft-delete 再オープン経路を意図的に塞いでいる
（同ファイル :34-36 が `UPDATE cash_register_closes SET deleted_at = NULL WHERE deleted_at IS NOT NULL`
を実行し、:44-45 が partial UNIQUE を完全 UNIQUE へ置換、:48 の COMMENT が append-only を宣言）。

したがって残るのは「テーブルに使われない列が残っている」という状態である。処置は 2 択。

1. drift チェッカに意図的非マップとして登録する
2. migration で `deleted_at` 列自体を落とす（列が残る限り、将来の実装が再び書き込む余地がある）

### 補足 — `CashRegisterCloseAdjustment` の `allModels()` 未登録は実害がある

テスト DB は model 単位の AutoMigrate でスキーマを構築するため、未登録のモデルは
テスト DB に表が作られない。本番は migration 003 が表を作るので製品影響は無いが、
このモデルに触るテストは書けない状態にある。1 行の追記で解消する。

### 再現に使ったコマンド

`docker compose exec -T backend go test -p 1 ./internal/model ./internal/lintscan -count=1`

---

## 追補 — 残る3件は独立ではなく、1つの決定に収束する（2026-08-01）

受領検証の続きで、残っていた失敗を実測したところ **route pin だけが独立で、他の3件は同一根因** だった。
§7 の「4 packet」という分解は誤りである。正しくは **1 つの方針決定 + その帰結** である。

### route pin — 独立・解消済み

`b65cf69ef` が `POST /api/v1/estimates/:id/successors` を追加していた（TASK-012 の見積後継ドラフト）。
`backend/docs/api.yaml:18274` に記載があり、`./internal/apicontract` も緑。意図的な追加で pin の更新漏れ。
`c77767f40` で 497 へ更新し `./cmd/api` は `ok`。

### 残る3件は「001 単一ファイル方針」の一点に収束する

| 実測 | 根拠 |
|---|---|
| `001_init.sql` の `CREATE TABLE` は **115** | `grep -c '^CREATE TABLE'` |
| `cash_register_close_adjustments` は 001 に **無い**。003 が作る | `grep -l` が 003 のみを返す |
| ERD gate は **001 だけ**を読む | `erd_table_count_drift_test.go:114` が `001_init.sql` を直接指す |
| ERD doc の宣言は **116** | `docs/architecture/erd.md:3,7,10` |

つまり ERD の 115 vs 116 は「doc の数え間違い」ではない。
**doc は 003 適用後の実スキーマ（116表）を正しく宣言しており、gate が 001 しか見ていないためズレている。**
数字をどちらかに寄せる修正は、どちらの方向でも誤りを固定する。

- marker を 115 にする → doc が本番実スキーマを過小申告する
- gate に 002/003 も読ませる → 001 単一ファイル方針を暗黙に撤回する

同じ構図が `cmd/migrate` の「top-level DDL は 001 のみ」と、
`CashRegisterClose.deleted_at` 未マップ（003 が soft-delete 経路を塞いだ帰結）にも当てはまる。

### したがって決めるべきは1つだけ

**002 / 003 を 001 へ統合して単一ファイル方針を維持するか、方針を撤回して記録するか。**

- 維持する場合: 002/003 を 001 へ折り込み、既存 DB は再構築。3 gate は自然に緑へ戻る。
  `deleted_at` は統合時に列ごと落とせる。
- 撤回する場合: 決定を記録したうえで、migrate gate を退役させ、ERD gate に全 DDL を読ませ、
  `deleted_at` を意図的非マップとして登録する。

**どちらを選ぶにせよ、3 gate を個別に黙らせてはならない。** 個別に直すと、
方針が生きているのか死んでいるのかが、どこにも記録されないまま消える。
