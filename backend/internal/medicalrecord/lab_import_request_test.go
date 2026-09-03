package medicalrecord

import (
	"strings"
	"testing"
)

func TestLabImportRevertRequest_RejectsOverMaxReason(t *testing.T) {
	var req labImportRevertRequest
	err := shouldBindJSON(t, map[string]any{"reason": strings.Repeat("r", 501)}, &req)
	if err == nil {
		t.Fatal("ShouldBindJSON() error = nil, want reason max=500 rejection")
	}
}

func TestLabImportRevertRequest_AcceptsAtMaxReason(t *testing.T) {
	var req labImportRevertRequest
	err := shouldBindJSON(t, map[string]any{"reason": strings.Repeat("r", 500)}, &req)
	if err != nil {
		t.Fatalf("ShouldBindJSON() error = %v, want nil at max=500", err)
	}
}
