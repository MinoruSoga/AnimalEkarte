package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
)

const (
	syntheticClosingSubtotal        = int64(1000)
	syntheticClosingTax             = int64(100)
	syntheticClosingTotal           = int64(1100)
	syntheticClosingCashSystemKey   = "cash"
	syntheticClosingCashDisplayName = "現金"
)

// SyntheticClosingRequest は S09 用の新規合成会計 5 件を作る入力。
type SyntheticClosingRequest struct {
	AppEnv             string
	DBHost             string
	TargetDate         time.Time
	PasswordHash       string
	ExistingBillingIDs []uint64
}

// SyntheticClosingResult は作成した使い捨て clinic と 5 件の完了時刻を返す。
type SyntheticClosingResult struct {
	ClinicID     uint64
	LoginEmail   string
	BillingIDs   []uint64
	CompletedAt  []time.Time
	CleanupToken string
}

// CreateSyntheticClosingFixture は新規 clinic / staff / 支払方法 / 明細 / 会計 5 件を原子的に作る。
// 既存 billings.id の UPDATE はしない。
func CreateSyntheticClosingFixture(ctx context.Context, db *gorm.DB, req SyntheticClosingRequest) (*SyntheticClosingResult, error) {
	if err := AllowUATSyntheticClosing(req.AppEnv, req.DBHost); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectExistingBillingIDs(req.ExistingBillingIDs); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if strings.TrimSpace(req.PasswordHash) == "" {
		return nil, apperrors.WrapInvalidInput("password hash is required")
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
	if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		return nil, apperrors.WrapInvalidInput("target date must be a weekday")
	}
	completed := []time.Time{
		time.Date(day.Year(), day.Month(), day.Day(), 10, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 13, 30, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 14, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 20, 0, 0, 0, jst),
		time.Date(day.Year(), day.Month(), day.Day(), 2, 0, 0, 0, jst).Add(24 * time.Hour),
	}

	var result *SyntheticClosingResult
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		clinicID := uint64(920000 + time.Now().UnixNano()%8000 + 1)
		if err := RejectReservedClinicID(clinicID); err != nil {
			return apperrors.WrapInvalidInput(err.Error())
		}

		company := &model.Company{Name: fmt.Sprintf("%s%d", syntheticClosingCompanyPrefix, clinicID)}
		if err := tx.Create(company).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic company")
		}
		clinic := &model.Clinic{
			ID:        clinicID,
			CompanyID: company.ID,
			Name:      fmt.Sprintf("%s%d", syntheticClosingClinicPrefix, clinicID),
			IsActive:  true,
		}
		if err := tx.Create(clinic).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic clinic")
		}
		settings := &model.ClinicSettings{
			ClinicID:            clinicID,
			ClosingAmStart:      "09:00",
			ClosingAmPmBoundary: "13:30",
			ClosingWeekdayEnd:   "19:00",
			ClosingSundayEnd:    "17:30",
		}
		if err := tx.Create(settings).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic closing settings")
		}

		account := &model.Account{
			Email:         SyntheticClosingLoginEmail(clinicID),
			PasswordHash:  req.PasswordHash,
			IsActive:      true,
			IsSystemAdmin: true,
		}
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic account")
		}
		staffRow := &model.Staff{
			ClinicID:  clinicID,
			AccountID: &account.ID,
			Name:      fmt.Sprintf("s09-staff-%d", clinicID),
			IsActive:  true,
			StaffType: model.StaffTypeDoctor,
		}
		if err := staff.CreateSyntheticClosingStaff(ctx, tx, staffRow); err != nil {
			return apperrors.Wrap(err, "create synthetic staff")
		}
		assignment := &model.StaffClinicAssignment{StaffID: staffRow.ID, ClinicID: clinicID, IsMain: true}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.Wrap(err, "assign synthetic staff clinic")
		}

		owner := &model.Owner{ClinicID: clinicID, Name: fmt.Sprintf("s09-owner-%d", clinicID)}
		if err := tx.Create(owner).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic owner")
		}
		species := &model.AnimalSpecies{Name: fmt.Sprintf("s09-species-%d", clinicID)}
		if err := tx.Create(species).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic species")
		}
		pet := &model.Pet{ClinicID: clinicID, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: fmt.Sprintf("s09-pet-%d", clinicID)}
		if err := tx.Create(pet).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic pet")
		}

		paymentMethod, err := resolveSyntheticClosingCashPaymentMethod(tx, clinicID)
		if err != nil {
			return err
		}

		ids := make([]uint64, 0, len(completed))
		for _, at := range completed {
			atCopy := at
			clinicIDCopy := clinicID
			billing := &model.Billing{
				ClinicID:      clinicID,
				OwnerID:       &owner.ID,
				PetID:         &pet.ID,
				Subtotal:      syntheticClosingSubtotal,
				TaxTotal:      syntheticClosingTax,
				TotalAmount:   syntheticClosingTotal,
				Status:        model.BillingStatusCompleted,
				ScheduledDate: day,
				CompletedAt:   &atCopy,
				Memo:          "s09-synthetic",
			}
			if err := tx.Create(billing).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic billing")
			}
			item := &model.BillingItem{
				BillingID: billing.ID,
				ClinicID:  &clinicIDCopy,
				Category:  model.ItemCategoryExamination,
				Name:      "s09-item",
				UnitPrice: syntheticClosingSubtotal,
				Quantity:  1,
				TaxType:   model.TaxTypeExcluded,
				TaxRate:   0.10,
				Source:    model.ItemSourceManual,
			}
			if err := tx.Create(item).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic billing item")
			}
			payment := &model.Payment{
				BillingID:       billing.ID,
				ClinicID:        clinicID,
				Subtotal:        syntheticClosingSubtotal,
				TaxTotal:        syntheticClosingTax,
				TotalAmount:     syntheticClosingTotal,
				BillingAmount:   syntheticClosingTotal,
				ReceivedAmount:  syntheticClosingTotal,
				Method:          model.PaymentMethodCash,
				PaymentMethodID: &paymentMethod.ID,
				PaidBy:          &staffRow.ID,
			}
			if err := tx.Create(payment).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic payment")
			}
			split := &model.PaymentSplit{
				ClinicID:        clinicID,
				BillingID:       billing.ID,
				Method:          model.PaymentMethodCash,
				PaymentMethodID: &paymentMethod.ID,
				Amount:          syntheticClosingTotal,
				ReceivedAmount:  syntheticClosingTotal,
				PaidBy:          &staffRow.ID,
			}
			if err := tx.Create(split).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic payment split")
			}
			ids = append(ids, billing.ID)
		}

		result = &SyntheticClosingResult{
			ClinicID:     clinicID,
			LoginEmail:   SyntheticClosingLoginEmail(clinicID),
			BillingIDs:   ids,
			CompletedAt:  completed,
			CleanupToken: SyntheticClosingCleanupToken(clinicID),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// resolveSyntheticClosingCashPaymentMethod は clinic INSERT の
// trg_create_default_payment_methods が既に入れた cash 行を再利用する。
// testdb は trigger を載せないので、無いときだけ INSERT する。
func resolveSyntheticClosingCashPaymentMethod(tx *gorm.DB, clinicID uint64) (*model.PaymentMethodMaster, error) {
	var existing model.PaymentMethodMaster
	err := tx.Where("clinic_id = ? AND system_key = ?", clinicID, syntheticClosingCashSystemKey).
		Take(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.Wrap(err, "load synthetic cash payment method")
	}

	cashKey := syntheticClosingCashSystemKey
	created := &model.PaymentMethodMaster{
		ClinicID:     clinicID,
		Name:         syntheticClosingCashDisplayName,
		SystemKey:    &cashKey,
		DisplayOrder: 1,
		IsActive:     true,
	}
	if err := tx.Create(created).Error; err != nil {
		var raced model.PaymentMethodMaster
		if loadErr := tx.Where("clinic_id = ? AND system_key = ?", clinicID, syntheticClosingCashSystemKey).
			Take(&raced).Error; loadErr == nil {
			return &raced, nil
		}
		return nil, apperrors.Wrap(err, "create synthetic payment method")
	}
	return created, nil
}

// DeleteSyntheticClosingFixture は合成 clinic とその子孫だけを消す。clinic 1/2 と接頭辞不一致は拒否する。
func DeleteSyntheticClosingFixture(ctx context.Context, db *gorm.DB, appEnv, dbHost string, clinicID uint64, cleanupToken string) error {
	if err := AllowUATSyntheticClosing(appEnv, dbHost); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectReservedClinicID(clinicID); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if !MatchSyntheticClosingCleanupToken(clinicID, cleanupToken) {
		return apperrors.WrapInvalidInput("cleanup token is invalid")
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var clinic model.Clinic
		if err := tx.First(&clinic, clinicID).Error; err != nil {
			return apperrors.Wrap(err, "load synthetic clinic")
		}
		if !strings.HasPrefix(clinic.Name, syntheticClosingClinicPrefix) {
			return apperrors.WrapInvalidInput("clinic name is not a synthetic closing fixture")
		}

		var staffs []model.Staff
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&staffs).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic staff")
		}
		accountIDs := make([]uint64, 0, len(staffs))
		for _, staff := range staffs {
			if staff.AccountID != nil {
				accountIDs = append(accountIDs, *staff.AccountID)
			}
		}

		scoped := []any{
			&model.PaymentSplit{},
			&model.Payment{},
			&model.BillingItem{},
			&model.Billing{},
			&model.PaymentMethodMaster{},
			&model.ClinicSettings{},
			&model.Pet{},
			&model.Owner{},
			&model.StaffClinicAssignment{},
		}
		for _, modelPtr := range scoped {
			if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Delete(modelPtr).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic clinic-scoped row")
			}
		}
		if err := staff.UnscopedDeleteSyntheticClosingStaffs(ctx, tx, clinicID); err != nil {
			return apperrors.Wrap(err, "delete synthetic staff")
		}
		if len(accountIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", accountIDs).Delete(&model.Account{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic accounts")
			}
		}
		if err := tx.Unscoped().Where("name LIKE ?", fmt.Sprintf("s09-species-%d", clinicID)).Delete(&model.AnimalSpecies{}).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic species")
		}
		if err := tx.Delete(&clinic).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic clinic")
		}
		if clinic.CompanyID != 0 {
			if err := tx.Where("id = ? AND name LIKE ?", clinic.CompanyID, syntheticClosingCompanyPrefix+"%").Delete(&model.Company{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic company")
			}
		}
		return nil
	})
}
