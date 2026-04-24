# BE 業務系（非マスタ）コード規約違反 完全スキャン

## 目的
BEマスタ以外（業務系）のコードが「handler/service/repository の責任分離・冗長コード排除・統一実装・統一命名」の
規約に準拠しているかを体系的に検査するためのチェックリストだ。

下記チェックリスト × 対象ファイルリストの**全組み合わせ**を検査し、
PASS/FAIL を表で出力せよ。

**新パターンの発見・起票は禁止。チェックリストに定義された18パターンのみを報告する。**

**対象外（マスタ関連は `tmp/BE/be_master_check.md` を使用）:**
- `/v1/masters/*` エンドポイント（マスタ管理系）
- LINE予約設定・予約スケジュール・予約スタッフ系
- マスタ専用 service / repository（vaccine, medicine, procedure, cage 等）

---

## チェックリスト（固定・18パターン）

### ■ 責任分離 (Responsibility Separation)

#### P1: FindByID before Delete/Update（service）
Delete/Update メソッドの先頭（バリデーション前）で FindByID を呼んでいるか？
- 違反例: `repo.Update(...)` を先に呼んでから `FindByID` を呼ぶ / FindByID なしで Update
- 正しい例: `FindByID → validate → buildFields → repo.Update`
- 対象: 「Service」リストのファイル全件（Delete/Update メソッドのみ）

#### P4: Update/Upsert の clinicScope（repository）
UPDATE/UPSERT クエリに `Scopes(clinicScope(clinicID))` があるか？
- 違反例: `db.Model(...).Where("id = ?", id).Updates(fields)`（clinicScope なし）
- 正しい例: `db.Model(...).Scopes(clinicScope(clinicID)).Where("id = ?", id).Updates(fields)`
- **例外（P4 は `-` とする）**: `clinic_repository.go` / `company_repository.go` / `account_repository.go` / `password_reset_token_repository.go` / `audit_repository.go` はテナント外テーブル。clinicScope 不要。
- 対象: 「Repository」リストのファイル全件（Update/Upsert メソッドのみ）

#### P9: apperrors.FromGORM（repository）
GORM エラーが `apperrors.FromGORM(err, "resource", id)` で変換されているか？
- 違反例: `return nil, err`
- 正しい例: `return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))`
- 対象: 「Repository」リストのファイル全件（全エラーリターン）

#### P10: FK依存チェック before Delete（service）
service の Delete で依存チェック（`CountUsageBy*` → count > 0 → `WrapConflict`）があるか？
- 違反例: `FindByID → repo.Delete()` の直行（依存チェックなし）
- 正しい例: `FindByID → CountUsage → count > 0 なら WrapConflict → repo.Delete`
- **注意**: 依存関係のない末端エンティティ（Treatment, Vital 等）は依存チェック不要のため `-` とする。FK で参照されているかで判断すること。
- 対象: 「Service」リストのファイル全件（Delete メソッドのみ）

#### P14: handler → repository 直接呼び出し禁止
handler が repository を直接インジェクション・呼び出ししていないか？
- 違反例: handler struct に `repo XxxRepository` フィールド / `h.repo.FindByID(...)`
- 正しい例: handler struct に `svc XxxService` のみ。DB アクセスは必ず service 経由
- **注意**: `Handler` 構造体は `repos *repository.Repositories` を保持しているが、各ハンドラメソッドが `h.repos.Xxx` を直接呼び出していれば違反。`h.svc.*` 経由であれば OK。
- 対象: 「Handler」リストのファイル全件

---

### ■ 冗長コード排除 (Redundancy Elimination)

#### P2: CountUsage の deleted_at IS NULL（repository）
CountBy*/CountUsage* メソッドの WHERE 句に `deleted_at IS NULL` があるか？
- 違反例: `Where("owner_id = ?", ownerID)`（IS NULL なし）
- 正しい例: `Where("owner_id = ? AND deleted_at IS NULL", ownerID)`
- 対象: 「Repository」リストのファイル全件（CountBy*/CountUsage* メソッドのみ。対象メソッドがなければ `-`）

#### P3: Preload の deleted_at IS NULL（repository）
`gorm.DeletedAt` を持つエンティティの Preload に `"deleted_at IS NULL"` 条件があるか？

ソフトデリート対象エンティティ（全42件 — model/ 配下で gorm.DeletedAt を持つ struct）:
`Account`, `StaffClinicAssignment`, `Billing`, `BillingItem`, `Payment`,
`Cage`, `Checkup`, `CheckupType`, `ChiefComplaintType`, `ClinicalPlan`,
`Consultation`, `DiagnosisType`, `DiagnosisName`, `Estimate`, `EstimateItem`,
`Examination`, `ExaminationType`, `HospitalizationPlan`, `Hospitalization`, `TreatmentPlan`,
`InquiryTemplate`, `Insurance`, `InventoryItem`, `MedicalRecord`, `Medicine`,
`MerchandiseItem`, `Occupation`, `Owner`, `PaymentMethodMaster`, `PermissionGroup`,
`Pet`, `Procedure`, `Reservation`, `ReservationType`, `ReservationTypeGroup`,
`Staff`, `ShiftTemplate`, `Treatment`, `TrimmingCourse`, `TrimmingOption`,
`Vaccination`, `Vaccine`

- 違反例: `Preload("Owner")` / `Preload("Pet")` / `Preload("Doctor")`（条件なし）
- 正しい例: `Preload("Owner", "deleted_at IS NULL")`
- 判定方法: 上記エンティティ名が `Preload("XxxName")` の第1引数に含まれていたら、第2引数 `"deleted_at IS NULL"` の有無を確認する
- **注意: `Preload("Doctor")` は `Staff` モデルへのエイリアス。必ず対象に含める**
- 対象: 「Repository」リストのファイル全件

#### P7: toXxxResponse() 変換（handler）
handler が c.JSON でモデルを直接返していないか？
- 違反例: `c.JSON(http.StatusOK, entity)`
- 正しい例: `c.JSON(http.StatusOK, toEntityResponse(entity))`
- 対象: 「Handler」リストのファイル全件（全 c.JSON 呼び出し）

---

### ■ 統一された実装方法 (Unified Implementation)

#### P5: RequirePermission（書き込み系 routes）
POST/PUT/PATCH/DELETE ルートに RequirePermission が設定されているか？
- 違反例: `protected.POST("/owners", h.CreateOwner)`（パーミッションチェックなし）
- 正しい例: `owners.POST("", h.RequirePermission(string(model.ResourceOwners), "create"), h.CreateOwner)`
- **免除（P5 は `-` とする）**:
  - 認証エンドポイント: `/login`, `/logout`, `/auth/refresh`, `/auth/forgot-password`, `/auth/reset-password`
  - 自己操作: `/users/me/password`, `/me`
  - LIFF公開API（JWT認証なし・LINE IDトークン認証のエンドポイント群）
- 対象: 「Routes」リストのファイル全件

#### P6: DELETE ルートは "delete" パーミッション（routes）
DELETE ルートに `RequirePermission(resource, "delete")` が設定されているか？（"edit" ではなく "delete"）
- 違反例: `h.RequirePermission(string(model.ResourceOwners), "edit")` を DELETE に使用
- 正しい例: `h.RequirePermission(string(model.ResourceOwners), "delete")`
- **免除**: P5 と同じ免除対象
- 対象: 「Routes」リストのファイル全件

#### P8: apperrors.Wrap（service）
service 内の全エラーリターンが `apperrors.Wrap` で包まれているか？
- 違反例: `return nil, err`（Wrap なし）
- 正しい例: `return nil, apperrors.Wrap(err, "failed to update owner")`
- 対象: 「Service」リストのファイル全件（全エラーリターン）

#### P11: slog.ErrorContext on error paths（service）
repository 層から伝播したエラー（DB障害・インフラ起因）のリターン前に `slog.ErrorContext(ctx, "...", "error", err)` があるか？
- 違反例: `return nil, apperrors.Wrap(err, "...")`のみ（ログなし）
- 正しい例: `slog.ErrorContext(ctx, "failed to ...", "error", err); return nil, apperrors.Wrap(...)`
- **除外（ログ不要）**: `WrapInvalidInput`・NotFound の存在確認・`WrapConflict`（ユーザー起因の正常フロー）
- 対象: 「Service」リストのファイル全件（`s.repo.*` 呼び出しが返したエラーのリターン箇所のみ）

#### P12: ShouldBindJSON 統一処理（handler）
`ShouldBindJSON` エラーが統一形式で処理されているか？
- 違反例: `c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})`
- 正しい例: `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))`
- 対象: 「Handler」リストのファイル全件（全 ShouldBindJSON 呼び出し）

#### P15: POST は 201 + Location ヘッダ（handler）
Create 系ハンドラが `http.StatusCreated(201)` と `Location` ヘッダを返しているか？
- 違反例1: `c.JSON(http.StatusOK, toXxxResponse(entity))`（200）
- 違反例2: `c.JSON(http.StatusCreated, toXxxResponse(entity))`（Location なし）
- 正しい例: `c.Header("Location", fmt.Sprintf("/api/v1/{resource}/%d", entity.ID)); c.JSON(http.StatusCreated, toXxxResponse(entity))`
- **注意**: 業務系は `/api/v1/{resource}/%d` を使用（マスタ系の `/v1/masters/` とは異なる）
- 対象: 「Handler」リストのファイル全件（Create 系メソッドのみ）

---

### ■ 統一された命名規則 (Unified Naming)

#### P13: const/buildFunc 定義順序（service）
service ファイルで `buildXxxUpdateFields` と `const colXxx` が interface 定義より**前**に配置されているか？

正しい順序:
```
1. const (colXxx = "xxx")
2. func buildXxxUpdateFields(...)
3. type XxxService interface
4. type xxxService struct
5. func NewXxxService(...)
6. func (s *xxxService) メソッド群
```
- 対象: 「Service」リストのファイル全件

#### P16: メソッド名統一（repository）
repository メソッド名がプロジェクト標準に統一されているか？
- 一覧取得: `FindAll` / `FindByClinicID`（`GetAll`, `List`, `Fetch` は違反）
- 単件取得: `FindByID`（`GetByID`, `Get`, `Find` は違反）
- 作成: `Create`
- 更新: `Update`
- 削除: `Delete`
- カウント: `CountBy{Xxx}` または `CountUsageBy{Xxx}`
- 対象: 「Repository」リストのファイル全件（interface + service から呼ばれるメソッド名）

#### P17: Input 構造体命名統一（service）
service の Input 構造体が `CreateXxxInput` / `UpdateXxxInput` の命名規則に従っているか？
- 違反例: `XxxCreateRequest`, `CreateXxxParams`, `XxxInput`
- 正しい例: `CreateOwnerInput`, `UpdateMedicalRecordInput`
- 対象: 「Service」リストのファイル全件

#### P18: toXxxResponse 関数名統一（handler）
handler のレスポンス変換関数が `toXxxResponse` / `toXxxListResponse` の命名規則に従っているか？
- 違反例: `convertToXxx`, `buildXxxResponse`, `mapXxx`, `newXxxResponse`
- 正しい例: `toOwnerResponse`, `toMedicalRecordListResponse`
- 対象: 「Handler」リストのファイル全件

---

## 対象ファイルリスト（全件）

### Service（P1, P8, P10, P11, P13, P17 を検査）
- backend/internal/service/account_service.go
- backend/internal/service/accounting_report_service.go
- backend/internal/service/accounting_service.go
- backend/internal/service/appointment_admin_service.go
- backend/internal/service/appointment_notification_service.go
- backend/internal/service/appointment_service.go
- backend/internal/service/audit_service.go
- backend/internal/service/billing_confirmation_service.go
- backend/internal/service/billing_item_service.go
- backend/internal/service/billing_service.go
- backend/internal/service/care_plan_item_service.go
- backend/internal/service/cash_register_service.go
- backend/internal/service/checkup_service.go
- backend/internal/service/clinic_service.go
- backend/internal/service/clinical_plan_service.go
- backend/internal/service/company_service.go
- backend/internal/service/daily_record_service.go
- backend/internal/service/estimate_service.go
- backend/internal/service/examination_service.go
- backend/internal/service/hospitalization_service.go
- backend/internal/service/inquiry_service.go
- backend/internal/service/inventory_service.go
- backend/internal/service/liff_service.go
- backend/internal/service/line_customer_service.go
- backend/internal/service/line_messaging_service.go
- backend/internal/service/medical_record_image_service.go
- backend/internal/service/medical_record_service.go
- backend/internal/service/owner_service.go
- backend/internal/service/password_reset_service.go
- backend/internal/service/pet_service.go
- backend/internal/service/refund_service.go
- backend/internal/service/reservation_service.go
- backend/internal/service/shift_entry_service.go
- backend/internal/service/staff_clinic_assignment_service.go
- backend/internal/service/treatment_plan_service.go
- backend/internal/service/treatment_service.go
- backend/internal/service/trimming_service.go
- backend/internal/service/vaccination_service.go
- backend/internal/service/vital_service.go

### Repository（P2, P3, P4, P9, P16 を検査）
> P3 は全リポジトリが対象。P2/P4 はメソッドが存在する場合のみ検査（なければ `-`）。
- backend/internal/repository/account_repository.go
- backend/internal/repository/accounting_repository.go
- backend/internal/repository/appointment_admin_repository.go
- backend/internal/repository/appointment_repository.go
- backend/internal/repository/audit_repository.go
- backend/internal/repository/billing_confirmation_repository.go
- backend/internal/repository/billing_item_repository.go
- backend/internal/repository/care_plan_item_repository.go
- backend/internal/repository/cash_register_close_repository.go
- backend/internal/repository/checkup_repository.go
- backend/internal/repository/clinic_repository.go
- backend/internal/repository/clinical_plan_repository.go
- backend/internal/repository/company_repository.go
- backend/internal/repository/daily_record_repository.go
- backend/internal/repository/estimate_repository.go
- backend/internal/repository/examination_repository.go
- backend/internal/repository/hospitalization_repository.go
- backend/internal/repository/inquiry_repository.go
- backend/internal/repository/inventory_repository.go
- backend/internal/repository/line_customer_repository.go
- backend/internal/repository/medical_record_image_repository.go
- backend/internal/repository/medical_record_repository.go
- backend/internal/repository/owner_repository.go
- backend/internal/repository/password_reset_token_repository.go
- backend/internal/repository/pet_repository.go
- backend/internal/repository/refund_repository.go
- backend/internal/repository/reservation_repository.go
- backend/internal/repository/shift_entry_repository.go
- backend/internal/repository/treatment_plan_repository.go
- backend/internal/repository/treatment_repository.go
- backend/internal/repository/trimming_repository.go
- backend/internal/repository/vaccination_repository.go
- backend/internal/repository/vital_repository.go

### Handler（P7, P12, P14, P15, P18 を検査）
- backend/internal/handler/accounting_handler.go
- backend/internal/handler/accounting_report_handler.go
- backend/internal/handler/appointment_admin_handler.go
- backend/internal/handler/appointment_handler.go
- backend/internal/handler/auth_handler.go
- backend/internal/handler/billing_confirmation_handler.go
- backend/internal/handler/billing_item_handler.go
- backend/internal/handler/care_plan_item_handler.go
- backend/internal/handler/cash_register_handler.go
- backend/internal/handler/checkup_handler.go
- backend/internal/handler/clinic_handler.go
- backend/internal/handler/clinical_plan_handler.go
- backend/internal/handler/company_handler.go
- backend/internal/handler/daily_record_handler.go
- backend/internal/handler/estimate_handler.go
- backend/internal/handler/examination_handler.go
- backend/internal/handler/hospitalization_handler.go
- backend/internal/handler/inquiry_handler.go
- backend/internal/handler/inventory_handler.go
- backend/internal/handler/liff_handler.go
- backend/internal/handler/line_customer_handler.go
- backend/internal/handler/medical_record_handler.go
- backend/internal/handler/medical_record_image_handler.go
- backend/internal/handler/owner_handler.go
- backend/internal/handler/pet_handler.go
- backend/internal/handler/refund_handler.go
- backend/internal/handler/reservation_handler.go
- backend/internal/handler/shift_handler.go
- backend/internal/handler/treatment_handler.go
- backend/internal/handler/treatment_plan_handler.go
- backend/internal/handler/trimming_handler.go
- backend/internal/handler/vaccination_handler.go
- backend/internal/handler/vital_handler.go

### Routes（P5, P6 を検査）
> 以下のファイルに含まれる Register*Routes 関数 / register*RoutesWithAuth 関数を検査対象とする。
> 認証エンドポイント（/login, /logout, /auth/*, /me）および LIFF 公開 API は P5/P6 の検査対象外。
- backend/internal/handler/handler.go（registerOwnerRoutesWithAuth + registerMedicalRecordRoutesWithAuth + registerHospitalizationRoutesWithAuth + registerTrimmingRoutesWithAuth + registerExaminationRoutesWithAuth + registerVaccinationRoutesWithAuth + registerAccountingRoutesWithAuth + registerInventoryRoutesWithAuth + registerEstimateRoutesWithAuth）
- backend/internal/handler/pet_handler.go（RegisterPetRoutes）
- backend/internal/handler/reservation_handler.go（RegisterReservationRoutes）
- backend/internal/handler/vital_handler.go（RegisterVitalRoutes）
- backend/internal/handler/treatment_handler.go（RegisterTreatmentRoutes）
- backend/internal/handler/treatment_plan_handler.go（RegisterTreatmentPlanMedicalRecordRoutes, RegisterTreatmentPlanHospitalizationRoutes）
- backend/internal/handler/billing_item_handler.go（RegisterBillingItemRoutes）
- backend/internal/handler/billing_confirmation_handler.go（RegisterBillingConfirmationRoutes）
- backend/internal/handler/clinical_plan_handler.go（RegisterClinicalPlanRoutes）
- backend/internal/handler/checkup_handler.go（RegisterGlobalCheckupRoutes, RegisterCheckupRoutes）
- backend/internal/handler/inquiry_handler.go（RegisterInquiryRoutes）
- backend/internal/handler/daily_record_handler.go（RegisterDailyRecordRoutes）
- backend/internal/handler/care_plan_item_handler.go（RegisterCarePlanItemRoutes）
- backend/internal/handler/shift_handler.go（RegisterShiftRoutes）
- backend/internal/handler/company_handler.go（RegisterCompanyRoutes）
- backend/internal/handler/clinic_handler.go（RegisterClinicRoutes）
- backend/internal/handler/accounting_report_handler.go（RegisterAccountingReportRoutes）
- backend/internal/handler/medical_record_image_handler.go（RegisterMedicalRecordImageRoutes）
- backend/internal/handler/cash_register_handler.go（RegisterCashRegisterRoutes）

---

## 実行方法（AgentTeam 推奨）

以下の4チームで並列実行せよ。各チームは担当ファイルのみを読む。

| チーム | 担当パターン | 担当ファイル |
|--------|------------|------------|
| Team-Service | P1, P8, P10, P11, P13, P17 | 上記「Service」リスト |
| Team-Repository | P2, P3, P4, P9, P16 | 上記「Repository」リスト |
| Team-Handler | P7, P12, P14, P15, P18 | 上記「Handler」リスト |
| Team-Routes | P5, P6 | 上記「Routes」リスト（Register*Routes 関数を検査） |

---

## 出力フォーマット（必須）

| ファイル | P1 | P2 | P3 | P4 | P5 | P6 | P7 | P8 | P9 | P10 | P11 | P12 | P13 | P14 | P15 | P16 | P17 | P18 | 違反詳細 |
|---------|----|----|----|----|----|----|----|----|----|----|-----|-----|-----|-----|-----|-----|-----|-----|---------|
| owner_service.go | OK | - | - | - | - | - | - | FAIL | - | OK | - | OK | - | - | - | OK | - | P8:行112 Wrap未使用 |
| owner_repository.go | - | OK | FAIL | OK | - | - | - | - | OK | - | - | - | - | - | OK | - | - | P3:行88 Preload("Pet")に IS NULL なし |

凡例:
- `OK` = 問題なし
- `FAIL` = 違反あり（違反詳細列にファイル名:行番号と内容を必ず記載）
- `-` = 該当パターンなし（このファイルに対象メソッドが存在しない）

---

## 禁止事項（遵守必須）

1. **新パターンの発見・起票禁止** — P1〜P18 以外の問題を見つけても記録しない
2. **推測判定禁止** — 必ずファイルを Read してから判定する。コードを読まずに OK/FAIL を出力しない
3. **曖昧出力禁止** — 「〜かもしれない」「要確認」は使わない。`OK` か `FAIL` かのみ
4. **ファイル追加禁止** — 上記リスト外のファイルをスキャンしない
5. **スキャン中の即時起票禁止** — 全ファイルスキャン完了後に PASS/FAIL 表と違反サマリを出力してから起票する
6. **スキップ禁止** — ファイルリストの全件を読むこと

---

## 完了条件

1. 上記全ファイル × 全パターンの PASS/FAIL 表が出力される
2. FAIL セルの一覧をまとめた「違反サマリ」を出力する
3. `docs/tasks/open/code-quality/` と `docs/tasks/closed/code-quality/` の既存タスクタイトルと照合し、**未起票の違反のみ**を新規タスクとして `docs/tasks/open/code-quality/` に起票する（タスク番号は既存の最大番号+1から採番）
