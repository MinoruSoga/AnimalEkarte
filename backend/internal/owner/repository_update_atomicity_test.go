package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	record := makeTestOwner(t, db, clinicID, "更新前")
	const callbackName = "owner:update_and_find_reload_failure"
	reloadErr := errors.New("forced reload failure")
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement != nil && query.Statement.Table == "owners" {
			query.AddError(reloadErr)
		}
	}))
	callbackRegistered := true
	t.Cleanup(func() {
		if callbackRegistered {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	got, err := repo.UpdateAndFind(ctx, clinicID, record.ID, map[string]any{"name": "更新後"})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, reloadErr)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRegistered = false

	var persisted model.Owner
	require.NoError(t, db.WithContext(ctx).First(&persisted, record.ID).Error)
	assert.Equal(t, "更新前", persisted.Name)
}
