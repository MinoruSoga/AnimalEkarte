package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToBillingConfirmationResponse(t *testing.T) {
	confirmedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	returnedAt := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	confirmedBy := uint64(3)
	returnedBy := uint64(4)

	tests := []struct {
		name  string
		input *model.BillingConfirmation
		check func(t *testing.T, resp billingConfirmationResponse)
	}{
		{
			name: "pending confirmation with nil pointers",
			input: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 5,
				Status:          model.ConfirmationStatusPending,
				ReturnReason:    "",
				Memo:            "",
			},
			check: func(t *testing.T, resp billingConfirmationResponse) {
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "5", resp.MedicalRecordID)
				assert.Equal(t, "pending", resp.Status)
				assert.Nil(t, resp.ConfirmedBy)
				assert.Nil(t, resp.ConfirmedAt)
				assert.Nil(t, resp.ReturnedBy)
				assert.Nil(t, resp.ReturnedAt)
			},
		},
		{
			name: "confirmed with ConfirmedBy and ConfirmedAt populated",
			input: &model.BillingConfirmation{
				ID:              2,
				MedicalRecordID: 6,
				Status:          model.ConfirmationStatusConfirmed,
				ConfirmedBy:     &confirmedBy,
				ConfirmedAt:     &confirmedAt,
				Memo:            "確認済み",
			},
			check: func(t *testing.T, resp billingConfirmationResponse) {
				assert.Equal(t, "confirmed", resp.Status)
				require.NotNil(t, resp.ConfirmedBy)
				assert.Equal(t, "3", *resp.ConfirmedBy)
				require.NotNil(t, resp.ConfirmedAt)
				assert.Equal(t, "確認済み", resp.Memo)
			},
		},
		{
			name: "returned with ReturnedBy and ReturnedAt populated",
			input: &model.BillingConfirmation{
				ID:              3,
				MedicalRecordID: 7,
				Status:          model.ConfirmationStatusReturned,
				ReturnedBy:      &returnedBy,
				ReturnedAt:      &returnedAt,
				ReturnReason:    "金額確認必要",
			},
			check: func(t *testing.T, resp billingConfirmationResponse) {
				assert.Equal(t, "returned", resp.Status)
				require.NotNil(t, resp.ReturnedBy)
				assert.Equal(t, "4", *resp.ReturnedBy)
				require.NotNil(t, resp.ReturnedAt)
				assert.Equal(t, "金額確認必要", resp.ReturnReason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toBillingConfirmationResponse(tt.input)
			tt.check(t, resp)
		})
	}
}
