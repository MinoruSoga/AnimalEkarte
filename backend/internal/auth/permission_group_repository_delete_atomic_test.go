package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestPermissionGroupRepository_Delete_ConflictsWhenAssigned(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	g := makePermissionGroup(t, db, clinicA, "割当済み削除対象グループ")
	staff := makeDoctor(t, db, clinicA, "割当済みグループスタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffPermissionGroup{StaffID: staff.ID, GroupID: g.ID}).Error)

	err := repo.Delete(ctx, clinicA, g.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

	got, findErr := repo.FindByID(ctx, clinicA, g.ID)
	require.NoError(t, findErr)
	assert.Equal(t, g.ID, got.ID)
}

func TestPermissionGroupRepository_Delete_UnusedSucceeds(t *testing.T) {
	db := setupPermissionGroupRepositoryTestDB(t)
	repo := NewPermissionGroupRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	g := makePermissionGroup(t, db, clinicA, "未割当削除対象グループ")
	require.NoError(t, repo.Delete(ctx, clinicA, g.ID))

	_, err := repo.FindByID(ctx, clinicA, g.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
