package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const petOwnerTransferMaxRows = 1000

var petOwnerTransferStatuses = []model.ReservationStatus{
	model.ReservationStatusPending,
	model.ReservationStatusConfirmed,
}

type PetOwnerTransfer struct {
	db *gorm.DB
}

func NewPetOwnerTransfer(db *gorm.DB) *PetOwnerTransfer {
	return &PetOwnerTransfer{db: db}
}

func (t *PetOwnerTransfer) TransferUpcomingPetReservations(
	ctx context.Context,
	clinicID, petID, previousOwnerID, newOwnerID uint64,
	from time.Time,
) ([]uint64, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("pet owner reservation transfer requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, t.db)

	var petRow model.Pet
	if err := db.Where("clinic_id = ? AND id = ?", clinicID, petID).First(&petRow).Error; err != nil {
		return nil, mapPetOwnerTransferError(err)
	}
	if petRow.OwnerID != newOwnerID {
		return nil, apperrors.WrapConflict("ペットの飼主が更新されていないため予約を移管できません")
	}

	var newOwner model.Owner
	if err := db.
		Clauses(clause.Locking{Strength: "SHARE", Options: "NOWAIT"}).
		Where("clinic_id = ? AND id = ?", clinicID, newOwnerID).
		First(&newOwner).Error; err != nil {
		return nil, mapPetOwnerTransferError(err)
	}

	targets := make([]model.Reservation, 0)
	if err := db.
		Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
		Where("clinic_id = ? AND pet_id = ? AND start_time >= ? AND status IN ?",
			clinicID, petID, from, petOwnerTransferStatuses).
		Order("id ASC").
		Limit(petOwnerTransferMaxRows + 1).
		Find(&targets).Error; err != nil {
		return nil, mapPetOwnerTransferError(err)
	}
	if len(targets) > petOwnerTransferMaxRows {
		return nil, apperrors.WrapInvalidInput("移管対象の予約が1000件を超えています")
	}

	targetIDs := make([]uint64, 0, len(targets))
	for i := range targets {
		target := &targets[i]
		if target.OwnerID == nil || *target.OwnerID != previousOwnerID {
			return nil, apperrors.WrapConflict("予約の飼主が移管元と一致しないため移管できません")
		}
		targetIDs = append(targetIDs, target.ID)
	}

	if len(targetIDs) > 0 {
		var records []model.MedicalRecord
		if err := db.
			Select("medical_records.id").
			Joins("JOIN appointments ON appointments.id = medical_records.appointment_id AND appointments.clinic_id = medical_records.clinic_id").
			Where("medical_records.clinic_id = ? AND medical_records.appointment_id IN ?", clinicID, targetIDs).
			Limit(1).
			Find(&records).Error; err != nil {
			return nil, mapPetOwnerTransferError(err)
		}
		if len(records) > 0 {
			return nil, apperrors.WrapConflict("カルテが作成済みの予約があるため移管できません")
		}
	}

	var lineCustomerID *uint64
	if hasLineAssociatedTarget(targets) {
		resolved, err := t.resolveNewOwnerLineCustomer(db, clinicID, &newOwner)
		if err != nil {
			return nil, err
		}
		lineCustomerID = resolved
	}

	for i := range targets {
		target := &targets[i]
		updates := map[string]any{
			"owner_id":   newOwnerID,
			"updated_at": time.Now().UTC(),
		}
		if target.LineCustomerID != nil {
			updates["line_customer_id"] = lineCustomerID
			updates["customer_fields"] = json.RawMessage("{}")
		}
		result := db.Model(&model.Reservation{}).
			Where("clinic_id = ? AND id = ? AND pet_id = ? AND owner_id = ? AND status IN ? AND start_time >= ?",
				clinicID, target.ID, petID, previousOwnerID, petOwnerTransferStatuses, from).
			Updates(updates)
		if result.Error != nil {
			return nil, mapPetOwnerTransferError(result.Error)
		}
		if result.RowsAffected != 1 {
			return nil, apperrors.WrapConflict("予約が移管中に変更されたため移管できません")
		}
	}

	sort.Slice(targetIDs, func(i, j int) bool { return targetIDs[i] < targetIDs[j] })
	return targetIDs, nil
}

func hasLineAssociatedTarget(targets []model.Reservation) bool {
	for i := range targets {
		if targets[i].LineCustomerID != nil {
			return true
		}
	}
	return false
}

func (t *PetOwnerTransfer) resolveNewOwnerLineCustomer(
	db *gorm.DB,
	clinicID uint64,
	newOwner *model.Owner,
) (*uint64, error) {
	if newOwner.LineUserID == nil || strings.TrimSpace(*newOwner.LineUserID) == "" || newOwner.LineIDConfirmedAt == nil {
		return nil, nil
	}
	var customers []model.LineCustomer
	if err := db.
		Clauses(clause.Locking{Strength: "SHARE", Options: "NOWAIT"}).
		Select("id").
		Where("clinic_id = ? AND owner_id = ? AND line_user_id = ?", clinicID, newOwner.ID, *newOwner.LineUserID).
		Limit(2).
		Find(&customers).Error; err != nil {
		return nil, mapPetOwnerTransferError(err)
	}
	if len(customers) != 1 {
		return nil, nil
	}
	return &customers[0].ID, nil
}

func mapPetOwnerTransferError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
		return apperrors.WrapConflict("対象が他の操作で更新中のため移管できません")
	}
	return apperrors.FromGORM(err, "reservation", "")
}
