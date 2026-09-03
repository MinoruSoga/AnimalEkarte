package medicalrecord

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestParseAndCanonicalizeCheckupPackage_InvalidInputDoesNotLeakDecoderErrors(t *testing.T) {
	tests := []struct {
		name      string
		raw       []byte
		wantMsg   string
		forbidden []string
	}{
		{
			name:      "unknown JSON field",
			raw:       []byte(`{"namespace":"n","version":"1","clinical_approval_ref":"r","types":[],"fields":[],"evil":true}`),
			wantMsg:   "マニフェストの形式が正しくありません",
			forbidden: []string{"json:", "unknown field", "invalid checkup package manifest"},
		},
		{
			name:      "invalid min_value",
			raw:       checkupPackageManifestJSON(t, "not-a-number", "10.0000"),
			wantMsg:   "健診パッケージの数値範囲が不正です",
			forbidden: []string{`field "dental_score"`, "invalid min_value"},
		},
		{
			name:      "empty min_value",
			raw:       checkupPackageManifestJSON(t, "", "10.0000"),
			wantMsg:   "健診パッケージの数値範囲が不正です",
			forbidden: []string{`field "dental_score"`, "min_value empty string"},
		},
		{
			name:      "invalid max_value",
			raw:       checkupPackageManifestJSON(t, "0.0000", "oops"),
			wantMsg:   "健診パッケージの数値範囲が不正です",
			forbidden: []string{`field "dental_score"`, "invalid max_value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndCanonicalizeCheckupPackage(tt.raw)
			if got != nil {
				t.Fatalf("canonical = %#v, want nil", got)
			}
			assertCheckupPackageFixedInvalidInput(t, err, tt.wantMsg, tt.forbidden...)
		})
	}
}

func checkupPackageManifestJSON(t *testing.T, minValue, maxValue string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"namespace":             "ns.demo",
		"version":               "1.0.0",
		"clinical_approval_ref": "clinical-ref-opaque-001",
		"types": []map[string]any{
			{
				"key": "dental_basic", "name": "Dental Basic", "description": "d",
				"interval": "1y", "target_age": "all", "sort_order": 1, "is_active": true,
			},
		},
		"fields": []map[string]any{
			{
				"key": "dental_score", "type_key": "dental_basic", "name": "Score",
				"field_type": "number", "unit": "pt", "min_value": minValue, "max_value": maxValue,
				"options": []string{}, "is_provisional": false, "sort_order": 1,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return raw
}

func assertCheckupPackageFixedInvalidInput(t *testing.T, err error, wantMessage string, forbiddenSubstrings ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil, want invalid input")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
	got := err.Error()
	if !strings.Contains(got, wantMessage) {
		t.Fatalf("error = %q, want to contain %q", got, wantMessage)
	}
	for _, leak := range forbiddenSubstrings {
		if strings.Contains(got, leak) {
			t.Fatalf("error leaked %q: %q", leak, got)
		}
	}
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T (%v), want AppError", err, err)
	}
	if appErr.Message != wantMessage {
		t.Fatalf("AppError.Message = %q, want %q", appErr.Message, wantMessage)
	}
}
