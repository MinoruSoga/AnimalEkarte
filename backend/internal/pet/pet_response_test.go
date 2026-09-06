package pet

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestResponseIncludesVersion(t *testing.T) {
	body, err := json.Marshal(toResponse(&model.Pet{Version: 7}))
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal(body, &response))
	require.Contains(t, response, "version")
	assert.Equal(t, float64(7), response["version"])
}

func TestToPetListResponseIncludesOwnerReportDetailFields(t *testing.T) {
	birthDate := time.Date(2015, 4, 14, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2016, 5, 20, 0, 0, 0, 0, time.UTC)
	lastVisit := time.Date(2024, 8, 25, 0, 0, 0, 0, time.UTC)
	weight := 7.35
	acquisitionType := model.AcquisitionTypePurchase
	insuranceID := uint64(9)
	bloodType := "DEA1.1陽性"
	microchipNumber := "392140000123456"

	resp := toPetListResponse(&model.Pet{
		ID:              7,
		OwnerID:         42,
		AnimalSpeciesID: 3,
		PetNumber:       "42-1",
		Name:            "ポチ",
		NameKana:        "ぽち",
		Gender:          model.PetGenderMale,
		Status:          model.PetStatusAlive,
		BirthDate:       &birthDate,
		Breed:           "柴犬",
		Color:           "赤",
		BloodType:       &bloodType,
		MicrochipNumber: &microchipNumber,
		Weight:          &weight,
		NeuteredDate:    &neuteredDate,
		AcquisitionType: &acquisitionType,
		DangerLevel:     model.DangerLevelMedium,
		Food:            "療法食",
		Environment:     "室内",
		LastVisit:       &lastVisit,
		InsuranceID:     &insuranceID,
		Remarks:         "咬傷注意",
		AnimalSpecies:   &model.AnimalSpecies{ID: 3, Name: "犬", SortOrder: 1},
		Insurance:       &model.Insurance{ID: insuranceID, Name: "アニコム", CoverageRate: 70, ContactPhone: "03-0000-0000"},
	})

	assert.Equal(t, "柴犬", resp.Breed)
	assert.Equal(t, "赤", resp.Color)
	assert.Equal(t, "DEA1.1陽性", *resp.BloodType)
	assert.Equal(t, "392140000123456", *resp.MicrochipNumber)
	// neutered_date は canonical 規約 (localTimePtr) でローカル化されるため、
	// tz 表現ではなくカレンダー日付で検証する (格納値 2016-05-20 を保持)。
	require.NotNil(t, resp.NeuteredDate)
	assert.Equal(t, "2016-05-20", resp.NeuteredDate.Format("2006-01-02"))
	require.NotNil(t, resp.AcquisitionType)
	assert.Equal(t, "purchased", *resp.AcquisitionType)
	assert.Equal(t, "medium", resp.DangerLevel)
	assert.Equal(t, "療法食", resp.Food)
	assert.Equal(t, "室内", resp.Environment)
	assert.Equal(t, &insuranceID, resp.InsuranceID)
	assert.Equal(t, "咬傷注意", resp.Remarks)
	require.NotNil(t, resp.Insurance)
	assert.Equal(t, "アニコム", resp.Insurance.Name)
	assert.Equal(t, 70, resp.Insurance.CoverageRate)
}

// TestToResponseSerializesDeceasedAt は
// PR#186 P2-2 Bug#1 のリグレッションテスト。死亡記録された pet の
// deceased_at が pet 詳細 (toResponse) で serialize されることを保証する。
// 修正前は両方の response DTO にフィールド自体が存在せず、フロントに値が
// 一切渡らなかった。
// BUG-003: staff 向け Response には deceased_reason も載せる（owner/LIFF DTO は別契約）。
func TestToResponseSerializesDeceasedAt(t *testing.T) {
	deceasedAt := time.Date(2026, 7, 10, 3, 0, 0, 0, time.UTC)
	deceasedReason := "老衰"

	pet := &model.Pet{
		ID:             7,
		OwnerID:        42,
		Name:           "ポチ",
		Status:         model.PetStatusDeceased,
		DeceasedAt:     &deceasedAt,
		DeceasedReason: &deceasedReason,
	}

	detail := toResponse(pet)
	require.NotNil(t, detail.DeceasedAt)
	assert.True(t, deceasedAt.Equal(*detail.DeceasedAt))
	require.NotNil(t, detail.DeceasedReason)
	assert.Equal(t, deceasedReason, *detail.DeceasedReason)

	body, err := json.Marshal(detail)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"deceased_reason":"老衰"`)
}

// TestToResponseOmitsDeceasedAtWhenAlive は、
// 生存中ペット（DeceasedAt が nil）で DeceasedAt が nil のままであることを
// 保証する（誤って死亡日を捏造しない）。
// BUG-003: 生存ペットでは deceased_reason も JSON から物理的に欠落する。
func TestToResponseOmitsDeceasedAtWhenAlive(t *testing.T) {
	pet := &model.Pet{
		ID:      7,
		OwnerID: 42,
		Name:    "ポチ",
		Status:  model.PetStatusAlive,
	}

	detail := toResponse(pet)
	assert.Nil(t, detail.DeceasedAt)
	assert.Nil(t, detail.DeceasedReason)

	body, err := json.Marshal(detail)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "deceased_at")
	assert.NotContains(t, string(body), "deceased_reason")
}
