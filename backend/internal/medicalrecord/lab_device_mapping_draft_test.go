package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const mappingSentinelValue = "77.77"

func TestLabMappingDraftService_DictionarySkipsLLM(t *testing.T) {
	svc, store, masters, drafter := newMappingHarness()
	fieldID := uint64(7)
	masters.resolution = &LabDeviceMasterResolution{
		Mapped: []LabDeviceResolvedItem{{
			DeviceItemCode:  "BUN-P",
			ValueShape:      model.LabDeviceValueShapeNumeric,
			Unit:            "mg/dl",
			FieldName:       "BUN",
			ExamTypeFieldID: fieldID,
			ExamTypeID:      3,
		}},
	}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	require.Len(t, view.Items, 1)
	assert.Equal(t, model.LabMappingItemSourceDictionary, view.Items[0].ItemSource)
	assert.Nil(t, view.Items[0].LLMRequest)
	assert.Equal(t, mappingSentinelValue, view.Items[0].ValueRaw)
	assert.False(t, view.Items[0].UnitMismatch)
	assert.Zero(t, drafter.calls)
	assert.Empty(t, store.masters)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, view.Draft.Status)
}

func TestLabMappingDraftService_LLMOmitsMeasuredValue(t *testing.T) {
	svc, _, masters, drafter := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	drafter.result = LabMappingLLMResult{ExamTypeFieldID: 7, ResponseJSON: json.RawMessage(`{"exam_type_field_id":7}`)}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	require.Equal(t, 1, drafter.calls)
	body, err := json.Marshal(drafter.last)
	require.NoError(t, err)
	assert.NotContains(t, string(body), mappingSentinelValue)
	assert.Equal(t, "BUN-P", drafter.last.Code)
	require.Len(t, view.Items, 1)
	assert.Equal(t, model.LabMappingItemSourceLLM, view.Items[0].ItemSource)
	assert.Equal(t, "BUN", view.Items[0].ProposedFieldName)
	assert.JSONEq(t, `{"exam_type_field_id":7}`, string(view.Items[0].LLMResponse))
	assert.NotContains(t, string(view.Items[0].LLMRequest), mappingSentinelValue)
}

func TestLabMappingDraftService_UnitMismatchIsNotConverted(t *testing.T) {
	svc, _, masters, drafter := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{
		Mapped: []LabDeviceResolvedItem{{
			DeviceItemCode: "BUN-P", Unit: "g/dl", FieldName: "BUN", ExamTypeFieldID: 7, ExamTypeID: 3,
		}},
	}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	require.Len(t, view.Items, 1)
	assert.True(t, view.Items[0].UnitMismatch)
	assert.Equal(t, mappingSentinelValue, view.Items[0].ValueRaw)
	assert.Equal(t, "g/dl", view.Items[0].ChartUnit)
	assert.Zero(t, drafter.calls)
}

func TestLabMappingDraftService_StopSkipsLLMAndKeepsRaw(t *testing.T) {
	svc, store, _, drafter := newMappingHarness()
	require.NoError(t, svc.SetGenerationStopped(context.Background(), 1, true))
	payload := fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", "")
	view, err := svc.Create(context.Background(), 1, payload, "nx600")
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusFailed, view.Draft.Status)
	assert.Equal(t, "generation_stopped", view.Draft.ErrorMessage)
	assert.Equal(t, payload, view.Draft.RawPayload)
	assert.Empty(t, view.Items)
	assert.Zero(t, drafter.calls)
	assert.Empty(t, store.masters)
	assert.True(t, store.stopped)
}

func TestLabMappingDraftService_SecondDraftDoesNotWriteChartMappings(t *testing.T) {
	svc, store, masters, _ := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	payload := fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", "")
	first, err := svc.Create(context.Background(), 1, payload, "nx600")
	require.NoError(t, err)
	second, err := svc.Create(context.Background(), 1, fujiOne("S2", "BUN-P", "1.0", "mg/dl", ""), "nx600")
	require.NoError(t, err)
	assert.NotEqual(t, first.Draft.ID, second.Draft.ID)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, first.Draft.Status)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, store.drafts[first.Draft.ID].Status)
	assert.Empty(t, store.masters)
	assert.Empty(t, store.specs)
}

func TestLabMappingDraftService_ConfirmWritesOnlySubmittedRows(t *testing.T) {
	svc, store, masters, _ := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P", "CRE-P"}}
	payload := synthFujiHeaderOnly("NORMAL ", "2026-01-15", "10:30", "S1", [][]byte{
		fujiEqualsSlot("BUN-P", mappingSentinelValue, "mg/dl", ""),
		fujiEqualsSlot("CRE-P", "1.2", "mg/dl", ""),
	})
	view, err := svc.Create(context.Background(), 1, payload, "nx600")
	require.NoError(t, err)
	require.Len(t, view.Items, 2)
	store.masters = append(store.masters, model.LabDeviceItemMaster{
		ClinicID: 1, SourceType: "fuji_nx600", DeviceItemCode: "KEEP", Unit: "mg/dl", IsActive: true,
	})
	confirmed, err := svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: "BUN-P", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusConfirmed, confirmed.Draft.Status)
	require.Len(t, store.masters, 2)
	assert.Equal(t, "KEEP", store.masters[0].DeviceItemCode)
	assert.Equal(t, "BUN-P", store.masters[1].DeviceItemCode)
	assert.Empty(t, store.specs)
}

func TestLabMappingDraftService_SpeciesConfirmWritesSpecNotDeviceMaster(t *testing.T) {
	svc, store, masters, drafter := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"WBC"}}
	drafter.result = LabMappingLLMResult{ExamTypeFieldID: 7}
	payload := idexxSpeciesFrame("Canine", "WBC  5.50  K/uL")
	view, err := svc.Create(context.Background(), 1, payload, "vetlab")
	require.NoError(t, err)
	assert.Equal(t, "canine", view.Draft.Species)
	confirmed, err := svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: "WBC", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "K/uL",
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusConfirmed, confirmed.Draft.Status)
	assert.Empty(t, store.masters)
	require.Len(t, store.specs, 1)
	assert.Equal(t, "canine", store.specs[0].Species)
	assert.Equal(t, uint64(7), store.specs[0].ExamTypeFieldID)
}

func TestLabMappingDraftService_AuditErrorRollsBack(t *testing.T) {
	svc, store, masters, _ := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	svc.(*mappingHarness).audit.err = errors.New("audit failed")
	_, err = svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: "BUN-P", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
		}},
	})
	require.Error(t, err)
	assert.Empty(t, store.masters)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, store.drafts[view.Draft.ID].Status)
}

func TestLabMappingDraftService_ConfirmRejectsFieldOutsideClinic(t *testing.T) {
	svc, store, masters, _ := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	_, err = svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: "BUN-P", ExamTypeFieldID: 99, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
		}},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Empty(t, store.masters)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, store.drafts[view.Draft.ID].Status)
}

func TestLabMappingDraftService_ConfirmGuards(t *testing.T) {
	svc, store, masters, drafter := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)

	_, err = svc.Confirm(context.Background(), 1, view.Draft.ID, 0, ConfirmLabMappingInput{})
	assert.True(t, apperrors.IsInvalidInput(err))

	confirmed, err := svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{})
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusConfirmed, confirmed.Draft.Status)
	assert.Empty(t, store.masters)

	_, err = svc.Confirm(context.Background(), 1, view.Draft.ID, 4, ConfirmLabMappingInput{})
	assert.True(t, apperrors.IsConflict(err))

	failed, err := svc.Create(context.Background(), 1, []byte{0x02, 0x03}, "nx600")
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusFailed, failed.Draft.Status)
	_, err = svc.Confirm(context.Background(), 1, failed.Draft.ID, 4, ConfirmLabMappingInput{})
	assert.True(t, apperrors.IsInvalidInput(err))

	commentPayload := fujiOne("S3", "NOTE", "1", "mg/dl", "comment")
	before := drafter.calls
	commentView, err := svc.Create(context.Background(), 1, commentPayload, "nx600")
	require.NoError(t, err)
	assert.Equal(t, before, drafter.calls)
	require.NotEmpty(t, commentView.Items)
	assert.Equal(t, model.LabMappingItemSourceComment, commentView.Items[0].ItemSource)
	_, err = svc.Confirm(context.Background(), 1, commentView.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: commentView.Items[0].DeviceItemCode, ExamTypeFieldID: 7,
			ValueShape: model.LabDeviceValueShapeText, ChartUnit: "",
		}},
	})
	assert.True(t, apperrors.IsInvalidInput(err))

	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	again, err := svc.Create(context.Background(), 1, fujiOne("S4", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	_, err = svc.Confirm(context.Background(), 1, again.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{
			{DeviceItemCode: "BUN-P", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl"},
			{DeviceItemCode: "BUN-P", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl"},
		},
	})
	assert.True(t, apperrors.IsInvalidInput(err))
	_, err = svc.Confirm(context.Background(), 1, again.Draft.ID, 4, ConfirmLabMappingInput{
		Items: []ConfirmLabMappingItem{{
			DeviceItemCode: "BUN-P", ExamTypeFieldID: 7, ValueShape: model.LabDeviceValueShapeNumeric,
			ChartUnit: strings.Repeat("あ", 33),
		}},
	})
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestLabMappingDraftService_BlankPaths(t *testing.T) {
	svc, store, masters, drafter := newMappingHarness()
	masters.resolution = &LabDeviceMasterResolution{UnmappedCodes: []string{"BUN-P"}}
	store.candidates = nil
	view, err := svc.Create(context.Background(), 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	assert.Equal(t, labMappingNoChartFields, view.Items[0].LLMError)
	assert.Equal(t, model.LabMappingItemSourceBlank, view.Items[0].ItemSource)
	assert.Zero(t, drafter.calls)

	store.candidates = []model.ExamTypeField{{ID: 7, Name: "BUN", Unit: "mg/dl", ExamTypeID: 3}}
	bare := NewLabMappingDraftService(store, masters, nil, &scriptAudit{}, snapTx{store: store})
	view, err = bare.Create(context.Background(), 1, fujiOne("S2", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	assert.Equal(t, labMappingLLMUnconfigured, view.Items[0].LLMError)

	drafter.err = errors.New("llm down")
	before := len(store.drafts)
	_, err = svc.Create(context.Background(), 1, fujiOne("S3", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.Error(t, err)
	assert.Len(t, store.drafts, before)

	drafter.err = nil
	drafter.result = LabMappingLLMResult{ErrorCategory: labMappingLLMHTTP, RequestJSON: json.RawMessage(`{"code":"BUN-P"}`)}
	view, err = svc.Create(context.Background(), 1, fujiOne("S4", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingItemSourceBlank, view.Items[0].ItemSource)
	assert.Equal(t, labMappingLLMHTTP, view.Items[0].LLMError)
	assert.Nil(t, view.Items[0].LLMResponse)
}

func TestLabMappingDraftService_SelectFailures(t *testing.T) {
	svc, store, _, drafter := newMappingHarness()
	_, err := svc.Create(context.Background(), 1, nil, "")
	require.Error(t, err)
	assert.Empty(t, store.drafts)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = svc.Create(ctx, 1, fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), "nx600")
	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, store.drafts)

	multi := append(fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", ""), fujiOne("S2", "CRE-P", "1.2", "mg/dl", "")...)
	view, err := svc.Create(context.Background(), 1, multi, "nx600")
	require.NoError(t, err)
	assert.Equal(t, "multiple_results", view.Draft.ErrorMessage)
	assert.Empty(t, view.Items)
	assert.Zero(t, drafter.calls)

	same := fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", "")
	view, err = svc.Create(context.Background(), 1, append(append([]byte{}, same...), same...), "nx600")
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, view.Draft.Status)
	require.Len(t, view.Items, 1)
}

func TestLabMappingHelpers_SpeciesAndUnits(t *testing.T) {
	assert.Equal(t, "canine", normalizeLabDeviceSpecies(" dog "))
	assert.Equal(t, "feline", normalizeLabDeviceSpecies("CAT"))
	assert.Equal(t, "", normalizeLabDeviceSpecies("poodle"))
	assert.Equal(t, "canine", labDeviceSpeciesFromText("結果 犬 WBC"))
	assert.Equal(t, "feline", labDeviceSpeciesFromText("feline then canine"))
	assert.Equal(t, "", labDeviceSpeciesFromText("catheter and hotdog"))
	assert.Equal(t, "", labDeviceSpeciesFromText("犬種と猫砂"))
	assert.Equal(t, "", labDeviceSpeciesFromText("xcanine"))
	assert.False(t, unitsMismatch("mg/dL", "mg/dl"))
	assert.False(t, unitsMismatch("µmol/L", "umol/l"))
	assert.True(t, unitsMismatch("mg/dl", "g/dl"))
	assert.False(t, unitsMismatch("", "g/dl"))

	frame, err := decodeIDEXXLongFrame([]byte("犬 WBC  5.50  K/uL"), []byte("raw-canine"))
	require.NoError(t, err)
	assert.Equal(t, "canine", frame.Species)
	frame, err = decodeIDEXXLongFrame([]byte("catheter WBC  5.50  K/uL"), []byte("raw-cat"))
	require.NoError(t, err)
	assert.Equal(t, "", frame.Species)
}

type mappingHarness struct {
	LabMappingDraftService
	audit *scriptAudit
}

func newMappingHarness() (LabMappingDraftService, *memMappingStore, *mappingMasterFake, *mappingDrafterFake) {
	store := newMemMappingStore()
	masters := &mappingMasterFake{}
	drafter := &mappingDrafterFake{}
	audit := &scriptAudit{}
	svc := NewLabMappingDraftService(store, masters, drafter, audit, snapTx{store: store})
	return &mappingHarness{LabMappingDraftService: svc, audit: audit}, store, masters, drafter
}

func fujiOne(specimen, code, value, unit, flag string) []byte {
	return synthFujiHeaderOnly("NORMAL ", "2026-01-15", "10:30", specimen, [][]byte{
		fujiEqualsSlot(code, value, unit, flag),
	})
}

func idexxSpeciesFrame(species, analyte string) []byte {
	body := species + "\n" + analyte
	out := make([]byte, 0, len(body)+2)
	out = append(out, 0x02)
	out = append(out, body...)
	out = append(out, 0x03)
	return out
}

type scriptAudit struct{ err error }

func (a *scriptAudit) LogEntryTx(context.Context, *AuditEntry) error {
	if a.err != nil {
		return a.err
	}
	return nil
}

type snapTx struct{ store *memMappingStore }

func (t snapTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	snap := t.store.clone()
	if err := fn(ctx); err != nil {
		t.store.restore(snap)
		return err
	}
	return nil
}

type memMappingStore struct {
	stopped    bool
	nextID     uint64
	drafts     map[uint64]model.LabDeviceMappingDraft
	items      map[uint64][]model.LabDeviceMappingDraftItem
	masters    []model.LabDeviceItemMaster
	specs      []model.LabDeviceMappingSpec
	fields     map[uint64]model.ExamTypeField
	candidates []model.ExamTypeField
}

func newMemMappingStore() *memMappingStore {
	return &memMappingStore{
		nextID: 1,
		drafts: map[uint64]model.LabDeviceMappingDraft{},
		items:  map[uint64][]model.LabDeviceMappingDraftItem{},
		fields: map[uint64]model.ExamTypeField{
			7: {ID: 7, ClinicID: 1, ExamTypeID: 3, Name: "BUN", Unit: "mg/dl"},
		},
		candidates: []model.ExamTypeField{{ID: 7, ClinicID: 1, ExamTypeID: 3, Name: "BUN", Unit: "mg/dl"}},
	}
}

func (s *memMappingStore) clone() memMappingStore {
	cp := memMappingStore{
		stopped:    s.stopped,
		nextID:     s.nextID,
		drafts:     make(map[uint64]model.LabDeviceMappingDraft, len(s.drafts)),
		items:      make(map[uint64][]model.LabDeviceMappingDraftItem, len(s.items)),
		masters:    append([]model.LabDeviceItemMaster(nil), s.masters...),
		specs:      append([]model.LabDeviceMappingSpec(nil), s.specs...),
		fields:     make(map[uint64]model.ExamTypeField, len(s.fields)),
		candidates: append([]model.ExamTypeField(nil), s.candidates...),
	}
	for id, draft := range s.drafts {
		cp.drafts[id] = draft
	}
	for id, items := range s.items {
		cp.items[id] = append([]model.LabDeviceMappingDraftItem(nil), items...)
	}
	for id, field := range s.fields {
		cp.fields[id] = field
	}
	return cp
}

func (s *memMappingStore) restore(cp memMappingStore) { *s = cp }

func (s *memMappingStore) GenerationStopped(context.Context, uint64) (bool, error) {
	return s.stopped, nil
}

func (s *memMappingStore) SetGenerationStopped(_ context.Context, _ uint64, stopped bool) error {
	s.stopped = stopped
	return nil
}

func (s *memMappingStore) InsertDraft(_ context.Context, draft *model.LabDeviceMappingDraft, items []model.LabDeviceMappingDraftItem) error {
	if draft.ID == 0 {
		draft.ID = s.nextID
		s.nextID++
	}
	s.drafts[draft.ID] = *draft
	copied := append([]model.LabDeviceMappingDraftItem(nil), items...)
	for i := range copied {
		copied[i].DraftID = draft.ID
		copied[i].ClinicID = draft.ClinicID
	}
	s.items[draft.ID] = copied
	return nil
}

func (s *memMappingStore) GetDraft(_ context.Context, _, id uint64) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	return s.copyDraft(id)
}

func (s *memMappingStore) LockDraft(_ context.Context, _, id uint64) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	return s.copyDraft(id)
}

func (s *memMappingStore) copyDraft(id uint64) (*model.LabDeviceMappingDraft, []model.LabDeviceMappingDraftItem, error) {
	draft, ok := s.drafts[id]
	if !ok {
		return nil, nil, apperrors.WrapNotFound("lab_device_mapping_draft", "missing")
	}
	items := append([]model.LabDeviceMappingDraftItem(nil), s.items[id]...)
	return &draft, items, nil
}

func (s *memMappingStore) MarkConfirmed(_ context.Context, _, id, staffID uint64, at time.Time) error {
	draft, ok := s.drafts[id]
	if !ok || draft.Status != model.LabMappingDraftStatusUnconfirmed {
		return apperrors.WrapConflict("mapping draft is already confirmed")
	}
	draft.Status = model.LabMappingDraftStatusConfirmed
	draft.ConfirmedAt = &at
	draft.ConfirmedByStaffID = &staffID
	s.drafts[id] = draft
	return nil
}

func (s *memMappingStore) UpdateConfirmedItems(_ context.Context, _, draftID uint64, items []model.LabDeviceMappingDraftItem) error {
	current := s.items[draftID]
	for _, item := range items {
		found := false
		for i := range current {
			if current[i].DeviceItemCode == item.DeviceItemCode {
				current[i] = item
				found = true
			}
		}
		if !found {
			return apperrors.WrapNotFound("lab_device_mapping_draft_item", item.DeviceItemCode)
		}
	}
	s.items[draftID] = current
	return nil
}

func (s *memMappingStore) UpsertDeviceMaster(_ context.Context, row model.LabDeviceItemMaster) error {
	for i := range s.masters {
		if s.masters[i].ClinicID == row.ClinicID && s.masters[i].SourceType == row.SourceType && s.masters[i].DeviceItemCode == row.DeviceItemCode {
			s.masters[i] = row
			return nil
		}
	}
	s.masters = append(s.masters, row)
	return nil
}

func (s *memMappingStore) UpsertSpeciesSpec(_ context.Context, spec model.LabDeviceMappingSpec) error {
	for i := range s.specs {
		if s.specs[i].ClinicID == spec.ClinicID && s.specs[i].SourceType == spec.SourceType && s.specs[i].Species == spec.Species && s.specs[i].DeviceItemCode == spec.DeviceItemCode {
			s.specs[i] = spec
			return nil
		}
	}
	s.specs = append(s.specs, spec)
	return nil
}

func (s *memMappingStore) FindExamField(_ context.Context, clinicID, fieldID uint64) (*model.ExamTypeField, error) {
	field, ok := s.fields[fieldID]
	if !ok || field.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("exam_type_field", "missing")
	}
	copy := field
	return &copy, nil
}

func (s *memMappingStore) ListFieldCandidates(context.Context, uint64) ([]model.ExamTypeField, error) {
	return append([]model.ExamTypeField(nil), s.candidates...), nil
}

type mappingMasterFake struct {
	resolution *LabDeviceMasterResolution
}

func (m *mappingMasterFake) ResolveItemsForSpecies(_ context.Context, _ uint64, _ string, _ string, codes []string) (*LabDeviceMasterResolution, error) {
	if m.resolution != nil {
		return m.resolution, nil
	}
	return &LabDeviceMasterResolution{UnmappedCodes: append([]string(nil), codes...)}, nil
}

func (m *mappingMasterFake) List(context.Context, uint64, string) ([]model.LabDeviceItemMaster, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) EnsureDefaults(context.Context, uint64) (int64, []model.LabDeviceItemMaster, error) {
	return 0, nil, errors.New("not used")
}
func (m *mappingMasterFake) Update(context.Context, uint64, uint64, UpdateLabDeviceItemMasterInput) (*model.LabDeviceItemMaster, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) ResolveItems(context.Context, uint64, string, []string) (*LabDeviceMasterResolution, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) ListDevices(context.Context, uint64) ([]model.LabDevice, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) CreateDevice(context.Context, uint64, CreateLabDeviceInput) (*model.LabDevice, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) UpdateDevice(context.Context, uint64, uint64, UpdateLabDeviceInput) (*model.LabDevice, error) {
	return nil, errors.New("not used")
}
func (m *mappingMasterFake) SaveConfiguration(context.Context, uint64, uint64, SaveLabDeviceConfigurationInput) (*SaveLabDeviceConfigurationResult, error) {
	return nil, errors.New("not used")
}

type mappingDrafterFake struct {
	calls  int
	last   LabMappingLLMRequest
	result LabMappingLLMResult
	err    error
}

func (d *mappingDrafterFake) DraftItem(_ context.Context, req LabMappingLLMRequest) (LabMappingLLMResult, error) {
	d.calls++
	d.last = req
	if d.err != nil {
		return LabMappingLLMResult{}, d.err
	}
	result := d.result
	if result.RequestJSON == nil {
		body, err := json.Marshal(req)
		if err != nil {
			return LabMappingLLMResult{}, err
		}
		result.RequestJSON = body
	}
	return result, nil
}
