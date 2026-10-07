package medicalrecord

import (
	"net/url"
)

type listCageQuery struct {
	CageType string
}

func newListCageQuery(values url.Values) listCageQuery {
	return listCageQuery{CageType: values.Get("cage_type")}
}

func (q listCageQuery) toServiceFilter() *string {
	return optionalStringQueryFilter(q.CageType)
}

type createCageRequest struct {
	Name        string `json:"name"        binding:"required"`
	CageType    string `json:"cage_type"   binding:"required,oneof=icu dog cat general"`
	CageSize    string `json:"cage_size"   binding:"required,oneof=small medium large"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func (r *createCageRequest) toServiceInput() *CreateCageInput {
	return &CreateCageInput{
		Name:        r.Name,
		CageType:    r.CageType,
		CageSize:    r.CageSize,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

type updateCageRequest struct {
	Name        *string `json:"name"`
	CageType    *string `json:"cage_type"    binding:"omitempty,oneof=icu dog cat general"`
	CageSize    *string `json:"cage_size"    binding:"omitempty,oneof=small medium large"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateCageRequest) toServiceInput() *UpdateCageInput {
	return &UpdateCageInput{
		Name:        r.Name,
		CageType:    r.CageType,
		CageSize:    r.CageSize,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

// mergeCageRequest は POST /masters/cages/:id/merge の入力（EMR-306）。
// member_cage_ids は1件以上必須。重複・自己参照・未連結状態などのドメイン検証は
// サービス側が担う（handler は malformed/empty のみ 400 に落とす）。
type mergeCageRequest struct {
	MemberCageIDs []uint64 `json:"member_cage_ids" binding:"required,min=1"`
}
