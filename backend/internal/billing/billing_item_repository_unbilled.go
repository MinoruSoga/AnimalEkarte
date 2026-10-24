package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Unbilled candidate queries for BillingItemRepository.
// 請求元の固定値 projection と billing_source_bindings が未接続（EMR-481）のため、旧 SQL
// （billing_items.vaccination_id・exam_id・appointment_id・trimming_* と billings.status、billing_confirmations を読む。
// いずれも廃止された列・表）を削除し、候補を返す 3 つのメソッドは固定エラー（422 finance_consumer_pending）を返す。
// メソッド名は BillingItemRepository の契約と lintscan の inventory のために残している。held for EMR-481（EMR-505 / Q7）。

// FindUnbilledVaccinationItemsByPetID は未請求の予防接種候補。held for EMR-481。
func (*billingItemRepository) FindUnbilledVaccinationItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, int, error) {
	return nil, 0, apperrors.WrapFinanceConsumerPending(unbilledPendingMessage)
}

// FindUnbilledExamItemsByPetID は未請求の検査候補。held for EMR-481。
func (*billingItemRepository) FindUnbilledExamItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, int, error) {
	return nil, 0, apperrors.WrapFinanceConsumerPending(unbilledPendingMessage)
}

// FindUnbilledTrimmingItemsByPetID は未請求のトリミング候補。held for EMR-481。
func (*billingItemRepository) FindUnbilledTrimmingItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, apperrors.WrapFinanceConsumerPending(unbilledPendingMessage)
}

// CountNonAccountingTrimmingByPetAndDate は同日同ペットの「未会計対象化」トリミング appointment 件数を返す(#77)。
// トリミング予約区分で status が accounting/completed/cancelled でない = まだ会計待ちに進んでいない取り残し候補。
func (r *billingItemRepository) CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Joins("JOIN reservation_types rt ON rt.id = appointments.reservation_type_id AND rt.clinic_id = appointments.clinic_id AND rt.deleted_at IS NULL").
		Where("appointments.clinic_id = ? AND appointments.pet_id = ? AND appointments.deleted_at IS NULL", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = appointments.pet_id AND p.clinic_id = appointments.clinic_id)").
		Where("rt.category = ?", model.ReservationTypeCategoryTrimming).
		Where("appointments.status NOT IN ?", []model.ReservationStatus{
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
			model.ReservationStatusCancelled,
		}).
		Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", date).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "appointment", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, petID))
	}
	return count, nil
}
