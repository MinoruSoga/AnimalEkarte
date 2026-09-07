package reservation

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationTypeGroupRepository is the data access interface for reservation type groups.
type ReservationTypeGroupRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	CountUsageByReservationTypeGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error)
	Create(ctx context.Context, g *model.ReservationTypeGroup) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeGroupRepository struct{ db *gorm.DB }

// New constructs a ReservationTypeGroupRepository.
func NewReservationTypeGroupRepository(db *gorm.DB) ReservationTypeGroupRepository {
	return &reservationTypeGroupRepository{db: db}
}

func (r *reservationTypeGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	var list []model.ReservationTypeGroup
	if err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&list).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_group", "")
	}
	return list, nil
}

func (r *reservationTypeGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	return persistence.FindByIDScoped[model.ReservationTypeGroup](ctx, r.db, "reservation_type_group", clinicID, id)
}

func (r *reservationTypeGroupRepository) CountUsageByReservationTypeGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("group_id = ? AND deleted_at IS NULL", groupID).Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type", "")
	}
	return count, nil
}

func (r *reservationTypeGroupRepository) Create(ctx context.Context, g *model.ReservationTypeGroup) error {
	db := r.db.WithContext(ctx)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := g.IsActive
	if err := db.Create(g).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_group", "")
	}
	if !wantActive {
		if err := db.Model(g).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_group", fmt.Sprintf("%d", g.ID))
		}
		g.IsActive = false
	}
	return nil
}

func (r *reservationTypeGroupRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
	if err := r.update(ctx, clinicID, id, buildReservationTypeGroupUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationTypeGroupRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, id, fields)
}

func (r *reservationTypeGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM reservation_types
			WHERE reservation_types.group_id = reservation_type_groups.id
			  AND reservation_types.clinic_id = ?
			  AND reservation_types.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.ReservationTypeGroup{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDeleteIfUnusedMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *reservationTypeGroupRepository) normalizeDeleteIfUnusedMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	return apperrors.WrapConflict("このグループには予約区分が設定されています。先に予約区分のグループを変更してください。")
}

func (r *reservationTypeGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, ids, "sort_order")
}
