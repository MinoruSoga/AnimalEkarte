package owner

import (
	"fmt"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// SummaryResponse is the owner summary embedded by pet and other domain
// responses.
// BUG-374 同様: フロントの generated/models.Owner.name を期待するため json:"name" に統一。
type SummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// PetAnimalSpeciesNested is the species embed on owner detail pets.
type PetAnimalSpeciesNested struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	SortOrder *int   `json:"sort_order"`
}

// PetInsuranceNested is the insurance embed on owner detail pets.
type PetInsuranceNested struct {
	ID           uint64  `json:"id"`
	Name         string  `json:"name"`
	CoverageRate float64 `json:"coverage_rate"`
	ContactPhone string  `json:"contact_phone"`
}

// PetInOwnerResponse is a pet row nested under OwnerResponse.
// EMR-491: 名前・動物種・性別などの基本情報は pet の現行 basic revision 由来で、nullable な列は JSON null で返す
// （空文字への置換や旧 pets 列へのフォールバックはしない）。last_visit は pets.last_visit の廃止後に取得元が無く、
// 導出もしないため常に null。
type PetInOwnerResponse struct {
	ID              uint64     `json:"id"`
	OwnerID         uint64     `json:"owner_id"`
	AnimalSpeciesID uint64     `json:"animal_species_id"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name"`
	PetNameKana     *string    `json:"pet_name_kana"`
	Gender          *string    `json:"gender"`
	Status          *string    `json:"status"`
	BirthDate       *time.Time `json:"birth_date"`
	Breed           *string    `json:"breed"`
	Color           *string    `json:"color"`
	BloodType       *string    `json:"blood_type"`
	MicrochipNumber *string    `json:"microchip_number"`
	DangerLevel     *string    `json:"danger_level"`
	NeuteredDate    *time.Time `json:"neutered_date"`
	NeuteredStatus  string     `json:"neutered_status"`
	AcquisitionType *string    `json:"acquisition_type"`
	Food            *string    `json:"food"`
	Environment     *string    `json:"environment"`
	LastVisit       *time.Time `json:"last_visit"`
	InsuranceID     *uint64    `json:"insurance_id,omitempty"`
	Remarks         *string    `json:"remarks"`
	// PetVersion / PetBasicRevisionID は pet の共通 subject の版と現行 basic revision の ID（基本情報の更新時に照合する元）。
	PetVersion         int64  `json:"pet_version"`
	PetBasicRevisionID uint64 `json:"pet_basic_revision_id"`
	// Weight は現行 revision の registration_weight（登録時の体重）。小数の桁を保ったまま JSON 数値で返す。
	Weight *model.ExactNumeric `json:"weight"`
	// DeceasedReason は含めない（セキュリティレビュー指摘: この構造体は未curationの
	// LIFF LinkLiffAccount 経路でも再利用されるため、スタッフ用の死因自由記述を飼主向け
	// レスポンスに載せてしまう。UI側でも死亡バナー表示に必要なのは DeceasedAt のみで
	// DeceasedReason の読み取り消費者は存在しない — 意図的に未追加）。
	DeceasedAt *time.Time `json:"deceased_at"`
	// EMR-174: 名前の由来 / 出逢いのストーリーは任意記録。
	// （deceased_reason と違い飼主向け公開に機微情報を含まないため owner 経路でも返す）。
	NameOrigin    *string                 `json:"name_origin"`
	MeetingStory  *string                 `json:"meeting_story"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	AnimalSpecies *PetAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance     *PetInsuranceNested     `json:"insurance,omitempty"`
}

// OwnerResponse is the owner detail/list HTTP response DTO (domain-owned wire).
// Field names use owner_name / owner_name_kana — not model.Owner's name / name_kana.
// EMR-491: 氏名・住所・連絡先は owner の現行 basic revision 由来で、nullable な列は JSON null で返す。
type OwnerResponse struct {
	ID uint64 `json:"id"`
	// ClinicID は所属医院 (#86 拠点横断一覧で医院名表示に使用)
	ClinicID               uint64               `json:"clinic_id"`
	OwnerVersion           int64                `json:"owner_version"`
	OwnerBasicRevisionID   uint64               `json:"owner_basic_revision_id"`
	OwnerName              string               `json:"owner_name"`
	OwnerNameKana          *string              `json:"owner_name_kana"`
	BirthDate              *time.Time           `json:"birth_date"`
	Company                *string              `json:"company"`
	PostalCode             *string              `json:"postal_code"`
	Address1               *string              `json:"address1"`
	Address2               *string              `json:"address2"`
	HomePostalCode         *string              `json:"home_postal_code"`
	HomeAddress1           *string              `json:"home_address1"`
	HomeAddress2           *string              `json:"home_address2"`
	Phone                  *string              `json:"phone"`
	CompanyPhone           *string              `json:"company_phone"`
	Email                  *string              `json:"email"`
	Remarks                string               `json:"remarks"`
	IsDangerous            bool                 `json:"is_dangerous"`
	DiscountRate           float64              `json:"discount_rate"`
	MembershipType         string               `json:"membership_type"`
	LineIDConfirmedAt      *time.Time           `json:"line_id_confirmed_at,omitempty"`
	LineIDConfirmedBy      *uint64              `json:"line_id_confirmed_by,omitempty"`
	DeliveryExcluded       bool                 `json:"delivery_excluded"`
	DeliveryExcludedReason *string              `json:"delivery_excluded_reason,omitempty"`
	DeliveryCaution        bool                 `json:"delivery_caution"`
	DeliveryCautionReason  *string              `json:"delivery_caution_reason,omitempty"`
	IsTransferred          bool                 `json:"is_transferred"`
	TransferAt             *time.Time           `json:"transfer_at,omitempty"`
	DMPreference           *bool                `json:"dm_preference,omitempty"`
	Pets                   []PetInOwnerResponse `json:"pets"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

// enumPtr は名前付き文字列型の nullable 値を *string へ写す。nil は nil のまま返し、空文字へは置換しない。
func enumPtr[T ~string](value *T) *string {
	if value == nil {
		return nil
	}
	text := string(*value)
	return &text
}

// toPetInOwnerResponse は hydrate 済みの pet を応答 DTO へ写す。基本情報が未解決のときは
// 空値で返さず unresolved（422）の error を返す（旧 pets 列へのフォールバックはしない）。
func toPetInOwnerResponse(p *model.Pet) (PetInOwnerResponse, error) {
	basic, err := commonbasic.PetOf(p)
	if err != nil {
		return PetInOwnerResponse{}, err
	}
	if p.BasicSubject == nil {
		return PetInOwnerResponse{}, apperrors.WrapCommonBasicUnresolved(
			fmt.Sprintf("pet %d/%d: basic subject is not hydrated", p.ClinicID, p.ID))
	}
	var weight *model.ExactNumeric
	if basic.RegistrationWeight != nil {
		copied := *basic.RegistrationWeight
		weight = &copied
	}
	resp := PetInOwnerResponse{
		ID:              p.ID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: basic.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            basic.Name,
		PetNameKana:     basic.NameKana,
		Gender:          enumPtr(basic.Gender),
		Status:          enumPtr(basic.Status),
		BirthDate:       httpapi.LocalTimePtr(basic.BirthDate),
		Breed:           basic.Breed,
		Color:           basic.Color,
		BloodType:       basic.BloodType,
		MicrochipNumber: basic.MicrochipNumber,
		DangerLevel:     enumPtr(basic.DangerLevel),
		Weight:          weight,
		NeuteredDate:    httpapi.LocalTimePtr(basic.NeuteredDate),
		NeuteredStatus:  string(basic.NeuteredStatus),
		AcquisitionType: enumPtr(basic.AcquisitionType),
		Food:            basic.Food,
		Environment:     basic.Environment,
		InsuranceID:     p.InsuranceID,
		Remarks:         basic.Remarks,
		DeceasedAt:      httpapi.LocalTimePtr(basic.DeceasedAt),
		NameOrigin:      basic.NameOrigin,
		MeetingStory:    basic.MeetingStory,
		CreatedAt:       httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(p.UpdatedAt),
	}
	resp.PetVersion = p.BasicSubject.Version
	resp.PetBasicRevisionID = p.BasicSubject.CurrentRevisionID
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
			CoverageRate: float64(p.Insurance.CoverageRate),
			ContactPhone: p.Insurance.ContactPhone,
		}
	}
	return resp, nil
}

// toOwnerResponse は hydrate 済みの飼主（入れ子の pets を含む）を応答 DTO へ写す。
// 基本情報が 1 件でも未解決なら空値で返さず unresolved（422）の error を返す。
func toOwnerResponse(o *model.Owner) (OwnerResponse, error) {
	basic, err := commonbasic.OwnerOf(o)
	if err != nil {
		return OwnerResponse{}, err
	}
	if o.BasicSubject == nil {
		return OwnerResponse{}, apperrors.WrapCommonBasicUnresolved(
			fmt.Sprintf("owner %d/%d: basic subject is not hydrated", o.ClinicID, o.ID))
	}
	pets := make([]PetInOwnerResponse, 0, len(o.Pets))
	for i := range o.Pets {
		petResp, petErr := toPetInOwnerResponse(&o.Pets[i])
		if petErr != nil {
			return OwnerResponse{}, petErr
		}
		pets = append(pets, petResp)
	}
	// owners.line_id_confirmed_at/by は drop 済みで、置換先（owner_line_binding_states）の reader がこの branch に無い。
	// line_id_confirmed_at / line_id_confirmed_by は値を作らず未設定のまま返す（omitempty により応答から省略される）。
	return OwnerResponse{
		ID:                     o.ID,
		ClinicID:               o.ClinicID,
		OwnerVersion:           o.BasicSubject.Version,
		OwnerBasicRevisionID:   o.BasicSubject.CurrentRevisionID,
		OwnerName:              basic.Name,
		OwnerNameKana:          basic.NameKana,
		BirthDate:              httpapi.LocalTimePtr(basic.BirthDate),
		Company:                basic.Company,
		PostalCode:             basic.PostalCode,
		Address1:               basic.Address1,
		Address2:               basic.Address2,
		HomePostalCode:         basic.HomePostalCode,
		HomeAddress1:           basic.HomeAddress1,
		HomeAddress2:           basic.HomeAddress2,
		Phone:                  basic.Phone,
		CompanyPhone:           basic.CompanyPhone,
		Email:                  basic.Email,
		Remarks:                o.Remarks,
		IsDangerous:            o.IsDangerous,
		DiscountRate:           o.DiscountRate,
		MembershipType:         string(o.MembershipType),
		DeliveryExcluded:       o.DeliveryExcluded,
		DeliveryExcludedReason: o.DeliveryExcludedReason,
		DeliveryCaution:        o.DeliveryCaution,
		DeliveryCautionReason:  o.DeliveryCautionReason,
		IsTransferred:          o.IsTransferred,
		TransferAt:             httpapi.LocalTimePtr(o.TransferAt),
		DMPreference:           o.DMPreference,
		Pets:                   pets,
		CreatedAt:              httpapi.LocalTime(o.CreatedAt),
		UpdatedAt:              httpapi.LocalTime(o.UpdatedAt),
	}, nil
}
