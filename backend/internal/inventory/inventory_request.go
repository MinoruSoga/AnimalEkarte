package inventory

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// errInventoryLastRestockedRetired は廃止された last_restocked の書込を拒否する（400）。
// inventory_items.last_restocked は 001_init.sql L10885 で DROP 済みで、黙って捨てると保存済みと思い込ませる。
// null は未送信と同じ扱いにする。
func errInventoryLastRestockedRetired() error {
	return apperrors.WrapInvalidInput("last_restocked は廃止されました。在庫の最終入荷日は保存されません")
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
	if r.LastRestocked != nil {
		return nil, errInventoryLastRestockedRetired()
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
	if r.LastRestocked != nil {
		return nil, errInventoryLastRestockedRetired()
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
	}, nil
}
