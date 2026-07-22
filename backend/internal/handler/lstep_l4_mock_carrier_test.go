package handler

// L④ moved the lifecycle-owned tests to internal/lstep. This no-op carrier is
// retained for legacy owner-handler tests until that consumer migrates.

import (
	"context"
	"time"
)

type mockLstepLifecycleService struct{}

func (*mockLstepLifecycleService) HandlePetDeath(context.Context, uint64, uint64, time.Time, string, *uint64) error {
	return nil
}
func (*mockLstepLifecycleService) HandlePetRevival(context.Context, uint64, uint64, *uint64) error {
	return nil
}
func (*mockLstepLifecycleService) HandleOwnerOptOut(context.Context, uint64, uint64, string) error {
	return nil
}
func (*mockLstepLifecycleService) HandleOwnerOptIn(context.Context, uint64, uint64) error {
	return nil
}
func (*mockLstepLifecycleService) HandleOwnerDeletion(context.Context, uint64, uint64) error {
	return nil
}
