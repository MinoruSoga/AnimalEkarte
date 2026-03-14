package handler

type createConsultationRequest struct {
	Name          string   `json:"name"           binding:"required"`
	Price         *float64 `json:"price"`
	IsActive      bool     `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	SortOrder     int      `json:"sort_order"`
}

type updateConsultationRequest struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	SortOrder     int      `json:"sort_order"`
}
