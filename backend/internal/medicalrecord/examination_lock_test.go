package medicalrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestExaminationResultsLocked(t *testing.T) {
	t.Parallel()

	rev := uint64(1)
	cases := []struct {
		name string
		exam *model.Examination
		want bool
	}{
		{name: "nil", exam: nil, want: false},
		{name: "pending", exam: &model.Examination{Status: model.ExaminationStatusPending}, want: false},
		{name: "result_entered", exam: &model.Examination{Status: model.ExaminationStatusResultEntered}, want: false},
		{name: "completed first-pass seal", exam: &model.Examination{Status: model.ExaminationStatusCompleted}, want: true},
		{
			name: "completed post-unconfirm working copy",
			exam: &model.Examination{Status: model.ExaminationStatusCompleted, CurrentRevisionVersion: &rev},
			want: false,
		},
		{name: "confirmed", exam: &model.Examination{Status: model.ExaminationStatusConfirmed}, want: true},
		{
			name: "confirmed with revision",
			exam: &model.Examination{Status: model.ExaminationStatusConfirmed, CurrentRevisionVersion: &rev},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, examinationResultsLocked(tc.exam))
			assert.Equal(t, tc.exam != nil && tc.exam.Status == model.ExaminationStatusConfirmed, examinationFullyLocked(tc.exam))
		})
	}
}

func TestExaminationLockErrorMessages(t *testing.T) {
	t.Parallel()

	completed := &model.Examination{Status: model.ExaminationStatusCompleted}
	confirmed := &model.Examination{Status: model.ExaminationStatusConfirmed}

	require.True(t, apperrors.IsConflict(errExaminationResultsLocked(completed)))
	assert.Contains(t, errExaminationResultsLocked(completed).Error(), "完了済み")
	assert.Contains(t, errExaminationResultsLocked(confirmed).Error(), "確定済み")
	assert.Contains(t, errExaminationDeleteLocked(completed).Error(), "完了済み")
	assert.Contains(t, errExaminationDeleteLocked(confirmed).Error(), "確定済み")
}
