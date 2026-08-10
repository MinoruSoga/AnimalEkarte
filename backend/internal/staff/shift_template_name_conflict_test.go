package staff

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

func uniqueShiftTemplateNameErr() error {
	return apperrors.FromGORM(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: apperrors.ConstraintShiftTemplateName,
	}, "shift_template", "")
}

func TestShiftTemplateService_Create_NameConflictMapsDomainCode(t *testing.T) {
	repo := &mockShiftTemplateRepository{
		createFn: func(_ context.Context, _ *model.ShiftTemplate) error {
			return uniqueShiftTemplateNameErr()
		},
	}
	svc := NewShiftTemplateService(repo)

	tpl, err := svc.Create(context.Background(), 10, &CreateShiftTemplateInput{
		Name:      "早番",
		ShiftType: string(model.ShiftTypeMorning),
		StartTime: "08:00",
		EndTime:   "13:00",
		IsActive:  boolPtr(true),
	})

	require.Error(t, err)
	assert.Nil(t, tpl)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeShiftTemplateNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "早番", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "shift_template '' already exists")
	assert.NotContains(t, err.Error(), "uk_shift_templates_clinic_name")
}

func TestShiftTemplateService_Update_NameConflictMapsDomainCode(t *testing.T) {
	name := "重複テンプレ"
	repo := &mockShiftTemplateRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
			return &model.ShiftTemplate{
				ID:        id,
				ClinicID:  clinicID,
				Name:      "旧",
				ShiftType: model.ShiftTypeMorning,
				StartTime: strPtr("08:00:00"),
				EndTime:   strPtr("13:00:00"),
			}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.ShiftTemplate, error) {
			return nil, uniqueShiftTemplateNameErr()
		},
	}
	svc := NewShiftTemplateService(repo)

	result, err := svc.Update(context.Background(), 10, 1, &UpdateShiftTemplateInput{Name: &name})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsNameConflict(err, apperrors.CodeShiftTemplateNameConflict))
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "重複テンプレ", appErr.Params["name"])
}

func TestShiftTemplateService_Create_GenericDBErrorNotElevated(t *testing.T) {
	repo := &mockShiftTemplateRepository{
		createFn: func(_ context.Context, _ *model.ShiftTemplate) error {
			return apperrors.WrapAlreadyExists("shift_template", "")
		},
	}
	svc := NewShiftTemplateService(repo)

	_, err := svc.Create(context.Background(), 10, &CreateShiftTemplateInput{
		Name:      "X",
		ShiftType: string(model.ShiftTypeMorning),
		StartTime: "08:00",
		EndTime:   "13:00",
		IsActive:  boolPtr(true),
	})

	require.Error(t, err)
	assert.False(t, apperrors.IsNameConflict(err, apperrors.CodeShiftTemplateNameConflict))
	assert.True(t, apperrors.IsAlreadyExists(err))
}
