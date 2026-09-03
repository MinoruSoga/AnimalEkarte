package inventory

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type listInventoryQuery struct {
	Category string
	Status   string
}

func newListInventoryQuery(values url.Values) listInventoryQuery {
	return listInventoryQuery{
		Category: values.Get("category"),
		Status:   values.Get("status"),
	}
}

type listInventoryFilters struct {
	Category *string
	Status   *string
}

func (q listInventoryQuery) toServiceFilters() listInventoryFilters {
	return listInventoryFilters{
		Category: httpapi.NilIfEmpty(q.Category),
		Status:   httpapi.NilIfEmpty(q.Status),
	}
}

// createInventoryRequest は在庫アイテム作成リクエスト。
// SD-4: status は request body に含めても無視（OpenAPI readOnly）。
// 公開 JSON の status は quantity/min_stock_level から導出する。
type createInventoryRequest struct {
	Name          string  `json:"name"            binding:"required,max=255"`
	Category      string  `json:"category"        binding:"required,oneof=medicine consumable food other"`
	Quantity      int     `json:"quantity"        binding:"min=0"` // BUG-466: 負数在庫の作成を拒否
	Unit          string  `json:"unit"            binding:"required,max=50"`
	MinStockLevel int     `json:"min_stock_level"`
	Location      string  `json:"location"        binding:"omitempty,max=255"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      string  `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
}

func (r *createInventoryRequest) toServiceInput() (*CreateInventoryInput, error) {
	expiryDate, err := httpapi.ParseDate(r.ExpiryDate)
	if err != nil {
		return nil, err
	}
	lastRestocked, err := httpapi.ParseDate(r.LastRestocked)
	if err != nil {
		return nil, err
	}
	return &CreateInventoryInput{
		Name:          r.Name,
		Category:      r.Category,
		Quantity:      r.Quantity,
		Unit:          r.Unit,
		MinStockLevel: r.MinStockLevel,
		Location:      r.Location,
		ExpiryDate:    expiryDate,
		Supplier:      r.Supplier,
		LastRestocked: lastRestocked,
	}, nil
}

// updateInventoryRequest は在庫アイテム更新リクエスト。
// SD-4: status は request body に含めても無視（OpenAPI readOnly）。
type updateInventoryRequest struct {
	Name          *string `json:"name" binding:"omitempty,max=255"`
	Category      *string `json:"category"        binding:"omitempty,oneof=medicine consumable food other"`
	Quantity      *int    `json:"quantity"        binding:"omitempty,min=0"` // BUG-466: 省略可・負数拒否
	Unit          *string `json:"unit" binding:"omitempty,max=50"`
	MinStockLevel *int    `json:"min_stock_level"`
	Location      *string `json:"location" binding:"omitempty,max=255"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      *string `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
}

func (r *updateInventoryRequest) toServiceInput() (*UpdateInventoryInput, error) {
	var category *model.InventoryCategory
	if r.Category != nil {
		cat := model.InventoryCategory(*r.Category)
		category = &cat
	}
	expiryDate, err := httpapi.ParseDate(r.ExpiryDate)
	if err != nil {
		return nil, err
	}
	lastRestocked, err := httpapi.ParseDate(r.LastRestocked)
	if err != nil {
		return nil, err
	}
	return &UpdateInventoryInput{
		Name:          r.Name,
		Category:      category,
		Quantity:      r.Quantity,
		Unit:          r.Unit,
		MinStockLevel: r.MinStockLevel,
		Location:      r.Location,
		ExpiryDate:    expiryDate,
		Supplier:      r.Supplier,
		LastRestocked: lastRestocked,
	}, nil
}
