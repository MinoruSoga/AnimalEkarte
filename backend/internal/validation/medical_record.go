package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

var (
	visitTypeRegex = regexp.MustCompile(`^(初診|再診)$`)
	statusRegex    = regexp.MustCompile(`^(作成中|確定済)$`)
)

// ValidateCreateMedicalRecord カルテ作成リクエストのバリデーション
func ValidateCreateMedicalRecord(req *model.CreateMedicalRecordRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	var errors []string

	// PetIDの検証
	if strings.TrimSpace(req.PetID) == "" {
		errors = append(errors, "pet_id is required")
	}

	// OwnerIDの検証
	if strings.TrimSpace(req.OwnerID) == "" {
		errors = append(errors, "owner_id is required")
	}

	// VisitDateの検証
	if strings.TrimSpace(req.VisitDate) == "" {
		errors = append(errors, "visit_date is required")
	} else {
		if _, err := time.Parse("2006-01-02T15:04:05Z07:00", req.VisitDate); err != nil {
			if _, err := time.Parse("2006-01-02 15:04:05", req.VisitDate); err != nil {
				if _, err := time.Parse("2006-01-02", req.VisitDate); err != nil {
					errors = append(errors, "visit_date must be a valid date format (YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or RFC3339)")
				}
			}
		}
	}

	// VisitTypeの検証
	if req.VisitType != "" {
		if !visitTypeRegex.MatchString(req.VisitType) {
			errors = append(errors, "visit_type must be '初診' or '再診'")
		}
	}

	// ChiefComplaintの検証
	if len(req.ChiefComplaint) > 1000 {
		errors = append(errors, "chief_complaint must be less than 1000 characters")
	}

	// Subjectiveの検証
	if len(req.Subjective) > 2000 {
		errors = append(errors, "subjective must be less than 2000 characters")
	}

	// Objectiveの検証
	if len(req.Objective) > 2000 {
		errors = append(errors, "objective must be less than 2000 characters")
	}

	// Assessmentの検証
	if len(req.Assessment) > 2000 {
		errors = append(errors, "assessment must be less than 2000 characters")
	}

	// Planの検証
	if len(req.Plan) > 2000 {
		errors = append(errors, "plan must be less than 2000 characters")
	}

	// Diagnosisの検証
	if len(req.Diagnosis) > 1000 {
		errors = append(errors, "diagnosis must be less than 1000 characters")
	}

	// Treatmentの検証
	if len(req.Treatment) > 2000 {
		errors = append(errors, "treatment must be less than 2000 characters")
	}

	// Prescriptionの検証
	if len(req.Prescription) > 2000 {
		errors = append(errors, "prescription must be less than 2000 characters")
	}

	// Notesの検証
	if len(req.Notes) > 2000 {
		errors = append(errors, "notes must be less than 2000 characters")
	}

	// SurgeryNotesの検証
	if len(req.SurgeryNotes) > 2000 {
		errors = append(errors, "surgery_notes must be less than 2000 characters")
	}

	// Statusの検証
	if req.Status != "" {
		if !statusRegex.MatchString(req.Status) {
			errors = append(errors, "status must be '作成中' or '確定済'")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateUpdateMedicalRecord カルテ更新リクエストのバリデーション
func ValidateUpdateMedicalRecord(req *model.UpdateMedicalRecordRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	var errors []string

	// VisitDateの検証
	if req.VisitDate != nil && *req.VisitDate != "" {
		if _, err := time.Parse("2006-01-02T15:04:05Z07:00", *req.VisitDate); err != nil {
			if _, err := time.Parse("2006-01-02 15:04:05", *req.VisitDate); err != nil {
				if _, err := time.Parse("2006-01-02", *req.VisitDate); err != nil {
					errors = append(errors, "visit_date must be a valid date format (YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, or RFC3339)")
				}
			}
		}
	}

	// VisitTypeの検証
	if req.VisitType != nil && *req.VisitType != "" {
		if !visitTypeRegex.MatchString(*req.VisitType) {
			errors = append(errors, "visit_type must be '初診' or '再診'")
		}
	}

	// ChiefComplaintの検証
	if req.ChiefComplaint != nil && len(*req.ChiefComplaint) > 1000 {
		errors = append(errors, "chief_complaint must be less than 1000 characters")
	}

	// Subjectiveの検証
	if req.Subjective != nil && len(*req.Subjective) > 2000 {
		errors = append(errors, "subjective must be less than 2000 characters")
	}

	// Objectiveの検証
	if req.Objective != nil && len(*req.Objective) > 2000 {
		errors = append(errors, "objective must be less than 2000 characters")
	}

	// Assessmentの検証
	if req.Assessment != nil && len(*req.Assessment) > 2000 {
		errors = append(errors, "assessment must be less than 2000 characters")
	}

	// Planの検証
	if req.Plan != nil && len(*req.Plan) > 2000 {
		errors = append(errors, "plan must be less than 2000 characters")
	}

	// Diagnosisの検証
	if req.Diagnosis != nil && len(*req.Diagnosis) > 1000 {
		errors = append(errors, "diagnosis must be less than 1000 characters")
	}

	// Treatmentの検証
	if req.Treatment != nil && len(*req.Treatment) > 2000 {
		errors = append(errors, "treatment must be less than 2000 characters")
	}

	// Prescriptionの検証
	if req.Prescription != nil && len(*req.Prescription) > 2000 {
		errors = append(errors, "prescription must be less than 2000 characters")
	}

	// Notesの検証
	if req.Notes != nil && len(*req.Notes) > 2000 {
		errors = append(errors, "notes must be less than 2000 characters")
	}

	// SurgeryNotesの検証
	if req.SurgeryNotes != nil && len(*req.SurgeryNotes) > 2000 {
		errors = append(errors, "surgery_notes must be less than 2000 characters")
	}

	// Statusの検証
	if req.Status != nil {
		if !statusRegex.MatchString(*req.Status) {
			errors = append(errors, "status must be '作成中' or '確定済'")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
