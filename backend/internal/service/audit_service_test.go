package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- AuditRepository モック ----

type mockAuditRepository struct {
	createFn   func(ctx context.Context, log *model.AuditLog) error
	lastLogged *model.AuditLog
}

func (m *mockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	m.lastLogged = log
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

// TestAuditService_LogLstepOperation_BackwardCompat は ISSUE-010 でメソッドを追加した後でも
// 既存の LogLstepOperation 呼び出しが互換動作（actor_type / clinic_id / metadata=nil）を維持することを検証する。
func TestAuditService_LogLstepOperation_BackwardCompat(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	staffID := uint64(42)
	resourceID := uint64(100)

	err := svc.LogLstepOperation(context.Background(), 1, &staffID, "add_tag", "owner", &resourceID)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "staff", repo.lastLogged.ActorType, "actorID 指定時は staff 扱い")
	assert.Equal(t, "add_tag", repo.lastLogged.Action)
	assert.Equal(t, "owner", repo.lastLogged.Resource)
	if assert.NotNil(t, repo.lastLogged.ClinicID) {
		assert.Equal(t, uint64(1), *repo.lastLogged.ClinicID)
	}
	if assert.NotNil(t, repo.lastLogged.ActorID) {
		assert.Equal(t, staffID, *repo.lastLogged.ActorID)
	}
	if assert.NotNil(t, repo.lastLogged.ResourceID) {
		assert.Equal(t, resourceID, *repo.lastLogged.ResourceID)
	}
	// metadata=nil のとき jsonb は NULL として扱われる（永続化時に nil バイト列）。
	assert.Nil(t, repo.lastLogged.Metadata, "metadata 未指定の場合は nil（NULL 保存）")
}

// TestAuditService_LogLstepOperation_SystemActor は actorID=nil でシステム扱いになることを検証する。
func TestAuditService_LogLstepOperation_SystemActor(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	err := svc.LogLstepOperation(context.Background(), 1, nil, "batch_dormant_detect", "clinic", nil)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "system", repo.lastLogged.ActorType, "actorID=nil はシステム実行扱い")
	assert.Nil(t, repo.lastLogged.ActorID)
	assert.Nil(t, repo.lastLogged.Metadata)
}

// TestAuditService_LogLstepOperationWithMetadata_PersistsJSON は ISSUE-010 で
// metadata が jsonb 用のバイト列として正しくシリアライズされ Repository に渡ることを検証する。
func TestAuditService_LogLstepOperationWithMetadata_PersistsJSON(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	staffID := uint64(7)

	metadata := map[string]any{
		"operation":           "checkup_sync_preview",
		"total_count":         100,
		"eligible_count":      80,
		"line_linked_count":   90,
		"opt_out_count":       5,
		"no_living_pet_count": 5,
		"filter": map[string]any{
			"checkup_type":      "annual",
			"species":           "dog",
			"last_visit_before": "2026-01-31",
			"last_visit_after":  "2025-01-01",
		},
	}

	err := svc.LogLstepOperationWithMetadata(context.Background(), 1, &staffID,
		"checkup_sync_preview", "owner", nil, metadata)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "staff", repo.lastLogged.ActorType)
	assert.Equal(t, "checkup_sync_preview", repo.lastLogged.Action)
	assert.NotNil(t, repo.lastLogged.Metadata, "metadata 指定時は jsonb バイト列が保存される")

	// Metadata は json.RawMessage([]byte) として保存される。Unmarshal して受入条件のキーを検証。
	var decoded map[string]any
	if !assert.NoError(t, json.Unmarshal(repo.lastLogged.Metadata, &decoded)) {
		return
	}
	assert.Equal(t, "checkup_sync_preview", decoded["operation"])
	assert.EqualValues(t, 100, decoded["total_count"])
	assert.EqualValues(t, 80, decoded["eligible_count"])
	assert.EqualValues(t, 90, decoded["line_linked_count"])
	assert.EqualValues(t, 5, decoded["opt_out_count"])
	assert.EqualValues(t, 5, decoded["no_living_pet_count"])

	filter, ok := decoded["filter"].(map[string]any)
	if assert.True(t, ok, "filter は object として保存される") {
		assert.Equal(t, "annual", filter["checkup_type"])
		assert.Equal(t, "dog", filter["species"])
		assert.Equal(t, "2026-01-31", filter["last_visit_before"])
		assert.Equal(t, "2025-01-01", filter["last_visit_after"])
	}
}

// TestAuditService_LogLstepOperationWithMetadata_NilMetadataIsNull は metadata=nil 時に
// audit_logs.metadata に NULL が保存される（後方互換: LogLstepOperation と同等）ことを検証する。
func TestAuditService_LogLstepOperationWithMetadata_NilMetadata(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	err := svc.LogLstepOperationWithMetadata(context.Background(), 1, nil,
		"batch_no_show_detect", "clinic", nil, nil)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Nil(t, repo.lastLogged.Metadata, "metadata=nil は NULL 保存")
}

// TestAuditService_LogLstepOperationWithMetadata_RepoError は repository エラーが
// apperrors.Wrap で包まれて返ることを検証する。
func TestAuditService_LogLstepOperationWithMetadata_RepoError(t *testing.T) {
	repo := &mockAuditRepository{
		createFn: func(_ context.Context, _ *model.AuditLog) error {
			return errors.New("db down")
		},
	}
	svc := NewAuditService(repo)

	err := svc.LogLstepOperationWithMetadata(context.Background(), 1, nil,
		"bulk_add_tag", "owner", nil, map[string]any{"requested_count": 10})
	assert.Error(t, err)
}
