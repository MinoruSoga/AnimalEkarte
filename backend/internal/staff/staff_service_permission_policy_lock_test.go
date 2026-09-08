package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestSetPermissionGroupIDs_PolicyLockPrecedesStaffLockAndFailsClosed(t *testing.T) {
	for _, failLock := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordered", true: "lock failure"}[failLock], func(t *testing.T) {
			var events []string
			lockFailure := errors.New("policy lock failed")
			state := &permissionAssignmentTxState{}
			svc := &staffService{
				tx: rollbackPermissionAssignmentTransactor{state: state},
				repo: &mockStaffRepository{lockInClinicFn: func(context.Context, uint64, uint64) (*model.Staff, error) {
					events = append(events, "staff")
					return &model.Staff{ID: 10, ClinicID: 1}, nil
				}},
				permissionGroupRepo: &mockPermissionGroupRepository{
					lockPermissionPolicyFn: func(ctx context.Context, clinicID uint64) error {
						require.Equal(t, true, ctx.Value(permissionAssignmentTxMarker{}))
						require.Equal(t, uint64(1), clinicID)
						events = append(events, "policy")
						if failLock {
							return lockFailure
						}
						return nil
					},
					updateStaffGroupsFn: func(context.Context, uint64, uint64, []uint64) error {
						events = append(events, "write")
						return nil
					},
				},
				permissionAudit: permissionAssignmentAuditLoggerFunc(func(context.Context, *PermissionAssignmentAuditEntry) error {
					events = append(events, "audit")
					return nil
				}),
			}
			err := svc.SetPermissionGroupIDs(permissionAssignmentAuditContext(), 1, 10, []uint64{1})
			if failLock {
				require.ErrorIs(t, err, lockFailure)
				require.Equal(t, []string{"policy"}, events)
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"policy", "staff", "write", "audit"}, events)
			}
		})
	}
}
