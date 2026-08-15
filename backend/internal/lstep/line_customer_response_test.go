package lstep

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToLineCustomerResponse(t *testing.T) {
	t.Run("converts LINE customer with linked owner", func(t *testing.T) {
		ownerID := uint64(9)
		created := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
		updated := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
		m := &model.LineCustomer{
			ID:               1,
			ClinicID:         2,
			LineUserID:       "U1234567890",
			DisplayName:      "たなか",
			RealName:         "田中太郎",
			AdditionalFields: []byte(`{"note":"VIP"}`),
			OwnerID:          &ownerID,
			Owner:            &model.Owner{ID: ownerID, Name: "田中太郎"},
			CreatedAt:        created,
			UpdatedAt:        updated,
		}

		resp := toLineCustomerResponse(m)

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(2), resp.ClinicID)
		assert.Equal(t, "U1234567890", resp.LineUserID)
		assert.Equal(t, "たなか", resp.DisplayName)
		assert.Equal(t, "田中太郎", resp.RealName)
		assert.JSONEq(t, `{"note":"VIP"}`, string(resp.AdditionalFields))
		require := assert.New(t)
		require.NotNil(resp.OwnerID)
		require.Equal(ownerID, *resp.OwnerID)
		require.Equal("田中太郎", resp.OwnerName)
	})

	t.Run("converts LINE customer with no linked owner", func(t *testing.T) {
		m := &model.LineCustomer{
			ID:               3,
			ClinicID:         4,
			LineUserID:       "U0000000000",
			DisplayName:      "未連携ユーザー",
			AdditionalFields: []byte(`{}`),
			OwnerID:          nil,
			Owner:            nil,
		}

		resp := toLineCustomerResponse(m)

		assert.Nil(t, resp.OwnerID)
		assert.Equal(t, "", resp.OwnerName)
	})
}
