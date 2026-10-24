package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ConfirmLabMappingItem is one row a person submitted.
type ConfirmLabMappingItem struct {
	DeviceItemCode  string
	ExamTypeFieldID uint64
	ValueShape      string
	ChartUnit       string
}

// ConfirmLabMappingInput is the table the person confirmed. Empty Items writes nothing.
type ConfirmLabMappingInput struct {
	Items []ConfirmLabMappingItem
}

// LabMappingDraftView is a draft plus its items. Raw bytes and LLM bodies stay on the models and off the API.
type LabMappingDraftView struct {
	Draft model.LabDeviceMappingDraft
	Items []model.LabDeviceMappingDraftItem
}

var errLabMappingGenerationStopped = errors.New("generation_stopped")

// LabMappingDraftService drafts, reads, confirms, and stops lab item maps.
type LabMappingDraftService interface {
	Create(ctx context.Context, clinicID uint64, payload []byte, hint string) (*LabMappingDraftView, error)
	Get(ctx context.Context, clinicID, id uint64) (*LabMappingDraftView, error)
	Confirm(ctx context.Context, clinicID, id, staffID uint64, input ConfirmLabMappingInput) (*LabMappingDraftView, error)
	SetGenerationStopped(ctx context.Context, clinicID uint64, stopped bool) error
}

type labMappingDraftService struct {
	store   LabMappingDraftStore
	masters LabDeviceItemMasterService
	drafter LabItemDrafter
	audit   AuditTxLogger
	tx      Transactor
}

// NewLabMappingDraftService initializes a LabMappingDraftService.
// drafter may be an unconfigured client. audit and tx are required for confirm.
func NewLabMappingDraftService(
	store LabMappingDraftStore,
	masters LabDeviceItemMasterService,
	drafter LabItemDrafter,
	audit AuditTxLogger,
	tx Transactor,
) LabMappingDraftService {
	return &labMappingDraftService{store: store, masters: masters, drafter: drafter, audit: audit, tx: tx}
}

func (s *labMappingDraftService) Create(
	ctx context.Context,
	clinicID uint64,
	payload []byte,
	hint string,
) (*LabMappingDraftView, error) {
	if len(payload) == 0 || len(payload) > labDeviceMaxPayloadBytes {
		return nil, apperrors.WrapInvalidInput("invalid_payload")
	}
	raw := append([]byte(nil), payload...)
	frame, selErr := selectOneLabResult(raw, hint)
	if selErr != nil {
		return s.persistFailed(ctx, clinicID, raw, mappingSelectionCategory(selErr), "", "", labDeviceFingerprint(raw))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	stopped, err := s.store.GenerationStopped(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if stopped {
		return s.persistFailed(ctx, clinicID, raw, "generation_stopped", mappingSourceOrEmpty(frame), frame.Species, frame.SourceFingerprint)
	}
	items, err := s.draftItems(ctx, clinicID, frame)
	if err != nil && !errors.Is(err, errLabMappingGenerationStopped) {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	draft := &model.LabDeviceMappingDraft{
		ClinicID:       clinicID,
		SourceType:     mappingSourceOrEmpty(frame),
		Species:        frame.Species,
		Status:         model.LabMappingDraftStatusUnconfirmed,
		RawPayload:     raw,
		RawFingerprint: frame.SourceFingerprint,
	}
	if errors.Is(err, errLabMappingGenerationStopped) {
		draft.Status = model.LabMappingDraftStatusFailed
		draft.ErrorMessage = "generation_stopped"
	}
	return s.insertView(ctx, draft, items)
}

func (s *labMappingDraftService) persistFailed(
	ctx context.Context,
	clinicID uint64,
	raw []byte,
	category, sourceType, species, fingerprint string,
) (*LabMappingDraftView, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	draft := &model.LabDeviceMappingDraft{
		ClinicID:       clinicID,
		SourceType:     sourceType,
		Species:        normalizeLabDeviceSpecies(species),
		Status:         model.LabMappingDraftStatusFailed,
		RawPayload:     raw,
		RawFingerprint: fingerprint,
		ErrorMessage:   category,
	}
	return s.insertView(ctx, draft, nil)
}

func (s *labMappingDraftService) insertView(
	ctx context.Context,
	draft *model.LabDeviceMappingDraft,
	items []model.LabDeviceMappingDraftItem,
) (*LabMappingDraftView, error) {
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("lab mapping transaction is required")
	}
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		return s.store.InsertDraft(txCtx, draft, items)
	}); err != nil {
		return nil, err
	}
	return &LabMappingDraftView{Draft: *draft, Items: items}, nil
}

func (s *labMappingDraftService) draftItems(
	ctx context.Context,
	clinicID uint64,
	frame LabDeviceFrame,
) ([]model.LabDeviceMappingDraftItem, error) {
	codes := make([]string, 0, len(frame.Items))
	for _, item := range frame.Items {
		if isLabMappingComment(item) {
			continue
		}
		codes = append(codes, item.Code)
	}
	resolution, err := s.masters.ResolveItemsForSpecies(ctx, clinicID, mappingSourceOrEmpty(frame), frame.Species, codes)
	if err != nil {
		return nil, err
	}
	mapped := make(map[string]LabDeviceResolvedItem, len(resolution.Mapped))
	for _, row := range resolution.Mapped {
		mapped[row.DeviceItemCode] = row
	}
	needsLLM := false
	for _, item := range frame.Items {
		if isLabMappingComment(item) {
			continue
		}
		if _, ok := mapped[item.Code]; !ok {
			needsLLM = true
			break
		}
	}
	var candidates []model.ExamTypeField
	if needsLLM {
		candidates, err = s.store.ListFieldCandidates(ctx, clinicID)
		if err != nil {
			return nil, err
		}
	}
	items := make([]model.LabDeviceMappingDraftItem, 0, len(frame.Items))
	seen := make(map[string]struct{}, len(frame.Items))
	for _, item := range frame.Items {
		code := clipRunes(item.Code, labMappingCodeMax)
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		row, err := s.draftOne(ctx, clinicID, frame, item, len(items), mapped, candidates)
		if err != nil {
			return items, err
		}
		items = append(items, row)
	}
	return items, nil
}

func (s *labMappingDraftService) draftOne(
	ctx context.Context,
	clinicID uint64,
	frame LabDeviceFrame,
	item LabDeviceItem,
	sort int,
	mapped map[string]LabDeviceResolvedItem,
	candidates []model.ExamTypeField,
) (model.LabDeviceMappingDraftItem, error) {
	if isLabMappingComment(item) {
		return draftItemFromFrame(clinicID, sort, item, model.LabMappingItemSourceComment), nil
	}
	if row, ok := mapped[item.Code]; ok {
		drafted := draftItemFromFrame(clinicID, sort, item, model.LabMappingItemSourceDictionary)
		fieldID := row.ExamTypeFieldID
		drafted.ExamTypeFieldID = &fieldID
		drafted.ChartUnit = clipRunes(row.Unit, labMappingChartUnitMax)
		drafted.ProposedFieldName = row.FieldName
		drafted.UnitMismatch = unitsMismatch(item.Unit, row.Unit)
		return drafted, nil
	}
	drafted := draftItemFromFrame(clinicID, sort, item, model.LabMappingItemSourceBlank)
	if len(candidates) == 0 {
		drafted.LLMError = labMappingNoChartFields
		return drafted, nil
	}
	if s.drafter == nil {
		drafted.LLMError = labMappingLLMUnconfigured
		return drafted, nil
	}
	stopped, err := s.store.GenerationStopped(ctx, clinicID)
	if err != nil {
		return model.LabDeviceMappingDraftItem{}, err
	}
	if stopped {
		return model.LabDeviceMappingDraftItem{}, errLabMappingGenerationStopped
	}
	if err := ctx.Err(); err != nil {
		return model.LabDeviceMappingDraftItem{}, err
	}
	result, err := s.drafter.DraftItem(ctx, LabMappingLLMRequest{
		SourceType: mappingSourceOrEmpty(frame),
		Species:    frame.Species,
		Code:       drafted.DeviceItemCode,
		DeviceUnit: drafted.DeviceUnit,
		Candidates: labMappingCandidates(candidates),
	})
	if err != nil {
		return model.LabDeviceMappingDraftItem{}, err
	}
	drafted.LLMRequest = labMappingRawJSONPtr(result.RequestJSON)
	drafted.LLMResponse = labMappingRawJSONPtr(result.ResponseJSON)
	if result.ErrorCategory != "" {
		drafted.LLMError = result.ErrorCategory
		return drafted, nil
	}
	field, ok := labMappingCandidate(candidates, result.ExamTypeFieldID)
	if !ok {
		drafted.LLMError = labMappingUnlistedField
		return drafted, nil
	}
	fieldID := field.ID
	drafted.ItemSource = model.LabMappingItemSourceLLM
	drafted.ExamTypeFieldID = &fieldID
	drafted.ProposedFieldName = field.Name
	drafted.ChartUnit = clipRunes(field.Unit, labMappingChartUnitMax)
	drafted.UnitMismatch = unitsMismatch(item.Unit, field.Unit)
	return drafted, nil
}

func labMappingCandidates(fields []model.ExamTypeField) []LabMappingFieldCandidate {
	out := make([]LabMappingFieldCandidate, 0, len(fields))
	for _, field := range fields {
		out = append(out, LabMappingFieldCandidate{ID: field.ID, Name: field.Name, Unit: field.Unit})
	}
	return out
}

// labMappingRawJSONPtr は LLM 本文を NULL 許容列へ写す。本文が無い（空）場合は NULL とし、空 JSON を捏造しない。
func labMappingRawJSONPtr(raw json.RawMessage) *json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	return &raw
}

func labMappingCandidate(fields []model.ExamTypeField, id uint64) (model.ExamTypeField, bool) {
	if id == 0 {
		return model.ExamTypeField{}, false
	}
	for _, field := range fields {
		if field.ID == id {
			return field, true
		}
	}
	return model.ExamTypeField{}, false
}

func (s *labMappingDraftService) Get(ctx context.Context, clinicID, id uint64) (*LabMappingDraftView, error) {
	draft, items, err := s.store.GetDraft(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	return &LabMappingDraftView{Draft: *draft, Items: items}, nil
}

func (s *labMappingDraftService) SetGenerationStopped(ctx context.Context, clinicID uint64, stopped bool) error {
	if s.tx == nil {
		return apperrors.WrapInternalServerError("lab mapping transaction is required")
	}
	return s.tx.WithTx(ctx, func(txCtx context.Context) error {
		return s.store.SetGenerationStopped(txCtx, clinicID, stopped)
	})
}

func (s *labMappingDraftService) Confirm(
	ctx context.Context,
	clinicID, id, staffID uint64,
	input ConfirmLabMappingInput,
) (*LabMappingDraftView, error) {
	if s.audit == nil {
		return nil, apperrors.WrapInternalServerError("lab mapping confirm audit is required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("lab mapping transaction is required")
	}
	if staffID == 0 {
		return nil, apperrors.WrapInvalidInput("authenticated staff is required")
	}
	var view *LabMappingDraftView
	err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		confirmed, err := s.confirmInTx(txCtx, clinicID, id, staffID, input)
		if err != nil {
			return err
		}
		view = confirmed
		return nil
	})
	if err != nil {
		return nil, err
	}
	return view, nil
}

func (s *labMappingDraftService) confirmInTx(
	ctx context.Context,
	clinicID, id, staffID uint64,
	input ConfirmLabMappingInput,
) (*LabMappingDraftView, error) {
	draft, items, err := s.store.LockDraft(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}
	if draft.Status == model.LabMappingDraftStatusFailed {
		return nil, apperrors.WrapInvalidInput("failed draft cannot be confirmed")
	}
	if draft.Status != model.LabMappingDraftStatusUnconfirmed {
		return nil, apperrors.WrapConflict("mapping draft is already confirmed")
	}
	byCode := make(map[string]model.LabDeviceMappingDraftItem, len(items))
	for _, item := range items {
		byCode[item.DeviceItemCode] = item
	}
	submitted, err := s.submittedRows(ctx, clinicID, draft, byCode, input.Items)
	if err != nil {
		return nil, err
	}
	for i := range submitted {
		if draft.Species == "" {
			if err := s.store.UpsertDeviceMaster(ctx, submitted[i].master); err != nil {
				return nil, err
			}
			continue
		}
		if err := s.store.UpsertSpeciesSpec(ctx, submitted[i].spec); err != nil {
			return nil, err
		}
	}
	updates := make([]model.LabDeviceMappingDraftItem, 0, len(submitted))
	codes := make([]string, 0, len(submitted))
	for _, row := range submitted {
		updates = append(updates, row.item)
		codes = append(codes, row.item.DeviceItemCode)
	}
	if err := s.store.UpdateConfirmedItems(ctx, clinicID, draft.ID, updates); err != nil {
		return nil, err
	}
	confirmedAt := time.Now().UTC()
	if err := s.store.MarkConfirmed(ctx, clinicID, draft.ID, staffID, confirmedAt); err != nil {
		return nil, err
	}
	if err := s.audit.LogEntryTx(ctx, labMappingConfirmAudit(clinicID, staffID, draft, codes, confirmedAt)); err != nil {
		return nil, err
	}
	return s.Get(ctx, clinicID, draft.ID)
}

type confirmedMappingRow struct {
	item   model.LabDeviceMappingDraftItem
	master model.LabDeviceItemMaster
	spec   model.LabDeviceMappingSpec
}

func (s *labMappingDraftService) submittedRows(
	ctx context.Context,
	clinicID uint64,
	draft *model.LabDeviceMappingDraft,
	byCode map[string]model.LabDeviceMappingDraftItem,
	input []ConfirmLabMappingItem,
) ([]confirmedMappingRow, error) {
	seen := make(map[string]struct{}, len(input))
	rows := make([]confirmedMappingRow, 0, len(input))
	for _, item := range input {
		code := clipRunes(item.DeviceItemCode, labMappingCodeMax)
		if code == "" || code != item.DeviceItemCode {
			return nil, apperrors.WrapInvalidInput("device_item_code is invalid")
		}
		if _, ok := seen[code]; ok {
			return nil, apperrors.WrapInvalidInput("device_item_code is duplicated")
		}
		seen[code] = struct{}{}
		current, ok := byCode[code]
		if !ok {
			return nil, apperrors.WrapInvalidInput("device_item_code is not on this draft")
		}
		if current.ItemSource == model.LabMappingItemSourceComment {
			return nil, apperrors.WrapInvalidInput("comment items are not mapped")
		}
		if !isLabDeviceValueShape(item.ValueShape) {
			return nil, apperrors.WrapInvalidInput("value_shape is invalid")
		}
		if len([]rune(item.ChartUnit)) > labMappingChartUnitMax {
			return nil, apperrors.WrapInvalidInput("chart unit is too long")
		}
		field, err := s.store.FindExamField(ctx, clinicID, item.ExamTypeFieldID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return nil, apperrors.WrapInvalidInput("exam_type_field_id is not in this clinic")
			}
			return nil, err
		}
		fieldID := field.ID
		updated := current
		updated.ExamTypeFieldID = &fieldID
		updated.ChartUnit = item.ChartUnit
		updated.UnitMismatch = unitsMismatch(current.DeviceUnit, item.ChartUnit)
		updated.ProposedFieldName = field.Name
		updated.Confirmed = true
		row := confirmedMappingRow{item: updated}
		if draft.Species == "" {
			row.master = model.LabDeviceItemMaster{
				ClinicID:        clinicID,
				SourceType:      draft.SourceType,
				DeviceItemCode:  code,
				Unit:            item.ChartUnit,
				ValueShape:      item.ValueShape,
				ExamTypeFieldID: &fieldID,
				IsActive:        true,
			}
		} else {
			row.spec = model.LabDeviceMappingSpec{
				ClinicID:         clinicID,
				SourceType:       draft.SourceType,
				Species:          draft.Species,
				DeviceItemCode:   code,
				Unit:             item.ChartUnit,
				ValueShape:       item.ValueShape,
				ExamTypeFieldID:  fieldID,
				ConfirmedDraftID: draft.ID,
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func labMappingConfirmAudit(
	clinicID, staffID uint64,
	draft *model.LabDeviceMappingDraft,
	codes []string,
	at time.Time,
) *AuditEntry {
	return &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &staffID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionLabDeviceMappingConfirm,
		Resource:   "lab_device_mapping_draft",
		ResourceID: &draft.ID,
		Metadata: map[string]any{
			"source_type":       draft.SourceType,
			"species":           draft.Species,
			"submitted_count":   len(codes),
			"device_item_codes": codes,
			"confirmed_at":      at.Format(time.RFC3339),
		},
	}
}
