// Package repository provides data access implementations for BillingItem entity.
package billing

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// BillingItemRepository は billing_items テーブルの CRUD を担うインターフェース
type BillingItemRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	ValidateCreateReferences(ctx context.Context, clinicID, billingID uint64, merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64) (model.ItemCategory, error)
	ValidateVaccinationCreateReference(ctx context.Context, clinicID, billingID, vaccinationID uint64) (*vaccinationBillingValues, error)
	ValidateExamCreateReference(ctx context.Context, clinicID, billingID, examID uint64) error
	LockActiveStaffAssignment(ctx context.Context, clinicID, staffID uint64) error
	// FindUnbilledVaccinationItemsByPetID は未請求 vaccination 候補を返す。
	// unbillableCount は vaccine master 欠損/負価格などで除外した件数（BUG-013 warning 用）。
	// 除外行は error にせず skip する（infra error のみ error）。
	FindUnbilledVaccinationItemsByPetID(ctx context.Context, clinicID, petID uint64) (items []model.BillingItem, unbillableCount int, err error)
	FindUnbilledExamItemsByPetID(ctx context.Context, clinicID, petID uint64) (items []model.BillingItem, unbillableCount int, err error)
	Create(ctx context.Context, item *model.BillingItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	// UpdateBillingTotalsForCompletedCorrection は確定済み会計の明細訂正時のみ totals 再計算を許可する（BUG-009）。
	// cancelled は引き続き拒否。通常経路は UpdateBillingTotals を使う。
	UpdateBillingTotalsForCompletedCorrection(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	// HasItemByOwnerSince は指定飼い主の請求アイテムに names いずれかが存在するか返す（FEAT-379）。
	HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// HasFoodPurchaseByOwnerSince は names 指定時は名前で、未指定時は category=food で判定する（FEAT-379）。
	HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// FindUnbilledTrimmingItemsByPetID は指定ペットの未請求トリミングコース/オプションを返す(#77)。
	FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	// CountNonAccountingTrimmingByPetAndDate は同日同ペットの「未会計対象化」トリミング appointment 件数を返す(#77)。
	CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

type activeStaffAssignmentLocker interface {
	LockActiveStaffAssignment(ctx context.Context, clinicID, staffID uint64) error
}

type billingItemRepository struct {
	db *gorm.DB
	activeStaffAssignmentLocker
}

// NewBillingItemRepository は BillingItemRepository を初期化して返す
func NewBillingItemRepository(db *gorm.DB) BillingItemRepository {
	return &billingItemRepository{
		db:                          db,
		activeStaffAssignmentLocker: NewBillingConfirmationRepository(db),
	}
}

// BE-refactor.md R1-1 follow-up (go-reviewer指摘・D2と同型): billing_item_service の
// CreateItem/UpdateItem/DeleteItem は Create/Update/Delete + recalculateTotals
// （FindByBillingID + UpdateBillingTotals）を WithTx 内で txCtx 付きで呼ぶ。
// dbOrTx 未参加のままだと SavePaymentSplits と同じ部分コミットが起こりうる
// （明細書込は独立 tx で即コミット、直後の合計再計算のみ失敗すると ambient tx の rollback で
// 明細だけ残り billing.subtotal/tax_total/total_amount と不整合になる）。
// FindByID も Update 後の再読込（txCtx 付き）で呼ばれるため対象に含める。
func (r *billingItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error) {
	var item model.BillingItem
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *billingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
	items := make([]model.BillingItem, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.billing_id = ?", billingID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", "")
	}
	return items, nil
}

type billingItemBillingReference struct {
	MedicalRecordID *uint64
	OwnerID         *uint64
	PetID           *uint64
	Status          model.BillingStatus
}

type billingItemMedicalRecordReference struct {
	ID            uint64
	AppointmentID *uint64
	OwnerID       *uint64
	PetID         *uint64
}

type billingItemAppointmentReference struct {
	OwnerID *uint64
	PetID   *uint64
}

type billingItemMerchandiseReference struct {
	ID       uint64
	Category model.ItemCategory
}

type vaccinationBillingValues struct {
	Name      string
	UnitPrice int64
}

func sameOptionalBillingReference(left, right *uint64) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func invalidBillingItemReferenceCombination() error {
	return apperrors.WrapInvalidInput("参照先の組み合わせが正しくありません")
}

// ValidateCreateReferences validates every request-derived billing_items FK
// against the authenticated clinic and its parent graph. It must run in the
// same transaction as Create so that row locks keep the validated graph stable
// until persistence commits. The billing parent is locked FOR UPDATE because
// the same transaction recalculates and updates its totals.
func (r *billingItemRepository) ValidateCreateReferences(
	ctx context.Context,
	clinicID, billingID uint64,
	merchandiseItemID, treatmentID, appointmentID, trimmingCourseID, trimmingOptionID *uint64,
) (model.ItemCategory, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return "", apperrors.WrapInternalServerError("billing item reference validation requires an active transaction")
	}
	tx = tx.WithContext(ctx)

	var billingRef billingItemBillingReference
	if err := tx.
		Table("billings").
		Select("medical_record_id", "owner_id", "pet_id", "status").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", billingID, clinicID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&billingRef).Error; err != nil {
		return "", apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	// BUG-463: defense in depth — finalized billings must not accept new items.
	if billingRef.Status == model.BillingStatusCompleted ||
		billingRef.Status == model.BillingStatusCancelled {
		return "", apperrors.WrapConflict("確定済みまたは取消済みの会計明細は登録できません")
	}

	var merchandiseRef billingItemMerchandiseReference
	if merchandiseItemID != nil {
		if err := tx.
			Table("merchandise_items").
			Select("id", "category").
			Where("id = ? AND clinic_id = ? AND is_active = TRUE AND deleted_at IS NULL", *merchandiseItemID, clinicID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&merchandiseRef).Error; err != nil {
			return "", apperrors.FromGORM(err, "merchandise_item", fmt.Sprintf("%d", *merchandiseItemID))
		}
	}

	var medicalRecordRef *billingItemMedicalRecordReference
	if billingRef.MedicalRecordID != nil && (treatmentID != nil || appointmentID != nil) {
		var ref billingItemMedicalRecordReference
		if err := tx.
			Table("medical_records").
			Select("id", "appointment_id", "owner_id", "pet_id").
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *billingRef.MedicalRecordID, clinicID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&ref).Error; err != nil {
			return "", apperrors.FromGORM(err, "medical_record", fmt.Sprintf("%d", *billingRef.MedicalRecordID))
		}
		if !sameOptionalBillingReference(billingRef.OwnerID, ref.OwnerID) ||
			!sameOptionalBillingReference(billingRef.PetID, ref.PetID) {
			return "", invalidBillingItemReferenceCombination()
		}
		medicalRecordRef = &ref
	}

	if treatmentID != nil {
		var treatmentRef struct {
			MedicalRecordID uint64
		}
		if err := tx.
			Table("treatments").
			Select("treatments.medical_record_id").
			Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
			Where("treatments.id = ? AND treatments.deleted_at IS NULL", *treatmentID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&treatmentRef).Error; err != nil {
			return "", apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", *treatmentID))
		}
		if billingRef.MedicalRecordID == nil ||
			treatmentRef.MedicalRecordID != *billingRef.MedicalRecordID {
			return "", invalidBillingItemReferenceCombination()
		}
	}

	// BUG-506: unbilled → complete clients may omit appointment_id while still
	// sending trimming_course_id / trimming_option_id. Resolve the unique
	// accounting-status trimming appointment for the billing pet (fail-closed
	// when zero or ambiguous). Keep invalid combinations rejected.
	effectiveAppointmentID := appointmentID
	if effectiveAppointmentID == nil && (trimmingCourseID != nil || trimmingOptionID != nil) {
		resolved, err := resolveUniqueTrimmingAppointmentID(
			tx, clinicID, billingRef.PetID, trimmingCourseID, trimmingOptionID,
		)
		if err != nil {
			return "", err
		}
		effectiveAppointmentID = resolved
	}

	if effectiveAppointmentID != nil {
		var appointmentRef billingItemAppointmentReference
		if err := tx.
			Table("appointments").
			Select("owner_id", "pet_id").
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", *effectiveAppointmentID, clinicID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&appointmentRef).Error; err != nil {
			return "", apperrors.FromGORM(err, "appointment", fmt.Sprintf("%d", *effectiveAppointmentID))
		}
		if !sameOptionalBillingReference(billingRef.OwnerID, appointmentRef.OwnerID) ||
			!sameOptionalBillingReference(billingRef.PetID, appointmentRef.PetID) {
			return "", invalidBillingItemReferenceCombination()
		}
		// S11: trimming appointment may differ from billing medical_record appointment.
		enforceMedicalAppointment := (medicalRecordRef != nil && treatmentID != nil) ||
			(medicalRecordRef != nil && trimmingCourseID == nil && trimmingOptionID == nil)
		if enforceMedicalAppointment &&
			(medicalRecordRef.AppointmentID == nil || *medicalRecordRef.AppointmentID != *effectiveAppointmentID) {
			return "", invalidBillingItemReferenceCombination()
		}
	}

	if (trimmingCourseID != nil || trimmingOptionID != nil) && effectiveAppointmentID == nil {
		return "", invalidBillingItemReferenceCombination()
	}
	if trimmingCourseID != nil {
		var id uint64
		// Parent appointments clinic correlation (SEC-SWEEP-02-BILL-B1a): child clinic
		// alone is insufficient when appointment_id is a corrupt cross-tenant FK.
		// No appointments.deleted_at — matches TRIM-B1 / MR-B1 appointments-parent pattern.
		// Use unaliased table names so AST lint sees appointments.id=appointment_trimming_details.appointment_id.
		if err := tx.
			Table("appointment_trimming_details").
			Select("trimming_courses.id").
			Joins("JOIN appointments ON appointments.id = appointment_trimming_details.appointment_id AND appointments.clinic_id = appointment_trimming_details.clinic_id").
			Joins("JOIN trimming_courses ON trimming_courses.id = appointment_trimming_details.course_id AND trimming_courses.clinic_id = appointment_trimming_details.clinic_id AND trimming_courses.deleted_at IS NULL").
			Where("appointment_trimming_details.appointment_id = ? AND appointment_trimming_details.clinic_id = ? AND appointment_trimming_details.course_id = ?", *effectiveAppointmentID, clinicID, *trimmingCourseID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&id).Error; err != nil {
			return "", apperrors.FromGORM(err, "trimming_course", fmt.Sprintf("%d", *trimmingCourseID))
		}
	}
	if trimmingOptionID != nil {
		var id uint64
		if err := tx.
			Table("appointment_trimming_options AS ato").
			Select("topt.id").
			Joins("JOIN appointments AS a ON a.id = ato.appointment_id AND a.clinic_id = ? AND a.deleted_at IS NULL", clinicID).
			Joins("JOIN trimming_options AS topt ON topt.id = ato.option_id AND topt.clinic_id = a.clinic_id AND topt.deleted_at IS NULL").
			Where("ato.appointment_id = ? AND ato.option_id = ?", *effectiveAppointmentID, *trimmingOptionID).
			Clauses(clause.Locking{Strength: "SHARE"}).
			Take(&id).Error; err != nil {
			return "", apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", *trimmingOptionID))
		}
	}

	return merchandiseRef.Category, nil
}

// resolveUniqueTrimmingAppointmentID finds the single accounting-status trimming
// appointment for pet that carries the given course and/or option. Zero or many
// matches fail closed (InvalidInput). Requires ambient tx.
func resolveUniqueTrimmingAppointmentID(
	tx *gorm.DB,
	clinicID uint64,
	petID *uint64,
	trimmingCourseID, trimmingOptionID *uint64,
) (*uint64, error) {
	if petID == nil {
		return nil, invalidBillingItemReferenceCombination()
	}
	if trimmingCourseID == nil && trimmingOptionID == nil {
		return nil, invalidBillingItemReferenceCombination()
	}

	appointmentQuery := tx.
		Table("appointments AS a").
		Select("a.id").
		Joins("JOIN reservation_types AS rt ON rt.id = a.reservation_type_id AND rt.clinic_id = a.clinic_id AND rt.deleted_at IS NULL").
		Where(
			"a.clinic_id = ? AND a.pet_id = ? AND a.deleted_at IS NULL AND a.status = ? AND rt.category = ?",
			clinicID, *petID, model.ReservationStatusAccounting, model.ReservationTypeCategoryTrimming,
		)
	if trimmingCourseID != nil {
		appointmentQuery = appointmentQuery.Joins(
			"JOIN appointment_trimming_details AS atd ON atd.appointment_id = a.id AND atd.clinic_id = a.clinic_id AND atd.course_id = ?",
			*trimmingCourseID,
		)
	}
	if trimmingOptionID != nil {
		appointmentQuery = appointmentQuery.Joins(
			"JOIN appointment_trimming_options AS ato ON ato.appointment_id = a.id AND ato.option_id = ?",
			*trimmingOptionID,
		)
	}

	var ids []uint64
	if err := appointmentQuery.Limit(2).Pluck("a.id", &ids).Error; err != nil {
		return nil, apperrors.FromGORM(err, "appointment", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, *petID))
	}
	if len(ids) != 1 {
		return nil, invalidBillingItemReferenceCombination()
	}
	id := ids[0]
	return &id, nil
}

func (r *billingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "billing_item", "")
	}
	return nil
}

// Update は clinic 述語を EXISTS subquery で強制する（BUG-417: GORM の Joins() は
// UPDATE/DELETE SQL へ伝播せず実質 no-op だった——service 層の事前 FindByID gate に
// 依存しない repository 層 defense-in-depth を回復）。
func (r *billingItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingItem{}).
		Where("billing_items.id = ?", id).
		Where("EXISTS (SELECT 1 FROM billings WHERE billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL)", clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete は clinic 述語を EXISTS subquery で強制し、soft-delete と同じ原子的更新で
// vaccination provenanceを解放する。これにより削除した接種イベントを再度取り込める。
func (r *billingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingItem{}).
		Where("billing_items.id = ?", id).
		Where("EXISTS (SELECT 1 FROM billings WHERE billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL)", clinicID).
		Updates(map[string]any{
			"vaccination_id": nil,
			"exam_id":        nil,
			"clinic_id":      nil,
			"deleted_at":     time.Now(),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	// BUG-463: exclude completed/cancelled so totals cannot be rewritten after finalization
	// even if a caller bypasses service-level status guards.
	return r.updateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount, false)
}

// UpdateBillingTotalsForCompletedCorrection は BUG-009: 理由付き確定済み明細訂正の totals 再計算専用。
func (r *billingItemRepository) UpdateBillingTotalsForCompletedCorrection(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	return r.updateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount, true)
}

func (r *billingItemRepository) updateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64, allowCompleted bool) error {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", billingID)
	if allowCompleted {
		// cancelled のみ除外（completed は訂正経路で許可）
		q = q.Where("status <> ?", model.BillingStatusCancelled)
	} else {
		q = q.Where("status NOT IN (?, ?)", model.BillingStatusCompleted, model.BillingStatusCancelled)
	}
	result := q.Updates(map[string]any{
		"subtotal":     subtotal,
		"tax_total":    taxTotal,
		"total_amount": totalAmount,
	})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		// Distinguish finalized (Conflict) from missing/out-of-scope (NotFound).
		var row struct {
			Status model.BillingStatus
		}
		err := persistence.DBOrTx(ctx, r.db).
			Model(&model.Billing{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Select("status").
			Where("id = ?", billingID).
			Take(&row).Error
		if err != nil {
			return apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
		}
		if row.Status == model.BillingStatusCancelled || (!allowCompleted && row.Status == model.BillingStatusCompleted) {
			return apperrors.WrapConflict("確定済みまたは取消済みの会計合計は更新できません")
		}
		return apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	return nil
}

func (r *billingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.completed_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.name IN ? AND billing_items.deleted_at IS NULL", names).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}

func (r *billingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.completed_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.deleted_at IS NULL")
	if len(names) > 0 {
		q = q.Where("billing_items.name IN ?", names)
	} else {
		q = q.Where("billing_items.category = ?", string(model.ItemCategoryFood))
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}
