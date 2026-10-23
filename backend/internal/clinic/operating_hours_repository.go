package clinic

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type operatingHoursRepository struct{ db *gorm.DB }

// NewOperatingHoursRepository constructs the clinic_operating_hours data port.
func NewOperatingHoursRepository(db *gorm.DB) OperatingHoursRepository {
	return &operatingHoursRepository{db: db}
}

// FindByClinicID は clinic_id の共通医院時間行を返す。行なしは found=false で、エラーでも補完でもない。
// DB エラーはそのまま伝播する（呼出側が未設定や空成功へ変換しない）。
func (r *operatingHoursRepository) FindByClinicID(ctx context.Context, clinicID uint64) (model.ClinicOperatingHours, bool, error) {
	var rows []model.ClinicOperatingHours
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Limit(1).
		Find(&rows).Error
	if err != nil {
		return model.ClinicOperatingHours{}, false, apperrors.FromGORM(err, "clinic_operating_hours", strconv.FormatUint(clinicID, 10))
	}
	if len(rows) == 0 {
		return model.ClinicOperatingHours{}, false, nil
	}
	return rows[0], true, nil
}
