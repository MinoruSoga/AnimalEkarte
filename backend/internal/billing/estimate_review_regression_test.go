package billing

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimateUpdatePreparationDoesNotMutateRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithEstimateSvc(&mockEstimateService{
		getFn: func(context.Context, uint64, uint64) (*model.Estimate, error) {
			return &model.Estimate{ID: 1, Items: []model.EstimateItem{
				{ID: 7, EstimateID: 1, Name: "合成品目", SortOrder: 0},
			}}, nil
		},
	})
	items := []createEstimateItemRequest{{Name: "合成品目", Quantity: 1}}
	req := updateEstimateRequest{Items: &items}
	c, _ := gin.CreateTestContext(nil)
	c.Request = httptest.NewRequest("PATCH", "/", nil)
	setClinicID(c)
	setStaffID(c)

	prepared, err := h.prepareUpdateEstimateRequest(c, 1, 1, &req)
	require.NoError(t, err)
	assert.Nil(t, items[0].ExistingItemID, "preparing a legacy row must leave caller input unchanged")
	assert.Nil(t, (*req.Items)[0].ExistingItemID)
	require.NotNil(t, prepared.Items)
	require.NotNil(t, (*prepared.Items)[0].ExistingItemID)
	assert.Equal(t, uint64(7), *(*prepared.Items)[0].ExistingItemID)
}
