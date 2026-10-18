package support

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestLLMBudgetValidation(t *testing.T) {
	ctx := context.Background()
	g := NewLLMBudgetGranter(nil)
	valid := []BudgetGrant{{Scope: "global", Model: "m", Interval: time.Second}}
	// nil DB は panic ではなく error
	_, err := g.GrantAll(ctx, valid)
	require.Error(t, err)
	for name, grants := range map[string][]BudgetGrant{
		"empty":        {},
		"no_scope":     {{Scope: "", Model: "m", Interval: time.Second}},
		"no_model":     {{Scope: "s", Model: "", Interval: time.Second}},
		"bad_interval": {{Scope: "s", Model: "m", Interval: 0}},
		"duplicate":    {{Scope: "s", Model: "m", Interval: time.Second}, {Scope: "s", Model: "m", Interval: 2 * time.Second}},
	} {
		_, err := g.GrantAll(ctx, grants)
		assert.Error(t, err, name)
	}
}

func TestEMR410Review_BudgetInvalidGrants(t *testing.T) {
	granter := NewLLMBudgetGranter(&gorm.DB{})
	cases := []struct {
		name   string
		grants []BudgetGrant
	}{
		{name: "empty", grants: nil},
		{name: "empty scope", grants: []BudgetGrant{{Scope: "", Model: "m", Interval: time.Second}}},
		{name: "empty model", grants: []BudgetGrant{{Scope: "s", Model: "", Interval: time.Second}}},
		{name: "zero interval", grants: []BudgetGrant{{Scope: "s", Model: "m", Interval: 0}}},
		{name: "negative interval", grants: []BudgetGrant{{Scope: "s", Model: "m", Interval: -time.Second}}},
		{name: "duplicate scope and model", grants: []BudgetGrant{
			{Scope: "s", Model: "m", Interval: time.Second},
			{Scope: "s", Model: "m", Interval: 2 * time.Second},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := granter.GrantAll(context.Background(), tc.grants)
			require.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
		})
	}
}
