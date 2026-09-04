package reservation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestShouldEnforceClosedDayConstraintOnUpdate_IgnoresShortcutRoute(t *testing.T) {
	for _, route := range []string{"reception", "exam_room", "record_shortcut"} {
		route := route
		t.Run(route, func(t *testing.T) {
			assert.True(t, shouldEnforceClosedDayConstraintOnUpdate(model.ReservationStatusPending, &route))
		})
	}
}
