package owner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- LinkLineUserID ----

func TestOwnerServiceLine_LinkLineUserID(t *testing.T) {
	lineUserID := "U1234567890"

	tests := []struct {
		name               string
		lineUserID         *string
		actorUserID        *uint64
		findByIDFn         func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
		findByLineUserIDFn func(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
		updateLineUserIDFn func(ctx context.Context, clinicID, id uint64, lineUserID *string) error
		auditSvc           AuditLogger
		wantErr            bool
		wantConflict       bool
	}{
		{
			name:       "links a new line user id successfully",
			lineUserID: &lineUserID,
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "unlinks when lineUserID is nil",
			lineUserID: nil,
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, gotLineUserID *string) error {
				assert.Nil(t, gotLineUserID)
				return nil
			},
			wantErr: false,
		},
		{
			name: "returns error when owner not found",
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "10")
			},
			lineUserID: &lineUserID,
			wantErr:    true,
		},
		{
			name:       "returns conflict when line user id already linked to a different owner",
			lineUserID: &lineUserID,
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 99}, nil
			},
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:       "does not conflict when the existing owner with this line user id is the same owner",
			lineUserID: &lineUserID,
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return &model.Owner{ID: 10}, nil // same as target id below
			},
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "continues when FindByLineUserID returns typed not found",
			lineUserID: &lineUserID,
			findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "line-user-id")
			},
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:       "returns wrapped error when repository update fails",
			lineUserID: &lineUserID,
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:        "audit log success path with actorUserID is exercised (best-effort, does not affect result)",
			lineUserID:  &lineUserID,
			actorUserID: uint64Ptr(5),
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return nil
			},
			auditSvc: &mockAuditService{},
			wantErr:  false,
		},
		{
			name:        "audit log failure is swallowed (best-effort) and does not propagate",
			lineUserID:  &lineUserID,
			actorUserID: uint64Ptr(5),
			updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
				return nil
			},
			auditSvc: &mockAuditService{logLstepOperationErr: errors.New("audit write failed")},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
					if tt.findByIDFn != nil {
						return tt.findByIDFn(ctx, clinicID, id)
					}
					return &model.Owner{ID: id, ClinicID: clinicID}, nil
				},
				findByLineUserIDFn: tt.findByLineUserIDFn,
				updateLineUserIDFn: tt.updateLineUserIDFn,
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, tt.auditSvc)

			err := svc.LinkLineUserID(context.Background(), 1, 10, tt.lineUserID, tt.actorUserID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestOwnerServiceLine_LinkLineUserID_UnknownLookupErrorFailsClosed(t *testing.T) {
	lookupErr := errors.New("database unavailable")
	updateCalled := false
	lineUserID := "U1234567890"
	repo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Owner, error) {
			return &model.Owner{ID: id, ClinicID: clinicID}, nil
		},
		findByLineUserIDFn: func(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
			return nil, lookupErr
		},
		updateLineUserIDFn: func(_ context.Context, _, _ uint64, _ *string) error {
			updateCalled = true
			return nil
		},
	}
	svc := NewService(repo, nil, nil, nil)

	err := svc.LinkLineUserID(context.Background(), 1, 10, &lineUserID, nil)

	assert.Error(t, err)
	assert.ErrorIs(t, err, lookupErr)
	assert.False(t, updateCalled)
}

// ---- ConfirmLineID ----

func TestOwnerServiceLine_ConfirmLineID_AuditLogging(t *testing.T) {
	lineUserID := "U1234567890"

	tests := []struct {
		name        string
		actorUserID *uint64
		auditSvc    AuditLogger
	}{
		{
			name:        "records audit log with actor id when auditSvc succeeds",
			actorUserID: uint64Ptr(42),
			auditSvc:    &mockAuditService{},
		},
		{
			name:        "swallows audit log failure (best-effort) and still returns the reloaded owner",
			actorUserID: uint64Ptr(42),
			auditSvc:    &mockAuditService{logLstepOperationErr: errors.New("audit write failed")},
		},
		{
			name:        "actorUserID nil is accepted (system confirmation)",
			actorUserID: nil,
			auditSvc:    &mockAuditService{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Owner, error) {
					return &model.Owner{ID: id, ClinicID: 1, LineUserID: &lineUserID}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return nil
				},
			}
			svc := NewService(repo, nil, &mockLstepTagSyncService{}, tt.auditSvc)

			owner, err := svc.ConfirmLineID(context.Background(), 1, 10, tt.actorUserID)
			assert.NoError(t, err)
			assert.NotNil(t, owner)
		})
	}
}
