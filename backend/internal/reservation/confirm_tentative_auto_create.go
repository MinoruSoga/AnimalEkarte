package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const confirmMedicalRecordTimeout = 30 * time.Second

// The reservation has committed before this independent, best-effort chart write.
// Preserve request trace values while dropping cancellation and the completed tx.
func (h *CRUDHandler) autoCreateMedicalRecordAfterConfirm(ctx context.Context, clinicID uint64, reservation *model.Reservation) {
	ctx, cancel := context.WithTimeout(persistence.DetachTx(ctx), confirmMedicalRecordTimeout)
	defer cancel()
	h.medicalRecord.AutoCreateFromReservation(ctx, clinicID, reservation)
}
