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
type createInventoryRequest struct {
	Name          string  `json:"name"            binding:"required"`
	Category      string  `json:"category"        binding:"required,oneof=medicine consumable food other"`
	Quantity      int     `json:"quantity"`
	Unit          string  `json:"unit"            binding:"required"`
	MinStockLevel int     `json:"min_stock_level"`
	Location      string  `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      string  `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        string  `json:"status"          binding:"omitempty,oneof=sufficient low out_of_stock"`
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
		Status:        r.Status,
	}, nil
}

// updateInventoryRequest は在庫アイテム更新リクエスト。
type updateInventoryRequest struct {
	Name          *string `json:"name"`
	Category      *string `json:"category"        binding:"omitempty,oneof=medicine consumable food other"`
	Quantity      *int    `json:"quantity"`
	Unit          *string `json:"unit"`
	MinStockLevel *int    `json:"min_stock_level"`
	Location      *string `json:"location"`
	ExpiryDate    *string `json:"expiry_date"`
	Supplier      *string `json:"supplier"`
	LastRestocked *string `json:"last_restocked"`
	Status        *string `json:"status"          binding:"omitempty,oneof=sufficient low out_of_stock"`
}

func (r *updateInventoryRequest) toServiceInput() (*UpdateInventoryInput, error) {
	var category *model.InventoryCategory
	if r.Category != nil {
		cat := model.InventoryCategory(*r.Category)
		category = &cat
	}
	var status *model.InventoryStatus
	if r.Status != nil {
		s := model.InventoryStatus(*r.Status)
		status = &s
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
		Status:        status,
	}, nil
}
