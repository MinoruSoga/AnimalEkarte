package pet

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type nullableStringRequestField struct {
	set   bool
	value *string
}

func (f *nullableStringRequestField) UnmarshalJSON(data []byte) error {
	f.set = true
	if string(data) == "null" {
		f.value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput("文字列の形式が正しくありません")
	}
	f.value = &value
	return nil
}

func (f nullableStringRequestField) toServiceInput() **string {
	if !f.set {
		return nil
	}
	return &f.value
}

type listPetFilters = PetListFilters

type listPetQuery struct {
	OwnerID         string
	Search          string
	Species         string
	IncludeDeceased string
}

func newListPetQuery(values url.Values) listPetQuery {
	return listPetQuery{
		OwnerID:         values.Get("owner_id"),
		Search:          values.Get("search"),
		Species:         values.Get("species"),
		IncludeDeceased: values.Get("include_deceased"),
	}
}

// parseOptionalBoolQueryFilter はクエリパラメータを optional な bool にパースする。
// 空文字は defaultVal を返す。パース失敗時は WrapInvalidInput を返す。
func parseOptionalBoolQueryFilter(value, field string, defaultVal bool) (bool, error) {
	if value == "" {
		return defaultVal, nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, apperrors.WrapInvalidInput("invalid " + field)
	}
	return b, nil
}

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func (q listPetQuery) toServiceFilters() (listPetFilters, error) {
	if len(q.Search) > 255 {
		return listPetFilters{}, apperrors.WrapInvalidInput("search must be at most 255 characters")
	}
	ownerID, err := parseOptionalUintQueryFilter(q.OwnerID, "owner_id")
	if err != nil {
		return listPetFilters{}, err
	}
	speciesID, err := parseOptionalUintQueryFilter(q.Species, "species")
	if err != nil {
		return listPetFilters{}, err
	}
	// #266: 生死フィルタのデフォルトは生存のみ（include_deceased 未指定 = false）。
	includeDeceased, err := parseOptionalBoolQueryFilter(q.IncludeDeceased, "include_deceased", false)
	if err != nil {
		return listPetFilters{}, err
	}
	return listPetFilters{
		OwnerID:         ownerID,
		Search:          q.Search,
		AnimalSpeciesID: speciesID,
		IncludeDeceased: includeDeceased,
	}, nil
}

// createPetRequest はペット作成のバインド struct
type createPetRequest struct {
	OwnerID         uint64    `json:"owner_id"          binding:"required"`
	AnimalSpeciesID uint64    `json:"animal_species_id" binding:"required"`
	Name            string    `json:"name"              binding:"required,max=255"`
	NameKana        string    `json:"name_kana"        binding:"omitempty,max=100"`
	Gender          string    `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	Status          string    `json:"status"            binding:"omitempty,oneof=alive deceased"`
	BirthDate       *jsonDate `json:"birth_date"`
	Breed           string    `json:"breed"            binding:"omitempty,max=100"`
	Color           string    `json:"color"            binding:"omitempty,max=100"`
	BloodType       string    `json:"blood_type"       binding:"omitempty,max=32"`
	MicrochipNumber string    `json:"microchip_number" binding:"omitempty,max=64"`
	Weight          *float64  `json:"weight"`
	NeuteredDate    *jsonDate `json:"neutered_date"`
	AcquisitionType string    `json:"acquisition_type"`
	DangerLevel     string    `json:"danger_level"`
	DangerReason    *string   `json:"danger_reason"    binding:"omitempty,max=500"`
	Food            string    `json:"food"             binding:"omitempty,max=500"`
	Environment     string    `json:"environment"      binding:"omitempty,max=500"`
	Phone           string    `json:"phone"            binding:"omitempty,max=30"`
	InsuranceID     *uint64   `json:"insurance_id"`
	Remarks         string    `json:"remarks"          binding:"omitempty,max=2000"`
}

func (r *createPetRequest) toServiceInput() *CreatePetInput {
	return &CreatePetInput{
		OwnerID:         r.OwnerID,
		AnimalSpeciesID: r.AnimalSpeciesID,
		Name:            r.Name,
		PetNameKana:     r.NameKana,
		Gender:          r.Gender,
		Status:          r.Status,
		BirthDate:       jsonDatePtr(r.BirthDate),
		Breed:           r.Breed,
		Color:           r.Color,
		BloodType:       r.BloodType,
		MicrochipNumber: r.MicrochipNumber,
		Weight:          r.Weight,
		NeuteredDate:    jsonDatePtr(r.NeuteredDate),
		AcquisitionType: r.AcquisitionType,
		DangerLevel:     r.DangerLevel,
		DangerReason:    r.DangerReason,
		Food:            r.Food,
		Environment:     r.Environment,
		Phone:           r.Phone,
		InsuranceID:     r.InsuranceID,
		Remarks:         r.Remarks,
	}
}

// updatePetRequest はペット更新のバインド struct（全フィールドポインタ型）
// InsuranceID は **uint64: JSON未送信=nil, "insurance_id":null=&nil, "insurance_id":123=&&123
type updatePetRequest struct {
	OwnerID         *uint64 `json:"owner_id"`
	AnimalSpeciesID *uint64 `json:"animal_species_id"`
	PetNumber       *string `json:"pet_number" binding:"omitempty,max=50"` // 自動採番後も手動変更可
	Name            *string `json:"name"              binding:"omitempty,max=255"`
	NameKana        *string `json:"name_kana"        binding:"omitempty,max=100"`
	Gender          *string `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	// Status は意図的に持たない(BUG-415)。status の書込は Create と /:id/death
	// (HandlePetDeath/HandlePetRevival、監査ログ+deceased_at 同一tx・fail-closed) に
	// 一本化済み。ここに "status" を JSON で送っても未知フィールドとして無視される。
	BirthDate       *jsonDate                  `json:"birth_date"`
	Breed           *string                    `json:"breed"            binding:"omitempty,max=100"`
	Color           *string                    `json:"color"            binding:"omitempty,max=100"`
	BloodType       *string                    `json:"blood_type"       binding:"omitempty,max=32"`
	MicrochipNumber *string                    `json:"microchip_number" binding:"omitempty,max=64"`
	Weight          *float64                   `json:"weight"`
	NeuteredDate    *jsonDate                  `json:"neutered_date"`
	AcquisitionType *string                    `json:"acquisition_type"`
	DangerLevel     *string                    `json:"danger_level"`
	DangerReason    nullableStringRequestField `json:"danger_reason"`
	Food            *string                    `json:"food"             binding:"omitempty,max=500"`
	Environment     *string                    `json:"environment"      binding:"omitempty,max=500"`
	Phone           *string                    `json:"phone"            binding:"omitempty,max=30"`
	LastVisit       *jsonDate                  `json:"last_visit"`
	InsuranceID     **uint64                   `json:"insurance_id"`
	Remarks         *string                    `json:"remarks"          binding:"omitempty,max=2000"`
}

func (r *updatePetRequest) toServiceInput() *UpdatePetInput {
	return &UpdatePetInput{
		OwnerID:         r.OwnerID,
		AnimalSpeciesID: r.AnimalSpeciesID,
		PetNumber:       r.PetNumber,
		Name:            r.Name,
		PetNameKana:     r.NameKana,
		Gender:          r.Gender,
		BirthDate:       jsonDatePtr(r.BirthDate),
		Breed:           r.Breed,
		Color:           r.Color,
		BloodType:       r.BloodType,
		MicrochipNumber: r.MicrochipNumber,
		Weight:          r.Weight,
		NeuteredDate:    jsonDatePtr(r.NeuteredDate),
		AcquisitionType: r.AcquisitionType,
		DangerLevel:     r.DangerLevel,
		DangerReason:    r.DangerReason.toServiceInput(),
		Food:            r.Food,
		Environment:     r.Environment,
		Phone:           r.Phone,
		LastVisit:       jsonDatePtr(r.LastVisit),
		InsuranceID:     r.InsuranceID,
		Remarks:         r.Remarks,
	}
}
