package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
)

// applyBillingOwnerPetSearch scopes billings by owner/pet display names with
// katakana→hiragana symmetry (same translate/NormalizeKana contract as owners list).

// applyBillingPaymentMethodFilter scopes by payments / payment_splits method.
// Op: is (default), is_not, is_empty, is_not_empty. Empty method with is/is_not is a no-op.
func applyBillingPaymentMethodFilter(q *gorm.DB, method *string, op string) *gorm.DB {
	op = strings.ToLower(strings.TrimSpace(op))
	if op == "" {
		op = "is"
	}
	hasPayment := `EXISTS (
		SELECT 1 FROM payments p
		WHERE p.billing_id = billings.id
		  AND p.clinic_id = billings.clinic_id
		  AND p.deleted_at IS NULL
	)`
	hasSplit := `EXISTS (
		SELECT 1 FROM payment_splits ps
		WHERE ps.billing_id = billings.id
		  AND ps.clinic_id = billings.clinic_id
	)`
	switch op {
	case "is_empty":
		return q.Where("NOT (" + hasPayment + " OR " + hasSplit + ")")
	case "is_not_empty":
		return q.Where("(" + hasPayment + " OR " + hasSplit + ")")
	}
	if method == nil || strings.TrimSpace(*method) == "" {
		return q
	}
	m := strings.TrimSpace(*method)
	methodMatch := `(
		EXISTS (
			SELECT 1 FROM payments p
			WHERE p.billing_id = billings.id
			  AND p.clinic_id = billings.clinic_id
			  AND p.deleted_at IS NULL
			  AND p.method = ?
		)
		OR EXISTS (
			SELECT 1 FROM payment_splits ps
			WHERE ps.billing_id = billings.id
			  AND ps.clinic_id = billings.clinic_id
			  AND ps.method = ?
		)
	)`
	if op == "is_not" {
		return q.Where("NOT "+methodMatch, m, m)
	}
	return q.Where(methodMatch, m, m)
}

func applyBillingOwnerPetSearch(q *gorm.DB, search string) *gorm.DB {
	rawPattern := "%" + textsearch.EscapeLike(search) + "%"
	normalizedPattern := "%" + textsearch.EscapeLike(textsearch.NormalizeKana(search)) + "%"
	return q.Where(
		`(
			EXISTS (
				SELECT 1 FROM owners o
				WHERE o.id = billings.owner_id
				  AND o.clinic_id = billings.clinic_id
				  AND o.deleted_at IS NULL
				  AND (
				    o.name ILIKE ? ESCAPE '\'
				    OR translate(o.name, ?, ?) ILIKE ? ESCAPE '\'
				    OR translate(COALESCE(o.name_kana, ''), ?, ?) ILIKE ? ESCAPE '\'
				  )
			)
			OR EXISTS (
				SELECT 1 FROM pets p
				WHERE p.id = billings.pet_id
				  AND p.clinic_id = billings.clinic_id
				  AND p.deleted_at IS NULL
				  AND (
				    p.name ILIKE ? ESCAPE '\'
				    OR translate(p.name, ?, ?) ILIKE ? ESCAPE '\'
				    OR translate(COALESCE(p.name_kana, ''), ?, ?) ILIKE ? ESCAPE '\'
				  )
			)
		)`,
		rawPattern,
		textsearch.KanaSourceChars, textsearch.KanaTargetChars, normalizedPattern,
		textsearch.KanaSourceChars, textsearch.KanaTargetChars, normalizedPattern,
		rawPattern,
		textsearch.KanaSourceChars, textsearch.KanaTargetChars, normalizedPattern,
		textsearch.KanaSourceChars, textsearch.KanaTargetChars, normalizedPattern,
	)
}

// AccountingListFilters is the list query contract for FindAll / List.
// StatusOp: "is" (default) | "is_not".
// PaymentMethodOp: "is" (default) | "is_not" | "is_empty" | "is_not_empty".
type AccountingListFilters struct {
	PetID           *uint64
	OwnerID         *uint64
	Status          *string
	StatusOp        string
	StartDate       *string
	EndDate         *string
	Search          string
	PaymentMethod   *string
	PaymentMethodOp string
}

type AccountingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error)
	// FindAllForClinics は複数医院の会計を横断検索する (#86 段階3)。clinicIDs はハンドラ層で所属検証済みであること。
	FindAllForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	// FindByIDForClinics は複数医院スコープで会計を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error)
	// LockAndFindByID は FOR UPDATE で請求を行ロック取得する（TOCTOU 防止）。トランザクション内でのみ使用。
	LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error
	Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error)
	SavePayment(ctx context.Context, payment *model.Payment) error
	// SavePaymentSplits は billing の payment_splits を delete-then-recreate で保存する。
	SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error
	// BUG-370 / #120: 未納者一覧（startDate〜endDate の BETWEEN）
	FindUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error)
	FindUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error)
	// #182: 飼主単位の未納残高（会計画面表示用・status=waiting 合計）
	SumUnpaidByOwner(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error)
	// #114: 月次未納繰越集計（firstDay=YYYY-MM-01, lastDay=YYYY-MM-DD 月末）
	FindMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error)
	// BUG-368: レジ締め日次集計
	GetDailySummary(ctx context.Context, clinicID uint64, date time.Time) (*DailySummaryResult, error)
	// FEAT-368: 集計・締め機能
	GetCloseAggregate(ctx context.Context, input GetCloseAggregateInput) (*CloseAggregateResult, error)
	GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error)
	GetMonthlyReportByPeriod(ctx context.Context, clinicID uint64, start, end time.Time) (*MonthlyReportResult, error)
	// #247 DEC-16⑥: カテゴリ×支払 per-billing 配賦入力（締め・月次共通）
	GetCategoryPaymentAllocationData(ctx context.Context, clinicID uint64, periodStart, periodEnd time.Time) (*CategoryPaymentAllocationData, error)
	// SumPaidByOwner は飼い主の支払済み請求合計（LTV）を返す（Lステップタグ同期用）。
	SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// MaxSingleVisitAmountByOwner は飼い主の1回来院最大支払額を返す（CPMスポット判定用）。
	MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// FindOwnersByAnnualRevenue は直近365日の完了済み請求額合計を飼い主ごとに集計し、降順で返す（LTV上位％判定用）。
	FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]OwnerAnnualRevenue, error)
	// FindByCompletionRequestID は BUG-018 冪等キーで会計を検索する（soft-deleted 含む）。
	// 見つからない場合は (nil, nil)。
	FindByCompletionRequestID(ctx context.Context, clinicID uint64, requestID string) (*model.Billing, error)
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func scopedPaymentSplitsPreload(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("payment_splits.clinic_id IN ?", clinicIDs).
			Where(`EXISTS (
				SELECT 1
				FROM billings AS payment_split_billing
				WHERE payment_split_billing.id = payment_splits.billing_id
				  AND payment_split_billing.clinic_id = payment_splits.clinic_id
				  AND payment_split_billing.deleted_at IS NULL
			)`)
	}
}

func scopedRefundsPreload(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("billing_refunds.clinic_id IN ?", clinicIDs).
			Where(`EXISTS (
				SELECT 1
				FROM billings AS refund_billing
				WHERE refund_billing.id = billing_refunds.billing_id
				  AND refund_billing.clinic_id = billing_refunds.clinic_id
				  AND refund_billing.deleted_at IS NULL
			)`)
	}
}

func scopedBillingStaffPreload(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			// Historical accounting rows keep the name of an inactive staff
			// member while the staff remains assigned to the billing clinic.
			// Active status is enforced separately on new writes.
			Where("staffs.deleted_at IS NULL").
			Where(`EXISTS (
				SELECT 1
				FROM staff_clinic_assignments AS billing_staff_assignment
				WHERE billing_staff_assignment.staff_id = staffs.id
				  AND billing_staff_assignment.clinic_id IN ?
				  AND billing_staff_assignment.deleted_at IS NULL
			)`, clinicIDs)
	}
}

func scopedStaffAssignmentsPreload(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Where("staff_clinic_assignments.clinic_id IN ?", clinicIDs).
			Where("staff_clinic_assignments.deleted_at IS NULL")
	}
}

func staffHasClinicAssignment(staff *model.Staff, clinicID uint64) bool {
	if staff == nil || staff.DeletedAt.Valid {
		return false
	}
	for i := range staff.ClinicAssignments {
		assignment := staff.ClinicAssignments[i]
		if assignment.StaffID == staff.ID &&
			assignment.ClinicID == clinicID &&
			!assignment.DeletedAt.Valid {
			return true
		}
	}
	return false
}

func sanitizeBillingStaff(
	staff **model.Staff,
	expectedStaffID *uint64,
	clinicID uint64,
) {
	if *staff == nil {
		return
	}
	if expectedStaffID == nil ||
		(*staff).ID != *expectedStaffID ||
		!staffHasClinicAssignment(*staff, clinicID) {
		*staff = nil
		return
	}
	// ClinicAssignments is loaded only to verify the exact billing/staff
	// correlation. Clone before clearing it because GORM may reuse one preload
	// pointer for the same staff across multiple billing parents.
	sanitized := **staff
	sanitized.ClinicAssignments = nil
	*staff = &sanitized
}

// sanitizeBillingRelations enforces exact parent correlation after GORM's
// batched preloads. A multi-clinic preload can safely query the authorized
// clinic set, but an association from clinic A must still never attach a row
// from clinic B (or a pet owned by someone other than the billing owner).
func sanitizeBillingRelations(billing *model.Billing) {
	if billing.Owner != nil &&
		(billing.OwnerID == nil ||
			billing.Owner.ID != *billing.OwnerID ||
			billing.Owner.ClinicID != billing.ClinicID) {
		billing.Owner = nil
	}
	if billing.Pet != nil &&
		(billing.PetID == nil ||
			billing.Pet.ID != *billing.PetID ||
			billing.Pet.ClinicID != billing.ClinicID ||
			billing.OwnerID == nil ||
			billing.Owner == nil ||
			billing.Pet.OwnerID != *billing.OwnerID) {
		billing.Pet = nil
	}
	for i := range billing.Payments {
		sanitizeBillingStaff(
			&billing.Payments[i].PaidByStaff,
			billing.Payments[i].PaidBy,
			billing.ClinicID,
		)
	}
	for i := range billing.Refunds {
		sanitizeBillingStaff(
			&billing.Refunds[i].RefundedByStaff,
			billing.Refunds[i].RefundedBy,
			billing.ClinicID,
		)
	}
}

func sanitizeBillingSliceRelations(billings []model.Billing) {
	for i := range billings {
		sanitizeBillingRelations(&billings[i])
	}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(persistence.ClinicScope(clinicID))
	return r.findBillingsWithFilters(ctx, q, []uint64{clinicID}, filters, page, limit)
}

func (r *accountingRepository) FindAllForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(persistence.ClinicScopeIn(clinicIDs))
	return r.findBillingsWithFilters(ctx, q, clinicIDs, filters, page, limit)
}

// findBillingsWithFilters はフィルタ・ページネーション適用後に返金合計を付与して返す共通実装。
// FindAll / FindAllForClinics の clinic スコープ差分は呼び出し元で適用済みのクエリ q を受け取る。
// clinicIDs は Owner/Pet Preload の P3.1 clinic_id 述語用（AUD-002）。
// search / status / payment_method はサーバサイド（ページ横断）。
func (r *accountingRepository) findBillingsWithFilters(ctx context.Context, q *gorm.DB, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q = q.Where("(billings.pet_id IS NULL OR EXISTS (SELECT 1 FROM pets p WHERE p.id = billings.pet_id AND p.clinic_id = billings.clinic_id))")
	if filters.PetID != nil {
		q = q.Where("pet_id = ?", *filters.PetID)
	}
	if filters.OwnerID != nil {
		q = q.Where("owner_id = ?", *filters.OwnerID)
	}
	if filters.Status != nil && *filters.Status != "" {
		op := strings.ToLower(strings.TrimSpace(filters.StatusOp))
		if op == "is_not" {
			q = q.Where("status <> ?", *filters.Status)
		} else {
			q = q.Where("status = ?", *filters.Status)
		}
	}
	if filters.StartDate != nil {
		q = q.Where("scheduled_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		q = q.Where("scheduled_date <= ?", *filters.EndDate)
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		q = applyBillingOwnerPetSearch(q, search)
	}
	q = applyBillingPaymentMethodFilter(q, filters.PaymentMethod, filters.PaymentMethodOp)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).Preload("Pet", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", scopedBillingStaffPreload(clinicIDs)).Preload("Payments.PaidByStaff.ClinicAssignments", scopedStaffAssignmentsPreload(clinicIDs)).Preload("Items", "deleted_at IS NULL").Preload("PaymentSplits", scopedPaymentSplitsPreload(clinicIDs)).
		Scopes(persistence.Paginate(page, limit)).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	sanitizeBillingSliceRelations(billings)
	if err := r.attachRefundTotals(ctx, billings); err != nil {
		return nil, 0, err
	}
	return billings, total, nil
}

// attachRefundTotals は billing スライスの各要素に返金合計をサブクエリで一括付与する。
func (r *accountingRepository) attachRefundTotals(ctx context.Context, billings []model.Billing) error {
	if len(billings) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(billings))
	for i := range billings {
		ids = append(ids, billings[i].ID)
	}
	type refundSum struct {
		BillingID uint64
		Total     int64
	}
	var sums []refundSum
	if err := r.db.WithContext(ctx).
		Model(&model.BillingRefund{}).
		Unscoped().
		Select(
			"billing_refunds.billing_id,"+
				" COALESCE(SUM(billing_refunds.amount), 0) AS total",
		).
		Joins(
			"JOIN billings ON billings.id = billing_refunds.billing_id"+
				" AND billings.clinic_id = billing_refunds.clinic_id"+
				" AND billings.deleted_at IS NULL",
		).
		Where("billing_refunds.billing_id IN ?", ids).
		Group("billing_refunds.billing_id").
		Scan(&sums).Error; err != nil {
		return apperrors.FromGORM(err, "billing_refund", "")
	}
	sumMap := make(map[uint64]int64, len(sums))
	for _, s := range sums {
		sumMap[s.BillingID] = s.Total
	}
	for i := range billings {
		billings[i].TotalRefundedAmount = sumMap[billings[i].ID]
	}
	return nil
}

func (r *accountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return r.findBillingByIDWithScope(persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)), []uint64{clinicID}, id)
}

func (r *accountingRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	return r.findBillingByIDWithScope(persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScopeIn(clinicIDs)), clinicIDs, id)
}

// findBillingByIDWithScope は clinic スコープ適用済みのクエリで billing を1件取得し、返金合計を計算して返す。
// clinicIDs は Owner/Pet Preload の P3.1 clinic_id 述語用（AUD-002）。
func (r *accountingRepository) findBillingByIDWithScope(q *gorm.DB, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := q.
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Preload("Payments.PaidByStaff", scopedBillingStaffPreload(clinicIDs)).
		Preload("Payments.PaidByStaff.ClinicAssignments", scopedStaffAssignmentsPreload(clinicIDs)).
		Preload("Refunds", scopedRefundsPreload(clinicIDs)).
		Preload("Refunds.RefundedByStaff", scopedBillingStaffPreload(clinicIDs)).
		Preload("Refunds.RefundedByStaff.ClinicAssignments", scopedStaffAssignmentsPreload(clinicIDs)).
		Preload("Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Pet", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("PaymentSplits", scopedPaymentSplitsPreload(clinicIDs)).
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	sanitizeBillingRelations(&billing)
	var total int64
	for i := range billing.Refunds {
		total += billing.Refunds[i].Amount
	}
	billing.TotalRefundedAmount = total
	return &billing, nil
}

// LockAndFindByID は FOR UPDATE で請求を行ロック取得する。
// refund_service の CreateRefund・accounting_service_correction の CorrectCreditPayment の
// トランザクション内で使用し、TOCTOU を防止する。
// BE-refactor.md R1-1 (D2): dbOrTx で ambient tx に参加する。参加しないと FOR UPDATE ロックが
// 別セッションで即座に解放され、TOCTOU 防止が機能しない（過去は r.db.WithContext(ctx) 直参照だった）。
func (r *accountingRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Preload("Payments.PaidByStaff", scopedBillingStaffPreload([]uint64{clinicID})).
		Preload("Payments.PaidByStaff.ClinicAssignments", scopedStaffAssignmentsPreload([]uint64{clinicID})).
		Preload("Refunds", scopedRefundsPreload([]uint64{clinicID})).
		Preload("Refunds.RefundedByStaff", scopedBillingStaffPreload([]uint64{clinicID})).
		Preload("Refunds.RefundedByStaff.ClinicAssignments", scopedStaffAssignmentsPreload([]uint64{clinicID})).
		Preload("PaymentSplits", scopedPaymentSplitsPreload([]uint64{clinicID})).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	sanitizeBillingRelations(&billing)
	// Preload した Refunds から TotalRefundedAmount を計算（FindByID と同じ）
	var total int64
	for i := range billing.Refunds {
		total += billing.Refunds[i].Amount
	}
	billing.TotalRefundedAmount = total
	return &billing, nil
}

// FindByCompletionRequestID は BUG-018 の冪等キーで会計を検索する。
// soft-deleted 行も含め（Unscoped）、key 再利用を拒否するため。見つからなければ (nil, nil)。
func (r *accountingRepository) FindByCompletionRequestID(ctx context.Context, clinicID uint64, requestID string) (*model.Billing, error) {
	if requestID == "" {
		return nil, nil
	}
	var billing model.Billing
	err := persistence.DBOrTx(ctx, r.db).
		Unscoped().
		Scopes(persistence.ClinicScope(clinicID)).
		Where("completion_request_id = ?", requestID).
		First(&billing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.FromGORM(err, "billing", requestID)
	}
	return &billing, nil
}

// BE-refactor.md X-12: 会計完了(completed)時、Create は accounting_service_core.Create の
// Transactor.WithTx から txCtx 付きで呼ばれ、後続の reservation.CompleteForAccounting と単一 tx に
// 参加する（dbOrTx が無ければ従来どおり db.WithContext(ctx) と等価・挙動保存）。
func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := persistence.DBOrTx(ctx, r.db).Create(accounting).Error; err != nil {
		if persistence.IsUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.FromGORM(err, "billing", "")
	}
	return nil
}

// Update は指定フィールドのみを更新し、更新後のレコードを返す。
// map[string]any を使うことで GORM のゼロ値スキップ問題を回避する。
// P2: service 層で逆遷移を拒否（修正 1）、repo は RowsAffected チェックで clinic scope/soft-delete を検証
// BE-refactor.md R1-2: Cancel が本メソッドを ambient tx（監査と原子化）から呼ぶため dbOrTx で参加する。
// ambient tx が無ければ従来どおり db.WithContext(ctx) と等価（挙動保存）。
func (r *accountingRepository) Update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := persistence.DBOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", scopedBillingStaffPreload([]uint64{clinicID})).Preload("Payments.PaidByStaff.ClinicAssignments", scopedStaffAssignmentsPreload([]uint64{clinicID})).Preload("Refunds", scopedRefundsPreload([]uint64{clinicID})).Preload("Refunds.RefundedByStaff", scopedBillingStaffPreload([]uint64{clinicID})).Preload("Refunds.RefundedByStaff.ClinicAssignments", scopedStaffAssignmentsPreload([]uint64{clinicID})).Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("PaymentSplits", scopedPaymentSplitsPreload([]uint64{clinicID})).
		Scopes(persistence.ClinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	sanitizeBillingRelations(&billing)
	for i := range billing.Refunds {
		billing.TotalRefundedAmount += billing.Refunds[i].Amount
	}
	return &billing, nil
}

// BE-refactor.md R1-1 (D2): accounting_service_core.Update・accounting_service_correction.
// CorrectCreditPayment の両方が本メソッドを ambient tx 内から txCtx 付きで呼ぶため dbOrTx で参加する。
func (r *accountingRepository) SavePayment(ctx context.Context, payment *model.Payment) error {
	if payment == nil || payment.BillingID == 0 {
		return apperrors.WrapInvalidInput("payment requires billing_id")
	}
	// map[string]any を使用してゼロ値（Subtotal=0 等）も確実に更新する。
	// struct の Assign では GORM がゼロ値フィールドをスキップする問題がある。
	fields := map[string]any{
		"subtotal":         payment.Subtotal,
		"tax_total":        payment.TaxTotal,
		"total_amount":     payment.TotalAmount,
		"insurance_name":   payment.InsuranceName,
		"insurance_ratio":  payment.InsuranceRatio,
		"insurance_amount": payment.InsuranceAmount,
		"discount_amount":  payment.DiscountAmount,
		"billing_amount":   payment.BillingAmount,
		"received_amount":  payment.ReceivedAmount,
		"change_amount":    payment.ChangeAmount,
		"method":           payment.Method,
		"paid_by":          payment.PaidBy,
	}

	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		clinicID, err := lockBillingClinic(tx, payment.BillingID)
		if err != nil {
			return err
		}
		// TASK-445: payments.clinic_id is an internal tenant key derived from the
		// locked billing row (never from request body). Set always so create and
		// update stay consistent with the billing tenant.
		payment.ClinicID = clinicID
		fields["clinic_id"] = payment.ClinicID
		if payment.PaidBy != nil {
			if err := lockActiveBillingStaffs(
				tx,
				clinicID,
				[]uint64{*payment.PaidBy},
			); err != nil {
				return err
			}
		}

		var existing model.Payment
		err = tx.
			Where("billing_id = ?", payment.BillingID).
			First(&existing).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.FromGORM(
					err,
					"payment",
					fmt.Sprintf("billing_id=%d", payment.BillingID),
				)
			}
			if err := tx.Create(payment).Error; err != nil {
				return apperrors.FromGORM(
					err,
					"payment",
					fmt.Sprintf("billing_id=%d", payment.BillingID),
				)
			}
			return nil
		}

		if err := tx.
			Model(&model.Payment{}).
			Where("billing_id = ?", payment.BillingID).
			Updates(fields).Error; err != nil {
			return apperrors.FromGORM(
				err,
				"payment",
				fmt.Sprintf("billing_id=%d", payment.BillingID),
			)
		}
		payment.ID = existing.ID
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to save payment")
	}
	return nil
}

// SavePaymentSplits は billing の payment_splits を delete-then-recreate で保存する。
// splits が空の場合は既存レコードを削除のみ行う。
// P4: DELETE に clinic_id = ? を付与しテナント越境削除を防ぐ（splits[0].ClinicID = 呼び出し元の clinicID）。
// BE-refactor.md R1-1 (D2): persistence.DBOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx があれば
// その中のネスト tx（SAVEPOINT）として参加する。過去は r.db.WithContext(ctx).Transaction(...) で
// 常に独立した新規 tx を開始しており、ambient tx が rollback しても本メソッドの書込は
// 既にコミット済みのため巻き戻らない部分コミットのバグがあった。
func (r *accountingRepository) SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error {
	if len(splits) == 0 {
		return nil
	}
	billingID := splits[0].BillingID
	clinicID := splits[0].ClinicID
	if billingID == 0 || clinicID == 0 {
		return apperrors.WrapInvalidInput(
			"payment splits require billing_id and clinic_id",
		)
	}
	for i := range splits {
		if splits[i].BillingID != billingID ||
			splits[i].ClinicID != clinicID {
			return apperrors.WrapInvalidInput(
				"payment splits must belong to one billing and clinic",
			)
		}
	}
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var billing model.Billing
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", billingID).
			First(&billing).Error; err != nil {
			return apperrors.FromGORM(
				err,
				"billing",
				fmt.Sprintf("%d", billingID),
			)
		}
		staffIDs := make([]uint64, 0, len(splits))
		for i := range splits {
			if splits[i].PaidBy != nil {
				staffIDs = append(staffIDs, *splits[i].PaidBy)
			}
		}
		if err := lockActiveBillingStaffs(tx, clinicID, staffIDs); err != nil {
			return err
		}
		if err := tx.Where("billing_id = ? AND clinic_id = ?", billingID, clinicID).Delete(&model.PaymentSplit{}).Error; err != nil {
			return apperrors.FromGORM(err, "payment_split", fmt.Sprintf("billing_id=%d", billingID))
		}
		if err := tx.Create(&splits).Error; err != nil {
			return apperrors.FromGORM(err, "payment_split", fmt.Sprintf("billing_id=%d", billingID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to save payment splits")
	}
	return nil
}
