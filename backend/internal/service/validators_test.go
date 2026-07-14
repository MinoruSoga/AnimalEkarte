package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestValidateRequiredName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "accepts short name", input: "犬", wantErr: false},
		{name: "accepts exactly 255 chars", input: strings.Repeat("あ", MasterNameMaxLength), wantErr: false},
		{name: "rejects empty", input: "", wantErr: true},
		{name: "rejects whitespace only", input: "   ", wantErr: true},
		{name: "rejects 256 chars", input: strings.Repeat("あ", MasterNameMaxLength+1), wantErr: true},
		{name: "rejects null byte", input: "bad\u0000name", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequiredName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOptionalName(t *testing.T) {
	// nil はスキップされる
	assert.NoError(t, validateOptionalName(nil))
	// 非 nil は通常検証
	ok := "犬"
	assert.NoError(t, validateOptionalName(&ok))
	ng := strings.Repeat("X", MasterNameMaxLength+1)
	err := validateOptionalName(&ng)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestValidateEmailFormat(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "accepts valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "accepts email with numbers",
			email:   "user123@example.co.jp",
			wantErr: false,
		},
		{
			name:    "accepts email with plus sign",
			email:   "user+test@example.com",
			wantErr: false,
		},
		{
			name:    "accepts email with hyphen",
			email:   "user-name@example-domain.com",
			wantErr: false,
		},
		{
			name:    "skips validation for empty email",
			email:   "",
			wantErr: false,
		},
		{
			name:    "rejects email without domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "rejects email without @",
			email:   "user.example.com",
			wantErr: true,
		},
		{
			name:    "rejects email with space",
			email:   "user @example.com",
			wantErr: true,
		},
		{
			name:    "rejects email with invalid characters",
			email:   "user#@example.com",
			wantErr: true,
		},
		{
			name:    "rejects email without TLD",
			email:   "user@example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailFormat(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePhoneFormat(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{
			name:    "accepts phone with hyphens (090-1234-5678)",
			phone:   "090-1234-5678",
			wantErr: false,
		},
		{
			name:    "accepts phone without hyphens (09012345678)",
			phone:   "09012345678",
			wantErr: false,
		},
		{
			name:    "accepts landline with hyphens (03-1234-5678)",
			phone:   "03-1234-5678",
			wantErr: false,
		},
		{
			name:    "accepts landline without hyphens (0312345678)",
			phone:   "0312345678",
			wantErr: false,
		},
		{
			name:    "skips validation for empty phone",
			phone:   "",
			wantErr: false,
		},
		{
			name:    "rejects phone without leading 0",
			phone:   "90-1234-5678",
			wantErr: true,
		},
		{
			name:    "rejects phone with letters",
			phone:   "090-ABCD-5678",
			wantErr: true,
		},
		{
			name:    "rejects phone too short",
			phone:   "090-123-456",
			wantErr: true,
		},
		{
			name:    "rejects phone with space",
			phone:   "090 1234 5678",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePhoneFormat(tt.phone)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePostalCodeFormat(t *testing.T) {
	tests := []struct {
		name       string
		postalCode string
		wantErr    bool
	}{
		{
			name:       "accepts postal code with hyphen (123-4567)",
			postalCode: "123-4567",
			wantErr:    false,
		},
		{
			name:       "accepts postal code without hyphen (1234567)",
			postalCode: "1234567",
			wantErr:    false,
		},
		{
			name:       "skips validation for empty postal code",
			postalCode: "",
			wantErr:    false,
		},
		{
			name:       "rejects postal code too short",
			postalCode: "12-3456",
			wantErr:    true,
		},
		{
			name:       "rejects postal code too long",
			postalCode: "1234-56789",
			wantErr:    true,
		},
		{
			name:       "rejects postal code with letters",
			postalCode: "ABC-DEFG",
			wantErr:    true,
		},
		{
			name:       "rejects postal code with space",
			postalCode: "123 4567",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePostalCodeFormat(tt.postalCode)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"valid password", "password123", false},
		{"too short", "pwd12", true},
		{"no letters", "12345678", true},
		{"no digits", "password", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.pw)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTaxType(t *testing.T) {
	assert.NoError(t, validateTaxType(""))
	assert.NoError(t, validateTaxType(string(model.TaxTypeIncluded)))
	assert.Error(t, validateTaxType("invalid_tax"))
}

func TestValidateItemCategory(t *testing.T) {
	// Called by go test for BE7-1; production callers use validateItemCategory via billing_item CreateItem.
	valid := []model.ItemCategory{
		model.ItemCategoryExamination,
		model.ItemCategoryTest,
		model.ItemCategoryProcedure,
		model.ItemCategorySurgery,
		model.ItemCategoryMedicine,
		model.ItemCategoryFood,
		model.ItemCategoryGoods,
		model.ItemCategoryOther,
		model.ItemCategoryVaccine,
		model.ItemCategoryTrimming,
		model.ItemCategoryHotel,
		model.ItemCategoryTraining,
	}
	for _, cat := range valid {
		assert.NoError(t, validateItemCategory(string(cat)), "category %q", cat)
	}
	assert.Error(t, validateItemCategory("unknown"))
	assert.Error(t, validateItemCategory("invalid_category"))
}

func TestValidateItemSource(t *testing.T) {
	assert.NoError(t, validateItemSource(string(model.ItemSourceMedicalRecord)))
	assert.Error(t, validateItemSource("invalid_source"))
}

func TestValidateNonNegativePrice(t *testing.T) {
	assert.NoError(t, validateNonNegativePrice(nil))
	var pos int64 = 100
	assert.NoError(t, validateNonNegativePrice(&pos))
	var neg int64 = -100
	assert.Error(t, validateNonNegativePrice(&neg))
}

func TestValidateMedicalImageType(t *testing.T) {
	assert.NoError(t, validateMedicalImageType(""))
	assert.NoError(t, validateMedicalImageType(string(model.MedicalImageTypeXray)))
	assert.Error(t, validateMedicalImageType("invalid_image_type"))
}

func TestValidateVaccineSpecies(t *testing.T) {
	assert.NoError(t, validateVaccineSpecies(""))
	assert.NoError(t, validateVaccineSpecies(string(model.VaccineSpeciesDog)))
	assert.Error(t, validateVaccineSpecies("invalid_species"))
}

func TestValidateAnesthesiaType(t *testing.T) {
	assert.NoError(t, validateAnesthesiaType(""))
	assert.NoError(t, validateAnesthesiaType(string(model.AnesthesiaTypeNone)))
	assert.Error(t, validateAnesthesiaType("invalid_anesthesia"))
}

func TestValidateCageType(t *testing.T) {
	assert.NoError(t, validateCageType(""))
	assert.NoError(t, validateCageType(string(model.CageTypeICU)))
	assert.Error(t, validateCageType("invalid_cage_type"))
}

func TestValidateCoverageRate(t *testing.T) {
	assert.NoError(t, validateCoverageRate(50))
	assert.Error(t, validateCoverageRate(-1))
	assert.Error(t, validateCoverageRate(101))
}

func TestValidateOptionalCoverageRate(t *testing.T) {
	assert.NoError(t, validateOptionalCoverageRate(nil))
	val := 50
	assert.NoError(t, validateOptionalCoverageRate(&val))
}

func TestValidateCageSize(t *testing.T) {
	assert.NoError(t, validateCageSize(""))
	assert.NoError(t, validateCageSize(string(model.CageSizeSmall)))
	assert.Error(t, validateCageSize("invalid_cage_size"))
}

func TestValidateDiscountRate(t *testing.T) {
	assert.NoError(t, validateDiscountRate(10.5))
	assert.Error(t, validateDiscountRate(-0.1))
	assert.Error(t, validateDiscountRate(100.1))
}

func TestValidateMembershipType(t *testing.T) {
	assert.NoError(t, validateMembershipType(""))
	assert.NoError(t, validateMembershipType(model.MembershipTypeMember))
	assert.Error(t, validateMembershipType("invalid_membership"))
}

func TestValidateCreateOwnerInput(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName:      "owner",
			Email:          "owner@example.com",
			Phone:          "090-1234-5678",
			PostalCode:     "123-4567",
			DiscountRate:   0,
			MembershipType: model.MembershipTypeMember,
		}
		assert.NoError(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName: "",
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid email", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName: "owner",
			Email:     "invalid_email",
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid phone", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName: "owner",
			Phone:     "invalid_phone",
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid postal", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName:  "owner",
			PostalCode: "invalid_postal",
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid discount", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName:    "owner",
			DiscountRate: -1,
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid membership", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName:      "owner",
			MembershipType: "invalid",
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})

	t.Run("invalid pets", func(t *testing.T) {
		input := &CreateOwnerInput{
			OwnerName: "owner",
			Pets: []CreatePetForOwnerInput{
				{Name: "pet", Gender: "invalid"},
			},
		}
		assert.Error(t, validateCreateOwnerInput(input))
	})
}

func TestValidateUpdateOwnerInput(t *testing.T) {
	name := "owner"
	email := "owner@example.com"
	phone := "090-1234-5678"
	postal := "123-4567"
	discount := 10.0
	membership := model.MembershipTypeMember

	t.Run("valid", func(t *testing.T) {
		input := &UpdateOwnerInput{
			OwnerName:      &name,
			Email:          &email,
			Phone:          &phone,
			PostalCode:     &postal,
			DiscountRate:   &discount,
			MembershipType: &membership,
		}
		assert.NoError(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		badName := ""
		input := &UpdateOwnerInput{OwnerName: &badName}
		assert.Error(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid email", func(t *testing.T) {
		badEmail := "bad"
		input := &UpdateOwnerInput{Email: &badEmail}
		assert.Error(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid phone", func(t *testing.T) {
		badPhone := "bad"
		input := &UpdateOwnerInput{Phone: &badPhone}
		assert.Error(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid postal", func(t *testing.T) {
		badPostal := "bad"
		input := &UpdateOwnerInput{PostalCode: &badPostal}
		assert.Error(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid discount", func(t *testing.T) {
		badDiscount := -1.0
		input := &UpdateOwnerInput{DiscountRate: &badDiscount}
		assert.Error(t, validateUpdateOwnerInput(input))
	})

	t.Run("invalid membership", func(t *testing.T) {
		badMembership := model.MembershipType("bad")
		input := &UpdateOwnerInput{MembershipType: &badMembership}
		assert.Error(t, validateUpdateOwnerInput(input))
	})
}

func TestValidatePetGender(t *testing.T) {
	assert.NoError(t, validatePetGender(""))
	assert.NoError(t, validatePetGender(string(model.PetGenderMale)))
	assert.Error(t, validatePetGender("invalid_gender"))
}

func TestValidatePetStatus(t *testing.T) {
	assert.NoError(t, validatePetStatus(""))
	assert.NoError(t, validatePetStatus(string(model.PetStatusAlive)))
	assert.Error(t, validatePetStatus("invalid_status"))
}

func TestValidatePetAcquisitionType(t *testing.T) {
	assert.NoError(t, validatePetAcquisitionType(""))
	assert.NoError(t, validatePetAcquisitionType(string(model.AcquisitionTypePurchase)))
	assert.Error(t, validatePetAcquisitionType("invalid_acquisition"))
}

func TestValidatePetDangerLevel(t *testing.T) {
	assert.NoError(t, validatePetDangerLevel(""))
	assert.NoError(t, validatePetDangerLevel(string(model.DangerLevelLow)))
	assert.Error(t, validatePetDangerLevel("invalid_danger"))
}

func TestValidateCreatePetInput(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		w := 5.5
		input := &CreatePetInput{
			Name:            "pet",
			Weight:          &w,
			Gender:          string(model.PetGenderMale),
			Status:          string(model.PetStatusAlive),
			AcquisitionType: string(model.AcquisitionTypePurchase),
			DangerLevel:     string(model.DangerLevelLow),
		}
		assert.NoError(t, validateCreatePetInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		input := &CreatePetInput{Name: ""}
		assert.Error(t, validateCreatePetInput(input))
	})

	t.Run("invalid weight", func(t *testing.T) {
		w := -1.0
		input := &CreatePetInput{Name: "pet", Weight: &w}
		assert.Error(t, validateCreatePetInput(input))
	})

	t.Run("invalid gender", func(t *testing.T) {
		input := &CreatePetInput{Name: "pet", Gender: "invalid"}
		assert.Error(t, validateCreatePetInput(input))
	})

	t.Run("invalid status", func(t *testing.T) {
		input := &CreatePetInput{Name: "pet", Status: "invalid"}
		assert.Error(t, validateCreatePetInput(input))
	})

	t.Run("invalid acquisition", func(t *testing.T) {
		input := &CreatePetInput{Name: "pet", AcquisitionType: "invalid"}
		assert.Error(t, validateCreatePetInput(input))
	})

	t.Run("invalid danger", func(t *testing.T) {
		input := &CreatePetInput{Name: "pet", DangerLevel: "invalid"}
		assert.Error(t, validateCreatePetInput(input))
	})
}

func TestValidateUpdatePetInput(t *testing.T) {
	name := "pet"
	weight := 5.5
	gender := string(model.PetGenderMale)
	status := string(model.PetStatusAlive)
	acq := string(model.AcquisitionTypePurchase)
	danger := string(model.DangerLevelLow)

	t.Run("valid", func(t *testing.T) {
		input := &UpdatePetInput{
			Name:            &name,
			Weight:          &weight,
			Gender:          &gender,
			Status:          &status,
			AcquisitionType: &acq,
			DangerLevel:     &danger,
		}
		assert.NoError(t, validateUpdatePetInput(input))
	})

	t.Run("invalid name", func(t *testing.T) {
		badName := ""
		input := &UpdatePetInput{Name: &badName}
		assert.Error(t, validateUpdatePetInput(input))
	})

	t.Run("invalid weight", func(t *testing.T) {
		badWeight := -1.0
		input := &UpdatePetInput{Weight: &badWeight}
		assert.Error(t, validateUpdatePetInput(input))
	})

	t.Run("invalid gender", func(t *testing.T) {
		badGender := "bad"
		input := &UpdatePetInput{Gender: &badGender}
		assert.Error(t, validateUpdatePetInput(input))
	})

	t.Run("invalid status", func(t *testing.T) {
		badStatus := "bad"
		input := &UpdatePetInput{Status: &badStatus}
		assert.Error(t, validateUpdatePetInput(input))
	})

	t.Run("invalid acquisition", func(t *testing.T) {
		badAcq := "bad"
		input := &UpdatePetInput{AcquisitionType: &badAcq}
		assert.Error(t, validateUpdatePetInput(input))
	})

	t.Run("invalid danger", func(t *testing.T) {
		badDanger := "bad"
		input := &UpdatePetInput{DangerLevel: &badDanger}
		assert.Error(t, validateUpdatePetInput(input))
	})
}
