package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

var (
	statusRegex = regexp.MustCompile(`^(作成中|確定済)$`)
)

// ValidateCreateMedicalRecord カルテ作成リクエストのバリデーション
func ValidateCreateMedicalRecord(req *model.CreateMedicalRecordRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	var errs []string

	// Dateの検証
	if strings.TrimSpace(req.Date) == "" {
		errs = append(errs, "date is required")
	} else {
		if _, err := time.Parse("2006-01-02", req.Date); err != nil {
			errs = append(errs, "date must be a valid date format (YYYY-MM-DD)")
		}
	}

	// OwnerNameの検証
	if strings.TrimSpace(req.OwnerName) == "" {
		errs = append(errs, "owner_name is required")
	}

	// PetNameの検証
	if strings.TrimSpace(req.PetName) == "" {
		errs = append(errs, "pet_name is required")
	}

	// Doctorの検証
	if strings.TrimSpace(req.Doctor) == "" {
		errs = append(errs, "doctor is required")
	}

	// ChiefComplaintの検証
	if len(req.ChiefComplaint) > 1000 {
		errs = append(errs, "chief_complaint must be less than 1000 characters")
	}

	// TreatmentPolicyの検証
	if len(req.TreatmentPolicy) > 2000 {
		errs = append(errs, "treatment_policy must be less than 2000 characters")
	}

	// PhysicalExamの検証
	if len(req.PhysicalExam) > 2000 {
		errs = append(errs, "physical_exam must be less than 2000 characters")
	}

	// DiagnosisDetailsの検証
	if len(req.DiagnosisDetails) > 2000 {
		errs = append(errs, "diagnosis_details must be less than 2000 characters")
	}

	// Statusの検証
	if req.Status != "" {
		if !statusRegex.MatchString(req.Status) {
			errs = append(errs, "status must be '作成中' or '確定済'")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// ValidateUpdateMedicalRecord カルテ更新リクエストのバリデーション
func ValidateUpdateMedicalRecord(req *model.UpdateMedicalRecordRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	var errs []string

	// Dateの検証
	if req.Date != nil && *req.Date != "" {
		if _, err := time.Parse("2006-01-02", *req.Date); err != nil {
			errs = append(errs, "date must be a valid date format (YYYY-MM-DD)")
		}
	}

	// OwnerNameの検証
	if req.OwnerName != nil && strings.TrimSpace(*req.OwnerName) == "" {
		errs = append(errs, "owner_name must not be empty")
	}

	// PetNameの検証
	if req.PetName != nil && strings.TrimSpace(*req.PetName) == "" {
		errs = append(errs, "pet_name must not be empty")
	}

	// Doctorの検証
	if req.Doctor != nil && strings.TrimSpace(*req.Doctor) == "" {
		errs = append(errs, "doctor must not be empty")
	}

	// ChiefComplaintの検証
	if req.ChiefComplaint != nil && len(*req.ChiefComplaint) > 1000 {
		errs = append(errs, "chief_complaint must be less than 1000 characters")
	}

	// TreatmentPolicyの検証
	if req.TreatmentPolicy != nil && len(*req.TreatmentPolicy) > 2000 {
		errs = append(errs, "treatment_policy must be less than 2000 characters")
	}

	// PhysicalExamの検証
	if req.PhysicalExam != nil && len(*req.PhysicalExam) > 2000 {
		errs = append(errs, "physical_exam must be less than 2000 characters")
	}

	// DiagnosisDetailsの検証
	if req.DiagnosisDetails != nil && len(*req.DiagnosisDetails) > 2000 {
		errs = append(errs, "diagnosis_details must be less than 2000 characters")
	}

	// Statusの検証
	if req.Status != nil && *req.Status != "" {
		if !statusRegex.MatchString(*req.Status) {
			errs = append(errs, "status must be '作成中' or '確定済'")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}
