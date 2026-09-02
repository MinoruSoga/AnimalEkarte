package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func requireReservationRowLockTransaction(ctx context.Context) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("reservation row lock requires an ambient transaction")
	}
	return nil
}

func (r *reservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	var count int64
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Where("clinic_id = ? AND reservation_type_id = ? AND start_time = ? AND status NOT IN ('cancelled') AND deleted_at IS NULL",
			clinicID, reservationTypeID, startTime)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// CountByTypeAndStartTimes は複数の開始時刻の予約件数を一括取得する（GROUP BY start_time）。
// BE-refactor.md R2-4 (D8): liff_service.FilterSlotsByCapacity の N+1（日付ごとの各スロットで
// CountByTypeAndStartTime を個別発行）を解消するためのバッチ経路。CountByTypeAndStartTime を
// 置き換えるものではなく（reservation_service の単発チェックは従来どおり）、追加の一括経路として
// 提供する。戻り値は startTime.Unix() 秒 → count のマップ（time.Time を map key にすると
// Location/monotonic 差異で等価判定が壊れるため Unix 秒で正規化する）。
// startTimes が空の場合は空マップを返す（クエリを発行しない）。
func (r *reservationRepository) CountByTypeAndStartTimes(ctx context.Context, clinicID, reservationTypeID uint64, startTimes []time.Time, excludeID *uint64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(startTimes))
	if len(startTimes) == 0 {
		return result, nil
	}
	var rows []countByTypeAndStartTimeRow
	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Select("start_time, COUNT(*) AS count").
		Where("clinic_id = ? AND reservation_type_id = ? AND start_time IN ? AND status NOT IN ('cancelled') AND deleted_at IS NULL",
			clinicID, reservationTypeID, startTimes).
		Group("start_time")
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	for _, row := range rows {
		result[row.StartTime.Unix()] = row.Count
	}
	return result, nil
}

// CountByCustomerAndDateRange は顧客・期間での予約件数を返す。
// 日次・月次制限チェックで使用する。
func (r *reservationRepository) CountByCustomerAndDateRange(ctx context.Context, clinicID, customerID uint64, start, end time.Time) (int64, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_customer_id = ? AND status NOT IN ('cancelled') AND start_time >= ? AND start_time < ? AND deleted_at IS NULL",
			customerID, start, end,
		).Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// FindAllByCategory はカテゴリでフィルタした予約一覧を返す（BE-119 トリミング管理 API）。
func (r *reservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	reservations := make([]model.Reservation, 0)
	var total int64

	q := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Where("appointments.clinic_id = ?", clinicID).
		Joins("JOIN reservation_types ON reservation_types.id = appointments.reservation_type_id AND reservation_types.clinic_id = appointments.clinic_id AND reservation_types.deleted_at IS NULL").
		Where("reservation_types.category = ?", category)

	if petID != nil || ownerID != nil {
		q = q.Joins("JOIN pets filter_pets ON filter_pets.id = appointments.pet_id AND filter_pets.clinic_id = appointments.clinic_id AND filter_pets.deleted_at IS NULL")
	}
	if petID != nil {
		q = q.Where("filter_pets.id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN owners filter_owners ON filter_owners.id = filter_pets.owner_id AND filter_owners.clinic_id = appointments.clinic_id AND filter_owners.deleted_at IS NULL").
			Where("filter_owners.id = ?", *ownerID)
	}
	if startDate != nil {
		start, err := ParseJSTDate(*startDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time >= ?", start)
	}
	if endDate != nil {
		end, err := ParseJSTDate(*endDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time < ?", end.AddDate(0, 0, 1))
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "appointment", "")
	}
	if err := q.
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("TrimmingDetail", "clinic_id = ?", clinicID).
		Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("TrimmingDetail.Options", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.Paginate(page, limit)).
		Order("appointments.start_time DESC").
		Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "appointment", "")
	}
	return reservations, total, nil
}

// CountByDateAndSource は日付・ソースの予約件数を返す。
// 確認番号生成で使用する。
func (r *reservationRepository) CountByDateAndSource(ctx context.Context, clinicID uint64, date time.Time, source model.ReservationSource) (int64, error) {
	var count int64
	start, end := AppointmentDayRange(date)
	err := persistence.DBOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ? AND source = ? AND deleted_at IS NULL", start, end, source).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// FindNoShowCandidates は終了から4時間以上経過した confirmed/pending 予約のうち、
// 確定済みカルテが存在しないものを返す（BE-014）。
func (r *reservationRepository) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	return r.FindNoShowCandidatesAt(ctx, clinicID, time.Now().UTC())
}

// FindNoShowCandidatesAt evaluates the complete candidate predicate against
// the durable scheduler timestamp instead of database wall-clock time.
func (r *reservationRepository) FindNoShowCandidatesAt(
	ctx context.Context,
	clinicID uint64,
	evaluatedAt time.Time,
) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ? AND deleted_at IS NULL AND status IN ? AND end_time <= CAST(? AS timestamptz) - interval '4 hours'",
			clinicID,
			[]string{string(model.ReservationStatusConfirmed), string(model.ReservationStatusPending)},
			evaluatedAt).
		Where(`NOT EXISTS (
			SELECT 1 FROM medical_records mr
			WHERE mr.clinic_id = appointments.clinic_id
			  AND mr.appointment_id = appointments.id
			  AND mr.status = ?
			  AND mr.deleted_at IS NULL
		)`, model.MedicalRecordStatusFinalized).
		Order("id ASC").
		Limit(noShowCandidateMax).
		Find(&reservations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, nil
}

func AppointmentDayRange(date time.Time) (start, end time.Time) {
	dateJST := date.In(config.JST)
	start = time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, dateJST.Location())
	end = start.AddDate(0, 0, 1)
	return start, end
}

func ParseJSTDate(value string) (time.Time, error) {
	t, err := time.ParseInLocation(time.DateOnly, value, config.JST)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}
	return t, nil
}

// AssertOwnerInClinic は owners を clinic スコープで存在確認する（AUD-001）。
// 別 clinic / 未存在を区別せず NotFound を返す。dbOrTx で ambient tx に参加する。
func (r *reservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	var id uint64
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Owner{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		Select("id").
		Where("id = ?", ownerID).
		Take(&id).Error
	if err != nil {
		return apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", ownerID))
	}
	return nil
}

// FindPetOwnerInClinic は pets とその owner の双方が同一 clinic に属する場合だけ OwnerID を返す。
// transaction 内では両行を共有ロックし、検証後から予約writeまでの clinic/owner 関係変更を防ぐ。
func (r *reservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	var pet model.Pet
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Pet{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("pets.id", "pets.owner_id").
		Joins("JOIN owners ON owners.id = pets.owner_id AND owners.clinic_id = pets.clinic_id AND owners.deleted_at IS NULL").
		Where("pets.id = ? AND pets.clinic_id = ? AND pets.deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	return pet.OwnerID, nil
}

// FindPetByIDInClinic は clinic スコープでペットを読み、死亡 write ガード（SD-10）に必要な列を返す。
// transaction 内では行を共有ロックし、検証後から write までの deceased_at 変更を防ぐ。
func (r *reservationRepository) FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	var pet model.Pet
	db := persistence.DBOrTx(ctx, r.db).Model(&model.Pet{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Select("pets.id", "pets.owner_id", "pets.deceased_at", "pets.status").
		Where("pets.id = ? AND pets.clinic_id = ? AND pets.deleted_at IS NULL", petID, clinicID).
		First(&pet).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "pet", fmt.Sprintf("%d", petID))
	}
	return &pet, nil
}

// AssertLineCustomerInClinic は line_customers を clinic スコープで存在確認する（AUD-001）。
func (r *reservationRepository) AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error {
	var id uint64
	db := persistence.DBOrTx(ctx, r.db).Model(&model.LineCustomer{})
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		Select("id").
		Where("id = ?", lineCustomerID).
		Take(&id).Error
	if err != nil {
		return apperrors.FromGORM(err, "line_customer", fmt.Sprintf("%d", lineCustomerID))
	}
	return nil
}
