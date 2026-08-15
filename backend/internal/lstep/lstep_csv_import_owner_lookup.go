package lstep

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CSVImportOwnerLookup はCSV内の候補IDに限定してOwnerの存在を照合する。
type CSVImportOwnerLookup struct{}

// NewLstepCSVImportOwnerLookup は CSV 内の LINE User ID だけを照合する軽量 lookup を返す。
func NewLstepCSVImportOwnerLookup() *CSVImportOwnerLookup {
	return &CSVImportOwnerLookup{}
}

func (r *CSVImportOwnerLookup) FindExistingLineUserIDs(
	ctx context.Context,
	db *gorm.DB,
	clinicID uint64,
	lineUserIDs []string,
) (map[string]struct{}, error) {
	matched := make(map[string]struct{}, len(lineUserIDs))
	if len(lineUserIDs) == 0 {
		return matched, nil
	}

	var existing []string
	if err := db.WithContext(ctx).
		Model(&model.Owner{}).
		Where("clinic_id = ?", clinicID).
		Where("line_user_id IN ?", lineUserIDs).
		Order("id ASC").
		Pluck("line_user_id", &existing).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", "line_user_id_batch")
	}
	for _, lineUserID := range existing {
		matched[lineUserID] = struct{}{}
	}
	return matched, nil
}
