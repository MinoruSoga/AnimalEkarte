package medicalrecord

// conversation_summary_service_test.go — EMR-485 会話要約 service の単体テスト（DB 不要）。
//
// service の責務だけを検証する:
//   - 入力検証（source 非空白・要約非空・clinic/カルテ特定・期待版の符号）
//   - source の前後空白除去と期待版の透過
//   - repository エラー（conflict / NotFound）の透過
//   - 未保存は unknown（3列 nil）のまま返し、推定値を生成しない
// revision 採番・冪等再送・clinic 隔離は repository 側の実DBテストが担う。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type stubConversationSummaryRepository struct {
	findCalls int
	saveCalls int
	lastCmd   SaveConversationSummaryCommand
	found     *ConversationSummary
	saved     *ConversationSummary
	findErr   error
	saveErr   error
}

func (s *stubConversationSummaryRepository) FindByMedicalRecordID(_ context.Context, _, _ uint64) (*ConversationSummary, error) {
	s.findCalls++
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.found, nil
}

func (s *stubConversationSummaryRepository) Save(_ context.Context, cmd SaveConversationSummaryCommand) (*ConversationSummary, error) {
	s.saveCalls++
	s.lastCmd = cmd
	if s.saveErr != nil {
		return nil, s.saveErr
	}
	return s.saved, nil
}

func conversationSummaryStringPtr(v string) *string { return &v }

func TestConversationSummaryService_SaveRejectsBlankSource(t *testing.T) {
	repo := &stubConversationSummaryRepository{}
	service := NewConversationSummaryService(repo)

	for _, source := range []string{"", "   ", "\t\n"} {
		saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
			ClinicID:        1,
			MedicalRecordID: 2,
			Summary:         "要約本文",
			Source:          source,
		})
		assert.Nil(t, saved)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "source 空白は invalid input: %v", err)
	}
	assert.Equal(t, 0, repo.saveCalls, "検証で落ちた保存は repository を呼ばない")
}

func TestConversationSummaryService_SaveRejectsBlankSummary(t *testing.T) {
	repo := &stubConversationSummaryRepository{}
	service := NewConversationSummaryService(repo)

	saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
		ClinicID:        1,
		MedicalRecordID: 2,
		Summary:         "  \n ",
		Source:          "lstep",
	})
	assert.Nil(t, saved)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "空の要約は invalid input: %v", err)
	assert.Equal(t, 0, repo.saveCalls)
}

func TestConversationSummaryService_SaveRejectsUnknownScope(t *testing.T) {
	repo := &stubConversationSummaryRepository{}
	service := NewConversationSummaryService(repo)

	cases := []struct {
		name            string
		clinicID        uint64
		medicalRecordID uint64
	}{
		{name: "clinic 未指定", clinicID: 0, medicalRecordID: 2},
		{name: "カルテ未指定", clinicID: 1, medicalRecordID: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
				ClinicID:        tc.clinicID,
				MedicalRecordID: tc.medicalRecordID,
				Summary:         "要約本文",
				Source:          "lstep",
			})
			assert.Nil(t, saved)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "scope 未特定は invalid input: %v", err)
		})
	}
	assert.Equal(t, 0, repo.saveCalls)
}

func TestConversationSummaryService_SaveRejectsNegativeExpectedRevision(t *testing.T) {
	repo := &stubConversationSummaryRepository{}
	service := NewConversationSummaryService(repo)

	saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
		ClinicID:         1,
		MedicalRecordID:  2,
		Summary:          "要約本文",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(-1),
	})
	assert.Nil(t, saved)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "負の期待版は invalid input: %v", err)
	assert.Equal(t, 0, repo.saveCalls)
}

func TestConversationSummaryService_SaveTrimsSourceAndForwardsExpectedRevision(t *testing.T) {
	repo := &stubConversationSummaryRepository{
		saved: &ConversationSummary{
			MedicalRecordID: 2,
			Summary:         conversationSummaryStringPtr("要約本文"),
			Revision:        conversationSummaryInt64Ptr(3),
			Source:          conversationSummaryStringPtr("lstep"),
		},
	}
	service := NewConversationSummaryService(repo)

	saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
		ClinicID:         1,
		MedicalRecordID:  2,
		Summary:          "要約本文",
		Source:           "  lstep  ",
		ExpectedRevision: conversationSummaryInt64Ptr(2),
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	require.NotNil(t, saved.Revision)
	assert.Equal(t, int64(3), *saved.Revision)

	require.Equal(t, 1, repo.saveCalls)
	assert.Equal(t, "lstep", repo.lastCmd.Source, "source は前後空白を除去して保存する")
	assert.Equal(t, "要約本文", repo.lastCmd.Summary, "要約本文は改変しない")
	require.NotNil(t, repo.lastCmd.ExpectedRevision)
	assert.Equal(t, int64(2), *repo.lastCmd.ExpectedRevision)
	assert.Equal(t, uint64(1), repo.lastCmd.ClinicID)
	assert.Equal(t, uint64(2), repo.lastCmd.MedicalRecordID)
}

func TestConversationSummaryService_SavePropagatesConflict(t *testing.T) {
	conflict := apperrors.WrapConflict(conversationSummaryConflictMessage)
	repo := &stubConversationSummaryRepository{saveErr: conflict}
	service := NewConversationSummaryService(repo)

	saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
		ClinicID:         1,
		MedicalRecordID:  2,
		Summary:          "要約本文",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(5),
	})
	assert.Nil(t, saved)
	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "repository の conflict をそのまま返す: %v", err)
}

func TestConversationSummaryService_SaveRejectsMissingRepository(t *testing.T) {
	service := NewConversationSummaryService(nil)

	saved, err := service.Save(context.Background(), SaveConversationSummaryInput{
		ClinicID:        1,
		MedicalRecordID: 2,
		Summary:         "要約本文",
		Source:          "lstep",
	})
	assert.Nil(t, saved)
	require.Error(t, err)
	assert.False(t, apperrors.IsInvalidInput(err), "保存先未設定は入力エラーではない")
	assert.False(t, apperrors.IsNotFound(err))
}

func TestConversationSummaryService_GetReturnsUnsavedAsUnknown(t *testing.T) {
	repo := &stubConversationSummaryRepository{
		found: &ConversationSummary{MedicalRecordID: 42},
	}
	service := NewConversationSummaryService(repo)

	got, err := service.Get(context.Background(), 1, 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(42), got.MedicalRecordID)
	assert.Nil(t, got.Summary, "未保存の要約本文は nil のまま返す")
	assert.Nil(t, got.Revision, "未保存の revision は nil のまま返す（0 を生成しない）")
	assert.Nil(t, got.Source, "未保存の source は nil のまま返す")
	assert.Equal(t, 1, repo.findCalls)
}

func TestConversationSummaryService_GetPropagatesNotFound(t *testing.T) {
	repo := &stubConversationSummaryRepository{findErr: apperrors.WrapNotFound("conversation_summary", "42")}
	service := NewConversationSummaryService(repo)

	got, err := service.Get(context.Background(), 1, 42)
	assert.Nil(t, got)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "repository の NotFound をそのまま返す: %v", err)
}

func TestConversationSummaryService_GetRejectsUnknownScope(t *testing.T) {
	repo := &stubConversationSummaryRepository{}
	service := NewConversationSummaryService(repo)

	got, err := service.Get(context.Background(), 0, 42)
	assert.Nil(t, got)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Equal(t, 0, repo.findCalls)
}
