package medicalrecord

import (
	"context"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabDeviceMultipleExamTypesMessage is the persist-time rejection message (BRT-98).
const LabDeviceMultipleExamTypesMessage = "mapped items resolve to more than one exam type"

// LabDeviceMultipleExamTypesErrorCode is the error_code stored on the job when needs_review
// is caused by items mapping to more than one exam type (ADR-007 §7 / F-1).
const LabDeviceMultipleExamTypesErrorCode = "lab_device_multiple_exam_types"

// LabDeviceResolvedItem is a catalog row that can be written to exam_results.
type LabDeviceResolvedItem struct {
	DeviceItemCode  string
	ValueShape      string
	Unit            string
	FieldName       string
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
	Unit            string
	ExamTypeFieldID *uint64
	IsActive        bool
}

type CreateLabDeviceInput struct {
	Name       string
	SourceType string
	ExamTypeID *uint64
	IsActive   bool
	SortOrder  int
}

type SaveLabDeviceConfigurationItemInput struct {
	ID uint64
	UpdateLabDeviceItemMasterInput
}

type SaveLabDeviceConfigurationInput struct {
	Device UpdateLabDeviceInput
	Items  []SaveLabDeviceConfigurationItemInput
}

type SaveLabDeviceConfigurationResult struct {
	Device *model.LabDevice
	Items  []model.LabDeviceItemMaster
}

type UpdateLabDeviceInput struct {
	Name       string
	ExamTypeID *uint64
	IsActive   bool
	SortOrder  int
}

// LabDeviceItemMasterService is the write-owner API for device item masters.
type LabDeviceItemMasterService interface {
	List(ctx context.Context, clinicID uint64, sourceType string) ([]model.LabDeviceItemMaster, error)
	EnsureDefaults(ctx context.Context, clinicID uint64) (inserted int64, items []model.LabDeviceItemMaster, err error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error)
	ResolveItems(ctx context.Context, clinicID uint64, sourceType string, codes []string) (*LabDeviceMasterResolution, error)
	ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error)
	CreateDevice(ctx context.Context, clinicID uint64, input CreateLabDeviceInput) (*model.LabDevice, error)
	UpdateDevice(ctx context.Context, clinicID, id uint64, input UpdateLabDeviceInput) (*model.LabDevice, error)
	SaveConfiguration(ctx context.Context, clinicID, id uint64, input SaveLabDeviceConfigurationInput) (*SaveLabDeviceConfigurationResult, error)
}

type labDeviceItemMasterService struct {
	repo LabDeviceItemMasterRepository
	tx   Transactor
}

// NewLabDeviceItemMasterService initializes a LabDeviceItemMasterService.
func NewLabDeviceItemMasterService(repo LabDeviceItemMasterRepository, tx ...Transactor) LabDeviceItemMasterService {
	return &labDeviceItemMasterService{repo: repo, tx: firstTransactor(tx)}
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
	defaults := labDeviceDefaults()
	devices := make([]model.LabDevice, 0, len(defaults))
	for _, device := range defaults {
		devices = append(devices, model.LabDevice{
			ClinicID:   clinicID,
			SourceType: string(device.SourceType),
			Name:       device.Name,
			IsActive:   true,
			SortOrder:  device.SortOrder,
		})
	}
	if _, err := s.repo.EnsureDevices(ctx, devices); err != nil {
		return 0, nil, err
	}
	items, err := s.repo.List(ctx, clinicID, "")
	if err != nil {
		return 0, nil, err
	}
	return inserted, items, nil
}

func (s *labDeviceItemMasterService) ListDevices(ctx context.Context, clinicID uint64) ([]model.LabDevice, error) {
	return s.repo.ListDevices(ctx, clinicID)
}

func (s *labDeviceItemMasterService) CreateDevice(
	ctx context.Context,
	clinicID uint64,
	input CreateLabDeviceInput,
) (*model.LabDevice, error) {
	name, sourceType, examTypeID, err := normalizeLabDeviceWrite(input.Name, input.SourceType, input.ExamTypeID)
	if err != nil {
		return nil, err
	}
	if !isLabDeviceSourceType(sourceType) {
		return nil, apperrors.WrapInvalidInput("unknown lab device source_type")
	}
	if err := s.validateExamType(ctx, clinicID, examTypeID); err != nil {
		return nil, err
	}
	row := &model.LabDevice{
		ClinicID:   clinicID,
		SourceType: sourceType,
		Name:       name,
		ExamTypeID: examTypeID,
		IsActive:   input.IsActive,
		SortOrder:  input.SortOrder,
	}
	if err := s.repo.CreateDevice(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *labDeviceItemMasterService) UpdateDevice(
	ctx context.Context,
	clinicID, id uint64,
	input UpdateLabDeviceInput,
) (*model.LabDevice, error) {
	name, _, examTypeID, err := normalizeLabDeviceWrite(input.Name, "", input.ExamTypeID)
	if err != nil {
		return nil, err
	}
	if err := s.validateExamType(ctx, clinicID, examTypeID); err != nil {
		return nil, err
	}
	return s.repo.UpdateDevice(ctx, clinicID, id, map[string]any{
		"name":         name,
		"exam_type_id": examTypeID,
		"is_active":    input.IsActive,
		"sort_order":   input.SortOrder,
	})
}

func (s *labDeviceItemMasterService) SaveConfiguration(
	ctx context.Context,
	clinicID, id uint64,
	input SaveLabDeviceConfigurationInput,
) (*SaveLabDeviceConfigurationResult, error) {
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("lab device master transaction is not configured")
	}
	var result SaveLabDeviceConfigurationResult
	err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		device, err := s.UpdateDevice(txCtx, clinicID, id, input.Device)
		if err != nil {
			return err
		}
		items := make([]model.LabDeviceItemMaster, 0, len(input.Items))
		seen := make(map[uint64]struct{}, len(input.Items))
		for _, item := range input.Items {
			if item.ID == 0 {
				return apperrors.WrapInvalidInput("item id is required")
			}
			if _, exists := seen[item.ID]; exists {
				return apperrors.WrapInvalidInput("duplicate item id")
			}
			seen[item.ID] = struct{}{}
			current, findErr := s.repo.FindByID(txCtx, clinicID, item.ID)
			if findErr != nil {
				return findErr
			}
			if current.SourceType != device.SourceType {
				return apperrors.WrapInvalidInput("item is not owned by this device")
			}
			updated, updateErr := s.Update(txCtx, clinicID, item.ID, item.UpdateLabDeviceItemMasterInput)
			if updateErr != nil {
				return updateErr
			}
			items = append(items, *updated)
		}
		result = SaveLabDeviceConfigurationResult{Device: device, Items: items}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *labDeviceItemMasterService) validateExamType(ctx context.Context, clinicID uint64, examTypeID *uint64) error {
	if examTypeID == nil {
		return nil
	}
	if *examTypeID == 0 {
		return apperrors.WrapInvalidInput("exam_type_id is invalid")
	}
	if _, err := s.repo.FindExamType(ctx, clinicID, *examTypeID); err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.WrapInvalidInput("exam_type_id is not in this clinic")
		}
		return err
	}
	return nil
}

func normalizeLabDeviceWrite(name, sourceType string, examTypeID *uint64) (string, string, *uint64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", nil, apperrors.WrapInvalidInput("name is required")
	}
	if len(name) > 100 {
		return "", "", nil, apperrors.WrapInvalidInput("name is too long")
	}
	if sourceType != "" && !isLabDeviceSourceType(sourceType) {
		return "", "", nil, apperrors.WrapInvalidInput("unknown lab device source_type")
	}
	return name, sourceType, examTypeID, nil
}

func (s *labDeviceItemMasterService) Update(
	ctx context.Context,
	clinicID, id uint64,
	input UpdateLabDeviceItemMasterInput,
) (*model.LabDeviceItemMaster, error) {
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
			ExamTypeFieldID: field.ID,
			ExamTypeID:      field.ExamTypeID,
		})
	}
	return res, nil
}
