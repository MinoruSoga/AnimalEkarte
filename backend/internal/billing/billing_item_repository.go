// Package billing provides billing item persistence.
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
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateBillingItemInput) error
	Delete(ctx context.Context, clinicID, id uint64) error
	// DeleteAllByBillingID は EMR-253: takeover 対象 billing の既存明細を一括 soft-delete する。
	// 個別 Delete と同じ原子的更新（vaccination/exam provenance + denormalized clinic_id を解放）。
	// 0 件（明細なし）は正常系として nil を返す。
	DeleteAllByBillingID(ctx context.Context, clinicID, billingID uint64) error
	// UpdateBillingTotals は廃止列への書込みだったため、何も書かず固定エラーを返す（held for EMR-506。呼出しとともに削除する）。
	// medicalrecord の退院会計（hospitalization_discharge_tx.go）が interface 経由でまだ呼ぶため宣言だけ残している。
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
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
	ClaimState      *string
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
// until persistence commits. The billing parent is locked FOR UPDATE so that the
// claim_state checked here cannot change before the item is inserted.
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
		Select("medical_record_id", "owner_id", "pet_id", "claim_state").
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", billingID, clinicID).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&billingRef).Error; err != nil {
		return "", apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	// BUG-463: defense in depth — only draft billings accept new items. A NULL or unknown claim_state
	// (001_init.sql:9454) is rejected as well, with the same polarity as the service guard rejectIfBillingFinalized.
	if err := rejectIfBillingFinalized(&model.Billing{ClaimState: billingRef.ClaimState}, "登録"); err != nil {
		return "", err
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
	if err := validateTrimmingBillingItemRefs(
		tx, clinicID, treatmentID, trimmingCourseID, trimmingOptionID,
		effectiveAppointmentID, billingRef, medicalRecordRef,
	); err != nil {
		return "", err
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
func (r *billingItemRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateBillingItemInput) error {
	return r.update(ctx, clinicID, id, buildBillingItemUpdate(&cmd))
}

func (r *billingItemRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
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

// Delete は clinic 述語を EXISTS subquery で強制する soft-delete。更新するのは deleted_at だけ。
// 請求元の解放は billing_source_bindings 側の操作になり（明細側の vaccination_id・exam_id 列は 001_init.sql:10891 で廃止）、
// clinic_id は NOT NULL（同 6199）で、trigger が親請求から再設定する（同 6251）ため NULL 化してはならない。
func (r *billingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingItem{}).
		Where("billing_items.id = ?", id).
		Where("EXISTS (SELECT 1 FROM billings WHERE billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL)", clinicID).
		Updates(map[string]any{
			"deleted_at": time.Now(),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

// DeleteAllByBillingID は EMR-253: takeover 対象 billing の既存明細を一括 soft-delete する。
// 個別 Delete と同じく更新するのは deleted_at だけ（廃止列と clinic_id の扱いは Delete を参照）。
// EXISTS 副問合せで親 billings の clinic scope を強制する。0 件（明細なし）は正常系。
func (r *billingItemRepository) DeleteAllByBillingID(ctx context.Context, clinicID, billingID uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.BillingItem{}).
		Where("billing_items.billing_id = ?", billingID).
		Where("billing_items.deleted_at IS NULL").
		Where("EXISTS (SELECT 1 FROM billings WHERE billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL)", clinicID).
		Updates(map[string]any{
			"deleted_at": time.Now(),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("billing_id=%d", billingID))
	}
	return nil
}

// UpdateBillingTotals は廃止列（billings.subtotal・tax_total・total_amount。001_init.sql:10889-10891）への書込みだったため、
// 何も書かず固定エラーを返す。合計は算定器（CalculateBilling）の結果として算定固定（snapshot）に保存する設計で、その writer は
// 未接続。medicalrecord の退院会計（hospitalization_discharge_tx.go）が interface 経由でまだ呼ぶため宣言だけ残す。held for EMR-506。
func (r *billingItemRepository) UpdateBillingTotals(_ context.Context, _, _ uint64, _, _, _ int64) error {
	return apperrors.WrapFinanceConsumerPending("会計ヘッダの合計の保存は、算定固定（snapshot）の writer が未接続のため利用できません")
}

// HasItemByOwnerSince は FEAT-379 の LSTEP タグ判定用で、billings.completed_at（廃止列）を基準にしていた。
// 置換する基準日（確定イベントの日時か、受領日か）は PO 判断待ち（Q9）で、DDL に決定済みの置換列はない。
// 推測せず、false も返さず、固定エラーで拒否する（呼出しの LSTEP タグ同期は error を伝播する）。held for EMR-506/509。
func (r *billingItemRepository) HasItemByOwnerSince(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
	return false, apperrors.WrapFinanceConsumerPending("請求履歴に基づくタグ判定は、請求確定日の基準（確定イベント・受領日）が未決定のため利用できません")
}

// HasFoodPurchaseByOwnerSince は HasItemByOwnerSince と同じ理由で保留する（names 未指定時の category=food 判定を含む）。
func (r *billingItemRepository) HasFoodPurchaseByOwnerSince(_ context.Context, _, _ uint64, _ time.Time, _ []string) (bool, error) {
	return false, apperrors.WrapFinanceConsumerPending("請求履歴に基づくタグ判定は、請求確定日の基準（確定イベント・受領日）が未決定のため利用できません")
}
