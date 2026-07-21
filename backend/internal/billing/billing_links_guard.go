package billing

// billing_links_guard.go — BE9-2C B②: service/accounting_service_builders.go から
// AUD-005 系の請求リンク整合ガードを移動（accounting 残留 consumer は delegate 経由・B④で解消）。

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func AssertBillingLinksMatchMedicalRecord(mr *model.MedicalRecord, ownerID, petID *uint64) error {
	if ownerID != nil && mr.OwnerID != nil && *ownerID != *mr.OwnerID {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", mr.ID))
	}
	if petID != nil && mr.PetID != nil && *petID != *mr.PetID {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", mr.ID))
	}
	return nil
}
