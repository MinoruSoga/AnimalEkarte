package reservation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToReservationStaffResponseEmitsNonNilEmptyCapableCourses(t *testing.T) {
	response := toReservationStaffResponse(&model.Staff{}, nil, nil)

	require.NotNil(t, response.CapableCourses)
	assert.Empty(t, response.CapableCourses)

	payload, err := json.Marshal(response)
	require.NoError(t, err)
	assert.Contains(t, string(payload), `"capable_courses":[]`)
}

func TestToReservationStaffResponseMapsCapableCourses(t *testing.T) {
	reservationType := &model.ReservationType{ID: 42, Name: "一般診療"}
	response := toReservationStaffResponse(
		&model.Staff{ID: 7, Name: "担当者"},
		nil,
		[]model.StaffReservationCapability{
			{ReservationTypeID: reservationType.ID, ReservationType: reservationType},
			{ReservationTypeID: 99},
		},
	)

	require.Len(t, response.CapableCourses, 2)
	assert.Equal(t, capableCourseResponse{ID: 42, Name: "一般診療"}, response.CapableCourses[0])
	assert.Equal(t, capableCourseResponse{ID: 99}, response.CapableCourses[1])
}
