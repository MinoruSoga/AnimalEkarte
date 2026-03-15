package handler

import "github.com/animal-ekarte/backend/internal/model"

// cageSummaryResponse はネストされたレスポンスで使用するケージの要約型
type cageSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toCageSummary は *model.Cage を *cageSummaryResponse に変換する。nilの場合はnilを返す。
func toCageSummary(c *model.Cage) *cageSummaryResponse {
	if c == nil {
		return nil
	}
	return &cageSummaryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}

