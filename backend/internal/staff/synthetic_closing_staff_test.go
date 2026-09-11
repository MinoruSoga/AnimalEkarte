package staff

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateSyntheticClosingStaff_RejectsUnsafeInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	require.Error(t, CreateSyntheticClosingStaff(ctx, nil, &model.Staff{ClinicID: 9}))
	require.Error(t, CreateSyntheticClosingStaff(ctx, nil, nil))

	err := CreateSyntheticClosingStaff(ctx, nil, &model.Staff{ClinicID: 1})
	require.Error(t, err)
	require.ErrorContains(t, err, "reserved")
}

func TestUnscopedDeleteSyntheticClosingStaffs_RejectsReservedClinic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	require.Error(t, UnscopedDeleteSyntheticClosingStaffs(ctx, nil, 9))
	err := UnscopedDeleteSyntheticClosingStaffs(ctx, nil, 2)
	require.Error(t, err)
	require.ErrorContains(t, err, "reserved")
}
