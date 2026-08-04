package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const examinationPrintDraftWatermark = "DRAFT / 未確定"

// ExaminationPrintBoundary classifies print output for clinic vs owner-facing use.
type ExaminationPrintBoundary string

const (
	ExaminationPrintBoundaryOfficial ExaminationPrintBoundary = "official"
	ExaminationPrintBoundaryDraft    ExaminationPrintBoundary = "draft"
)

// ExaminationPrintItem is one immutable item row from a revision snapshot.
// IsAssessed is the stored revision flag — never recomputed for print.
type ExaminationPrintItem struct {
	ID              uint64
	ExamTypeFieldID *uint64
	Name            string
	InspectionValue string
	NormalValue     string
	Result          string
	Unit            string
	ReferenceValue  string
	RefMin          *float64
	RefMax          *float64
	QualitativeMin  *string
	QualitativeMax  *string
	IsAssessed      bool
	IsAbnormal      bool
	Status          model.ExaminationResultStatus
	SortOrder       int
}

// ExaminationPrintSnapshot is a single atomic revision DTO for print/PDF surfaces.
// It is built only from examination_revisions + examination_revision_items.
type ExaminationPrintSnapshot struct {
	ExaminationID   uint64
	ClinicID        uint64
	Version         uint64
	Kind            model.ExaminationRevisionKind
	Status          model.ExaminationStatus
	PrintBoundary   ExaminationPrintBoundary
	Watermark       string
	Date            time.Time
	ResultSummary   string
	Machine         string
	ExamTypeID      uint64
	MedicalRecordID *uint64
	PetID           *uint64
	DoctorID        *uint64
	Display         model.ExaminationDisplaySnapshot
	Items           []ExaminationPrintItem
}

// FindPrintSnapshot loads one clinic-scoped revision (and its items) for print.
// When version is nil, the parent's current_revision_version is used; if that is
// also nil the call fails closed (legacy rows without revisions are not printable).
func (r *examinationRepository) FindPrintSnapshot(
	ctx context.Context,
	clinicID, examinationID uint64,
	version *uint64,
) (*ExaminationPrintSnapshot, error) {
	db := persistence.DBOrTx(ctx, r.db)

	resolvedVersion := version
	if resolvedVersion == nil {
		var parent model.Examination
		if err := db.
			Select("id", "clinic_id", "current_revision_version").
			Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, examinationID).
			First(&parent).Error; err != nil {
			return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examinationID))
		}
		if parent.CurrentRevisionVersion == nil {
			return nil, apperrors.WrapNotFound("examination_print_snapshot", fmt.Sprintf("%d", examinationID))
		}
		resolvedVersion = parent.CurrentRevisionVersion
	}

	var revision model.ExaminationRevision
	if err := db.
		Where(
			"clinic_id = ? AND examination_id = ? AND version = ?",
			clinicID,
			examinationID,
			*resolvedVersion,
		).
		First(&revision).Error; err != nil {
		return nil, apperrors.FromGORM(err, "examination_revision", fmt.Sprintf("examination_id=%d version=%d", examinationID, *resolvedVersion))
	}

	items := make([]model.ExaminationRevisionItem, 0)
	if err := db.
		Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicID, examinationID, revision.Version).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "examination_revision_item", fmt.Sprintf("examination_id=%d version=%d", examinationID, revision.Version))
	}

	var display model.ExaminationDisplaySnapshot
	if err := json.Unmarshal(revision.DisplaySnapshot, &display); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode examination display snapshot")
	}

	printItems := make([]ExaminationPrintItem, 0, len(items))
	for i := range items {
		item := items[i]
		printItems = append(printItems, ExaminationPrintItem{
			ID:              item.ID,
			ExamTypeFieldID: cloneOptionalUint64(item.ExamTypeFieldID),
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Result:          item.Result,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          cloneOptionalFloat64(item.RefMin),
			RefMax:          cloneOptionalFloat64(item.RefMax),
			QualitativeMin:  cloneOptionalString(item.QualitativeMin),
			QualitativeMax:  cloneOptionalString(item.QualitativeMax),
			IsAssessed:      item.IsAssessed,
			IsAbnormal:      item.IsAbnormal,
			Status:          item.Status,
			SortOrder:       item.SortOrder,
		})
	}

	boundary := ExaminationPrintBoundaryDraft
	watermark := examinationPrintDraftWatermark
	if revision.Kind == model.ExaminationRevisionKindOfficial && revision.Status == model.ExaminationStatusConfirmed {
		boundary = ExaminationPrintBoundaryOfficial
		watermark = ""
	}

	return &ExaminationPrintSnapshot{
		ExaminationID:   revision.ExaminationID,
		ClinicID:        revision.ClinicID,
		Version:         revision.Version,
		Kind:            revision.Kind,
		Status:          revision.Status,
		PrintBoundary:   boundary,
		Watermark:       watermark,
		Date:            revision.Date,
		ResultSummary:   revision.ResultSummary,
		Machine:         revision.Machine,
		ExamTypeID:      revision.ExamTypeID,
		MedicalRecordID: cloneOptionalUint64(revision.MedicalRecordID),
		PetID:           cloneOptionalUint64(revision.PetID),
		DoctorID:        cloneOptionalUint64(revision.DoctorID),
		Display:         display,
		Items:           printItems,
	}, nil
}

// GetPrintSnapshot returns a clinic-scoped atomic revision print DTO.
func (s *examinationService) GetPrintSnapshot(
	ctx context.Context,
	clinicID, examinationID uint64,
	version *uint64,
) (*ExaminationPrintSnapshot, error) {
	if s.revisions == nil {
		return nil, apperrors.WrapInternalServerError("examination revision repository capability is required")
	}
	snapshot, err := s.revisions.FindPrintSnapshot(ctx, clinicID, examinationID, version)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read examination print snapshot", "error", err)
		return nil, apperrors.Wrap(err, "failed to read examination print snapshot")
	}
	return snapshot, nil
}

type examinationPrintItemResponse struct {
	ID              uint64   `json:"id"`
	ExamTypeFieldID *uint64  `json:"exam_type_field_id,omitempty"`
	Name            string   `json:"name"`
	InspectionValue string   `json:"inspection_value"`
	NormalValue     string   `json:"normal_value"`
	Result          string   `json:"result"`
	Unit            string   `json:"unit"`
	ReferenceValue  string   `json:"reference_value"`
	RefMin          *float64 `json:"ref_min,omitempty"`
	RefMax          *float64 `json:"ref_max,omitempty"`
	QualitativeMin  *string  `json:"qualitative_min,omitempty"`
	QualitativeMax  *string  `json:"qualitative_max,omitempty"`
	IsAssessed      bool     `json:"is_assessed"`
	IsAbnormal      bool     `json:"is_abnormal"`
	Status          string   `json:"status"`
	SortOrder       int      `json:"sort_order"`
}

type examinationPrintDisplayResponse struct {
	MedicalRecordNo        string `json:"medical_record_no"`
	PetName                string `json:"pet_name"`
	MedicalRecordOwnerName string `json:"medical_record_owner_name"`
	PetOwnerName           string `json:"pet_owner_name"`
	SpeciesName            string `json:"species_name"`
	ExamTypeName           string `json:"exam_type_name"`
	DoctorName             string `json:"doctor_name"`
}

type examinationPrintSnapshotResponse struct {
	ExaminationID   uint64                           `json:"examination_id"`
	ClinicID        uint64                           `json:"clinic_id"`
	Version         uint64                           `json:"version"`
	Kind            string                           `json:"kind"`
	Status          string                           `json:"status"`
	PrintBoundary   string                           `json:"print_boundary"`
	Watermark       string                           `json:"watermark,omitempty"`
	Date            time.Time                        `json:"date"`
	ResultSummary   string                           `json:"result_summary"`
	Machine         string                           `json:"machine"`
	ExamTypeID      uint64                           `json:"exam_type_id"`
	MedicalRecordID *uint64                          `json:"medical_record_id,omitempty"`
	PetID           *uint64                          `json:"pet_id,omitempty"`
	DoctorID        *uint64                          `json:"doctor_id,omitempty"`
	Display         examinationPrintDisplayResponse  `json:"display"`
	Items           []examinationPrintItemResponse   `json:"items"`
}

func toExaminationPrintSnapshotResponse(snapshot *ExaminationPrintSnapshot) examinationPrintSnapshotResponse {
	items := make([]examinationPrintItemResponse, 0, len(snapshot.Items))
	for i := range snapshot.Items {
		item := snapshot.Items[i]
		items = append(items, examinationPrintItemResponse{
			ID:              item.ID,
			ExamTypeFieldID: item.ExamTypeFieldID,
			Name:            item.Name,
			InspectionValue: item.InspectionValue,
			NormalValue:     item.NormalValue,
			Result:          item.Result,
			Unit:            item.Unit,
			ReferenceValue:  item.ReferenceValue,
			RefMin:          item.RefMin,
			RefMax:          item.RefMax,
			QualitativeMin:  item.QualitativeMin,
			QualitativeMax:  item.QualitativeMax,
			// Stored revision flag only — never reassess for print.
			IsAssessed: item.IsAssessed,
			IsAbnormal: item.IsAbnormal,
			Status:     string(item.Status),
			SortOrder:  item.SortOrder,
		})
	}
	return examinationPrintSnapshotResponse{
		ExaminationID:   snapshot.ExaminationID,
		ClinicID:        snapshot.ClinicID,
		Version:         snapshot.Version,
		Kind:            string(snapshot.Kind),
		Status:          string(snapshot.Status),
		PrintBoundary:   string(snapshot.PrintBoundary),
		Watermark:       snapshot.Watermark,
		Date:            httpapi.LocalTime(snapshot.Date),
		ResultSummary:   snapshot.ResultSummary,
		Machine:         snapshot.Machine,
		ExamTypeID:      snapshot.ExamTypeID,
		MedicalRecordID: snapshot.MedicalRecordID,
		PetID:           snapshot.PetID,
		DoctorID:        snapshot.DoctorID,
		Display: examinationPrintDisplayResponse{
			MedicalRecordNo:        snapshot.Display.MedicalRecordNo,
			PetName:                snapshot.Display.PetName,
			MedicalRecordOwnerName: snapshot.Display.MedicalRecordOwnerName,
			PetOwnerName:           snapshot.Display.PetOwnerName,
			SpeciesName:            snapshot.Display.SpeciesName,
			ExamTypeName:           snapshot.Display.ExamTypeName,
			DoctorName:             snapshot.Display.DoctorName,
		},
		Items: items,
	}
}
