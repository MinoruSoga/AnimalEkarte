package pet

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
)

type PetAnimalSpeciesNested struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type PetInsuranceNested struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	CoverageRate int    `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
}

// PetOwnerNested はペット行に埋め込む飼主サマリ（#266: pets 一覧のペット行粒度化）。
// petListResponse / PetResponse の両方で共有する（detail 側は追加フィールドを無視すれば足りるため
// 型を分けない）。OwnerNumber は独立した採番カラムではなく Owner.ID のエイリアス
// （FE 既存実装が ownerNumber: owner.id として扱っていた表示上の呼称に合わせる）。
type PetOwnerNested struct {
	ID          uint64 `json:"id"`
	OwnerNumber uint64 `json:"owner_number"`
	Name        string `json:"name"`
	NameKana    string `json:"name_kana"`
	Phone       string `json:"phone"`
	IsDangerous bool   `json:"is_dangerous"`
}

func toPetOwnerNested(o *model.Owner) *PetOwnerNested {
	if o == nil {
		return nil
	}
	return &PetOwnerNested{
		ID:          o.ID,
		OwnerNumber: o.ID,
		Name:        o.Name,
		NameKana:    o.NameKana,
		Phone:       o.Phone,
		IsDangerous: o.IsDangerous,
	}
}

type PetResponse struct {
	ID              uint64     `json:"id"`
	Version         int        `json:"version"`
	ClinicID        uint64     `json:"clinic_id"`
	OwnerID         uint64     `json:"owner_id"`
	AnimalSpeciesID uint64     `json:"animal_species_id"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	Status          string     `json:"status"`
	BirthDate       *time.Time `json:"birth_date,omitempty"`
	Breed           string     `json:"breed"`
	Color           string     `json:"color"`
	BloodType       *string    `json:"blood_type,omitempty"`
	MicrochipNumber *string    `json:"microchip_number,omitempty"`
	Weight          *float64   `json:"weight,omitempty"`
	NeuteredDate    *time.Time `json:"neutered_date,omitempty"`
	AcquisitionType *string    `json:"acquisition_type,omitempty"`
	DangerLevel     string     `json:"danger_level"`
	DangerReason    *string    `json:"danger_reason,omitempty"`
	Food            string     `json:"food"`
	Environment     string     `json:"environment"`
	Phone           string     `json:"phone"`
	LastVisit       *time.Time `json:"last_visit,omitempty"`
	InsuranceID     *uint64    `json:"insurance_id,omitempty"`
	Remarks         string     `json:"remarks"`
	// DeceasedReason は staff 向け GET /pets/{id} (本 DTO) のみに載せる（BUG-003）。
	// owner.PetInOwnerResponse / LIFF 向け DTO には載せない（飼主経路への死因漏洩防止）。
	// omitempty: 生存ペットや未記録時は JSON から物理的に欠落させる。
	DeceasedReason *string                 `json:"deceased_reason,omitempty"`
	DeceasedAt     *time.Time              `json:"deceased_at,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Owner          *PetOwnerNested         `json:"owner,omitempty"`
	AnimalSpecies  *PetAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance      *PetInsuranceNested     `json:"insurance,omitempty"`
}

// petFirstVisitResponse は #158 飼主レポートのペット初診日（最古カルテ date 由来）。
// FirstVisitDate は派生値で、カルテが無い場合は null（捏造しない）。
type petFirstVisitResponse struct {
	FirstVisitDate *time.Time `json:"first_visit_date"`
}

func toPetFirstVisitResponse(date *time.Time) petFirstVisitResponse {
	// MedicalRecord.date は medical_record 詳細経路で localTime(r.Date) として datetime 配信される。
	// 同じ date 値の派生である初診日も localTimePtr で time.Local へ変換し、経路間の tz 表現割れ
	// (`…Z` vs `…+09:00`) を防ぐ。nil (カルテ無し) は素通しで null を維持する。
	return petFirstVisitResponse{FirstVisitDate: httpapi.LocalTimePtr(date)}
}

// petListResponse はリスト表示に必要な最小限フィールドのみ返す（GET /v1/pets 専用）
type petListResponse struct {
	ID uint64 `json:"id"`
	// ClinicID: #266/#86 拠点横断一覧で FE (OwnersList.tsx) が「別医院の行は編集・削除を抑止」
	// 判定に使う。PetResponse(詳細) には既にあるが petListResponse は最小限フィールド構成のため
	// 欠けていた（#266 pets 一覧のペット行粒度化で FE がこの一覧に依存するようになり露見）。
	ClinicID        uint64                  `json:"clinic_id"`
	OwnerID         uint64                  `json:"owner_id"`
	AnimalSpeciesID uint64                  `json:"animal_species_id"`
	PetNumber       string                  `json:"pet_number"`
	Name            string                  `json:"name"`
	PetNameKana     string                  `json:"pet_name_kana"`
	Gender          string                  `json:"gender"`
	Status          string                  `json:"status"`
	BirthDate       *time.Time              `json:"birth_date,omitempty"`
	Breed           string                  `json:"breed"`
	Color           string                  `json:"color"`
	BloodType       *string                 `json:"blood_type,omitempty"`
	MicrochipNumber *string                 `json:"microchip_number,omitempty"`
	Weight          *float64                `json:"weight,omitempty"`
	NeuteredDate    *time.Time              `json:"neutered_date,omitempty"`
	AcquisitionType *string                 `json:"acquisition_type,omitempty"`
	DangerLevel     string                  `json:"danger_level"`
	DangerReason    *string                 `json:"danger_reason,omitempty"`
	Food            string                  `json:"food"`
	Environment     string                  `json:"environment"`
	LastVisit       *time.Time              `json:"last_visit,omitempty"`
	InsuranceID     *uint64                 `json:"insurance_id,omitempty"`
	Remarks         string                  `json:"remarks"`
	Owner           *PetOwnerNested         `json:"owner,omitempty"`
	AnimalSpecies   *PetAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance       *PetInsuranceNested     `json:"insurance,omitempty"`
}

func toPetListResponse(p *model.Pet) petListResponse {
	var acquisitionType *string
	if p.AcquisitionType != nil {
		s := string(*p.AcquisitionType)
		acquisitionType = &s
	}
	resp := petListResponse{
		ID:              p.ID,
		ClinicID:        p.ClinicID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		// 日付フィールドは canonical 規約に従い localTimePtr で time.Local へ変換する
		// (格納は date.go の ParseInLocation(time.Local) によるローカル暦日。
		// owner 経路と byte 一致させ、pgx の深夜 UTC 表現の漏れを防ぐ)。
		BirthDate:       httpapi.LocalTimePtr(p.BirthDate),
		Breed:           p.Breed,
		Color:           p.Color,
		BloodType:       p.BloodType,
		MicrochipNumber: p.MicrochipNumber,
		Weight:          p.Weight,
		NeuteredDate:    httpapi.LocalTimePtr(p.NeuteredDate),
		AcquisitionType: acquisitionType,
		DangerLevel:     string(p.DangerLevel),
		DangerReason:    p.DangerReason,
		Food:            p.Food,
		Environment:     p.Environment,
		LastVisit:       httpapi.LocalTimePtr(p.LastVisit),
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
	}
	resp.Owner = toPetOwnerNested(p.Owner)
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &PetAnimalSpeciesNested{
			ID:        p.AnimalSpecies.ID,
			Name:      p.AnimalSpecies.Name,
			SortOrder: p.AnimalSpecies.SortOrder,
		}
	}
	if p.Insurance != nil {
		resp.Insurance = &PetInsuranceNested{
			ID:           p.Insurance.ID,
			Name:         p.Insurance.Name,
			CoverageRate: p.Insurance.CoverageRate,
			ContactPhone: p.Insurance.ContactPhone,
		}
	}
	return resp
}

// SummaryResponse is the nested pet carrier retained for trimming and other
// compatibility responses until those domains own their response DTOs.
type SummaryResponse struct {
	ID        uint64   `json:"id"`
	Name      string   `json:"name"`
	PetNumber string   `json:"pet_number"`
	Weight    *float64 `json:"weight,omitempty"`
	// Status は死亡ペット判定（入院一覧/詳細の petIsDeceased）向け。alive/deceased。
	Status string `json:"status,omitempty"`
	// Breed は犬種等（#231: トリミング一覧の犬種列向け。空文字は未設定を示す）。
	Breed         string                        `json:"breed,omitempty"`
	AnimalSpecies *AnimalSpeciesSummaryResponse `json:"animal_species,omitempty"`
	// Owner は飼主名を必要とする一覧（トリミング/予約等）向け。Pet.Owner を Preload した場合のみ埋まる。
	// owner.SummaryResponse は他 package 参照のため tygo が any に潰す。wire は id+name のみなので
	// tstype で明示する（trimming の tstype:"-" と同系統。JSON タグ・runtime 形状は不変）。
	Owner *owner.SummaryResponse `json:"owner,omitempty" tstype:"{ id: number; name: string }"`
}

// AnimalSpeciesSummaryResponse is the species summary nested in SummaryResponse.
type AnimalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func toPetResponse(p *model.Pet) PetResponse {
	var acquisitionType *string
	if p.AcquisitionType != nil {
		s := string(*p.AcquisitionType)
		acquisitionType = &s
	}
	resp := PetResponse{
		ID:              p.ID,
		Version:         p.Version,
		ClinicID:        p.ClinicID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		// 日付フィールドは canonical 規約に従い localTimePtr で time.Local へ変換する
		// (CreatedAt/UpdatedAt と同様に同関数内でローカル化を統一)。
		BirthDate:       httpapi.LocalTimePtr(p.BirthDate),
		Breed:           p.Breed,
		Color:           p.Color,
		BloodType:       p.BloodType,
		MicrochipNumber: p.MicrochipNumber,
		Weight:          p.Weight,
		NeuteredDate:    httpapi.LocalTimePtr(p.NeuteredDate),
		AcquisitionType: acquisitionType,
		DangerLevel:     string(p.DangerLevel),
		DangerReason:    p.DangerReason,
		Food:            p.Food,
		Environment:     p.Environment,
		Phone:           p.Phone,
		LastVisit:       httpapi.LocalTimePtr(p.LastVisit),
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
		DeceasedReason:  p.DeceasedReason,
		DeceasedAt:      httpapi.LocalTimePtr(p.DeceasedAt),
		CreatedAt:       httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(p.UpdatedAt),
	}
	resp.Owner = toPetOwnerNested(p.Owner)
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &PetAnimalSpeciesNested{
			ID:        p.AnimalSpecies.ID,
			Name:      p.AnimalSpecies.Name,
			SortOrder: p.AnimalSpecies.SortOrder,
		}
	}
	if p.Insurance != nil {
		resp.Insurance = &PetInsuranceNested{
			ID:           p.Insurance.ID,
			Name:         p.Insurance.Name,
			CoverageRate: p.Insurance.CoverageRate,
			ContactPhone: p.Insurance.ContactPhone,
		}
	}
	return resp
}
