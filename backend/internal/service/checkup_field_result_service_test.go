package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- #211 健診結果値サービスの mock ----

type mockCheckupTypeFieldRepository struct {
	findByCheckupTypeIDFn func(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
}

func (m *mockCheckupTypeFieldRepository) FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	return m.findByCheckupTypeIDFn(ctx, clinicID, checkupTypeID)
}

type mockCheckupFieldResultRepository struct {
	replaceCalled bool
	captured      []model.CheckupFieldResult
	findByPetIDFn func(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
}

func (m *mockCheckupFieldResultRepository) FindByCheckupID(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
	return m.captured, nil
}

func (m *mockCheckupFieldResultRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	if m.findByPetIDFn != nil {
		return m.findByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

func (m *mockCheckupFieldResultRepository) ReplaceForCheckup(_ context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, error) {
	m.replaceCalled = true
	for i := range results {
		results[i].CheckupID = checkupID
		results[i].ClinicID = clinicID
	}
	m.captured = results
	return results, nil
}

// 健診パッケージ（clinic A の checkup_type=10）が持つフィールド定義。
// 別 clinic / 別パッケージのフィールドはこの集合に現れない（FindByCheckupTypeID が clinic スコープのため）。
func clinicAFields() []model.CheckupTypeField {
	maxScore := 4.0
	minScore := 0.0
	return []model.CheckupTypeField{
		{ID: 101, ClinicID: 1, CheckupTypeID: 10, Name: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1},
		{ID: 102, ClinicID: 1, CheckupTypeID: 10, Name: "歯石付着度スコア", FieldType: model.CheckupFieldTypeNumber, MinValue: &minScore, MaxValue: &maxScore, SortOrder: 2},
		{ID: 103, ClinicID: 1, CheckupTypeID: 10, Name: "歯科ケアアドバイス", FieldType: model.CheckupFieldTypeMultiSelect,
			Options: datatypes.JSON([]byte(`[{"value":"daily_brushing","label":"毎日の歯磨き"},{"value":"scaling","label":"定期スケーリング"}]`)), SortOrder: 3},
	}
}

func newCheckupFieldResultServiceForTest(resultRepo *mockCheckupFieldResultRepository) (CheckupFieldResultService, *mockCheckupFieldResultRepository) {
	checkupRepo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
			// clinic A・カルテ 5・歯科パッケージ(10) の健診記録。
			return &model.Checkup{ID: checkupID, ClinicID: 1, MedicalRecordID: 5, CheckupTypeID: 10}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 5, ClinicID: 1, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	fieldRepo := &mockCheckupTypeFieldRepository{
		findByCheckupTypeIDFn: func(_ context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
			// clinic スコープを模倣: clinic A の歯科パッケージのみフィールドを返す。
			if clinicID == 1 && checkupTypeID == 10 {
				return clinicAFields(), nil
			}
			return []model.CheckupTypeField{}, nil
		},
	}
	svc := NewCheckupFieldResultService(checkupRepo, mrRepo, fieldRepo, resultRepo)
	return svc, resultRepo
}

// SC3: 別 clinic / 別パッケージの checkup_type_field_id を渡す書き込みは拒否され、永続化されない。
// 同一パッケージのフィールドは従来どおり保存できる（#124 同型ガード・examination ReplaceItems と同型）。
//
// temp-revert RED 実証: service の field 所有権検証ループ（fieldByID メンバシップ確認）を外すと、
// この "rejected_not_persisted" ケースが ReplaceForCheckup を呼び persist してしまい失敗する。
func TestCheckupFieldResultService_ReplaceForCheckup_RejectsCrossPackageField(t *testing.T) {
	t.Run("cross-package field id rejected and NOT persisted", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{}
		svc, repo := newCheckupFieldResultServiceForTest(resultRepo)

		// field 999 は clinic A の歯科パッケージに属さない（別 clinic / 別種別の項目）。
		_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: uint64Ptr(999), ValueBool: boolPtr(true)},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "別パッケージのフィールド参照は invalid input で拒否されるべき: %v", err)
		assert.False(t, repo.replaceCalled, "拒否された書き込みは永続化されてはならない")
	})

	t.Run("same-package field id succeeds and is persisted", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{}
		svc, repo := newCheckupFieldResultServiceForTest(resultRepo)

		saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: uint64Ptr(101), ValueBool: boolPtr(true)},
		})
		require.NoError(t, err)
		assert.True(t, repo.replaceCalled, "正当な書き込みは永続化されるべき")
		require.Len(t, saved, 1)
		assert.Equal(t, "歯石除去必要の有無", saved[0].FieldName, "フィールド定義からスナップショットされるべき")
		assert.Equal(t, model.CheckupFieldTypeBoolean, saved[0].FieldType)
	})
}

// SC2: number 型は min/max から status / is_abnormal をサーバ側で導出する。
func TestCheckupFieldResultService_ReplaceForCheckup_ComputesNumberStatus(t *testing.T) {
	cases := []struct {
		name       string
		value      float64
		wantStatus model.ExaminationResultStatus
		wantAbnorm bool
	}{
		{"in range", 2, model.ExaminationResultStatusNormal, false},
		{"above max", 5, model.ExaminationResultStatusHigh, true},
		{"below min", -1, model.ExaminationResultStatusLow, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resultRepo := &mockCheckupFieldResultRepository{}
			svc, _ := newCheckupFieldResultServiceForTest(resultRepo)
			v := tc.value
			saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, []UpsertCheckupFieldResultInput{
				{CheckupTypeFieldID: uint64Ptr(102), ValueNumber: &v},
			})
			require.NoError(t, err)
			require.Len(t, saved, 1)
			assert.Equal(t, tc.wantStatus, saved[0].Status)
			assert.Equal(t, tc.wantAbnorm, saved[0].IsAbnormal)
		})
	}
}

// multi_select は options に存在しない値を request 境界で拒否する。
func TestCheckupFieldResultService_ReplaceForCheckup_RejectsUnknownOption(t *testing.T) {
	resultRepo := &mockCheckupFieldResultRepository{}
	svc, repo := newCheckupFieldResultServiceForTest(resultRepo)

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(103), ValueList: []string{"daily_brushing", "unknown_value"}},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "選択肢に無い値は拒否されるべき: %v", err)
	assert.False(t, repo.replaceCalled)
}

// 確定済みカルテへの結果書き込みは拒否される（checkup Create/Update と対称）。
func TestCheckupFieldResultService_ReplaceForCheckup_RejectsFinalizedRecord(t *testing.T) {
	resultRepo := &mockCheckupFieldResultRepository{}
	checkupRepo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: checkupID, ClinicID: 1, MedicalRecordID: 5, CheckupTypeID: 10}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 5, ClinicID: 1, Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	fieldRepo := &mockCheckupTypeFieldRepository{
		findByCheckupTypeIDFn: func(_ context.Context, _, _ uint64) ([]model.CheckupTypeField, error) {
			return clinicAFields(), nil
		},
	}
	svc := NewCheckupFieldResultService(checkupRepo, mrRepo, fieldRepo, resultRepo)

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(101), ValueBool: boolPtr(true)},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテは conflict で拒否されるべき: %v", err)
	assert.False(t, resultRepo.replaceCalled)
}
