package lstep

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LineCustomerRepository は予約顧客のデータアクセスインターフェース
type LineCustomerRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error)
	// CountAll returns the unscoped (by safety cap) total for clinic-scoped line customers (G2F-05).
	CountAll(ctx context.Context, clinicID uint64) (int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
	FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error)
	UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error
}

type lineCustomerRepository struct{ db *gorm.DB }

func NewLineCustomerRepository(db *gorm.DB) LineCustomerRepository {
	return &lineCustomerRepository{db: db}
}

// lineCustomerListMax is a hard safety cap for List (G2F-05).
// Truncation is surfaced to clients via LineCustomerListResult.Truncated + Total.
const lineCustomerListMax = 200

func (r *lineCustomerRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error) {
	items := make([]model.LineCustomer, 0)
	// clinic_id 述語必須: LineCustomerService.LinkOwner が service 層で ownerID の所属クリニックを
	// 検証する（FE-refactor.md 残件 3 対応）が、多層防御として読み取り側でも Owner を
	// 自クリニックにスコープしフォールバックさせる。
	err := r.db.WithContext(ctx).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("created_at DESC").
		Limit(lineCustomerListMax).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_customer", "")
	}
	return items, nil
}

func (r *lineCustomerRepository) CountAll(ctx context.Context, clinicID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.LineCustomer{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Count(&total).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "line_customer", "count")
	}
	return total, nil
}

func (r *lineCustomerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
	var c model.LineCustomer
	// clinic_id 述語必須: LineCustomerService.LinkOwner が service 層で ownerID の所属クリニックを
	// 検証する（FE-refactor.md 残件 3 対応）が、多層防御として読み取り側でも Owner を
	// 自クリニックにスコープしフォールバックさせる
	// (GetLiffProfile / GetHealthCard のクロステナント露出防止)。
	err := r.db.WithContext(ctx).
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Owner.Pets", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Owner.Pets.AnimalSpecies").
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_customer", fmt.Sprintf("%d", id))
	}
	return &c, nil
}

func (r *lineCustomerRepository) FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error) {
	var c model.LineCustomer
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("line_user_id = ?", lineUserID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c = model.LineCustomer{
			ClinicID:    clinicID,
			LineUserID:  lineUserID,
			DisplayName: displayName,
		}
		if err2 := r.db.WithContext(ctx).Create(&c).Error; err2 != nil {
			return nil, apperrors.FromGORM(err2, "line_customer", "")
		}
		return &c, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_customer", lineUserID)
	}
	return &c, nil
}

func (r *lineCustomerRepository) UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error {
	result := r.db.WithContext(ctx).
		Model(&model.LineCustomer{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Update("additional_fields", fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "line_customer", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *lineCustomerRepository) UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.LineCustomer{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Update("owner_id", ownerID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "line_customer", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("line_customer", fmt.Sprintf("%d", id))
	}
	return nil
}
