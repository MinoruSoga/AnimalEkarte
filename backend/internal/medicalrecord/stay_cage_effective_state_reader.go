package medicalrecord

import (
	"context"
	"slices"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type stayCageEffectiveStateReader struct{ db *gorm.DB }

// NewStayCageEffectiveStateReader provides the confirmed normal-case subset of
// the EMR-499 core. Unspecified reservation/correction cases remain unknown.
func NewStayCageEffectiveStateReader(db *gorm.DB) StayCageEffectiveStateReader {
	return &stayCageEffectiveStateReader{db: db}
}

func (r *stayCageEffectiveStateReader) ReadEffectiveStatesForCageAssignment(ctx context.Context, clinicID uint64, stayCaseIDs []uint64) ([]StayCageEffectiveState, error) {
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic is required")
	}
	ids := slices.Clone(stayCaseIDs)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	for _, id := range ids {
		if id == 0 {
			return nil, apperrors.WrapInvalidInput("stay case is required")
		}
	}
	if len(ids) == 0 {
		return []StayCageEffectiveState{}, nil
	}
	if r.db == nil {
		return nil, apperrors.WrapServiceUnavailable("stay reader database is not configured")
	}
	var result []StayCageEffectiveState
	read := func(txCtx context.Context) error {
		var err error
		result, err = r.readLocked(txCtx, clinicID, ids)
		return err
	}
	if persistence.TxFromContext(ctx) != nil {
		if err := read(ctx); err != nil {
			return nil, err
		}
	} else if err := persistence.NewTransactor(r.db).WithTx(ctx, read); err != nil {
		return nil, err
	}
	return result, nil
}

// readLocked retains FOR SHARE on the stay rows through the event read and,
// for allocation writes, through the caller's commit. The adopted
// app_private.stay_movement_guard BEFORE INSERT takes FOR UPDATE on this same
// stay row, so a movement INSERT cannot cross this protected read. All writer
// lock ordering still requires owner integration. The adopted immutable-history
// trigger rejects event UPDATE/DELETE; real DB application/acceptance is untested.
func (r *stayCageEffectiveStateReader) readLocked(ctx context.Context, clinicID uint64, ids []uint64) ([]StayCageEffectiveState, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("stay read requires ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	var stays []model.StayCase
	if err := db.Clauses(clause.Locking{Strength: "SHARE"}).
		Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, ids).
		Order("id ASC").Find(&stays).Error; err != nil {
		return nil, apperrors.FromGORM(err, "stay case", "")
	}
	var events []model.StayMovementEvent
	if err := db.Where("clinic_id = ? AND stay_case_id IN ?", clinicID, ids).
		Order("stay_case_id ASC, revision ASC, id ASC").Find(&events).Error; err != nil {
		return nil, apperrors.FromGORM(err, "stay movement", "")
	}
	byStay := make(map[uint64]model.StayCase, len(stays))
	for _, stay := range stays {
		byStay[stay.ID] = stay
	}
	movements := make(map[uint64][]model.StayMovementEvent)
	for _, event := range events {
		movements[event.StayCaseID] = append(movements[event.StayCaseID], event)
	}
	result := make([]StayCageEffectiveState, 0, len(ids))
	for _, id := range ids {
		stay, ok := byStay[id]
		if !ok {
			result = append(result, StayCageEffectiveState{ClinicID: clinicID, StayCaseID: id, State: StayCageUnknown})
			continue
		}
		result = append(result, projectStayCageEffectiveState(stay, movements[id]))
	}
	return result, nil
}
