package httpapi

// PaginatedResponse is the common shape for a paginated list response (moved from
// internal/handler's PaginatedResponse — BE9-2C, query_helpers.go target:httpapi
// extraction; see this package's doc comment).
type PaginatedResponse[T any] struct {
	Data  T     `json:"data"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// NewPaginatedResponse builds a PaginatedResponse with type inference.
func NewPaginatedResponse[T any](data T, total int64, page, limit int) PaginatedResponse[T] {
	return PaginatedResponse[T]{Data: data, Total: total, Page: page, Limit: limit}
}
