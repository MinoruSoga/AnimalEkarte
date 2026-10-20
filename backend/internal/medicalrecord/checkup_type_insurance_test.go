package medicalrecord

// EMR-456: 健診種別の保険設定（is_non_insurance）の service / request / response 契約。
// NULL=未保存（旧欠落）、true=保険対象外、false=保険対象。新規初期値は保険対象外を API が明示保存し、
// 更新は省略=変更なし・明示 null=400・boolean=保存（null 契約の統一）。更新の version は必須で、
// 不一致は部分保存せず Conflict。設定変更は同一 tx で監査する。

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestCheckupTypeService_Create_InsuranceSetting(t *testing.T) {
	tests := []struct {
		name  string
		input *bool
		want  bool
	}{
		{name: "unspecified saves new-entry default non-insurance (true)", input: nil, want: true},
		{name: "explicit false (insurance eligible) is kept", input: boolPtr(false), want: false},
		{name: "explicit true (non-insurance) is kept", input: boolPtr(true), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var persisted *model.CheckupType
			repo := &mockCheckupTypeRepository{
				createFn: func(_ context.Context, ct *model.CheckupType) error {
					ct.ID = 5
					ct.Version = 1
					persisted = ct
					return nil
				},
			}
			audit := &recordingCheckupTypeAudit{}
			svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, audit)

			got, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{
				Name: "定期健診", IsNonInsurance: tt.input, ActorID: &testCheckupTypeActorID,
			})

			require.NoError(t, err)
			require.NotNil(t, persisted.IsNonInsurance, "create must never persist NULL")
			assert.Equal(t, tt.want, *persisted.IsNonInsurance)
			require.NotNil(t, got.IsNonInsurance)
			assert.Equal(t, tt.want, *got.IsNonInsurance)
			require.Len(t, audit.entries, 1)
			entry := audit.entries[0]
			assert.Equal(t, checkupTypeAuditActionCreate, entry.Action)
			assert.Equal(t, checkupTypeAuditResource, entry.Resource)
			assert.Equal(t, testCheckupTypeActorID, *entry.ActorID)
			assert.Equal(t, uint64(1), *entry.ClinicID)
			assert.Equal(t, uint64(5), *entry.ResourceID)
			assert.Nil(t, entry.OldValue)
			assert.Equal(t, persisted.IsNonInsurance, entry.NewValue.(map[string]any)["is_non_insurance"])
		})
	}
}

func TestCheckupTypeService_Create_FailClosedAudit(t *testing.T) {
	tests := []struct {
		name    string
		audit   AuditTxLogger
		tx      Transactor
		actorID *uint64
		check   func(t *testing.T, err error)
	}{
		{
			name: "missing actor is rejected", audit: &recordingCheckupTypeAudit{}, tx: checkupTypePassthroughTx{},
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsInvalidInput(err)) },
		},
		{
			name: "audit write failure fails the create", audit: &recordingCheckupTypeAudit{err: errors.New("audit down")},
			tx: checkupTypePassthroughTx{}, actorID: &testCheckupTypeActorID,
			check: func(t *testing.T, err error) { assert.Error(t, err) },
		},
		{
			name: "nil audit dependency is a server error", audit: nil, tx: checkupTypePassthroughTx{}, actorID: &testCheckupTypeActorID,
			check: func(t *testing.T, err error) { assertCheckupTypeInternalError(t, err) },
		},
		{
			name: "nil transactor is a server error before any write", audit: &recordingCheckupTypeAudit{}, tx: nil, actorID: &testCheckupTypeActorID,
			check: func(t *testing.T, err error) { assertCheckupTypeInternalError(t, err) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			repo := &mockCheckupTypeRepository{createFn: func(_ context.Context, _ *model.CheckupType) error {
				created = true
				return nil
			}}
			svc := NewCheckupTypeService(repo, tt.tx, tt.audit)
			got, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "x", ActorID: tt.actorID})
			assert.Nil(t, got)
			tt.check(t, err)
			if tt.tx == nil {
				assert.False(t, created)
			}
		})
	}
}

func TestCheckupTypeService_Update_InsuranceSettingAndVersion(t *testing.T) {
	name := "改名"
	tests := []struct {
		name         string
		stored       *bool
		input        UpdateCheckupTypeInput
		wantField    bool
		wantValue    bool
		wantAudit    bool
		wantExpected int64
	}{
		{name: "unspecified keeps stored value",
			stored: boolPtr(true), input: UpdateCheckupTypeInput{Name: &name, Version: int64Ptr(3)}, wantExpected: 3},
		{name: "unspecified keeps legacy NULL (no default re-applied)",
			stored: nil, input: UpdateCheckupTypeInput{Name: &name, Version: int64Ptr(3)}, wantExpected: 3},
		{name: "explicit false on legacy NULL is saved and audited",
			stored: nil, input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(3)},
			wantField: true, wantValue: false, wantAudit: true, wantExpected: 3},
		{name: "explicit true over false is saved and audited",
			stored: boolPtr(false), input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(true), Version: int64Ptr(3)},
			wantField: true, wantValue: true, wantAudit: true, wantExpected: 3},
		{name: "same value is saved without an insurance audit",
			stored: boolPtr(true), input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(true), Version: int64Ptr(3)},
			wantField: true, wantValue: true, wantExpected: 3},
		{name: "client version is used as the expected version",
			stored: boolPtr(true), input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(2)},
			wantField: true, wantValue: false, wantAudit: true, wantExpected: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := &model.CheckupType{ID: 9, ClinicID: 1, Name: "健診", IsNonInsurance: tt.stored, Version: 3}
			var gotCmd UpdateCheckupTypeInput
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					cp := *stored
					return &cp, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error) {
					gotCmd = cmd
					after := *stored
					if v, ok := buildCheckupTypeUpdate(&cmd)[colCheckupTypeNonIns]; ok {
						b := v.(bool)
						after.IsNonInsurance = &b
					}
					return &after, nil
				},
			}
			audit := &recordingCheckupTypeAudit{}
			svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, audit)
			input := tt.input
			input.ActorID = &testCheckupTypeActorID

			got, err := svc.Update(context.Background(), 1, 9, &input)

			require.NoError(t, err)
			fields := buildCheckupTypeUpdate(&gotCmd)
			v, ok := fields[colCheckupTypeNonIns]
			assert.Equal(t, tt.wantField, ok)
			if tt.wantField {
				assert.Equal(t, tt.wantValue, v)
			} else {
				assert.Equal(t, tt.stored, got.IsNonInsurance)
			}
			_, writesVersion := fields["version"]
			assert.False(t, writesVersion, "version is maintained by the storage trigger, never SET by the app")
			require.NotNil(t, gotCmd.expectedVersion)
			assert.Equal(t, tt.wantExpected, *gotCmd.expectedVersion)
			if tt.wantAudit {
				require.Len(t, audit.entries, 1)
				assert.Equal(t, checkupTypeAuditActionUpdate, audit.entries[0].Action)
				assert.Equal(t, tt.stored, audit.entries[0].OldValue.(map[string]any)["is_non_insurance"])
				assert.Equal(t, got.IsNonInsurance, audit.entries[0].NewValue.(map[string]any)["is_non_insurance"])
			} else {
				assert.Empty(t, audit.entries)
			}
		})
	}
}

func TestCheckupTypeService_Update_RejectsWithoutWrite(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateCheckupTypeInput
		findErr   error
		updateErr error
		auditErr  error
		noActor   bool
		check     func(t *testing.T, err error)
		wantWrite bool
	}{
		{name: "version only is not a field change",
			input: UpdateCheckupTypeInput{Version: int64Ptr(3)},
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsInvalidInput(err)) }},
		{name: "missing client version is rejected before any read/write (no fallback)",
			input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false)},
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsInvalidInput(err)) }},
		{name: "other clinic or missing row is not found",
			input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(3)}, findErr: apperrors.WrapNotFound("checkup_type", "9"),
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsNotFound(err)) }},
		{name: "version conflict is surfaced as conflict",
			input:     UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(1)},
			updateErr: apperrors.WrapConflict("stale"), wantWrite: true,
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsConflict(err)) }},
		{name: "audit failure fails the update (tx rollback)",
			input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(3)}, auditErr: errors.New("audit down"), wantWrite: true,
			check: func(t *testing.T, err error) { assert.Error(t, err) }},
		{name: "missing actor on an insurance change is rejected (fail-closed)",
			input: UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(3)}, noActor: true, wantWrite: true,
			check: func(t *testing.T, err error) { assert.True(t, apperrors.IsInvalidInput(err)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrote := false
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.CheckupType{ID: 9, ClinicID: 1, IsNonInsurance: boolPtr(true), Version: 3}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupTypeInput) (*model.CheckupType, error) {
					wrote = true
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return &model.CheckupType{ID: 9, ClinicID: 1, IsNonInsurance: boolPtr(false), Version: 4}, nil
				},
			}
			svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{err: tt.auditErr})
			input := tt.input
			if !tt.noActor {
				input.ActorID = &testCheckupTypeActorID
			}
			got, err := svc.Update(context.Background(), 1, 9, &input)
			assert.Nil(t, got)
			tt.check(t, err)
			assert.Equal(t, tt.wantWrite, wrote)
		})
	}
}

func TestCheckupTypeService_Update_RequiresClientVersion(t *testing.T) {
	// review 4320626569 の解消条件: 2 人が版 3 を読み、一方が版 4 へ保存した後、
	// 他方の版 3 保存は 409・書込なし。service が読取時点の版で補う fallback は廃止済み。
	name := "改名"
	var updateCalls int
	repo := &mockCheckupTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
			return &model.CheckupType{ID: 9, ClinicID: 1, Name: "健診", IsNonInsurance: boolPtr(true), Version: 4}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error) {
			updateCalls++
			return nil, apperrors.WrapConflict(errMsgCheckupTypeVersionConflict)
		},
	}
	svc := NewCheckupTypeService(repo, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})

	got, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{
		Name: &name, Version: int64Ptr(3), ActorID: &testCheckupTypeActorID,
	})

	assert.Nil(t, got)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "stale client version must be a conflict, not a silent overwrite")
	assert.Equal(t, 1, updateCalls, "the update was attempted exactly once and rolled back")
}

func TestCheckupTypeRequest_InsuranceBinding(t *testing.T) {
	t.Run("update absent means unchanged (nil reaches the service)", func(t *testing.T) {
		var req updateCheckupTypeRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"a","version":2}`), &req))
		assert.Nil(t, req.toServiceInput().IsNonInsurance)
	})
	t.Run("update explicit null is rejected (never treated as unchanged)", func(t *testing.T) {
		var req updateCheckupTypeRequest
		assert.Error(t, json.Unmarshal([]byte(`{"name":"a","is_non_insurance":null,"version":2}`), &req))
	})
	t.Run("create explicit null is rejected (never treated as the default)", func(t *testing.T) {
		var req createCheckupTypeRequest
		assert.Error(t, json.Unmarshal([]byte(`{"name":"a","is_non_insurance":null}`), &req))
	})
	t.Run("update version is required on the wire", func(t *testing.T) {
		// ShouldBindJSON が required,min=1 を強制する。encoding/json は binding を見ないため
		// タグの存在をここで固定し、HTTP の 400 は handler test が実測する。
		f, ok := reflect.TypeOf(updateCheckupTypeRequest{}).FieldByName("Version")
		require.True(t, ok)
		assert.Contains(t, f.Tag.Get("binding"), "required")
		assert.Contains(t, f.Tag.Get("binding"), "min=1")
	})
	t.Run("update false/true and version are carried", func(t *testing.T) {
		var req updateCheckupTypeRequest
		require.NoError(t, json.Unmarshal([]byte(`{"is_non_insurance":false,"version":4}`), &req))
		in := req.toServiceInput()
		require.NotNil(t, in.IsNonInsurance)
		assert.False(t, *in.IsNonInsurance)
		require.NotNil(t, in.Version)
		assert.Equal(t, int64(4), *in.Version)
	})
	t.Run("create absent stays nil so the service applies the new-entry default", func(t *testing.T) {
		var req createCheckupTypeRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"a"}`), &req))
		assert.Nil(t, req.toServiceInput().IsNonInsurance)
	})
	t.Run("create true/false is carried", func(t *testing.T) {
		var req createCheckupTypeRequest
		require.NoError(t, json.Unmarshal([]byte(`{"name":"a","is_non_insurance":false}`), &req))
		in := req.toServiceInput()
		require.NotNil(t, in.IsNonInsurance)
		assert.False(t, *in.IsNonInsurance)
	})
	t.Run("non-boolean payload is rejected", func(t *testing.T) {
		var req updateCheckupTypeRequest
		assert.Error(t, json.Unmarshal([]byte(`{"is_non_insurance":"yes","version":2}`), &req))
	})
}

func TestCheckupTypeResponse_InsuranceSetting(t *testing.T) {
	tests := []struct {
		name  string
		value *bool
		want  string
	}{
		{name: "legacy missing is null, not false", value: nil, want: `null`},
		{name: "explicit false", value: boolPtr(false), want: `false`},
		{name: "explicit true", value: boolPtr(true), want: `true`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(toCheckupTypeResponse(&model.CheckupType{ID: 1, IsNonInsurance: tt.value, Version: 7}))
			require.NoError(t, err)
			var got map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(raw, &got))
			assert.JSONEq(t, tt.want, string(got["is_non_insurance"]))
			assert.JSONEq(t, `7`, string(got["version"]))
		})
	}
}

func assertCheckupTypeInternalError(t *testing.T, err error) {
	t.Helper()
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestCheckupTypeService_InvalidDirectInputsStopBeforeRepository(t *testing.T) {
	svc := NewCheckupTypeService(&mockCheckupTypeRepository{}, checkupTypePassthroughTx{}, &recordingCheckupTypeAudit{})
	got, err := svc.Create(context.Background(), 1, nil)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err))
	for _, version := range []int64{0, -1} {
		got, err = svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{Name: checkupRuntimeString("a"), Version: &version})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
	}
}
func checkupRuntimeString(value string) *string { return &value }

type checkupRuntimeTxMarker struct{}
type checkupRuntimeTx struct{}

func (checkupRuntimeTx) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, checkupRuntimeTxMarker{}, true))
}

type checkupRuntimeAudit struct {
	log func(context.Context, *AuditEntry) error
}

func (a checkupRuntimeAudit) LogEntryTx(ctx context.Context, entry *AuditEntry) error {
	return a.log(ctx, entry)
}

// This verifies context composition, not a database commit/rollback.
func TestCheckupTypeService_CreateAndUpdateShareAuditTransactionContext(t *testing.T) {
	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			checks, writes, audits := 0, 0, 0
			check := func(ctx context.Context) { require.Equal(t, true, ctx.Value(checkupRuntimeTxMarker{})); checks++ }
			repo := &mockCheckupTypeRepository{
				findByIDFn: func(ctx context.Context, _, _ uint64) (*model.CheckupType, error) {
					check(ctx)
					return &model.CheckupType{ID: 9, ClinicID: 1, Version: 7, IsNonInsurance: nil}, nil
				},
				createFn: func(ctx context.Context, item *model.CheckupType) error {
					check(ctx)
					writes++
					item.ID = 9
					item.Version = 1
					return nil
				},
				updateFieldsFn: func(ctx context.Context, _, _ uint64, cmd UpdateCheckupTypeInput) (*model.CheckupType, error) {
					check(ctx)
					writes++
					require.Equal(t, int64(7), *cmd.expectedVersion)
					return &model.CheckupType{ID: 9, ClinicID: 1, Version: 8, IsNonInsurance: cmd.IsNonInsurance}, nil
				},
			}
			audit := checkupRuntimeAudit{log: func(ctx context.Context, entry *AuditEntry) error {
				check(ctx)
				audits++
				require.Equal(t, uint64(1), *entry.ClinicID)
				require.Equal(t, testCheckupTypeActorID, *entry.ActorID)
				if operation == "update" {
					require.Nil(t, entry.OldValue.(map[string]any)["is_non_insurance"])
				}
				return nil
			}}
			svc := NewCheckupTypeService(repo, checkupRuntimeTx{}, audit)
			if operation == "create" {
				got, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "a", ActorID: &testCheckupTypeActorID})
				require.NoError(t, err)
				require.NotNil(t, got.IsNonInsurance)
				require.True(t, *got.IsNonInsurance)
				require.NotNil(t, got.SortOrder)
			} else {
				got, err := svc.Update(context.Background(), 1, 9, &UpdateCheckupTypeInput{IsNonInsurance: boolPtr(false), Version: int64Ptr(7), ActorID: &testCheckupTypeActorID})
				require.NoError(t, err)
				require.False(t, *got.IsNonInsurance)
				require.Equal(t, int64(8), got.Version)
			}
			require.Equal(t, 1, writes)
			require.Equal(t, 1, audits)
			require.GreaterOrEqual(t, checks, 2)
		})
	}
}
