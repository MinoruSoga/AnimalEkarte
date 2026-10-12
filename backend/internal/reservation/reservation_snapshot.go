package reservation

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const reservationSnapshotMaxRows = 10000
const reservationSnapshotTimeout = 10 * time.Second

// The opt-in capability is private so legacy consumers keep their pagination contract.
type reservationSnapshotLister interface {
	listSnapshot(context.Context, []uint64, listReservationFilters) ([]model.Reservation, error)
}
type reservationSnapshotFinder interface {
	findSnapshot(context.Context, []uint64, listReservationFilters) ([]model.Reservation, error)
}

func validateReservationSnapshotScope(filters listReservationFilters) error {
	if filters.Date == nil && (filters.StartDate == nil || filters.EndDate == nil) && filters.PetID == nil && filters.OwnerID == nil {
		return apperrors.WrapInvalidInput("予約の期間・ペット・飼主を指定してください")
	}
	if (filters.StartDate == nil) != (filters.EndDate == nil) {
		return apperrors.WrapInvalidInput("start_date と end_date を両方指定してください")
	}
	if filters.StartDate != nil && !filters.StartDate.Before(*filters.EndDate) {
		return apperrors.WrapInvalidInput("予約の期間が不正です")
	}
	return nil
}

func (s *reservationService) listSnapshot(ctx context.Context, clinicIDs []uint64, filters listReservationFilters) ([]model.Reservation, error) {
	if err := validateReservationSnapshotScope(filters); err != nil {
		return nil, err
	}
	finder, ok := s.repo.(reservationSnapshotFinder)
	if !ok {
		return nil, apperrors.WrapNotImplemented("予約一覧の一括取得は未設定です")
	}
	items, err := finder.findSnapshot(ctx, clinicIDs, filters)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read reservation snapshot")
	}
	return items, nil
}

func (r *reservationRepository) findSnapshot(ctx context.Context, clinicIDs []uint64, filters listReservationFilters) ([]model.Reservation, error) {
	if err := validateReservationSnapshotScope(filters); err != nil {
		return nil, err
	}
	rows := make([]model.Reservation, 0)
	if len(clinicIDs) == 0 {
		return rows, nil
	}
	ctx, cancel := context.WithTimeout(ctx, reservationSnapshotTimeout)
	defer cancel()
	// Own the read transaction: count, appointments and every preload use the same
	// PostgreSQL snapshot and connection, including when a row is rescheduled mid-read.
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := reservationListQuery(tx, clinicIDs, filters)
		var total int64
		if err := q.Count(&total).Error; err != nil {
			return apperrors.FromGORM(err, "reservation", "")
		}
		if total > reservationSnapshotMaxRows {
			return apperrors.WrapInvalidInput("予約が10000件を超えています。期間・ペット・飼主を絞り込んでください")
		}
		if err := reservationListPreloads(q, clinicIDs, false).
			Limit(reservationSnapshotMaxRows).Order("start_time ASC").Order("id ASC").Find(&rows).Error; err != nil {
			return apperrors.FromGORM(err, "reservation", "")
		}
		if int64(len(rows)) != total {
			return apperrors.WrapInternalServerError("reservation snapshot row count mismatch")
		}
		return nil
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	return rows, nil
}
