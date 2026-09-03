package medicalrecord

import (
	"strings"
	"testing"
)

func TestCreateMedicalRecordAddendumRequest_ToServiceInput(t *testing.T) {
	req := CreateMedicalRecordAddendumRequest{AfterText: "after", Reason: "reason"}

	input := req.toServiceInput(1, 2)

	if input.MedicalRecordID != 1 {
		t.Errorf("MedicalRecordID = %d, want 1", input.MedicalRecordID)
	}
	if input.AuthorUserID != 2 {
		t.Errorf("AuthorUserID = %d, want 2", input.AuthorUserID)
	}
	if input.AfterText != req.AfterText {
		t.Errorf("AfterText = %q, want %q", input.AfterText, req.AfterText)
	}
}

func TestCreateMedicalRecordAddendumRequest_RejectsOverMax(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "after_text over max=1000",
			body: map[string]any{"after_text": strings.Repeat("a", 1001), "reason": "reason"},
		},
		{
			name: "reason over max=500",
			body: map[string]any{"after_text": "after", "reason": strings.Repeat("r", 501)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req CreateMedicalRecordAddendumRequest
			if err := shouldBindJSON(t, tt.body, &req); err == nil {
				t.Fatal("ShouldBindJSON() error = nil, want over-max rejection")
			}
		})
	}
}

func TestCreateMedicalRecordAddendumRequest_AcceptsAtMax(t *testing.T) {
	var req CreateMedicalRecordAddendumRequest
	err := shouldBindJSON(t, map[string]any{
		"after_text": strings.Repeat("a", 1000),
		"reason":     strings.Repeat("r", 500),
	}, &req)
	if err != nil {
		t.Fatalf("ShouldBindJSON() error = %v, want nil at max", err)
	}
}
