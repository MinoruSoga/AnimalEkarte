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

func TestAuditService_LogEntry(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)
	clinicID := uint64(1)
	actorID := uint64(2)
	resourceID := uint64(3)

	err := svc.LogEntry(context.Background(), &AuditLogInput{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  "staff",
		Action:     "update",
		Resource:   "permission_group",
		ResourceID: &resourceID,
		OldValue:   map[string]any{"name": "old"},
		NewValue:   map[string]any{"name": "new"},
		IPAddress:  "127.0.0.1",
		UserAgent:  "test-agent",
	})
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "staff", repo.lastLogged.ActorType)
	assert.Equal(t, "update", repo.lastLogged.Action)
	assert.Equal(t, "permission_group", repo.lastLogged.Resource)
	assert.Equal(t, "127.0.0.1", repo.lastLogged.IPAddress)
	assert.NotNil(t, repo.lastLogged.OldValue)
	assert.NotNil(t, repo.lastLogged.NewValue)
}

func TestAuditService_LogEntry_Validation(t *testing.T) {
	tests := []struct {
		name  string
		input *AuditLogInput
	}{
		{
			name: "clinic_id missing",
			input: &AuditLogInput{
				ActorID:   ptrUint64ForAuditTest(2),
				ActorType: model.AuditActorTypeStaff,
				Action:    "update",
				Resource:  "permission_group",
			},
		},
		{
			name: "staff actor requires actor_id",
			input: &AuditLogInput{
				ClinicID:  ptrUint64ForAuditTest(1),
				ActorType: model.AuditActorTypeStaff,
				Action:    "update",
				Resource:  "permission_group",
			},
		},
		{
			name: "system actor must not have actor_id",
			input: &AuditLogInput{
				ClinicID:  ptrUint64ForAuditTest(1),
				ActorID:   ptrUint64ForAuditTest(2),
				ActorType: model.AuditActorTypeSystem,
				Action:    "batch",
				Resource:  "owner",
			},
		},
		{
			name: "unknown actor type",
			input: &AuditLogInput{
				ClinicID:  ptrUint64ForAuditTest(1),
				ActorID:   ptrUint64ForAuditTest(2),
				ActorType: "account",
				Action:    "update",
				Resource:  "permission_group",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepository{}
			svc := NewAuditService(repo)

			err := svc.LogEntry(context.Background(), tt.input)

			assert.Error(t, err)
			assert.Nil(t, repo.lastLogged, "invalid audit input must not reach repository")
		})
	}
}

func TestAuditService_LogAuthLogin_RequiresClinicAndStaff(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)
	staffID := uint64(2)

	err := svc.LogAuthLogin(context.Background(), nil, &staffID, model.AuditActionAuthLoginFailure, "127.0.0.1", "test-agent")

	assert.Error(t, err)
	assert.Nil(t, repo.lastLogged)
}

func ptrUint64ForAuditTest(v uint64) *uint64 {
	return &v
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

// TestAuditService_LogMedicalRecordChange_Create は create アクションで
// resource="medical_record", old_value=nil, new_value 設定の確認。
func TestAuditService_LogMedicalRecordChange_Create(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	staffID := uint64(10)
	recordID := uint64(99)
	newValue := map[string]any{"status": "draft", "doctor_id": uint64(5)}

	err := svc.LogMedicalRecordChange(context.Background(), 1, &staffID, "create", recordID, nil, newValue)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "medical_record", repo.lastLogged.Resource)
	assert.Equal(t, "create", repo.lastLogged.Action)
	assert.Equal(t, "staff", repo.lastLogged.ActorType)
	if assert.NotNil(t, repo.lastLogged.ResourceID) {
		assert.Equal(t, recordID, *repo.lastLogged.ResourceID)
	}
	assert.Nil(t, repo.lastLogged.OldValue, "create 時は old_value=nil")
	assert.NotNil(t, repo.lastLogged.NewValue, "create 時は new_value に値がある")
}

// TestAuditService_LogMedicalRecordChange_Update は update アクションで
// old_value / new_value 両方にデータが入ることを確認。
func TestAuditService_LogMedicalRecordChange_Update(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	staffID := uint64(10)
	recordID := uint64(99)
	oldValue := map[string]any{"status": "draft"}
	newValue := map[string]any{"status": "finalized"}

	err := svc.LogMedicalRecordChange(context.Background(), 1, &staffID, "update", recordID, oldValue, newValue)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "update", repo.lastLogged.Action)
	assert.NotNil(t, repo.lastLogged.OldValue)
	assert.NotNil(t, repo.lastLogged.NewValue)

	var decodedOld, decodedNew map[string]any
	assert.NoError(t, json.Unmarshal(repo.lastLogged.OldValue, &decodedOld))
	assert.NoError(t, json.Unmarshal(repo.lastLogged.NewValue, &decodedNew))
	assert.Equal(t, "draft", decodedOld["status"])
	assert.Equal(t, "finalized", decodedNew["status"])
}

// TestAuditService_LogMedicalRecordChange_SystemActor は actorID=nil でシステム扱いになることを確認。
func TestAuditService_LogMedicalRecordChange_SystemActor(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	err := svc.LogMedicalRecordChange(context.Background(), 1, nil, "delete", 99, map[string]any{"status": "draft"}, nil)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "system", repo.lastLogged.ActorType)
	assert.Nil(t, repo.lastLogged.ActorID)
}

// TestAuditService_LogVitalChange は vital リソースの監査ログに medical_record_id が
// metadata に含まれることを確認する（AUDIT-H1）。
func TestAuditService_LogVitalChange(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	staffID := uint64(20)
	vitalID := uint64(55)
	medicalRecordID := uint64(77)

	err := svc.LogVitalChange(context.Background(), 1, &staffID, "create", vitalID, medicalRecordID, nil, map[string]any{"weight": 3.5})
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "vital", repo.lastLogged.Resource)
	assert.Equal(t, "create", repo.lastLogged.Action)
	if assert.NotNil(t, repo.lastLogged.ResourceID) {
		assert.Equal(t, vitalID, *repo.lastLogged.ResourceID)
	}
	assert.NotNil(t, repo.lastLogged.Metadata, "metadata に medical_record_id を含む")

	var meta map[string]any
	assert.NoError(t, json.Unmarshal(repo.lastLogged.Metadata, &meta))
	assert.EqualValues(t, medicalRecordID, meta["medical_record_id"])
}

// TestAuditService_LogClinicSwitch_StaffActor はスタッフによるクリニック切替で
// resource="auth", action="switch_clinic", actor_type="staff" かつ
// old/new_value に clinic_id が含まれることを確認する（FEAT-374 Phase 2）。
func TestAuditService_LogClinicSwitch_StaffActor(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	actorID := uint64(5)
	fromClinicID := uint64(1)
	toClinicID := uint64(2)

	err := svc.LogClinicSwitch(context.Background(), &actorID, fromClinicID, toClinicID, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "auth", repo.lastLogged.Resource)
	assert.Equal(t, "switch_clinic", repo.lastLogged.Action)
	assert.Equal(t, "staff", repo.lastLogged.ActorType)
	if assert.NotNil(t, repo.lastLogged.ActorID) {
		assert.Equal(t, actorID, *repo.lastLogged.ActorID)
	}
	if assert.NotNil(t, repo.lastLogged.ClinicID) {
		assert.Equal(t, toClinicID, *repo.lastLogged.ClinicID)
	}
	assert.Equal(t, "192.168.1.1", repo.lastLogged.IPAddress)
	assert.Equal(t, "Mozilla/5.0", repo.lastLogged.UserAgent)

	var oldVal, newVal map[string]any
	if assert.NotNil(t, repo.lastLogged.OldValue) {
		assert.NoError(t, json.Unmarshal(repo.lastLogged.OldValue, &oldVal))
		assert.EqualValues(t, fromClinicID, oldVal["clinic_id"])
	}
	if assert.NotNil(t, repo.lastLogged.NewValue) {
		assert.NoError(t, json.Unmarshal(repo.lastLogged.NewValue, &newVal))
		assert.EqualValues(t, toClinicID, newVal["clinic_id"])
	}
}

// TestAuditService_LogClinicSwitch_SystemActor は actorID=nil でシステム扱いになることを確認する。
func TestAuditService_LogClinicSwitch_SystemActor(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	err := svc.LogClinicSwitch(context.Background(), nil, 1, 3, "", "")
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "system", repo.lastLogged.ActorType)
	assert.Nil(t, repo.lastLogged.ActorID)
}

// TestAuditActionConstants_Billing は Issue #122 で追加した billing 系 action 定数の
// 存在・値・{resource}.{operation} 形式を検証する。
func TestAuditActionConstants_Billing(t *testing.T) {
	assert.Equal(t, "billing.cancel", model.AuditActionBillingCancel)
	assert.Equal(t, "billing.post_close_edit", model.AuditActionBillingPostCloseEdit)
	assert.Equal(t, "billing_refund.create", model.AuditActionBillingRefundCreate)
}

// TestAuditService_LogAddendumCreate は addendum の監査ログで
// new_value に before_text/after_text/reason が含まれることを確認する（AUDIT-H1）。
func TestAuditService_LogAddendumCreate(t *testing.T) {
	repo := &mockAuditRepository{}
	svc := NewAuditService(repo)

	actorID := uint64(30)
	addendumID := uint64(11)
	medicalRecordID := uint64(22)
	addendum := &model.MedicalRecordAddendum{
		MedicalRecordID: medicalRecordID,
		AuthorUserID:    actorID,
		BeforeText:      "before content",
		AfterText:       "after content",
		Reason:          "correction",
	}

	err := svc.LogAddendumCreate(context.Background(), 1, &actorID, addendumID, medicalRecordID, addendum)
	assert.NoError(t, err)

	if !assert.NotNil(t, repo.lastLogged) {
		return
	}
	assert.Equal(t, "medical_record_addendum", repo.lastLogged.Resource)
	assert.Equal(t, "create", repo.lastLogged.Action)
	if assert.NotNil(t, repo.lastLogged.ResourceID) {
		assert.Equal(t, addendumID, *repo.lastLogged.ResourceID)
	}
	assert.Nil(t, repo.lastLogged.OldValue, "addendum create は old_value なし")
	assert.NotNil(t, repo.lastLogged.NewValue)

	var newVal map[string]any
	assert.NoError(t, json.Unmarshal(repo.lastLogged.NewValue, &newVal))
	assert.Equal(t, "before content", newVal["before_text"])
	assert.Equal(t, "after content", newVal["after_text"])
	assert.Equal(t, "correction", newVal["reason"])

	assert.NotNil(t, repo.lastLogged.Metadata)
	var meta map[string]any
	assert.NoError(t, json.Unmarshal(repo.lastLogged.Metadata, &meta))
	assert.EqualValues(t, medicalRecordID, meta["medical_record_id"])
}
