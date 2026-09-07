package billing

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyntheticClosingRequest は S09 用の新規合成会計 5 件を作る入力。
type SyntheticClosingRequest struct {
	AppEnv             string
	DBHost             string
	TargetDate         time.Time
	ExistingBillingIDs []uint64
}

// SyntheticClosingResult は作成した使い捨て clinic と 5 件の完了時刻を返す。
type SyntheticClosingResult struct {
	ClinicID    uint64
	BillingIDs  []uint64
	CompletedAt []time.Time
}

// CreateSyntheticClosingFixture は新規 clinic / owner / pet / 会計 5 件を作る。
// 既存 billings.id の UPDATE はしない。
func CreateSyntheticClosingFixture(ctx context.Context, db *gorm.DB, req SyntheticClosingRequest) (*SyntheticClosingResult, error) {
	if err := AllowUATSyntheticClosing(req.AppEnv, req.DBHost); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectExistingBillingIDs(req.ExistingBillingIDs); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if db == nil {
		return nil, apperrors.WrapInvalidInput("db is required")
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, apperrors.Wrap(err, "load Asia/Tokyo")
	}
	inJST := req.TargetDate.In(jst)
	day := time.Date(inJST.Year(), inJST.Month(), inJST.Day(), 0, 0, 0, 0, jst)
	completed := []time.Time{
		time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 13, 30, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 14, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 20, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 2, 0, 0, 0, jst).Add(24 * time.Hour),
	}

	clinicID := uint64(900000 + time.Now().UnixNano()%90000 + 1)
	if err := RejectReservedClinicID(clinicID); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}

	company := &model.Company{Name: fmt.Sprintf("s09-synthetic-%d", clinicID)}
	if err := db.WithContext(ctx).Create(company).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic company")
	}
	clinic := &model.Clinic{ID: clinicID, CompanyID: company.ID, Name: fmt.Sprintf("s09-clinic-%d", clinicID), IsActive: true}
	if err := db.WithContext(ctx).Create(clinic).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic clinic")
	}
	settings := &model.ClinicSettings{
		ClinicID:            clinicID,
		ClosingAmStart:      "09:00",
		ClosingAmPmBoundary: "13:30",
		ClosingWeekdayEnd:   "19:00",
		ClosingSundayEnd:    "17:30",
	}
	if err := db.WithContext(ctx).Create(settings).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic closing settings")
	}

	owner := &model.Owner{ClinicID: clinicID, Name: fmt.Sprintf("s09-owner-%d", clinicID)}
	if err := db.WithContext(ctx).Create(owner).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic owner")
	}
	species := &model.AnimalSpecies{Name: fmt.Sprintf("s09-species-%d", clinicID)}
	if err := db.WithContext(ctx).Create(species).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic species")
	}
	pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: fmt.Sprintf("s09-pet-%d", clinicID)}
	if err := db.WithContext(ctx).Create(pet).Error; err != nil {
		return nil, apperrors.Wrap(err, "create synthetic pet")
	}

	ids := make([]uint64, 0, len(completed))
	for _, at := range completed {
		atCopy := at
		billing := &model.Billing{
			ClinicID:      clinicID,
			OwnerID:       &owner.ID,
			PetID:         &pet.ID,
			Subtotal:      1000,
			TaxTotal:      100,
			TotalAmount:   1100,
			Status:        model.BillingStatusCompleted,
			ScheduledDate: day,
			CompletedAt:   &atCopy,
			Memo:          "s09-synthetic",
		}
		if err := db.WithContext(ctx).Create(billing).Error; err != nil {
			return nil, apperrors.Wrap(err, "create synthetic billing")
		}
		ids = append(ids, billing.ID)
	}

	return &SyntheticClosingResult{ClinicID: clinicID, BillingIDs: ids, CompletedAt: completed}, nil
}
