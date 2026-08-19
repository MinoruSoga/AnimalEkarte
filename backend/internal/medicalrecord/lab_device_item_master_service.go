package medicalrecord

import (
	"context"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabDeviceMultipleExamTypesMessage is the persist-time rejection (BRT-98 uses AssertSingleExamType).
const LabDeviceMultipleExamTypesMessage = "mapped items resolve to more than one exam type"

// LabDeviceResolvedItem is a catalog row that can be written to exam_results.
type LabDeviceResolvedItem struct {
	DeviceItemCode  string
	ValueShape      string
	Unit            string
	FieldName       string
	DisplayName     string
	ExamTypeFieldID uint64
	ExamTypeID      uint64
}

// LabDeviceMasterResolution splits codes into mapped vs needs_review.
type LabDeviceMasterResolution struct {
	Mapped        []LabDeviceResolvedItem
	UnmappedCodes []string
}

// UpdateLabDeviceItemMasterInput is the editable subset. Codes are not changed.
type UpdateLabDeviceItemMasterInput struct {
	DisplayName     string
	Unit            string
	ExamTypeFieldID *uint64
	IsActive        bool
}

// LabDeviceItemMasterService is the write-owner API for device item masters.
type LabDeviceItemMasterService interface {
	List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error)
	EnsureDefaults(ctx context.Context, clinicID uint64) (inserted int64, items []model.LabDeviceItemMaster, err error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error)
	ResolveItems(ctx context.Context, clinicID uint64, sourceType string, codes []string) (*LabDeviceMasterResolution, error)
}

type labDeviceItemMasterService struct {
	repo LabDeviceItemMasterRepository
}

// NewLabDeviceItemMasterService initializes a LabDeviceItemMasterService.
func NewLabDeviceItemMasterService(repo LabDeviceItemMasterRepository) LabDeviceItemMasterService {
	return &labDeviceItemMasterService{repo: repo}
}

// UniqueMappedExamTypeIDs returns distinct exam_type_id values among mapped rows.
func UniqueMappedExamTypeIDs(mapped []LabDeviceResolvedItem) []uint64 {
	seen := make(map[uint64]struct{}, 2)
	ids := make([]uint64, 0, 2)
	for _, item := range mapped {
		if item.ExamTypeID == 0 {
			continue
		}
		if _, ok := seen[item.ExamTypeID]; ok {
			continue
		}
		seen[item.ExamTypeID] = struct{}{}
		ids = append(ids, item.ExamTypeID)
	}
	return ids
}

// AssertSingleExamType rejects a measurement that would create two exams (BRT-98 persist).
func AssertSingleExamType(mapped []LabDeviceResolvedItem) error {
	if len(UniqueMappedExamTypeIDs(mapped)) > 1 {
		return apperrors.WrapInvalidInput(LabDeviceMultipleExamTypesMessage)
	}
	return nil
}

func (s *labDeviceItemMasterService) List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error) {
	if sourceType != "" && !isLabDeviceSourceType(sourceType) {
		return nil, apperrors.WrapInvalidInput("unknown lab device source_type")
	}
	return s.repo.List(ctx, clinicID, sourceType)
}

func (s *labDeviceItemMasterService) EnsureDefaults(ctx context.Context, clinicID uint64) (int64, []model.LabDeviceItemMaster, error) {
	catalog := labDeviceItemCatalog()
	rows := make([]model.LabDeviceItemMaster, 0, len(catalog))
	for _, item := range catalog {
		rows = append(rows, model.LabDeviceItemMaster{
			ClinicID:       clinicID,
			SourceType:     string(item.SourceType),
			DeviceItemCode: item.DeviceItemCode,
			DisplayName:    item.DisplayName,
			Unit:           item.Unit,
			ValueShape:     item.ValueShape,
			SortOrder:      item.SortOrder,
			IsActive:       true,
		})
	}
	inserted, err := s.repo.EnsureCatalog(ctx, rows)
	if err != nil {
		return 0, nil, err
	}
	items, err := s.repo.List(ctx, clinicID, "")
	if err != nil {
		return 0, nil, err
	}
	return inserted, items, nil
}

func (s *labDeviceItemMasterService) Update(
	ctx context.Context,
	clinicID, id uint64,
	input UpdateLabDeviceItemMasterInput,
) (*model.LabDeviceItemMaster, error) {
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return nil, apperrors.WrapInvalidInput("display_name is required")
	}
	if len(displayName) > 100 {
		return nil, apperrors.WrapInvalidInput("display_name is too long")
	}
	unit := strings.TrimSpace(input.Unit)
	if len(unit) > 32 {
		return nil, apperrors.WrapInvalidInput("unit is too long")
	}
	if input.ExamTypeFieldID != nil {
		if *input.ExamTypeFieldID == 0 {
			return nil, apperrors.WrapInvalidInput("exam_type_field_id is invalid")
		}
		if _, err := s.repo.FindExamTypeField(ctx, clinicID, *input.ExamTypeFieldID); err != nil {
			if apperrors.IsNotFound(err) {
				return nil, apperrors.WrapInvalidInput("exam_type_field_id is not in this clinic")
			}
			return nil, err
		}
	}
	fields := map[string]any{
		"display_name":       displayName,
		"unit":               unit,
		"is_active":          input.IsActive,
		"exam_type_field_id": input.ExamTypeFieldID,
	}
	return s.repo.Update(ctx, clinicID, id, fields)
}

func (s *labDeviceItemMasterService) ResolveItems(
	ctx context.Context,
	clinicID uint64,
	sourceType string,
	codes []string,
) (*LabDeviceMasterResolution, error) {
	res := &LabDeviceMasterResolution{
		Mapped:        make([]LabDeviceResolvedItem, 0),
		UnmappedCodes: make([]string, 0),
	}
	if !isLabDeviceSourceType(sourceType) {
		return nil, apperrors.WrapInvalidInput("unknown lab device source_type")
	}
	if len(codes) == 0 {
		return res, nil
	}
	rows, err := s.repo.FindByClinicSourceCodes(ctx, clinicID, sourceType, codes)
	if err != nil {
		return nil, err
	}
	byCode := make(map[string]model.LabDeviceItemMaster, len(rows))
	fieldIDs := make([]uint64, 0, len(rows))
	for i := range rows {
		byCode[rows[i].DeviceItemCode] = rows[i]
		if rows[i].IsActive && rows[i].ExamTypeFieldID != nil {
			fieldIDs = append(fieldIDs, *rows[i].ExamTypeFieldID)
		}
	}
	fields, err := s.repo.FindExamTypeFields(ctx, clinicID, fieldIDs)
	if err != nil {
		return nil, err
	}
	for _, code := range codes {
		row, ok := byCode[code]
		if !ok || !row.IsActive || row.ExamTypeFieldID == nil {
			res.UnmappedCodes = append(res.UnmappedCodes, code)
			continue
		}
		field, ok := fields[*row.ExamTypeFieldID]
		if !ok || field.ExamTypeID == 0 {
			res.UnmappedCodes = append(res.UnmappedCodes, code)
			continue
		}
		res.Mapped = append(res.Mapped, LabDeviceResolvedItem{
			DeviceItemCode:  code,
			ValueShape:      row.ValueShape,
			Unit:            row.Unit,
			FieldName:       field.Name,
			DisplayName:     row.DisplayName,
			ExamTypeFieldID: field.ID,
			ExamTypeID:      field.ExamTypeID,
		})
	}
	return res, nil
}
