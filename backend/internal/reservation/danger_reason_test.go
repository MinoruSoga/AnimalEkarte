package reservation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestDangerReason_StaffReservationSummaryIncludesField(t *testing.T) {
	reason := "診察台で噛む"
	response := toPetSummary(&model.Pet{
		ID:           20,
		Name:         "ポチ",
		DangerLevel:  model.DangerLevelHigh,
		DangerReason: &reason,
	})

	body, err := json.Marshal(response)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.Equal(t, reason, payload["danger_reason"])
}

func TestDangerReason_LiffPetSerializersOmitField(t *testing.T) {
	reason := "staff only"
	profileBody, err := json.Marshal(toLiffProfileResponse(&model.LineCustomer{
		Owner: &model.Owner{
			Pets: []model.Pet{{
				ID:           1,
				Name:         "ポチ",
				DangerReason: &reason,
			}},
		},
	}))
	require.NoError(t, err)
	assertJSONDoesNotContainDangerReason(t, "toLiffProfileResponse", profileBody)

	ownerID := uint64(2)
	healthCardService := newHealthCardTestService(
		&mockLiffCustomerRepository{
			findByIDFn: func(_ context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:       customerID,
					ClinicID: clinicID,
					OwnerID:  &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Name: "飼主",
						Pets: []model.Pet{{
							ID:           1,
							ClinicID:     clinicID,
							OwnerID:      ownerID,
							Name:         "ポチ",
							DangerLevel:  model.DangerLevelHigh,
							DangerReason: &reason,
						}},
					},
				}, nil
			},
		},
		&mockVaccinationRepository{
			findByOwnerFn: func(_ context.Context, _, _ uint64) ([]model.Vaccination, error) {
				return []model.Vaccination{}, nil
			},
		},
	)
	healthCard, err := healthCardService.GetHealthCard(context.Background(), 1, 3)
	require.NoError(t, err)
	require.Len(t, healthCard.Pets, 1)
	healthCardBody, err := json.Marshal(toLiffHealthCardResponse(healthCard))
	require.NoError(t, err)
	assertJSONDoesNotContainDangerReason(t, "toLiffHealthCardResponse", healthCardBody)
}

func assertJSONDoesNotContainDangerReason(t *testing.T, serializer string, body []byte) {
	t.Helper()
	var payload any
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.NotContains(t, string(body), `"danger_reason"`,
		"%s must physically omit danger_reason", serializer)
}
