package apicontract

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type reservationStaffContractProperty struct {
	Type       string `yaml:"type"`
	Deprecated bool   `yaml:"deprecated"`
	Items      struct {
		Ref string `yaml:"$ref"`
	} `yaml:"items"`
}

type reservationStaffContractSchema struct {
	Type       string                                      `yaml:"type"`
	Deprecated bool                                        `yaml:"deprecated"`
	Required   []string                                    `yaml:"required"`
	Properties map[string]reservationStaffContractProperty `yaml:"properties"`
}

type reservationStaffContractOperation struct {
	Deprecated bool `yaml:"deprecated"`
}

type reservationStaffContractPath struct {
	Get *reservationStaffContractOperation `yaml:"get"`
	Put *reservationStaffContractOperation `yaml:"put"`
}

type reservationStaffContractSpec struct {
	Components struct {
		Schemas map[string]reservationStaffContractSchema `yaml:"schemas"`
	} `yaml:"components"`
	Paths map[string]reservationStaffContractPath `yaml:"paths"`
}

func loadReservationStaffContractSpec(t *testing.T) reservationStaffContractSpec {
	t.Helper()

	src, err := os.ReadFile(openapiPath) //nolint:gosec // fixed repository contract path
	require.NoError(t, err)

	var spec reservationStaffContractSpec
	require.NoError(t, yaml.Unmarshal(src, &spec))
	return spec
}

func TestOpenAPIReservationStaffUsesPositiveCapabilityContract(t *testing.T) {
	spec := loadReservationStaffContractSpec(t)

	capableCourse, ok := spec.Components.Schemas["CapableCourse"]
	require.True(t, ok, "CapableCourse schema must be documented")
	assert.Equal(t, "object", capableCourse.Type)
	assert.False(t, capableCourse.Deprecated)
	assert.ElementsMatch(t, []string{"id", "name"}, capableCourse.Required)
	assert.Contains(t, capableCourse.Properties, "id")
	assert.Contains(t, capableCourse.Properties, "name")

	staff, ok := spec.Components.Schemas["ReservationStaff"]
	require.True(t, ok, "ReservationStaff schema must be documented")
	assert.Contains(t, staff.Required, "capable_courses")
	capableCourses, ok := staff.Properties["capable_courses"]
	require.True(t, ok, "ReservationStaff.capable_courses must be documented")
	assert.Equal(t, "array", capableCourses.Type)
	assert.Equal(t, "#/components/schemas/CapableCourse", capableCourses.Items.Ref)
	assert.False(t, capableCourses.Deprecated)

	capablePath, ok := spec.Paths["/masters/staffs/{id}/capable-reservation-types"]
	require.True(t, ok, "capability master path must remain documented")
	assert.NotNil(t, capablePath.Get)
	assert.NotNil(t, capablePath.Put)
}

func TestOpenAPIReservationStaffRetainsDeprecatedExclusionCompatibility(t *testing.T) {
	spec := loadReservationStaffContractSpec(t)

	excludedCourse, ok := spec.Components.Schemas["ExcludedCourse"]
	require.True(t, ok, "legacy ExcludedCourse schema must remain during deprecation")
	assert.True(t, excludedCourse.Deprecated)
	assert.Contains(t, spec.Components.Schemas, "StaffReservationExclusion")

	staff := spec.Components.Schemas["ReservationStaff"]
	excludedCourses, ok := staff.Properties["excluded_courses"]
	require.True(t, ok, "legacy ReservationStaff.excluded_courses must remain during deprecation")
	assert.True(t, excludedCourses.Deprecated)
	assert.Equal(t, "#/components/schemas/ExcludedCourse", excludedCourses.Items.Ref)

	for _, schemaName := range []string{"CreateReservationStaffRequest", "UpdateReservationStaffRequest"} {
		schema, ok := spec.Components.Schemas[schemaName]
		require.True(t, ok, "%s schema must remain documented", schemaName)
		excludedTypeIDs, ok := schema.Properties["excluded_type_ids"]
		require.True(t, ok, "%s.excluded_type_ids must remain during deprecation", schemaName)
		assert.True(t, excludedTypeIDs.Deprecated)
	}

	legacyPath, ok := spec.Paths["/masters/staffs/{id}/excluded-reservation-types"]
	require.True(t, ok, "legacy exclusion path must remain during deprecation")
	require.NotNil(t, legacyPath.Get, "legacy exclusion GET must remain during deprecation")
	require.NotNil(t, legacyPath.Put, "legacy exclusion PUT must remain during deprecation")
	assert.True(t, legacyPath.Get.Deprecated)
	assert.True(t, legacyPath.Put.Deprecated)
}
