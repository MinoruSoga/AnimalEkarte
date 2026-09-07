package main

import (
	"context"
	"strings"
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
	require.Len(t, router.Routes(), 64)
}

func TestNewReservationComposition_InjectsLineCredentialCipherClosures(t *testing.T) {
	repository := &wiringLineReservationSettingRepository{
		persisted: &model.LineReservationSetting{
			ClinicID:          7,
			LineChannelSecret: "encrypted:secret",
			LineAccessToken:   "encrypted:token",
		},
	}
	composition := newReservationComposition(
		reservationRepositories{
			LineReservationSettings: repository,
		},
		reservationServiceDependencies{
			EncryptCredential: func(value string) (string, error) {
				return "rewired:" + value, nil
			},
			DecryptCredential: func(_ context.Context, value string) string {
				return strings.TrimPrefix(value, "encrypted:")
			},
		},
	)

	_, _, err := composition.LineReservationSettings.Save(
		context.Background(),
		7,
		&reservation.UpsertLineReservationSettingInput{},
	)

	require.NoError(t, err)
	require.NotNil(t, repository.persisted)
	// R-05 Phase B: reservation settings path no longer re-encrypts or writes
	// LineChannelSecret (canonical SoT is clinic_integrations). OnConflict also
	// excludes the column, so empty model field must not assert a rewired secret.
	assert.Empty(t, repository.persisted.LineChannelSecret)
	assert.Equal(t, "rewired:token", repository.persisted.LineAccessToken)
}
