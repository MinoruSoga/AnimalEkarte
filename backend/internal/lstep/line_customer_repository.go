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
	// EMR-504: line_customers.owner_id は DROP 済み。飼主は owner_line_binding_states（state='linked'）から
	// attachLinkedOwners で読む（gorm:"-" の Owner は Preload できない）。
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("created_at DESC").
		Limit(lineCustomerListMax).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_customer", "")
	}
	if err := r.attachLinkedOwners(ctx, clinicID, items, false); err != nil {
		return nil, err
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
	// EMR-504: 飼主は binding（state='linked'）から clinic スコープで読み込み、customer.Owner に載せる
	// (GetLiffProfile / GetHealthCard のクロステナント露出防止)。Pet.AnimalSpecies は gorm:"-" のため
	// Preload しない（種別名の再配線は EMR-491）。
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_customer", fmt.Sprintf("%d", id))
	}
	customers := []model.LineCustomer{c}
	if err := r.attachLinkedOwners(ctx, clinicID, customers, true); err != nil {
		return nil, err
	}
	return &customers[0], nil
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

// UpdateOwnerLink は EMR-504 / EMR-549 C1 で固定 422 にした。line_customers.owner_id は DROP 済みで、
// 置換先 owner_line_binding_states の writer（EMR-469 W09）は未接続。旧列へは書かず、成功も装わない。
func (r *lineCustomerRepository) UpdateOwnerLink(_ context.Context, _, _ uint64, _ *uint64) error {
	return apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineBindingPending,
		"LINE顧客と飼い主の紐付けは現在変更できません（紐付け機能の準備中です）",
	)
}

// attachLinkedOwners は owner_line_binding_states（state='linked'）が指す飼主を各顧客の Owner に設定する。
// binding の無い顧客（未連携・unlink_pending・legacy_unknown）は Owner=nil のまま返す。
// 飼主は同一医院の未削除行だけを読み、withPets のときは同一医院の未削除ペットも Preload する。
func (r *lineCustomerRepository) attachLinkedOwners(ctx context.Context, clinicID uint64, customers []model.LineCustomer, withPets bool) error {
	if len(customers) == 0 {
		return nil
	}
	customerIDs := make([]uint64, len(customers))
	for i := range customers {
		customerIDs[i] = customers[i].ID
	}
	var bindings []model.OwnerLineBindingState
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND line_customer_id IN ? AND state = ?", clinicID, customerIDs, persistence.LineBindingStateLinked).
		Find(&bindings).Error; err != nil {
		return apperrors.FromGORM(err, "line_customer", "binding")
	}
	ownerIDByCustomer := make(map[uint64]uint64, len(bindings))
	ownerIDs := make([]uint64, 0, len(bindings))
	for i := range bindings {
		if bindings[i].LineCustomerID == nil {
			continue
		}
		ownerIDByCustomer[*bindings[i].LineCustomerID] = bindings[i].OwnerID
		ownerIDs = append(ownerIDs, bindings[i].OwnerID)
	}
	if len(ownerIDs) == 0 {
		return nil
	}
	query := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID))
	if withPets {
		query = query.Preload("Pets", "clinic_id = ? AND deleted_at IS NULL", clinicID)
	}
	var owners []model.Owner
	if err := query.Where("id IN ? AND deleted_at IS NULL", ownerIDs).Find(&owners).Error; err != nil {
		return apperrors.FromGORM(err, "owner", "")
	}
	ownersByID := make(map[uint64]*model.Owner, len(owners))
	for i := range owners {
		ownersByID[owners[i].ID] = &owners[i]
	}
	for i := range customers {
		if ownerID, ok := ownerIDByCustomer[customers[i].ID]; ok {
			customers[i].Owner = ownersByID[ownerID]
		}
	}
	return nil
}
