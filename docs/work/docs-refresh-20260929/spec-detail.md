# DOCS-REFRESH-20260929 — spec-detail

作業場所: `/private/tmp/ae-dr29-spec-detail`（branch `docs-refresh-20260929/spec-detail`）
対象: `docs/spec/screens/20*`–`40*`、`99-medical-record-flow.md`、`common-dialogs.md`、`settings/**`

## 対象と判定

| ファイル | 判定 | 主な根拠 |
|---|---|---|
| `20-master-settings.md` | kept | マスタ一覧・権限表は `backend/internal/*/routes*.go`・`settings-routes.tsx` と一致 |
| `21-login.md` | kept | `AuthProvider`・`is_system_admin`・ログインフローは `frontend/src/features/auth/` と一致 |
| `22-estimate-list.md` / `23-estimate-form.md` / `26-estimate-detail.md` | kept | `frontend/src/features/estimates/`・見積 API と一致 |
| `24-shift-calendar.md` | kept | `ShiftCalendarPage`/`ShiftCell`/`ShiftFormDialog`/`ClinicHolidayModal`・`/api/v1/shifts`・`clinic-holidays` は `frontend/src/features/shifts/`・`backend/internal/staff/`・`backend/internal/clinic/` と一致 |
| `25-checkups-list.md` | kept | `/checkups` ガードは `ResourceCheckups`、`select-pet`/`new` は `medical-records` create+edit の二重ガード（`clinical-care-routes.tsx:367-422`）と一致。健診パッケージ要件は承認待ちソースのまま未確定扱いを維持 |
| `27-inventory-form.md` | kept | `inventory.go` モデル（`is_inventory` 逆フラグなし）と一致 |
| `28-line-reservation.md` | kept | FE ガード乖離（`ResourceReservations` vs `ResourceHospitalSettings` API）、`/line-reservation/slots` は `ResourceMasterReservationType` と一致（`operations-routes.tsx:73-114`） |
| `29-closing-aggregation.md` | kept | `resolvePeriodRange`・`GetCloseAggregate`・`BuildAllocationBillings`・`RefundParentWeights`・`sumRefundsForCompletedBillings` は `backend/internal/billing/` に実在。closing 時刻デフォルト（09:00/14:00/18:30/17:30）は `001_init.sql:2059-2061`・`clinic_settings.go` と一致。12:00/18:30 の記述は歴史的 PO 決定として維持 |
| `30-unpaid-list.md` | kept | `unpaid`/`unpaid-balance`/`unpaid-period` ルート・`latest_scheduled`（EMR-189）・`group_by=monthly`→`period` エイリアス（`unpaid-tab-model.ts`）と一致 |
| `31-lstep-integration.md` | **corrected** | `lstep-tag-config` 変異系権限を `hospital-settings` → システム管理者専用に訂正（`backend/internal/lstep/routes.go:52-65,176-185`） |
| `32-accounting-reports.md` | kept | `/reports/monthly`・`/reports/monthly/csv` は `accounting_report_handler.go`・`ResourceAccountingReports` と一致。締め後編集は `ResourceAccountingPostCloseEdit`+`post_close_reason` 必須と一致 |
| `34-lstep-delivery-monitor.md` | kept | `delivery-monitor/summary|logs` は `ResourceLstepAnalytics:view`、`staleTime` 1分・ポーリングなしと一致（`lstep_delivery_monitor_handler.go`・`get-lstep-delivery-trigger-*.ts`） |
| `35-internal-manual.md` | kept | `manual_articles` テーブル（`001_init.sql:2864`）・`ResourceManualEdit` view/edit/delete・`features/manual/content` + `import.meta.glob` + fuse.js と一致 |
| `36-aggregation-dashboard.md` | kept | `GET /owners/aggregations` は `owners:view`、`CalculateCPMStage`（V1）/`CalculateCPMStageV2`（Lステップ側のみ）の記述は `lstep/aggregation_handler.go`・`lstep_tag_sync_service.go` と一致 |
| `37-line-reserve-owner-flow.md` | kept | 全ページコンポーネント・`getStepProgress`・`useReservationFlow`・`ValidateAndCreate`・`liff_validation.go` 制限値・`CancelByID`・`filterCalendarAppointments`・`useFetchState`/`resolveFetchError`/`useLiff`（`shared-liff/`）と一致 |
| `38-liff-pet-health.md` | **corrected** | レートリミットを「読み取り系 30回/分」→ 認証済み系 60回/分に訂正（`backend/internal/reservation/routes.go:167-188`） |
| `39-owner-report.md` | kept | `/owners/:id/report/pets`（`owners:view`）・`/pets/:id/first-visit`・`/pets/:id/treatment-history`（`medical-records:view`）と一致（`pet/routes.go`・`medicalrecord/routes_records.go`） |
| `40-identity-links.md` | kept | api.yaml の identity-links 11 パス 13 ルート・`identity-links:view/edit` リソースと一致 |
| `99-medical-record-flow.md` | kept | `/inquiries`・`/vitals`・`/clinical-plan`・`/treatments`（PUT bulk は edit）・`/addenda` の権限表は `routes_records.go:33-68` と一致 |
| `common-dialogs.md` | kept | `StaffSelectionModal`（`features/medical-records/components/`）等の参照先は実在 |
| `settings/master-*.md` 各種 | kept | `masters/staffs`（`ResourceMasterStaff`）・`occupations`・`shift-templates`（`ResourceShifts`）・`reservation-types`（`available-slots` POST=`edit`/DELETE=`delete`）・`inquiry-templates`/`interview/templates` 両ルート・`medicine dose-params`（`ResourceMasterMedical`）・`medicine.inventory_id` は API のみ対応・UI 未公開の記述と一致。payment-methods の ADR-003 固定キー seed（`payment_methods.csv`）・reservation-types seed（診察/お手入れ/ワクチン/健診、`reservation_visible=false`）・closing-time デフォルトとも一致 |

## 実施した訂正（抜粋・重要度順）

- `docs/spec/screens/31-lstep-integration.md`: `POST|DELETE /api/v1/lstep-tag-config/{auto-managed-prefixes,condition-tag-mappings,send-purpose-tag-prefixes}` の必須権限 `hospital-settings` create/delete → **システム管理者のみ**（`requireSystemAdmin()`・RBAC 非対象）。GET は `hospital-settings:view` のまま。SOLO-09 / LSA-04 / DEC-30 の意図的閉鎖（根拠: `backend/internal/lstep/routes.go:52-65,176-185`）。日付付き訂正注記を表直下に追加。
- `docs/spec/screens/38-liff-pet-health.md`: レートリミット「読み取り系は 30回/分」→「認証済みエンドポイント（`/health-card` 含む）は 60回/分」。30回/分は公開 `/settings` と `/my-reservations` のみ（根拠: `backend/internal/reservation/routes.go:167-188`）。同日付訂正。

## 外部依存・BLOCKED（未解消のまま残すもの）

- **docs-symbol-drift が fail のまま**: `FAIL テーブル数: docs の宣言値 128 が実装の実測値 130 と不一致` ×3（検査トークン 608 件）。原因は `docs/architecture/erd.md`・`docs/spec/specification.md` の宣言値で、両ファイルは本エージェントの編集許可範囲外（他子エージェント担当）。本スコープでは修正不能のため BLOCKED として報告。
- **健診パッケージ要件（`14_健診パッケージ_クライアント要件_2026-08-24.md`）**: 承認待ちのため実装確定ソース扱い不可。`25-checkups-list.md` は現行実装（`ResourceCheckups` + `medical-records` create/edit ガード）の記述のみとし、要件を反映した表現への言及・変更は行わない（UNKNOWN/外部依存として維持）。
- **Plane 状態**: 読み取り専用方針のため本セッションでは参照していない。実行状態が不明な項目は UNKNOWN。

## 提案削除候補（理由・後継リンク）

なし（本スコープ内に削除相当の陳腐化ファイルは確認されず）。

## 検証

- `git diff --check -- docs/spec` → exit 0（whitespace エラーなし）
- `bash scripts/check-docs-symbol-drift.sh` → `NG docs-symbol-drift: 3 件のドリフト（検査トークン 608 件）`（ベースラインと同一。全件 allowlist 外ファイル起因）
- `git status --porcelain` → `M docs/spec/screens/31-lstep-integration.md`、`M docs/spec/screens/38-liff-pet-health.md`、?? `docs/work/docs-refresh-20260929/spec-detail.md` のみ
- allowlist 外差分: **なし**
