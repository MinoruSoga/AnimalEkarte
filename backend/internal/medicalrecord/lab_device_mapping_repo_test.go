package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"gorm.io/gorm/schema"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestLabMappingDraftRepository_WritesRequireAmbientTx(t *testing.T) {
	repo := NewLabMappingDraftRepository(nil)
	ctx := context.Background()
	require.Error(t, repo.SetGenerationStopped(ctx, 1, true))
	require.Error(t, repo.InsertDraft(ctx, &model.LabDeviceMappingDraft{ClinicID: 1}, nil))
	_, _, err := repo.LockDraft(ctx, 1, 1)
	require.Error(t, err)
	require.Error(t, repo.MarkConfirmed(ctx, 1, 1, 1, time.Now()))
	require.Error(t, repo.UpdateConfirmedItems(ctx, 1, 1, nil))
	require.Error(t, repo.UpsertDeviceMaster(ctx, model.LabDeviceItemMaster{}))
	require.Error(t, repo.UpsertSpeciesSpec(ctx, model.LabDeviceMappingSpec{}))
}

func TestFindMappingSpecs_EmptySpeciesDoesNotTouchDB(t *testing.T) {
	repo := NewLabDeviceItemMasterRepository(nil)
	rows, err := repo.FindMappingSpecs(context.Background(), 1, "fuji_nx600", "", []string{"BUN-P"})
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestLabMappingDraftRepository_ConfirmAndSpeciesOverlay(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.LabDeviceItemMaster{},
		&model.LabDeviceMappingDraft{},
		&model.LabDeviceMappingDraftItem{},
		&model.LabDeviceMappingSpec{},
		&model.LabDeviceMappingGeneration{},
	))
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	examA := &model.ExaminationType{ClinicID: clinicA, Name: "mapping-a"}
	require.NoError(t, db.Create(examA).Error)
	fieldA := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examA.ID, Name: "BUN", Unit: "mg/dl"}
	require.NoError(t, db.Create(fieldA).Error)
	examB := &model.ExaminationType{ClinicID: clinicB, Name: "mapping-b"}
	require.NoError(t, db.Create(examB).Error)
	fieldB := &model.ExamTypeField{ClinicID: clinicB, ExamTypeID: examB.ID, Name: "BUN", Unit: "mg/dl"}
	require.NoError(t, db.Create(fieldB).Error)
	t.Cleanup(func() {
		db.Exec("DELETE FROM lab_device_mapping_draft_items WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_specs WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_drafts WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_mapping_generation WHERE clinic_id IN (1,2)")
		db.Exec("DELETE FROM lab_device_item_masters WHERE clinic_id IN (1,2) AND device_item_code IN ('NEW','KEEP')")
		db.Exec("DELETE FROM exam_type_fields WHERE id IN (?,?)", fieldA.ID, fieldB.ID)
		db.Exec("DELETE FROM exam_types WHERE id IN (?,?)", examA.ID, examB.ID)
	})

	store := NewLabMappingDraftRepository(db)
	masters := NewLabDeviceItemMasterService(NewLabDeviceItemMasterRepository(db))
	tx := persistence.NewTransactor(db)
	require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
		return store.InsertDraft(txCtx, &model.LabDeviceMappingDraft{
			ClinicID: clinicA, SourceType: "fuji_nx600", Status: model.LabMappingDraftStatusUnconfirmed,
			RawPayload: []byte{0x02, 0x03}, RawFingerprint: "draft-new",
		}, []model.LabDeviceMappingDraftItem{{
			DeviceItemCode: "NEW", ItemSource: model.LabMappingItemSourceBlank, ValueRaw: mappingSentinelValue,
			DeviceUnit: "mg/dl", ChartUnit: "g/dl", UnitMismatch: true,
		}})
	}))

	res, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, res.Mapped)

	var draft model.LabDeviceMappingDraft
	require.NoError(t, db.Where("clinic_id = ? AND raw_fingerprint = ?", clinicA, "draft-new").First(&draft).Error)
	require.NoError(t, db.Create(&model.LabDeviceMappingSpec{
		ClinicID: clinicA, SourceType: "fuji_nx600", Species: "canine", DeviceItemCode: "NEW",
		Unit: "g/dl", ValueShape: model.LabDeviceValueShapeNumeric, ExamTypeFieldID: fieldA.ID, ConfirmedDraftID: draft.ID,
	}).Error)
	canine, err := masters.ResolveItemsForSpecies(ctx, clinicA, "fuji_nx600", "canine", []string{"NEW"})
	require.NoError(t, err)
	require.Len(t, canine.Mapped, 1)
	assert.Equal(t, fieldA.ID, canine.Mapped[0].ExamTypeFieldID)
	plain, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, plain.Mapped)
	feline, err := masters.ResolveItemsForSpecies(ctx, clinicA, "fuji_nx600", "feline", []string{"NEW"})
	require.NoError(t, err)
	assert.Empty(t, feline.Mapped)

	require.NoError(t, db.Create(&model.LabDeviceItemMaster{
		ClinicID: clinicA, SourceType: "fuji_nx600", DeviceItemCode: "KEEP",
		Unit: "mg/dl", ValueShape: model.LabDeviceValueShapeNumeric, ExamTypeFieldID: &fieldA.ID, IsActive: true,
	}).Error)
	audit := &scriptAudit{}
	svc := NewLabMappingDraftService(store, masters, nil, audit, tx)
	require.NoError(t, svc.SetGenerationStopped(ctx, clinicA, true))
	kept, err := masters.ResolveItems(ctx, clinicA, "fuji_nx600", []string{"KEEP"})
	require.NoError(t, err)
	require.Len(t, kept.Mapped, 1)

	_, err = svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldB.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.True(t, apperrors.IsInvalidInput(err))

	audit.err = assert.AnError
	_, err = svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldA.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.Error(t, err)
	var masterCount int64
	require.NoError(t, db.Model(&model.LabDeviceItemMaster{}).
		Where("clinic_id = ? AND device_item_code = ?", clinicA, "NEW").Count(&masterCount).Error)
	assert.Zero(t, masterCount)
	var still model.LabDeviceMappingDraft
	require.NoError(t, db.First(&still, draft.ID).Error)
	assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, still.Status)

	audit.err = nil
	confirmed, err := svc.Confirm(ctx, clinicA, draft.ID, 4, ConfirmLabMappingInput{Items: []ConfirmLabMappingItem{{
		DeviceItemCode: "NEW", ExamTypeFieldID: fieldA.ID, ValueShape: model.LabDeviceValueShapeNumeric, ChartUnit: "mg/dl",
	}}})
	require.NoError(t, err)
	assert.Equal(t, model.LabMappingDraftStatusConfirmed, confirmed.Draft.Status)
	require.NoError(t, db.Model(&model.LabDeviceItemMaster{}).
		Where("clinic_id = ? AND device_item_code = ?", clinicA, "NEW").Count(&masterCount).Error)
	assert.Equal(t, int64(1), masterCount)
	saved, err := svc.Get(ctx, clinicA, draft.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(saved.Items))
	assert.False(t, saved.Items[0].UnitMismatch, "matching submitted unit must overwrite true")
	for _, tc := range []struct {
		unit     string
		mismatch bool
	}{{"g/dl", true}, {"", false}, {"mg/dl", false}} {
		updated := saved.Items[0]
		updated.ChartUnit = tc.unit
		updated.UnitMismatch = tc.mismatch
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			return store.UpdateConfirmedItems(txCtx, clinicA, draft.ID, []model.LabDeviceMappingDraftItem{updated})
		}))
		var persisted model.LabDeviceMappingDraftItem
		require.NoError(t, db.Where("clinic_id = ? AND draft_id = ? AND device_item_code = ?", clinicA, draft.ID, "NEW").First(&persisted).Error)
		assert.Equal(t, tc.mismatch, persisted.UnitMismatch)
	}
}

func TestLabMappingDraftRepository_InvalidHTTPBodiesPersist(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ExaminationType{}, &model.ExamTypeField{}, &model.LabDeviceMappingDraft{}, &model.LabDeviceMappingDraftItem{}))
	exam := model.ExaminationType{ClinicID: 1, Name: "mapping-http"}
	require.NoError(t, db.Create(&exam).Error)
	field := model.ExamTypeField{ClinicID: 1, ExamTypeID: exam.ID, Name: "BUN", Unit: "mg/dl"}
	require.NoError(t, db.Create(&field).Error)
	var draftIDs []uint64
	t.Cleanup(func() {
		if len(draftIDs) > 0 {
			db.Unscoped().Where("clinic_id = ? AND draft_id IN ?", 1, draftIDs).Delete(&model.LabDeviceMappingDraftItem{})
			db.Unscoped().Where("clinic_id = ? AND id IN ?", 1, draftIDs).Delete(&model.LabDeviceMappingDraft{})
		}
		db.Unscoped().Where("clinic_id = ?", 1).Delete(&field)
		db.Unscoped().Where("clinic_id = ?", 1).Delete(&exam)
	})
	for _, status := range []int{http.StatusOK, http.StatusBadGateway} {
		for _, tc := range []struct {
			name   string
			body   []byte
			binary bool
		}{
			{"text", []byte("invalid model body"), false},
			{"nul", []byte{'x', 0, 'y'}, true},
			{"invalid_utf8", []byte{'x', 0xff, 'y'}, true},
		} {
			t.Run(http.StatusText(status)+"/"+tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status); _, _ = w.Write(tc.body) }))
				t.Cleanup(server.Close)
				store := NewLabMappingDraftRepository(db)
				svc := NewLabMappingDraftService(store, &mappingMasterFake{}, NewOpenAILabItemDrafter(server.Client(), "test-key", server.URL), &scriptAudit{}, persistence.NewTransactor(db))
				payload := fujiOne("S1", "BUN-P", mappingSentinelValue, "mg/dl", "")
				view, err := svc.Create(context.Background(), 1, payload, "nx600")
				require.NoError(t, err)
				draftIDs = append(draftIDs, view.Draft.ID)
				saved, err := svc.Get(context.Background(), 1, view.Draft.ID)
				require.NoError(t, err)
				assert.Equal(t, model.LabMappingDraftStatusUnconfirmed, saved.Draft.Status)
				assert.True(t, bytes.Equal(payload, saved.Draft.RawPayload), "raw payload must persist")
				require.Equal(t, 1, len(saved.Items))
				assert.Equal(t, model.LabMappingItemSourceBlank, saved.Items[0].ItemSource)
				category := labMappingLLMParse
				if status != http.StatusOK {
					category = labMappingLLMHTTP
				}
				assert.Equal(t, category, saved.Items[0].LLMError)
				assert.True(t, json.Valid(saved.Items[0].LLMResponse))
				if tc.binary {
					var retainedBody []byte
					if saved.Items[0].LLMResponse[0] == '"' {
						var retained string
						require.NoError(t, json.Unmarshal(saved.Items[0].LLMResponse, &retained))
						retainedBody = []byte(retained)
					} else {
						var retained struct {
							Encoding string `json:"encoding"`
							Body     []byte `json:"body"`
						}
						require.NoError(t, json.Unmarshal(saved.Items[0].LLMResponse, &retained))
						assert.Equal(t, "base64", retained.Encoding)
						retainedBody = retained.Body
					}
					assert.True(t, bytes.Equal(tc.body, retainedBody), "exact response bytes must persist")
				} else {
					var retained string
					require.NoError(t, json.Unmarshal(saved.Items[0].LLMResponse, &retained))
					assert.True(t, retained == string(tc.body), "response must persist")
				}
			})
		}
	}
}

func TestLabMappingDraftRepository_ClinicUniqueSchema(t *testing.T) {
	t.Run("gorm_index", func(t *testing.T) {
		parsed, err := schema.Parse(&model.LabDeviceMappingDraft{}, &sync.Map{}, schema.NamingStrategy{})
		require.NoError(t, err)
		index := parsed.LookIndex("uq_lab_device_mapping_drafts_id_clinic")
		require.NotNil(t, index)
		require.Equal(t, 2, len(index.Fields))
		assert.Equal(t, "clinic_id", index.Fields[0].DBName)
		assert.Equal(t, "id", index.Fields[1].DBName)
	})
	t.Run("migration", func(t *testing.T) {
		body, err := os.ReadFile("../../migrations/024_lab_device_mapping_drafts_clinic_unique.sql")
		require.NoError(t, err)
		var statements []string
		for _, statement := range strings.Split(string(body), ";") {
			normalized := strings.Join(strings.Fields(statement), " ")
			if normalized != "" {
				statements = append(statements, normalized)
			}
		}
		assert.Equal(t, []string{
			"ALTER TABLE lab_device_mapping_draft_items DROP CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic",
			"ALTER TABLE lab_device_mapping_specs DROP CONSTRAINT fk_lab_device_mapping_specs_draft_clinic",
			"ALTER TABLE lab_device_mapping_drafts DROP CONSTRAINT uq_lab_device_mapping_drafts_id_clinic",
			"ALTER TABLE lab_device_mapping_drafts ADD CONSTRAINT uq_lab_device_mapping_drafts_id_clinic UNIQUE (clinic_id, id)",
			"ALTER TABLE lab_device_mapping_draft_items ADD CONSTRAINT fk_lab_device_mapping_draft_items_draft_clinic FOREIGN KEY (clinic_id, draft_id) REFERENCES lab_device_mapping_drafts (clinic_id, id) ON DELETE RESTRICT",
			"ALTER TABLE lab_device_mapping_specs ADD CONSTRAINT fk_lab_device_mapping_specs_draft_clinic FOREIGN KEY (clinic_id, confirmed_draft_id) REFERENCES lab_device_mapping_drafts (clinic_id, id) ON DELETE RESTRICT",
		}, statements)
	})
}
