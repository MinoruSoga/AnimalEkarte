# BE マスタ系 コード規約違反 完全スキャン

## 目的
BEマスタ関連コードが「handler/service/repository の責任分離・冗長コード排除・統一実装・統一命名」の
規約に準拠しているかを体系的に検査するためのチェックリストだ。

下記チェックリスト × 対象ファイルリストの**全組み合わせ**を検査し、
PASS/FAIL を表で出力せよ。

**新パターンの発見・起票は禁止。チェックリストに定義された18パターンのみを報告する。**

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
- 対象: 「Repository - マスタ系」リストのファイル全件（Update/SwapSortOrder/Upsert メソッドのみ）

#### P9: apperrors.FromGORM（repository）
GORM エラーが `apperrors.FromGORM(err, "resource", id)` で変換されているか？
- 違反例: `return nil, err`
- 正しい例: `return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))`
- 対象: 「Repository - マスタ系」リストのファイル全件（全エラーリターン）

#### P10: FK依存チェック before Delete（service）
service の Delete で `CountUsageBy{ID}` → count > 0 なら `WrapConflict` のチェックがあるか？
- 違反例: `FindByID → repo.Delete()` の直行（依存チェックなし）
- 正しい例: `FindByID → CountUsage → count > 0 なら WrapConflict → repo.Delete`
- 対象: 「Service」リストのファイル全件（Delete メソッドのみ）

#### P14: handler → repository 直接呼び出し禁止
handler が repository を直接インジェクション・呼び出ししていないか？
- 違反例: handler struct に `repo XxxRepository` フィールド / `h.repo.FindByID(...)`
- 正しい例: handler struct に `svc XxxService` のみ。DB アクセスは必ず service 経由
- 対象: 「Handler」リストのファイル全件

---

### ■ 冗長コード排除 (Redundancy Elimination)

#### P2: CountUsage の deleted_at IS NULL（repository）
CountBy*/CountUsage* メソッドの WHERE 句に `deleted_at IS NULL` があるか？
- 違反例: `Where("vaccine_id = ?", vaccineID)`（IS NULL なし）
- 正しい例: `Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID)`
- 対象: 「Repository - マスタ系」リストのファイル全件

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

- 違反例: `Preload("ReservationType")` / `Preload("Doctor")`（条件なし）
- 正しい例: `Preload("ReservationType", "deleted_at IS NULL")`
- 判定方法: 上記エンティティ名が `Preload("XxxName")` の第1引数に含まれていたら、第2引数 `"deleted_at IS NULL"` の有無を確認する
- **注意: `Preload("Doctor")` は `Staff` モデルへのエイリアス。必ず対象に含める**
- 対象: 「Repository - マスタ系」リスト全件 ＋ 「Repository - 非マスタ系」リスト全件

#### P7: toXxxResponse() 変換（handler）
handler が c.JSON でモデルを直接返していないか？
- 違反例: `c.JSON(http.StatusOK, entity)`
- 正しい例: `c.JSON(http.StatusOK, toEntityResponse(entity))`
- 対象: 「Handler」リストのファイル全件（全 c.JSON 呼び出し）

---

### ■ 統一された実装方法 (Unified Implementation)

#### P5: RequirePermission（書き込み系 routes）
POST/PUT/PATCH/DELETE ルートに RequirePermission が設定されているか？
- 違反例: `masters.POST("/...", h.Create)`（パーミッションチェックなし）
- 正しい例: `masters.POST("/...", RequirePermission("edit"), h.Create)` または `perm(resource, "create")`
- **注意: `staff_handler.go:RegisterMasterRoutes` は `perm := func(...) { return h.RequirePermission(...) }` というローカルヘルパーを使用。`perm(...)` も等価で OK**
- 対象: 「Routes」リストのファイル全件

#### P6: DELETE ルートは "delete" パーミッション（routes）
DELETE ルートに `RequirePermission("delete")` または `perm(resource, "delete")` が設定されているか？
- 違反例: `perm(resource, "edit")` を DELETE に使用
- 正しい例: `perm(resource, "delete")`
- 対象: 「Routes」リストのファイル全件

#### P8: apperrors.Wrap（service）
service 内の全エラーリターンが `apperrors.Wrap` で包まれているか？
- 違反例: `return nil, err`（Wrap なし）
- 正しい例: `return nil, apperrors.Wrap(err, "failed to update staff")`
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
- 正しい例: `c.Header("Location", fmt.Sprintf("/v1/masters/{resource}/%d", entity.ID)); c.JSON(http.StatusCreated, toXxxResponse(entity))`
- 参照実装: `animal_species_handler.go`, `vaccine_handler.go`
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
- 対象: 「Repository - マスタ系」リストのファイル全件（interface + service から呼ばれるメソッド名）

#### P17: Input 構造体命名統一（service）
service の Input 構造体が `CreateXxxInput` / `UpdateXxxInput` の命名規則に従っているか？
- 違反例: `XxxCreateRequest`, `CreateXxxParams`, `XxxInput`
- 正しい例: `CreateVaccineInput`, `UpdateStaffInput`
- 対象: 「Service」リストのファイル全件

#### P18: toXxxResponse 関数名統一（handler）
handler のレスポンス変換関数が `toXxxResponse` / `toXxxListResponse` の命名規則に従っているか？
- 違反例: `convertToXxx`, `buildXxxResponse`, `mapXxx`, `newXxxResponse`
- 正しい例: `toVaccineResponse`, `toVaccineListResponse`
- 対象: 「Handler」リストのファイル全件

---

## 対象ファイルリスト（全件）

### Service（P1, P8, P10, P11, P13, P17 を検査）
- backend/internal/service/staff_service.go
- backend/internal/service/procedure_service.go
- backend/internal/service/vaccine_service.go
- backend/internal/service/checkup_type_service.go
- backend/internal/service/cage_service.go
- backend/internal/service/diagnosis_service.go
- backend/internal/service/permission_group_service.go
- backend/internal/service/payment_method_master_service.go
- backend/internal/service/trimming_course_service.go
- backend/internal/service/trimming_option_service.go
- backend/internal/service/shift_template_service.go
- backend/internal/service/reservation_type_service.go
- backend/internal/service/reservation_type_group_service.go
- backend/internal/service/reservation_type_liff_service.go
- backend/internal/service/closing_settings_service.go
- backend/internal/service/clinic_holiday_service.go
- backend/internal/service/animal_species_service.go
- backend/internal/service/chief_complaint_service.go
- backend/internal/service/exam_type_service.go
- backend/internal/service/medicine_service.go
- backend/internal/service/occupation_service.go
- backend/internal/service/inquiry_template_service.go
- backend/internal/service/merchandise_item_service.go
- backend/internal/service/insurance_service.go
- backend/internal/service/consultation_service.go
- backend/internal/service/hospitalization_plan_service.go
- backend/internal/service/reservation_staff_service.go
- backend/internal/service/line_reservation_setting_service.go
- backend/internal/service/reservation_schedule_service.go

### Repository - マスタ系（P2, P3, P4, P9, P16 を検査）
- backend/internal/repository/vaccine_repository.go
- backend/internal/repository/medicine_repository.go
- backend/internal/repository/procedure_repository.go
- backend/internal/repository/checkup_type_repository.go
- backend/internal/repository/cage_repository.go
- backend/internal/repository/clinic_holiday_repository.go
- backend/internal/repository/reservation_staff_repository.go
- backend/internal/repository/staff_repository.go
- backend/internal/repository/animal_species_repository.go
- backend/internal/repository/chief_complaint_repository.go
- backend/internal/repository/exam_type_repository.go
- backend/internal/repository/trimming_course_repository.go
- backend/internal/repository/trimming_option_repository.go
- backend/internal/repository/shift_template_repository.go
- backend/internal/repository/reservation_type_repository.go
- backend/internal/repository/reservation_type_group_repository.go
- backend/internal/repository/reservation_type_liff_repository.go
- backend/internal/repository/diagnosis_repository.go
- backend/internal/repository/permission_group_repository.go
- backend/internal/repository/payment_method_master_repository.go
- backend/internal/repository/inquiry_template_repository.go
- backend/internal/repository/occupation_repository.go
- backend/internal/repository/insurance_repository.go
- backend/internal/repository/merchandise_item_repository.go
- backend/internal/repository/consultation_repository.go
- backend/internal/repository/hospitalization_plan_repository.go
- backend/internal/repository/closing_special_period_repository.go
- backend/internal/repository/clinic_settings_repository.go
- backend/internal/repository/line_reservation_setting_repository.go
- backend/internal/repository/reservation_schedule_repository.go
- backend/internal/repository/reservation_type_occupation_repository.go
- backend/internal/repository/reservation_type_unavailable_time_repository.go
- backend/internal/repository/staff_clinic_assignment_repository.go

### Repository - 非マスタ系（P3 のみ検査）
> P3 はマスタ以外のリポジトリも Preload 経由でソフトデリート対象エンティティを読み込む。
> マスタ系の P3 は Team-Repository-Master が担当するため、**マスタ系リストに含まれないファイルのみ**を記載する。
- backend/internal/repository/appointment_admin_repository.go
- backend/internal/repository/hospitalization_repository.go
- backend/internal/repository/examination_repository.go
- backend/internal/repository/checkup_repository.go
- backend/internal/repository/reservation_repository.go
- backend/internal/repository/vaccination_repository.go
- backend/internal/repository/medical_record_repository.go
- backend/internal/repository/treatment_repository.go
- backend/internal/repository/treatment_plan_repository.go
- backend/internal/repository/daily_record_repository.go
- backend/internal/repository/clinical_plan_repository.go
- backend/internal/repository/care_plan_item_repository.go
- backend/internal/repository/estimate_repository.go
- backend/internal/repository/billing_item_repository.go
- backend/internal/repository/trimming_repository.go
- backend/internal/repository/owner_repository.go
- backend/internal/repository/pet_repository.go
- backend/internal/repository/appointment_repository.go
- backend/internal/repository/vital_repository.go
- backend/internal/repository/accounting_repository.go
- backend/internal/repository/shift_entry_repository.go
- backend/internal/repository/line_customer_repository.go
- backend/internal/repository/inventory_repository.go

### Handler（P7, P12, P14, P15, P18 を検査）
- backend/internal/handler/staff_handler.go
- backend/internal/handler/closing_settings_handler.go
- backend/internal/handler/clinic_holiday_handler.go
- backend/internal/handler/animal_species_handler.go
- backend/internal/handler/chief_complaint_handler.go
- backend/internal/handler/exam_type_handler.go
- backend/internal/handler/medicine_handler.go
- backend/internal/handler/procedure_handler.go
- backend/internal/handler/trimming_course_handler.go
- backend/internal/handler/trimming_option_handler.go
- backend/internal/handler/cage_handler.go
- backend/internal/handler/checkup_type_handler.go
- backend/internal/handler/vaccine_handler.go
- backend/internal/handler/diagnosis_handler.go
- backend/internal/handler/permission_group_handler.go
- backend/internal/handler/payment_method_master_handler.go
- backend/internal/handler/reservation_type_handler.go
- backend/internal/handler/reservation_type_group_handler.go
- backend/internal/handler/reservation_type_liff_handler.go
- backend/internal/handler/shift_template_handler.go
- backend/internal/handler/insurance_handler.go
- backend/internal/handler/occupation_handler.go
- backend/internal/handler/inquiry_template_handler.go
- backend/internal/handler/merchandise_item_handler.go
- backend/internal/handler/consultation_handler.go
- backend/internal/handler/hospitalization_plan_handler.go
- backend/internal/handler/reservation_staff_handler.go
- backend/internal/handler/line_reservation_setting_handler.go
- backend/internal/handler/reservation_schedule_handler.go

### Routes（P5, P6 を検査）
> マスタ関連ルートは各ハンドラファイル内の `Register*Routes` 関数に直接定義されている。
- backend/internal/handler/staff_handler.go（RegisterMasterRoutes — 全 /v1/masters/* ルートを包含）
- backend/internal/handler/payment_method_master_handler.go（RegisterPaymentMethodMasterRoutes）
- backend/internal/handler/closing_settings_handler.go（RegisterClosingSettingsRoutes）
- backend/internal/handler/clinic_holiday_handler.go（RegisterClinicHolidayRoutes）
- backend/internal/handler/shift_template_handler.go（RegisterShiftTemplateRoutes）
- backend/internal/handler/reservation_line_routes.go（RegisterLineReservationRoutes / RegisterLiffRoutes）

---

## 実行方法（AgentTeam 推奨）

以下の5チームで並列実行せよ。各チームは担当ファイルのみを読む。

| チーム | 担当パターン | 担当ファイル |
|--------|------------|------------|
| Team-Service | P1, P8, P10, P11, P13, P17 | 上記「Service」リスト |
| Team-Repository-Master | P2, P3, P4, P9, P16 | 上記「Repository - マスタ系」リスト |
| Team-Repository-Preload | P3 | 上記「Repository - 非マスタ系」リスト |
| Team-Handler | P7, P12, P14, P15, P18 | 上記「Handler」リスト |
| Team-Routes | P5, P6 | 上記「Routes」リスト（Register*Routes 関数を検査） |

---

## 出力フォーマット（必須）

| ファイル | P1 | P2 | P3 | P4 | P5 | P6 | P7 | P8 | P9 | P10 | P11 | P12 | P13 | P14 | P15 | P16 | P17 | P18 | 違反詳細 |
|---------|----|----|----|----|----|----|----|----|----|----|-----|-----|-----|-----|-----|-----|-----|-----|---------|
| staff_service.go | FAIL | - | - | - | - | - | - | OK | - | OK | FAIL | - | OK | - | - | - | OK | - | P1:行337 Update前にFindByIDなし / P11:行362 ErrorContext欠落 |
| vaccine_repository.go | - | FAIL | - | OK | - | - | - | - | FAIL | - | - | - | - | - | - | OK | - | - | P2:行98 IS NULL欠落 / P9:行45 FromGORM未使用 |

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
