package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// nested_summary_response.go — medicalrecord-local nested DTOs for checkup/vaccination
// JSON (exported for tygo). Field sets are domain-specific and must not be merged with
// billing/reservation summaries.

// StaffSummaryResponse is the nested staff JSON for medicalrecord responses.
type StaffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary returns nil when s is nil.
func toStaffSummary(s *model.Staff) *StaffSummaryResponse {
	if s == nil {
		return nil
	}
	return &StaffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// OwnerSummaryResponse is the nested owner JSON for medicalrecord responses.
type OwnerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary returns nil when o is nil. 固定シグネチャの変換関数のため R-NIL（EMR-491）に従い MustOwnerBasic を使う。
// 取得経路（medicalrecord の feeding repository）が hydrate すること。未 hydrate は panic（500）で、空値にはしない。
func toOwnerSummary(o *model.Owner) *OwnerSummaryResponse {
	if o == nil {
		return nil
	}
	return &OwnerSummaryResponse{
		ID:        o.ID,
		OwnerName: commonbasic.MustOwnerBasic(o).Name,
	}
}

// AnimalSpeciesSummaryResponse is the nested species JSON for medicalrecord responses.
type AnimalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// PetSummaryResponse is the nested pet JSON for vaccination/checkup lists.
type PetSummaryResponse struct {
	ID        uint64              `json:"id"`
	Name      string              `json:"name"`
	PetNumber string              `json:"pet_number"`
	Weight    *model.ExactNumeric `json:"weight"`
	Status    *string             `json:"status"`
	Breed     *string             `json:"breed"`

	Gender         *string                       `json:"gender"`
	BirthDate      *string                       `json:"birth_date"`
	DeceasedAt     *time.Time                    `json:"deceased_at"`
	NeuteredDate   *string                       `json:"neutered_date"`
	NeuteredStatus string                        `json:"neutered_status,omitempty"`
	AnimalSpecies  *AnimalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	Owner          *OwnerSummaryResponse         `json:"owner,omitempty"`
}

// toPetSummary returns nil when p is nil.
// 固定シグネチャの変換関数のため R-NIL（EMR-491）に従い MustPetBasic を使う。取得経路（medicalrecord の feeding repository）が
// hydrate すること。未 hydrate は panic（500）で、空値にはしない。
func toPetSummary(p *model.Pet) *PetSummaryResponse {
	if p == nil {
		return nil
	}
	basic := commonbasic.MustPetBasic(p)
	resp := &PetSummaryResponse{
		ID:        p.ID,
		Name:      basic.Name,
		PetNumber: p.PetNumber,
		Weight:    copyBasicWeight(basic.RegistrationWeight),
		Status:    nullableEnumString(basic.Status),
		Breed:     basic.Breed,
		Gender:    nullableEnumString(basic.Gender),
		Owner:     toOwnerSummary(p.Owner),
	}
	if basic.BirthDate != nil {
		d := basic.BirthDate.Format("2006-01-02")
		resp.BirthDate = &d
	}
	resp.DeceasedAt = httpapi.LocalTimePtr(basic.DeceasedAt)
	if basic.NeuteredDate != nil {
		d := basic.NeuteredDate.Format("2006-01-02")
		resp.NeuteredDate = &d
	}
	if basic.NeuteredStatus != "" {
		resp.NeuteredStatus = string(basic.NeuteredStatus)
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &AnimalSpeciesSummaryResponse{
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
