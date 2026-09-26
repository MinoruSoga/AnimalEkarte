package billing

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
)

const (
	syntheticClosingSubtotal        = int64(1000)
	syntheticClosingTax             = int64(100)
	syntheticClosingTotal           = int64(1100)
	syntheticClosingCashSystemKey   = "cash"
	syntheticClosingCashDisplayName = "現金"
)

// SyntheticClosingRequest は S09 用の新規合成会計 5 件を作る入力。
type SyntheticClosingRequest struct {
	AppEnv             string
	DBHost             string
	TargetDate         time.Time
	PasswordHash       string
	ExistingBillingIDs []uint64
}

// SyntheticClosingResult は作成した使い捨て clinic と 5 件の完了時刻を返す。
type SyntheticClosingResult struct {
	ClinicID     uint64
	LoginEmail   string
	BillingIDs   []uint64
	CompletedAt  []time.Time
	CleanupToken string
}

// CreateSyntheticClosingFixture は新規 clinic / staff / 支払方法 / 明細 / 会計 5 件を原子的に作る。
// 既存 billings.id の UPDATE はしない。
func CreateSyntheticClosingFixture(ctx context.Context, db *gorm.DB, req SyntheticClosingRequest) (*SyntheticClosingResult, error) {
	if err := AllowUATSyntheticClosing(req.AppEnv, req.DBHost); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectExistingBillingIDs(req.ExistingBillingIDs); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if strings.TrimSpace(req.PasswordHash) == "" {
		return nil, apperrors.WrapInvalidInput("password hash is required")
	}
	if db == nil {
		return nil, apperrors.WrapInvalidInput("db is required")
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, apperrors.Wrap(err, "load Asia/Tokyo")
	}
	inJST := req.TargetDate.In(jst)
	day := time.Date(inJST.Year(), inJST.Month(), inJST.Day(), 0, 0, 0, 0, jst)
	if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		return nil, apperrors.WrapInvalidInput("target date must be a weekday")
	}
	completed := []time.Time{
		time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 13, 30, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 14, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 20, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 2, 0, 0, 0, jst).Add(24 * time.Hour),
	}

	var result *SyntheticClosingResult
	// ambient tx が ctx にあれば SAVEPOINT で join し、無ければ新規 tx を開く。
	// 委譲 collaborator には tx 束縛 ctx を渡し、persistence.DBOrTx が
	// 呼び出し側 ambient tx ではなくこの fixture tx を解決するようにする。
	base := persistence.DBOrTx(ctx, db)
	err = base.Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		clinicID := uint64(920000 + time.Now().UnixNano()%8000 + 1)
		if err := RejectReservedClinicID(clinicID); err != nil {
			return apperrors.WrapInvalidInput(err.Error())
		}

		company := &model.Company{Name: fmt.Sprintf("%s%d", syntheticClosingCompanyPrefix, clinicID)}
		if err := tx.Create(company).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic company")
		}
		clinic := &model.Clinic{
			ID:        clinicID,
			CompanyID: company.ID,
			Name:      fmt.Sprintf("%s%d", syntheticClosingClinicPrefix, clinicID),
			IsActive:  true,
		}
		if err := tx.Create(clinic).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic clinic")
		}
		settings := &model.ClinicSettings{
			ClinicID:            clinicID,
			ClosingAmStart:      "09:00",
			ClosingAmPmBoundary: "13:30",
			ClosingWeekdayEnd:   "19:00",
			ClosingSundayEnd:    "17:30",
		}
		if err := tx.Create(settings).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic closing settings")
		}

		account := &model.Account{
			Email:         SyntheticClosingLoginEmail(clinicID),
			PasswordHash:  req.PasswordHash,
			IsActive:      true,
			IsSystemAdmin: true,
		}
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic account")
		}
		staffRow := &model.Staff{
			ClinicID:  clinicID,
			AccountID: &account.ID,
			Name:      fmt.Sprintf("s09-staff-%d", clinicID),
			IsActive:  true,
			StaffType: model.StaffTypeDoctor,
		}
		if err := staff.CreateSyntheticClosingStaff(txCtx, tx, staffRow); err != nil {
			return apperrors.Wrap(err, "create synthetic staff")
		}
		assignment := &model.StaffClinicAssignment{StaffID: staffRow.ID, ClinicID: clinicID, IsMain: true}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.Wrap(err, "assign synthetic staff clinic")
		}

		owner := &model.Owner{ClinicID: clinicID, Name: fmt.Sprintf("s09-owner-%d", clinicID)}
		if err := tx.Create(owner).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic owner")
		}
		species := &model.AnimalSpecies{Name: fmt.Sprintf("s09-species-%d", clinicID)}
		if err := tx.Create(species).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic species")
		}
		pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: fmt.Sprintf("s09-pet-%d", clinicID)}
		if err := tx.Create(pet).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic pet")
		}

		paymentMethod, err := resolveSyntheticClosingCashPaymentMethod(tx, clinicID)
		if err != nil {
			return err
		}

		ids := make([]uint64, 0, len(completed))
		for _, at := range completed {
			atCopy := at
			clinicIDCopy := clinicID
			billing := &model.Billing{
				ClinicID:      clinicID,
				OwnerID:       &owner.ID,
				PetID:         &pet.ID,
				Subtotal:      syntheticClosingSubtotal,
				TaxTotal:      syntheticClosingTax,
				TotalAmount:   syntheticClosingTotal,
				Status:        model.BillingStatusCompleted,
				ScheduledDate: day,
				CompletedAt:   &atCopy,
				Memo:          "s09-synthetic",
			}
			if err := tx.Create(billing).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic billing")
			}
			item := &model.BillingItem{
				BillingID: billing.ID,
				ClinicID:  &clinicIDCopy,
				Category:  model.ItemCategoryExamination,
				Name:      "s09-item",
				UnitPrice: syntheticClosingSubtotal,
				Quantity:  1,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   0.10,
				Source:    model.ItemSourceManual,
			}
			if err := tx.Create(item).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic billing item")
			}
			payment := &model.Payment{
				BillingID:       billing.ID,
				ClinicID:        clinicID,
				Subtotal:        syntheticClosingSubtotal,
				TaxTotal:        syntheticClosingTax,
				TotalAmount:     syntheticClosingTotal,
				BillingAmount:   syntheticClosingTotal,
				ReceivedAmount:  syntheticClosingTotal,
				Method:          model.PaymentMethodCash,
				PaymentMethodID: &paymentMethod.ID,
				PaidBy:          &staffRow.ID,
			}
			if err := tx.Create(payment).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic payment")
			}
			split := &model.PaymentSplit{
				ClinicID:        clinicID,
				BillingID:       billing.ID,
				Method:          model.PaymentMethodCash,
				PaymentMethodID: &paymentMethod.ID,
				Amount:          syntheticClosingTotal,
				ReceivedAmount:  syntheticClosingTotal,
				PaidBy:          &staffRow.ID,
			}
			if err := tx.Create(split).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic payment split")
			}
			ids = append(ids, billing.ID)
		}

		result = &SyntheticClosingResult{
			ClinicID:     clinicID,
			LoginEmail:   SyntheticClosingLoginEmail(clinicID),
			BillingIDs:   ids,
			CompletedAt:  completed,
			CleanupToken: SyntheticClosingCleanupToken(clinicID),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// resolveSyntheticClosingCashPaymentMethod は clinic INSERT の
// trg_create_default_payment_methods が既に入れた cash 行を再利用する。
// testdb は trigger を載せないので、無いときだけ INSERT する。
func resolveSyntheticClosingCashPaymentMethod(tx *gorm.DB, clinicID uint64) (*model.PaymentMethodMaster, error) {
	var existing model.PaymentMethodMaster
	err := tx.Where("clinic_id = ? AND system_key = ?", clinicID, syntheticClosingCashSystemKey).
		Take(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, "load synthetic cash payment method")
	}

	cashKey := syntheticClosingCashSystemKey
	created := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         syntheticClosingCashDisplayName,
		SystemKey:    &cashKey,
		DisplayOrder: 1,
		IsActive:     true,
	}
	if err := tx.Create(created).Error; err != nil {
		var raced model.PaymentMethodMaster
		if loadErr := tx.Where("clinic_id = ? AND system_key = ?", clinicID, syntheticClosingCashSystemKey).
			Take(&raced).Error; loadErr == nil {
			return &raced, nil
		}
		return nil, apperrors.Wrap(err, "create synthetic payment method")
	}
	return created, nil
}

// SyntheticClosingAuditRowsPolicy は teardown トランザクション内で clinic スコープの
// audit_logs 行を解決する EMR-211 用の受け口。EMR-211 は option (b) に決裁済みで、
// 既定経路は SyntheticClosingAuditAnonymizePolicy — 監査行を削除せず sentinel に
// 付け替える — を束ねる。nil は監査行を温存したまま RESTRICT FK で teardown が
// 失敗する旧来の fail-closed 動作として、受け口検証用途に残る。
type SyntheticClosingAuditRowsPolicy func(ctx context.Context, tx *gorm.DB, clinicID uint64) error

// syntheticClosingAppendOnlyTables は append-only トリガー（BEFORE UPDATE OR DELETE →
// RAISE EXCEPTION）で通常経路の削除を拒否するテーブル。teardown は sanctioned
// exception として、トランザクション内で DISABLE TRIGGER USER → DELETE → ENABLE
// TRIGGER USER を行う。DDL はトランザクション対象なので、失敗時の rollback で
// トリガー状態も元に戻る。
// 一覧は migrations/001_init.sql の prevent_*_mutation トリガー群と
// internal/lintscan の teardown lint で突合される。
var syntheticClosingAppendOnlyTables = []string{
	"cash_register_close_adjustments",
	"cash_register_closes",
	"examination_revision_items",
	"examination_revisions",
	"lab_import_exam_retraction_items",
	"lab_import_exam_retractions",
	"lab_import_revert_receipts",
	"lab_import_usage_receipts",
}

// syntheticClosingDeleteStatements は fixture 生成物の全子孫を、001_init.sql の
// ブロッキング FK（ON DELETE RESTRICT / NO ACTION / 省略）を子→親の順で網羅する
// 削除系列。exams→examination_revisions の RESTRICT サイクルだけは先に
// current_revision_version を NULL へ戻して切る（exams 自体は append-only ではない）。
// CASCADE / SET NULL の子（estimate_items・exam_results・staff_notes 等）は親削除で
// 自動処理されるため対象外。audit_logs は EMR-211 (b) で削除せず sentinel へ付け替える
// ため系列に含めない（synthetic_closing_audit.go）。
// この系列は internal/lintscan の teardown lint が 001_init.sql から導出した
// RESTRICT 閉包と突合する — 新規テーブル追加時はここにも追加が必要。
var syntheticClosingDeleteStatements = []string{
	"DELETE FROM appointment_trimming_details WHERE clinic_id = ?",
	"DELETE FROM appointment_trimming_options WHERE clinic_id = ?",
	"DELETE FROM billing_items WHERE clinic_id = ?",
	"DELETE FROM billing_refunds WHERE clinic_id = ?",
	"DELETE FROM campaigns WHERE clinic_id = ?",
	"DELETE FROM care_logs WHERE clinic_id = ?",
	"DELETE FROM cash_register_close_adjustments WHERE clinic_id = ?",
	"DELETE FROM checkup_field_results WHERE clinic_id = ?",
	"DELETE FROM checkup_package_import_receipts WHERE clinic_id = ?",
	"DELETE FROM checkups WHERE clinic_id = ?",
	"DELETE FROM chief_complaint_types WHERE clinic_id = ?",
	"DELETE FROM clinic_holidays WHERE clinic_id = ?",
	"DELETE FROM clinic_integrations WHERE clinic_id = ?",
	"DELETE FROM clinic_settings WHERE clinic_id = ?",
	"DELETE FROM daily_records WHERE clinic_id = ?",
	"DELETE FROM diagnosis_names WHERE clinic_id = ?",
	"DELETE FROM diagnosis_types WHERE clinic_id = ?",
	"DELETE FROM estimates WHERE clinic_id = ?",
	"DELETE FROM exam_reference_ranges WHERE clinic_id = ?",
	"DELETE FROM examination_revision_items WHERE clinic_id = ?",
	"DELETE FROM hospitalization_plans WHERE clinic_id = ?",
	"DELETE FROM hospitalizations WHERE clinic_id = ?",
	"DELETE FROM inquiry_templates WHERE clinic_id = ?",
	"DELETE FROM insurances WHERE clinic_id = ?",
	"DELETE FROM lab_device_item_masters WHERE clinic_id = ?",
	"DELETE FROM lab_device_station_settings WHERE clinic_id = ?",
	"DELETE FROM lab_device_waits WHERE clinic_id = ?",
	"DELETE FROM lab_devices WHERE clinic_id = ?",
	"DELETE FROM lab_import_events WHERE clinic_id = ?",
	"DELETE FROM lab_import_exam_retraction_items WHERE clinic_id = ?",
	"DELETE FROM lab_import_job_items WHERE clinic_id = ?",
	"DELETE FROM lab_import_revert_receipts WHERE clinic_id = ?",
	"DELETE FROM lab_import_usage_receipts WHERE clinic_id = ?",
	"DELETE FROM line_link_tokens WHERE clinic_id = ?",
	"DELETE FROM line_reservation_settings WHERE clinic_id = ?",
	"DELETE FROM line_send_logs WHERE clinic_id = ?",
	"DELETE FROM lstep_delivery_trigger_log WHERE clinic_id = ?",
	"DELETE FROM lstep_friend_attribute_snapshots WHERE clinic_id = ?",
	"DELETE FROM lstep_migration_progress WHERE clinic_id = ?",
	"DELETE FROM lstep_settings WHERE clinic_id = ?",
	"DELETE FROM lstep_sync_error_counters WHERE clinic_id = ?",
	"DELETE FROM lstep_tag_cache WHERE clinic_id = ?",
	"DELETE FROM lstep_tag_code_mappings WHERE clinic_id = ?",
	"DELETE FROM lstep_trigger_priorities WHERE clinic_id = ?",
	"DELETE FROM medical_record_addenda WHERE clinic_id = ?",
	"DELETE FROM medical_record_image_upload_quota WHERE clinic_id = ?",
	"DELETE FROM merchandise_items WHERE clinic_id = ?",
	"DELETE FROM occupations WHERE clinic_id = ?",
	"DELETE FROM owner_identity_group_members WHERE clinic_id = ?",
	"DELETE FROM payment_splits WHERE clinic_id = ?",
	"DELETE FROM payments WHERE clinic_id = ?",
	"DELETE FROM permission_groups WHERE clinic_id = ?",
	"DELETE FROM pet_chronic_conditions WHERE clinic_id = ?",
	"DELETE FROM pet_identity_group_members WHERE clinic_id = ?",
	"DELETE FROM pet_owners WHERE clinic_id = ?",
	"DELETE FROM prescriptions WHERE clinic_id = ?",
	"DELETE FROM reservation_type_available_slots WHERE clinic_id = ?",
	"DELETE FROM reservation_type_groups WHERE clinic_id = ?",
	"DELETE FROM reservation_type_occupations WHERE clinic_id = ?",
	"DELETE FROM reservation_type_unavailable_times WHERE clinic_id = ?",
	"DELETE FROM shared_files WHERE clinic_id = ?",
	"DELETE FROM shift_templates WHERE clinic_id = ?",
	"DELETE FROM staff_clinic_assignments WHERE clinic_id = ?",
	"DELETE FROM staff_reservation_capabilities WHERE clinic_id = ?",
	"DELETE FROM treatment_plans WHERE clinic_id = ?",
	"DELETE FROM treatments WHERE clinic_id = ?",
	"DELETE FROM trimming_courses WHERE clinic_id = ?",
	"DELETE FROM vital_records WHERE clinic_id = ?",
	"DELETE FROM trimming_options WHERE clinic_id = ?",
	"DELETE FROM vaccinations WHERE clinic_id = ?",
	"DELETE FROM cash_register_closes WHERE clinic_id = ?",
	"DELETE FROM checkup_type_fields WHERE clinic_id = ?",
	"DELETE FROM examination_revisions WHERE clinic_id = ?",
	"DELETE FROM cages WHERE clinic_id = ?",
	"DELETE FROM lab_import_exam_retractions WHERE clinic_id = ?",
	"DELETE FROM exam_type_fields WHERE clinic_id = ?",
	"DELETE FROM lstep_csv_imports WHERE clinic_id = ?",
	"DELETE FROM payment_methods WHERE clinic_id = ?",
	"DELETE FROM billings WHERE clinic_id = ?",
	"DELETE FROM pet_identity_groups WHERE created_clinic_id = ?",
	"DELETE FROM medicine_dose_params WHERE clinic_id = ?",
	"DELETE FROM medicines WHERE clinic_id = ?",
	"DELETE FROM inventory_items WHERE clinic_id = ?",
	"DELETE FROM procedures WHERE clinic_id = ?",
	"DELETE FROM consultations WHERE clinic_id = ?",
	"DELETE FROM vaccines WHERE clinic_id = ?",
	"DELETE FROM medical_records WHERE clinic_id = ?",
	"DELETE FROM checkup_types WHERE clinic_id = ?",
	"DELETE FROM exams WHERE clinic_id = ?",
	"DELETE FROM owner_identity_groups WHERE created_clinic_id = ?",
	"DELETE FROM lab_import_jobs WHERE clinic_id = ?",
	"DELETE FROM exam_types WHERE clinic_id = ?",
	"DELETE FROM line_customers WHERE clinic_id = ?",
	"DELETE FROM pets WHERE clinic_id = ?",
	"DELETE FROM owners WHERE clinic_id = ?",
}

// syntheticClosingDeleteTargetRe は syntheticClosingDeleteStatements 各文の対象
// テーブルとスコープ列を取り出す。lint 側（internal/lintscan）はテーブル抽出のみでよく、
// 同系列を別正規表現で抽出する。
var syntheticClosingDeleteTargetRe = regexp.MustCompile(`^DELETE FROM (\w+) WHERE (\w+) = \?$`)

// DeleteSyntheticClosingFixture は合成 clinic とその子孫だけを消す。clinic 1/2 と接頭辞不一致は拒否する。
// audit_logs は EMR-211 (b) の既定 policy で削除せず sentinel clinic/staff へ
// 付け替える（synthetic_closing_audit.go）ため、監査行を持つ合成 clinic の
// teardown も監査証跡を保ったまま完遂する。
func DeleteSyntheticClosingFixture(ctx context.Context, db *gorm.DB, appEnv, dbHost string, clinicID uint64, cleanupToken string) error {
	return DeleteSyntheticClosingFixtureWithAuditPolicy(ctx, db, appEnv, dbHost, clinicID, cleanupToken, SyntheticClosingAuditAnonymizePolicy)
}

// DeleteSyntheticClosingFixtureWithAuditPolicy は DeleteSyntheticClosingFixture と同じだが、
// auditRowsPolicy が非 nil のとき teardown トランザクション内でそれを呼び、
// clinic スコープの audit_logs 行の扱いを委譲する（EMR-211 の受け口）。
func DeleteSyntheticClosingFixtureWithAuditPolicy(ctx context.Context, db *gorm.DB, appEnv, dbHost string, clinicID uint64, cleanupToken string, auditRowsPolicy SyntheticClosingAuditRowsPolicy) error {
	if err := AllowUATSyntheticClosing(appEnv, dbHost); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectReservedClinicID(clinicID); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if !MatchSyntheticClosingCleanupToken(clinicID, cleanupToken) {
		return apperrors.WrapInvalidInput("cleanup token is invalid")
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}

	// create と同じ契約: ambient tx があれば SAVEPOINT join し、委譲 cleanup が
	// 呼び出し側 ambient tx ではなくこの teardown tx へ解決されるよう
	// tx 束縛 ctx を渡す。裸 ctx のまま委譲すると DBOrTx が ambient tx を
	// 選び、teardown tx の外へ commit される残存 bug になる。
	base := persistence.DBOrTx(ctx, db)
	return base.Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		var clinic model.Clinic
		if err := tx.First(&clinic, clinicID).Error; err != nil {
			return apperrors.Wrap(err, "load synthetic clinic")
		}
		if !strings.HasPrefix(clinic.Name, syntheticClosingClinicPrefix) {
			return apperrors.WrapInvalidInput("clinic name is not a synthetic closing fixture")
		}

		var staffs []model.Staff
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&staffs).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic staff")
		}
		accountIDs := make([]uint64, 0, len(staffs))
		for _, staff := range staffs {
			if staff.AccountID != nil {
				accountIDs = append(accountIDs, *staff.AccountID)
			}
		}

		existingColumns, err := syntheticClosingExistingColumns(tx)
		if err != nil {
			return err
		}

		disabledTriggers := make([]string, 0, len(syntheticClosingAppendOnlyTables))
		for _, table := range syntheticClosingAppendOnlyTables {
			if existingColumns[table] == nil {
				continue
			}
			if err := tx.Exec(fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER USER", table)).Error; err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("disable append-only triggers on %s", table))
			}
			disabledTriggers = append(disabledTriggers, table)
		}

		if existingColumns["exams"]["current_revision_version"] {
			// exams ↔ examination_revisions は双方 RESTRICT NOT DEFERRABLE のため、
			// current_revision_version を NULL に戻して先にサイクルを切る。
			if err := tx.Exec("UPDATE exams SET current_revision_version = NULL WHERE clinic_id = ?", clinicID).Error; err != nil {
				return apperrors.Wrap(err, "clear synthetic exam revision pointers")
			}
		}

		for _, stmt := range syntheticClosingDeleteStatements {
			target := syntheticClosingDeleteTargetRe.FindStringSubmatch(stmt)
			if len(target) < 3 {
				return apperrors.Wrap(fmt.Errorf("unparsable delete statement %q", stmt), "teardown plan")
			}
			if !existingColumns[target[1]][target[2]] {
				continue
			}
			if err := tx.Exec(stmt, clinicID).Error; err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("delete synthetic %s", target[1]))
			}
		}

		if auditRowsPolicy != nil {
			if err := auditRowsPolicy(txCtx, tx, clinicID); err != nil {
				return apperrors.Wrap(err, "resolve synthetic audit_logs")
			}
		}

		if err := reservation.UnscopedDeleteSyntheticClosingReservations(txCtx, tx, clinicID); err != nil {
			return apperrors.Wrap(err, "delete synthetic reservations")
		}
		if err := staff.UnscopedDeleteSyntheticClosingStaffs(txCtx, tx, clinicID); err != nil {
			return apperrors.Wrap(err, "delete synthetic staff")
		}
		if len(accountIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", accountIDs).Delete(&model.Account{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic accounts")
			}
		}
		if err := tx.Unscoped().Where("name LIKE ?", fmt.Sprintf("s09-species-%d", clinicID)).Delete(&model.AnimalSpecies{}).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic species")
		}
		if err := tx.Delete(&clinic).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic clinic")
		}
		if clinic.CompanyID != 0 {
			if err := tx.Where("id = ? AND name LIKE ?", clinic.CompanyID, syntheticClosingCompanyPrefix+"%").Delete(&model.Company{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic company")
			}
		}

		for _, table := range disabledTriggers {
			if err := tx.Exec(fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER USER", table)).Error; err != nil {
				return apperrors.Wrap(err, fmt.Sprintf("re-enable append-only triggers on %s", table))
			}
		}
		return nil
	})
}

// syntheticClosingExistingColumns は現在スキーマの テーブル→列 集合を返す。
// testdb（AutoMigrate の GORM double）は migration 全テーブル・全列を持たず、
// migration 側で後付けされた列（例: treatments.clinic_id）も欠けるため、
// 存在しない対象への DELETE/ALTER/UPDATE はスキップする。migrate 済み実 DB では
// lint 側が系列の網羅性を 001_init.sql と突合する。
func syntheticClosingExistingColumns(tx *gorm.DB) (map[string]map[string]bool, error) {
	type row struct {
		TableName  string
		ColumnName string
	}
	var rows []row
	if err := tx.Raw(
		`SELECT table_name, column_name FROM information_schema.columns
		WHERE table_schema = current_schema()`,
	).Scan(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "list existing columns for synthetic teardown")
	}
	existing := make(map[string]map[string]bool, len(rows))
	for _, r := range rows {
		if existing[r.TableName] == nil {
			existing[r.TableName] = make(map[string]bool)
		}
		existing[r.TableName][r.ColumnName] = true
	}
	return existing, nil
}
