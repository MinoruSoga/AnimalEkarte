package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type AccountingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
	// FindAllForClinics は複数医院の会計を横断検索する (#86 段階3)。clinicIDs はハンドラ層で所属検証済みであること。
	FindAllForClinics(ctx context.Context, clinicIDs []uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error)
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
	// CompleteAccountingAppointments は会計完了に伴い対象 appointment を完了へ進める。
	// (1) 同日同一ペットの会計待ち(accounting)予約、(2) billing.medical_record_id 経由の診察 appointment(status 非依存)。
	CompleteAccountingAppointments(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error)
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
	// SumPaidByOwner は飼い主の支払済み請求合計（LTV）を返す（Lステップタグ同期用）。
	SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// MaxSingleVisitAmountByOwner は飼い主の1回来院最大支払額を返す（CPMスポット判定用）。
	MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	// FindOwnersByAnnualRevenue は直近365日の完了済み請求額合計を飼い主ごとに集計し、降順で返す（LTV上位％判定用）。
	FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]OwnerAnnualRevenue, error)
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(clinicScope(clinicID))
	return r.findBillingsWithFilters(ctx, q, petID, ownerID, status, startDate, endDate, page, limit)
}

func (r *accountingRepository) FindAllForClinics(ctx context.Context, clinicIDs []uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(clinicScopeIn(clinicIDs))
	return r.findBillingsWithFilters(ctx, q, petID, ownerID, status, startDate, endDate, page, limit)
}

// findBillingsWithFilters はフィルタ・ページネーション適用後に返金合計を付与して返す共通実装。
// FindAll / FindAllForClinics の clinic スコープ差分は呼び出し元で適用済みのクエリ q を受け取る。
func (r *accountingRepository) findBillingsWithFilters(ctx context.Context, q *gorm.DB, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startDate != nil {
		q = q.Where("scheduled_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("scheduled_date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", "deleted_at IS NULL").Preload("Items", "deleted_at IS NULL").Preload("PaymentSplits").
		Offset((page - 1) * limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
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
		Select("billing_id, COALESCE(SUM(amount), 0) AS total").
		Where("billing_id IN ?", ids).
		Group("billing_id").
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
	return r.findBillingByIDWithScope(r.db.WithContext(ctx).Scopes(clinicScope(clinicID)), id)
}

func (r *accountingRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	return r.findBillingByIDWithScope(r.db.WithContext(ctx).Scopes(clinicScopeIn(clinicIDs)), id)
}

// findBillingByIDWithScope は clinic スコープ適用済みのクエリで billing を1件取得し、返金合計を計算して返す。
func (r *accountingRepository) findBillingByIDWithScope(q *gorm.DB, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := q.
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Preload("Payments.PaidByStaff", "deleted_at IS NULL").
		Preload("Refunds").
		Preload("Refunds.RefundedByStaff").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("PaymentSplits").
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
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
	err := dbOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Preload("Payments.PaidByStaff", "deleted_at IS NULL").
		Preload("PaymentSplits").
		Scopes(clinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	// Preload した Refunds から TotalRefundedAmount を計算（FindByID と同じ）
	var total int64
	for i := range billing.Refunds {
		total += billing.Refunds[i].Amount
	}
	billing.TotalRefundedAmount = total
	return &billing, nil
}

// BE-refactor.md X-12: 会計完了(completed)時、Create は accounting_service_core.Create の
// Transactor.WithTx から txCtx 付きで呼ばれ、後続の CompleteAccountingAppointments と単一 tx に
// 参加する（dbOrTx が無ければ従来どおり db.WithContext(ctx) と等価・挙動保存）。
func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := dbOrTx(ctx, r.db).Create(accounting).Error; err != nil {
		if isUniqueConstraintErr(err) {
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
	result := dbOrTx(ctx, r.db).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		// clinic scope 外 or soft-delete 済み（service 層逆遷移ガード後に clinic scope 外になる場合）
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := dbOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").Preload("Payments", "deleted_at IS NULL").Preload("Payments.PaidByStaff", "deleted_at IS NULL").Preload("Refunds").Preload("Refunds.RefundedByStaff").Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("PaymentSplits").
		Scopes(clinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	return &billing, nil
}

// BE-refactor.md R1-1 (D2): accounting_service_core.Update・accounting_service_correction.
// CorrectCreditPayment の両方が本メソッドを ambient tx 内から txCtx 付きで呼ぶため dbOrTx で参加する。
func (r *accountingRepository) SavePayment(ctx context.Context, payment *model.Payment) error {
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

	var existing model.Payment
	err := dbOrTx(ctx, r.db).
		Where("billing_id = ?", payment.BillingID).
		First(&existing).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			// DB エラー → 変換して返す
			return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
		}
		// レコードなし → 新規作成
		if err := dbOrTx(ctx, r.db).Create(payment).Error; err != nil {
			return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
		}
		return nil
	}

	// 既存レコード → map で更新（ゼロ値も反映）
	if err := dbOrTx(ctx, r.db).
		Model(&model.Payment{}).
		Where("billing_id = ?", payment.BillingID).
		Updates(fields).Error; err != nil {
		return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
	}
	payment.ID = existing.ID
	return nil
}

// SavePaymentSplits は billing の payment_splits を delete-then-recreate で保存する。
// splits が空の場合は既存レコードを削除のみ行う。
// P4: DELETE に clinic_id = ? を付与しテナント越境削除を防ぐ（splits[0].ClinicID = 呼び出し元の clinicID）。
// BE-refactor.md R1-1 (D2): dbOrTx(ctx, r.db).Transaction(...) にすることで、ambient tx があれば
// その中のネスト tx（SAVEPOINT）として参加する。過去は r.db.WithContext(ctx).Transaction(...) で
// 常に独立した新規 tx を開始しており、ambient tx が rollback しても本メソッドの書込は
// 既にコミット済みのため巻き戻らない部分コミットのバグがあった。
func (r *accountingRepository) SavePaymentSplits(ctx context.Context, splits []model.PaymentSplit) error {
	if len(splits) == 0 {
		return nil
	}
	billingID := splits[0].BillingID
	clinicID := splits[0].ClinicID
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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

// BE-refactor.md X-12: accounting_service_core.Create/Update の Transactor.WithTx から txCtx
// 付きで呼ばれ、billing 本体の書込（Create/Update）と同一 tx に参加する。dbOrTx が無ければ
// 従来どおり db.WithContext(ctx) と等価（挙動保存）。medical_record サブクエリも読み取り一貫性
// のため dbOrTx に揃える（同一 tx 内の書込に対する読み取りを ambient tx から行う）。
func (r *accountingRepository) CompleteAccountingAppointments(ctx context.Context, clinicID uint64, medicalRecordID, ownerID, petID *uint64, scheduledDate time.Time) (int64, error) {
	var totalAffected int64

	// (1) 同日同一ペットの会計待ち(accounting)予約を完了化する（トリミング + 受付カンバンで会計待ちに進めた診察）。
	if ownerID != nil && petID != nil && !scheduledDate.IsZero() {
		result := dbOrTx(ctx, r.db).
			Model(&model.Reservation{}).
			Where("clinic_id = ? AND owner_id = ? AND pet_id = ? AND status = ? AND deleted_at IS NULL",
				clinicID, *ownerID, *petID, model.ReservationStatusAccounting).
			Where("DATE(start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", scheduledDate).
			Update("status", model.ReservationStatusCompleted)
		if result.Error != nil {
			return totalAffected, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("clinic=%d owner=%d pet=%d scheduled_date=%s", clinicID, *ownerID, *petID, scheduledDate.Format(time.DateOnly)))
		}
		totalAffected += result.RowsAffected
	}

	// (2) billing.medical_record_id 経由で診察 appointment を直接完了化する（status 非依存・orphan 根絶）。
	//     診察は billing_confirmation の医師確認だけで会計可能なため、受付カンバンで会計待ち(accounting)に
	//     進めずに会計すると (1) の条件に合致せず、会計後も診察カードが受付ボードに残る。これを防ぐ。
	if medicalRecordID != nil {
		result := dbOrTx(ctx, r.db).
			Model(&model.Reservation{}).
			Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
			Where("status NOT IN ?", []model.ReservationStatus{model.ReservationStatusCompleted, model.ReservationStatusCancelled, model.ReservationStatusNoShow}).
			Where("id IN (?)",
				dbOrTx(ctx, r.db).Model(&model.MedicalRecord{}).
					Select("appointment_id").
					Where("id = ? AND clinic_id = ? AND appointment_id IS NOT NULL AND deleted_at IS NULL", *medicalRecordID, clinicID)).
			Update("status", model.ReservationStatusCompleted)
		if result.Error != nil {
			return totalAffected, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("clinic=%d medical_record=%d", clinicID, *medicalRecordID))
		}
		totalAffected += result.RowsAffected
	}

	return totalAffected, nil
}
