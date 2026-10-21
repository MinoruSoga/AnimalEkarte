package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/pet"
)

// ReservationAdminRepository は管理者向け予約管理のデータアクセスインターフェース
type ReservationAdminRepository interface {
	FindAllByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error)
	FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	// FindTimeRangesByDateRange は [from, to) 半開区間の予約から id/doctor_id/start_time/end_time/status のみを
	// Preload なしで一括取得する(G7-1: 日付ループN+1回避のプリフェッチ用軽量クエリ)。
	FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error)
	Create(ctx context.Context, r *model.Reservation) error
	SoftDelete(ctx context.Context, clinicID, id uint64) error
	// LIFF用
	FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	// CancelByID returns the owner of the row actually cancelled under the UPDATE lock.
	CancelByID(ctx context.Context, clinicID, customerID, id uint64) (*uint64, error)
	// 通知用（キャンセル前に関連エンティティを含めて取得）
	FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// FindLiffPetPairs は LIFF マイ予約一覧用に、各予約の EMR-520 pet ペア（ローカル pet.id と
	// common 現行 revision 名）を1クエリで取得する。認可パスは同一スナップショットで評価し、
	// 未解決（未紐付・紐付け取り消し・別医院・削除済み等）の予約は map に含めない
	// （呼び出し側で pet_id/pet_name 両方 null に読み替える）。appointmentIDs が空ならクエリを発行しない。
	FindLiffPetPairs(ctx context.Context, clinicID, customerID uint64, appointmentIDs []uint64) (map[uint64]LiffPetPair, error)
}

// LiffPetPair は LIFF 一覧応答の pet_id（ローカル pet.id）と pet_name（common 現行 revision 名）。
// common UUID や revision ID/番号は公開しない。
type LiffPetPair struct {
	PetID   uint64
	PetName string
}

type reservationAdminRepository struct{ db *gorm.DB }

// Safety caps for admin list queries (G2F-07). Calendar month remains the
// primary bound; these hard ceilings prevent pathological Preload growth.
// Customer history returns the most recent rows only (full paging is a
// follow-up outside this repository unit).
const (
	maxAdminMonthRows      = persistence.MaxMasterListRows
	maxCustomerHistoryRows = 100
)

func NewReservationAdminRepository(db *gorm.DB) ReservationAdminRepository {
	return &reservationAdminRepository{db: db}
}

func (r *reservationAdminRepository) FindAllByMonth(ctx context.Context, clinicID uint64, year int, month time.Month) ([]model.Reservation, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("LineCustomer", "clinic_id = ?", clinicID).
		Scopes(persistence.ClinicScope(clinicID), reservationRelationsMatchParentClinic).
		Where("start_time >= ? AND start_time < ?", start, end).
		Order("start_time ASC").
		Limit(maxAdminMonthRows).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	dateJST := date.In(time.Local)
	start := time.Date(dateJST.Year(), dateJST.Month(), dateJST.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)

	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("CreatedByStaff", reservationCreatedByStaffPreload).
		Preload("LineCustomer", "clinic_id = ?", clinicID).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID), reservationRelationsMatchParentClinic).
		Where("start_time >= ? AND start_time < ?", start, end).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

// FindTimeRangesByDateRange は日付範囲計算専用の軽量クエリ。slot 計算は Status/DoctorID/StartTime/EndTime
// しか使わないため、FindAllByDay の6 Preload を伴わない Select 限定版として提供する(G7-1)。
func (r *reservationAdminRepository) FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error) {
	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Select("id", "doctor_id", "start_time", "end_time", "status").
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_time >= ? AND start_time < ?", from, to).
		Order("start_time ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) Create(ctx context.Context, ra *model.Reservation) error {
	if err := assertReservationCreatedBy(ctx, r.db, ra.ClinicID, ra.CreatedBy); err != nil {
		return err
	}
	if err := persistence.DBOrTx(ctx, r.db).Create(ra).Error; err != nil {
		return apperrors.FromGORM(err, "appointment", "")
	}
	return nil
}

func (r *reservationAdminRepository) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, r.db, &model.Reservation{}, "appointment", clinicID, id)
}

func (r *reservationAdminRepository) FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	items := make([]model.Reservation, 0)
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Scopes(persistence.ClinicScope(clinicID), reservationRelationsMatchParentClinic).
		Where("line_customer_id = ? AND status NOT IN ('tentative') AND deleted_at IS NULL", customerID).
		Order("start_time DESC").
		Limit(maxCustomerHistoryRows).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	return items, nil
}

func (r *reservationAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	var appt model.Reservation
	err := r.db.WithContext(ctx).
		Preload("ReservationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", staffAssignedToClinicsCond, []uint64{clinicID}).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID), reservationRelationsMatchParentClinic).Where("id = ?", id).First(&appt).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", fmt.Sprintf("%d", id))
	}
	return &appt, nil
}

func (r *reservationAdminRepository) CancelByID(ctx context.Context, clinicID, customerID, id uint64) (*uint64, error) {
	// BUG-029: LIFF cancel must keep the row visible as status=cancelled
	// (my-reservations + staff calendar). Do not set deleted_at.
	// RETURNING obtains owner_id from the same UPDATE that locks and cancels the
	// row, including a staff owner change committed while that UPDATE waited.
	var cancelled model.Reservation
	result := persistence.DBOrTx(ctx, r.db).
		Model(&cancelled).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "owner_id"}}}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND line_customer_id = ? AND status != ? AND deleted_at IS NULL",
			id, customerID, model.ReservationStatusCancelled).
		Updates(map[string]any{
			"status": model.ReservationStatusCancelled,
		})
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "appointment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("appointment", fmt.Sprintf("%d", id))
	}
	return cancelled.OwnerID, nil
}

// FindLiffPetPairs は EMR-520 の認可パスを1スナップショットで評価する。
// appointments a → 非削除 pet（pet.owner_id = a.owner_id）→ 非削除 owner →
// owner_line_binding_states（同一医院/owner・state='linked'・line_customer_id = a.line_customer_id）
// → pet.JoinCurrentCommonBasicName（common 現行名）。name-match・customer_fields・
// キャッシュ・r.Pet.Name には依らない。
func (r *reservationAdminRepository) FindLiffPetPairs(ctx context.Context, clinicID, customerID uint64, appointmentIDs []uint64) (map[uint64]LiffPetPair, error) {
	pairs := make(map[uint64]LiffPetPair, len(appointmentIDs))
	if len(appointmentIDs) == 0 {
		return pairs, nil
	}
	type row struct {
		AppointmentID uint64 `gorm:"column:appointment_id"`
		PetID         uint64 `gorm:"column:pet_id"`
		PetName       string `gorm:"column:pet_name"`
	}
	var rows []row
	err := pet.JoinCurrentCommonBasicName(
		r.db.WithContext(ctx).
			Table("appointments a").
			Select("a.id AS appointment_id, p.id AS pet_id, rev.name AS pet_name").
			Joins("JOIN pets p ON p.id = a.pet_id AND p.clinic_id = a.clinic_id AND p.deleted_at IS NULL AND p.owner_id = a.owner_id").
			Joins("JOIN owners o ON o.id = a.owner_id AND o.clinic_id = a.clinic_id AND o.deleted_at IS NULL").
			Joins("JOIN owner_line_binding_states lb ON lb.clinic_id = a.clinic_id AND lb.owner_id = a.owner_id AND lb.state = 'linked' AND lb.line_customer_id = a.line_customer_id"),
		"p", "rev",
	).
		Where("a.clinic_id = ? AND a.line_customer_id = ? AND a.id IN ? AND a.deleted_at IS NULL",
			clinicID, customerID, appointmentIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "appointment", "")
	}
	for _, rw := range rows {
		pairs[rw.AppointmentID] = LiffPetPair{PetID: rw.PetID, PetName: rw.PetName}
	}
	return pairs, nil
}
