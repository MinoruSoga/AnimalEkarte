package clinicale2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	padPetCount            = 20
	medicalRecordCount     = 21
	ownerSearchToken       = "e2e-owner"
	outsideFirstPagePrefix = "e2e-zebra-"
)

// Request は disposable clinic を作る入力。PasswordHash はログに出さない。
type Request struct {
	AppEnv       string
	DBHost       string
	PasswordHash string
}

// Result は Playwright が参照する合成 ID / 氏名。秘密は含めない。
type Result struct {
	ClinicID                uint64 `json:"clinicId"`
	OwnerName               string `json:"ownerName"`
	OwnerSearch             string `json:"ownerSearch"`
	PetID                   uint64 `json:"petId"`
	PetName                 string `json:"petName"`
	OutsideFirstPagePetID   uint64 `json:"outsideFirstPagePetId"`
	OutsideFirstPagePetName string `json:"outsideFirstPagePetName"`
	EstimateTitle           string `json:"estimateTitle"`
	MedicalRecordCount      int    `json:"medicalRecordCount"`
}

// Create は新規 clinic / staff / owner / pet / 確定カルテと allowlist 用の行を INSERT する。
func Create(ctx context.Context, db *gorm.DB, req Request) (*Result, error) {
	if err := Allow(req.AppEnv, req.DBHost); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}
	if strings.TrimSpace(req.PasswordHash) == "" {
		return nil, apperrors.WrapInvalidInput("password hash is required")
	}
	if db == nil {
		return nil, apperrors.WrapInvalidInput("db is required")
	}

	clinicID := clinicIDBase + uint64(time.Now().UnixNano()%8000)
	if err := RejectReservedClinicID(clinicID); err != nil {
		return nil, apperrors.WrapInvalidInput(err.Error())
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, apperrors.Wrap(err, "load Asia/Tokyo")
	}
	day := time.Now().In(jst)
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, jst)

	var result *Result
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		company := &model.Company{Name: fmt.Sprintf("%s%d", companyNamePrefix, clinicID)}
		if err := tx.Create(company).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic company")
		}
		clinic := &model.Clinic{
			ID:        clinicID,
			CompanyID: company.ID,
			Name:      fmt.Sprintf("%s%d", clinicNamePrefix, clinicID),
			IsActive:  true,
		}
		if err := tx.Create(clinic).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic clinic")
		}

		account := &model.Account{
			Email:         LoginEmail(clinicID),
			PasswordHash:  req.PasswordHash,
			IsActive:      true,
			IsSystemAdmin: true,
		}
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic account")
		}
		staff := &model.Staff{
			ClinicID:  clinicID,
			AccountID: &account.ID,
			Name:      fmt.Sprintf("e2e-staff-%d", clinicID),
			IsActive:  true,
			StaffType: model.StaffTypeDoctor,
		}
		if err := tx.Create(staff).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic staff")
		}
		assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
		if err := tx.Create(assignment).Error; err != nil {
			return apperrors.Wrap(err, "assign synthetic staff clinic")
		}

		owner := &model.Owner{
			ClinicID: clinicID,
			Name:     fmt.Sprintf("%s-%d", ownerSearchToken, clinicID),
			NameKana: ownerSearchToken,
		}
		if err := tx.Create(owner).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic owner")
		}
		species := &model.AnimalSpecies{Name: fmt.Sprintf("e2e-species-%d", clinicID)}
		if err := tx.Create(species).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic species")
		}

		mainPet := &model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owner.ID,
			AnimalSpeciesID: species.ID,
			Name:            fmt.Sprintf("e2e-pet-%d", clinicID),
		}
		if err := tx.Create(mainPet).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic pet")
		}
		for i := 0; i < padPetCount; i++ {
			pad := &model.Pet{
				ClinicID:        clinicID,
				OwnerID:         owner.ID,
				AnimalSpeciesID: species.ID,
				Name:            fmt.Sprintf("e2e-a%02d-%d", i, clinicID),
			}
			if err := tx.Create(pad).Error; err != nil {
				return apperrors.Wrap(err, "create pad pet")
			}
		}
		outsidePet := &model.Pet{
			ClinicID:        clinicID,
			OwnerID:         owner.ID,
			AnimalSpeciesID: species.ID,
			Name:            fmt.Sprintf("%s%d", outsideFirstPagePrefix, clinicID),
		}
		if err := tx.Create(outsidePet).Error; err != nil {
			return apperrors.Wrap(err, "create outside-first-page pet")
		}

		for i := 0; i < medicalRecordCount; i++ {
			petID := mainPet.ID
			record := &model.MedicalRecord{
				ClinicID: clinicID,
				RecordNo: fmt.Sprintf("E2E-%d-%02d", clinicID, i+1),
				Date:     day,
				OwnerID:  &owner.ID,
				PetID:    &petID,
				DoctorID: &staff.ID,
				Status:   model.MedicalRecordStatusFinalized,
			}
			if err := tx.Create(record).Error; err != nil {
				return apperrors.Wrap(err, "create synthetic medical record")
			}
		}

		var firstRecord model.MedicalRecord
		if err := tx.Where("clinic_id = ?", clinicID).Order("id ASC").First(&firstRecord).Error; err != nil {
			return apperrors.Wrap(err, "load first synthetic medical record")
		}

		examType := &model.ExaminationType{ClinicID: clinicID, Name: fmt.Sprintf("e2e-exam-%d", clinicID), IsActive: true}
		if err := tx.Create(examType).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic exam type")
		}
		exam := &model.Examination{
			ClinicID:        clinicID,
			MedicalRecordID: &firstRecord.ID,
			PetID:           &mainPet.ID,
			ExamTypeID:      examType.ID,
			DoctorID:        &staff.ID,
			Date:            day,
			Status:          model.ExaminationStatusCompleted,
		}
		if err := tx.Create(exam).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic examination")
		}

		vaccine := &model.Vaccine{ClinicID: clinicID, Name: fmt.Sprintf("e2e-vac-%d", clinicID), IsActive: true}
		if err := tx.Create(vaccine).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic vaccine")
		}
		vaccination := &model.Vaccination{
			ClinicID:        clinicID,
			MedicalRecordID: &firstRecord.ID,
			PetID:           &mainPet.ID,
			VaccineID:       vaccine.ID,
			Date:            day,
			DoctorID:        &staff.ID,
		}
		if err := tx.Create(vaccination).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic vaccination")
		}

		checkupType := &model.CheckupType{ClinicID: clinicID, Name: fmt.Sprintf("e2e-chk-%d", clinicID), IsActive: true}
		if err := tx.Create(checkupType).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic checkup type")
		}
		checkup := &model.Checkup{
			ClinicID:        clinicID,
			MedicalRecordID: firstRecord.ID,
			PetID:           &mainPet.ID,
			CheckupTypeID:   checkupType.ID,
			Date:            day,
			DoctorID:        &staff.ID,
		}
		if err := tx.Create(checkup).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic checkup")
		}

		hospitalization := &model.Hospitalization{
			ClinicID:            clinicID,
			OwnerID:             owner.ID,
			PetID:               mainPet.ID,
			HospitalizationType: model.HospitalizationTypeInpatient,
			StartDate:           day,
			EndDate:             day.Add(24 * time.Hour),
			Status:              model.HospitalizationStatusAdmitted,
			DoctorID:            &staff.ID,
		}
		if err := tx.Create(hospitalization).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic hospitalization")
		}

		estimateTitle := fmt.Sprintf("e2e-est-%d", clinicID)
		estimate := &model.Estimate{
			ClinicID:        clinicID,
			EstimateNo:      fmt.Sprintf("E2-%d", clinicID),
			MedicalRecordID: &firstRecord.ID,
			Title:           estimateTitle,
			OwnerID:         &owner.ID,
			PetID:           &mainPet.ID,
			Status:          model.EstimateStatusDraft,
			CreatedBy:       &staff.ID,
		}
		if err := tx.Create(estimate).Error; err != nil {
			return apperrors.Wrap(err, "create synthetic estimate")
		}

		result = &Result{
			ClinicID:                clinicID,
			OwnerName:               owner.Name,
			OwnerSearch:             ownerSearchToken,
			PetID:                   mainPet.ID,
			PetName:                 mainPet.Name,
			OutsideFirstPagePetID:   outsidePet.ID,
			OutsideFirstPagePetName: outsidePet.Name,
			EstimateTitle:           estimateTitle,
			MedicalRecordCount:      medicalRecordCount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Delete は合成 clinic とその子孫だけを消す。clinic 1/2 と接頭辞不一致は拒否する。
func Delete(ctx context.Context, db *gorm.DB, appEnv, dbHost string, clinicID uint64) error {
	if err := Allow(appEnv, dbHost); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if err := RejectReservedClinicID(clinicID); err != nil {
		return apperrors.WrapInvalidInput(err.Error())
	}
	if db == nil {
		return apperrors.WrapInvalidInput("db is required")
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var clinic model.Clinic
		if err := tx.First(&clinic, clinicID).Error; err != nil {
			return apperrors.Wrap(err, "load synthetic clinic")
		}
		if !strings.HasPrefix(clinic.Name, clinicNamePrefix) {
			return apperrors.WrapInvalidInput("clinic name is not a clinical e2e fixture")
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

		var estimates []model.Estimate
		if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Find(&estimates).Error; err != nil {
			return apperrors.Wrap(err, "list synthetic estimates")
		}
		estimateIDs := make([]uint64, 0, len(estimates))
		for _, estimate := range estimates {
			estimateIDs = append(estimateIDs, estimate.ID)
		}

		if len(estimateIDs) > 0 {
			if err := tx.Unscoped().Where("estimate_id IN ?", estimateIDs).Delete(&model.EstimateItem{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic estimate items")
			}
		}
		scoped := []any{
			&model.Estimate{},
			&model.Checkup{},
			&model.Vaccination{},
			&model.Examination{},
			&model.Hospitalization{},
			&model.MedicalRecord{},
			&model.Pet{},
			&model.Owner{},
			&model.Vaccine{},
			&model.ExaminationType{},
			&model.CheckupType{},
			&model.StaffClinicAssignment{},
			&model.Staff{},
		}
		for _, modelPtr := range scoped {
			if err := tx.Unscoped().Where("clinic_id = ?", clinicID).Delete(modelPtr).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic clinic-scoped row")
			}
		}
		if len(accountIDs) > 0 {
			if err := tx.Unscoped().Where("id IN ?", accountIDs).Delete(&model.Account{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic accounts")
			}
		}
		if err := tx.Unscoped().Where("name LIKE ?", fmt.Sprintf("e2e-species-%d", clinicID)).Delete(&model.AnimalSpecies{}).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic species")
		}
		if err := tx.Delete(&clinic).Error; err != nil {
			return apperrors.Wrap(err, "delete synthetic clinic")
		}
		if clinic.CompanyID != 0 {
			if err := tx.Where("id = ? AND name LIKE ?", clinic.CompanyID, companyNamePrefix+"%").Delete(&model.Company{}).Error; err != nil {
				return apperrors.Wrap(err, "delete synthetic company")
			}
		}
		return nil
	})
}

// EncodeResult は stdout 用の 1 行 JSON。秘密は載らない。
func EncodeResult(result *Result) ([]byte, error) {
	if result == nil {
		return nil, apperrors.WrapInvalidInput("result is required")
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, apperrors.Wrap(err, "encode clinical fixture")
	}
	return payload, nil
}
