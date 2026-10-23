package main

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
)

type wiringLineReservationSettingRepository struct {
	reservation.LineReservationSettingRepository
	persisted *model.LineReservationSetting
}

func (r *wiringLineReservationSettingRepository) FindByClinicID(
	context.Context,
	uint64,
) (*model.LineReservationSetting, error) {
	return r.persisted, nil
}

func (r *wiringLineReservationSettingRepository) FindAll(
	context.Context,
) ([]model.LineReservationSetting, error) {
	if r.persisted == nil {
		return nil, nil
	}
	return []model.LineReservationSetting{*r.persisted}, nil
}

func (r *wiringLineReservationSettingRepository) Save(
	_ context.Context,
	_ uint64,
	setting *model.LineReservationSetting,
) error {
	r.persisted = setting
	return nil
}

func TestNewReservationCompositionBuildsTargetOwnedGraph(t *testing.T) {
	gin.SetMode(gin.TestMode)

	occupations := staff.NewOccupationRepository(nil)
	repositories := newReservationRepositories(
		nil,
		staff.NewRepository(nil),
		staff.NewShiftEntryRepository(nil),
		occupations,
	)
	require.NotNil(t, repositories.Reservations)
	require.NotNil(t, repositories.ReservationStaff)
	require.NotNil(t, repositories.ReservationSchedules)
	require.Same(t, occupations, repositories.Occupations)

	composition := newReservationComposition(
		repositories,
		reservationServiceDependencies{},
	)
	require.NotNil(t, composition.Reservations)
	require.NotNil(t, composition.ReservationTypes)
	require.NotNil(t, composition.ReservationStaff)
	require.NotNil(t, composition.Liff)
	require.NotNil(t, composition.Notifier)
	require.NotNil(t, composition.DrainNotifications)
	require.NotPanics(t, composition.DrainNotifications)
	require.NotPanics(t, reservationNotificationDrainer(nil))

	noop := func(c *gin.Context) {
		c.Next()
	}
	handler := composition.newHandler(reservationHandlerDependencies{
		LiffAuth: noop,
		LiffRateLimit: func(_ int) gin.HandlerFunc {
			return noop
		},
		LinkLiffAccount: noop,
		RequirePermission: func(_, _ string) gin.HandlerFunc {
			return noop
		},
	})
	require.NotNil(t, handler)

	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))
	handler.RegisterLiffRoutes(router)
	// 64 → 67: EMR-308 adds PUT/DELETE /reservations/pets/:petId/next-visit-intent
	// and POST /reservations/:id/confirm-tentative.
	require.Len(t, router.Routes(), 67)
}

func TestNewReservationComposition_LineSettingsSaveHasNoCredentialWiring(t *testing.T) {
	repository := &wiringLineReservationSettingRepository{
		persisted: &model.LineReservationSetting{ClinicID: 7},
	}
	composition := newReservationComposition(
		reservationRepositories{
			LineReservationSettings: repository,
		},
		reservationServiceDependencies{},
	)

	_, _, err := composition.LineReservationSettings.Save(
		context.Background(),
		7,
		&reservation.UpsertLineReservationSettingInput{},
	)

	require.NoError(t, err)
	require.NotNil(t, repository.persisted)
	// EMR-504: LINE credentials are no longer part of the reservation setting (clinic_integrations
	// is their only store), so the composed Save path needs no cipher closures and writes none.
	assert.Equal(t, uint64(7), repository.persisted.ClinicID)
}
