package lintscan

// dbortx_inventory_lint_test.go — BE-refactor follow-up (ambient-transaction harness): dbOrTx 参加メソッドの inventory ゲート
//
// ─── Background ──────────────────────────────────────────────────────────────────────
//
// repository メソッドが `dbOrTx(ctx, r.db)` を使うと、Transactor.WithTx / repos.Transaction の
// ambient transaction 内から呼ばれた場合に自動で同一 tx に参加する（無ければ base db にフォールバック
// = `r.db.WithContext(ctx)` と等価）。R1-1/R1-2（accounting/refund/billing_item 等）は、WithTx 内で
// 呼ばれる読取/書込が `r.db.WithContext` 直参照で tx 非参加だった部分コミット/TOCTOU バグを、この
// dbOrTx への統一で塞いだ。
//
// 本ゲートは「ambient tx 参加が意図されたメソッド集合」を inventory として固定する:
//   - 固定メソッドが必要な参加形（DBOrTx / 明示helper / TxFromContext）をやめた
//     （`r.db.WithContext` へ revert 等）→ tx 参加の regression → fail。
//   - 未登録の新規メソッドが dbOrTx を使い始めた → allowlist 追加を強制 → レビュー時に「ambient tx
//     参加が正しいか・atomicity/isolation テストを添えたか」を必ず問う。
//
// ─── 意図的にやらないこと（taint 解析の限界・#124 の教訓） ────────────────────────────
//
// 「WithTx 内で呼ばれるのに dbOrTx を使っていない（= 参加漏れ）メソッド」の検出は、service→repository を
// 跨ぐ手続き間データフロー解析（どの repo メソッドが WithTx クロージャ内で呼ばれるか）が必須で、go/ast
// 単体では信頼できる規則が書けない（master_fk_write_inventory_lint と同じ taint 断念）。よって本ゲートは
// 「明示された ambient-tx 参加 surface の固定と regression 検出」に絞る。参加漏れの正本ガードは各 tx フローの
// atomicity テスト（accounting_repository_tx_atomicity_test.go / refund_repository_sum_tx_participation_test.go
// / checkup_field_result_tx_atomicity_test.go 等）が担う。
//
// ─── Static scanning blind spots (BE9-1) ───────────────────────────────────────────
//
// This gate is a syntactic match over literal AST call shapes (see `funcUsesDBOrTx` and
// `ambientTxParticipationExpectations` below). It cannot see:
//   - a DBOrTx-equivalent participation hidden behind a raw SQL string, or a renamed/aliased
//     import of the DBOrTx helper (shadowing/aliasing is a documented existing limitation, not
//     new to BE9-1);
//   - an arbitrary method that SHOULD participate in an ambient tx but is neither a literal
//     DBOrTx user nor registered with an explicit helper/required-tx expectation (the
//     taint-analysis limitation already documented above);
//   - any dbOrTx-shaped call reachable only through a background job, cron, worker, or other
//     code path that is not represented as a literal AST call shape in a plain .go file this
//     scanner visits.
// The tx atomicity/isolation runtime tests referenced throughout this file's allowlist comments
// remain required for exactly these blind spots; this static inventory gate is a complement
// layered on top of them, not a substitute.
//
// ─── Technique ──────────────────────────────────────────────────────────────────────
//
// preload_clinic_scope_lint_test.go の moduleInternalSource / legacyLintKey（内部で
// internal/lintscan.WalkInternalTreeT を使用。BE9-1・旧 repoSourceFS go:embed を置換）と
// baseFileName / receiverMethodKey（audit_tx_inventory_lint_test.go）を再利用。モジュール全体の
// internal/ 配下（internal/repository 以外の internal/service・internal/model 等も含む）を走査し、
// 各 FuncDecl 本体に DBOrTx 呼び出し、または明示登録された同等以上の参加形があるものを
// (keyFile | ReceiverType.Method) で列挙し、allowlist と双方向突合する。
// internal/repository/** 由来のキーは legacyLintKey により旧
// repoSourceFS 相当の形（basename / 1階層以上の相対パス）へ正規化され、既存 allowlist のキー形が
// そのまま一致し続ける。
//
// 注（syntactic match・sibling lint と同一設計）: `funcUsesDBOrTx` は
//   1) Ident `dbOrTx`（legacy wrapper shape; reintroduction regression detection）
//   2) Ident `DBOrTx`（same-package free name inside persistence）
//   3) Selector `persistence.DBOrTx`（canonical shared kernel）
// を検出する。go/types の意味解決はしない。シャドーイングや別名 import は誤検知/見逃しの既知限界
// （preload/audit-tx lint と同じ割り切り）。
//
// ambient Reorder 方針: `Reorder` がローカル helper / `persistence.ReorderByClinicID` のみを
// 呼ぶ経路はメソッド本体に `dbOrTx`/`DBOrTx` が現れない。参加保証の正本は
// `persistence.ReorderByClinicID`/`ReorderGlobal` 内の `DBOrTx`（free func・本 inventory の
// receiver 走査外）であり、TestDBOrTxInventory_ReorderHelpersUseDBOrTx で固定する。
// ドメイン Reorder を method inventory に載せる必要はない（allowlist 膨張と偽回帰を避ける）。

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// dbOrTxParticipatingMethods は現worktreeで ambient transaction に参加する repository
// メソッドを固定する（key = "<file> | <ReceiverType>.<Method>"）。大半は DBOrTx を直接使い、
// 明示された例外は ambientTxParticipationExpectations の同等以上の形を使う。
// R1-1/R1-2 の tx 参加 surface を含む。
// 追加/削除時はこのマップを更新し、新規は対応する atomicity/isolation テストを添えること。
var dbOrTxParticipatingMethods = map[string]struct{}{
	// auth persistence (BE9 auth Phase 1): global account/reset/blacklist data and
	// clinic-scoped permission groups participate in the caller's ambient transaction.
	"auth/account_repository.go|accountRepository.Create":                                                 {},
	"auth/account_repository.go|accountRepository.CompareAndSwapPasswordHash":                             {},
	"auth/account_repository.go|accountRepository.FindByEmail":                                            {},
	"auth/account_repository.go|accountRepository.FindByID":                                               {},
	"auth/account_repository.go|accountRepository.FindByIDForUpdate":                                      {},
	"auth/account_repository.go|accountRepository.UpdatePasswordHash":                                     {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.Create":                         {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.ConsumeByID":                    {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.DeleteByAccountID":              {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.DeleteByID":                     {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.DeleteIssued":                   {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindLatestByAccountIDForUpdate": {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindByTokenHash":                {},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindByTokenHashForUpdate":       {},
	"auth/permission_group_repository.go|permissionGroupRepository.CountUsageByGroupID":                   {},
	"auth/permission_group_repository.go|permissionGroupRepository.Create":                                {},
	"auth/permission_group_repository.go|permissionGroupRepository.CreateWithRules":                       {},
	"auth/permission_group_repository.go|permissionGroupRepository.Delete":                                {},
	"auth/permission_group_repository.go|permissionGroupRepository.DeleteSoftDeletedByClinicID":           {},
	"auth/permission_group_repository.go|permissionGroupRepository.FindAll":                               {},
	"auth/permission_group_repository.go|permissionGroupRepository.FindAllEffectivePermissionsByStaffID":  {},
	"auth/permission_group_repository.go|permissionGroupRepository.FindAllGroupIDsByStaffID":              {},
	"auth/permission_group_repository.go|permissionGroupRepository.FindByID":                              {},
	"auth/permission_group_repository.go|permissionGroupRepository.LockByIDForUpdate":                     {},
	"auth/permission_group_repository.go|permissionGroupRepository.Reorder":                               {},
	"auth/permission_group_repository.go|permissionGroupRepository.Update":                                {},
	"auth/permission_group_repository.go|permissionGroupRepository.UpdateWithRules":                       {},
	"auth/permission_group_repository.go|permissionGroupRepository.UpdateRules":                           {},
	"auth/permission_group_repository.go|permissionGroupRepository.UpdateStaffGroups":                     {},
	"auth/permission_group_repository.go|permissionGroupRepository.replaceRules":                          {},
	"auth/token_blacklist_repository.go|tokenBlacklistRepository.Create":                                  {},
	"auth/token_blacklist_repository.go|tokenBlacklistRepository.DeleteExpired":                           {},
	"auth/token_blacklist_repository.go|tokenBlacklistRepository.ExistsByJTI":                             {},
	// accounting (R1-1 money-path atomicity; appointment completion moved to the reservation
	// write owner in BE9-2E-0 while retaining ambient transaction participation)
	"billing/accounting_repository.go|accountingRepository.Create":             {},
	"billing/accounting_repository.go|accountingRepository.FindByID":           {}, // commit-before-reload must observe the caller's ambient writes
	"billing/accounting_repository.go|accountingRepository.FindByIDForClinics": {}, // same multi-clinic read contract as FindByID
	"billing/accounting_repository.go|accountingRepository.LockAndFindByID":    {},
	"billing/accounting_repository.go|accountingRepository.SavePayment":        {},
	"billing/accounting_repository.go|accountingRepository.SavePaymentSplits":  {},
	"billing/accounting_repository.go|accountingRepository.Update":             {},
	// BUG-018 idempotency probe: the completion-request key must be looked up inside the
	// caller's ambient transaction so a replay cannot observe a pre-commit gap and create a
	// second billing for the same key. Unscoped + ClinicScope keeps soft-deleted keys reserved.
	"billing/accounting_repository.go|accountingRepository.FindByCompletionRequestID": {},
	// Audit writes deliberately require an already-open ambient transaction and call
	// persistence.TxFromContext directly. The explicit expectation below prevents weakening
	// this fail-closed contract back to fallback DBOrTx behavior.
	"audit/repository.go|repository.CreateTx": {},
	// billing_confirmation (SD-2 系ガード監査: 会計医師確認 Confirm/Return が確定済みカルテ書込
	// ガード対象と判明。billingConfirmationService.Confirm/Return の LockByIDForUpdate ambient tx
	// に参加させる)
	// Confirm/Return actor lock, GetOrCreate, and update share one transaction so
	// an unassigned actor cannot create a pending row before authorization.
	// Runtime: TestBillingConfirmationService_RuntimeActorIsolation and
	// TestBillingConfirmationRepository_LockActiveStaffAssignment.
	"billing/billing_confirmation_repository.go|billingConfirmationRepository.Create":                    {},
	"billing/billing_confirmation_repository.go|billingConfirmationRepository.FindByMedicalRecordID":     {},
	"billing/billing_confirmation_repository.go|billingConfirmationRepository.LockActiveStaffAssignment": {},
	"billing/billing_confirmation_repository.go|billingConfirmationRepository.Update":                    {},
	// billing_item (R1-1)
	"billing/billing_item_repository.go|billingItemRepository.Create":          {},
	"billing/billing_item_repository.go|billingItemRepository.Delete":          {},
	"billing/billing_item_repository.go|billingItemRepository.FindByBillingID": {},
	"billing/billing_item_repository.go|billingItemRepository.FindByID":        {},
	"billing/billing_item_repository.go|billingItemRepository.Update":          {},
	// BUG-506 create runs only inside AccountingService.WithTx and requires the ambient tx.
	// Runtime: billing_item_reference_repository_test.go and accounting Complete tests.
	"billing/billing_item_service.go|billingItemService.createItemInAmbientTx": {},
	// UpdateBillingTotals is a thin wrapper; ambient-tx body is updateBillingTotals.
	"billing/billing_item_repository.go|billingItemRepository.updateBillingTotals": {},
	// Create validates every request-derived FK under shared locks in the same
	// transaction. Runtime: billing_item_reference_repository_test.go.
	"billing/billing_item_repository.go|billingItemRepository.ValidateCreateReferences":           {},
	"billing/billing_item_repository.go|billingItemRepository.ValidateVaccinationCreateReference": {},
	// AE-LAB exam billing FK validation. TxFromContext fail-closed + exam_types SHARE lock.
	// Runtime: TestBillingItemRepository_ValidateExamCreateReference_FailClosedWithoutAmbientTx
	// and TestBillingItemRepository_ValidateExamCreateReferenceLocksExamTypeInAmbientTransaction.
	"billing/billing_item_exam.go|billingItemRepository.ValidateExamCreateReference": {},
	// campaign
	"billing/campaign_repository.go|campaignRepository.FindAllApplicableForItem": {}, // BE8-4 batch9: moved from campaign_repository.go
	"billing/campaign_repository.go|campaignRepository.FindApplicableForItem":    {}, // BE8-4 batch9: moved from campaign_repository.go
	"billing/campaign_repository.go|campaignRepository.ReplaceTargets":           {}, // BE8-4 batch9: moved from campaign_repository.go; G6-2 repo-internal tx replace
	// BE-X06-BIL-CAMPAIGN-01 / BIL-03: Update+ReplaceTargets+FindByID share ambient tx.
	// Runtime: campaign_repository_tx_atomicity_test.go
	"billing/campaign_repository.go|campaignRepository.Create":   {},
	"billing/campaign_repository.go|campaignRepository.Delete":   {},
	"billing/campaign_repository.go|campaignRepository.FindAll":  {},
	"billing/campaign_repository.go|campaignRepository.FindByID": {},
	"billing/campaign_repository.go|campaignRepository.Update":   {},
	// BE-X06-LSTEP-SETTINGS-01 / LSA-06: settings write graph joins ambient tx.
	// Runtime: lstep_settings_tx_atomicity_test.go
	"lstep/lstep_settings_repository.go|lstepSettingsRepository.FindByClinicAndService":   {},
	"lstep/lstep_settings_repository.go|lstepSettingsRepository.Upsert":                   {},
	"lstep/lstep_settings_repository.go|lstepSettingsRepository.DeleteByClinicAndService": {},
	// Runtime: TestLstepSettingsRepository_DeleteByClinicServiceAndKey_RollsBackWhenAmbientTxFails
	"lstep/lstep_settings_repository.go|lstepSettingsRepository.DeleteByClinicServiceAndKey": {},
	// Runtime: TestLstepSettingsRepository_FindCredentialByClinicServiceKey_SeesUncommittedUpsert
	"lstep/lstep_settings_repository.go|lstepSettingsRepository.FindCredentialByClinicServiceKey":    {},
	"lstep/lstep_sync_settings_repository.go|lstepSyncSettingsRepository.FindByClinicID":             {},
	"lstep/lstep_sync_settings_repository.go|lstepSyncSettingsRepository.Upsert":                     {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.FindByClinicID":                   {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.Save":                             {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.UpdateCPMVersion":                 {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.UpdateDormantThresholds":          {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.UpdateCPMV2Thresholds":            {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.UpdateCPMV1Thresholds":            {},
	"clinic/clinic_settings_repository.go|clinicSettingsRepository.UpdateHealthPreventionThresholds": {},
	// U-X01X05-CLINIC / POC-02+POC-05: special-period reads/writes join ambient tx via DBOrTx.
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.FindAll":      {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.FindByID":     {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.FindByDate":   {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.Create":       {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.Update":       {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.Delete":       {},
	"clinic/closing_special_period_repository.go|closingSpecialPeriodRepository.CheckOverlap": {},

	// daily_record: parent/clinic relation validation and audit-coupled writes must
	// remain on the service-owned ambient transaction.
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.CreateCareLog":                  {},
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.CreateStaffNote":                {},
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.CreateVitalRecord":              {}, // AUD-006: FindOrCreate+CreateVital same ambient tx
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.FindByHospitalizationID":        {},
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.FindByHospitalizationIDAndDate": {},
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.FindOrCreateByDate":             {}, // BE8-4 batch6: moved from daily_record_repository.go
	"medicalrecord/vital_repository.go|vitalRepository.FindByID":                                    {}, // response re-fetch must observe and govern the same tx mutation
	// treatment_plan (U-MR-TREATMENT-PLAN / MRD-02): Create + clinicScopeQuery join ambient tx
	// so service write+reload stays atomic under Transactor.WithTx.
	"medicalrecord/treatment_plan_repository.go|treatmentPlanRepository.Create":           {},
	"medicalrecord/treatment_plan_repository.go|treatmentPlanRepository.clinicScopeQuery": {},
	// care_plan_item / hospitalization (BE9-2D ⑤: DischargeWithBilling の repos.Transaction→
	// Transactor.WithTx 化。FOR UPDATE 直列化・退院status更新・care plan read を billing 書込と
	// 同一 ambient tx に参加させる＝二重会計防止。BE9-2E-0ではCreate/Updateのclinic/master
	// 検証も同一txへ収束。Create/FindByID/Updateのrollback proofは
	// TestHospitalizationRepository_CRUDParticipatesInAmbientTransaction。)
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.FindByHospitalizationID": {},
	// U-X01X03-MR-CARE / MRA-02: Create/Update/Delete/FindByID join ambient tx for write+reload + audited hard-delete.
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.FindByID":                  {},
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.Create":                    {},
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.Update":                    {},
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.Delete":                    {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.Create":                {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.FindByID":              {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.LockByIDForUpdate":     {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.Update":                {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.UpdateIfNotDischarged": {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.Delete":                {}, // MRB-05: delete + audit same ambient tx
	// checkup_field (#211 tx-internal replace)
	"medicalrecord/checkup_field_repository.go|checkupFieldResultRepository.FindByCheckupID":   {},
	"medicalrecord/checkup_field_repository.go|checkupFieldResultRepository.ReplaceForCheckup": {},
	// Checkup relation validation, mutation, and readback share the service tx.
	// Runtime: checkup_repository_tx_atomicity_test.go; clinic/relation coverage:
	// checkup_repository_test.go and checkup_service_test.go.
	"medicalrecord/checkup_repository.go|checkupRepository.Create":                {},
	"medicalrecord/checkup_repository.go|checkupRepository.Delete":                {},
	"medicalrecord/checkup_repository.go|checkupRepository.FindByClinicID":        {},
	"medicalrecord/checkup_repository.go|checkupRepository.FindByID":              {},
	"medicalrecord/checkup_repository.go|checkupRepository.FindByMedicalRecordID": {},
	"medicalrecord/checkup_repository.go|checkupRepository.FindByOwnerID":         {},
	// Required ambient tx + parent-before-child row lock for finalize/delete serialization.
	// Runtime: TestDB_CheckupRepository_ParentThenChildLocksSerializeMedicalRecordFinalization.
	"medicalrecord/checkup_repository.go|checkupRepository.LockByIDForUpdate": {},
	"medicalrecord/checkup_repository.go|checkupRepository.Update":            {},
	// clinical_plan (BUG-010 residual 929fef0fa / 90ee096bf): clinical plan Update and Delete
	// record a staff-actor audit entry fail-closed in the SAME transaction as the business
	// write, so every method on this repository must join the caller's ambient tx. The reads
	// (FindByMedicalRecordID) and the two unexported guards (existsInClinic, parentStillDraft)
	// participate as well: they re-check ownership and draft state from inside that tx, and a
	// non-ambient read would let a concurrent finalize slip between the guard and the write.
	// Runtime: TestClinicalPlanService_Update_AuditFailureRollsBackDB,
	// TestClinicalPlanService_Update_AuditSuccessCommitsDB.
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.Create":                {},
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.Delete":                {},
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.FindByMedicalRecordID": {},
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.Update":                {},
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.existsInClinic":        {},
	"medicalrecord/clinical_plan_repository.go|clinicalPlanRepository.parentStillDraft":      {},
	// estimate (SD-2 系ガード監査: 見積書 Create/Update/Delete が確定済みカルテ書込ガード対象と判明。
	// estimateService の LockByIDForUpdate ambient tx に参加させる。FindByID は
	// UpdateIfNotLocked/normalizeDeleteIfNotLockedMiss の tx 内再取得のため併せて追加)
	"billing/estimate_repository.go|estimateRepository.Create":                 {},
	"billing/estimate_repository.go|estimateRepository.FindByID":               {},
	"billing/estimate_repository.go|estimateRepository.UpdateIfNotLocked":      {},
	"billing/estimate_repository.go|estimateRepository.DeleteIfNotLocked":      {},
	"billing/estimate_repository.go|estimateRepository.CountItemsByEstimateID": {},
	"billing/estimate_repository.go|estimateRepository.ReplaceItems":           {},
	// Runtime: TestEstimateRepository_AllocateNextEstimateNo_SeesUncommittedInsertAndRollsBack
	// + TestEstimateRepository_AllocateNextEstimateNo_AdvisoryLockBlocksConcurrentSession
	"billing/estimate_repository.go|estimateRepository.AllocateNextEstimateNo": {},
	// W-013 cash register close writes/reads join ambient tx (post-close correction atomicity).
	// Runtime: cash_register_close_repository_tx_atomicity_test.go
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.Create":              {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.CreateAdjustment":    {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.FindAll":             {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.FindByID":            {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.HasCloseOnDate":      {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.FindByDateAndPeriod": {},
	"billing/cash_register_close_repository.go|cashRegisterCloseRepository.Void":                {},
	// examination (BE-refactor.md R1-2 tx-internal replace; Create/FindByID/Update added for X-11
	// finalize-child-write-race — must join the LockByIDForUpdate ambient tx or the FK check on
	// examinations.medical_record_id deadlocks against the FOR UPDATE row lock; Delete added for
	// H-8d — same finalize-lock race as Update, now WithTx-wrapped in examinationService.Delete)
	"medicalrecord/examination_repository.go|examinationRepository.CountItemsByExamID": {}, // Runtime: TestDB_ExaminationRepository_CountItemsByExamIDReadsAmbientTxAndRollsBack.
	"medicalrecord/examination_repository.go|examinationRepository.Create":             {},
	"medicalrecord/examination_repository.go|examinationRepository.Delete":             {},
	// FindAll/FindByJobID must observe relation writes made earlier in the service
	// tx. Runtime: examination_repository_relation_read_tx_test.go.
	"medicalrecord/examination_repository.go|examinationRepository.FindAll":              {},
	"medicalrecord/examination_repository.go|examinationRepository.FindAllItemsByExamID": {},
	"medicalrecord/examination_repository.go|examinationRepository.FindByID":             {},
	"medicalrecord/examination_repository.go|examinationRepository.FindByJobID":          {},
	// Required ambient tx lock serializes status/move/delete/result replacement.
	// Runtime: TestDB_ExaminationRepository_LockByIDForUpdateSerializesConcurrentStatusUpdate.
	"medicalrecord/examination_repository.go|examinationRepository.LockByIDForUpdate":    {},
	"medicalrecord/examination_repository.go|examinationRepository.ReplaceItemsByExamID": {},
	"medicalrecord/examination_repository.go|examinationRepository.Update":               {},

	// TASK-032: lab import job/event/receipt/retraction repos participate in ambient tx.
	// Lock/CAS methods require ambient tx (no base DB fallback).
	"medicalrecord/lab_import_repository.go|labImportJobRepository.Create":                            {},
	"medicalrecord/lab_import_repository.go|labImportJobRepository.Update":                            {},
	"medicalrecord/lab_import_repository.go|labImportJobRepository.FindByID":                          {},
	"medicalrecord/lab_import_repository.go|labImportJobRepository.LockByIDForUpdate":                 {},
	"medicalrecord/lab_import_repository.go|labImportJobRepository.CompareAndSetStatus":               {},
	"medicalrecord/lab_import_repository.go|labImportEventRepository.Create":                          {},
	"medicalrecord/lab_import_repository.go|labImportEventRepository.FindByJob":                       {},
	"medicalrecord/lab_import_repository.go|labImportEventRepository.HasEventType":                    {},
	"medicalrecord/lab_import_repository.go|labImportUsageReceiptRepository.Create":                   {},
	"medicalrecord/lab_import_repository.go|labImportUsageReceiptRepository.LockByJobForUpdate":       {},
	"medicalrecord/lab_import_repository.go|labImportUsageReceiptRepository.CountByJob":               {},
	"medicalrecord/lab_import_repository.go|labImportUsageReceiptRepository.CountManualMutationByJob": {},
	"medicalrecord/lab_import_repository.go|labImportRevertReceiptRepository.FindByIdempotencyKey":    {},
	"medicalrecord/lab_import_repository.go|labImportRevertReceiptRepository.LockByIdempotencyKey":    {},
	"medicalrecord/lab_import_repository.go|labImportRevertReceiptRepository.Create":                  {},
	"medicalrecord/lab_import_repository.go|labImportRetractionRepository.CreateWithItems":            {},
	"medicalrecord/lab_import_repository.go|LabImportDuplicateCheckerDB.IsDuplicate":                  {},
	// TASK-032 revert service helpers that read under ambient tx.
	"medicalrecord/lab_import_revert_service.go|labImportRevertService.lockLinkedExamsByJob": {},
	"medicalrecord/lab_import_revert_service.go|labImportRevertService.assertRevertSafe":     {},
	"medicalrecord/lab_import_revert_service.go|labImportRevertService.assertExamRelations":  {},
	"medicalrecord/lab_import_usage_tracker.go|labImportUsageTracker.RecordClinicalUse":      {},

	// TASK-027 Slice A: append, revision-only read, and status/pointer CAS must observe the
	// same service-owned transaction. Runtime rollback/visibility proof:
	// TestExaminationRevision_RepositoryMethodsParticipateInAmbientTransaction.
	"medicalrecord/examination_revision_repository.go|examinationRepository.AppendOfficialRevision": {},
	"medicalrecord/examination_revision_repository.go|examinationRepository.ConfirmWithRevisionCAS": {},
	"medicalrecord/examination_revision_repository.go|examinationRepository.FindOfficialByID":       {},
	// TASK-031: print snapshot is read-only revision load; ambient tx optional via DBOrTx.
	"medicalrecord/examination_print_snapshot.go|examinationRepository.FindPrintSnapshot": {},
	// TASK-374: clinic-scoped package import apply/preflight participate in ambient tx.
	"medicalrecord/checkup_package_import_service.go|checkupPackageImportService.Apply":                 {},
	"medicalrecord/checkup_package_import_service.go|checkupPackageImportService.preflightCollisions":   {},
	"medicalrecord/checkup_package_import_service.go|checkupPackageImportService.validateActorInClinic": {},
	// TASK-027 Slice B: unconfirm/edit/reconfirm revision writes and pointer CAS all fail closed
	// without the service-owned ambient transaction. Runtime rollback proof:
	// TestExaminationRevision_UnconfirmAuditFailureRollsBackWorkingRevisionAndPointer.
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendWorkingRevisionFromOfficial": {},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendWorkingRevisionFromCurrent":  {},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendOfficialRevisionFromWorking": {},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AdvanceRevisionCAS":                {},

	"medicalrecord/exam_type_repository.go|examTypeRepository.FindByID":                      {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.Create":                        {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.Update":                        {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.Delete":                        {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.Reorder":                       {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.CountUsageByExamTypeID":        {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.CountChildrenByParentID":       {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.AnimalSpeciesExists":           {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.CountExamResultsByFieldID":     {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.CountReferenceRangesByFieldID": {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.CreateField":                   {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.DeleteField":                   {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.FindReferenceRangesByFieldIDs": {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.LockFieldByID":                 {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.ReorderFields":                 {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.ReplaceReferenceRanges":        {},
	"medicalrecord/exam_type_repository.go|examTypeRepository.UpdateField":                   {},
	// medical_record_addendum
	"medicalrecord/medical_record_addendum_repository.go|medicalRecordAddendumRepository.Create":                {},
	"medicalrecord/medical_record_addendum_repository.go|medicalRecordAddendumRepository.FindByID":              {},
	"medicalrecord/medical_record_addendum_repository.go|medicalRecordAddendumRepository.FindByMedicalRecordID": {},
	// medical_record (X-11 Appendix-A finalize-child-write-race fix)
	// SD-2: 確定済みカルテ画像ガード — Create/Delete/FindByID が LockByIDForUpdate の
	// ambient tx に参加する（medical_record_image_service.go の WithTx 内から呼ばれる）。
	// BE9-2D sub-batch④a: moved to internal/medicalrecord (facade kept in internal/repository).
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.Create":   {},
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.Delete":   {},
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.FindByID": {},

	"medicalrecord/medical_record_repository.go|medicalRecordRepository.LockByIDForUpdate": {},
	// Auto-create holds one ambient transaction from the non-blocking advisory lock through the
	// clinic/pet/JST-day duplicate count and INSERT. Runtime proofs:
	// TestMedicalRecordRepository_AcquireAutoCreateLock_IsNonBlockingWhenContended and
	// TestMedicalRecordRepository_CountByPetAndDate_IsScopedAndJoinsAmbientTransaction.
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.AcquireAutoCreateLock":           {},
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.CountByPetAndDate":               {},
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.CountEstimatesByMedicalRecordID": {}, // delete/estimate creation serialization under the medical-record row lock
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.Create":                          {},
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.Delete":                          {}, // draft-only CAS soft delete must share finalization transactions
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.FindByAppointmentID":             {}, // appointment row-lock serialization proof in medicalrecord package
	// F-1: delete takes appointments FOR UPDATE before medical_records. Runtime:
	// TestMedicalRecordService_DeleteWaitsOnAppointmentRowLockBeforeInConsultationCommit
	// TestMedicalRecordRepository_LockLinkedAppointmentForUpdate_WaitsOnAppointmentRowLock
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.LockLinkedAppointmentForUpdate": {},
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.Update":                         {},
	"medicalrecord/medical_record_repository.go|medicalRecordRepository.findMedicalRecordByID":          {},
	// vaccination (BE9-2E-0: patient/master validation, readback, and writes must share the
	// service-owned transaction. Runtime proofs live in vaccination_transaction_concurrency_test.go;
	// clinic-scoped read/preload coverage lives in vaccination_clinic_isolation_test.go.)
	"medicalrecord/vaccination_repository.go|vaccinationRepository.Create":                      {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.Delete":                      {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.FindAll":                     {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.FindByID":                    {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.FindByOwner":                 {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.FindOwnersByVaccineDeadline": {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.LockByIDForUpdate":           {},
	"medicalrecord/vaccination_repository.go|vaccinationRepository.Update":                      {},
	"medicalrecord/vaccine_repository.go|vaccineRepository.FindByID":                            {},
	// BUG-425: tag-code replacement is one transaction; every replacement write must use DBOrTx.
	"lstep/lstep_tag_code_mapping_repository.go|lstepTagCodeMappingRepository.Create":                         {},
	"lstep/lstep_tag_code_mapping_repository.go|lstepTagCodeMappingRepository.SoftDelete":                     {},
	"lstep/lstep_tag_code_mapping_repository.go|lstepTagCodeMappingRepository.SoftDeleteByClinicIDAndTagName": {},
	// Public LINE account linking: token lock/CAS, owner row lock/update, and
	// fail-closed audit must all remain on the service-owned ambient transaction.
	// Runtime: line_link_token_repository_test.go,
	// line_link_transaction_integration_test.go, and
	// owner/repository_line_link_tx_test.go.
	"lstep/line_link_token_repository.go|lineLinkTokenRepository.Consume":              {},
	"lstep/line_link_token_repository.go|lineLinkTokenRepository.Create":               {},
	"lstep/line_link_token_repository.go|lineLinkTokenRepository.LockUsableByRawToken": {},
	// medicine_dose_param (R1-2 dose-param tx)
	"medicalrecord/medicine_dose_param_repository.go|medicineDoseParamRepository.Create":                   {},
	"medicalrecord/medicine_dose_param_repository.go|medicineDoseParamRepository.Delete":                   {},
	"medicalrecord/medicine_dose_param_repository.go|medicineDoseParamRepository.FindByMedicineAndSpecies": {},
	"medicalrecord/medicine_dose_param_repository.go|medicineDoseParamRepository.FindByMedicineID":         {},
	"medicalrecord/medicine_dose_param_repository.go|medicineDoseParamRepository.Update":                   {},
	// global animal_species master: audited Create/Update/Delete/Reorder share WithTx;
	// FindAll/FindByID also participate so Update's post-write reload and service pre-read
	// see ambient writes. Runtime:
	// TestAnimalSpeciesRepository_*_RollsBackWhenAmbientTxFails /
	// TestAnimalSpeciesService_*_AuditFailureRollsBack
	// (internal/pet/animal_species_repository_test.go, animal_species_service_test.go).
	// Reorder delegates to persistence.ReorderGlobal (DBOrTx inside helper; not inventoried).
	"pet/animal_species_repository.go|animalSpeciesRepository.FindAll":  {},
	"pet/animal_species_repository.go|animalSpeciesRepository.FindByID": {},
	"pet/animal_species_repository.go|animalSpeciesRepository.Create":   {},
	"pet/animal_species_repository.go|animalSpeciesRepository.Update":   {},
	"pet/animal_species_repository.go|animalSpeciesRepository.Delete":   {},
	// pet (BUG-407: lstepLifecycleService.HandlePetDeath/HandlePetRevival が status/deceased_at
	// 更新と一次監査ログ書込を Transactor.WithTx で束ね fail-closed 化。BE9-2Eで
	// internal/pet へ移動。runtime proof は internal/pet/repository_tx_atomicity_test.go)
	"pet/repository.go|repository.Update": {},
	// 死亡登録/復活の条件付き UPDATE（CAS）。期待 status を述語に含め RowsAffected==0 を 409 へ写像する
	// ため、呼び出し元の ambient transaction に参加して同一 snapshot 上で判定する必要がある。Runtime:
	// TestPetRepository_RecordDeath_ConcurrentRequestsPreserveWinner /
	// TestPetRepository_ClearDeath_ConcurrentRevivalRequests
	// (internal/pet/repository_tx_atomicity_test.go)
	"pet/repository.go|repository.RecordDeath": {},
	"pet/repository.go|repository.ClearDeath":  {},
	// pet_owners の clinic 相関 read/count と全置換は、後続 service の置換 + audit と同じ
	// ambient transaction に参加する。Runtime:
	// TestPetOwnerRepository_FindByPetID_AmbientTransaction /
	// TestPetOwnerRepository_ReplaceForPet_AmbientTransaction /
	// TestPetOwnerRepository_CountByOwnerID_AmbientTransaction
	// (internal/pet/pet_owner_repository_tx_atomicity_test.go) /
	// TestPetOwnerRepository_FindSharedPetsByOwnerID_AmbientTransaction
	// (internal/pet/pet_owner_repository_clinic_isolation_test.go)
	"pet/pet_owner_repository.go|petOwnerRepository.FindByPetID":             {},
	"pet/pet_owner_repository.go|petOwnerRepository.FindSharedPetsByOwnerID": {},
	"pet/pet_owner_repository.go|petOwnerRepository.ReplaceForPet":           {},
	"pet/pet_owner_repository.go|petOwnerRepository.CountByOwnerID":          {},
	// BE9 pet create write owner: direct create, owner lock, number allocation, and reload
	// remain in the caller's ambient transaction. Runtime:
	// TestPetRepository_Create_AmbientTransactionNeverEscapesBaseDB; rollback-on-reload:
	// TestPetRepository_Create_ReloadFailureRollsBackPet.
	"pet/owner_registration.go|writer.Create":                     {},
	"pet/owner_registration.go|writer.CreateForOwnerRegistration": {},
	// BE9 owner reads stay inside the caller's ambient transaction. Public wrapper proof:
	// TestOwnerRepository_FindByID_UsesAmbientTransaction; rollback-on-reload:
	// TestOwnerRepository_CreateWithPets_ReloadFailureRollsBackGraph.
	// LINE webhook CAS writes preserve clinic/linked-ID/event-order predicates and
	// roll back with the caller's transaction. Runtime:
	// TestOwnerRepository_LineWebhookCASUpdates_RollBackWithAmbientTransaction.
	"owner/repository.go|ownerRepository.findOwnerByID":        {},
	"owner/repository.go|ownerRepository.CountPetsByOwnerID":   {},
	"owner/repository.go|ownerRepository.Delete":               {},
	"owner/repository.go|ownerRepository.LockForDelete":        {},
	"owner/repository.go|ownerRepository.LockLineLinkOwner":    {},
	"owner/repository.go|ownerRepository.UpdateLineBlockedAt":  {},
	"owner/repository.go|ownerRepository.UpdateLineFollowedAt": {},
	"owner/repository.go|ownerRepository.UpdateLineUserID":     {},
	// prescription (X-11 Appendix-A finalize-child-write-race fix — same FK-deadlock rationale as examination)
	// procedure (MRC-07 ambient tx for delete usage check)
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.FindAll":                 {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.FindByID":                {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.Create":                  {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.Update":                  {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.Delete":                  {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.CountUsageByProcedureID": {},
	"medicalrecord/procedure_repository.go|procedureRepositoryImpl.CountChildrenByParentID": {},
	// SEC-CS-F13: cage soft-delete + hospitalization usage re-check join ambient tx.
	// FindByID takes FOR SHARE under ambient tx (hospitalization cage FK validation).
	// Runtime: cage_delete_concurrency_test.go ConcurrentAssignFirst / DeleteFirst /
	// CountUsage_AmbientTxSeesUncommittedHospitalization /
	// LockByIDForUpdate_RequiresAmbientTransaction.
	"medicalrecord/cage_repository.go|cageRepositoryImpl.FindAll":              {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.FindByID":             {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.LockByIDForUpdate":    {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.Create":               {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.Update":               {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.Delete":               {},
	"medicalrecord/cage_repository.go|cageRepositoryImpl.CountUsageByCageID":   {},
	"medicalrecord/prescription_repository.go|prescriptionRepository.Create":   {}, // BE8-4 batch7: moved from prescription_repository.go
	"medicalrecord/prescription_repository.go|prescriptionRepository.FindByID": {}, // MRC-01: response re-fetch must observe and govern the same tx mutation
	"medicalrecord/prescription_repository.go|prescriptionRepository.Update":   {}, // BE8-4 batch7: moved from prescription_repository.go
	// prescription Delete (BE-refactor.md H-8e: prescriptionService.Delete が finalize ロック確認・
	// Delete を s.transactor.WithTx で束ねるようになったための追加。examination Delete=H-8d と同型)
	"medicalrecord/prescription_repository.go|prescriptionRepository.Delete": {}, // BE8-4 batch7: moved from prescription_repository.go
	// refund (R1-1 TOCTOU)
	"billing/refund_repository.go|refundRepository.Create":                         {}, // BE8-4 batch8: moved from refund_repository.go
	"billing/refund_repository.go|refundRepository.SumByBillingID":                 {}, // BE8-4 batch8: moved from refund_repository.go
	"billing/refund_repository.go|refundRepository.SumByBillingIDAndPaymentMethod": {}, // BE8-4 batch8: moved from refund_repository.go
	// reservation (uniform dbOrTx)
	"reservation/reservation_repository.go|reservationRepository.AssertLineCustomerInClinic":                       {}, // AUD-001
	"reservation/reservation_repository.go|reservationRepository.AssertOwnerInClinic":                              {}, // AUD-001
	"reservation/reservation_repository.go|reservationRepository.AcquireBookingLock":                               {}, // X-9 (Appendix-A phantom-booking fix)
	"reservation/reservation_repository.go|reservationRepository.FindPetOwnerInClinic":                             {}, // AUD-001
	"reservation/reservation_repository.go|reservationRepository.CountByCustomerAndDateRange":                      {},
	"reservation/reservation_repository.go|reservationRepository.CountByDateAndSource":                             {},
	"reservation/reservation_repository.go|reservationRepository.CountByTypeAndStartTime":                          {},
	"reservation/reservation_repository.go|reservationRepository.CountByTypeAndStartTimes":                         {},
	"reservation/reservation_repository.go|reservationRepository.CountConflicts":                                   {},
	"reservation/reservation_repository.go|reservationRepository.CountMedicalRecordsByReservationID":               {},
	"reservation/reservation_repository.go|reservationRepository.CountOnDutyDoctors":                               {},
	"reservation/reservation_repository.go|reservationRepository.Create":                                           {},
	"reservation/reservation_repository.go|reservationRepository.Delete":                                           {},
	"reservation/reservation_repository.go|reservationRepository.ExistsByReservationTypeID":                        {},
	"reservation/reservation_repository.go|reservationRepository.ExistsByStaffID":                                  {},
	"reservation/reservation_repository.go|reservationRepository.FindAll":                                          {},
	"reservation/reservation_repository.go|reservationRepository.FindAllByCategory":                                {},
	"reservation/reservation_repository.go|reservationRepository.FindClinicIDsByStaffID":                           {},
	"reservation/reservation_repository.go|reservationRepository.findReservationByID":                              {},
	"reservation/reservation_repository.go|reservationRepository.HasDoctorConflict":                                {},
	"reservation/reservation_repository.go|reservationRepository.LockAndFindByID":                                  {},
	"reservation/reservation_repository.go|reservationRepository.update":                                           {},
	"reservation/reservation_intent_repository.go|reservationRepository.CompleteForAccounting":                     {}, // BE9-2E-0 write owner
	"reservation/reservation_intent_repository.go|reservationRepository.DeleteForTrimming":                         {}, // BE9-2E-0 typed delete + ambient-tx rollback test
	"reservation/reservation_intent_repository.go|reservationRepository.AssertMedicalRecordDoctorInClinic":         {}, // BE9-2E-0 doctor guard; TestVaccinationService_DoctorAssignmentDeletionWaitsForValidationTransaction
	"reservation/reservation_intent_repository.go|reservationRepository.MarkNoShowAt":                              {}, // durable scheduler no-show transition at an explicit slot
	"reservation/reservation_intent_repository.go|reservationRepository.UpdateForTrimming":                         {}, // BE9-2E-0 typed update + ambient-tx rollback test
	"reservation/reservation_intent_repository.go|reservationRepository.acquireAppointmentLifecycleLock":           {}, // no-show/finalization serialization
	"reservation/reservation_intent_repository.go|reservationRepository.assertActiveTrimmingReservationType":       {}, // new trimming writes require active clinic-scoped master
	"reservation/reservation_intent_repository.go|reservationRepository.assertGeneralMedicalRecordReservationType": {}, // BackfillForMedicalRecord category guard under appointment tx
	"reservation/reservation_intent_repository.go|reservationRepository.assertStaffAssignedToClinic":               {}, // BE9-2E-0 doctor tenant guard
	"reservation/reservation_intent_repository.go|reservationRepository.assertTrimmingReservationType":             {}, // BE9-2E-0 master-FK tenant guard
	"reservation/appointment_admin_repository.go|reservationAdminRepository.Create":                                {}, // AUD-001
	// BUG-424: unavailable-time reads participate in the booking transaction.
	"reservation/reservation_type_unavailable_time_repository.go|reservationTypeUnavailableTimeRepository.FindAll": {},
	// trimming master reads join the service-owned transaction and hold SHARE locks through
	// appointment/detail/junction writes. Runtime proof: TestTrimmingMasterFindByID_HoldsShareLockForAmbientTransaction.
	"trimming/trimming_course_repository.go|trimmingCourseRepository.CountUsageByTrimmingCourseID":      {},
	"trimming/trimming_course_repository.go|trimmingCourseRepository.Create":                            {},
	"trimming/trimming_course_repository.go|trimmingCourseRepository.Delete":                            {},
	"trimming/trimming_course_repository.go|trimmingCourseRepository.FindAll":                           {},
	"trimming/trimming_course_repository.go|trimmingCourseRepository.FindByID":                          {},
	"trimming/trimming_course_repository.go|trimmingCourseRepository.Update":                            {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.CountUsageByCourseTypeID": {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.Create":                   {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.Delete":                   {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.FindAll":                  {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.FindByID":                 {},
	"trimming/trimming_course_type_repository.go|trimmingCourseTypeRepository.Update":                   {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.CountUsageByTrimmingOptionID":      {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.Create":                            {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.Delete":                            {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.FindAll":                           {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.FindByID":                          {},
	"trimming/trimming_option_repository.go|trimmingOptionRepository.Update":                            {},
	// reservationtype domain package (methods that previously used dbOrTx; Update/Delete remain
	// r.db.WithContext by design — behavior preserved from flat file; facade keeps service imports)
	"reservation/reservation_type_repository.go|reservationTypeRepository.CountChildrenByParentID":       {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.CountUsageByReservationTypeID": {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.Create":                        {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.FindAll":                       {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.FindAllWithChildren":           {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.FindByID":                      {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.FindByIDWithChildren":          {},
	"reservation/reservation_type_repository.go|reservationTypeRepository.Update":                        {},
	// staff_clinic_assignment (moved into internal/staff). Runtime lock/replace proofs live in
	// staff_clinic_assignment_*_test.go and staff_assignment_concurrency_test.go.
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.CountByStaffAndClinic":      {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.Create":                     {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.Delete":                     {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.DeleteByStaffAndClinicIDs":  {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.FindByStaffAndClinic":       {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.FindByStaffID":              {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.LockActiveByStaff":          {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.LockActiveByStaffAndClinic": {},
	"staff/staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.RestoreOrCreate":            {},
	// staff (uniform DBOrTx; runtime atomicity/isolation proofs live in internal/staff)
	"staff/staff_repository.go|staffRepository.CountBlockingReferencesByStaffID": {},
	"staff/staff_repository.go|staffRepository.Create":                           {},
	"staff/staff_repository.go|staffRepository.Delete":                           {},
	"staff/staff_repository.go|staffRepository.FindAll":                          {},
	"staff/staff_repository.go|staffRepository.FindByAccountID":                  {},
	"staff/staff_repository.go|staffRepository.FindByID":                         {},
	"staff/staff_repository.go|staffRepository.FindByIDInClinic":                 {},
	"staff/staff_repository.go|staffRepository.LockActiveByIDForShare":           {},
	"staff/staff_repository.go|staffRepository.LockActiveByIDForUpdate":          {},
	"staff/staff_repository.go|staffRepository.LockActiveByIDForUpdateInClinic":  {},
	"staff/staff_repository.go|staffRepository.Reorder":                          {},
	"staff/staff_repository.go|staffRepository.Update":                           {},
	"staff/staff_repository.go|staffRepository.UpdatePrimaryClinicID":            {},
	"staff/staff_repository.go|staffRepository.activeSystemAdminStaffQuery":      {},
	// ADR-006 論点#1 案A: reservation_staff_repository.go から移動した予約用途 write
	"staff/staff_repository.go|staffRepository.CreateForReservation":        {},
	"staff/staff_repository.go|staffRepository.UpdateForReservation":        {},
	"staff/staff_repository.go|staffRepository.SwapSortOrderForReservation": {},
	// occupation (moved into internal/staff; service-owned transaction and master-FK tests)
	"staff/occupation_repository.go|occupationRepository.CountUsageByOccupationID": {},
	"staff/occupation_repository.go|occupationRepository.Create":                   {},
	"staff/occupation_repository.go|occupationRepository.Delete":                   {},
	"staff/occupation_repository.go|occupationRepository.FindAll":                  {},
	"staff/occupation_repository.go|occupationRepository.FindByID":                 {},
	"staff/occupation_repository.go|occupationRepository.Update":                   {},
	"staff/occupation_repository.go|occupationRepository.WithTx":                   {},
	"staff/occupation_repository.go|occupationRepository.lockActiveByID":           {},
	// shift_entry (uniform DBOrTx; Save/Delete concurrency proofs live in internal/staff)
	"staff/shift_entry_repository.go|shiftEntryRepository.Create": {},
	// ADR-006 論点#1 案A: reservation_schedule_repository.go から移動した予約用途 write
	"staff/shift_entry_repository.go|shiftEntryRepository.SaveByStaffDate":         {},
	"staff/shift_entry_repository.go|shiftEntryRepository.Delete":                  {},
	"staff/shift_entry_repository.go|shiftEntryRepository.DeleteByStaffDate":       {},
	"staff/shift_entry_repository.go|shiftEntryRepository.ExistsByStaffID":         {},
	"staff/shift_entry_repository.go|shiftEntryRepository.FindAll":                 {},
	"staff/shift_entry_repository.go|shiftEntryRepository.FindByID":                {},
	"staff/shift_entry_repository.go|shiftEntryRepository.FindClinicIDsByStaffID":  {},
	"staff/shift_entry_repository.go|shiftEntryRepository.FindOnDutyStaffs":        {},
	"staff/shift_entry_repository.go|shiftEntryRepository.LockActiveByIDForUpdate": {},
	"staff/shift_entry_repository.go|shiftEntryRepository.Update":                  {},
	"staff/shift_entry_repository.go|shiftEntryRepository.ReplaceBreaks":           {},
	// shift_template moved into internal/staff; the repository-owned break replacement and
	// service transaction tests pin atomicity.
	"staff/shift_template_repository.go|shiftTemplateRepository.Create":                  {},
	"staff/shift_template_repository.go|shiftTemplateRepository.Delete":                  {},
	"staff/shift_template_repository.go|shiftTemplateRepository.FindAll":                 {},
	"staff/shift_template_repository.go|shiftTemplateRepository.FindByID":                {},
	"staff/shift_template_repository.go|shiftTemplateRepository.LockActiveByIDForUpdate": {},
	"staff/shift_template_repository.go|shiftTemplateRepository.Update":                  {},
	"staff/shift_template_repository.go|shiftTemplateRepository.UpdateBreaks":            {},
	"staff/shift_template_repository.go|shiftTemplateRepository.WithTx":                  {},
	// staff_provisioning (TASK-609). Runtime ambient-tx proofs:
	// staff_provisioning_repository_tx_atomicity_test.go +
	// staff_provisioning_repository_integration_test.go (Apply path).
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.AcquireBatchLock":               {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.AssignPermissionGroups":         {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.ClinicExists":                   {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.CreateAccount":                  {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.CreateAssignment":               {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.CreateStaff":                    {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.EmailExists":                    {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.FindAccountByID":                {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.FindReceiptsInScope":            {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.FindStaffByAccountID":           {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.HasMasterStaffCreate":           {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.LockOccupationForShare":         {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.OccupationBelongsToClinic":      {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.PermissionGroupsBelongToClinic": {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.StaffAssignedToClinic":          {},
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.WriteAudit":                     {},
	// identitylink (#239 Phase 1). Runtime ambient-tx proofs:
	// identitylink/repository_tx_atomicity_test.go
	"identitylink/repository.go|repository.conn":             {},
	"identitylink/repository.go|repository.requireAmbientTx": {},
	// trimming detail (uniform dbOrTx)
	"trimming/trimming_repository.go|appointmentTrimmingDetailRepository.Create":              {},
	"trimming/trimming_repository.go|appointmentTrimmingDetailRepository.FindByAppointmentID": {},
	"trimming/trimming_repository.go|appointmentTrimmingDetailRepository.SetOptions":          {},
	"trimming/trimming_repository.go|appointmentTrimmingDetailRepository.Update":              {},
	// vital (X-11 Appendix-A finalize-child-write-race fix — same FK-deadlock rationale as examination)
	// BE9-2D sub-batch④a: moved to internal/medicalrecord (facade kept in internal/repository).
	// FindByMedicalRecordID: BE9-2D ④b — treatmentService の dose 体重解決が保存 tx 内から読む
	// read の tx 参加維持（旧 repos.Transaction の tx-bound clone と等価にする）
	"medicalrecord/vital_repository.go|vitalRepository.FindByMedicalRecordID": {},
	"medicalrecord/vital_repository.go|vitalRepository.Create":                {},
	"medicalrecord/vital_repository.go|vitalRepository.Update":                {},
	"medicalrecord/vital_repository.go|vitalRepository.Delete":                {},
	// G6-2 (BE-refactor.md tx-mechanism-consolidation): repo-internal r.db.WithContext(ctx).Transaction
	// → dbOrTx(ctx, r.db).Transaction conversion, no ambient-tx caller into any of these (verified per-file).
	"manualarticle/repository.go|repository.Upsert":      {}, // BE8-4 batch3: moved from manual_article_repository.go
	"owner/repository.go|ownerRepository.CreateWithPets": {},
	// SEC-CS-F15: UpdateAndFindApplying owns DBOrTx + LockByIDForUpdate; UpdateAndFind is a thin
	// delegate without its own DBOrTx shape (removed from allowlist after F15).
	// Runtime: owner_discount_toctou_test.go.
	"owner/repository.go|ownerRepository.UpdateAndFindApplying":                                                {},
	"owner/repository.go|ownerRepository.LockByIDForUpdate":                                                    {},
	"owner/repository.go|ownerRepository.RecordLstepOptOut":                                                    {},
	"owner/repository.go|ownerRepository.ClearLstepOptOut":                                                     {},
	"reservation/reservation_type_liff_repository.go|reservationTypeLiffRepository.UpdateSortOrder":            {},
	"reservation/reservation_type_liff_repository.go|reservationTypeLiffRepository.Update":                     {},
	"reservation/reservation_type_liff_repository.go|reservationTypeLiffRepository.Delete":                     {},
	"reservation/reservation_type_liff_repository.go|reservationTypeLiffRepository.DeleteWithDependencyChecks": {},
	"reservation/reservation_type_liff_repository.go|reservationTypeLiffRepository.FindByID":                   {},
	// treatment (BE9-2D ④b: WithTx 化に伴う ambient tx 参加。④b Batch A で medicalrecord へ移動済み、
	// lockDraftMedicalRecord 行ロック・在庫減算・逸脱監査と同一 ambient tx へ参加させる)
	"medicalrecord/treatment_repository.go|treatmentRepository.Create":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.Delete":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.Update":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.BulkUpdateSortOrder": {},
	// SEC-CS-F09/F10: treatment / treatment-plan discount recheck under FOR UPDATE in write TX.
	// Runtime: treatment_discount_toctou_test.go, treatment_plan_discount_toctou_test.go.
	"medicalrecord/treatment_repository.go|treatmentRepository.LockByIDForUpdate":          {},
	"medicalrecord/treatment_plan_repository.go|treatmentPlanRepository.LockByIDForUpdate": {},
	// X-6 (Appendix-A tx-atomicity fix, commit d7eff8c8): medicine/inventory repo-internal
	// r.db.WithContext(ctx).Transaction → dbOrTx(ctx, r.db).Transaction. Allowlist backfill
	// discovered during G6-2 (X-6 landed without registering these).
	"medicalrecord/medicine_repository.go|medicineRepository.Create":                  {},
	"medicalrecord/medicine_repository.go|medicineRepository.Update":                  {},
	"medicalrecord/medicine_repository.go|medicineRepository.Delete":                  {},
	"medicalrecord/medicine_repository.go|medicineRepository.FindAll":                 {},
	"medicalrecord/medicine_repository.go|medicineRepository.FindByID":                {},
	"medicalrecord/medicine_repository.go|medicineRepository.CountChildrenByParentID": {},
	"medicalrecord/medicine_repository.go|medicineRepository.CountUsageByMedicineID":  {},
	"inventory/repository.go|repository.Create":                                       {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.Update":                                       {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.Delete":                                       {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.FindAll":                                      {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.FindByID":                                     {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.DecreaseStock":                                {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.CountUsageByInventoryID":                      {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.UpdateNameByMedicineCategory":                 {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.DeleteByNameAndMedicineCategory":              {}, // BE8-4 batch18: moved from inventory_repository.go
	// BUG-465 landed DBOrTx on merchandise Create/Update without allowlist (main red).
	// Runtime: merchandise_item_repository_test.go AmbientTxRollback cases.
	"inventory/merchandise_item_repository.go|merchandiseItemRepository.Create": {},
	"inventory/merchandise_item_repository.go|merchandiseItemRepository.Update": {},
	// BE-ACT-CAMPAIGN-TARGET-SERIALIZATION: FindByID joins ambient tx and takes FOR SHARE so
	// campaign target attachment serializes with concurrent merchandise soft-delete.
	// Runtime: TestMerchandiseItemRepository_FindByID_HoldsShareLockForAmbientTransaction.
	"inventory/merchandise_item_repository.go|merchandiseItemRepository.FindByID": {},
	// BE-ACT-MERCHANDISE-ATOMIC-DELETE: soft-delete + billing/estimate/campaign-target usage
	// re-check join the service-owned ambient transaction. Runtime:
	// TestMerchandiseItemService_Delete_ConcurrentAttachFirstYieldsConflict /
	// TestMerchandiseItemRepository_CountUsageByMerchandiseItemID_IncludesCampaignTargets /
	// TestMerchandiseItemRepository_CountUsage_AmbientTxSeesUncommittedCampaignTarget.
	"inventory/merchandise_item_repository.go|merchandiseItemRepository.CountUsageByMerchandiseItemID": {},
	"inventory/merchandise_item_repository.go|merchandiseItemRepository.Delete":                        {},
	// X-7 (Appendix-A tx-atomicity fix, commit 2a7a4dfc): clinic repository tx conversion.
	// Permission-group ownership moved to internal/auth in BE9 auth Phase 1.
	"clinic/clinic_repository.go|clinicRepository.Create":                            {},
	"clinic/clinic_repository.go|clinicRepository.Update":                            {},
	"clinic/clinic_repository.go|clinicRepository.Delete":                            {},
	"clinic/clinic_repository.go|clinicRepository.FindAll":                           {},
	"clinic/clinic_repository.go|clinicRepository.FindByID":                          {},
	"clinic/clinic_repository.go|clinicRepository.FindCompany":                       {},
	"clinic/clinic_repository.go|clinicRepository.LockActiveByID":                    {},
	"clinic/clinic_repository.go|clinicRepository.LockByIDForUpdate":                 {},
	"clinic/clinic_repository.go|clinicRepository.CountOwnersByClinicID":             {},
	"clinic/clinic_repository.go|clinicRepository.CountStaffByClinicID":              {},
	"clinic/clinic_repository.go|clinicRepository.CountBlockingReferencesByClinicID": {},
	// X-8 (Appendix-A tx-atomicity fix, commit 1e2d483c): reservation_staff repo-internal tx
	// conversion. Allowlist backfill discovered during G6-2 (X-8 landed without registering these).
	"reservation/reservation_staff_repository.go|reservationStaffRepository.UpdateExcludedReservationTypes": {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.UpdateReservationCapabilities":  {},
	// BE-refactor.md H-7: FindByID を dbOrTx 化し、reservationStaffService.Update の
	// tx 内所有権確認（GetByID）を ambient tx に参加させ TOCTOU 窓を閉じる。
	"reservation/reservation_staff_repository.go|reservationStaffRepository.FindByID": {},
	// Stage B: FindAllExcluded* are pure facades (universe \ capable) with no body-level
	// DBOrTx — inventory the leaf methods instead. Runtime leaf coverage:
	// TestReservationStaffRepository_LeafReads_SeeUncommittedAmbientWrites
	// (optional composition probe still calls FindAllExcluded* facades).
	"reservation/reservation_staff_repository.go|reservationStaffRepository.hasActiveClinicAssignment":                {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.filterStaffIDsWithActiveAssignment":       {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.listActiveReservationTypeUniverse":        {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.FindAllReservationCapabilities":           {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.FindAllReservationCapabilitiesByStaffIDs": {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.LockForMutation":                          {},
	"reservation/reservation_staff_repository.go|reservationStaffRepository.SupportsReservationType":                  {}, // assignment/capability SHARE-lock concurrency proof
	// Durable scheduler reads use an explicit slot timestamp and participate in the caller's tx.
	"reservation/reservation_repository.go|reservationRepository.FindNoShowCandidatesAt": {},
	// SD-10 deceased write guard: ambient SHARE-lock read. Runtime:
	// TestReservationRepository_FindPetByIDInClinic_SeesUncommittedDeceasedUpdate
	// + TestReservationRepository_FindPetByIDInClinic_ShareLockBlocksConcurrentWriter
	"reservation/reservation_repository.go|reservationRepository.FindPetByIDInClinic": {},

	// Runtime: TestExamReferenceRangeRepository_FindAnimalSpeciesID_HoldsExamShareLockUntilAmbientTransactionCommits.
	"medicalrecord/exam_reference_range_repository.go|examinationRepository.FindAnimalSpeciesID": {},

	// Runtime: TestExamReferenceRangeRepository_ResolveByFieldIDs_HoldsReferenceRangeShareLockUntilAmbientTransactionCommits.
	"medicalrecord/exam_reference_range_repository.go|examinationRepository.ResolveByFieldIDs": {},

	// AE-LAB lab_device master/device catalog. Writers roll back with ambient tx; readers
	// see uncommitted writes. Runtime: lab_device_item_master_repository_tx_atomicity_test.go
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindByClinicSourceCodes": {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindExamTypeField":       {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.EnsureCatalog":           {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.Update":                  {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindDeviceByID":          {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindByID":                {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.ListDevices":             {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.UpdateDevice":            {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.EnsureDevices":           {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.CreateDevice":            {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.List":                    {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindExamType":            {},
	"medicalrecord/lab_device_item_master_repository.go|labDeviceItemMasterRepository.FindExamTypeFields":      {},
	// Private DBOrTx helper for receive jobs/waits/station. Runtime:
	// TestLabDeviceReceiveRepository_CreateJobWithItems_RollsBackWhenAmbientTxFails
	"medicalrecord/lab_device_receive_repository.go|labDeviceReceiveRepository.q": {},
}

type ambientTxParticipationShape uint8

const (
	ambientTxViaLocalDBOrTxHelper ambientTxParticipationShape = iota + 1
	ambientTxRequired
	ambientTxRequiredViaLocalHelper
)

type ambientTxParticipationExpectation struct {
	shape      ambientTxParticipationShape
	helperName string
}

// ambientTxParticipationExpectations pins methods that intentionally use a stronger or
// wrapped ambient-transaction shape instead of a literal DBOrTx call in the method body.
//
// The helper shape is accepted only while both edges remain visible in the same source file:
// the method must call the named helper, and that helper must call DBOrTx. The required shape
// must keep a literal persistence.TxFromContext call, so it cannot silently weaken back to
// fallback-on-base-DB behavior.
var ambientTxParticipationExpectations = map[string]ambientTxParticipationExpectation{
	// TASK-027 Slice A append is fail-closed: it must receive an ambient transaction and
	// never fall back to the repository base DB. Runtime proof is named in the allowlist above.
	"medicalrecord/examination_revision_repository.go|examinationRepository.AppendOfficialRevision": {
		shape: ambientTxRequired,
	},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendWorkingRevisionFromOfficial": {
		shape:      ambientTxRequiredViaLocalHelper,
		helperName: "lockRevisionWorkflowParent",
	},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendWorkingRevisionFromCurrent": {
		shape: ambientTxRequired,
	},
	"medicalrecord/examination_revision_workflow_repository.go|examinationRepository.AppendOfficialRevisionFromWorking": {
		shape:      ambientTxRequiredViaLocalHelper,
		helperName: "lockRevisionWorkflowParent",
	},
	"auth/account_repository.go|accountRepository.CompareAndSwapPasswordHash": {
		shape: ambientTxRequired,
	},
	"auth/account_repository.go|accountRepository.FindByIDForUpdate": {
		shape: ambientTxRequired,
	},
	"auth/account_repository.go|accountRepository.UpdatePasswordHash": {
		shape: ambientTxRequired,
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.Create": {
		shape:      ambientTxViaLocalDBOrTxHelper,
		helperName: "silentPasswordResetTokenDB",
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.ConsumeByID": {
		shape: ambientTxRequired,
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.DeleteByAccountID": {
		shape: ambientTxRequired,
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.DeleteIssued": {
		shape:      ambientTxViaLocalDBOrTxHelper,
		helperName: "silentPasswordResetTokenDB",
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindLatestByAccountIDForUpdate": {
		shape: ambientTxRequired,
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindByTokenHash": {
		shape:      ambientTxViaLocalDBOrTxHelper,
		helperName: "silentPasswordResetTokenDB",
	},
	"auth/password_reset_token_repository.go|passwordResetTokenRepository.FindByTokenHashForUpdate": {
		shape: ambientTxRequired,
	},
	"audit/repository.go|repository.CreateTx": {
		shape: ambientTxRequired,
	},
	"billing/billing_confirmation_repository.go|billingConfirmationRepository.LockActiveStaffAssignment": {
		shape: ambientTxRequired,
	},
	"billing/billing_item_repository.go|billingItemRepository.ValidateCreateReferences": {
		shape: ambientTxRequired,
	},
	"billing/billing_item_service.go|billingItemService.createItemInAmbientTx": {
		shape: ambientTxRequired,
	},
	"billing/billing_item_repository.go|billingItemRepository.ValidateVaccinationCreateReference": {
		shape: ambientTxRequired,
	},
	"billing/billing_item_exam.go|billingItemRepository.ValidateExamCreateReference": {
		shape: ambientTxRequired,
	},
	"pet/owner_registration.go|writer.CreateForOwnerRegistration": {
		shape:      ambientTxRequiredViaLocalHelper,
		helperName: "createPetsInTransaction",
	},
	"auth/token_blacklist_repository.go|tokenBlacklistRepository.Create": {
		shape:      ambientTxViaLocalDBOrTxHelper,
		helperName: "silentTokenBlacklistDB",
	},
	"auth/token_blacklist_repository.go|tokenBlacklistRepository.ExistsByJTI": {
		shape:      ambientTxViaLocalDBOrTxHelper,
		helperName: "silentTokenBlacklistDB",
	},
	// staff provision audit: TxFromContext-only fail-closed (never weakens to DBOrTx).
	"staff/staff_provisioning_repository.go|staffProvisioningRepository.WriteAudit": {
		shape: ambientTxRequired,
	},
	// identitylink writes: requireAmbientTx returns TxFromContext handle only.
	"identitylink/repository.go|repository.requireAmbientTx": {
		shape: ambientTxRequired,
	},
}

// funcUsesDBOrTx reports whether a function body contains a call to dbOrTx / DBOrTx /
// persistence.DBOrTx(...). Does not chase helpers (see Reorder
// ambient policy in file header).
func funcUsesDBOrTx(fd *ast.FuncDecl) bool {
	if fd.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := ce.Fun.(type) {
		case *ast.Ident:
			// Legacy wrapper `dbOrTx` or same-package canonical `DBOrTx`.
			if fun.Name == "dbOrTx" || fun.Name == "DBOrTx" {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			if fun.Sel != nil && fun.Sel.Name == "DBOrTx" {
				if id, ok := fun.X.(*ast.Ident); ok {
					if id.Name == "persistence" {
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return found
}

type ambientTxProducerMatcher func(*ast.CallExpr) bool

func isPersistenceTxFromContext(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel == nil || selector.Sel.Name != "TxFromContext" {
		return false
	}
	packageIdent, ok := selector.X.(*ast.Ident)
	return ok && packageIdent.Name == "persistence"
}

func isLocalHelperCall(helperName string) ambientTxProducerMatcher {
	return func(call *ast.CallExpr) bool {
		ident, ok := call.Fun.(*ast.Ident)
		return ok && ident.Name == helperName
	}
}

func expressionContainsProducer(expr ast.Expr, matches ambientTxProducerMatcher) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if ok && matches(call) {
			found = true
			return false
		}
		return true
	})
	return found
}

func assignedProducerHandles(fd *ast.FuncDecl, matches ambientTxProducerMatcher) map[string]struct{} {
	handles := make(map[string]struct{})
	if fd.Body == nil {
		return handles
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		assignment, ok := n.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != len(assignment.Rhs) {
			return true
		}
		for i := range assignment.Rhs {
			if !expressionContainsProducer(assignment.Rhs[i], matches) {
				continue
			}
			ident, ok := assignment.Lhs[i].(*ast.Ident)
			if ok && ident.Name != "_" {
				handles[ident.Name] = struct{}{}
			}
		}
		return true
	})
	return handles
}

func producerHandlesRemainDerived(
	fd *ast.FuncDecl,
	handles map[string]struct{},
	matches ambientTxProducerMatcher,
) bool {
	if fd.Body == nil {
		return false
	}
	valid := true
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if !valid {
			return false
		}
		assignment, ok := n.(*ast.AssignStmt)
		if !ok || len(assignment.Lhs) != len(assignment.Rhs) {
			return true
		}
		for i := range assignment.Lhs {
			ident, ok := assignment.Lhs[i].(*ast.Ident)
			if !ok {
				continue
			}
			if _, tracked := handles[ident.Name]; !tracked {
				continue
			}
			if (matches != nil && expressionContainsProducer(assignment.Rhs[i], matches)) ||
				expressionDerivedFromHandle(assignment.Rhs[i], handles) {
				continue
			}
			valid = false
			return false
		}
		return true
	})
	return valid
}

func expressionDerivedFromHandle(expr ast.Expr, handles map[string]struct{}) bool {
	switch typed := expr.(type) {
	case *ast.Ident:
		_, ok := handles[typed.Name]
		return ok
	case *ast.ParenExpr:
		return expressionDerivedFromHandle(typed.X, handles)
	case *ast.SelectorExpr:
		return expressionDerivedFromHandle(typed.X, handles)
	case *ast.CallExpr:
		selector, ok := typed.Fun.(*ast.SelectorExpr)
		return ok && expressionDerivedFromHandle(selector.X, handles)
	case *ast.IndexExpr:
		return expressionDerivedFromHandle(typed.X, handles)
	case *ast.IndexListExpr:
		return expressionDerivedFromHandle(typed.X, handles)
	case *ast.TypeAssertExpr:
		return expressionDerivedFromHandle(typed.X, handles)
	default:
		return false
	}
}

// funcUsesProducedDBHandle requires the producer result to flow into the receiver side of a
// selector call. Merely invoking TxFromContext/helper and then issuing the query through r.db
// does not satisfy this shape.
func funcUsesProducedDBHandle(fd *ast.FuncDecl, matches ambientTxProducerMatcher) bool {
	if fd.Body == nil {
		return false
	}
	handles := assignedProducerHandles(fd, matches)
	if len(handles) > 0 && !producerHandlesRemainDerived(fd, handles, matches) {
		return false
	}
	found := false
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if expressionContainsProducer(selector.X, matches) ||
			expressionDerivedFromHandle(selector.X, handles) {
			found = true
			return false
		}
		return true
	})
	return found
}

func funcUsesNamedDBHandle(fd *ast.FuncDecl, handleName string) bool {
	if fd.Body == nil || handleName == "" {
		return false
	}
	handles := map[string]struct{}{handleName: {}}
	if !producerHandlesRemainDerived(fd, handles, nil) {
		return false
	}
	found := false
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && expressionDerivedFromHandle(selector.X, handles) {
			found = true
			return false
		}
		return true
	})
	return found
}

func funcForwardsProducedHandleToLocalHelper(
	fd *ast.FuncDecl,
	matches ambientTxProducerMatcher,
	helperName string,
) (int, bool) {
	if fd.Body == nil || helperName == "" {
		return 0, false
	}
	handles := assignedProducerHandles(fd, matches)
	if len(handles) == 0 || !producerHandlesRemainDerived(fd, handles, matches) {
		return 0, false
	}
	foundIndex := 0
	found := false
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != helperName {
			return true
		}
		for i, argument := range call.Args {
			if expressionContainsProducer(argument, matches) ||
				expressionDerivedFromHandle(argument, handles) {
				foundIndex = i
				found = true
				return false
			}
		}
		return true
	})
	return foundIndex, found
}

func parameterNameAt(fd *ast.FuncDecl, index int) (string, bool) {
	if fd.Type == nil || fd.Type.Params == nil || index < 0 {
		return "", false
	}
	current := 0
	for _, field := range fd.Type.Params.List {
		if len(field.Names) == 0 {
			if current == index {
				return "", false
			}
			current++
			continue
		}
		for _, name := range field.Names {
			if current == index {
				return name.Name, name.Name != ""
			}
			current++
		}
	}
	return "", false
}

func funcReturnsDBOrTxHandle(fd *ast.FuncDecl) bool {
	if fd.Body == nil {
		return false
	}
	handles := assignedProducerHandles(fd, funcUsesDBOrTxCall)
	if len(handles) > 0 &&
		!producerHandlesRemainDerived(fd, handles, funcUsesDBOrTxCall) {
		return false
	}
	foundReturn := false
	valid := true
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if !valid {
			return false
		}
		returnStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, result := range returnStmt.Results {
			if expressionContainsProducer(result, funcUsesDBOrTxCall) ||
				expressionDerivedFromHandle(result, handles) {
				foundReturn = true
				continue
			}
			valid = false
			return false
		}
		return true
	})
	return valid && foundReturn
}

func funcUsesDBOrTxCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name == "dbOrTx" || fun.Name == "DBOrTx"
	case *ast.SelectorExpr:
		if fun.Sel == nil || fun.Sel.Name != "DBOrTx" {
			return false
		}
		packageIdent, ok := fun.X.(*ast.Ident)
		return ok && packageIdent.Name == "persistence"
	default:
		return false
	}
}

func funcUsesRequiredAmbientTx(fd *ast.FuncDecl) bool {
	return funcUsesProducedDBHandle(fd, isPersistenceTxFromContext)
}

func funcMatchesAmbientTxExpectation(
	file *ast.File,
	method *ast.FuncDecl,
	expectation ambientTxParticipationExpectation,
) bool {
	switch expectation.shape {
	case ambientTxViaLocalDBOrTxHelper:
		for _, decl := range file.Decls {
			helper, ok := decl.(*ast.FuncDecl)
			if !ok ||
				helper.Recv != nil ||
				helper.Name == nil ||
				helper.Name.Name != expectation.helperName {
				continue
			}
			return funcReturnsDBOrTxHandle(helper) &&
				funcUsesProducedDBHandle(method, isLocalHelperCall(expectation.helperName))
		}
		return false
	case ambientTxRequired:
		return funcUsesRequiredAmbientTx(method)
	case ambientTxRequiredViaLocalHelper:
		argumentIndex, ok := funcForwardsProducedHandleToLocalHelper(
			method,
			isPersistenceTxFromContext,
			expectation.helperName,
		)
		if !ok {
			return false
		}
		for _, decl := range file.Decls {
			helper, ok := decl.(*ast.FuncDecl)
			if !ok ||
				helper.Recv != nil ||
				helper.Name == nil ||
				helper.Name.Name != expectation.helperName {
				continue
			}
			parameterName, ok := parameterNameAt(helper, argumentIndex)
			return ok && funcUsesNamedDBHandle(helper, parameterName)
		}
		return false
	default:
		return false
	}
}

func detectAmbientTxParticipationExpectation(
	file *ast.File,
	method *ast.FuncDecl,
) (ambientTxParticipationExpectation, bool) {
	if funcUsesRequiredAmbientTx(method) {
		return ambientTxParticipationExpectation{shape: ambientTxRequired}, true
	}
	for _, decl := range file.Decls {
		helper, ok := decl.(*ast.FuncDecl)
		if !ok || helper.Recv != nil || helper.Name == nil {
			continue
		}
		helperName := helper.Name.Name
		if funcReturnsDBOrTxHandle(helper) &&
			funcUsesProducedDBHandle(method, isLocalHelperCall(helperName)) {
			return ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: helperName,
			}, true
		}
		if _, ok := funcForwardsProducedHandleToLocalHelper(
			method,
			isPersistenceTxFromContext,
			helperName,
		); ok {
			expectation := ambientTxParticipationExpectation{
				shape:      ambientTxRequiredViaLocalHelper,
				helperName: helperName,
			}
			if funcMatchesAmbientTxExpectation(file, method, expectation) {
				return expectation, true
			}
		}
	}
	return ambientTxParticipationExpectation{}, false
}

func parseAmbientTxSourceFile(t *testing.T, keyFile string, src []byte) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, keyFile, src, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", keyFile, err)
	}
	return file
}

func walkRepositoryForAmbientTxExpectations(
	t *testing.T,
) map[string]ambientTxParticipationExpectation {
	t.Helper()
	found := make(map[string]ambientTxParticipationExpectation)
	for rawKey, src := range moduleInternalSource(t) {
		keyFile := legacyLintKey(rawKey)
		file := parseAmbientTxSourceFile(t, keyFile, src)
		for _, decl := range file.Decls {
			method, ok := decl.(*ast.FuncDecl)
			if !ok || method.Recv == nil {
				continue
			}
			expectation, ok := detectAmbientTxParticipationExpectation(file, method)
			if !ok {
				continue
			}
			found[keyFile+"|"+receiverMethodKey(method)] = expectation
		}
	}
	return found
}

// walkRepositoryForDBOrTx enumerates every allowlisted ambient-transaction participant
// (module-wide; BE9-1), keyed by "<file> | <ReceiverType>.<Method>". Most participants call
// DBOrTx directly. Explicit expectations above preserve wrapped DBOrTx and required-transaction
// shapes without making the scanner chase arbitrary helpers. Discovery is module-wide via
// moduleInternalSource (internal/repository plus every other internal/ package); keys for
// internal/repository/** files are legacyLintKey-normalized so existing allowlist entries keep
// matching unchanged.
func walkRepositoryForDBOrTx(t *testing.T) map[string]struct{} {
	t.Helper()
	found := map[string]struct{}{}
	tree := moduleInternalSource(t)
	for rawKey, src := range tree {
		keyFile := legacyLintKey(rawKey)
		f := parseAmbientTxSourceFile(t, keyFile, src)
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil {
				continue
			}
			methodKey := keyFile + "|" + receiverMethodKey(fd)
			if expectation, ok := ambientTxParticipationExpectations[methodKey]; ok {
				if funcMatchesAmbientTxExpectation(f, fd, expectation) {
					found[methodKey] = struct{}{}
				}
				continue
			}
			_, usesEquivalentShape := detectAmbientTxParticipationExpectation(f, fd)
			if funcUsesDBOrTx(fd) || usesEquivalentShape {
				found[methodKey] = struct{}{}
			}
		}
	}
	return found
}

// reconcileDBOrTxInventory is the pure bidirectional check.
func reconcileDBOrTxInventory(found, allow map[string]struct{}) []string {
	var violations []string
	for key := range found {
		if _, ok := allow[key]; !ok {
			violations = append(violations,
				"repository method "+key+" newly participates via DBOrTx but is NOT on dbOrTxParticipatingMethods. "+
					"Add it to the allowlist AND ensure a tx atomicity/isolation test covers its ambient-tx "+
					"participation (e.g. *_tx_atomicity_test.go). This gate forces that review.")
		}
	}
	for key := range allow {
		if _, ok := found[key]; !ok {
			violations = append(violations,
				"allowlisted ambient-tx method "+key+" no longer matches its required DBOrTx, local-helper, or "+
					"TxFromContext participation shape (or was renamed/removed). This is a tx-participation "+
					"REGRESSION (R1-1/R1-2): the method may silently escape or weaken an ambient WithTx → "+
					"partial-commit/TOCTOU. Restore the required shape, or if the method was intentionally "+
					"removed/renamed, delete the stale allowlist entry.")
		}
	}
	return violations
}

// ─── Gate tests ─────────────────────────────────────────────────────────────────────

// TestDBOrTxInventory_MatchesAllowlist is the gate. Floor guards against a vacuous pass.
func TestDBOrTxInventory_MatchesAllowlist(t *testing.T) {
	found := walkRepositoryForDBOrTx(t)
	if len(found) < 30 {
		t.Fatalf("only %d dbOrTx-using methods found; discovery/AST likely broke (would vacuously pass). "+
			"Expected the R1-1/R1-2 + uniform reservation/staff/shift/trimming surface (~80).", len(found))
	}
	for _, v := range reconcileDBOrTxInventory(found, dbOrTxParticipatingMethods) {
		t.Error(v)
	}
	detectedExpectations := walkRepositoryForAmbientTxExpectations(t)
	for key, detected := range detectedExpectations {
		registered, ok := ambientTxParticipationExpectations[key]
		if !ok {
			t.Errorf(
				"repository method %s uses a non-direct ambient-tx shape but has no "+
					"ambientTxParticipationExpectations entry", key,
			)
			continue
		}
		if registered != detected {
			t.Errorf(
				"repository method %s ambient-tx shape changed: registered=%+v detected=%+v",
				key,
				registered,
				detected,
			)
		}
	}
}

// TestDBOrTxInventory_DiscoveryReachesModuleWideAndNestedPackages pins that the module-wide
// discovery set (moduleInternalSource, backed by lintscan.WalkInternalTreeT; BE9-1)
// walkRepositoryForDBOrTx iterates over reaches: (a) a real 2+-level production package,
// (b) at least one file from a DIFFERENT top-level internal/ package, and (c) arbitrary deeper
// nesting (scanner capability, proven via a synthetic tree).
//
// Renamed + strengthened from the pre-BE9-1
// TestDBOrTxInventory_WalksAllEmbeddedFilesIncludingSubpackages, which only pinned the go:embed
// glob's 1-level reach within internal/repository.
func TestDBOrTxInventory_DiscoveryReachesModuleWideAndNestedPackages(t *testing.T) {
	tree := moduleInternalSource(t)
	if _, ok := tree["infra/smtp/sender.go"]; !ok {
		t.Fatal("module-wide discovery does not include infra/smtp/sender.go; " +
			"walkRepositoryForDBOrTx may have narrowed and would silently drop nested production packages")
	}
	// Reaching this line already proves every discovered file parsed cleanly: walkRepositoryForDBOrTx
	// calls t.Fatalf internally on any parse failure for ANY discovered file.
	found := walkRepositoryForDBOrTx(t)
	sawTargetPackageKey := false
	for k := range found {
		if strings.Contains(k, "/") {
			sawTargetPackageKey = true
			break
		}
	}
	if !sawTargetPackageKey {
		t.Fatal("walkRepositoryForDBOrTx found no dbOrTx-using method keyed under a target package path " +
			"(e.g. reservation/reservation_repository.go|...); discovery may have collapsed filenames")
	}

	assertDiscoversFileFromDifferentTopLevelPackage(t, tree)
	assertLintscanReachesTwoOrMoreNestingLevels(t)
}

// TestDBOrTxInventory_Analyzer pins the dbOrTx detector on inline fixtures.
func TestDBOrTxInventory_Analyzer(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "method using dbOrTx is detected",
			src: `package p
func (r *fooRepository) Bar() { _ = dbOrTx(ctx, r.db).Find(&x) }`,
			want: true,
		},
		{
			name: "method using only r.db.WithContext is NOT detected",
			src: `package p
func (r *fooRepository) Baz() { _ = r.db.WithContext(ctx).Find(&x) }`,
			want: false,
		},
		{
			name: "dbOrTx inside nested closure is detected",
			src: `package p
func (r *fooRepository) Qux() { _ = dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error { return nil }) }`,
			want: true,
		},
		{
			name: "persistence.DBOrTx selector is detected",
			src: `package p
func (r *fooRepository) Canonical() { _ = persistence.DBOrTx(ctx, r.db).Find(&x) }`,
			want: true,
		},
		{
			name: "same-package DBOrTx Ident is detected",
			src: `package p
func (r *fooRepository) Cap() { _ = DBOrTx(ctx, r.db).Find(&x) }`,
			want: true,
		},
		{
			name: "unrelated selector is NOT detected",
			src: `package p
func (r *fooRepository) Nope() { _ = other.DBOrTx(ctx, r.db).Find(&x) }`,
			want: false,
		},
		{
			name: "reorder helper alone is NOT detected (ambient path documented separately)",
			src: `package p
func (r *fooRepository) Reorder() { _ = reorderByClinicID(ctx, r.db, m, "x", 1, ids) }`,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "fixture.go", []byte(tc.src), 0)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			fd := f.Decls[0].(*ast.FuncDecl)
			if funcUsesDBOrTx(fd) != tc.want {
				t.Fatalf("funcUsesDBOrTx = %v, want %v", funcUsesDBOrTx(fd), tc.want)
			}
		})
	}
}

// TestDBOrTxInventory_AnalyzerDetectsUsageUnderNestedPathFilename is the dbortx-specific
// counterpart of the preload/audit-tx "AnalyzerDetectsViolationUnderNestedPathFilename" tests
// (BE9-1: dbortx lacked one — see the precedent map in the BE9-1 task notes). funcUsesDBOrTx
// itself takes no filename — it inspects only the parsed *ast.FuncDecl body — so this proves
// dbOrTx-usage detection is identical regardless of the filename/location shape used to parse the
// SAME source: a repository-root-shaped filename, a DIFFERENT top-level internal/ package
// filename, and a 2+-level nested filename must all detect the SAME dbOrTx usage.
func TestDBOrTxInventory_AnalyzerDetectsUsageUnderNestedPathFilename(t *testing.T) {
	src := `package p
func (r *fooRepository) Bar() { _ = dbOrTx(ctx, r.db).Find(&x) }`
	filenames := []string{
		"repository/x.go",          // root-shaped
		"service/x.go",             // a different top-level internal/ package
		"repository/two/deep/x.go", // 2+-level nested
	}
	for _, fn := range filenames {
		t.Run(fn, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, fn, []byte(src), 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			fd := f.Decls[0].(*ast.FuncDecl)
			if !funcUsesDBOrTx(fd) {
				t.Fatalf("dbOrTx usage not detected under filename %q", fn)
			}
		})
	}
}

// TestDBOrTxInventory_ReorderHelpersUseDBOrTx pins the ambient-tx contract for Reorder:
// persistence.ReorderByClinicID / ReorderGlobal must call DBOrTx so domain Reorder methods
// that only delegate to those helpers still join ambient WithTx (paymentmethod, cage, …).
func TestDBOrTxInventory_ReorderHelpersUseDBOrTx(t *testing.T) {
	tree := moduleInternalSource(t)

	src, ok := tree["persistence/scope.go"]
	if !ok {
		t.Fatal("persistence/scope.go not found in module-wide discovery set")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "persistence/scope.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := map[string]bool{
		"ReorderByClinicID": false,
		"ReorderGlobal":     false,
	}
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv != nil || fd.Name == nil {
			continue
		}
		if _, ok := want[fd.Name.Name]; !ok {
			continue
		}
		if !funcUsesDBOrTx(fd) {
			t.Errorf("persistence.%s must call DBOrTx (ambient Reorder contract)", fd.Name.Name)
			continue
		}
		want[fd.Name.Name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("persistence.%s not found in scope.go", name)
		}
	}

	// paymentmethod.Reorder must keep delegating to the DBOrTx-aware reorder helper
	// (not r.db.WithContext Transaction).
	pmSrc, ok := tree["billing/payment_method_master_repository.go"]
	if !ok {
		t.Fatal("billing/payment_method_master_repository.go not found in module-wide discovery set")
	}
	pmFset := token.NewFileSet()
	pmF, err := parser.ParseFile(pmFset, "billing/payment_method_master_repository.go", pmSrc, 0)
	if err != nil {
		t.Fatalf("parse paymentmethod: %v", err)
	}
	foundReorder := false
	for _, decl := range pmF.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name == nil || fd.Name.Name != "Reorder" || fd.Recv == nil {
			continue
		}
		foundReorder = true
		// Must call a local helper or persistence.ReorderByClinicID.
		callsHelper := false
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := ce.Fun.(*ast.Ident); ok && id.Name == "reorderByClinicID" {
				callsHelper = true
				return false
			}
			if sel, ok := ce.Fun.(*ast.SelectorExpr); ok && sel.Sel != nil && sel.Sel.Name == "ReorderByClinicID" {
				callsHelper = true
				return false
			}
			return true
		})
		if !callsHelper {
			t.Error("paymentmethod repository Reorder must call a DBOrTx-aware reorder helper")
		}
	}
	if !foundReorder {
		t.Error("paymentmethod.repository.Reorder not found")
	}
}

// TestDBOrTxInventory_GateDetectsViolations freezes the gate's failure modes.
func TestDBOrTxInventory_GateDetectsViolations(t *testing.T) {
	base := map[string]struct{}{"billing/accounting_repository.go|accountingRepository.SavePayment": {}}

	t.Run("clean baseline reports nothing", func(t *testing.T) {
		if v := reconcileDBOrTxInventory(base, base); len(v) != 0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})
	t.Run("new unlisted dbOrTx method fails", func(t *testing.T) {
		found := map[string]struct{}{
			"billing/accounting_repository.go|accountingRepository.SavePayment": {},
			"new_repository.go|newRepository.DoTx":                              {},
		}
		v := reconcileDBOrTxInventory(found, base)
		if len(v) != 1 || !strings.Contains(v[0], "newly participates via DBOrTx") {
			t.Fatalf("expected new-method violation, got %v", v)
		}
	})
	t.Run("reverted method (stale entry) fails as regression", func(t *testing.T) {
		v := reconcileDBOrTxInventory(map[string]struct{}{}, base)
		if len(v) != 1 || !strings.Contains(v[0], "REGRESSION") {
			t.Fatalf("expected regression violation, got %v", v)
		}
	})
}

func TestDBOrTxInventory_EquivalentParticipationAnalyzer(t *testing.T) {
	cases := []struct {
		name        string
		src         string
		expectation ambientTxParticipationExpectation
		want        bool
	}{
		{
			name: "local helper backed by DBOrTx is accepted",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	return persistence.DBOrTx(ctx, fallback)
}
func (r *fooRepository) Create() { _ = silentDB(ctx, r.db).Create(&x) }`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: true,
		},
		{
			name: "DBOrTx-backed helper assigned before use is accepted",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	db := persistence.DBOrTx(ctx, fallback)
	return db.Session(&gorm.Session{})
}
func (r *fooRepository) Create() {
	db := silentDB(ctx, r.db)
	_ = db.Create(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: true,
		},
		{
			name: "local helper that bypasses DBOrTx is rejected",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	return fallback.WithContext(ctx)
}
func (r *fooRepository) Create() { _ = silentDB(ctx, r.db).Create(&x) }`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: false,
		},
		{
			name: "helper that discards DBOrTx and returns fallback is rejected",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	_ = persistence.DBOrTx(ctx, fallback)
	return fallback.WithContext(ctx)
}
func (r *fooRepository) Create() { _ = silentDB(ctx, r.db).Create(&x) }`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: false,
		},
		{
			name: "helper that rebinds DBOrTx handle to fallback is rejected",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	db := persistence.DBOrTx(ctx, fallback)
	db = fallback.WithContext(ctx)
	return db
}
func (r *fooRepository) Create() { _ = silentDB(ctx, r.db).Create(&x) }`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: false,
		},
		{
			name: "method that stops calling its expected helper is rejected",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	return persistence.DBOrTx(ctx, fallback)
}
func (r *fooRepository) Create() { _ = r.db.WithContext(ctx).Create(&x) }`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: false,
		},
		{
			name: "method that discards helper result and uses base DB is rejected",
			src: `package p
func silentDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	return persistence.DBOrTx(ctx, fallback)
}
func (r *fooRepository) Create() {
	_ = silentDB(ctx, r.db)
	_ = r.db.WithContext(ctx).Create(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxViaLocalDBOrTxHelper,
				helperName: "silentDB",
			},
			want: false,
		},
		{
			name: "required ambient transaction is accepted",
			src: `package p
func (r *fooRepository) Update() {
	tx := persistence.TxFromContext(ctx)
	_ = tx.WithContext(ctx).Updates(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape: ambientTxRequired,
			},
			want: true,
		},
		{
			name: "required ambient transaction cannot weaken to fallback DBOrTx",
			src: `package p
func (r *fooRepository) Update() {
	_ = persistence.DBOrTx(ctx, r.db).Updates(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape: ambientTxRequired,
			},
			want: false,
		},
		{
			name: "required ambient transaction result cannot be discarded",
			src: `package p
func (r *fooRepository) Update() {
	_ = persistence.TxFromContext(ctx)
	_ = r.db.WithContext(ctx).Updates(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape: ambientTxRequired,
			},
			want: false,
		},
		{
			name: "required ambient transaction handle cannot be rebound to base DB",
			src: `package p
func (r *fooRepository) Update() {
	tx := persistence.TxFromContext(ctx)
	tx = r.db
	_ = tx.WithContext(ctx).Updates(&x)
}`,
			expectation: ambientTxParticipationExpectation{
				shape: ambientTxRequired,
			},
			want: false,
		},
		{
			name: "required ambient transaction forwarded to local DB helper is accepted",
			src: `package p
func writeWithTx(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(&x).Error
}
func (r *fooRepository) Create() error {
	tx := persistence.TxFromContext(ctx)
	return writeWithTx(ctx, tx)
}`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxRequiredViaLocalHelper,
				helperName: "writeWithTx",
			},
			want: true,
		},
		{
			name: "forwarded helper that ignores required transaction is rejected",
			src: `package p
func writeWithTx(ctx context.Context, tx *gorm.DB) error {
	return baseDB.WithContext(ctx).Create(&x).Error
}
func (r *fooRepository) Create() error {
	tx := persistence.TxFromContext(ctx)
	return writeWithTx(ctx, tx)
}`,
			expectation: ambientTxParticipationExpectation{
				shape:      ambientTxRequiredViaLocalHelper,
				helperName: "writeWithTx",
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "fixture.go", []byte(tc.src), 0)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}
			var method *ast.FuncDecl
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if ok && fd.Recv != nil {
					method = fd
					break
				}
			}
			if method == nil {
				t.Fatal("fixture method not found")
			}
			got := funcMatchesAmbientTxExpectation(f, method, tc.expectation)
			if got != tc.want {
				t.Fatalf("funcMatchesAmbientTxExpectation = %v, want %v", got, tc.want)
			}
		})
	}
}
