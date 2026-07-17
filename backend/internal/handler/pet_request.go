package handler

import (
	"net/url"
	"strconv"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/service"
)

// listPetFilters は互換のためのエイリアス。実体は repository.PetListFilters
// （medical_record_request.go の listMedicalRecordFilters と同型パターン）。
type listPetFilters = repository.PetListFilters

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

func (q listPetQuery) toServiceFilters() (listPetFilters, error) {
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
	Name            string    `json:"name"              binding:"required"`
	NameKana        string    `json:"name_kana"`
	Gender          string    `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	Status          string    `json:"status"            binding:"omitempty,oneof=alive deceased"`
	BirthDate       *jsonDate `json:"birth_date"`
	Breed           string    `json:"breed"`
	Color           string    `json:"color"`
	BloodType       string    `json:"blood_type"       binding:"omitempty,max=32"`
	MicrochipNumber string    `json:"microchip_number" binding:"omitempty,max=64"`
	Weight          *float64  `json:"weight"`
	NeuteredDate    *jsonDate `json:"neutered_date"`
	AcquisitionType string    `json:"acquisition_type"`
	DangerLevel     string    `json:"danger_level"`
	Food            string    `json:"food"`
	Environment     string    `json:"environment"`
	Phone           string    `json:"phone"`
	InsuranceID     *uint64   `json:"insurance_id"`
	Remarks         string    `json:"remarks"`
}

func (r *createPetRequest) toServiceInput() *service.CreatePetInput {
	return &service.CreatePetInput{
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
	PetNumber       *string `json:"pet_number"` // 自動採番後も手動変更可
	Name            *string `json:"name"`
	NameKana        *string `json:"name_kana"`
	Gender          *string `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	// Status は意図的に持たない(BUG-415)。status の書込は Create と /:id/death
	// (HandlePetDeath/HandlePetRevival、監査ログ+deceased_at 同一tx・fail-closed) に
	// 一本化済み。ここに "status" を JSON で送っても未知フィールドとして無視される。
	BirthDate       *jsonDate `json:"birth_date"`
	Breed           *string   `json:"breed"`
	Color           *string   `json:"color"`
	BloodType       *string   `json:"blood_type"       binding:"omitempty,max=32"`
	MicrochipNumber *string   `json:"microchip_number" binding:"omitempty,max=64"`
	Weight          *float64  `json:"weight"`
	NeuteredDate    *jsonDate `json:"neutered_date"`
	AcquisitionType *string   `json:"acquisition_type"`
	DangerLevel     *string   `json:"danger_level"`
	Food            *string   `json:"food"`
	Environment     *string   `json:"environment"`
	Phone           *string   `json:"phone"`
	LastVisit       *jsonDate `json:"last_visit"`
	InsuranceID     **uint64  `json:"insurance_id"`
	Remarks         *string   `json:"remarks"`
}

func (r *updatePetRequest) toServiceInput() *service.UpdatePetInput {
	return &service.UpdatePetInput{
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
		Food:            r.Food,
		Environment:     r.Environment,
		Phone:           r.Phone,
		LastVisit:       jsonDatePtr(r.LastVisit),
		InsuranceID:     r.InsuranceID,
		Remarks:         r.Remarks,
	}
}
