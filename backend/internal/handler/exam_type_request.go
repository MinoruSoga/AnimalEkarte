package handler

type createExaminationTypeRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

type updateExaminationTypeRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

type examTypeItemInput struct {
	Name            string `json:"name"             binding:"required"`
	InspectionValue string `json:"inspection_value"`
	NormalValue     string `json:"normal_value"`
	SortOrder       int    `json:"sort_order"`
}

type replaceExamTypeItemsRequest struct {
	Items []examTypeItemInput `json:"items" binding:"required"`
}
