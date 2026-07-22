package handler

import "context"

// This narrow carrier remains for the real legacy owner deletion route consumer. Delete it
// with OwnerDeletionLifecycle when that HTTP surface moves in BE9-2E.
type mockOwnerDeletionLifecycle struct{}

func (*mockOwnerDeletionLifecycle) HandleOwnerDeletion(context.Context, uint64, uint64) error {
	return nil
}
