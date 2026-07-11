# BE-refactor.md — バックエンド リファクタリング実行計画（第 3 期）

- **作成日**: 2026-07-11 / **基準 HEAD**: `70f4c298`（origin/main と一致・ci.yml 全ジョブ success 確認済み） / **検証**: 全項目の対象コード・file:line・等価性主張を計画者が直接 read/grep で実証済み（初回検証 `3a6d84d9`、復元時 `70f4c298` で主要アンカー再確認）
- **対象読者**: この計画書とリポジトリのコード以外の文脈を持たない実行者（AI/人間）
- **性質**: 全項目が**挙動保存**（HTTP レスポンス・DB 書込結果・権限判定・ログ以外の観測可能挙動を変えない。ログの追加/キー整形のみ例外として項目内に明示。**エラーメッセージ文言は全項目で現状維持**）。挙動変更は全て §4 別トラックへ隔離。
- **文書履歴**: 統合バックログ版（`b8fc2e3b`）を本実行計画で置換する。統合版の X-15〜X-18/H 系の詳細証拠は git 履歴（`git show b8fc2e3b:BE-refactor.md`）が正本 — §4 から参照する。

> **過去エピックとの関係**: 第 1 期（D1-D13/R1-R3）・第 2 期（BE-refactor-v2 = G 系 28 件 + R1〜R9/R16〜R21）・挙動変更トラック（X-1〜X-13 + **X-14 バッチ 1〜3/U1〜U7 — 残 6 エントリ**）・tygo enum_style union 化（FE6-1）・LIFF health-card 実装（FE8-1/FE8-2 — 旧 M3 解消）は実行済み（`git log` が正本）。本書は第 3 期 = ①第 2 期から持ち越した挙動保存残件 R10〜R15、②X 波・G6-2 実行結果の全体掃き出しで検出した新規負債（R22〜R32）、を計画化したもの。
>
> **前提の現況（2026-07-11 夜 実測）**: `git status --porcelain -- backend/` = 空、未 push 0、origin/main の ci.yml 全ジョブ success（run 29152374862）。**R0-1 は通過可能な状態**。
>
> **行番号の腐敗耐性**: 本書の行番号は実測時点のもの。X-14 U2-U7（`7a0a6cdc`・service 層 35 ファイル）と lint 是正（`0676efa7`・7 ファイル）で一部ずれている可能性があるため、**各項目は行番号よりもメソッド名・grep アンカーを正とする**（各項目に検索コマンドを併記済み）。

---

## §1 現状理解（構造マップ）

### 1.1 システム概要

動物病院向け電子カルテ SaaS のバックエンド。Go 1.25 / Gin / GORM / PostgreSQL 18。モジュール `github.com/animal-ekarte/backend`。
レイヤ: **handler → service → repository → model**。マルチテナント: ほぼ全テーブルが `clinic_id` を持ち、クリニック間分離が最重要インバリアント。デプロイは Cloudflare Workers + Containers（`backend/worker/` は薄いプロキシ + migrate 実行口）。

### 1.2 パッケージ構成

| パッケージ | 役割 |
|---|---|
| `cmd/` | api（本体）/ migrate / coverage-ratchet / lstep-migrate / seed-export / seed-old-db / stage-import |
| `internal/handler` | `*_request.go`（バインド→`service.XxxInput`）→ `*_response.go`（model→wire DTO）。エラーは全て `RespondError(c, err)` |
| `internal/service` | 業務ロジック。DI は `service.go` の `NewServices()` に一段化済み（旧二段階 DI は G9-1 で解消） |
| `internal/repository` | データアクセス。`base.go`: `clinicScope`/`clinicScopeIn`/`medicalRecordTenantScope`/`dbOrTx(ctx,db)`。`helpers.go`: `findByIDScoped`/`updateScopedByID`(:70)/`deleteScopedByID`(:86)/`reorderByClinicID`(:15)/`reorderGlobal`(:37)。tx 機構は①`transactor.go` の `WithTx`(ctx-txKey) ②`repositories.go` の `Transaction`(repo-swap) の 2 系統併存（意図的・CLAUDE.md に規約化済み） |
| `internal/errors` | `apperrors` エイリアス。センチネル + `Wrap*` + `FromGORM`。bare `return err` 禁止（wrapcheck が CI ゲート） |
| `internal/apicontract` | api.yaml ↔ 実装の drift 検査（route gate + date-format gate） |
| `internal/{config,dbconn,infra,logger,middleware,model,seedbundle}` | 設定 / cmd 系 DSN / LINE・Lstep・S3・crypto / slog / GORM モデル / seed マニフェスト |

### 1.3 X 波（挙動変更トラック）が導入した新パターン — 本計画の掃き出し対象

2026-07-10〜11 に X-1〜X-14（U7 まで）が集中実行され、以下の新イディオムが入った:

- **`LockDraftByID`**（medical_record_repository.go:345）: カルテ行の `FOR UPDATE` ロック。**名前に反して draft 限定ではなく status 不問でロックする**（finalized チェックは呼び出し側の責務 — doc コメント・テストで意図確認済み）。5 service（treatment/examination/vital/prescription/checkup_field_result）が tx 内で使用。
- **`AcquireBookingLock`**（X-9）: 予約枠のファントム挿入防止 advisory lock。既存の競合チェックの**前段に追加**されたもので、旧チェック群は現役。
- **マスタ FK ガード**: request 由来の clinic-scoped master FK を `FindByID(ctx, clinicID, id)` で検証するガードが X-4/X-5/X-14 各バッチで追加された。**バッチ毎に複数の書き方**が生まれており（→ R24）、統一が本計画の主要作業。U2-U7（`7a0a6cdc`）で 22 サイト分のガードがさらに追加済み — **R24 の対象は着手時 grep での再列挙が正**。
- **repo 内部 tx の dbOrTx 標準化**: 旧 13 ファイルの `r.db.WithContext(ctx).Transaction` は全て `dbOrTx(ctx, r.db).Transaction` へ変換済み・lint 台帳登録済み。残る生イディオムは tx 機構の作成者 2 箇所（transactor.go:29 / repositories.go:243 — **変換禁止**）と reorder ヘルパ 2 箇所（→ R23）のみ。

### 1.4 ガード lint テスト（絶対に緑を維持・allowlist の無断編集禁止）

| テスト | 固定内容 |
|---|---|
| `repository/dbortx_inventory_lint_test.go` | dbOrTx 参加メソッド約 80 件の双方向台帳（キー: `<file>\|<Receiver>.<Method>`）。**revert も新規追加も fail** — R23/R31 はこの台帳更新を伴う |
| `repository/preload_clinic_scope_lint_test.go` | clinic-scoped マスタ Preload の clinic_id 述語必須 |
| `repository/migration_cascade_lint_test.go` | ON DELETE CASCADE 件数固定 |
| `repository/audit_tx_inventory_lint_test.go` | 臨床 hard-delete の tx 内監査台帳（全エントリ audited-tx-internal） |
| `repository/test_schema_enum_parity_test.go` | テスト DB enum ↔ 001_init.sql の完全一致 |
| `service/master_fk_write_inventory_lint_test.go` | request 由来マスタ FK write の全数名簿。**残 statusKnownUnguarded は 6 エントリ**（§4 X-14 残・本計画では status を変えない） |
| `model/schema_drift_test.go` + `all_models` 網羅 | GORM model ↔ 実 DDL（DB 接続必要） |
| `apicontract/openapi_route_drift_test.go` / `openapi_date_format_drift_test.go` | 実装ルート ↔ api.yaml / date format（`liff_response.go\|last_visit_date` は根拠コメント付き pin 済み — §4 参照） |

### 1.5 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**（パッケージ単位スコープのみ許可）。
- repository/model パッケージのテストは DB 必要（`ekarte_db_test` は `setupTestDB` が自動作成。共有スキーマは `setupSharedTestSchema` — Treatment 登録済み `069fda6e`）。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` が無出力であることを確認してからコミット。
- 権限拒否されたコマンドは人間に提示して実行依頼し、結果を得てから判定（結果なしでコミット禁止）。

---

## §2 項目 R0: 安全網の構築（最初に必ず実行）

### R0-1. 前提確認

```
docker compose ps                     # backend / db が Up
git rev-parse HEAD                    # 記録
git status --porcelain -- backend/   # ★空であること。1 ファイルでも dirty なら着手せず中断・報告★
```

`BE-refactor.md`・`FE-refactor.md`・frontend/・`.claude/` の dirty は無視してよい（触らない・ステージしない）。

### R0-2. ベースライン記録（全て green が着手条件）

```
docker compose exec backend go test ./internal/config/ -count=1
docker compose exec backend go test ./internal/apicontract/ -count=1
docker compose exec backend go test ./internal/model/ -count=1        # DB 必要
docker compose exec backend go test ./internal/service/ -count=1
docker compose exec backend go test ./internal/repository/ -count=1   # DB 必要・数分
```

パッケージ毎の PASS/FAIL とテスト数を記録。1 つでも FAIL なら着手せず中断・報告。

### R0-3. コミット規約

- **1 項目 = 1 コミット**。メッセージは `<type>(backend): <説明> (R<n>)`（type: refactor/test/docs/chore）。
- `Co-Authored-By` を含めない。**push しない**（完了報告でハッシュ一覧提示）。
- 戻し方: 直前項目は `git reset --hard HEAD~1`、それより前は `git revert`。

---

## §3 作業項目リスト（この順に実行する）

> 完了条件を満たせない項目は変更を破棄し、SKIP/BLOCKED として理由を記録して依存されていない次の項目へ進む（依存先 SKIP は連鎖 SKIP）。実装バグを新たに発見したら直さず BLOCKED 記録。テストを実挙動に合わせて曲げるのも禁止。
> ID は第 2 期からの通し番号（R10〜R15 = 持ち越し・R16〜R21 は完了済み欠番・R22〜R32 = 新規）。

### — Phase 1: 即効の小粒（持ち越し分） —

### R11. lstep_csv_helpers.go の複合日時リテラルを time.DateTime へ ✅ DONE (29c7e151)

- **対象**: `backend/internal/service/lstep_csv_helpers.go:108`（`csvDateFormats` の 3 番目の要素。アンカー: `grep -n '2006-01-02 15:04:05' backend/internal/service/lstep_csv_helpers.go`）
- **問題**: `"2006-01-02 15:04:05"` は Go 1.20+ の `time.DateTime` と完全一致。隣接の `time.DateOnly`（:106）と不統一。
- **変更内容**: 当該リテラルを `time.DateTime` に置換（値同一）。`"2006/01/02 15:04:05"`・`"2006-01-02T15:04:05"` の真の複合レイアウトは触らない。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'Csv' -count=1` → PASS
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### R12. ConfigureTimeZone の JST 再導出を config.JST 再利用に統一 ✅ DONE (046c4d04)

- **対象**: `backend/internal/config/timezone.go:23-30`
- **問題**: package init で panic-fail-fast 済みのキャッシュ `JST`（:14-20）があるのに `time.LoadLocation` を再実行。直上のドキュメント（:11-14）が「再導出せずキャッシュを使え」と明記した直後の関数が違反。
- **変更内容**（シグネチャ・戻り値型不変）:
  ```go
  func ConfigureTimeZone() error {
      time.Local = JST
      return nil
  }
  ```
  `fmt` import が未使用化する場合は除去。
- **完了条件**: `docker compose exec backend go test ./internal/config/ -count=1` → PASS
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### R10. 孤児メソッド CountWorkingStaffByReservationTypeID の削除 ✅ DONE (a5fc0e3a)

- **対象**: `backend/internal/repository/reservation_type_occupation_repository.go`（interface doc:25・宣言:27・実装:98）、同 `_test.go`、mock 3 箇所（`service/liff_service_mock_test.go`・`service/reservation_type_service_test.go`・`service/reservation_type_service_occupation_test.go`）、`service/liff_service_availability.go`（コメント内の言及）
- **問題**: G7-1 が本番呼び出しをバッチ版 `CountWorkingStaffByReservationTypeIDs`（複数形）へ切替えた結果、単数形は interface/実装/テスト/mock のみに残る死にコード。
- **変更内容**: 上記の宣言・実装・専用テスト関数・mock メソッド・コメント言及を削除。**複数形（バッチ版）は現役 — 絶対に触らない**。削除前に `grep -rn 'CountWorkingStaffByReservationTypeID\b' backend/internal --include='*.go'` で対象を再列挙し、削除後に同 grep → 0 件を確認。
- **完了条件**: 上記 grep 0 件。`docker compose exec backend go test ./internal/repository/ -run TestReservationTypeOccupation -count=1` と `./internal/service/ -run 'TestLiff|TestReservationType' -count=1` → PASS。
- **リスク / 戻し方**: mock 削除漏れはコンパイルエラーで即検出。`git reset --hard HEAD~1`。
- **依存**: R0

### — Phase 2: repository ヘルパ統合（持ち越し分・保護テスト敷設済み） —

### R13. 業務 repo の Update 13 サイトを updateScopedByID へ統合 ✅ DONE (1696ee86)

- **対象**（初回実測の行・アンカーはメソッド名）: `estimate_repository.go:81` / `checkup_repository.go:149` / `inventory_repository.go:73`※ / `diagnosis_repository.go:70,202` / `examination_repository.go:116`※ / `permission_group_repository.go:77` / `reservation_type_liff_repository.go:68` / `reservation_staff_repository.go:95`※ / `closing_special_period_repository.go:80` / `shift_entry_repository.go:94`※ / `reservation_repository.go:184`※ / `accounting_repository.go:217`※（※= dbOrTx 変種）
- **問題**: マスタ系 19 repo で確立済みの `updateScopedByID`（helpers.go:70、シグネチャ `updateScopedByID(ctx, db, m, resource, clinicID, id, fields)` — **db は第 2 引数**）と同型の手書きボディが業務 repo に残存。
- **変更内容（サイト毎の機械的手順）**: 置換前に既存ボディとヘルパを突合し、①RowsAffected==0→WrapNotFound の扱い ②更新後再取得の有無と形 ③Select/Omit 句の有無 が**完全一致する場合のみ**コアブロックを `updateScopedByID(ctx, r.db または dbOrTx(ctx, r.db), ...)` に置換する。再取得が `FindByID` 呼び出しのサイトはその行を残す。**再取得が inline `First()`/`Preload` 群のサイト（実測: closing_special_period・accounting）は再取得部を変えずコアブロックのみ置換し、それも不一致ならスキップして報告**。1 つでも差異が残る形なら当該サイトはスキップ + 理由記録。
- **除外（触ってはならない）**: treatment / care_plan_item / clinical_plan（サブクエリ隔離）、billing_item（JOIN スコープ）、medical_record の Update（draft-status 条件 + Conflict 変換）、pet_chronic_condition（F3 — RowsAffected 検査なしのため置換は挙動変更）。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS（dbortx lint・isolation テスト群含む）。置換/スキップの全サイト一覧を報告。
- **リスク / 戻し方**: 等価性の見落とし → サイト毎突合の厳守と package 全テストで担保。`git reset --hard HEAD~1`。
- **依存**: R0（保護テストは第 2 期で敷設済み）

### R14. 業務 repo の Delete 5 サイトを deleteScopedByID へ統合 ✅ DONE (2417d9c7)

- **対象**: `estimate_repository.go:95` / `checkup_repository.go:164` / `inventory_repository.go:87`（dbOrTx） / `medical_record_repository.go:269` / `appointment_admin_repository.go:101`（メソッド名 SoftDelete・resource ラベル "appointment"）
- **問題・手順・除外**: R13 と同一。**訂正情報**: `medical_record_repository.go` の Delete は `Scopes(clinicScope(clinicID)).Where("id = ?", id)` のみで **draft 述語を持たない**（draft 述語は別メソッド `DeleteDraftByAppointmentID` 側）— 素直に fold 可能。`reservation_staff` の Delete は JOIN スコープのため**対象外**。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS。
- **リスク / 戻し方**: R13 と同一。`git reset --hard HEAD~1`。
- **依存**: R13（同一ファイル群の連続変更）

### R23. reorder ヘルパ 2 本の dbOrTx 化（repo 内部 tx 生イディオムの最終残存） ✅ DONE (dfc1234d)

- **対象**: `backend/internal/repository/helpers.go:15`（`reorderByClinicID`）・`:37`（`reorderGlobal`）と、その全呼び出しサイト（マスタ系 repo — `grep -rn 'reorderByClinicID\|reorderGlobal' backend/internal/repository --include='*.go' | grep -v _test` で実行時に全列挙。計画時実測: 非テスト 28 箇所）
- **問題**: repository/CLAUDE.md の tx 規約は `r.db.WithContext(ctx).Transaction` を「使わない」と明記するが、この 2 ヘルパだけが同イディオムで独立 tx を張ったまま残存（G6-2 変換の対象外だった free 関数）。規約と実装の矛盾。
- **変更内容**: (1) 両ヘルパのシグネチャに `ctx context.Context` を追加し、本体を `dbOrTx(ctx, db).Transaction(...)` へ変更（呼び出し側は `r.db` を渡したまま = ambient tx が無ければ従来と完全同一の独立 tx）。(2) 全呼び出しサイトに `ctx` を配線（各 Reorder メソッドは ctx を保持済み）。(3) **事前検証は計画時に実施済み**: service 層の Reorder 呼び出しに `WithTx`/`repos.Transaction` 閉包（txCtx）経由は **0 件**（grep 実測）— 実行時に同じ grep で再確認し、新規に増えていた場合のみそのサイトを変換せず BLOCKED 報告。(4) `dbortx_inventory_lint_test.go` の walker は receiver 付きメソッドを対象とするため **free 関数の本ヘルパは台帳対象外の可能性が高い** — repository パッケージテストを実行し、reconcile がエントリを要求した場合のみ既存 G6-2 ブロックの様式・根拠コメントに倣って追加する（要求されなければ台帳は触らない）。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -count=1` → 全 PASS（lint 台帳 reconcile 含む）。実行時再確認の grep 結果（ambient 呼び出し元ゼロ）を報告に添付。
- **リスク / 戻し方**: lint 台帳の様式ミスは lint 自身が検出。`git reset --hard HEAD~1`。
- **依存**: R13/R14（helpers.go と repo ファイル群の連続変更のため直後に）

### — Phase 3: X 波の掃き出し是正 —

### R28. shared_file_repository の重複 deleted_at 述語 4 箇所の削除 ✅ DONE (80a681ab)

- **対象**（アンカー: `grep -n 'deleted_at IS NULL' backend/internal/repository/shared_file_repository.go`。初回実測 :41/:53/:65/:79）: 単純 Where 3 箇所 + **複合 WHERE 1 箇所**（`Where("deleted_at IS NULL AND EXTRACT(EPOCH FROM created_at) < ?", thresholdUnix)`）
- **問題**: X-13（`cb1bee11`）で `SharedFile.DeletedAt` が `gorm.DeletedAt` 化され GORM が `deleted_at IS NULL` を自動付与するようになったのに、手書き述語が残り**全クエリで同一条件が二重に付く**。
- **変更内容**: 単純箇所は述語（または Where 句ごと）削除、**複合箇所は `deleted_at IS NULL AND ` の前半のみ除去**して EPOCH 条件を残す。`Unscoped()` は本ファイルに 0 件を確認済み。生成 SQL の意味は不変。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run TestSharedFileRepository -count=1` → PASS。`grep -c 'deleted_at IS NULL' backend/internal/repository/shared_file_repository.go` → 0。
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### R25. accountingService.Create の成功パス重複解消（X-12 の残骸） ✅ DONE (8dcc7bdb)

- **対象**: `backend/internal/service/accounting_service_core.go` の `Create` メソッド後半（アンカー: コメント `// BE-refactor.md X-12:` から次の `func (s *accountingService) Update` まで。初回実測 :77-106）
- **問題**（実コードで検証済み）: completed 分岐（WithTx 内 Create + completeAccountingAppointments → 成功 slog → `syncCPMStageTag` → return）と waiting 分岐（素の Create → 成功 slog → return）で、`s.repo.Create` + `slog.ErrorContext(..., "failed to create accounting", ...)` + `apperrors.Wrap(err, "failed to create accounting")` + `slog.InfoContext(ctx, "accounting created", slog.Uint64("billing_id",...), slog.Uint64("clinic_id",...))` + `return billing, nil` が**逐語重複**している。さらに completed 側は外側で `"failed to create accounting in transaction"` の追い掛けラップ。**注意: 2 分岐は対称ではない — `syncCPMStageTag` は completed 側のみ**（これが挙動なので統一後も completed 限定を維持する）。
- **変更内容**（挙動保存 — **waiting ケースは従来どおり tx 外のまま**。tx 意味論には一切触れない）:
  ```go
  createBilling := func(cctx context.Context) error {
      if err := s.repo.Create(cctx, input.ClinicID, billing); err != nil {
          slog.ErrorContext(cctx, "failed to create accounting", "error", err)
          return apperrors.Wrap(err, "failed to create accounting")
      }
      return nil
  }
  if billing.Status == model.BillingStatusCompleted {
      if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
          if err := createBilling(txCtx); err != nil {
              return err
          }
          if err := s.completeAccountingAppointments(txCtx, input.ClinicID, billing); err != nil {
              return apperrors.Wrap(err, "failed to complete accounting appointments during create")
          }
          return nil
      }); err != nil {
          return nil, err //nolint:wrapcheck // 閉包内で文脈付き wrap 済み（"in transaction" の同義二重ラップを廃止）
      }
  } else if err := createBilling(ctx); err != nil {
      return nil, err //nolint:wrapcheck // createBilling 内で wrap 済み
  }
  slog.InfoContext(ctx, "accounting created",
      slog.Uint64("billing_id", billing.ID),
      slog.Uint64("clinic_id", input.ClinicID))
  if billing.Status == model.BillingStatusCompleted {
      s.syncCPMStageTag(ctx, input.ClinicID, billing)
  }
  return billing, nil
  ```
  既存の X-12 説明コメントは維持。エラーメッセージの観測可能な変化は「completed 側の `failed to create accounting in transaction` 外皮の消失」のみで、同義二重ラップの解消として本項目で明示許容する（他の文言は不変）。現実装の completed 分岐の順序は**成功 InfoContext → `syncCPMStageTag` → return**（実測済み）で、上のスケッチはこれを維持している。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestAccountingService_Create' -count=1` → PASS。`docker compose exec backend go test ./internal/repository/ -run TestAccountingRepository_CompleteAccountingAppointments -count=1` → PASS。diff 上で slog メッセージ・キー・`syncCPMStageTag` の条件と順序が不変であることを確認。
- **リスク / 戻し方**: 低（tx 構造は現状維持のため mock 期待の変更は原則不要。wrapcheck の nolint には根拠コメント必須）。`git reset --hard HEAD~1`。
- **依存**: R0

### R26. reservation_staff_service の同義二重ラップ解消 ✅ DONE (db64e5c7)

- **対象**: `backend/internal/service/reservation_staff_service.go` の `Update` メソッド（アンカー: `grep -n 'failed to update reservation staff' backend/internal/service/reservation_staff_service.go`。初回実測: WithTx 閉包 :162-177・外側ラップ :178）
- **問題**（実コードで検証済み）: 閉包内は 2 分岐とも文脈付き wrap 済み（`"failed to update reservation staff"` / `"failed to update excluded courses"`、いずれも slog 付き）なのに、外側が `"failed to update reservation staff"` で追い掛けラップする。staff 更新失敗パスでは同義二重、除外コース失敗パスでは冗長な外皮になる。
- **変更内容**: 外側を `return nil, nil, err //nolint:wrapcheck // tx 閉包内の 2 分岐とも文脈付き wrap 済み（同義二重ラップ回避）` に変更。**素の `return err` にしない理由**: `transactor.WithTx` はインターフェースメソッドのため wrapcheck lint が bare return を fail させる — nolint + 根拠コメントが既存イディオム（同形の先例を service 層 grep で確認して様式を合わせる）。エラーメッセージの観測可能な変化は外皮 1 枚の消失のみ（閉包内の文言は両分岐とも不変・sentinel 判定は Unwrap 連鎖で不変）。※accounting 側の同型は R25 が処理。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestReservationStaffService -count=1` → PASS（メッセージ文字列を assert するテストがあれば期待値を単一ラップ形に更新し、その旨報告）。wrapcheck 準拠は nolint 形式の目視 + CI に委ねる（push 後確認の申し送り）。
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### R27. prescription_service の LockDraftByID 失敗ログ欠落を是正（P11・ログ追加のみ） ✅ DONE (21f91c4d)

- **対象**: `backend/internal/service/prescription_service.go` の Create と Update の `LockDraftByID` エラーパス（アンカー: `grep -n 'failed to get medical record' backend/internal/service/prescription_service.go`。初回実測 :88-90/:133-135）
- **問題**（実コードで検証済み）: X-11 の 5 呼び出しサイトのうちここだけ repo エラー時の `slog.ErrorContext` が無い（P11 違反）。兄弟の `vital_service.go` は slog + wrap の形。
- **変更内容**: 2 箇所の `if err != nil {` ブロックに `slog.ErrorContext(txCtx, "failed to get medical record", "error", err)` を追加する（**wrap 文言は現状のまま維持** — 文言を vital に揃えるのはエラーメッセージ変更 = 挙動変更のためしない。slog のメッセージも wrap と同文言にして齟齬を作らない）。純粋なログ 1 行追加 ×2。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run TestPrescription -count=1` → PASS。diff が slog 2 行の追加のみ。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### R24. マスタ FK ガードの複数イディオムを共有ヘルパへ統一 ✅ DONE (812c3522)

- **対象**: `backend/internal/service/validators.go`（共有ヘルパ新設）+ 全ガードサイト。**着手時に必ず再列挙**: `grep -rn 'validateParentOwnership\|validateMasterFKs\|validateOwnedMerchandiseItemIDs\|failed to verify.*ownership' backend/internal/service --include='*.go' | grep -v _test` — **X-14 U2-U7（`7a0a6cdc`）が 22 サイト分のガードを追加済みのため、下表（計画時点の既知分）より大幅に増えている。grep の全件が対象**
- **問題**（初回計画時に全実装を read 済み）: 「request 由来マスタ FK の所有権検証」という単一の関心事が複数様式に分裂している。ただし **wrap は全サイト `apperrors.Wrap` で統一済み・wrap 文言も `failed to verify <entity> ownership` パターンに概ね一致**（テストでの文言 assert は grep 実測 0 件）。分裂しているのは slog の有無・メッセージ・キー:

| 様式 | サイト（計画時点の既知分） | entity（wrap 文言の <entity>・バイト同一で維持する） | slog |
|---|---|---|---|
| U1 クローン ×5（13 行逐語・唯一の英語 doc コメント） | `validateParentOwnership` @ checkup_type / consultation / exam_type / procedure / vaccine | `parent checkup type` 等（各実装の文言をコピー） | あり（`parent_id`/`clinic_id` キー） |
| 複数 optional FK ヘルパ | care_plan_item_service.go `validateMasterFKs` | `medicine` / `procedure` / `hospitalization plan` | **なし** |
| slice ヘルパ | campaign_service.go `validateOwnedMerchandiseItemIDs` | `merchandise item` | **なし** |
| inline ×2 系統 | billing_item_service.go（`trimming course` / `trimming option` — slog メッセージだけ別文言）、hospitalization_service.go（`cage` — Create/Update に逐語複製） | 同左 | あり（`error` キーのみ） |
| U2-U7 追加分 | inquiry / lab_import / liff / medical_record_subrecords / medicine / owner / pet / reservation 系 / staff 等 — **着手時 grep で列挙し、様式を読んでから同じ規則で統合** | 各実装の文言をコピー | 実装毎に確認 |

- **変更内容**:
  1. `validators.go` に共有ヘルパ 2 本を新設（entity 引数で**現行 wrap 文言をバイト同一に再現**する）:
  ```go
  // validateOwnedMasterFK は request 由来の clinic-scoped master FK の所有権を検証する。
  // find は対象マスタの FindByID を包んだアダプタ。nil の FK はスキップ（optional FK 規約）。
  func validateOwnedMasterFK(ctx context.Context, entity string, clinicID uint64, id *uint64,
      find func(ctx context.Context, clinicID, id uint64) error) error {
      if id == nil {
          return nil
      }
      if err := find(ctx, clinicID, *id); err != nil {
          slog.ErrorContext(ctx, "failed to verify "+entity+" ownership",
              "error", err, "id", *id, "clinic_id", clinicID)
          return apperrors.Wrap(err, "failed to verify "+entity+" ownership")
      }
      return nil
  }
  // validateOwnedMasterFKs は slice 版（campaign の merchandise_item_ids 等）。
  func validateOwnedMasterFKs(ctx context.Context, entity string, clinicID uint64, ids []uint64,
      find func(ctx context.Context, clinicID, id uint64) error) error {
      for i := range ids {
          if err := validateOwnedMasterFK(ctx, entity, clinicID, &ids[i], find); err != nil {
              return err
          }
      }
      return nil
  }
  ```
  2. 各サイトを置換。FindByID アダプタは `func(actx context.Context, cid, mid uint64) error { _, err := s.repo.FindByID(actx, cid, mid); return err }` の形で各サイトに置く（repo フィールド名はサイト毎に異なる）。既存のサイト別ヘルパ関数は、本体を共有ヘルパ呼び出しへ委譲するか、関数ごと削除して呼び出し元を直接置換（どちらでも可・ファイル内で一貫させる）。**wrap 文言が `failed to verify <entity> ownership` パターンに一致しないサイト（U2-U7 追加分にあり得る）は、ヘルパにメッセージを引数化して現行文言を維持するか、当該サイトをスキップして報告する — 文言変更は禁止**。
  3. **不変条件**: wrap 文言のバイト同一維持 = HTTP レスポンス不変。**明示許容する唯一の観測可能差分はログ**（無ログサイトに slog が付く・キー統一）。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'Rejects|ParentFK' -count=1` → 全 PASS。`docker compose exec backend go test ./internal/service/ -count=1` → 全 PASS。`docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1` → PASS（名簿 status 不変の確認）。置換サイト全数（grep 増分含む）と entity 文言の対応表を報告。
- **リスク / 戻し方**: U2-U7 増分の様式差の見落とし → 着手時 grep 全件処理 + 文言維持ルールで担保。`git reset --hard HEAD~1`。
- **依存**: R0（Phase 3 の中で最後に実施 — 対象ファイルが最多のため）

### R31. LockDraftByID を実態に合う名前へリネーム ✅ DONE (1eb09d2b)

- **対象**（計画時実測: 21 ファイル・69 ヒット — `grep -rn 'LockDraftByID' backend/internal --include='*.go'` で着手時に再列挙）: 定義 `medical_record_repository.go:345` + interface 宣言、service 呼び出し 5 サイト（treatment/examination/vital/prescription/checkup_field_result）、FK デッドロック根拠コメントでの言及 3 repo ファイル（examination/prescription/vital_repository.go）、`dbortx_inventory_lint_test.go` の台帳キー、repo テスト 2 ファイル（`medical_record_repository_test.go`・専用並行テスト `medical_record_finalize_lock_concurrency_test.go` — テスト関数名にも旧名が含まれるため関数名ごと更新）、service テスト内 mock 約 10 ファイル
- **問題**: 名前が「draft をロックする」と読めるが、実装は **status 不問で FOR UPDATE ロック**し finalized 判定は呼び出し側の責務（doc・テストで意図確認済み）。名前を信じた将来の呼び出し側が status チェックを省くと、X-10/X-11 で閉じた finalize レースが再発する。
- **変更内容**: `LockByIDForUpdate` へ機械リネーム（grep 全サイト列挙 → 一括置換）。doc コメントの「呼び出し側が finalized チェックを行う」旨は維持・強調。`dbortx_inventory_lint_test.go` の台帳キー（`<file>|<Receiver>.<Method>`）も同時更新（reconcile が強制する）。
- **完了条件**: 旧名 grep → 0 件。`docker compose exec backend go test ./internal/repository/ -run 'TestMedicalRecordRepository|FinalizeLock' -count=1` と `./internal/service/ -count=1` → PASS。
- **リスク / 戻し方**: リネーム漏れはコンパイル/lint で即検出。`git reset --hard HEAD~1`。
- **依存**: R24, R27（service 層の同時変更を避けるため後）

### — Phase 4: ドキュメント・API 契約・テスト基盤 —

### R22. 層別 CLAUDE.md の tx/ロック規約を実装に追随させる ✅ DONE (ba2f46f9)

- **対象**: `backend/internal/repository/CLAUDE.md`（「Tx 参加機構の使い分け (MANDATORY)」節 + 「タint」typo）、`backend/internal/service/CLAUDE.md`（新規ルールブロック追加）
- **問題**: ①X-11 の中核不変条件「カルテ子エンティティの書込は tx 内で行ロック（R31 後の `LockByIDForUpdate`）を取得してから finalized チェックを行う。子 repo の Create/Update は dbOrTx で ambient tx に参加させる（さもないと FOR UPDATE 行ロックと FK チェックがデッドロックする）」が**コード内コメントと lint 台帳コメントにしか存在せず**、service 層のルール文書に無い — 将来の新規カルテ子レコード実装が黙ってレースを再導入し得る。②repository/CLAUDE.md は生 tx イディオムを「使わない」と書くが reorder ヘルパの現状（R23 で解消）と、enforcement の実体（dbortx / audit_tx の 2 lint テスト名）が節内に明記されていない。③「タint 追跡」の typo（アンカー: `grep -n 'タint' backend/internal/repository/CLAUDE.md`）。
- **変更内容**（ドキュメントのみ）: (1) service/CLAUDE.md に「カルテ子エンティティ書込（MANDATORY）」ブロックを追加 — 上記不変条件 + 現行 5 サービスの先例 + 検証テスト名（`*_tx_atomicity_test.go` 群）。(2) repository/CLAUDE.md の tx 節に enforcement として 2 lint テスト名を明記し、reorder ヘルパが dbOrTx 化済み（R23）であることを反映。(3) typo「タint」→「taint」。
- **完了条件**: `grep -n 'タint' backend/internal/repository/CLAUDE.md` → 0 件。`grep -n 'LockByIDForUpdate' backend/internal/service/CLAUDE.md` → 1 件以上。ランタイム検証不要。
- **リスク / 戻し方**: なし。`git revert`。
- **依存**: R23, R31（最終的な名前・状態を文書化するため後）

### R32. CODING_RULES.md の forbidden-pattern 教材コード 14 ブロック是正（H-6） ✅ DONE (3b54c02d)

- **対象**: `backend/CODING_RULES.md`（約 948 行。変更 14 ブロック・実測フットプリント約 180 行・正味 約 −48 行。行番号は初回実測 — 各ブロックは節番号 + grep アンカーで再特定可能）
- **問題**: レビューゲート P 系規約と正反対のコードを「手順」として教えている。実在しない API（`apperrors.Wrapf`・`handleError`・`ErrInternal`）、`uint64` PK モデルへの `uuid.MustParse`、`gin.H{"error":...}` 直返し、`clinicScope` 欠落、P5 ルート権限ゲートの章欠落、さらに **PR チェックリストが自動実行禁止コマンド（`go test ./...` / `golangci-lint run ./...`）を指示**している。
- **変更内容**（ドキュメントのみ。ブロック毎の逐条仕様）:
  1. §1.3 :128-134 REWRITE — repo interface に `clinicID uint64` 第 2 引数、`Update(ctx, clinicID, id, fields map[string]any)`（P16 準拠）。
  2. §2.1 :179-181 REWRITE — ❌例の本文から `uuid.MustParse` を除去（context 欠落例としては維持）。
  3. §3.2 :259-292 DELETE-AND-POINT — 34 行の `errors.Is`+`gin.H` switch を `RespondError(c, err)` の 8 行例 + handler CLAUDE.md P7/P12 への参照 1 段落に置換。
  4. §3.3 :296-323 REWRITE — repo は `FromGORM` + `clinicScope` + uint64、service は `strconv.ParseUint` + `WrapInvalidInput`/`Wrap`（`Wrapf` は実在しないため全廃）。
  5. §3.1 :231-254 DELETE-AND-POINT — センチネル一覧の再掲（drift 済み: `ErrInternal` 非実在・4 種欠落・`Wrapf` 定義）を削除し「`internal/errors/errors.go` が正本・再掲しない」1 段落へ。
  6. §4.1 :358-360 REWRITE — `slog.String("owner_id", owner.ID.String())`（uint64 に .String() は無い・コンパイル不能）→ `slog.Uint64("owner_id", owner.ID)`。
  7. §5.1 :436-445 REWRITE — FindByID を uint64 + clinicID + `FromGORM` + `clinicScope` 形へ。
  8. §5.1 :466-475 REWRITE — Update を `Scopes(clinicScope)` + fields map + RowsAffected==0→WrapNotFound + 再取得の P4/P16 準拠形へ。
  9. §5.1 :478-487 REWRITE — Delete に clinicID/`clinicScope` を追加。
  10. §6.1 :587-620 REWRITE — 実在しない `h.handleError` を `RespondError` に、bind エラーは `WrapInvalidInput(parseBindError(err))`、201 には Location ヘッダ + `toOwnerResponse`（P7/P12/P15）。
  11. §6.3 :648-672 DELETE-AND-POINT — `gin.H{"data": ...}` 形式を全廃し `toXxxResponse()`/`RespondError` + P7/P18 参照へ。
  12. §6 に新設 §6.5「ルート登録と権限ゲート」ADD — `perm(model.ResourceX, "view|create|edit|delete")` の 4 verb 例 + PATCH（PUT 不使用）+ 免除ルート一覧 + P5/P6 参照（約 +12 行）。
  13. §7.1 :689-693 REWRITE — 死んだ `"github.com/google/uuid"` import 行を削除。
  14. §11 :934-935 REWRITE — `go test ./... -v`/`golangci-lint run ./...` を「変更パッケージに絞った `docker compose exec backend go test ./internal/<pkg>/ ...`（全体実行は自動実行禁止）」へ（backend/CLAUDE.md の禁止コマンド節を参照させる）。
  変更しないもの: §1.1 ディレクトリツリー（現状正確）・§5.2/§5.3/§5.4/§7.2/§7.3/§8/§9（現規約と整合を確認済み）。
- **完了条件**: `grep -n 'uuid\.\|Wrapf\|handleError\|ErrInternal\|gin.H{"error"\|gin.H{"data"' backend/CODING_RULES.md` → 0 件（`uuid` は文書都合の言及が残る場合はその行番号を報告）。`grep -n 'go test ./\.\.\.' backend/CODING_RULES.md` → 0 件。ランタイム検証不要。
- **リスク / 戻し方**: なし（文書のみ）。`git revert`。
- **依存**: R0（R22 と独立・どちらが先でも可）

### R15. api.yaml: Billing スキーマに medical_record プロパティを追記 ✅ DONE (6f5cb939)

- **対象**: `backend/docs/api.yaml` の Billing スキーマ（初回実測 :1472-1583。owner/pet/items/payments/payment_splits/refunds は記載済み。アンカー: `Billing:` スキーマ定義）
- **問題**: model `accounting.go` の `MedicalRecord *MedicalRecord json:"medical_record,omitempty"`（Preload 時のみ直列化）に対応するプロパティのみ未記載（スカラー `medical_record_id` はある）。
- **変更内容**: 既存の owner/pet と同じ流儀で `medical_record`（`$ref` または入れ子・nullable・readOnly・description: Preload 時のみ）を追記。実装コードは触らない。
- **完了条件**: `docker compose exec backend go test ./internal/apicontract/ -count=1` → PASS（route/date-format 両ゲート）。
- **リスク / 戻し方**: YAML 構文ミスはゲートのパースで検出。`git reset --hard HEAD~1`。
- **依存**: R0

### R29. medical_record_repository_test.go（816 行）の分割 ✅ DONE (0a4d9912)

- **対象**: `backend/internal/repository/medical_record_repository_test.go`（X-10/X-11 のテスト追加で 676→816 行、800 行上限超過）
- **変更内容**: Update-version 系と行ロック（R31 後の `LockByIDForUpdate`）系のテスト関数を `medical_record_repository_update_test.go` へ逐語移動（同一パッケージのためヘルパ共有はそのまま）。テストケースの中身は 1 文字も変えない。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run TestMedicalRecordRepository -count=1` → PASS かつ**総テスト数が分割前と同一**（分割前に -v で件数記録）。両ファイル 800 行未満。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R31（テスト関数名がリネーム後で確定してから）

### R30. repository テストの makeBilling* 7 変種をオプション構造体ヘルパへ統合 ✅ DONE (8fd74db9)

- **対象**（計画時 grep 実測 7 変種 — 着手時に `grep -rn 'func makeBilling\|func makeTrimmingBilling' backend/internal/repository --include='*_test.go'` で再列挙）: `makeBilling`（accounting_repository_unpaid_test.go — 唯一 ownerID/petID/amount/status/scheduledDate をパラメータ化した最富形）/ `makeBillingRet` / `makeBillingForAccountingTx` / `makeBillingForItemTx` / `makeBillingForRefund` / `makeTrimmingBilling`（billing_item_trimming_test.go）/ `makeTrimmingBillingWithCompletedAt`（billing_item_lstep_queries_test.go）。**対象外**: `makeTrimmingBillingItem`（BillingItem ビルダー）・`makeBillingConfirmationMedicalRecord`（MedicalRecord ビルダー）— 名前が似ているだけの別物。
- **問題**: 「最小構成の Billing を INSERT する」ビルダーが status/金額/日付フィールドだけ違えて 7 つ併存。
- **変更内容**: (1) 7 変種の全フィールド差分を表にして突合（実行時に各実装を読む。最富形 `makeBilling` のパラメータ集合を基準にするのが近道）。(2) 共有テストヘルパ（例: `billing_test_fixtures_test.go` の `makeBillingWith(t, db, opts billingFixtureOpts)`）を新設し、7 変種を thin wrapper 化 or 呼び出し置換。**各テストの INSERT される行の値が置換前後で完全同一**であることが条件（既定値の差異が 1 つでも吸収できなければ当該変種は残して報告）。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestBillingItemRepository|TestAccountingRepository' -count=1` → PASS。
- **リスク / 戻し方**: テストのみ。`git reset --hard HEAD~1`。
- **依存**: R0（最後に実施 — 価値対効果が最小のため、先行項目が押した場合の SKIP 候補筆頭）

---

## §4 別トラック（本計画では実行しない・記録のみ）

実行者はこの節に一切手を付けない。完了報告へ「未着手・別トラック」として転記するだけでよい。X-15〜X-18・H-1〜H-8 の詳細証拠と手順は `git show b8fc2e3b:BE-refactor.md` を参照。

| ID | 内容 | 状態/理由 |
|---|---|---|
| **X-14 残** | master-FK write allowlist の **statusKnownUnguarded 残 6 エントリ**（billingItemService.CreateItem[MerchandiseItemID=DEAD field 注記あり]・treatmentService.Create/Update[InventoryID]・trimmingCourseService.Update[CourseTypeID]・trimmingService.Create/Update[CourseID/OptionIDs]）へのガード実装 + isolation テスト | 挙動変更。U1〜U7 で 22+α 件は実装済み（`7a0a6cdc` ほか）。残 6 は本計画の対象外 |
| X-15 | 状態トグル系 DELETE 4 ルートの "edit" 権限（P6 例外の明文化 or delete 化） | PO 判断待ち |
| X-16 | 健診一覧のページネーション欠落・FindAlerts 下限なし | API 形状変更 = FE 同期必須 + PO 判断 |
| X-17 | RequireXRequestedWith のエラースキーマ逸脱（`{"error":...}`） | レスポンス形状変更。FE 側 grep 確認とセット |
| X-18 | password_reset の 30s タイムアウトが smtp.SendMail に非伝播 | DialContext + SetDeadline 化（挙動変更） |
| H-1 / H-2 | UpdateStaffGroups / UpdateExcludedReservationTypes の多施設スタッフ・クロステナント破壊（junction に clinic_id 無し） | HIGH・要 migration の挙動変更。別チケット必須 |
| H-3 | `billing_items.category` 索引欠落（FEAT-383 バッチの Seq Scan） | 要 migration |
| H-4 / H-5 | `audit_logs.clinic_id` / `lstep_csv_imports.uploaded_by_user_id` の Go ポインタ型 vs NOT NULL DDL | 型変更 = 挙動リスク。別チケット |
| H-7 | reservationStaffService.Update の tx 外所有権確認 TOCTOU | LOW・低頻度管理操作 |
| **H-8** | finalize-child-write-race の残存経路（treatment Update/Delete・examination Delete・prescription Delete）+ `BulkUpdateSortOrder` の finalize ゲート自体の欠落 | **HIGH・silent close 禁止**。R31 のリネーム後に着手すれば正しい名前で実装できる |
| F3 | pet_chronic_condition の RowsAffected 検査なし | ヘルパ統合はレース窓の挙動変更 = PO 判断 |
| F6 | Lstep 系死にコード群の keep/delete | 機能ロードマップ判断（Write API 停止由来の休眠可能性） |
| B-2 | Preload read-lint 未登録 3 マスタ | model association 追加が前提（設計変更） |
| M2 | `MeResponse.AvatarURL` 欠落の三方向 drift（FE6-5 の satisfies 契約ゲートでこのクラスは今後機械検出される） | AvatarURL を BE に足すか openapi/FE から消すかの決定のみ |
| **last_visit_date の wire 形式** | `liff_response.go` の `last_visit_date`（`*time.Time`/RFC3339）↔ api.yaml `format: date` — 現在は date-format gate の allowlist に根拠コメント付き pin 済み（`70f4c298`）で CI green だが、**可視化であって解消ではない** | date-only 化 or format: date-time 化の PO 判断 → 決定後に pin 削除 + 実装/宣言変更 |
| CI 構成の見直し提案 | Backend ジョブの Lint 直列ゲートが Test ステップを隠蔽する構造（TestMedicalRecordTenantScope バグが 2 日間露見しなかった実例） | ワークフロー変更 = オーナー判断（continue-on-error or 並列ジョブ化） |
| ~~tygo `enum_style: "union"` + pin~~ / ~~M3 health-card 未実装~~ / ~~push 滞留~~ | **全て解消済み**（FE6-1 / FE8-1・8cfb49e2 / origin/main = 70f4c298 全ジョブ green） | CLOSED |

## §5 やらないことリスト（禁止事項）

1. **挙動変更の一切**（§4 の X/H/F/M 系すべて）。許容される観測可能差分は各項目に明示した範囲のみ — R24=ログの追加/キー統一、R25/R26=同義二重ラップの外皮 1 枚の消失、R27=ログ 1 行追加。**エラーメッセージ文言の「統一・改善」はしない**（R24/R27 とも現行文言のバイト同一維持を明記済み。tx 構造の変更も無い — R25 は waiting ケースを tx 外のまま維持する）。
2. **`master_fk_write_inventory_lint_test.go` の status 変更と `cross_tenant_master_fk_write_test.go` の編集**（X-14 残 6 件は別トラック。R24 は名簿の status を変えない）。
3. **機能追加・仕様変更・migration の新規作成/編集・API 形状変更**。
4. **依存ライブラリの追加・更新**（go.mod / go.sum 不可侵）。tygo 関連（tygo.yaml / Makefile / docker-compose の codegen）も触らない（FE6-1/6-5 で整備完了済み — 現状維持が正）。
5. **transactor.go:29 と repositories.go:243 の生 Transaction**（tx 機構の作成者 — 変換禁止）。
6. **`make codegen` / `frontend/` 配下 / DB リセット / migration 適用 / `docker compose up|down|restart` の実行**（例外: テスト再現のための `ekarte_db_test` 内テーブル操作は使い捨て DB として許容 — 本体 `ekarte_db` は不可侵）。
7. **フルリポジトリ検証の無断実行**: `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` 禁止。完了条件記載のスコープ限定コマンドのみ。拒否時は人間に依頼し結果必須。
8. **lint 台帳（dbortx/audit_tx/date-format 等）の allowlist を「通すため」に編集しない** — R23/R31 で必要になる正当な台帳更新は各項目の手順に従い、根拠コメント付きで行う。`liff_response.go|last_visit_date` の pin は削除しない（PO 判断待ち）。
9. **push・PR 作成・外部書き込み・コミットの squash/rebase/amend 禁止**。進捗は本ファイルの項目見出しへ `✅ DONE (hash)` を追記する形でのみ記録。
10. **`.claude/` 配下・`FE-refactor.md`・frontend/ への変更禁止**。

## §6 実行者への指示文（このままコピペして渡す）

```
あなたはこのリポジトリのバックエンドリファクタリング実行者である。
リポジトリルートの BE-refactor.md が唯一の作業指示書である。以下を厳守せよ。

1. まず BE-refactor.md を全文読む。次に backend/internal/repository/CLAUDE.md、
   backend/internal/service/CLAUDE.md、backend/internal/handler/CLAUDE.md を読む。
2. §2 の R0（安全網）を最初に実行する。git status --porcelain -- backend/ が
   空でない場合、およびベースラインの 5 パッケージテストが 1 つでも赤い場合は、
   着手せず報告して終了。
3. §3 の作業項目を記載順（R11→R12→R10→R13→R14→R23→R28→R25→R26→R27→R24→
   R31→R22→R32→R15→R29→R30）に、1 項目ずつ実施する。
   - 1 項目 = 1 コミット。コミット後に次へ。並行作業禁止。
   - 各項目の「完了条件」のコマンドを全て実行し、満たさない限りコミットしない。
   - 変更した Go ファイルは gofmt -l（backend コンテナ内）無出力を確認する。
   - 権限拒否されたコマンドは人間に提示して実行を依頼し、結果を得てから判定。
   - 満たせない場合: 変更を破棄し SKIP/BLOCKED 記録、依存されていない次の項目へ。
4. R13/R14/R30 の「完全一致の場合のみ置換」プロトコルを厳守し、スキップした
   サイトと理由を必ず記録する。R23 の ambient 呼び出し元検証、R24 の実行時
   サイト再列挙と現行文言維持、R25 の slog 同一性確認も同様に証拠を残す。
5. §4（別トラック）と §5（やらないこと）に該当する作業は、必要だと感じても
   実行するな。実装バグを発見したら直さず BLOCKED として報告せよ。
6. push するな。コミットはローカルに残す。
7. 全項目終了後の完了報告に含める:
   - 項目ごとの DONE/SKIP/BLOCKED とコミットハッシュ
   - R13/R14 の置換/スキップ全サイト表
   - R23/R31 で更新した lint 台帳エントリ一覧
   - R24 の置換サイト全数と維持した文言一覧
   - ベースラインと完了時点の 5 パッケージテスト結果比較
```

---

## 付録: 実行順の依存関係まとめ

```
R0 ─┬─ R11, R12, R10                     （Phase 1: 相互独立）
    ├─ R13 ─ R14 ─ R23                   （Phase 2: repo ヘルパ系・同一ファイル群直列）
    ├─ R28, R25, R26, R27 ─ R24 ─ R31    （Phase 3: 掃き出し → FK 統一 → リネーム）
    ├─ (R23, R31 後) R22                 （tx/ロック規約の文書化は最終名で）
    ├─ R32, R15                          （独立ドキュメント/契約）
    └─ (R31 後) R29 ─ R30                （テスト分割・fixture 統合は最後）
```

- R24 は service 層最多ファイルの変更のため Phase 3 の最後。R31 は R24/R27 と同じ service 群を触るためその後。
- R22（CLAUDE.md）は R23 の reorder 変換と R31 のリネームを文書化するため両者の後。
- R29 は R31 のテスト関数名確定後。R30 は価値対効果が最小のため最終・SKIP 許容候補。
- R25 が accounting_service_core.go の二重ラップを包含するため、R26 のスコープは reservation_staff のみ。
