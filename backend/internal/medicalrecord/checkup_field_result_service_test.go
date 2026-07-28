package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockCheckupResultCheckupRepository struct {
	CheckupRepository
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Checkup, error)
}

func (m *mockCheckupResultCheckupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Checkup{ID: id, ClinicID: clinicID, MedicalRecordID: 10, CheckupTypeID: 100}, nil
}

type mockCheckupResultMedicalRecordRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

func (m *mockCheckupResultMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
}

// LockByIDForUpdate は X-11 finalize-lock テスト用に FindByID と同じ挙動へ委譲する。
func (m *mockCheckupResultMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}

type mockCheckupTypeFieldRepository struct {
	CheckupTypeFieldRepository
	findByCheckupTypeIDFn func(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
}

func (m *mockCheckupTypeFieldRepository) FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	if m.findByCheckupTypeIDFn != nil {
		return m.findByCheckupTypeIDFn(ctx, clinicID, checkupTypeID)
	}
	return []model.CheckupTypeField{
		{ID: 1, Name: "NumberField", FieldType: model.CheckupFieldTypeNumber, SortOrder: 1},
		{ID: 2, Name: "BoolField", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 2},
		{ID: 3, Name: "MultiField", FieldType: model.CheckupFieldTypeMultiSelect, Options: []byte("[{\"value\":\"A\"},{\"value\":\"B\"}]"), SortOrder: 3},
		{ID: 4, Name: "SingleField", FieldType: model.CheckupFieldTypeSingleSelect, Options: []byte("[{\"value\":\"X\"},{\"value\":\"Y\"}]"), SortOrder: 4},
		{ID: 5, Name: "TextField", FieldType: model.CheckupFieldTypeText, SortOrder: 5},
	}, nil
}

type mockCheckupFieldResultRepository struct {
	CheckupFieldResultRepository
	findByCheckupIDFn   func(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error)
	findByPetIDFn       func(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	replaceForCheckupFn func(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error)
}

func (m *mockCheckupFieldResultRepository) FindByCheckupID(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	if m.findByCheckupIDFn != nil {
		return m.findByCheckupIDFn(ctx, clinicID, checkupID)
	}
	return []model.CheckupFieldResult{}, nil
}

func (m *mockCheckupFieldResultRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	if m.findByPetIDFn != nil {
		return m.findByPetIDFn(ctx, clinicID, petID)
	}
	return []model.CheckupFieldResult{}, nil
}

func (m *mockCheckupFieldResultRepository) ReplaceForCheckup(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error) {
	if m.replaceForCheckupFn != nil {
		return m.replaceForCheckupFn(ctx, clinicID, checkupID, results)
	}
	return results, int64(len(results)), nil
}

type mockAuditTxLogger struct {
	logEntryTxFn func(ctx context.Context, input *AuditEntry) error
}

func (m *mockAuditTxLogger) LogEntryTx(ctx context.Context, input *AuditEntry) error {
	if m.logEntryTxFn != nil {
		return m.logEntryTxFn(ctx, input)
	}
	return nil
}

type mockCheckupTransactor struct {
	Transactor
}

func (m *mockCheckupTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// TestComputeCheckupNumberStatus は number 型の status/is_abnormal 導出の全分岐を検証する。
func TestComputeCheckupNumberStatus(t *testing.T) {
	tests := []struct {
		name           string
		value          *float64
		refMin         *float64
		refMax         *float64
		wantStatus     model.ExaminationResultStatus
		wantIsAbnormal bool
	}{
		{
			name:           "nil value returns normal",
			value:          nil,
			refMin:         fptr(1),
			refMax:         fptr(10),
			wantStatus:     model.ExaminationResultStatusNormal,
			wantIsAbnormal: false,
		},
		{
			name:           "value below refMin returns low",
			value:          fptr(0.5),
			refMin:         fptr(1),
			refMax:         fptr(10),
			wantStatus:     model.ExaminationResultStatusLow,
			wantIsAbnormal: true,
		},
		{
			name:           "value above refMax returns high",
			value:          fptr(11),
			refMin:         fptr(1),
			refMax:         fptr(10),
			wantStatus:     model.ExaminationResultStatusHigh,
			wantIsAbnormal: true,
		},
		{
			name:           "value within range returns normal",
			value:          fptr(5),
			refMin:         fptr(1),
			refMax:         fptr(10),
			wantStatus:     model.ExaminationResultStatusNormal,
			wantIsAbnormal: false,
		},
		{
			name:           "nil refMin skips low check",
			value:          fptr(-100),
			refMin:         nil,
			refMax:         fptr(10),
			wantStatus:     model.ExaminationResultStatusNormal,
			wantIsAbnormal: false,
		},
		{
			name:           "nil refMax skips high check",
			value:          fptr(1000),
			refMin:         fptr(1),
			refMax:         nil,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantIsAbnormal: false,
		},
		{
			name:           "both ref bounds nil returns normal",
			value:          fptr(42),
			refMin:         nil,
			refMax:         nil,
			wantStatus:     model.ExaminationResultStatusNormal,
			wantIsAbnormal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, isAbnormal := computeCheckupNumberStatus(tt.value, tt.refMin, tt.refMax)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantIsAbnormal, isAbnormal)
		})
	}
}

func TestCheckupFieldResultService_ListFields(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(nil, nil, repo, nil, nil, nil)
		res, err := svc.ListFields(ctx, 1, 100)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockCheckupTypeFieldRepository{
			findByCheckupTypeIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCheckupFieldResultService(nil, nil, repo, nil, nil, nil)
		_, err := svc.ListFields(ctx, 1, 100)
		assert.Error(t, err)
	})
}

func TestCheckupFieldResultService_ListByCheckup(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		resultRepo := &mockCheckupFieldResultRepository{
			findByCheckupIDFn: func(_ context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error) {
				return []model.CheckupFieldResult{{ID: 1, ClinicID: clinicID, CheckupID: checkupID}}, nil
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, resultRepo, nil, nil)
		res, err := svc.ListByCheckup(ctx, 1, 10, 50)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("checkup verify error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
				return nil, apperrors.WrapNotFound("checkup", "")
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, nil, nil, nil)
		_, err := svc.ListByCheckup(ctx, 1, 10, 50)
		assert.Error(t, err)
	})

	t.Run("results error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		resultRepo := &mockCheckupFieldResultRepository{
			findByCheckupIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, resultRepo, nil, nil)
		_, err := svc.ListByCheckup(ctx, 1, 10, 50)
		assert.Error(t, err)
	})

	// verifyCheckup: 非 NotFound の repo エラー（DB 障害等）は NotFound と異なりログ出力を伴う分岐。
	t.Run("checkup verify non-notfound repository error is logged and wrapped", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
				return nil, errors.New("db connection error")
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, nil, nil, nil)
		_, err := svc.ListByCheckup(ctx, 1, 10, 50)
		assert.Error(t, err)
		assert.False(t, apperrors.IsNotFound(err))
	})

	// verifyCheckup: checkup が別の medical_record に属する場合は NotFound（クロスカルテ紐付け拒否）。
	t.Run("checkup belongs to a different medical record returns not found", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Checkup, error) {
				return &model.Checkup{ID: id, ClinicID: clinicID, MedicalRecordID: 999, CheckupTypeID: 100}, nil
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, nil, nil, nil)
		_, err := svc.ListByCheckup(ctx, 1, 10, 50)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestCheckupFieldResultService_ListByPet(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{}
		svc := NewCheckupFieldResultService(nil, nil, nil, resultRepo, nil, nil)
		res, err := svc.ListByPet(ctx, 1, 20)
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("error", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{
			findByPetIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCheckupFieldResultService(nil, nil, nil, resultRepo, nil, nil)
		_, err := svc.ListByPet(ctx, 1, 20)
		assert.Error(t, err)
	})
}

func TestCheckupFieldResultService_ReplaceForCheckup(t *testing.T) {
	ctx := context.Background()

	numVal := 15.5
	refMin := 10.0
	refMax := 20.0
	boolVal := true

	t.Run("success replace with multiple types", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{
			findByCheckupTypeIDFn: func(_ context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
				return []model.CheckupTypeField{
					{ID: 1, Name: "Num", FieldType: model.CheckupFieldTypeNumber, MinValue: &refMin, MaxValue: &refMax},
					{ID: 2, Name: "Bool", FieldType: model.CheckupFieldTypeBoolean},
					{ID: 3, Name: "Multi", FieldType: model.CheckupFieldTypeMultiSelect, Options: []byte("[{\"value\":\"A\"}]")},
					{ID: 4, Name: "Single", FieldType: model.CheckupFieldTypeSingleSelect, Options: []byte("[{\"value\":\"X\"}]")},
					{ID: 5, Name: "Text", FieldType: model.CheckupFieldTypeText},
				}, nil
			},
		}
		resultRepo := &mockCheckupFieldResultRepository{
			findByCheckupIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
				return []model.CheckupFieldResult{{ID: 99, FieldName: "Old", FieldType: model.CheckupFieldTypeText}}, nil
			},
		}
		audit := &mockAuditTxLogger{}
		tx := &mockCheckupTransactor{}

		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, audit, tx)

		id1, id2, id3, id4, id5 := uint64(1), uint64(2), uint64(3), uint64(4), uint64(5)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id1, ValueNumber: &numVal},
			{CheckupTypeFieldID: &id2, ValueBool: &boolVal},
			{CheckupTypeFieldID: &id3, ValueList: []string{"A"}},
			{CheckupTypeFieldID: &id4, ValueText: "X"},
			{CheckupTypeFieldID: &id5, ValueText: "hello"},
		}

		res, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.NoError(t, err)
		assert.Len(t, res, 5)
	})

	t.Run("inputs is nil error", func(t *testing.T) {
		svc := NewCheckupFieldResultService(nil, nil, nil, nil, nil, nil)
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, nil)
		assert.Error(t, err)
	})

	t.Run("checkup verify error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
				return nil, apperrors.WrapNotFound("checkup", "")
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, nil, nil, nil, nil, nil)
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, []UpsertCheckupFieldResultInput{})
		assert.Error(t, err)
	})

	t.Run("parent medical record find error", func(t *testing.T) {
		// X-11: parent status チェックは s.transactor.WithTx 内の先頭（field validation の後）に
		// 移動したため、fieldRepo と transactor も到達可能な状態で構成する必要がある。
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, errors.New("db error")
			},
		}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, &mockCheckupTransactor{})
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, []UpsertCheckupFieldResultInput{})
		assert.Error(t, err)
	})

	t.Run("parent medical record finalized", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusFinalized}, nil
			},
		}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, &mockCheckupTransactor{})
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, []UpsertCheckupFieldResultInput{})
		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
	})

	t.Run("fieldRepo error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{
			findByCheckupTypeIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, []UpsertCheckupFieldResultInput{})
		assert.Error(t, err)
	})

	t.Run("validation input checkup_type_field_id is nil", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: nil},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("validation input field not in type", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		id := uint64(999)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("validation input single select value not allowed", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		id := uint64(4)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueText: "Z"},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("validation input single select allow empty", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		resultRepo := &mockCheckupFieldResultRepository{}
		tx := &mockCheckupTransactor{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, &mockAuditTxLogger{}, tx)
		id := uint64(4)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueText: ""},
		}
		res, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("validation input options unmarshal error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{
			findByCheckupTypeIDFn: func(_ context.Context, _ uint64, _ uint64) ([]model.CheckupTypeField, error) {
				return []model.CheckupTypeField{
					{ID: 4, Name: "Single", FieldType: model.CheckupFieldTypeSingleSelect, Options: []byte("bad-json")},
				}, nil
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		id := uint64(4)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueText: "X"},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	// parseCheckupOptionValues: options 未定義（len==0）フィールドに非空値が送信された場合、
	// 許可集合が空のため必ず「選択肢に存在しません」エラーになる分岐を検証する。
	t.Run("validation input single select with no options defined rejects any value", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{
			findByCheckupTypeIDFn: func(_ context.Context, _ uint64, _ uint64) ([]model.CheckupTypeField, error) {
				return []model.CheckupTypeField{
					{ID: 4, Name: "Single", FieldType: model.CheckupFieldTypeSingleSelect, Options: nil},
				}, nil
			},
		}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		id := uint64(4)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueText: "anything"},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("validation input multi select value not allowed", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, nil, nil, nil)
		id := uint64(3)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueList: []string{"C"}},
		}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("validation input multi select allow empty", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		resultRepo := &mockCheckupFieldResultRepository{}
		tx := &mockCheckupTransactor{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, &mockAuditTxLogger{}, tx)
		id := uint64(3)
		inputs := []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: &id, ValueList: []string{}},
		}
		res, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("transactor snapshot load error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		resultRepo := &mockCheckupFieldResultRepository{
			findByCheckupIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
				return nil, errors.New("db error")
			},
		}
		tx := &mockCheckupTransactor{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, &mockAuditTxLogger{}, tx)
		id := uint64(1)
		inputs := []UpsertCheckupFieldResultInput{{CheckupTypeFieldID: &id}}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("transactor replace save error", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		resultRepo := &mockCheckupFieldResultRepository{
			replaceForCheckupFn: func(_ context.Context, _, _ uint64, _ []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error) {
				return nil, 0, errors.New("save error")
			},
		}
		tx := &mockCheckupTransactor{}
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, &mockAuditTxLogger{}, tx)
		id := uint64(1)
		inputs := []UpsertCheckupFieldResultInput{{CheckupTypeFieldID: &id}}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, nil, inputs)
		assert.Error(t, err)
	})

	t.Run("transactor audit error with actorID", func(t *testing.T) {
		checkupRepo := &mockCheckupResultCheckupRepository{}
		medRepo := &mockCheckupResultMedicalRecordRepository{}
		fieldRepo := &mockCheckupTypeFieldRepository{}
		resultRepo := &mockCheckupFieldResultRepository{
			replaceForCheckupFn: func(_ context.Context, _, _ uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error) {
				return results, 1, nil
			},
		}
		audit := &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
				return errors.New("audit error")
			},
		}
		tx := &mockCheckupTransactor{}
		actor := uint64(100)
		svc := NewCheckupFieldResultService(checkupRepo, medRepo, fieldRepo, resultRepo, audit, tx)
		id := uint64(1)
		inputs := []UpsertCheckupFieldResultInput{{CheckupTypeFieldID: &id}}
		_, err := svc.ReplaceForCheckup(ctx, 1, 10, 50, &actor, inputs)
		assert.Error(t, err)
	})
}
