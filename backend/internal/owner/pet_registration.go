package owner

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// PetRegistrationDraft is the owner use case's immutable request for one
// nested pet. Identity, tenant, numbering, and associations are intentionally
// absent because the pet write owner derives them inside its transaction.
type PetRegistrationDraft struct {
	AnimalSpeciesID uint64
	Name            string
	NameKana        string
	Breed           string
	Color           string
	BloodType       *string
	MicrochipNumber *string
	Gender          model.PetGender
	Status          model.PetStatus
	BirthDate       *time.Time
	Weight          *float64
	NeuteredDate    *time.Time
	AcquisitionType *model.AcquisitionType
	DangerLevel     model.DangerLevel
	DangerReason    *string
	Food            string
	Environment     string
	Phone           string
	InsuranceID     *uint64
	Remarks         string
}

// PetRegistrationIntent is the owner-owned cross-domain command. The pet
// adapter translates it to the pet package's write-owner vocabulary.
type PetRegistrationIntent struct {
	ClinicID uint64
	OwnerID  uint64
	Pets     []PetRegistrationDraft
}

// PetRegistrar is the consumer-side capability used by owner registration.
// Implementations must join the ambient transaction and fail closed when it is
// absent.
type PetRegistrar interface {
	CreateForOwnerRegistration(
		ctx context.Context,
		intent PetRegistrationIntent,
	) ([]model.Pet, error)
}
