package repository

// dbortx_inventory_lint_test.go — BE-refactor follow-up (P1 harness): dbOrTx 参加メソッドの inventory ゲート
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
// ─── Technique ──────────────────────────────────────────────────────────────────────
//
// preload_clinic_scope_lint_test.go の repoSourceFS（go:embed *.go */*.go）と baseFileName / receiverMethodKey
// （audit_tx_inventory_lint_test.go）を再利用。各 FuncDecl 本体に `dbOrTx(` 呼び出しがあるものを
// (baseFile | ReceiverType.Method) で列挙し、allowlist と双方向突合する。
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
	"account_repository.go|accountRepository.Create":      {},
	"account_repository.go|accountRepository.FindByEmail": {},
	"account_repository.go|accountRepository.FindByID":    {},
	"account_repository.go|accountRepository.Update":      {},
	// accounting (R1-1 money-path atomicity; Create/CompleteAccountingAppointments added for
	// BE-refactor.md X-12 billing-complete-appt-post-tx atomicity fix)
	"accounting_repository.go|accountingRepository.CompleteAccountingAppointments": {},
	"accounting_repository.go|accountingRepository.Create":                         {},
	"accounting_repository.go|accountingRepository.LockAndFindByID":                {},
	"accounting_repository.go|accountingRepository.SavePayment":                    {},
	"accounting_repository.go|accountingRepository.SavePaymentSplits":              {},
	"accounting_repository.go|accountingRepository.Update":                         {},
	// audit (#211 tx-internal)
	"audit_repository.go|auditRepository.CreateTx": {},
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
	"dailyrecord/repository.go|repository.CreateVitalRecord":  {}, // BE8-4 batch6: moved from daily_record_repository.go
	"dailyrecord/repository.go|repository.FindOrCreateByDate": {}, // BE8-4 batch6: moved from daily_record_repository.go
	// checkup_field (#211 tx-internal replace)
	"checkup_field_repository.go|checkupFieldResultRepository.FindByCheckupID":   {},
	"checkup_field_repository.go|checkupFieldResultRepository.ReplaceForCheckup": {},
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
	"medical_record_image_repository.go|medicalRecordImageRepository.Create":   {},
	"medical_record_image_repository.go|medicalRecordImageRepository.Delete":   {},
	"medical_record_image_repository.go|medicalRecordImageRepository.FindByID": {},

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
	"prescription/repository.go|repository.Create": {}, // BE8-4 batch7: moved from prescription_repository.go
	"prescription/repository.go|repository.Update": {}, // BE8-4 batch7: moved from prescription_repository.go
	// prescription Delete (BE-refactor.md H-8e: prescriptionService.Delete が finalize ロック確認・
	// Delete を s.transactor.WithTx で束ねるようになったための追加。examination Delete=H-8d と同型)
	"prescription/repository.go|repository.Delete": {}, // BE8-4 batch7: moved from prescription_repository.go
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
	"staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.CountByStaffAndClinic": {},
	"staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.Create":                {},
	"staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.Delete":                {},
	"staff_clinic_assignment_repository.go|staffClinicAssignmentRepository.FindByStaffID":         {},
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
	"shift_entry_repository.go|shiftEntryRepository.Create":           {},
	"shift_entry_repository.go|shiftEntryRepository.Delete":           {},
	"shift_entry_repository.go|shiftEntryRepository.ExistsByStaffID":  {},
	"shift_entry_repository.go|shiftEntryRepository.FindAll":          {},
	"shift_entry_repository.go|shiftEntryRepository.FindByID":         {},
	"shift_entry_repository.go|shiftEntryRepository.FindOnDutyStaffs": {},
	"shift_entry_repository.go|shiftEntryRepository.Update":           {},
	"shift_entry_repository.go|shiftEntryRepository.ReplaceBreaks":    {}, // G6-2 repo-internal tx replace
	// trimming detail (uniform dbOrTx)
	"trimming_repository.go|appointmentTrimmingDetailRepository.Create":              {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.FindByAppointmentID": {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.SetOptions":          {},
	"trimming_repository.go|appointmentTrimmingDetailRepository.Update":              {},
	// vital (X-11 Appendix-A finalize-child-write-race fix — same FK-deadlock rationale as examination)
	"vital_repository.go|vitalRepository.Create": {},
	"vital_repository.go|vitalRepository.Update": {},
	"vital_repository.go|vitalRepository.Delete": {},
	// G6-2 (BE-refactor.md tx-mechanism-consolidation): repo-internal r.db.WithContext(ctx).Transaction
	// → dbOrTx(ctx, r.db).Transaction conversion, no ambient-tx caller into any of these (verified per-file).
	"manualarticle/repository.go|repository.Upsert":                                     {}, // BE8-4 batch3: moved from manual_article_repository.go
	"owner_repository.go|ownerRepository.CreateWithPets":                                {},
	"reservation_schedule_repository.go|reservationScheduleRepository.Save":             {},
	"reservation_type_liff_repository.go|reservationTypeLiffRepository.UpdateSortOrder": {},
	"shifttemplate/repository.go|repository.UpdateBreaks":                               {}, // BE8-4 batch12: moved from shift_template_repository.go
	"treatment_repository.go|treatmentRepository.BulkUpdateSortOrder":                   {},
	// X-6 (Appendix-A tx-atomicity fix, commit d7eff8c8): medicine/inventory repo-internal
	// r.db.WithContext(ctx).Transaction → dbOrTx(ctx, r.db).Transaction. Allowlist backfill
	// discovered during G6-2 (X-6 landed without registering these).
	"medicine_repository.go|medicineRepository.Create":                            {},
	"medicine_repository.go|medicineRepository.Update":                            {},
	"medicine_repository.go|medicineRepository.Delete":                            {},
	"medicine_repository.go|medicineRepository.FindAll":                           {},
	"medicine_repository.go|medicineRepository.FindByID":                          {},
	"medicine_repository.go|medicineRepository.CountChildrenByParentID":           {},
	"medicine_repository.go|medicineRepository.CountUsageByMedicineID":            {},
	"inventory_repository.go|inventoryRepository.Create":                          {},
	"inventory_repository.go|inventoryRepository.Update":                          {},
	"inventory_repository.go|inventoryRepository.Delete":                          {},
	"inventory_repository.go|inventoryRepository.FindAll":                         {},
	"inventory_repository.go|inventoryRepository.FindByID":                        {},
	"inventory_repository.go|inventoryRepository.DecreaseStock":                   {},
	"inventory_repository.go|inventoryRepository.CountUsageByInventoryID":         {},
	"inventory_repository.go|inventoryRepository.UpdateNameByMedicineCategory":    {},
	"inventory_repository.go|inventoryRepository.DeleteByNameAndMedicineCategory": {},
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

// walkRepositoryForDBOrTx enumerates every repository method that calls dbOrTx, keyed by
// "<file> | <ReceiverType>.<Method>". Includes domain subpackages embedded in repoSourceFS.
func walkRepositoryForDBOrTx(t *testing.T) map[string]struct{} {
	t.Helper()
	found := map[string]struct{}{}
	for _, name := range listEmbeddedRepoGoFiles(t) {
		src, err := repoSourceFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", name, err)
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		// Root allowlist entries use basenames; nested paths keep relative path to avoid collisions.
		keyFile := baseFileName(name)
		if strings.Contains(name, "/") {
			keyFile = name
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
		t.Fatalf("only %d dbOrTx-using methods found; embed/AST likely broke (would vacuously pass). "+
			"Expected the R1-1/R1-2 + uniform reservation/staff/shift/trimming surface (~80).", len(found))
	}
	for _, v := range reconcileDBOrTxInventory(found, dbOrTxParticipatingMethods) {
		t.Error(v)
	}
}

// TestDBOrTxInventory_WalksAllEmbeddedFilesIncludingSubpackages pins that walkRepositoryForDBOrTx
// processes every 1-level domain subpackage file listEmbeddedRepoGoFiles(t) returns, with no
// additional per-lint filtering. See TestRepoSourceEmbed_ReachesOneLevelSubpackages
// (preload_clinic_scope_lint_test.go) for the underlying embed-glob regression test that this
// wrapper depends on (BE-refactor.md BE8-0).
func TestDBOrTxInventory_WalksAllEmbeddedFilesIncludingSubpackages(t *testing.T) {
	names := listEmbeddedRepoGoFiles(t)
	nested := 0
	for _, n := range names {
		if strings.Contains(n, "/") {
			nested++
		}
	}
	if nested == 0 {
		t.Fatal("no 1-level subpackage files in the embedded set walkRepositoryForDBOrTx iterates over")
	}
	// Reaching this line already proves every nested file parsed cleanly: walkRepositoryForDBOrTx
	// calls t.Fatalf internally on any parse failure for ANY embedded file, subpackage included.
	found := walkRepositoryForDBOrTx(t)
	sawNestedKey := false
	for k := range found {
		if strings.Contains(k, "/") {
			sawNestedKey = true
			break
		}
	}
	if !sawNestedKey {
		t.Fatal("walkRepositoryForDBOrTx found no dbOrTx-using method keyed under a subpackage path " +
			"(e.g. reservationtype/repository.go|...); either the embed stopped reaching subpackages, " +
			"or the reservationtype dbOrTx usages were removed")
	}
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

// TestDBOrTxInventory_ReorderHelpersUseDBOrTx pins the ambient-tx contract for Reorder:
// repohelpers.ReorderByClinicID / ReorderGlobal must call DBOrTx so domain Reorder methods
// that only delegate to those helpers still join ambient WithTx (paymentmethod, cage, …).
func TestDBOrTxInventory_ReorderHelpersUseDBOrTx(t *testing.T) {
	src, err := repoSourceFS.ReadFile("repohelpers/scope.go")
	if err != nil {
		t.Fatalf("read repohelpers/scope.go: %v", err)
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
	pmSrc, err := repoSourceFS.ReadFile("paymentmethod/repository.go")
	if err != nil {
		t.Fatalf("read paymentmethod/repository.go: %v", err)
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
