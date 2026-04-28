package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- mock LineSendLogRepository ----

type mockLineSendLogRepo struct {
	createFn      func(ctx context.Context, log *model.LineSendLog) error
	findByOwnerFn func(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error)
}

func (m *mockLineSendLogRepo) Create(ctx context.Context, log *model.LineSendLog) error {
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}
func (m *mockLineSendLogRepo) FindByOwner(ctx context.Context, clinicID, ownerID uint64, limit int) ([]*model.LineSendLog, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID, limit)
	}
	return []*model.LineSendLog{}, nil
}

// ---- mock SharedFileService ----

type mockSharedFileSvc struct{}

func (m *mockSharedFileSvc) Upload(_ context.Context, _, _ uint64, _ *UploadSharedFileInput) (*SharedFileResponse, error) {
	return nil, nil
}
func (m *mockSharedFileSvc) GetSignedURL(_ context.Context, _, _ uint64) (string, error) {
	return "", nil
}
func (m *mockSharedFileSvc) FindAll(_ context.Context, _ uint64) ([]*SharedFileResponse, error) {
	return nil, nil
}
func (m *mockSharedFileSvc) Delete(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockSharedFileSvc) CleanupExpired(_ context.Context) error {
	return nil
}

// ---- helpers ----

func newLineSendSvc(ownerRepo repository.OwnerRepository, logRepo repository.LineSendLogRepository) LineSendService {
	return NewLineSendService(
		&mockLstepSettingsService{},
		ownerRepo,
		&mockSharedFileSvc{},
		&mockLstepTagCacheRepository{},
		&mockAuditService{},
		logRepo,
	)
}

// ---- tests ----

func TestGetSendLogs(t *testing.T) {
	tests := []struct {
		name    string
		logRepo *mockLineSendLogRepo
		wantLen int
		wantErr bool
	}{
		{
			name: "success returns logs",
			logRepo: &mockLineSendLogRepo{
				findByOwnerFn: func(_ context.Context, _, _ uint64, _ int) ([]*model.LineSendLog, error) {
					return []*model.LineSendLog{{ID: 1, Status: "sent"}, {ID: 2, Status: "failed"}}, nil
				},
			},
			wantLen: 2,
		},
		{
			name: "repo error",
			logRepo: &mockLineSendLogRepo{
				findByOwnerFn: func(_ context.Context, _, _ uint64, _ int) ([]*model.LineSendLog, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newLineSendSvc(&mockOwnerRepository{}, tt.logRepo)
			logs, err := svc.GetSendLogs(context.Background(), 1, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, logs, tt.wantLen)
			}
		})
	}
}

func TestSend_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 99, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_NoLineUserID(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}

func TestSend_LstepOptOut(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LstepOptOut: true}, nil
		},
	}
	svc := newLineSendSvc(ownerRepo, &mockLineSendLogRepo{})
	_, err := svc.Send(context.Background(), 1, &SendLineMessageInput{OwnerID: 1, MessageType: "text", Text: "hi"})
	assert.Error(t, err)
}
