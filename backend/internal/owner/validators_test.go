package owner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

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
	tests := []struct {
		name    string
		input   CreateOwnerInput
		wantErr bool
	}{
		{
			name: "valid",
			input: CreateOwnerInput{
				OwnerName:      "owner",
				Email:          "owner@example.com",
				Phone:          "090-1234-5678",
				PostalCode:     "123-4567",
				MembershipType: model.MembershipTypeMember,
			},
		},
		{name: "invalid name", input: CreateOwnerInput{}, wantErr: true},
		{name: "invalid email", input: CreateOwnerInput{OwnerName: "owner", Email: "invalid_email"}, wantErr: true},
		{name: "invalid phone", input: CreateOwnerInput{OwnerName: "owner", Phone: "invalid_phone"}, wantErr: true},
		{name: "invalid postal", input: CreateOwnerInput{OwnerName: "owner", PostalCode: "invalid_postal"}, wantErr: true},
		{name: "invalid discount", input: CreateOwnerInput{OwnerName: "owner", DiscountRate: -1}, wantErr: true},
		{
			name:    "invalid membership",
			input:   CreateOwnerInput{OwnerName: "owner", MembershipType: "invalid"},
			wantErr: true,
		},
		{
			name: "invalid nested pet",
			input: CreateOwnerInput{
				OwnerName: "owner",
				Pets:      []CreatePetForOwnerInput{{Name: "pet", Gender: "invalid"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateOwnerInput(&tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateUpdateOwnerInput(t *testing.T) {
	name := "owner"
	email := "owner@example.com"
	phone := "090-1234-5678"
	postal := "123-4567"
	discount := 10.0
	membership := model.MembershipTypeMember

	tests := []struct {
		name    string
		input   UpdateOwnerInput
		wantErr bool
	}{
		{
			name: "valid",
			input: UpdateOwnerInput{
				OwnerName:      &name,
				Email:          &email,
				Phone:          &phone,
				PostalCode:     &postal,
				DiscountRate:   &discount,
				MembershipType: &membership,
			},
		},
		{name: "invalid name", input: UpdateOwnerInput{OwnerName: ptrString("")}, wantErr: true},
		{name: "invalid email", input: UpdateOwnerInput{Email: ptrString("bad")}, wantErr: true},
		{name: "invalid phone", input: UpdateOwnerInput{Phone: ptrString("bad")}, wantErr: true},
		{name: "invalid postal", input: UpdateOwnerInput{PostalCode: ptrString("bad")}, wantErr: true},
		{
			name:    "invalid discount",
			input:   UpdateOwnerInput{DiscountRate: func() *float64 { v := -1.0; return &v }()},
			wantErr: true,
		},
		{
			name: "invalid membership",
			input: UpdateOwnerInput{
				MembershipType: func() *model.MembershipType {
					v := model.MembershipType("bad")
					return &v
				}(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateOwnerInput(&tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateNestedPetEnums(t *testing.T) {
	tests := []struct {
		name     string
		validate func(string) error
		valid    string
		invalid  string
	}{
		{
			name:     "gender",
			validate: validatePetGender,
			valid:    string(model.PetGenderMale),
			invalid:  "invalid_gender",
		},
		{
			name:     "status",
			validate: validatePetStatus,
			valid:    string(model.PetStatusAlive),
			invalid:  "invalid_status",
		},
		{
			name:     "acquisition type",
			validate: validatePetAcquisitionType,
			valid:    string(model.AcquisitionTypePurchase),
			invalid:  "invalid_acquisition",
		},
		{
			name:     "danger level",
			validate: validatePetDangerLevel,
			valid:    string(model.DangerLevelLow),
			invalid:  "invalid_danger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.validate(""))
			assert.NoError(t, tt.validate(tt.valid))
			assert.Error(t, tt.validate(tt.invalid))
		})
	}
}

func TestOwnerContactValidators(t *testing.T) {
	tests := []struct {
		name     string
		validate func(string) error
		valid    []string
		invalid  []string
	}{
		{
			name:     "email",
			validate: validateEmailFormat,
			valid:    []string{"", "owner@example.com"},
			invalid:  []string{"owner", "owner@example"},
		},
		{
			name:     "phone",
			validate: validatePhoneFormat,
			valid:    []string{"", "03-1234-5678", "09012345678"},
			invalid:  []string{"123", "phone"},
		},
		{
			name:     "postal code",
			validate: validatePostalCodeFormat,
			valid:    []string{"", "123-4567", "1234567"},
			invalid:  []string{"123", "postal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, value := range tt.valid {
				assert.NoError(t, tt.validate(value))
			}
			for _, value := range tt.invalid {
				assert.Error(t, tt.validate(value))
			}
		})
	}
}
