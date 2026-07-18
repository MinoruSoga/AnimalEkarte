package handler

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{
			name:    "valid password with letters and digits",
			pw:      "abcd1234",
			wantErr: false,
		},
		{
			name:    "valid password longer than minimum",
			pw:      "password123456",
			wantErr: false,
		},
		{
			name:    "empty password is too short",
			pw:      "",
			wantErr: true,
		},
		{
			name:    "shorter than 8 characters",
			pw:      "abc123",
			wantErr: true,
		},
		{
			name:    "exactly 8 characters is valid",
			pw:      "ab12cd34",
			wantErr: false,
		},
		{
			name:    "letters only, no digits",
			pw:      "abcdefgh",
			wantErr: true,
		},
		{
			name:    "digits only, no letters",
			pw:      "12345678",
			wantErr: true,
		},
		{
			name:    "japanese letters count as letters, still needs a digit",
			pw:      "ぱすわーど",
			wantErr: true,
		},
		{
			name:    "japanese letters plus digit is valid",
			pw:      "ぱすわーど1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.pw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("validatePassword() returned nil error, want error")
				}
				if !apperrors.IsInvalidInput(err) {
					t.Fatalf("validatePassword() error = %v, want invalid input error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validatePassword() returned error: %v, want nil", err)
			}
		})
	}
}
