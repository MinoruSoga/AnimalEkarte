package owner

import "context"

type mockLstepTagSyncService struct {
	syncOwnerAnimalClassificationTagFn func(context.Context, uint64, uint64) error
	syncExclusionTagsFn                func(context.Context, uint64, uint64) error
}

func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(
	ctx context.Context,
	clinicID, ownerID uint64,
) error {
	if m.syncOwnerAnimalClassificationTagFn != nil {
		return m.syncOwnerAnimalClassificationTagFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockLstepTagSyncService) SyncExclusionTags(
	ctx context.Context,
	clinicID, ownerID uint64,
) error {
	if m.syncExclusionTagsFn != nil {
		return m.syncExclusionTagsFn(ctx, clinicID, ownerID)
	}
	return nil
}

type mockAuditService struct {
	logLstepOperationErr error
	logLstepOperationFn  func(
		context.Context,
		uint64,
		*uint64,
		string,
		string,
		*uint64,
	) error
}

func (m *mockAuditService) LogLstepOperation(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action, resource string,
	resourceID *uint64,
) error {
	if m.logLstepOperationFn != nil {
		return m.logLstepOperationFn(ctx, clinicID, actorID, action, resource, resourceID)
	}
	return m.logLstepOperationErr
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}
