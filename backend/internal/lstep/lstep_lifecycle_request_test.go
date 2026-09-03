package lstep

import (
	"strings"
	"testing"
)

func TestLifecycleReason_BindingMax(t *testing.T) {
	atLimit := strings.Repeat("a", 500)
	overLimit := strings.Repeat("a", 501)

	t.Run("patchPetDeathRequest reason at 500 is accepted", func(t *testing.T) {
		var req patchPetDeathRequest
		err := bindJSONRequest(t, map[string]any{"deceased_at": "2026-01-01", "reason": atLimit}, &req)
		if err != nil {
			t.Fatalf("ShouldBindJSON error = %v, want nil", err)
		}
	})
	t.Run("patchPetDeathRequest reason over 500 is rejected", func(t *testing.T) {
		var req patchPetDeathRequest
		err := bindJSONRequest(t, map[string]any{"deceased_at": "2026-01-01", "reason": overLimit}, &req)
		if err == nil {
			t.Fatal("ShouldBindJSON error = nil, want over-max rejection")
		}
	})
	t.Run("postOwnerLstepOptOutRequest reason at 500 is accepted", func(t *testing.T) {
		var req postOwnerLstepOptOutRequest
		err := bindJSONRequest(t, map[string]any{"reason": atLimit}, &req)
		if err != nil {
			t.Fatalf("ShouldBindJSON error = %v, want nil", err)
		}
	})
	t.Run("postOwnerLstepOptOutRequest reason over 500 is rejected", func(t *testing.T) {
		var req postOwnerLstepOptOutRequest
		err := bindJSONRequest(t, map[string]any{"reason": overLimit}, &req)
		if err == nil {
			t.Fatal("ShouldBindJSON error = nil, want over-max rejection")
		}
	})
	t.Run("patchOwnerLstepOptOutRequest reason at 500 is accepted", func(t *testing.T) {
		var req patchOwnerLstepOptOutRequest
		err := bindJSONRequest(t, map[string]any{"opt_out": true, "reason": atLimit}, &req)
		if err != nil {
			t.Fatalf("ShouldBindJSON error = %v, want nil", err)
		}
	})
	t.Run("patchOwnerLstepOptOutRequest reason over 500 is rejected", func(t *testing.T) {
		var req patchOwnerLstepOptOutRequest
		err := bindJSONRequest(t, map[string]any{"opt_out": true, "reason": overLimit}, &req)
		if err == nil {
			t.Fatal("ShouldBindJSON error = nil, want over-max rejection")
		}
	})
}
