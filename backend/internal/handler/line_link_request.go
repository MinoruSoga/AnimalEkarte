package handler

import "github.com/animal-ekarte/backend/internal/service"

// linkAccountRequest は LinkLiffAccount のリクエスト。
type linkAccountRequest struct {
	LinkToken   string `json:"link_token" binding:"required"`
	LineIDToken string `json:"line_id_token" binding:"required"`
	Force       bool   `json:"force"`
}

func (r linkAccountRequest) toServiceInput() service.LinkAccountInput {
	return service.LinkAccountInput{
		LinkToken:   r.LinkToken,
		LineIDToken: r.LineIDToken,
		Force:       r.Force,
	}
}
