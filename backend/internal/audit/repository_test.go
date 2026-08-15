package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestRepositoryCreateTxRequiresAmbientTransaction(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "explicitly detached context",
			ctx:  persistence.DetachTx(context.Background()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotErr error
			assert.NotPanics(t, func() {
				gotErr = NewRepository(nil).CreateTx(tt.ctx, &model.AuditLog{})
			})

			require.Error(t, gotErr)
			var appErr *apperrors.AppError
			require.ErrorAs(t, gotErr, &appErr)
			assert.Equal(t, "INTERNAL", appErr.Code)
			assert.ErrorContains(t, gotErr, "ambient transaction")
		})
	}
}
