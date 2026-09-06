package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BUG-011: medicine master duplicate-name must map like other clinical-item masters
// (structured 409 code + params.name), not bare `medicine '' already exists`.

func TestMedicineService_Create_NameConflictMapsDomainCode(t *testing.T) {
	repo := &mockMedicineRepository{
		createFn: func(_ context.Context, _ *model.Medicine) error {
			return uniqueTreatmentNameErr("medicine", apperrors.ConstraintMedicineName)
		},
	}
	svc := newTestMedicineService(repo)

	got, err := svc.Create(context.Background(), 1, &CreateMedicineInput{Name: "V04薬剤"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeMedicineNameConflict))
	assert.True(t, apperrors.IsConflict(err))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "V04薬剤", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "medicine '' already exists")
	assert.Equal(t, "resource already exists", appErr.Message)
}

func TestMedicineService_Update_NameConflictMapsDomainCode(t *testing.T) {
	name := "V04薬剤"
	repo := &mockMedicineRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
			return &model.Medicine{ID: id, ClinicID: 1, Name: "旧薬剤"}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
			return nil, uniqueTreatmentNameErr("medicine", apperrors.ConstraintMedicineName)
		},
	}
	svc := newTestMedicineService(repo)

	got, err := svc.Update(context.Background(), 1, 1, &UpdateMedicineInput{Name: &name})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeMedicineNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "V04薬剤", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "medicine '' already exists")
}

func TestMedicineService_Create_GenericAlreadyExistsNotElevated(t *testing.T) {
	repo := &mockMedicineRepository{
		createFn: func(_ context.Context, _ *model.Medicine) error {
			return apperrors.WrapAlreadyExists("medicine", "")
		},
	}
	svc := newTestMedicineService(repo)
	_, err := svc.Create(context.Background(), 1, &CreateMedicineInput{Name: "X"})
	require.Error(t, err)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodeMedicineNameConflict))
	assert.True(t, apperrors.IsAlreadyExists(err))
}
