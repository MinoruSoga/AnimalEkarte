# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-16（docs/ 再編: 旧 docs/tasks/open の個別タスクを本書「個別タスク詳細」節へ統合。旧 docs/tasks/・docs/archive/ は削除 — 完了記録は git 履歴参照）
> 前回: 2026-07-15（対応済除去: PO-003・TEST-FLAKE-P2）
> 監査範囲: backend / frontend / CI-CD / migrations / docs / GitHub Issues / git 状態
> **push・外部書き込み・credential 変更はユーザー所有アクション。**（PR マージはユーザーが手動で行う。本台帳には載せない）
> **別台帳**: BE 保留 = `BE-pending.md` / PO 判断キュー = `q&a.html`。過去のバグ監査（旧 docs/archive/bug.md）・受付テレメトリ完了記録（旧 docs/archive/tasks/closed/change-ui.md）は git 履歴参照
> **本書の役割**: プロジェクト横断 TODO・個別タスク詳細・今期着手可能な BE 残・BE/FE リファクタ次期引き継ぎ・やらない判断の正本台帳。

---

## 運用規約

### Docker 検証規約（BE・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

### 台帳スコープ規則

- 今期着手可能な BE 残タスクのみを「BE 残タスク」節に記載する（シークレット・テスト・PERF 等）。対応済みは残さない。詳細・手順の正本は git 履歴（旧 docs/tasks/closed/ は削除済み）。
- その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。
- PR マージ判断・マージ状態・マージ用チェックリストは本台帳に載せない（ユーザー手動）。
- PO 決裁の正本は `q&a.html`。2026-07-15 時点で PO-001〜008 は回答済み。決裁済みの「やらない」は実装着手禁止のまま。決裁済みの「即実装可」は本台帳の「PO 決裁」節を正とする。
- 着手保留・次期送り・任意検証は `BE-pending.md` を正本とする。再検討トリガが立つか判断が出たら、実装単位として本台帳の該当節へ戻す。

---

## Project TODO

### P1 — Open Issues（2026-07-15 時点・台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #211 | 検査・健診パッケージ化 | **PO 決裁済**。**即実装**: (1) アドプリット（checkup_types id=15＋配下4フィールド）削除 (2) 尿比重 min/max 空化 — 同一 seed スライス。`db_reset` は USER。**残実装**: exam_results 複合 FK（新規 migration・今四半期）。**やらない**: CRUD UI／四季分割・腎臓ドック／select 異常ハイライト（#249 同時）／ライブ E2E。provisional 解除はクライアント臨床責任者確認後に seed 手動更新 |
| #201 | 薬量自動計算 | **実装・PO 運用確定済**。**エージェント残**: コード注記/audit metadata「運用確定待ち」→「暫定確定(2026-07-15 PO)」。**gh クローズは USER**（B5 文言）。admin_route・prescriptions 連携はスコープ外確定 |
| #212 | カバレッジ90%目標 | ratchet ゲート導入済。90% 到達自体は長期目標のまま未達 |
| #89/#97/#98/#99/#109 | シークレット移行・ローテーション | **USER BLOCKED**（リポジトリ Phase A 済）。4系統ローテ / P5-2 Secrets / #97 本文マスク / #109 フォールバック撤去。詳細は SEC-SECRETS-5 |

その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。

**クローズ済みで本表から除去**: #213 / #196 / #194 / #189 / #229（gh クローズは USER）／PO-003（受付テレメトリ常時有効固定・2026-07-15）／TEST-FLAKE-P2（隔離 DB 化・2026-07-15）

### P1 — lab_import 外部検査連携

> **PO-007: 今は作らない**（FE 着手禁止）。実データ経路開通後に最小 UX 定義してから実装可。正本: `q&a.html` PO-007。

### P2 — リファクタリング follow-up

- [ ] **PERF/FOLLOWUP 系** — 未消化（`FOLLOWUP-X14A`・`PERF-FOLLOWUP-01/02/05`・`PERF-M1/M2/M3`・`FEAT-searchable-select-targets` — 詳細は本書「個別タスク詳細」節）。コード裏付け: `accounting_handler.go:161` の `hasPermission` がハンドラ内に残存（旧 PERF-FOLLOWUP-03 相当）。**PERF-FOLLOWUP-07 は実装済みと確認**（`backend/internal/service/n1_lint_test.go` が本タスクを起票元として存在。旧 open ファイルが陳腐化していた）

> BE リファクタ（第7期）・FE リファクタ（第6期）は **完了**。次期引き継ぎは下記節を参照。

### P3 — インフラ / その他

- [ ] **[USER] P2 Terraform（internal ALB + VPC Origin）本番適用** — `infra/terraform/terraform.tfvars` はローカルに準備済み（gitignore 対象）。`terraform apply` の実行判断は USER
- [ ] **[USER 判断] stg-smoke の login/CRUD 復活の要否** — 撤去済み・復活時の手順は workflow 内コメントに明記
- [ ] **[USER] Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認** — リポジトリ外での確認が必要・未検証
- ⚠️ **ECS ロールバック経路の `.env.staging` 依存ギャップ** — 通常の STG 運用（Cloudflare 正系統）では影響なし。詳細: [ECS ロールバックランブック §2](docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)

### PO 決裁 — 即実装可（未消化）

> `q&a.html` PO-001〜008 は回答済み。正本は同 HTML。対応済みは本表から除去。

| 優先 | ID | 内容 | 備考 |
|------|-----|------|------|
| 1 | #211 A1+A2 | アドプリット seed 削除＋尿比重 min/max 空化 | 同一 seed スライス。`db_reset` は USER |
| 2 | #201 B2 | 「運用確定待ち」注記 →「暫定確定(2026-07-15 PO)」 | コード/監査メタ |
| 3 | #211 A6 | exam_results 複合 FK の新規 migration 起草 | 既適用 001 編集禁止。STG 適用は USER |
| 4 | PO-008 | 7-1/7-3 実装現行値ファクトシート1枚 | クライアント確認用。7-2/7-4 はブロッカー |
| — | PO-002 | Sentry Phase 1（例外+版数のみ・PII off） | ベンダ/課金は USER。security-review 必須 |
| — | PO-006 | ADR-003 TRIGGER Issue | **起票操作は USER**（案1B・二重保持解消もスコープに含める） |

**やらない（決裁確定）**: 健診 CRUD UI / 四季分割・腎臓ドック / select 異常ハイライト（今） / admin_route・prescriptions dose 連携 / 会計 DTO 化 / lab_import FE / #211 ライブ E2E（今） / Sentry Phase 2（実測後）

### ユーザー所有アクション一覧

| アクション | 根拠 |
|-----------|------|
| SEC-SECRETS-5: 4系統ローテーション＋ P5-2 GitHub Secrets 登録＋ #97 本文マスク | PUBLIC 履歴露出の実効無効化。手順: runbook §0.5 / `infra/cloudflare/README.md` |
| seed 003_demo 変更後のローカル/STG `db_reset`（SEC / #211 A1+A2） | migration-seed-safety。エージェントは DB reset 自動実行禁止 |
| #109 Phase C: `STG_DEMO_*` 登録後に performance-tests フォールバック撤去（エージェント可） | Secrets 未登録のまま撤去すると scheduled が壊れる |
| ECS ロールバック時のみ: SSM Parameter Store 登録＋IAM 権限 | 通常運用では不要 |
| Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認/設定 | 外部システム操作 |
| `terraform apply` 承認（tfvars 準備済み） | インフラ破壊的変更 |
| ADR-003（PO-006）独立 Issue 起票 | PO 承認済。案1B TRIGGER＋二重保持解消検討 |
| #201 gh Issue クローズ | PO 運用確定済。残記録文言は B5 |
| #229 gh Issue クローズ | ローカル実装完了済み |
| Sentry 等ベンダ確定・課金契約（PO-002） | 課金・外部契約 |

### 証跡サマリー（2026-07-15）

| 検査対象 | 結果 |
|---------|------|
| GitHub Issues（open） | 台帳掲載は上表。FEAT 群は gh 正 |
| Backend coverage ベースライン | `backend/.coverage-baseline` 存在（#194 CLOSED） |

---

## BE 残タスク

> 今期着手可能な BE 残タスクのみ。対応済みは残さない。詳細・手順の正本は git 履歴（旧 docs/tasks/closed/ は削除済み）。
> 次期送り・着手保留・任意検証は `BE-pending.md`。本書と重複させない。

**エージェント実装可能な残タスク:**

| ID | 優先度 | 内容 | 状態・条件 |
|----|--------|------|-----------|
| SEC-SECRETS-5 | **USER 残** | リポジトリ Phase A 済。**残（credential-impacting）**: 4系統ローテ、P5-2 `gh secret set`、#97 本文マスク、#109 `STG_DEMO_*` 登録後のフォールバック撤去。#98/#99 は Phase 8 まで PENDING | Issue クローズはローテ完了後。seed 変更後は `db_reset`（USER） |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。詳細経緯は git 履歴（旧 docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md）。P0/P1 は完了済み・2026-07-12 CLOSED |

**スコープ外**: `FEAT-searchable-select-targets`（個別タスク詳細参照）は FE 案件のため本節に含めない（Project TODO の P2 に掲載）。

---

## BE リファクタ引き継ぎ

- **第7期**: **完了**（BE7-0〜BE7-21）。詳細は git 履歴を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。
- **PO 決裁**: `q&a.html`（BE 関連の「やらない」は PO-004/005/006）。

### 次期監査への引き継ぎ

- **[要精査] `middleware/liff_auth.go:56,116`**: `LineCustomerRepository.FindOrCreateByLineUserID` を middleware から直接呼んでいる（層違反の疑い）。
- **[次期] apicontract のフィールドレベルゲート**: BE7-17 クラスの api.yaml 乖離を機械検出する拡張。
- **[次期] god-function 走査**: `examination_service.go` ReplaceItems（101行）/ medicine・reservation・cash_register の80行台。
- **[次期] `lstep_csv_import_service.go` が自前 `s.db`（gorm 直 import）を持つ妥当性**。
- **[次期] `reservation_type_handler.go` の weekly/specific 相互依存バリデーションの service 移動**（LOW）。
- **[次期] `AuthService`/`TokenService`（BE7-20/21 で抽出済み）と `validators_auth.go` の統合、P8/P11 準拠の付与**。
- **[次期] `AuditService.Log` の interface 露出整理**。

### 第7期で確定した「やらない」判断（次期でも踏襲推奨）

- duration リテラル（`24*time.Hour` 等）の一括定数化・`INTERVAL '365 days'` SQL 共通化はしない。
- `internal/middleware/response.go` の `respondError` と handler `RespondError` の統一はしない（X-17 で意図的見送り済み）。
- 犬猫種別判定ヘルパ `isDog/isCatSpeciesName`（部分一致・マーケ）と `doseSpeciesAliases`（完全一致・投薬 fail-closed）は**契約が意図的に異なるため統合禁止**。

---

## FE リファクタ引き継ぎ

- **第6期**: **完了**（FE6-0〜FE6-18）。詳細は git 履歴を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。
- **PO 決裁**: `q&a.html`。残 FE 即実装は PO-002（Sentry Phase 1）。

### 次期監査への引き継ぎ

- **OwnerSearchModal の React Query 化**: FE6-1 はバグ修正に留めた。feature 側 `useSearchOwners` フックへの構造改善は次期。
- **`ShiftFormDialog` の `use-shift-form.ts` 抽出 / `TreatmentRow` の EditableCell 化 / `ChangePasswordDialog` の api 層整理**: 実害なしの一貫性改善。
- **liff / line-reserve の `index.html` に CSP メタタグがない**（メインアプリのみ設定済み）。
- **`src/lib/` と `src/utils/` の役割分担が不文律**。規約明文化候補。
- **export されているが外部参照のない型シンボル約15件**（`CPMStageOption` 等）: 次期にまとめて掃除。
- **Pet属性ラベルの単一ソース化**: FE6-8 は二重定義＋ガードテストでの乖離検知に留めた。
- **曜日ラベル契約の統合**: master の `ReservationTypeAvailableSlotsCalendar` は月曜始まりヘッダー。契約が異なるため統合には設計が必要。

### 第6期で確定した「やらない」判断（次期でも踏襲推奨）

- `use-*-form` 系フックの共通スケルトン抽象化は、ドメインロジックが実質的に異なり害と判定済み。
- `src/features/owners/components/pet-edit-field-shared.tsx` のリネーム・`.ts` 化は不可（JSX 定数を含む）。
- `src/components/ui/`（shadcn 生成物）・`src/types/generated/`（tygo 生成物）は編集しない。
- `types/index.ts` の FA9 構造自体の変更はしない（FE6-18 でドキュメント明文化のみ実施済み）。

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| `BE-pending.md` | 着手保留・次期送り・任意検証の正本 |
| `q&a.html` | 内部 PO 判断キュー（決裁記録の正本。PO-001〜008 回答済み） |

> 旧 `todo.md` / `BE_todo.md` / `BE-refactor.md` / `FE-refactor.md` は本ファイルへ吸収済み（削除）。
> 旧 `docs/tasks/{open,pending,closed}/` は 2026-07-16 に廃止 — open は本書「個別タスク詳細」節へ統合、closed の詳細は git 履歴（追跡分のみ。gitignore 下の untracked 分は消失）。
> 旧 `docs/archive/`（bug.md・change-ui.md 含む 912 ファイル）は 2026-07-16 に削除 — 全て完了記録であり `git log --all -- docs/archive/` で復元可能。

## 個別タスク詳細（旧 docs/tasks/open 統合・2026-07-16）

> 旧 docs/tasks/ は gitignore 済み（untracked）だったため、本節が唯一の記録。
> 旧 open/README.md の優先度: FOLLOWUP-X14A(P1) > PERF-FOLLOWUP-05(High) > PERF-FOLLOWUP-01(High) > PERF-M1/M2(Medium) > PERF-FOLLOWUP-02(Medium) > PERF-M3(Low)。

### FOLLOWUP-X14A: treatmentService.Create の MedicineID→DecreaseStock フォールバックがクロステナント在庫減算を許す（P1）

- **問題**: X-14a(`6438c6ed`) は InventoryID 直接指定経路のみ塞いだ。`InventoryID=nil` かつ `MedicineID` 非 nil のとき、MedicineID の値がそのまま `inventory_items.id`（主キー）として `DecreaseStock` に渡り、他クリニックの inventory_items.id と数値衝突すれば他院在庫を無警告で減算できる書込 IDOR。
- **根拠**: `backend/internal/service/treatment_service.go:278-292`（`targetInvID = *input.MedicineID` フォールバック）、`backend/internal/repository/inventory_repository.go:84-95`（`DecreaseStock` は `Where("id = ?", id)` のみで clinic_id 述語なし）。medicines と inventory_items の ID 採番空間は無関係（`medicine_service.go:279`、`cross_tenant_master_fk_write_test.go:107` 注記）。`validateTreatmentMasterFKs` は medicines 行の所有権しか検証しない。
- **修正方針**: 案A（推奨）= Medicine→在庫行解決時に `Inventory.FindByID(ctx, clinicID, resolvedInvID)` で所有権検証してから DecreaseStock（`validateOwnedMasterFK` パターン）。案B = `DecreaseStock` に clinicID 引数追加し `Where("id = ? AND clinic_id = ?")`、RowsAffected==0 を NotFound 扱い、全呼び出し元を追随。A/B 併用可（多層防御）。
- **受け入れ条件**: ① MedicineID のみ指定で他院 inventory_items.id と衝突するケースの isolation テスト（修正前 FAIL→修正後 PASS: 404/拒否・在庫不変）② `docker compose exec backend go test ./internal/service/ -run 'TestTreatmentService|RejectsCrossClinic' -count=1` PASS ③ `master_fk_write_inventory_lint_test.go` の根拠コメントに MedicineID フォールバック検証済みを追記 ④ 同一院内正常系は無変更 PASS。
- **状態**: P1。悪用には ID 数値衝突が必要。STG 実データの衝突有無は SELECT で要確認。起票元 = clinic-isolation-auditor。

### PERF-FOLLOWUP-05: pwreset メール送信ゴルーチンの shutdown drain（High）

- **問題**: パスワードリセットメール送信が fire-and-forget goroutine（30s timeout, `context.Background()`）のため、サーバ shutdown 時に孤児化。STG で「リセットメール届かない」報告実績あり。
- **根拠**: 送信失敗の slog 記録は `password_reset_service.go:105-107` で対応済み。だが送信は依然 `go func()`（`:101-109`）。`main.go:191-213` の graceful shutdown は HTTP サーバのみ drain。
- **修正方針**: 短期（推奨）= WaitGroup 導入（`s.wg.Add(1)` / `defer s.wg.Done()`）+ shutdown hook で `s.wg.Wait()`。中期 = outbox 型ジョブキュー。残調査: SMTP 実タイムアウト値、goroutine リーク監視。

### PERF-FOLLOWUP-01: pets 複合インデックス追加（High）

- **問題**: PERF-3 の `pets.CountLivingByOwnerIDs`（`WHERE owner_id IN (...) GROUP BY owner_id`）に効く複合インデックス未整備。大規模クリニックで効果半減。
- **根拠（2026-07-12 スコープ訂正済）**: `owners(clinic_id, id)` は不要。実クエリ `pet_repository.go:149-166` は `deceased_at IS NULL` で生存判定。既存は `idx_pets_owner_id(owner_id) WHERE deleted_at IS NULL` と `idx_pets_deceased(clinic_id, deceased_at)` のみ。
- **修正方針**: `(clinic_id, owner_id, deceased_at)` 相当の複合インデックスを `backend/migrations/` に連番追加。`pg_indexes` で重複確認後、`EXPLAIN ANALYZE` で事前事後確認。STG 適用は USER。

### PERF-M1: SyncAnnual4CheckupTag の GetHealthPreventionThresholds N+1 解消（Medium・要再検証）

- **問題**: バッチループで PERF-1 hoist 済みの2関数と異なり、`SyncAnnual4CheckupTag` はオーナーごとに `GetHealthPreventionThresholds` を呼ぶ N+1 が残存（起票時点）。
- **根拠**: `backend/internal/service/lstep_health_tag_sync_checkup.go:181`、`lstep_health_tag_sync_batch.go` のループ。
- **修正方針**: `cachedThresholds` 引数付き private `syncAnnual4CheckupTagImpl` を切り出し（`SyncVaccineDeadlineTag` と同型）、バッチから pre-hoist 値を渡す。公開シグネチャ変更なし。`perf_n1_regression_test.go` 拡張で回帰保証。
- **要再検証**: `n1_lint_test.go`（PERF-FOLLOWUP-07 実装済み）の検出パターン・allowlist と突合し、残存を実測確認してから着手。

### PERF-M2: SyncFilariaTag / SyncFleaTickTag / SyncFoodPurchaseTag の閾値・設定 hoist（Medium・要再検証）

- **問題**: 同バッチループ内の3関数が clinic 定数（閾値・タグコードマッピング）をループ内で繰り返し取得している可能性（起票時点・調査フェーズあり）。
- **修正方針**: 各関数の `settingsSvc` / `tagCodeRepo` 呼び出しを調査 → ループ内 fetch があれば PERF-1 同型 hoist。call-count spy の回帰テストを `perf_n1_regression_test.go` に追記。M1 と同じく `n1_lint_test.go` と突合してから着手。

### PERF-FOLLOWUP-02: dormant/health-prevention バッチの無制限全件取得（Medium）

- **問題**: `SyncHealthPreventionTagsForClinic`→`ownerRepo.FindAllWithLineUserID`、`DetectDormantOwners`→`medRecordRepo.FindDormantOwnerEntries` が LIMIT なし全件取得。
- **修正方針**: カーソルベース推奨。`FindAllWithLineUserIDCursor(ctx, clinicID, afterID, limit)` 新設、1ページ約500件。page 境界（500/501件）unit test + カーソル前進 assert。対象: `lstep_health_tag_sync_batch.go` / `lstep_batch_dormant.go`。現状1,000件以下では非問題。

### PERF-M3: CreateCheckupSync — 空 OwnerIDs 早期 return で監査ログスキップ（Low）

- **問題**: `backend/internal/service/checkup_sync_service_create.go` の `CreateCheckupSync` が `len(input.OwnerIDs)==0` で即 return し監査ログ未記録（M-3 既知ギャップ）。
- **修正方針**: 案A（推奨）= 早期 return 前に `LogLstepOperationWithMetadata(..., "checkup_sync_create", ..., {"owner_count": 0})` を記録。テスト: 空 OwnerIDs で audit mock が1回呼ばれることをアサート。

### FEAT-searchable-select-targets: 検索可能 Combobox 化（FE・実装完了、目視確認のみ残）

- **実装状況**: P1〜P3 全実装完了（type-check/lint/隣接テスト green）。SearchableSelect = `frontend/src/components/ui/searchable-select.tsx`。適用済み: 予約区分・担当者(`ReservationFormFields.tsx:334,416`)、診断名1/2+カテゴリ(`DiagnosisHeaderDiagnosis.tsx:52,58,64`)、診療計画病名(`ClinicalPlanSection.tsx:47`)、主訴(`InterviewChiefComplaint.tsx:45`)、ワクチン(`VaccinationForm.tsx:72`)、検査種別・担当医(`ExaminationFormFields.tsx:56,63`)、健診種別・担当医(`CheckupForm.tsx:111,143`)、入院ケージ(`HospitalizationBasicInfo.tsx:106`)、薬剤親カテゴリ(`MedicineSidePanelSections.tsx:67`)、指名フィルタ(`ReceptionFilterPanel.tsx:59`)、医師フィルタ(`ReservationManagementCalendar.tsx:85`)、動物種(`NewOwnerInlineForm.tsx:83`/`PetEditModalFieldSections.tsx`)、スタッフフィルタ(`ShiftCalendar.tsx:107`、per-option `disabled` 追加)。
- **意図的スキップ**: `ShiftFormDialog` テンプレ選択（非制御アクショントリガー）／`ReservationTypeSidePanel` グループ選択（カラードット custom JSX・実件数<15）。保留候補: Lステップ TriggerType（`LstepDeliveryMonitorPageParts.tsx:71`）。
- **残作業**: [USER] 目視確認（検索・スクロール・選択・カスケード・per-option disabled）。

---

## docs/ 再編（2026-07-16）で発見した残課題

- [ ] **BUG候補: `frontend/src/hooks/use-reception-kanban.ts:18` の既存 type エラー** — FEAT-searchable 作業時に発見・無関係のため未修正のまま。要修正。
- [ ] **[DOC] `docs/architecture/auth.md`（旧 AUTH.md）の内容が自己申告 2026-06-12 のまま** — 以後はヘッダ一括付与のみで本文未更新。RBAC 34 リソースの実装突合が必要。
- [ ] **[DOC] `docs/architecture/data-flow.md` 同上（自己申告 2026-06-12）** — Request ID / 非同期同期の記述を実装と突合。
- [ ] **[DOC] `docs/ops/testing/INTEGRATION_TEST_PLAN.md` 同上（自己申告 2026-06-12）** — 統合テスト戦略の現行 CI 構成との突合。
- [ ] **[USER・任意] Notion EkarteSprint 文字化け3語の目視確認** — 2026-07-15 の保留9件適用は完了（読み戻し 9/9 PASS）。転送時に文字化けした3語（します／共有済み／事前提供）の適用先ページ（クレジット訂正フロー／検査④機器データ取込／検査⑥自動連携調査）の該当文のみ目視確認できればクローズ。
- [x] ~~`.gitignore` の `docs/tasks/` ルールと 376 追跡ファイルの矛盾~~ — 2026-07-16 docs/tasks 廃止で解消（死にルール3行も除去）
- [x] ~~`docs/openapi.yaml` と `backend/docs/api.yaml` の二重管理（302行ドリフト）~~ — 2026-07-16 削除・Swagger UI を正本へ再配線で解消
- [ ] **[DOC] `infra/docs/` と repo `docs/ops/` の役割分担が不明瞭** — `infra/docs/{architecture.md,deployment-guide.md}` が repo docs/ 体系の外に独立して存在し、`docs/ops/infra-architecture.md` と主題が重複する疑い。統合か分担明文化かの判断が必要。

### AnimalEkarte CSV import — USER actions

- [x] **方針 (2026-07-15):** フル 003_demo（~529MB・PHI 含みうる）は **Git に載せない**。正本バックアップは `old_db/sensitive-local/animalekarte-003-demo-full/`。リポジトリの `003_demo` は小さいデモのまま。
- [ ] **USER:** ローカルでフル seed を使う: `rsync -a ../old_db/sensitive-local/animalekarte-003-demo-full/ backend/migrations/seeds/003_demo/` のあと `make reset`（エージェントは reset しない）。誤 `git add` 防止のため該当 CSV に `skip-worktree` 推奨。
- [ ] **USER:** STG へのフル seed 適用は別途承認・手動実行（通常は小さいデモのまま）。
