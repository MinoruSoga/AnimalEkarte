// Trimming option persistence belongs to package trimming.
package trimming

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// TrimmingOptionRepository is the trimming option persistence contract.
type TrimmingOptionRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	Create(ctx context.Context, option *model.TrimmingOption) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error)
}

type trimmingOptionRepository struct{ db *gorm.DB }

// NewTrimmingOptionRepository constructs a trimming option repository.
func NewTrimmingOptionRepository(db *gorm.DB) TrimmingOptionRepository {
	return &trimmingOptionRepository{db: db}
}

func (r *trimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	options := make([]model.TrimmingOption, 0)
	if err := persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&options).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", "")
	}
	return options, nil
}

func (r *trimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	var option model.TrimmingOption
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	if err := db.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&option).Error; err != nil {
		return nil, apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", id))
	}
	return &option, nil
}

func (r *trimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := option.IsActive
	wantCombinable := option.IsCombinable
	if err := db.Create(option).Error; err != nil {
		return apperrors.FromGORM(err, "trimming_option", "")
	}
	if !wantActive {
		if err := db.Model(option).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", option.ID))
		}
		option.IsActive = false
	}
	if !wantCombinable {
		if err := db.Model(option).Update("is_combinable", false).Error; err != nil {
			return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("%d", option.ID))
		}
		option.IsCombinable = false
	}
	return nil
}

func (r *trimmingOptionRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	if err := r.update(ctx, clinicID, id, buildTrimmingOptionUpdate(&cmd)); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *trimmingOptionRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.TrimmingOption{}, "trimming_option", clinicID, id, fields)
}

// Delete はオプションの使用中判定ができないため、常に拒否する（EMR-549 C1 F3）。
// 旧判定の appointment_trimming_options は 001_init.sql L10923 で DROP 済みで、置換先
// （trimming_booking_revision_options）の使用数 reader は EMR-480 / EMR-458 の担当で未接続。論理削除は
// FK RESTRICT に守られないため、使用中のオプションも消えてしまう。使用数を確認できない削除は行わない。
func (r *trimmingOptionRepository) Delete(_ context.Context, _, _ uint64) error {
	return apperrors.WrapTrimmingMasterUsageUnverified()
}

func (r *trimmingOptionRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.TrimmingOption{}, "trimming_option", clinicID, ids, "sort_order")
}

// CountUsageByTrimmingOptionID は使用数を確認できないため、常に拒否エラーを返す（EMR-549 C1 F3）。
// 旧集計の appointment_trimming_options は DROP 済みで、置換先の reader が未接続。
func (r *trimmingOptionRepository) CountUsageByTrimmingOptionID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, apperrors.WrapTrimmingMasterUsageUnverified()
}
