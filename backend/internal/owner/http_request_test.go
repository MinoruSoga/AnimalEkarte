package owner

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestUpdateOwnerRequest_DMPreferenceNullableJSON(t *testing.T) {
	t.Run("omitted means no update", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{}`), &req))

		input := req.toServiceInput()

		assert.Nil(t, input.DMPreference)
	})

	t.Run("null clears dm_preference", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"dm_preference":null}`), &req))

		input := req.toServiceInput()

		require.NotNil(t, input.DMPreference)
		assert.Nil(t, *input.DMPreference)
	})

	t.Run("false is preserved", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"dm_preference":false}`), &req))

		input := req.toServiceInput()

		require.NotNil(t, input.DMPreference)
		require.NotNil(t, *input.DMPreference)
		assert.False(t, **input.DMPreference)
	})

	t.Run("non-boolean value returns invalid input error", func(t *testing.T) {
		var req updateOwnerRequest
		err := json.Unmarshal([]byte(`{"dm_preference":"yes"}`), &req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "真偽値の形式が正しくありません")
	})
}

func TestUpdateOwnerRequest_BirthDateNullableJSON(t *testing.T) {
	t.Run("omitted means no update", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{}`), &req))

		input := req.toServiceInput()
		fields := buildOwnerUpdate(input)

		assert.Nil(t, input.BirthDate)
		assert.NotContains(t, fields, colBirthDate)
	})

	t.Run("null clears birth_date", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"birth_date":null}`), &req))

		input := req.toServiceInput()
		fields := buildOwnerUpdate(input)

		require.NotNil(t, input.BirthDate)
		assert.Nil(t, *input.BirthDate)
		assert.Contains(t, fields, colBirthDate)
		assert.Nil(t, fields[colBirthDate])
	})

	t.Run("date updates birth_date without zero time", func(t *testing.T) {
		var req updateOwnerRequest
		require.NoError(t, json.Unmarshal([]byte(`{"birth_date":"1990-04-01"}`), &req))

		input := req.toServiceInput()
		fields := buildOwnerUpdate(input)

		require.NotNil(t, input.BirthDate)
		require.NotNil(t, *input.BirthDate)
		assert.Equal(t, "1990-04-01", (**input.BirthDate).Format("2006-01-02"))
		assert.False(t, (**input.BirthDate).IsZero())
		storedBirthDate, ok := fields[colBirthDate].(*time.Time)
		require.True(t, ok)
		assert.Equal(t, "1990-04-01", storedBirthDate.Format("2006-01-02"))
		assert.False(t, storedBirthDate.IsZero())
	})

	t.Run("empty string is rejected instead of producing zero time", func(t *testing.T) {
		var req updateOwnerRequest
		err := json.Unmarshal([]byte(`{"birth_date":""}`), &req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "日付の形式が正しくありません")
	})
}

// ---- newListOwnersQuery ----

func TestNewListOwnersQuery(t *testing.T) {
	t.Run("extracts search param", func(t *testing.T) {
		values := url.Values{"search": []string{"田中"}}
		query, err := newListOwnersQuery(values)
		require.NoError(t, err)
		assert.Equal(t, "田中", query.Search)
	})

	t.Run("returns empty search when absent", func(t *testing.T) {
		query, err := newListOwnersQuery(url.Values{})
		require.NoError(t, err)
		assert.Equal(t, "", query.Search)
	})

	t.Run("rejects search longer than 255", func(t *testing.T) {
		values := url.Values{"search": []string{strings.Repeat("あ", 256)}}
		_, err := newListOwnersQuery(values)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

// ---- createPetForOwnerRequest.toServiceInput ----

func TestCreatePetForOwnerRequest_ToServiceInput(t *testing.T) {
	t.Run("maps all fields", func(t *testing.T) {
		weight := 3.5
		insuranceID := uint64(7)
		birthDate := jsonDate{}
		require.NoError(t, json.Unmarshal([]byte(`"2020-01-15"`), &birthDate))
		neuteredDate := jsonDate{}
		require.NoError(t, json.Unmarshal([]byte(`"2021-06-01"`), &neuteredDate))

		req := createPetForOwnerRequest{
			Name:            "ポチ",
			AnimalSpeciesID: 2,
			NameKana:        "ポチ",
			Breed:           "柴犬",
			Color:           "茶",
			BloodType:       "DEA1.1",
			MicrochipNumber: "123456789012345",
			Gender:          "male",
			Status:          "alive",
			BirthDate:       &birthDate,
			Weight:          &weight,
			NeuteredDate:    &neuteredDate,
			AcquisitionType: "purchase",
			DangerLevel:     "none",
			Food:            "ドライフード",
			Environment:     "屋内",
			InsuranceID:     &insuranceID,
			Remarks:         "特記事項なし",
		}

		input := req.toServiceInput()

		assert.Equal(t, "ポチ", input.Name)
		assert.Equal(t, uint64(2), input.AnimalSpeciesID)
		assert.Equal(t, "ポチ", input.PetNameKana)
		assert.Equal(t, "柴犬", input.Breed)
		assert.Equal(t, "茶", input.Color)
		assert.Equal(t, "DEA1.1", input.BloodType)
		assert.Equal(t, "123456789012345", input.MicrochipNumber)
		assert.Equal(t, "male", input.Gender)
		assert.Equal(t, "alive", input.Status)
		require.NotNil(t, input.BirthDate)
		require.NotNil(t, input.Weight)
		assert.Equal(t, weight, *input.Weight)
		require.NotNil(t, input.NeuteredDate)
		assert.Equal(t, "purchase", input.AcquisitionType)
		assert.Equal(t, "none", input.DangerLevel)
		assert.Equal(t, "ドライフード", input.Food)
		assert.Equal(t, "屋内", input.Environment)
		require.NotNil(t, input.InsuranceID)
		assert.Equal(t, insuranceID, *input.InsuranceID)
		assert.Equal(t, "特記事項なし", input.Remarks)
	})

	t.Run("nil optional pointers stay nil", func(t *testing.T) {
		req := createPetForOwnerRequest{Name: "タマ", AnimalSpeciesID: 1}

		input := req.toServiceInput()

		assert.Nil(t, input.BirthDate)
		assert.Nil(t, input.Weight)
		assert.Nil(t, input.NeuteredDate)
		assert.Nil(t, input.InsuranceID)
	})
}

// ---- createOwnerRequest.toServiceInput ----

func TestCreateOwnerRequest_ToServiceInput(t *testing.T) {
	t.Run("converts pets slice via nested toServiceInput", func(t *testing.T) {
		req := createOwnerRequest{
			OwnerName:      "山田太郎",
			MembershipType: "member",
			Pets: []createPetForOwnerRequest{
				{Name: "ポチ", AnimalSpeciesID: 1},
				{Name: "タマ", AnimalSpeciesID: 2},
			},
		}

		input := req.toServiceInput()

		assert.Equal(t, "山田太郎", input.OwnerName)
		assert.Equal(t, model.MembershipType("member"), input.MembershipType)
		require.Len(t, input.Pets, 2)
		assert.Equal(t, "ポチ", input.Pets[0].Name)
		assert.Equal(t, "タマ", input.Pets[1].Name)
	})

	t.Run("empty pets produces empty (non-nil) slice", func(t *testing.T) {
		req := createOwnerRequest{OwnerName: "鈴木花子"}

		input := req.toServiceInput()

		assert.NotNil(t, input.Pets)
		assert.Empty(t, input.Pets)
	})
}

// ---- updateOwnerRequest.toServiceInput: MembershipType branch ----

func TestUpdateOwnerRequest_ToServiceInput_MembershipType(t *testing.T) {
	t.Run("converts non-nil MembershipType to model pointer", func(t *testing.T) {
		mt := "member"
		req := updateOwnerRequest{MembershipType: &mt}

		input := req.toServiceInput()

		require.NotNil(t, input.MembershipType)
		assert.Equal(t, model.MembershipType("member"), *input.MembershipType)
	})

	t.Run("nil MembershipType stays nil", func(t *testing.T) {
		req := updateOwnerRequest{}

		input := req.toServiceInput()

		assert.Nil(t, input.MembershipType)
	})
}

func TestNullableUint64RequestField_JSON(t *testing.T) {
	t.Run("omitted means unset", func(t *testing.T) {
		var f nullableUint64RequestField
		assert.Nil(t, f.toServiceInput())
	})
	t.Run("null clears", func(t *testing.T) {
		var f nullableUint64RequestField
		require.NoError(t, json.Unmarshal([]byte("null"), &f))
		got := f.toServiceInput()
		require.NotNil(t, got)
		assert.Nil(t, *got)
	})
	t.Run("number sets", func(t *testing.T) {
		var f nullableUint64RequestField
		require.NoError(t, json.Unmarshal([]byte("42"), &f))
		got := f.toServiceInput()
		require.NotNil(t, got)
		require.NotNil(t, *got)
		assert.Equal(t, uint64(42), **got)
	})
}
