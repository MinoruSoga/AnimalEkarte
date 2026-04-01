package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

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
