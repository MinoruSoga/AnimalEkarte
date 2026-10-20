package medicalrecord

import (
	"context"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// masterWriteRowsLocker serializes parent references with parent soft deletion.
// Every writer locks its complete set in ascending order before any row mutation.
type masterWriteRowsLocker interface {
	LockWriteRows(context.Context, uint64, []uint64) error
}

func masterWriteRowIDs(target uint64, parent *uint64) []uint64 {
	ids := make([]uint64, 0, 2)
	if target != 0 {
		ids = append(ids, target)
	}
	if parent != nil && *parent != 0 && *parent != target {
		ids = append(ids, *parent)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func lockMasterWriteRows(ctx context.Context, repo any, clinicID, target uint64, parent *uint64) error {
	ids := masterWriteRowIDs(target, parent)
	if len(ids) == 0 {
		return nil
	}
	locker, ok := repo.(masterWriteRowsLocker)
	if !ok {
		return apperrors.WrapInternalServerError("master parent write lock dependency is required")
	}
	return locker.LockWriteRows(ctx, clinicID, ids)
}
