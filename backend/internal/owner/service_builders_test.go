package owner

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestBuildOwnerUpdate(t *testing.T) {
	t.Run("empty input produces empty field map", func(t *testing.T) {
		fields := buildOwnerUpdate(&UpdateOwnerInput{})
		assert.Empty(t, fields)
	})

	t.Run("all fields set are mapped to their column names", func(t *testing.T) {
		name := "山田太郎"
		nameKana := "ヤマダタロウ"
		birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		company := "株式会社テスト"
		postalCode := "100-0001"
		address1 := "東京都千代田区"
		address2 := "1-1-1"
		homePostalCode := "200-0001"
		homeAddress1 := "神奈川県横浜市"
		homeAddress2 := "2-2-2"
		phone := "090-1234-5678"
		companyPhone := "03-1234-5678"
		email := "yamada@example.com"
		remarks := "備考"
		isDangerous := true
		discountRate := 10.5
		membershipType := model.MembershipTypeMember
		dmPref := true
		dmPreferencePtr := &dmPref
		birthDatePtr := &birthDate

		input := &UpdateOwnerInput{
			OwnerName:      &name,
			OwnerNameKana:  &nameKana,
			BirthDate:      &birthDatePtr,
			Company:        &company,
			PostalCode:     &postalCode,
			Address1:       &address1,
			Address2:       &address2,
			HomePostalCode: &homePostalCode,
			HomeAddress1:   &homeAddress1,
			HomeAddress2:   &homeAddress2,
			Phone:          &phone,
			CompanyPhone:   &companyPhone,
			Email:          &email,
			Remarks:        &remarks,
			IsDangerous:    &isDangerous,
			DiscountRate:   &discountRate,
			MembershipType: &membershipType,
			DMPreference:   &dmPreferencePtr,
		}
		fields := buildOwnerUpdate(input)

		assert.Equal(t, name, fields[colOwnerName])
		assert.Equal(t, nameKana, fields[colOwnerNameKana])
		require.NotNil(t, fields[colBirthDate])
		assert.Equal(t, birthDate, *fields[colBirthDate].(*time.Time))
		assert.False(t, fields[colBirthDate].(*time.Time).IsZero())
		assert.Equal(t, company, fields[colCompany])
		assert.Equal(t, postalCode, fields[colPostalCode])
		assert.Equal(t, address1, fields[colAddress1])
		assert.Equal(t, address2, fields[colAddress2])
		assert.Equal(t, homePostalCode, fields[colHomePostalCode])
		assert.Equal(t, homeAddress1, fields[colHomeAddress1])
		assert.Equal(t, homeAddress2, fields[colHomeAddress2])
		assert.Equal(t, phone, fields[colPhone])
		assert.Equal(t, companyPhone, fields[colCompanyPhone])
		assert.Equal(t, email, fields[colEmail])
		assert.Equal(t, remarks, fields[colRemarks])
		assert.Equal(t, isDangerous, fields[colIsDangerous])
		assert.Equal(t, discountRate, fields[colDiscountRate])
		assert.Equal(t, membershipType, fields[colMembershipType])
		assert.Equal(t, dmPref, *fields[colDMPreference].(*bool))
	})

	t.Run("DMPreference pointing at a nil *bool clears the column", func(t *testing.T) {
		input := &UpdateOwnerInput{}
		var nilBoolPtr *bool
		input.DMPreference = &nilBoolPtr

		fields := buildOwnerUpdate(input)

		assert.Contains(t, fields, colDMPreference)
		assert.Nil(t, fields[colDMPreference])
	})

	t.Run("BirthDate pointing at a nil time pointer clears the column", func(t *testing.T) {
		input := &UpdateOwnerInput{}
		var nilTimePtr *time.Time
		input.BirthDate = &nilTimePtr

		fields := buildOwnerUpdate(input)

		assert.Contains(t, fields, colBirthDate)
		assert.Nil(t, fields[colBirthDate])
	})
}

func TestNormalizeOwnerReason(t *testing.T) {
	t.Run("trims and returns pointer to normal reason", func(t *testing.T) {
		normalized, err := normalizeOwnerReason("  引っ越しのため  ", "reason")
		require.NoError(t, err)
		require.NotNil(t, normalized)
		assert.Equal(t, "引っ越しのため", *normalized)
	})

	t.Run("empty (or whitespace-only) reason normalizes to nil", func(t *testing.T) {
		normalized, err := normalizeOwnerReason("   ", "reason")
		require.NoError(t, err)
		assert.Nil(t, normalized)
	})

	t.Run("reason exceeding max length is rejected", func(t *testing.T) {
		tooLong := strings.Repeat("あ", ownerDeliveryReasonMaxLength+1)
		normalized, err := normalizeOwnerReason(tooLong, "reason")
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, normalized)
	})

	t.Run("reason at exactly max length is accepted", func(t *testing.T) {
		exact := strings.Repeat("あ", ownerDeliveryReasonMaxLength)
		normalized, err := normalizeOwnerReason(exact, "reason")
		require.NoError(t, err)
		require.NotNil(t, normalized)
		assert.Equal(t, exact, *normalized)
	})
}
