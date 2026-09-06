package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func uniqueTreatmentNameErr(resource, constraint string) error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: constraint,
	}, resource, "")
}

func TestTreatmentItemService_Create_NameConflictMapsDomainCode(t *testing.T) {
	t.Run("consultation", func(t *testing.T) {
		svc := NewConsultationService(&mockConsultationRepository{
			createFn: func(_ context.Context, _ *model.Consultation) error {
				return uniqueTreatmentNameErr("consultation", apperrors.ConstraintConsultationName)
			},
		})
		got, err := svc.Create(context.Background(), 1, &CreateConsultationInput{Name: "V04診察"})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeConsultationNameConflict))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "V04診察", appErr.Params["name"])
		assert.NotContains(t, appErr.Message, "consultation '' already exists")
	})

	t.Run("exam_type", func(t *testing.T) {
		svc := NewExamTypeService(&mockExamTypeRepository{
			createFn: func(_ context.Context, _ *model.ExaminationType) error {
				return uniqueTreatmentNameErr("examination_type", apperrors.ConstraintExamTypeName)
			},
		}, passthroughExamTypeTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateExamTypeInput{Name: "V04検査"})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeExamTypeNameConflict))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "V04検査", appErr.Params["name"])
	})

	t.Run("procedure", func(t *testing.T) {
		svc := NewProcedureService(&mockProcedureRepository{
			createFn: func(_ context.Context, _ *model.Procedure) error {
				return uniqueTreatmentNameErr("procedure", apperrors.ConstraintProcedureName)
			},
		}, &mockTransactor{})
		got, err := svc.Create(context.Background(), 1, &CreateProcedureInput{Name: "V04処置"})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeProcedureNameConflict))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "V04処置", appErr.Params["name"])
	})

	t.Run("vaccine", func(t *testing.T) {
		svc := NewVaccineService(&mockVaccineRepository{
			createFn: func(_ context.Context, _ *model.Vaccine) error {
				return uniqueTreatmentNameErr("vaccine", apperrors.ConstraintVaccineName)
			},
		})
		got, err := svc.Create(context.Background(), 1, &CreateVaccineInput{Name: "V04予防接種"})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeVaccineNameConflict))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "V04予防接種", appErr.Params["name"])
	})

	t.Run("checkup_type", func(t *testing.T) {
		svc := NewCheckupTypeService(&mockCheckupTypeRepository{
			createFn: func(_ context.Context, _ *model.CheckupType) error {
				return uniqueTreatmentNameErr("checkup_type", apperrors.ConstraintCheckupTypeName)
			},
		})
		got, err := svc.Create(context.Background(), 1, &CreateCheckupTypeInput{Name: "V04定期健診"})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeCheckupTypeNameConflict))
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "V04定期健診", appErr.Params["name"])
	})
}

func TestConsultationService_Update_NameConflictMapsDomainCode(t *testing.T) {
	name := "V04診察"
	svc := NewConsultationService(&mockConsultationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
			return &model.Consultation{ID: id, ClinicID: 1, Name: "旧"}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateConsultationInput) (*model.Consultation, error) {
			return nil, uniqueTreatmentNameErr("consultation", apperrors.ConstraintConsultationName)
		},
	})

	got, err := svc.Update(context.Background(), 1, 1, &UpdateConsultationInput{Name: &name})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeConsultationNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "V04診察", appErr.Params["name"])
}

func TestConsultationService_Create_GenericAlreadyExistsNotElevated(t *testing.T) {
	svc := NewConsultationService(&mockConsultationRepository{
		createFn: func(_ context.Context, _ *model.Consultation) error {
			return apperrors.WrapAlreadyExists("consultation", "")
		},
	})
	_, err := svc.Create(context.Background(), 1, &CreateConsultationInput{Name: "X"})
	require.Error(t, err)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodeConsultationNameConflict))
	assert.True(t, apperrors.IsAlreadyExists(err))
}
