package persistence

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAmbientTransactionContext(t *testing.T) {
	t.Parallel()

	base := context.WithValue(context.Background(), struct{ name string }{"request"}, "kept")
	tx := &gorm.DB{}
	txCtx := WithTxValue(base, tx)

	assert.Same(t, tx, TxFromContext(txCtx))

	detached := DetachTx(txCtx)
	assert.Nil(t, TxFromContext(detached))
	assert.Equal(t, "kept", detached.Value(struct{ name string }{"request"}))
	assert.NoError(t, detached.Err())
}

func TestPostgresConstraintClassification(t *testing.T) {
	t.Parallel()

	unique := fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505"})
	foreignKey := fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23503"})

	assert.True(t, IsUniqueConstraintErr(unique))
	assert.False(t, IsUniqueConstraintErr(foreignKey))
	assert.True(t, IsFKConstraintErr(foreignKey))
	assert.False(t, IsFKConstraintErr(unique))
	assert.False(t, IsUniqueConstraintErr(errors.New("other")))
	assert.False(t, IsFKConstraintErr(nil))
}

func TestSharedScopeAndJunctionSurfaceIsAvailable(t *testing.T) {
	t.Parallel()

	require.NotNil(t, ClinicScope(1))
	require.NotNil(t, ClinicScopeIn([]uint64{1, 2}))
	require.NotNil(t, MedicalRecordTenantScope("treatments", 1))
	require.NotNil(t, Paginate(1, 20))

	var _ = NewTransactor(&gorm.DB{})
	var _ = DBOrTx
	var _ = FindByIDScoped[struct{}]
	var _ = ReplaceChildRowsByParentID[struct{}]
	var _ = InsertJunctionRowsInBatches[struct{}]
}
