package apperrors

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsNameUniqueConflict_PermissionGroupName(t *testing.T) {
	// Simulate FromGORM-wrapped unique_violation with measured constraint.
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: ConstraintPermissionGroupName,
	}
	fromGORM := FromGORM(pgErr, "permission_group", "")
	require.True(t, IsAlreadyExists(fromGORM))

	// pgErr must remain recoverable for fail-closed ConstraintName mapping.
	var recovered *pgconn.PgError
	require.True(t, errors.As(fromGORM, &recovered))
	assert.Equal(t, ConstraintPermissionGroupName, recovered.ConstraintName)

	conflict := AsNameUniqueConflict(
		fromGORM,
		"執行",
		ConstraintPermissionGroupName,
		CodePermissionGroupNameConflict,
	)
	require.Error(t, conflict)
	assert.True(t, IsNameConflict(conflict, CodePermissionGroupNameConflict))
	assert.True(t, IsConflict(conflict))

	var appErr *AppError
	require.True(t, errors.As(conflict, &appErr))
	assert.Equal(t, CodePermissionGroupNameConflict, appErr.Code)
	assert.Equal(t, "執行", appErr.Params["name"])
	assert.NotContains(t, appErr.Message, "uk_permission")
	assert.NotContains(t, appErr.Message, "permission_group '' already exists")
	assert.Equal(t, "resource already exists", appErr.Message)
}

func TestAsNameUniqueConflict_DoesNotMapRulesConstraint(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: ConstraintPermissionGroupRules,
	}
	fromGORM := FromGORM(pgErr, "permission_group_rule", "")

	conflict := AsNameUniqueConflict(
		fromGORM,
		"執行",
		ConstraintPermissionGroupName,
		CodePermissionGroupNameConflict,
	)
	assert.Nil(t, conflict, "rules unique_violation must not elevate to name conflict")
	assert.True(t, IsAlreadyExists(fromGORM))
}

func TestAsNameUniqueConflict_MissingConstraintNameFailClosed(t *testing.T) {
	// AlreadyExists without recoverable pg constraint — do not elevate.
	err := WrapAlreadyExists("permission_group", "")
	conflict := AsNameUniqueConflict(
		err,
		"執行",
		ConstraintPermissionGroupName,
		CodePermissionGroupNameConflict,
	)
	assert.Nil(t, conflict)
}

func TestAsNameUniqueConflict_AnimalSpeciesName(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: ConstraintAnimalSpeciesName,
	}
	fromGORM := FromGORM(pgErr, "animal_species", "")
	conflict := AsNameUniqueConflict(
		fromGORM,
		"V04動物種類",
		ConstraintAnimalSpeciesName,
		CodeAnimalSpeciesNameConflict,
	)
	require.Error(t, conflict)
	assert.True(t, IsNameConflict(conflict, CodeAnimalSpeciesNameConflict))
	var appErr *AppError
	require.True(t, errors.As(conflict, &appErr))
	assert.Equal(t, "V04動物種類", appErr.Params["name"])
}

func TestAsNameUniqueConflict_ShiftTemplateAndLstepPrefix(t *testing.T) {
	shift := AsNameUniqueConflict(
		FromGORM(&pgconn.PgError{Code: "23505", ConstraintName: ConstraintShiftTemplateName}, "shift_template", ""),
		"早番",
		ConstraintShiftTemplateName,
		CodeShiftTemplateNameConflict,
	)
	require.Error(t, shift)
	assert.True(t, IsNameConflict(shift, CodeShiftTemplateNameConflict))
	assert.True(t, RespondWithConflictCode(shift))

	prefix := AsNameUniqueConflict(
		FromGORM(&pgconn.PgError{Code: "23505", ConstraintName: ConstraintLstepAutoManagedPrefix}, "lstep_auto_managed_prefix", ""),
		"checkup_",
		ConstraintLstepAutoManagedPrefix,
		CodeLstepAutoManagedPrefixConflict,
	)
	require.Error(t, prefix)
	assert.True(t, IsNameConflict(prefix, CodeLstepAutoManagedPrefixConflict))
	assert.True(t, RespondWithConflictCode(prefix))
}

func TestAsNameUniqueConflict_TreatmentItemNames(t *testing.T) {
	cases := []struct {
		constraint string
		code       string
		resource   string
		name       string
	}{
		{ConstraintConsultationName, CodeConsultationNameConflict, "consultation", "V04診察"},
		{ConstraintExamTypeName, CodeExamTypeNameConflict, "examination_type", "V04検査"},
		{ConstraintProcedureName, CodeProcedureNameConflict, "procedure", "V04処置"},
		{ConstraintVaccineName, CodeVaccineNameConflict, "vaccine", "V04予防接種"},
		{ConstraintCheckupTypeName, CodeCheckupTypeNameConflict, "checkup_type", "V04定期健診"},
		{ConstraintMedicineName, CodeMedicineNameConflict, "medicine", "V04薬剤"},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			conflict := AsNameUniqueConflict(
				FromGORM(&pgconn.PgError{Code: "23505", ConstraintName: tc.constraint}, tc.resource, ""),
				tc.name,
				tc.constraint,
				tc.code,
			)
			require.Error(t, conflict)
			assert.True(t, IsNameConflict(conflict, tc.code))
			assert.True(t, RespondWithConflictCode(conflict))
			var appErr *AppError
			require.True(t, errors.As(conflict, &appErr))
			assert.Equal(t, tc.name, appErr.Params["name"])
			assert.NotContains(t, appErr.Message, tc.resource+" '' already exists")
		})
	}
}

func TestConflictHTTPExtras_AndRespondFlag(t *testing.T) {
	err := WrapNameConflict(CodePermissionGroupNameConflict, "執行")
	extras := ConflictHTTPExtras(err)
	require.NotNil(t, extras)
	params, ok := extras["params"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "執行", params["name"])
	assert.True(t, RespondWithConflictCode(err))

	// Empty name → params map empty → no extras, but still needs code response.
	emptyName := WrapNameConflict(CodeAnimalSpeciesNameConflict, "")
	assert.Nil(t, ConflictHTTPExtras(emptyName))
	assert.True(t, RespondWithConflictCode(emptyName))

	// Generic already-exists must not force WithExtras.
	assert.False(t, RespondWithConflictCode(WrapAlreadyExists("x", "")))
	assert.Nil(t, ConflictHTTPExtras(WrapAlreadyExists("x", "")))
}

func TestFromGORM_UniquePreservesPgErrorForConstraintMapping(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uk_permission_groups",
		TableName:      "permission_groups",
		Detail:         "Key (clinic_id, name)=(1, x) already exists.",
	}
	err := FromGORM(pgErr, "permission_group", "")
	assert.True(t, IsAlreadyExists(err))
	// Response-facing Message must not embed constraint/table/SQL detail.
	var appErr *AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "permission_group '' already exists", appErr.Message)
	assert.NotContains(t, appErr.Message, "uk_permission")
	assert.NotContains(t, appErr.Message, "permission_groups")
	assert.NotContains(t, appErr.Message, "Key (")

	// Wrapped further (service style) still exposes pg constraint.
	wrapped := Wrap(err, "failed to create permission group")
	var recovered *pgconn.PgError
	require.True(t, errors.As(wrapped, &recovered))
	assert.Equal(t, "uk_permission_groups", recovered.ConstraintName)
}
