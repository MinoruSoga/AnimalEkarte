package service

import (
	"context"
	"encoding/json"
	"errors"
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
	// existing は置換前に service が取得する「既存（=削除対象）結果」のスナップショット。
	// #211 の削除監査テストでは、これを seed して削除が発生する状況を再現する。
	existing []model.CheckupFieldResult
	// existingErr は FindByCheckupID（置換前スナップショット）の失敗を注入する。
	// スナップショット失敗時に置換を中断する制御（監査なし silent 削除の防止）を検証する。
	existingErr   error
	findByPetIDFn func(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
}

func (m *mockCheckupFieldResultRepository) FindByCheckupID(_ context.Context, _, _ uint64) ([]model.CheckupFieldResult, error) {
	return m.existing, m.existingErr
}

func (m *mockCheckupFieldResultRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	if m.findByPetIDFn != nil {
		return m.findByPetIDFn(ctx, clinicID, petID)
	}
	return nil, nil
}

func (m *mockCheckupFieldResultRepository) ReplaceForCheckup(_ context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error) {
	m.replaceCalled = true
	for i := range results {
		results[i].CheckupID = checkupID
		results[i].ClinicID = clinicID
	}
	m.captured = results
	// 実 repo は既存全削除→挿入。mock は snapshot 件数（m.existing）を実削除数として返す
	// （#211: サービスの監査ゲートはスナップショットでなく実削除数 deletedCount に基づくため）。
	return results, int64(len(m.existing)), nil
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

// newCheckupFieldResultServiceForTest は実 AuditService（NewAuditService）に記録用の
// mockAuditRepository を噛ませて service を構築する。mock の AuditService ではなく実装を使うことで
// validateAuditLog（actor_type / clinic_id 整合）が実際に走り、監査が silent に弾かれる失敗モード
// （前段 security M2 の教訓）を検知できる。auditRepo.lastLogged は validateAuditLog 通過後にのみ set される。
func newCheckupFieldResultServiceForTest(resultRepo *mockCheckupFieldResultRepository) (CheckupFieldResultService, *mockCheckupFieldResultRepository, *mockAuditRepository) {
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
	auditRepo := &mockAuditRepository{}
	// 実 AuditService（*auditService）は AuditTxLogger も実装する。noopTransactor は fn を直接実行する
	// テスト用トランザクション（fn のエラーを伝播）。これにより #211 の fail-closed 制御フローを検証できる。
	svc := NewCheckupFieldResultService(checkupRepo, mrRepo, fieldRepo, resultRepo, NewAuditService(auditRepo).(AuditTxLogger), noopTransactor{})
	return svc, resultRepo, auditRepo
}

// SC3: 別 clinic / 別パッケージの checkup_type_field_id を渡す書き込みは拒否され、永続化されない。
// 同一パッケージのフィールドは従来どおり保存できる（#124 同型ガード・examination ReplaceItems と同型）。
//
// temp-revert RED 実証: service の field 所有権検証ループ（fieldByID メンバシップ確認）を外すと、
// この "rejected_not_persisted" ケースが ReplaceForCheckup を呼び persist してしまい失敗する。
func TestCheckupFieldResultService_ReplaceForCheckup_RejectsCrossPackageField(t *testing.T) {
	t.Run("cross-package field id rejected and NOT persisted", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{}
		svc, repo, _ := newCheckupFieldResultServiceForTest(resultRepo)

		// field 999 は clinic A の歯科パッケージに属さない（別 clinic / 別種別の項目）。
		_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, nil, []UpsertCheckupFieldResultInput{
			{CheckupTypeFieldID: uint64Ptr(999), ValueBool: boolPtr(true)},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "別パッケージのフィールド参照は invalid input で拒否されるべき: %v", err)
		assert.False(t, repo.replaceCalled, "拒否された書き込みは永続化されてはならない")
	})

	t.Run("same-package field id succeeds and is persisted", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{}
		svc, repo, _ := newCheckupFieldResultServiceForTest(resultRepo)

		saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, nil, []UpsertCheckupFieldResultInput{
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
			svc, _, _ := newCheckupFieldResultServiceForTest(resultRepo)
			v := tc.value
			saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, nil, []UpsertCheckupFieldResultInput{
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
	svc, repo, _ := newCheckupFieldResultServiceForTest(resultRepo)

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, nil, []UpsertCheckupFieldResultInput{
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
	svc := NewCheckupFieldResultService(checkupRepo, mrRepo, fieldRepo, resultRepo, NewAuditService(&mockAuditRepository{}).(AuditTxLogger), noopTransactor{})

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, nil, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(101), ValueBool: boolPtr(true)},
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテは conflict で拒否されるべき: %v", err)
	assert.False(t, resultRepo.replaceCalled)
}

// SC1 (#211): 既存結果を持つ checkup の置換は、削除を実 AuditService 経由で監査記録する。
// 誰が（actor_id）・どの clinic（clinic_id）・対象（checkup_id=resource_id）・何を削除したか（old_value）が
// 追跡可能であることを assert する。これは「監査なしの silent な患者結果削除」旧挙動の回帰ガード。
//
// temp-revert RED 実証: ReplaceForCheckup の `if len(existing) > 0 && s.auditSvc != nil { LogEntry(...) }`
// ブロックを削除すると auditRepo.lastLogged が nil のままになり本テストは RED（=旧 silent 削除を再現）。
func TestCheckupFieldResultService_ReplaceForCheckup_AuditsDeletion(t *testing.T) {
	resultRepo := &mockCheckupFieldResultRepository{
		existing: []model.CheckupFieldResult{
			{ID: 50, ClinicID: 1, CheckupID: 7, CheckupTypeFieldID: uint64Ptr(101),
				FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, ValueBool: boolPtr(true)},
		},
	}
	svc, repo, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
	actor := uint64(9)
	num := 2.0

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(102), ValueNumber: &num},
	})
	require.NoError(t, err)
	assert.True(t, repo.replaceCalled)

	require.NotNil(t, auditRepo.lastLogged, "既存結果の削除は監査記録されなければならない（silent 削除禁止）")
	logged := auditRepo.lastLogged
	require.NotNil(t, logged.ClinicID)
	assert.Equal(t, uint64(1), *logged.ClinicID, "監査は対象 clinic を記録する")
	require.NotNil(t, logged.ActorID)
	assert.Equal(t, uint64(9), *logged.ActorID, "監査は実行者（誰が）を記録する")
	assert.Equal(t, model.AuditActorTypeStaff, logged.ActorType)
	assert.Equal(t, model.AuditActionCheckupFieldResultReplace, logged.Action)
	assert.Equal(t, model.AuditResourceCheckupFieldResult, logged.Resource)
	require.NotNil(t, logged.ResourceID)
	assert.Equal(t, uint64(7), *logged.ResourceID, "監査は対象 checkup を記録する")
	assert.NotEmpty(t, logged.OldValue, "監査は削除された結果値（何を）を記録する")
	assert.Contains(t, string(logged.OldValue), "歯石除去必要の有無", "削除された結果値のスナップショットが old_value に含まれる")
}

// SC2 (#211): 空PUT の挙動。
//
//	(a) results 省略（nil）は拒否され、永続化も監査も発生しない（偶発的全消去の防御＝明示フラグ要求）。
//	(b) 明示的な空配列 [] による全削除は PUT セマンティクスとして許容されるが、既存結果があれば監査される。
//	(c) 既存結果が無い空配列は削除が発生しないため監査されない。
//
// 兄弟整合の根拠: 兄弟 examination.ReplaceItems（exam_results）は空入力で既存を silent・無監査削除する
// 同型欠陥を持つ（auditSvc 注入済みだが delete を監査せず防御も無い）。本タスクは checkup でこの
// 監査＋明示フラグ防御パターンを確立し、exam_results は別タスクで適用する（prompt 明記）。
func TestCheckupFieldResultService_ReplaceForCheckup_EmptyPut(t *testing.T) {
	t.Run("nil results (omitted) is rejected and not persisted/audited", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{
			existing: []model.CheckupFieldResult{
				{ID: 50, ClinicID: 1, CheckupID: 7, CheckupTypeFieldID: uint64Ptr(101),
					FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, ValueBool: boolPtr(true)},
			},
		}
		svc, repo, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
		actor := uint64(9)

		_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "省略された results は必須エラーで拒否されるべき: %v", err)
		assert.False(t, repo.replaceCalled, "拒否時は既存結果を削除してはならない")
		assert.Nil(t, auditRepo.lastLogged, "削除が起きないため監査も発生しない")
	})

	t.Run("explicit empty array clears and is audited when results existed", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{
			existing: []model.CheckupFieldResult{
				{ID: 50, ClinicID: 1, CheckupID: 7, CheckupTypeFieldID: uint64Ptr(101),
					FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, ValueBool: boolPtr(true)},
			},
		}
		svc, repo, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
		actor := uint64(9)

		saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, []UpsertCheckupFieldResultInput{})
		require.NoError(t, err, "明示的な空配列による全削除は PUT セマンティクスとして許容される")
		assert.True(t, repo.replaceCalled)
		assert.Len(t, saved, 0)
		require.NotNil(t, auditRepo.lastLogged, "空配列での全削除も監査されなければならない（silent 全削除禁止）")
		assert.Equal(t, model.AuditResourceCheckupFieldResult, auditRepo.lastLogged.Resource)
		// JSON キー順序に依存しないよう unmarshal して型付きで検証する。
		var meta map[string]any
		require.NoError(t, json.Unmarshal(auditRepo.lastLogged.Metadata, &meta))
		assert.Equal(t, float64(0), meta["new_count"], "全削除は new_count=0 として監査される")
		assert.Equal(t, float64(1), meta["deleted_count"], "削除件数が監査される")
	})

	t.Run("empty array with no existing results is not audited (no deletion)", func(t *testing.T) {
		resultRepo := &mockCheckupFieldResultRepository{} // existing = nil
		svc, repo, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
		actor := uint64(9)

		saved, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, []UpsertCheckupFieldResultInput{})
		require.NoError(t, err)
		assert.True(t, repo.replaceCalled)
		assert.Len(t, saved, 0)
		assert.Nil(t, auditRepo.lastLogged, "既存結果が無ければ削除は発生せず監査もされない")
	})
}

// SC1/SC3 補強（security LOW-2 / go M3）: 置換前スナップショット取得（FindByCheckupID）が失敗したら、
// 監査なしの silent 削除を避けるため置換を中断する。スナップショットを確立できないまま削除する経路は
// 旧 silent 削除の再来であり、許してはならない。
func TestCheckupFieldResultService_ReplaceForCheckup_AbortsWhenSnapshotFails(t *testing.T) {
	resultRepo := &mockCheckupFieldResultRepository{
		existingErr: errors.New("db unavailable"),
	}
	svc, repo, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
	actor := uint64(9)

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(101), ValueBool: boolPtr(true)},
	})
	require.Error(t, err, "スナップショット取得失敗時は置換を中断しなければならない")
	assert.False(t, repo.replaceCalled, "スナップショットを確立できないまま削除してはならない（silent 削除防止）")
	assert.Nil(t, auditRepo.lastLogged)
}

// SC3 (#211): 監査書込が失敗したら ReplaceForCheckup はエラーを返す（旧 best-effort → fail-closed への
// 意図的な挙動変更）。tx 内監査により、監査が書けないなら患者検診結果の削除も行わない。
// 実 DB での「削除がロールバックし既存結果が DB に残存する」原子性は repository の tx atomicity テストが
// 正本（checkup_field_result_tx_atomicity_test.go: RollsBackWhenAmbientTxFails）。本 service テストは
// 監査エラーを握り潰さず伝播する fail-closed 制御フローを検証する（noopTransactor が fn のエラーを伝播）。
//
// temp-revert RED 実証: service の監査ブロックを `return apperrors.Wrap(...)` から旧 best-effort の
// slog のみ（エラーを返さない）に戻すと、本テストは NoError となり RED（旧 silent 継続を再現）。
func TestCheckupFieldResultService_ReplaceForCheckup_AuditFailureIsFailClosed(t *testing.T) {
	resultRepo := &mockCheckupFieldResultRepository{
		existing: []model.CheckupFieldResult{
			{ID: 50, ClinicID: 1, CheckupID: 7, CheckupTypeFieldID: uint64Ptr(101),
				FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, ValueBool: boolPtr(true)},
		},
	}
	svc, _, auditRepo := newCheckupFieldResultServiceForTest(resultRepo)
	auditRepo.createFn = func(_ context.Context, _ *model.AuditLog) error {
		return errors.New("audit_logs insert failed")
	}
	actor := uint64(9)

	_, err := svc.ReplaceForCheckup(context.Background(), 1, 5, 7, &actor, []UpsertCheckupFieldResultInput{
		{CheckupTypeFieldID: uint64Ptr(102), ValueNumber: float64PtrLocal(2)},
	})
	require.Error(t, err, "監査書込失敗時は置換をエラーにし、削除をロールバックさせる（fail-closed）")
}

// float64PtrLocal は #211 テスト内で *float64 を生成するヘルパー。
func float64PtrLocal(v float64) *float64 { return &v }
