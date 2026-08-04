package model

import (
	"encoding/json"
	"time"
)

// AuditLog は権限変更・認証操作の記録。削除禁止テーブル。
type AuditLog struct {
	ID uint64 `gorm:"primaryKey"   json:"id"`
	// ClinicID は実DDLで bigint NOT NULL REFERENCES clinics(id)（X-3）。Go 型は *uint64 のまま維持する
	// — サービス層 validateAuditLog（audit_service.go）が永続化前に非nil/非ゼロを検証する経路を持ち、
	// 呼び出し元が検証前の構造体を一時的に組み立てる余地を残すため（DB制約が最終防衛線）。
	ClinicID   *uint64         `gorm:"not null"     json:"clinic_id"`
	ActorID    *uint64         `json:"actor_id"`
	ActorType  string          `gorm:"not null"     json:"actor_type"`
	Action     string          `gorm:"not null"     json:"action"`
	Resource   string          `gorm:"not null"     json:"resource"`
	ResourceID *uint64         `json:"resource_id"`
	OldValue   json.RawMessage `gorm:"type:jsonb"   json:"old_value"`
	NewValue   json.RawMessage `gorm:"type:jsonb"   json:"new_value"`
	// Metadata は LSTEP 操作の件数・抽出条件を保存する多次元メタデータ（ISSUE-010）。
	// resource_id 単一 ID では表現できない情報（例: 健診対象抽出のフィルタ条件 + 件数集計）を JSON で永続化する。
	Metadata json.RawMessage `gorm:"type:jsonb"   json:"metadata"`
	// IPAddress は実DDLで inet NULL（X-3）。空文字列は `''::inet` として 22P02 になるため、
	// 未設定/空を表す値は Go の nil として保持する（*string、"" ではない）。
	IPAddress *string   `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// 監査アクション定数
const (
	AuditActorTypeStaff  = "staff"
	AuditActorTypeSystem = "system"

	AuditActionPermissionGroupCreate        = "permission_group.create"
	AuditActionPermissionGroupUpdate        = "permission_group.update"
	AuditActionPermissionGroupDelete        = "permission_group.delete"
	AuditActionPermissionRulesUpdate        = "permission_rules.update"
	AuditActionStaffPermissionGroupsReplace = "staff.permission_groups.replace"
	AuditActionAuthLoginSuccess             = "auth.login.success"
	AuditActionAuthLoginFailure             = "auth.login.failure"
	AuditActionAuthLogout                   = "auth.logout"
	AuditActionAuthPasswordChange           = "auth.password.change"
	AuditActionAuthPasswordReset            = "auth.password.reset"
	AuditActionAuthPasswordAdminReplace     = "auth.password.admin_replace"

	// Lステップ / LINE連携 監査アクション
	AuditActionLstepSettingsSave     = "lstep.settings.save"
	AuditActionLstepTagSync          = "lstep.tag.sync"
	AuditActionLstepTagSyncBulk      = "lstep.tag.sync_bulk"
	AuditActionLineNotificationSend  = "line.notification.send"
	AuditActionOwnerLineUserIDUpdate = "owner.line_user_id.update"
	AuditActionOwnerLineUserIDUnlink = "owner.line_user_id.unlink"
	AuditActionReservationNoShow     = "reservation.no_show.auto"

	// 取扱説明書（マニュアル）編集 監査アクション
	AuditActionManualArticleUpsert = "manual_article.upsert"
	AuditActionManualArticleDelete = "manual_article.delete"

	// トリミング予約の臨床変更監査（BUG-422）
	AuditActionTrimmingCreate = "trimming.create"
	AuditActionTrimmingUpdate = "trimming.update"
	AuditActionTrimmingDelete = "trimming.delete"

	// 会計・返金 監査アクション（#122）
	AuditActionBillingCancel        = "billing.cancel"
	AuditActionBillingPostCloseEdit = "billing.post_close_edit"
	AuditActionBillingRefundCreate  = "billing_refund.create"
	// #189: 確定済み会計のクレジット（カード）金額の確定後訂正
	AuditActionBillingCreditCorrection = "billing.credit_correction"
	// BUG-440: 明細削除による予防接種 claim 解放（再取込可能化）の actor 監査
	AuditActionBillingVaccinationClaimRelease = "billing.vaccination_claim_release"

	// #201 薬量自動計算 監査アクション
	// dose パラメータ変更（作成/更新/削除）・per_weight 有効化・著しい逸脱上書き
	AuditActionMedicineDoseParamUpsert = "medicine_dose_param.upsert"
	AuditActionMedicineDoseParamDelete = "medicine_dose_param.delete"
	AuditActionMedicinePerWeightEnable = "medicine.per_weight.enable"
	AuditActionTreatmentDoseDeviation  = "treatment.dose.deviation"

	// lab import 監査アクション（Phase 4A）
	AuditActionLabImportPreviewRequested = "lab_import.preview.requested"
	AuditActionLabImportCommitRequested  = "lab_import.commit.requested"
	AuditActionLabImportCommitSucceeded  = "lab_import.commit.succeeded"
	AuditActionLabImportCommitFailed     = "lab_import.commit.failed"
	AuditActionLabImportSourceBlocked    = "lab_import.source.blocked"
	// AuditActionLabImportRevertSucceeded records compensating revert (TASK-032). Distinct from commit.
	AuditActionLabImportRevertSucceeded = "lab_import.revert.succeeded"

	// #211 健診結果値の置換（既存削除を伴う PUT）監査アクション
	AuditActionCheckupFieldResultReplace = "checkup_field_result.replace"
	// TASK-374 / #211: 健診 package の versioned clinic-scoped import apply
	AuditActionCheckupPackageImportApply = "checkup_package_import.apply"

	// BE-refactor.md R1-2 (D1): 検査結果値（exam_results）の置換（既存削除を伴う PUT）監査アクション。
	// checkup_field_result と同型の tx 内 fail-closed 監査。
	AuditActionExamResultReplace = "exam_result.replace"
	// #249 / DEC-53: parent examination mutation は authenticated actor と before/after を
	// mutation と同じ transaction で記録する。confirm は update と分離して状態遷移を識別する。
	AuditActionExaminationCreate    = "examination.create"
	AuditActionExaminationUpdate    = "examination.update"
	AuditActionExaminationConfirm   = "examination.confirm"
	AuditActionExaminationUnconfirm = "examination.unconfirm"
	AuditActionExaminationDelete    = "examination.delete"

	// BUG-010 residual: clinical plan (診察所見・診断詳細・治療方針) の versioned update 監査。
	// examination parent mutation と同型で before/after を同一 tx に書く。
	AuditActionClinicalPlanUpdate = "clinical_plan.update"
	// BUG-010 residual (delete): clinical_plans の破壊的削除（GORM soft-delete）でも
	// 法的臨床フィールドの削除前値を同一 DBOrTx に fail-closed で残す（MRA-01 と同型）。
	AuditActionClinicalPlanDelete = "clinical_plan.delete"

	AuditActionPetOwnerReplace                     = "pet_owner.replace"
	AuditActionHospitalizationDischargeWithBilling = "hospitalization.discharge_with_billing"

	// #239 identity link 手動 link/unlink（PHI を載せない ID のみ）
	AuditActionOwnerIdentityLinkCreate = "owner_identity_link.create"
	AuditActionOwnerIdentityLinkAdd    = "owner_identity_link.add_members"
	AuditActionOwnerIdentityLinkUnlink = "owner_identity_link.unlink"
	AuditActionPetIdentityLinkCreate   = "pet_identity_link.create"
	AuditActionPetIdentityLinkAdd      = "pet_identity_link.add_members"
	AuditActionPetIdentityLinkUnlink   = "pet_identity_link.unlink"

	// #255 staff batch provisioning（PII 非搭載: batch_id / digest / count / external_staff_id のみ）
	AuditActionStaffProvisionCreate  = "staff.provision.create"
	AuditActionStaffProvisionReceipt = "staff.provision.receipt"
)

// audit_logs.resource 定数
const (
	AuditResourceAccount   = "account"
	AuditResourceStaff     = "staff"
	AuditResourceLabImport = "lab_import"
	// #201 薬量自動計算
	AuditResourceMedicineDoseParam = "medicine_dose_param"
	AuditResourceMedicine          = "medicine"
	AuditResourceTreatmentDose     = "treatment_dose"
	// #211 健診パッケージ型付き結果値の置換（既存削除を伴う）監査
	AuditResourceCheckupFieldResult = "checkup_field_result"
	// TASK-374 / #211: 健診 package import provenance / apply 監査
	AuditResourceCheckupPackageImport = "checkup_package_import"
	// BE-refactor.md R1-2: 検査結果値の置換（既存削除を伴う）監査
	AuditResourceExamResult = "exam_result"
	// #249 / DEC-53: parent exams row の create/update/confirm/delete 監査。
	AuditResourceExamination = "examination"
	// BUG-010 residual: clinical_plans 行の更新監査。
	AuditResourceClinicalPlan = "clinical_plan"

	AuditResourceReservation     = "reservation"
	AuditResourceHospitalization = "hospitalization"
	AuditResourceTrimming        = "trimming"
	// U-X01X03-MR-CARE / MRA-01: hard-delete care plan items need durable audit resource.
	AuditResourceCarePlanItem = "care_plan_item"
	// #239 identity link groups（owner / pet 共通 resource 名）
	AuditResourceIdentityLink = "identity_link"
	// #255 staff batch provisioning receipt（clinic-scoped, PII-free）
	AuditResourceStaffProvisionBatch = "staff_provision_batch"
)

// care plan item destructive actions (MRA-01)
const (
	AuditActionCarePlanItemDelete = "care_plan_item.delete"
)

// LabBlockedReason は source_blocked 監査イベントの reason フィールドに使用できる
// 許可された値のみを表す型。free-form string は使用不可。
// 新しい reason を追加する場合はこのファイルに定数を追加すること。
type LabBlockedReason string

const (
	// LabBlockedReasonMDBSchemaUnconfirmed は drwan ソースの MDB スキーマが未確認のためブロックされた場合。
	LabBlockedReasonMDBSchemaUnconfirmed LabBlockedReason = "mdb_schema_not_confirmed"
	// LabBlockedReasonSourceNotImplemented は該当ソース種別が未実装のためブロックされた場合。
	LabBlockedReasonSourceNotImplemented LabBlockedReason = "source_not_implemented"
	// LabBlockedReasonSourceTypeBlocked は該当ソース種別がポリシーによりブロックされた場合。
	LabBlockedReasonSourceTypeBlocked LabBlockedReason = "source_type_blocked"
)

// validLabBlockedReasons は許可されたすべての LabBlockedReason 値のセット。
// 新しい定数を追加したらここにも追加すること。
var validLabBlockedReasons = map[LabBlockedReason]struct{}{
	LabBlockedReasonMDBSchemaUnconfirmed: {},
	LabBlockedReasonSourceNotImplemented: {},
	LabBlockedReasonSourceTypeBlocked:    {},
}

// ValidLabBlockedReason は r が許可済み LabBlockedReason 定数のいずれかと一致するか検証する。
// 任意の文字列キャスト (model.LabBlockedReason("arbitrary")) は false を返す。
func ValidLabBlockedReason(r LabBlockedReason) bool {
	if r == "" {
		return false
	}
	_, ok := validLabBlockedReasons[r]
	return ok
}

// LabAuditErrorCategory は commit_failed 監査イベントの error_category フィールドに使用できる
// 許可された値のみを表す型。free-form string は使用不可。
// 新しいカテゴリを追加する場合はこのファイルに定数を追加すること。
type LabAuditErrorCategory string

const (
	// LabAuditErrorCategoryInvalidInput はリクエストの入力値が不正な場合。
	LabAuditErrorCategoryInvalidInput LabAuditErrorCategory = "invalid_input"
	// LabAuditErrorCategoryNotFound はリソースが見つからない / スコープ外の場合。
	LabAuditErrorCategoryNotFound LabAuditErrorCategory = "not_found"
	// LabAuditErrorCategoryForbidden は操作が権限上禁止されている場合。
	LabAuditErrorCategoryForbidden LabAuditErrorCategory = "forbidden"
	// LabAuditErrorCategoryUnauthorized は認証されていない場合。
	LabAuditErrorCategoryUnauthorized LabAuditErrorCategory = "unauthorized"
	// LabAuditErrorCategoryInternal はその他の内部エラーの場合。
	LabAuditErrorCategoryInternal LabAuditErrorCategory = "internal"
)

// validLabAuditErrorCategories は許可されたすべての LabAuditErrorCategory 値のセット。
// 新しい定数を追加したらここにも追加すること。
var validLabAuditErrorCategories = map[LabAuditErrorCategory]struct{}{
	LabAuditErrorCategoryInvalidInput: {},
	LabAuditErrorCategoryNotFound:     {},
	LabAuditErrorCategoryForbidden:    {},
	LabAuditErrorCategoryUnauthorized: {},
	LabAuditErrorCategoryInternal:     {},
}

// ValidLabAuditErrorCategory は c が許可済み LabAuditErrorCategory 定数のいずれかと一致するか検証する。
// 任意の文字列キャスト (model.LabAuditErrorCategory("arbitrary")) は false を返す。
func ValidLabAuditErrorCategory(c LabAuditErrorCategory) bool {
	if c == "" {
		return false
	}
	_, ok := validLabAuditErrorCategories[c]
	return ok
}
