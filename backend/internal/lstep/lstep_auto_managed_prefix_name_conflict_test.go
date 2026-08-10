package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func uniqueLstepAutoManagedPrefixErr() error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: apperrors.ConstraintLstepAutoManagedPrefix,
	}, "lstep_auto_managed_prefix", "")
}

func TestCreateAutoManagedPrefix_NameConflictMapsDomainCode(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createAutoManagedPrefixFn: func(_ context.Context, _ *model.LstepAutoManagedPrefix) error {
			return uniqueLstepAutoManagedPrefixErr()
		},
	}
	svc := newTagConfigSvc(repo)

	item, err := svc.CreateAutoManagedPrefix(context.Background(), CreateAutoManagedPrefixInput{
		Prefix:   "checkup_",
		Category: "C1",
	})

	require.Error(t, err)
	assert.Nil(t, item)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeLstepAutoManagedPrefixConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "checkup_", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "lstep_auto_managed_prefix '' already exists")
	assert.NotContains(t, err.Error(), "lstep_auto_managed_prefixes_prefix_key")
}

func TestCreateAutoManagedPrefix_GenericDBErrorNotElevated(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createAutoManagedPrefixFn: func(_ context.Context, _ *model.LstepAutoManagedPrefix) error {
			return apperrors.WrapAlreadyExists("lstep_auto_managed_prefix", "")
		},
	}
	svc := newTagConfigSvc(repo)

	_, err := svc.CreateAutoManagedPrefix(context.Background(), CreateAutoManagedPrefixInput{
		Prefix:   "x_",
		Category: "C1",
	})

	require.Error(t, err)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodeLstepAutoManagedPrefixConflict))
	assert.True(t, apperrors.IsAlreadyExists(err))
}
