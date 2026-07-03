package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// staffAssignedToClinicsCond は Preload した staff（Doctor / CreatedByStaff）を、指定クリニック集合の
// いずれかに現在所属している場合のみ表示する条件。staff は staff_clinic_assignments による多医院所属のため
// staffs.clinic_id（主所属）単純スコープでは共有スタッフを誤って隠す。assignment-EXISTS で多医院所属を
// 尊重しつつ、別テナント単独所属スタッフ名の漏洩を防ぐ。予約は現在/未来データのため履歴表示の回帰はない。
const staffAssignedToClinicsCond = "deleted_at IS NULL AND EXISTS (SELECT 1 FROM staff_clinic_assignments sca WHERE sca.staff_id = staffs.id AND sca.clinic_id IN ? AND sca.deleted_at IS NULL)"

// ReservationCRUDRepository はコア CRUD 操作（5 メソッド）。
// reservation_service / trimming_service / liff_service で使用。
type ReservationCRUDRepository interface {
	// FindAll は指定した複数医院 (#86 拠点横断) の予約を検索する。clinicIDs はハンドラ層で所属検証済みであること。
	FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// FindByIDForClinics は複数医院スコープで予約を1件取得する (#86 詳細画面拠点横断)。clinicIDs はハンドラ層で所属検証済みであること。
	FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, reservation *model.Reservation) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// ReservationSlotRepository はトランザクション内の競合チェック操作（4 メソッド）。
// dbOrTx でコンテキストの tx を自動使用。reservation_service で使用。
type ReservationSlotRepository interface {
	// LockAndFindByID は FOR UPDATE で予約を行ロック取得する（updateWithConflictCheck 用）。
	LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// HasDoctorConflict は指定医師の時間枠重複を SELECT FOR UPDATE でチェックする。
	HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error)
	// CountOnDutyDoctors は当日の出勤医師数を返す。
	CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	// CountConflicts は時間枠の競合予約数を SELECT FOR UPDATE で返す。
	CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	// CountByTypeAndStartTime は同一予約区分・同一開始時刻の予約件数を返す。
	CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
}

// ReservationQueryRepository はクロスフィーチャーのクエリ・依存チェック（6 メソッド）。
// reservation_type_service / staff_service / liff_service / reservation_validators で使用。
type ReservationQueryRepository interface {
	ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error)
	ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error)
	// CountMedicalRecordsByReservationID は BE-refactor.md R2-5 (D12) で clinic_id 述語を追加した
	// （唯一の呼び出し元 reservationService.Delete が clinicID を既に保持・ownership 検証済み）。
	CountMedicalRecordsByReservationID(ctx context.Context, clinicID, reservationID uint64) (int64, error)
	// CountByCustomerAndDateRange は顧客・期間での予約件数を返す（日次・月次制限チェック用）。
	CountByCustomerAndDateRange(ctx context.Context, clinicID, customerID uint64, start, end time.Time) (int64, error)
	// CountByDateAndSource は日付・ソースの予約件数を返す（確認番号生成用）。
	CountByDateAndSource(ctx context.Context, clinicID uint64, date time.Time, source model.ReservationSource) (int64, error)
	// FindAllByCategory はカテゴリ（'general'/'trimming'）でフィルタした予約一覧を返す。
	// トリミング管理APIが appointments ベースで動作するために使用（BE-119）。
	FindAllByCategory(ctx context.Context, clinicID uint64, category model.ReservationTypeCategory, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	// FindNoShowCandidates は end_time が現在時刻以前の confirmed/pending 予約を返す（BE-014 ノーショウ検知用）。
	FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	// HasReservationByOwnerInRange は指定飼い主の start〜end 期間内に予約が存在するか返す（FEAT-383 次回来院リマインド用）。
	HasReservationByOwnerInRange(ctx context.Context, clinicID, ownerID uint64, start, end time.Time) (bool, error)
}

// ReservationRepository は 3 つのサブインターフェースを合成したフルインターフェース。
// DI 配線と、複数グループのメソッドが必要なサービスで使用する。
type ReservationRepository interface {
	ReservationCRUDRepository
	ReservationSlotRepository
	ReservationQueryRepository
}

type reservationRepository struct {
	db *gorm.DB
}

// NewReservationRepository は ReservationRepository の実装を返す。
func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	reservations := make([]model.Reservation, 0)
	var total int64

	// フェイルセーフ: 検証バグ等で空スライスが渡っても全件露出させない
	if len(clinicIDs) == 0 {
		return reservations, 0, nil
	}

	q := dbOrTx(ctx, r.db).Model(&model.Reservation{}).Scopes(clinicScopeIn(clinicIDs))
	switch {
	case date != nil:
		// 単日フィルタ（当日受付など）
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.Add(24 * time.Hour)
		q = q.Where("start_time >= ? AND start_time < ?", start, end)
	case startDate != nil && endDate != nil:
		// 期間レンジフィルタ（予約管理カレンダーの表示中の週/月）。endDate は排他的上限
		q = q.Where("start_time >= ? AND start_time < ?", *startDate, *endDate)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if source != nil {
		q = q.Where("source = ?", *source)
	}
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	if err := q.Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Pet.Owner", "deleted_at IS NULL").Preload("Pet.AnimalSpecies").Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).Preload("ReservationType.Group", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).Preload("Doctor", staffAssignedToClinicsCond, clinicIDs).
		Offset((page - 1) * limit).Limit(limit).Order("start_time ASC").Find(&reservations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, total, nil
}

func (r *reservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	return r.findReservationByID(ctx, []uint64{clinicID}, id)
}

func (r *reservationRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	return r.findReservationByID(ctx, clinicIDs, id)
}

// findReservationByID は認可済みクリニック集合を受け取り予約を1件取得する共通実装。
// Preload する診療区分マスタも同じ集合で clinic 隔離する（別クリニックの診療区分混入防止）。
func (r *reservationRepository) findReservationByID(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Reservation, error) {
	if len(clinicIDs) == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	var reservation model.Reservation
	err := dbOrTx(ctx, r.db).
		Preload("Owner", "deleted_at IS NULL").
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.Owner", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("ReservationType.Group", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs).
		Preload("Doctor", staffAssignedToClinicsCond, clinicIDs).
		Preload("CreatedByStaff", staffAssignedToClinicsCond, clinicIDs).
		Scopes(clinicScopeIn(clinicIDs)).Where("id = ?", id).First(&reservation).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	return &reservation, nil
}

func (r *reservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	if err := dbOrTx(ctx, r.db).Create(reservation).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("reservation", reservation.StartTime.String())
		}
		return apperrors.FromGORM(err, "reservation", "")
	}
	return nil
}

func (r *reservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	result := dbOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := dbOrTx(ctx, r.db).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Reservation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationRepository) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	var count int64
	err := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", reservationTypeID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

func (r *reservationRepository) ExistsByStaffID(ctx context.Context, clinicID, staffID uint64) (bool, error) {
	var count int64
	err := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
		Where("doctor_id = ? AND deleted_at IS NULL", staffID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", "")
	}
	return count > 0, nil
}

// CountMedicalRecordsByReservationID は予約を参照しているカルテの件数を返す（BUG-201）
// BE-refactor.md R2-5 (D12): clinic_id 述語を追加（medical_records は clinic_id カラムを直接持つ）。
func (r *reservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, clinicID, reservationID uint64) (int64, error) {
	var count int64
	if err := dbOrTx(ctx, r.db).
		Model(&model.MedicalRecord{}).
		Where("appointment_id = ? AND clinic_id = ? AND deleted_at IS NULL", reservationID, clinicID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", "")
	}
	return count, nil
}

// LockAndFindByID は FOR UPDATE で予約を行ロック取得する。
// updateWithConflictCheck のトランザクション内で使用する。
func (r *reservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	var appt model.Reservation
	err := dbOrTx(ctx, r.db).Raw(
		`SELECT * FROM appointments WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL FOR UPDATE`,
		clinicID, id,
	).Scan(&appt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
	}
	if appt.ID == 0 {
		return nil, apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

// HasDoctorConflict は指定医師の時間枠重複を SELECT FOR UPDATE でチェックする。
func (r *reservationRepository) HasDoctorConflict(ctx context.Context, clinicID, doctorID uint64, start, end time.Time, excludeID *uint64) (bool, error) {
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := dbOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled', 'no_show')
		  AND start_time < ?
		  AND end_time > ?
		  AND doctor_id = ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, doctorID, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return false, apperrors.Wrap(err, "lock reservations for doctor conflict check")
	}
	return len(existing) > 0, nil
}

// CountOnDutyDoctors は当日の出勤医師数を返す。
func (r *reservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	var count int64
	err := dbOrTx(ctx, r.db).Raw(`
		SELECT COUNT(DISTINCT se.staff_id)
		FROM shift_entries se
		JOIN staffs s ON s.id = se.staff_id
		WHERE se.clinic_id = ?
		  AND se.date = DATE(? AT TIME ZONE 'Asia/Tokyo')
		  AND se.shift_type NOT IN ('off', 'paid_leave')
		  AND s.staff_type = 'doctor'
		  AND s.is_active = true
		  AND s.deleted_at IS NULL`,
		clinicID, date,
	).Scan(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count on-duty doctors")
	}
	return count, nil
}

// CountConflicts は時間枠の競合予約数を SELECT FOR UPDATE で返す。
func (r *reservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	var existing []struct{ ID uint64 }
	excl := uint64(0)
	if excludeID != nil {
		excl = *excludeID
	}
	err := dbOrTx(ctx, r.db).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled')
		  AND start_time < ?
		  AND end_time > ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, end, start, excl, excl,
	).Scan(&existing).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "lock reservations for capacity check")
	}
	return int64(len(existing)), nil
}

func (r *reservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	var count int64
	q := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
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

// countByTypeAndStartTimeRow is the GROUP BY scan target for CountByTypeAndStartTimes.
type countByTypeAndStartTimeRow struct {
	StartTime time.Time
	Count     int64
}

// CountByTypeAndStartTimes は複数の開始時刻の予約件数を一括取得する（GROUP BY start_time）。
// BE-refactor.md R2-4 (D8): liff_service.filterSlotsByCapacity の N+1（日付ごとの各スロットで
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
	q := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
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
	err := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
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

	q := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
		Where("appointments.clinic_id = ?", clinicID).
		Joins("JOIN reservation_types ON reservation_types.id = appointments.reservation_type_id AND reservation_types.deleted_at IS NULL").
		Where("reservation_types.category = ?", category)

	if petID != nil {
		q = q.Where("appointments.pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN pets ON pets.id = appointments.pet_id AND pets.deleted_at IS NULL").
			Where("pets.owner_id = ?", *ownerID)
	}
	if startDate != nil {
		start, err := parseJSTDate(*startDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time >= ?", start)
	}
	if endDate != nil {
		end, err := parseJSTDate(*endDate)
		if err != nil {
			return nil, 0, apperrors.WrapInvalidInput(err.Error())
		}
		q = q.Where("appointments.start_time < ?", end.AddDate(0, 0, 1))
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "appointment", "")
	}
	if err := q.
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.Owner", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies").
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("TrimmingDetail.Course", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("TrimmingDetail.Options", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Offset((page - 1) * limit).Limit(limit).
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
	start, end := appointmentDayRange(date)
	err := dbOrTx(ctx, r.db).Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ? AND source = ? AND deleted_at IS NULL", start, end, source).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// FindNoShowCandidates は end_time が現在時刻以前の confirmed/pending 予約を返す（BE-014）。
func (r *reservationRepository) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	var reservations []model.Reservation
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND deleted_at IS NULL AND status IN ? AND end_time <= NOW() - interval '4 hours'",
			clinicID,
			[]string{string(model.ReservationStatusConfirmed), string(model.ReservationStatusPending)}).
		Find(&reservations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation", "")
	}
	return reservations, nil
}

// HasReservationByOwnerInRange は指定飼い主の start〜end 期間内に予約が存在するか返す（FEAT-383）。
func (r *reservationRepository) HasReservationByOwnerInRange(ctx context.Context, clinicID, ownerID uint64, start, end time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Reservation{}).
		Where("clinic_id = ? AND owner_id = ? AND deleted_at IS NULL AND start_time >= ? AND start_time < ?",
			clinicID, ownerID, start, end).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "reservation", fmt.Sprintf("clinic=%d owner=%d range", clinicID, ownerID))
	}
	return count > 0, nil
}

func appointmentDayRange(date time.Time) (start, end time.Time) {
	dateJST := date.In(jstLoc)
	start = time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, dateJST.Location())
	end = start.AddDate(0, 0, 1)
	return start, end
}

func parseJSTDate(value string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", value, jstLoc)
	if err != nil {
		return time.Time{}, apperrors.WrapInvalidInput("date must be YYYY-MM-DD format")
	}
	return t, nil
}
