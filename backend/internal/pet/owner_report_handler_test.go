package pet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// Keep the shared handler mock compatible with Service without making the
// unrelated legacy handler test file own Owner Report behavior.
func (*mockPetServiceHandler) ListOwnerReportPets(context.Context, []uint64, uint64) ([]model.Pet, error) {
	return nil, nil
}

// mockPetRepository is declared in service_test_helpers_test.go.
func (*mockPetRepository) FindOwnerReportPets(context.Context, []uint64, uint64) ([]model.Pet, error) {
	return nil, nil
}

type ownerReportPetServiceStub struct {
	Service
	listFn func(context.Context, []uint64, uint64) ([]model.Pet, error)
}

func (s *ownerReportPetServiceStub) ListOwnerReportPets(
	ctx context.Context,
	clinicIDs []uint64,
	ownerID uint64,
) ([]model.Pet, error) {
	return s.listFn(ctx, clinicIDs, ownerID)
}

func jsonObjectKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	return keys
}

func TestListOwnerReportPets_ReturnsOnlyCuratedFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	birthDate := time.Date(2021, time.May, 6, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2022, time.June, 7, 0, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2026, time.July, 25, 0, 0, 0, 0, time.UTC)
	bloodType := "DEA 1.1+"
	microchipNumber := "392000000000001"
	weight := 8.4
	acquisitionType := model.AcquisitionTypeRescued

	svc := &ownerReportPetServiceStub{
		listFn: func(_ context.Context, clinicIDs []uint64, ownerID uint64) ([]model.Pet, error) {
			assert.Equal(t, []uint64{1, 2}, clinicIDs)
			assert.Equal(t, uint64(41), ownerID)
			return []model.Pet{{
				ID:              9,
				ClinicID:        1,
				OwnerID:         ownerID,
				AnimalSpeciesID: 3,
				PetNumber:       "41-1",
				Name:            "モモ",
				NameKana:        "モモ",
				Gender:          model.PetGenderFemale,
				Status:          model.PetStatusAlive,
				BirthDate:       &birthDate,
				Breed:           "柴犬",
				Color:           "赤",
				BloodType:       &bloodType,
				MicrochipNumber: &microchipNumber,
				Weight:          &weight,
				NeuteredDate:    &neuteredDate,
				AcquisitionType: &acquisitionType,
				DangerLevel:     model.DangerLevelHigh,
				Food:            "療法食",
				Environment:     "室内",
				Phone:           "090-0000-0000",
				LastVisit:       &lastVisit,
				InsuranceID:     func() *uint64 { id := uint64(5); return &id }(),
				Remarks:         "診療メモ",
				DeceasedReason:  func() *string { reason := "staff only"; return &reason }(),
				AnimalSpecies:   &model.AnimalSpecies{ID: 3, Name: "犬", SortOrder: 1},
				Insurance:       &model.Insurance{ID: 5, ClinicID: 1, Name: "50%保険", CoverageRate: 50, ContactPhone: "03-0000-0000"},
			}}, nil
		},
	}

	h := newHandlerWithPetSvcHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/owners/41/report/pets", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "41"}}
	setClinicID(c)
	c.Set("clinic_ids", []uint64{1, 2})

	h.ListOwnerReportPets(c)

	require.Equal(t, http.StatusOK, w.Code)
	var response map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.ElementsMatch(t, []string{"data"}, jsonObjectKeys(response))

	data, ok := response["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 1)
	item, ok := data[0].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{
		"id",
		"name",
		"pet_name_kana",
		"gender",
		"status",
		"birth_date",
		"breed",
		"color",
		"blood_type",
		"microchip_number",
		"weight",
		"neutered_date",
		"acquisition_type",
		"food",
		"environment",
		"last_visit",
		"remarks",
		"animal_species",
		"insurance",
	}, jsonObjectKeys(item))

	species, ok := item["animal_species"].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"name"}, jsonObjectKeys(species))
	assert.Equal(t, "犬", species["name"])

	insurance, ok := item["insurance"].(map[string]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"name", "coverage_rate"}, jsonObjectKeys(insurance))
	assert.Equal(t, "50%保険", insurance["name"])
	assert.Equal(t, float64(50), insurance["coverage_rate"])
}

func TestListOwnerReportPets_ValidatesBoundaryAndMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(*gin.Context)
		serviceErr error
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "missing clinic context returns unauthorized",
			paramID:    "41",
			setupCtx:   func(*gin.Context) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed owner ID returns bad request",
			paramID:    "not-a-number",
			setupCtx:   setClinicID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown or unauthorized owner returns not found",
			paramID:    "41",
			setupCtx:   setClinicID,
			serviceErr: apperrors.WrapNotFound("owner", "41"),
			wantStatus: http.StatusNotFound,
			wantCalled: true,
		},
		{
			name:       "unexpected service failure returns internal error",
			paramID:    "41",
			setupCtx:   setClinicID,
			serviceErr: fmt.Errorf("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			svc := &ownerReportPetServiceStub{
				listFn: func(context.Context, []uint64, uint64) ([]model.Pet, error) {
					called = true
					return nil, tt.serviceErr
				},
			}
			h := newHandlerWithPetSvcHandler(svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/owners/"+tt.paramID+"/report/pets", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.ListOwnerReportPets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantCalled, called)
		})
	}
}

func TestPetRepository_FindOwnerReportPets_CorrelatesOwnerAndPreloadsToPetClinic(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Insurance{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE insurances CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE animal_species CASCADE").Error)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	targetOwner := makeTestOwner(t, db, clinicA, "レポート対象飼主")
	otherOwner := makeTestOwner(t, db, clinicA, "別飼主")
	localInsurance := testdb.MakeInsurance(t, db, clinicA, "医院A保険")
	crossClinicInsurance := testdb.MakeInsurance(t, db, clinicB, "医院B保険")
	require.NoError(t, db.Model(localInsurance).Update("coverage_rate", 70).Error)
	require.NoError(t, db.Model(crossClinicInsurance).Update("coverage_rate", 90).Error)

	livingPet := testdb.MakePetWithInsurance(
		t, db, clinicA, targetOwner.ID, &localInsurance.ID, "生存ペット",
	)
	deceasedPet := testdb.MakePetWithInsurance(
		t, db, clinicA, targetOwner.ID, nil, "死亡ペット",
	)
	deceasedAt := time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)
	require.NoError(t, db.Model(deceasedPet).Updates(map[string]any{
		"status":      model.PetStatusDeceased,
		"deceased_at": deceasedAt,
	}).Error)
	crossInsurancePet := testdb.MakePetWithInsurance(
		t, db, clinicA, targetOwner.ID, &crossClinicInsurance.ID, "越境保険ペット",
	)
	testdb.MakePetWithInsurance(t, db, clinicA, otherOwner.ID, nil, "別飼主ペット")
	testdb.MakePetWithInsurance(t, db, clinicB, targetOwner.ID, nil, "破損越境飼主ペット")
	deletedPet := testdb.MakePetWithInsurance(
		t, db, clinicA, targetOwner.ID, nil, "論理削除ペット",
	)
	require.NoError(t, db.Delete(deletedPet).Error)

	got, err := repo.FindOwnerReportPets(ctx, []uint64{clinicA, clinicB}, targetOwner.ID)

	require.NoError(t, err)
	gotIDs := make([]uint64, 0, len(got))
	byID := make(map[uint64]model.Pet, len(got))
	for _, pet := range got {
		gotIDs = append(gotIDs, pet.ID)
		byID[pet.ID] = pet
	}
	assert.ElementsMatch(t, []uint64{livingPet.ID, deceasedPet.ID, crossInsurancePet.ID}, gotIDs)

	require.NotNil(t, byID[livingPet.ID].Insurance)
	assert.Equal(t, localInsurance.ID, byID[livingPet.ID].Insurance.ID)
	assert.NotNil(t, byID[livingPet.ID].AnimalSpecies)
	assert.Equal(t, clinicA, byID[deceasedPet.ID].ClinicID, "死亡ペットも同院の対象飼主なら含める")
	assert.Nil(t, byID[crossInsurancePet.ID].Insurance, "認可集合に別院が含まれても pet と clinic が一致しない保険を preload しない")

	unauthorized, err := repo.FindOwnerReportPets(ctx, []uint64{clinicB}, targetOwner.ID)
	assert.Nil(t, unauthorized)
	assert.True(t, apperrors.IsNotFound(err), "認可医院外の飼主は存在を隠して NotFound にする: %v", err)
}
