package repository

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
// 本ゲートは「現在 dbOrTx を使う（= ambient tx 参加が意図された）メソッド集合」を inventory として固定する:
//   - 固定メソッドが `dbOrTx` 使用をやめた（`r.db.WithContext` へ revert 等）→ tx 参加の regression → fail。
//   - 未登録の新規メソッドが dbOrTx を使い始めた → allowlist 追加を強制 → レビュー時に「ambient tx
//     参加が正しいか・atomicity/isolation テストを添えたか」を必ず問う。
//
// ─── 意図的にやらないこと（taint 解析の限界・#124 の教訓） ────────────────────────────
//
// 「WithTx 内で呼ばれるのに dbOrTx を使っていない（= 参加漏れ）メソッド」の検出は、service→repository を
// 跨ぐ手続き間データフロー解析（どの repo メソッドが WithTx クロージャ内で呼ばれるか）が必須で、go/ast
// 単体では信頼できる規則が書けない（master_fk_write_inventory_lint と同じ taint 断念）。よって本ゲートは
// 「dbOrTx を使う surface の固定と regression 検出」に絞る。参加漏れの正本ガードは各 tx フローの
// atomicity テスト（accounting_repository_tx_atomicity_test.go / refund_repository_sum_tx_participation_test.go
// / checkup_field_result_tx_atomicity_test.go 等）が担う。
//
// ─── Static scanning blind spots (BE9-1) ───────────────────────────────────────────
//
// This gate is a syntactic match over literal AST call shapes (see the `funcUsesDBOrTx` doc
// comment below for the exact three shapes matched). It cannot see:
//   - a dbOrTx-equivalent participation hidden behind a raw SQL string, or a renamed/aliased
//     import of the dbOrTx helper (shadowing/aliasing is a documented existing limitation, not
//     new to BE9-1);
//   - a method that SHOULD participate in an ambient tx but calls a helper that itself lacks
//     dbOrTx (the taint-analysis limitation already documented above);
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
// 各 FuncDecl 本体に `dbOrTx(` 呼び出しがあるものを (keyFile | ReceiverType.Method) で列挙し、
// allowlist と双方向突合する。internal/repository/** 由来のキーは legacyLintKey により旧
// repoSourceFS 相当の形（basename / 1階層以上の相対パス）へ正規化され、既存 allowlist のキー形が
// そのまま一致し続ける。
//
// 注（syntactic match・sibling lint と同一設計）: `funcUsesDBOrTx` は
//   1) Ident `dbOrTx`（parent package wrapper）
//   2) Ident `DBOrTx`（same-package free name, e.g. inside repohelpers）
//   3) Selector `repohelpers.DBOrTx`（domain subpackages）
// を検出する。go/types の意味解決はしない。シャドーイングや別名 import は誤検知/見逃しの既知限界
// （preload/audit-tx lint と同じ割り切り）。
//
// ambient Reorder 方針: `Reorder` が `reorderByClinicID` / `repohelpers.ReorderByClinicID` のみを
// 呼ぶ経路はメソッド本体に `dbOrTx`/`DBOrTx` が現れない。参加保証の正本は
// `repohelpers.ReorderByClinicID`/`ReorderGlobal` 内の `DBOrTx`（free func・本 inventory の
// receiver 走査外）であり、TestDBOrTxInventory_ReorderHelpersUseDBOrTx で固定する。
// ドメイン Reorder を method inventory に載せる必要はない（allowlist 膨張と偽回帰を避ける）。

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// dbOrTxParticipatingMethods は現 HEAD で dbOrTx(ctx, r.db) を使う repository メソッド（実測 80 個）を
// 固定する（key = "<file> | <ReceiverType>.<Method>"）。R1-1/R1-2 の tx 参加 surface を含む。
// 追加/削除時はこのマップを更新し、新規は対応する atomicity/isolation テストを添えること。
var dbOrTxParticipatingMethods = map[string]struct{}{
	// account
	"account/repository.go|repository.Create":      {}, // BE8-4 batch23: moved from account_repository.go
	"account/repository.go|repository.FindByEmail": {}, // BE8-4 batch23: moved from account_repository.go
	"account/repository.go|repository.FindByID":    {}, // BE8-4 batch23: moved from account_repository.go
	"account/repository.go|repository.Update":      {}, // BE8-4 batch23: moved from account_repository.go
	// accounting (R1-1 money-path atomicity; Create/CompleteAccountingAppointments added for
	// BE-refactor.md X-12 billing-complete-appt-post-tx atomicity fix)
	"accounting_repository.go|accountingRepository.CompleteAccountingAppointments": {},
	"accounting_repository.go|accountingRepository.Create":                         {},
	"accounting_repository.go|accountingRepository.LockAndFindByID":                {},
	"accounting_repository.go|accountingRepository.SavePayment":                    {},
	"accounting_repository.go|accountingRepository.SavePaymentSplits":              {},
	"accounting_repository.go|accountingRepository.Update":                         {},
	// audit (#211 tx-internal)
	"audit/repository.go|repository.CreateTx": {}, // BE8-4 batch22: moved from audit_repository.go
	// billing_confirmation (SD-2 系ガード監査: 会計医師確認 Confirm/Return が確定済みカルテ書込
	// ガード対象と判明。billingConfirmationService.Confirm/Return の LockByIDForUpdate ambient tx
	// に参加させる)
	"billing_confirmation_repository.go|billingConfirmationRepository.Update": {},
	// billing_item (R1-1)
	"billing_item_repository.go|billingItemRepository.Create":              {},
	"billing_item_repository.go|billingItemRepository.Delete":              {},
	"billing_item_repository.go|billingItemRepository.FindByBillingID":     {},
	"billing_item_repository.go|billingItemRepository.FindByID":            {},
	"billing_item_repository.go|billingItemRepository.Update":              {},
	"billing_item_repository.go|billingItemRepository.UpdateBillingTotals": {},
	// campaign
	"campaign/repository.go|repository.FindAllApplicableForItem": {}, // BE8-4 batch9: moved from campaign_repository.go
	"campaign/repository.go|repository.FindApplicableForItem":    {}, // BE8-4 batch9: moved from campaign_repository.go
	"campaign/repository.go|repository.ReplaceTargets":           {}, // BE8-4 batch9: moved from campaign_repository.go; G6-2 repo-internal tx replace
	// daily_record (AUD-006: FindOrCreate+CreateVital same ambient tx)
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.CreateVitalRecord":  {}, // BE8-4 batch6: moved from daily_record_repository.go
	"medicalrecord/daily_record_repository.go|dailyRecordRepository.FindOrCreateByDate": {}, // BE8-4 batch6: moved from daily_record_repository.go
	// care_plan_item / hospitalization (BE9-2D ⑤: DischargeWithBilling の repos.Transaction→
	// Transactor.WithTx 化。FOR UPDATE 直列化・退院status更新・care plan read を billing 書込と
	// 同一 ambient tx に参加させる＝二重会計防止)
	"medicalrecord/care_plan_item_repository.go|carePlanItemRepository.FindByHospitalizationID":   {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.LockByIDForUpdate":     {},
	"medicalrecord/hospitalization_repository.go|hospitalizationRepository.UpdateIfNotDischarged": {},
	// checkup_field (#211 tx-internal replace)
	"medicalrecord/checkup_field_repository.go|checkupFieldResultRepository.FindByCheckupID":   {},
	"medicalrecord/checkup_field_repository.go|checkupFieldResultRepository.ReplaceForCheckup": {},
	// estimate (SD-2 系ガード監査: 見積書 Create/Update/Delete が確定済みカルテ書込ガード対象と判明。
	// estimateService の LockByIDForUpdate ambient tx に参加させる。FindByID は
	// UpdateIfNotLocked/normalizeDeleteIfNotLockedMiss の tx 内再取得のため併せて追加)
	"estimate_repository.go|estimateRepository.Create":                 {},
	"estimate_repository.go|estimateRepository.FindByID":               {},
	"estimate_repository.go|estimateRepository.UpdateIfNotLocked":      {},
	"estimate_repository.go|estimateRepository.DeleteIfNotLocked":      {},
	"estimate_repository.go|estimateRepository.CountItemsByEstimateID": {},
	// examination (BE-refactor.md R1-2 tx-internal replace; Create/FindByID/Update added for X-11
	// finalize-child-write-race — must join the LockByIDForUpdate ambient tx or the FK check on
	// examinations.medical_record_id deadlocks against the FOR UPDATE row lock; Delete added for
	// H-8d — same finalize-lock race as Update, now WithTx-wrapped in examinationService.Delete)
	"examination_repository.go|examinationRepository.Create":               {},
	"examination_repository.go|examinationRepository.Delete":               {},
	"examination_repository.go|examinationRepository.FindAllItemsByExamID": {},
	"examination_repository.go|examinationRepository.FindByID":             {},
	"examination_repository.go|examinationRepository.ReplaceItemsByExamID": {},
	"examination_repository.go|examinationRepository.Update":               {},
	// medical_record_addendum
	"medical_record_addendum_repository.go|medicalRecordAddendumRepository.Create":                {},
	"medical_record_addendum_repository.go|medicalRecordAddendumRepository.FindByID":              {},
	"medical_record_addendum_repository.go|medicalRecordAddendumRepository.FindByMedicalRecordID": {},
	// medical_record (X-11 Appendix-A finalize-child-write-race fix)
	// SD-2: 確定済みカルテ画像ガード — Create/Delete/FindByID が LockByIDForUpdate の
	// ambient tx に参加する（medical_record_image_service.go の WithTx 内から呼ばれる）。
	// BE9-2D sub-batch④a: moved to internal/medicalrecord (facade kept in internal/repository).
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.Create":   {},
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.Delete":   {},
	"medicalrecord/medical_record_image_repository.go|medicalRecordImageRepository.FindByID": {},

	"medical_record_repository.go|medicalRecordRepository.LockByIDForUpdate":     {},
	"medical_record_repository.go|medicalRecordRepository.Create":                {},
	"medical_record_repository.go|medicalRecordRepository.Update":                {},
	"medical_record_repository.go|medicalRecordRepository.findMedicalRecordByID": {},
	// medicine_dose_param (R1-2 dose-param tx)
	"medicine_dose_param_repository.go|medicineDoseParamRepository.Create":                   {},
	"medicine_dose_param_repository.go|medicineDoseParamRepository.Delete":                   {},
	"medicine_dose_param_repository.go|medicineDoseParamRepository.FindByMedicineAndSpecies": {},
	"medicine_dose_param_repository.go|medicineDoseParamRepository.FindByMedicineID":         {},
	"medicine_dose_param_repository.go|medicineDoseParamRepository.Update":                   {},
	// pet (BUG-407: lstepLifecycleService.HandlePetDeath/HandlePetRevival が status/deceased_at
	// 更新と一次監査ログ書込を Transactor.WithTx で束ね fail-closed 化。runtime proof は
	// pet_repository_tx_atomicity_test.go)
	"pet_repository.go|petRepository.Update": {},
	// prescription (X-11 Appendix-A finalize-child-write-race fix — same FK-deadlock rationale as examination)
	"medicalrecord/prescription_repository.go|prescriptionRepository.Create": {}, // BE8-4 batch7: moved from prescription_repository.go
	"medicalrecord/prescription_repository.go|prescriptionRepository.Update": {}, // BE8-4 batch7: moved from prescription_repository.go
	// prescription Delete (BE-refactor.md H-8e: prescriptionService.Delete が finalize ロック確認・
	// Delete を s.transactor.WithTx で束ねるようになったための追加。examination Delete=H-8d と同型)
	"medicalrecord/prescription_repository.go|prescriptionRepository.Delete": {}, // BE8-4 batch7: moved from prescription_repository.go
	// refund (R1-1 TOCTOU)
	"refund/repository.go|repository.Create":                         {}, // BE8-4 batch8: moved from refund_repository.go
	"refund/repository.go|repository.SumByBillingID":                 {}, // BE8-4 batch8: moved from refund_repository.go
	"refund/repository.go|repository.SumByBillingIDAndPaymentMethod": {}, // BE8-4 batch8: moved from refund_repository.go
	// reservation (uniform dbOrTx)
	"reservation_repository.go|reservationRepository.AssertLineCustomerInClinic":         {}, // AUD-001
	"reservation_repository.go|reservationRepository.AssertOwnerInClinic":                {}, // AUD-001
	"reservation_repository.go|reservationRepository.AcquireBookingLock":                 {}, // X-9 (Appendix-A phantom-booking fix)
	"reservation_repository.go|reservationRepository.FindPetOwnerInClinic":               {}, // AUD-001
	"reservation_repository.go|reservationRepository.CountByCustomerAndDateRange":        {},
	"reservation_repository.go|reservationRepository.CountByDateAndSource":               {},
	"reservation_repository.go|reservationRepository.CountByTypeAndStartTime":            {},
	"reservation_repository.go|reservationRepository.CountByTypeAndStartTimes":           {},
	"reservation_repository.go|reservationRepository.CountConflicts":                     {},
	"reservation_repository.go|reservationRepository.CountMedicalRecordsByReservationID": {},
	"reservation_repository.go|reservationRepository.CountOnDutyDoctors":                 {},
	"reservation_repository.go|reservationRepository.Create":                             {},
	"reservation_repository.go|reservationRepository.Delete":                             {},
	"reservation_repository.go|reservationRepository.ExistsByReservationTypeID":          {},
	"reservation_repository.go|reservationRepository.ExistsByStaffID":                    {},
	"reservation_repository.go|reservationRepository.FindAll":                            {},
	"reservation_repository.go|reservationRepository.FindAllByCategory":                  {},
	"reservation_repository.go|reservationRepository.findReservationByID":                {},
	"reservation_repository.go|reservationRepository.HasDoctorConflict":                  {},
	"reservation_repository.go|reservationRepository.LockAndFindByID":                    {},
	"reservation_repository.go|reservationRepository.Update":                             {},
	"appointment_admin_repository.go|reservationAdminRepository.Create":                  {}, // AUD-001
	// reservationtype domain package (methods that previously used dbOrTx; Update/Delete remain
	// r.db.WithContext by design — behavior preserved from flat file; facade keeps service imports)
	"reservationtype/repository.go|repository.CountChildrenByParentID":       {},
	"reservationtype/repository.go|repository.CountUsageByReservationTypeID": {},
	"reservationtype/repository.go|repository.Create":                        {},
	"reservationtype/repository.go|repository.FindAll":                       {},
	"reservationtype/repository.go|repository.FindAllWithChildren":           {},
	"reservationtype/repository.go|repository.FindByID":                      {},
	"reservationtype/repository.go|repository.FindByIDWithChildren":          {},
	// staff_clinic_assignment
	"staffclinicassignment/repository.go|repository.CountByStaffAndClinic": {}, // BE8-4 batch19: moved from staff_clinic_assignment_repository.go
	"staffclinicassignment/repository.go|repository.Create":                {}, // BE8-4 batch19: moved from staff_clinic_assignment_repository.go
	"staffclinicassignment/repository.go|repository.Delete":                {}, // BE8-4 batch19: moved from staff_clinic_assignment_repository.go
	"staffclinicassignment/repository.go|repository.FindByStaffID":         {}, // BE8-4 batch19: moved from staff_clinic_assignment_repository.go
	// staff (uniform dbOrTx)
	"staff_repository.go|staffRepository.CountBlockingReferencesByStaffID": {},
	"staff_repository.go|staffRepository.Create":                           {},
	"staff_repository.go|staffRepository.Delete":                           {},
	"staff_repository.go|staffRepository.FindAll":                          {},
	"staff_repository.go|staffRepository.FindByAccountID":                  {},
	"staff_repository.go|staffRepository.FindByID":                         {},
	"staff_repository.go|staffRepository.Reorder":                          {},
	"staff_repository.go|staffRepository.Update":                           {},
	"staff_repository.go|staffRepository.UpdatePrimaryClinicID":            {},
	// shift_entry (uniform dbOrTx)
	"shiftentry/repository.go|repository.Create":           {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.Delete":           {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.ExistsByStaffID":  {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.FindAll":          {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.FindByID":         {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.FindOnDutyStaffs": {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.Update":           {}, // BE8-4 batch13: moved from shift_entry_repository.go
	"shiftentry/repository.go|repository.ReplaceBreaks":    {}, // BE8-4 batch13: moved from shift_entry_repository.go; G6-2 repo-internal tx replace
	// trimming detail (uniform dbOrTx)
	"trimming_repository.go|appointmentTrimmingDetailRepository.Create":              {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.FindByAppointmentID": {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.SetOptions":          {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.Update":              {},
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
	"manualarticle/repository.go|repository.Upsert":                                     {}, // BE8-4 batch3: moved from manual_article_repository.go
	"owner_repository.go|ownerRepository.CreateWithPets":                                {},
	"reservation_schedule_repository.go|reservationScheduleRepository.Save":             {},
	"reservation_type_liff_repository.go|reservationTypeLiffRepository.UpdateSortOrder": {},
	"shifttemplate/repository.go|repository.UpdateBreaks":                               {}, // BE8-4 batch12: moved from shift_template_repository.go
	// treatment (BE9-2D ④b: WithTx 化に伴う ambient tx 参加。④b Batch A で medicalrecord へ移動済み、
	// lockDraftMedicalRecord 行ロック・在庫減算・逸脱監査と同一 ambient tx へ参加させる)
	"medicalrecord/treatment_repository.go|treatmentRepository.Create":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.Delete":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.Update":              {},
	"medicalrecord/treatment_repository.go|treatmentRepository.BulkUpdateSortOrder": {},
	// X-6 (Appendix-A tx-atomicity fix, commit d7eff8c8): medicine/inventory repo-internal
	// r.db.WithContext(ctx).Transaction → dbOrTx(ctx, r.db).Transaction. Allowlist backfill
	// discovered during G6-2 (X-6 landed without registering these).
	"medicine_repository.go|medicineRepository.Create":                   {},
	"medicine_repository.go|medicineRepository.Update":                   {},
	"medicine_repository.go|medicineRepository.Delete":                   {},
	"medicine_repository.go|medicineRepository.FindAll":                  {},
	"medicine_repository.go|medicineRepository.FindByID":                 {},
	"medicine_repository.go|medicineRepository.CountChildrenByParentID":  {},
	"medicine_repository.go|medicineRepository.CountUsageByMedicineID":   {},
	"inventory/repository.go|repository.Create":                          {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.Update":                          {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.Delete":                          {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.FindAll":                         {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.FindByID":                        {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.DecreaseStock":                   {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.CountUsageByInventoryID":         {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.UpdateNameByMedicineCategory":    {}, // BE8-4 batch18: moved from inventory_repository.go
	"inventory/repository.go|repository.DeleteByNameAndMedicineCategory": {}, // BE8-4 batch18: moved from inventory_repository.go
	// X-7 (Appendix-A tx-atomicity fix, commit 2a7a4dfc): clinic/permission_group repo-internal
	// tx conversion. Allowlist backfill discovered during G6-2 (X-7 landed without registering these).
	"clinic_repository.go|clinicRepository.Create":                               {},
	"clinic_repository.go|clinicRepository.Update":                               {},
	"clinic_repository.go|clinicRepository.Delete":                               {},
	"clinic_repository.go|clinicRepository.FindByID":                             {},
	"clinic_repository.go|clinicRepository.FindCompany":                          {},
	"permission_group_repository.go|permissionGroupRepository.Create":            {},
	"permission_group_repository.go|permissionGroupRepository.Reorder":           {},
	"permission_group_repository.go|permissionGroupRepository.UpdateRules":       {},
	"permission_group_repository.go|permissionGroupRepository.UpdateStaffGroups": {},
	// X-8 (Appendix-A tx-atomicity fix, commit 1e2d483c): reservation_staff repo-internal tx
	// conversion. Allowlist backfill discovered during G6-2 (X-8 landed without registering these).
	"reservation_staff_repository.go|reservationStaffRepository.Create":                         {},
	"reservation_staff_repository.go|reservationStaffRepository.Update":                         {},
	"reservation_staff_repository.go|reservationStaffRepository.Delete":                         {},
	"reservation_staff_repository.go|reservationStaffRepository.UpdateSortOrder":                {},
	"reservation_staff_repository.go|reservationStaffRepository.UpdateExcludedReservationTypes": {},
	"reservation_staff_repository.go|reservationStaffRepository.UpdateReservationCapabilities":  {},
	// BE-refactor.md H-7: FindByID を dbOrTx 化し、reservationStaffService.Update の
	// tx 内所有権確認（GetByID）を ambient tx に参加させ TOCTOU 窓を閉じる。
	"reservation_staff_repository.go|reservationStaffRepository.FindByID": {},
}

// funcUsesDBOrTx reports whether a function body contains a call to dbOrTx / DBOrTx /
// repohelpers.DBOrTx(...). Does not chase helpers (see Reorder ambient policy in file header).
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
			// parent package wrapper `dbOrTx` or same-package `DBOrTx` (repohelpers).
			if fun.Name == "dbOrTx" || fun.Name == "DBOrTx" {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			// domain subpackages: repohelpers.DBOrTx(ctx, r.db)
			if fun.Sel != nil && fun.Sel.Name == "DBOrTx" {
				if id, ok := fun.X.(*ast.Ident); ok && id.Name == "repohelpers" {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

// walkRepositoryForDBOrTx enumerates every method (module-wide; BE9-1) that calls dbOrTx, keyed
// by "<file> | <ReceiverType>.<Method>". Discovery is module-wide via moduleInternalSource
// (internal/repository plus every other internal/ package); keys for internal/repository/**
// files are legacyLintKey-normalized so existing allowlist entries keep matching unchanged.
func walkRepositoryForDBOrTx(t *testing.T) map[string]struct{} {
	t.Helper()
	found := map[string]struct{}{}
	tree := moduleInternalSource(t)
	for rawKey, src := range tree {
		keyFile := legacyLintKey(rawKey)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, keyFile, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", keyFile, err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil {
				continue
			}
			if funcUsesDBOrTx(fd) {
				found[keyFile+"|"+receiverMethodKey(fd)] = struct{}{}
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
				"repository method "+key+" newly uses dbOrTx but is NOT on dbOrTxParticipatingMethods. "+
					"Add it to the allowlist AND ensure a tx atomicity/isolation test covers its ambient-tx "+
					"participation (e.g. *_tx_atomicity_test.go). This gate forces that review.")
		}
	}
	for key := range allow {
		if _, ok := found[key]; !ok {
			violations = append(violations,
				"allowlisted dbOrTx method "+key+" no longer calls dbOrTx (reverted to r.db.WithContext, or "+
					"renamed/removed). If reverted, this is a tx-participation REGRESSION (R1-1/R1-2): the method "+
					"will silently NOT join an ambient WithTx → partial-commit/TOCTOU. Restore dbOrTx, or if the "+
					"method was intentionally removed/renamed, delete the stale allowlist entry.")
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
}

// TestDBOrTxInventory_DiscoveryReachesModuleWideAndNestedPackages pins that the module-wide
// discovery set (moduleInternalSource, backed by lintscan.WalkInternalTreeT; BE9-1)
// walkRepositoryForDBOrTx iterates over reaches: (a) 1-level+ repository domain subpackages
// (walkRepositoryForDBOrTx must still discover a dbOrTx usage keyed under one), (b) at least one
// file from a DIFFERENT top-level internal/ package, and (c) 2+-level nesting (scanner
// capability, proven via a synthetic tree).
//
// Renamed + strengthened from the pre-BE9-1
// TestDBOrTxInventory_WalksAllEmbeddedFilesIncludingSubpackages, which only pinned the go:embed
// glob's 1-level reach within internal/repository.
func TestDBOrTxInventory_DiscoveryReachesModuleWideAndNestedPackages(t *testing.T) {
	tree := moduleInternalSource(t)
	nested := 0
	for n := range tree {
		if strings.HasPrefix(n, "repository/") && strings.Contains(legacyLintKey(n), "/") {
			nested++
		}
	}
	if nested == 0 {
		t.Fatal("no 1-level+ subpackage repository files in the module-wide discovered set walkRepositoryForDBOrTx iterates over")
	}
	// Reaching this line already proves every discovered file parsed cleanly: walkRepositoryForDBOrTx
	// calls t.Fatalf internally on any parse failure for ANY discovered file, subpackage included.
	found := walkRepositoryForDBOrTx(t)
	sawNestedKey := false
	for k := range found {
		if strings.Contains(k, "/") {
			sawNestedKey = true
			break
		}
	}
	if !sawNestedKey {
		t.Fatal("walkRepositoryForDBOrTx found no dbOrTx-using method keyed under a repository subpackage path " +
			"(e.g. reservationtype/repository.go|...); either discovery stopped reaching subpackages, " +
			"or the reservationtype dbOrTx usages were removed")
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
			name: "repohelpers.DBOrTx selector is detected",
			src: `package p
func (r *fooRepository) Sel() { _ = repohelpers.DBOrTx(ctx, r.db).Find(&x) }`,
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
// repohelpers.ReorderByClinicID / ReorderGlobal must call DBOrTx so domain Reorder methods
// that only delegate to those helpers still join ambient WithTx (paymentmethod, cage, …).
func TestDBOrTxInventory_ReorderHelpersUseDBOrTx(t *testing.T) {
	tree := moduleInternalSource(t)

	src, ok := tree["repository/repohelpers/scope.go"]
	if !ok {
		t.Fatal("repository/repohelpers/scope.go not found in module-wide discovery set")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "repohelpers/scope.go", src, 0)
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
			t.Errorf("repohelpers.%s must call DBOrTx (ambient Reorder contract)", fd.Name.Name)
			continue
		}
		want[fd.Name.Name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("repohelpers.%s not found in scope.go", name)
		}
	}

	// paymentmethod.Reorder must keep delegating to the DBOrTx-aware reorder helper
	// (not r.db.WithContext Transaction).
	pmSrc, ok := tree["repository/paymentmethod/repository.go"]
	if !ok {
		t.Fatal("repository/paymentmethod/repository.go not found in module-wide discovery set")
	}
	pmFset := token.NewFileSet()
	pmF, err := parser.ParseFile(pmFset, "paymentmethod/repository.go", pmSrc, 0)
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
		// Must call reorderByClinicID (local wrapper → repohelpers.ReorderByClinicID).
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
			t.Error("paymentmethod.repository.Reorder must call reorderByClinicID or repohelpers.ReorderByClinicID")
		}
	}
	if !foundReorder {
		t.Error("paymentmethod.repository.Reorder not found")
	}
}

// TestDBOrTxInventory_GateDetectsViolations freezes the gate's failure modes.
func TestDBOrTxInventory_GateDetectsViolations(t *testing.T) {
	base := map[string]struct{}{"accounting_repository.go|accountingRepository.SavePayment": {}}

	t.Run("clean baseline reports nothing", func(t *testing.T) {
		if v := reconcileDBOrTxInventory(base, base); len(v) != 0 {
			t.Fatalf("expected 0, got %v", v)
		}
	})
	t.Run("new unlisted dbOrTx method fails", func(t *testing.T) {
		found := map[string]struct{}{
			"accounting_repository.go|accountingRepository.SavePayment": {},
			"new_repository.go|newRepository.DoTx":                      {},
		}
		v := reconcileDBOrTxInventory(found, base)
		if len(v) != 1 || !strings.Contains(v[0], "newly uses dbOrTx") {
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
