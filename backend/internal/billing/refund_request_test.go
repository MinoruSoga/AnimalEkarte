package billing

import (
	"strings"
	"testing"
)

func TestCreateRefundRequest_ReasonMax(t *testing.T) {
	t.Run("reason at 500 is accepted", func(t *testing.T) {
		err := bindJSONBody(t, map[string]any{
			"amount": 100,
			"reason": strings.Repeat("a", 500),
		}, &createRefundRequest{})
		if err != nil {
			t.Fatalf("ShouldBindJSON = %v, want nil", err)
		}
	})
	t.Run("reason over 500 is rejected", func(t *testing.T) {
		err := bindJSONBody(t, map[string]any{
			"amount": 100,
			"reason": strings.Repeat("a", 501),
		}, &createRefundRequest{})
		if err == nil {
			t.Fatal("ShouldBindJSON = nil, want over-max error")
		}
	})
	t.Run("omitted reason stays optional", func(t *testing.T) {
		err := bindJSONBody(t, map[string]any{"amount": 100}, &createRefundRequest{})
		if err != nil {
			t.Fatalf("ShouldBindJSON = %v, want nil for optional reason", err)
		}
	})
}
