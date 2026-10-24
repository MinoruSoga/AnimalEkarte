package reservation

import (
	"time"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go — reservation-local nested DTOs. Pet summary includes
// danger_level and must not be merged with medicalrecord/billing summaries.

// staffSummaryResponse is nested staff JSON for reservation responses.
type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary returns nil when s is nil.
func toStaffSummary(s *model.Staff) *staffSummaryResponse {
	if s == nil {
		return nil
	}
	return &staffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// ownerSummaryResponse is nested owner JSON for reservation responses.
// IsDangerous is staff-facing only (EMR-173 危険人物 mark); it must not be added to
// LIFF/owner-facing owner summaries.
type ownerSummaryResponse struct {
	ID          uint64 `json:"id"`
	OwnerName   string `json:"name"`
	IsDangerous bool   `json:"is_dangerous"`
}

// toOwnerSummary returns nil when o is nil. 固定シグネチャの変換関数のため R-NIL（EMR-491）に従い MustOwnerBasic を使う。
// 取得経路（reservation の feeding repository）が hydrate すること。未 hydrate は panic（500）で、空値にはしない。
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:          o.ID,
		OwnerName:   commonbasic.MustOwnerBasic(o).Name,
		IsDangerous: o.IsDangerous,
	}
}

// animalSpeciesSummaryResponse is nested species JSON for reservation responses.
type animalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// petSummaryResponse is the staff-facing reservation pet summary. DangerLevel intentionally
// exists only in this domain copy to support the Reception clinical-safety badge.
type petSummaryResponse struct {
	ID           uint64              `json:"id"`
	Name         string              `json:"name"`
	PetNumber    string              `json:"pet_number"`
	Weight       *model.ExactNumeric `json:"weight"`
	Status       *string             `json:"status"`
	BirthDate    *string             `json:"birth_date"`
	DeceasedAt   *time.Time          `json:"deceased_at"`
	DangerLevel  *string             `json:"danger_level"`
	DangerReason *string             `json:"danger_reason"`
	Breed        *string             `json:"breed"`

	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner         *ownerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary maps a pet into the reservation-local summary. nil の場合は nil を返す。
// 固定シグネチャの変換関数のため R-NIL（EMR-491）に従い MustPetBasic を使う。取得経路（reservation の feeding repository）が
// hydrate すること。未 hydrate は panic（500）で、空値にはしない。
func toPetSummary(p *model.Pet) *petSummaryResponse {
	if p == nil {
		return nil
	}
	basic := commonbasic.MustPetBasic(p)
	resp := &petSummaryResponse{
		ID:           p.ID,
		Name:         basic.Name,
		PetNumber:    p.PetNumber,
		Weight:       copyBasicWeight(basic.RegistrationWeight),
		Status:       nullableEnumString(basic.Status),
		DeceasedAt:   httpapi.LocalTimePtr(basic.DeceasedAt),
		DangerLevel:  nullableEnumString(basic.DangerLevel),
		DangerReason: basic.DangerReason,
		Breed:        basic.Breed,
		Owner:        toOwnerSummary(p.Owner),
	}
	if basic.BirthDate != nil {
		d := basic.BirthDate.Format(time.DateOnly)
		resp.BirthDate = &d
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   p.AnimalSpecies.ID,
			Name: p.AnimalSpecies.Name,
		}
	}
	return resp
}

// nullableEnumString は名前付き文字列型の nullable 値を *string へ写す。nil は nil のまま返し、空文字へは置換しない。
func nullableEnumString[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	text := string(*value)
	return &text
}

// copyBasicWeight は登録時体重（nullable）を別インスタンスとして返す。nil は nil のまま（JSON null）。
func copyBasicWeight(value *model.ExactNumeric) *model.ExactNumeric {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
