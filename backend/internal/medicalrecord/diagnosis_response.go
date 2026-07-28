package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// DiagnosisTypeResponse and DiagnosisNameResponse are exported (unlike this slice's other
// response DTOs) because internal/handler/clinical_plan_response.go — ClinicalPlan is
// boundary map §3.7 sub-batch ④, high-risk finalize-lock territory, out of this batch's
// scope and staying in internal/handler pending BE9-2D — embeds these exact response shapes
// in its own DiagnosisResponse/Diagnosis2Response fields. Exporting here (rather than
// duplicating the JSON shape in internal/handler) avoids two independently-maintained copies
// of the same API contract silently drifting apart.
type DiagnosisTypeResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToDiagnosisTypeResponse converts a model.DiagnosisType to its response shape.
func ToDiagnosisTypeResponse(c *model.DiagnosisType) DiagnosisTypeResponse {
	return DiagnosisTypeResponse{
		ID:          c.ID,
		ClinicID:    c.ClinicID,
		Name:        c.Name,
		IsActive:    c.IsActive,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		CreatedAt:   httpapi.LocalTime(c.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(c.UpdatedAt),
	}
}

// DiagnosisNameResponse — see DiagnosisTypeResponse's doc comment for why this is exported.
type DiagnosisNameResponse struct {
	ID              uint64    `json:"id"`
	ClinicID        uint64    `json:"clinic_id"`
	Name            string    `json:"name"`
	IsActive        bool      `json:"is_active"`
	Description     string    `json:"description"`
	DiagnosisTypeID uint64    `json:"diagnosis_type_id"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToDiagnosisNameResponse converts a model.DiagnosisName to its response shape.
func ToDiagnosisNameResponse(n *model.DiagnosisName) DiagnosisNameResponse {
	return DiagnosisNameResponse{
		ID:              n.ID,
		ClinicID:        n.ClinicID,
		Name:            n.Name,
		IsActive:        n.IsActive,
		Description:     n.Description,
		DiagnosisTypeID: n.DiagnosisTypeID,
		SortOrder:       n.SortOrder,
		CreatedAt:       httpapi.LocalTime(n.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(n.UpdatedAt),
	}
}
