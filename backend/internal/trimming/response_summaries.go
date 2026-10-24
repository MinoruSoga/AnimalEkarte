package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toStaffSummary(staff *model.Staff) *staffSummaryResponse {
	if staff == nil {
		return nil
	}
	return &staffSummaryResponse{ID: staff.ID, Name: staff.Name}
}

type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary は固定シグネチャの変換関数（呼び出しが trimming の response 組み立て全体に渡る）のため、
// R-NIL に従い MustOwnerBasic を使う。取得経路（reservation の feeding repository）が必ず hydrate すること。
func toOwnerSummary(owner *model.Owner) *ownerSummaryResponse {
	if owner == nil {
		return nil
	}
	return &ownerSummaryResponse{ID: owner.ID, OwnerName: commonbasic.MustOwnerBasic(owner).Name}
}

type animalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type petSummaryResponse struct {
	ID            uint64                        `json:"id"`
	Name          string                        `json:"name"`
	PetNumber     string                        `json:"pet_number"`
	Weight        *model.ExactNumeric           `json:"weight"`
	Status        *string                       `json:"status"`
	Breed         *string                       `json:"breed"`
	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *ownerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary も R-NIL に従い MustPetBasic を使う。AnimalSpecies は HydratePets が revision の
// animal_species_id から載せたものを使う。weight は登録時体重（revision.registration_weight）で、
// nullable 列は JSON null のまま返す。
func toPetSummary(pet *model.Pet) *petSummaryResponse {
	if pet == nil {
		return nil
	}
	basic := commonbasic.MustPetBasic(pet)
	response := &petSummaryResponse{
		ID:        pet.ID,
		Name:      basic.Name,
		PetNumber: pet.PetNumber,
		Weight:    basic.RegistrationWeight,
		Status:    petStatusString(basic.Status),
		Breed:     basic.Breed,
		Owner:     toOwnerSummary(pet.Owner),
	}
	if pet.AnimalSpecies != nil {
		response.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   pet.AnimalSpecies.ID,
			Name: pet.AnimalSpecies.Name,
		}
	}
	return response
}

// petStatusString は nullable な status を JSON null のまま返す（空文字にしない）。
func petStatusString(status *model.PetStatus) *string {
	if status == nil {
		return nil
	}
	value := string(*status)
	return &value
}

func localTime(value time.Time) time.Time {
	return httpapi.LocalTime(value)
}
