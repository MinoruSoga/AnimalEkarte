package medicalrecord

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// exactNumericCanonical renders a stored numeric as a canonical decimal string
// (Int * 10^Exp). Stored values originate from canonical wire input, so this
// only reformats digits without rounding.
func exactNumericCanonical(n model.ExactNumeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	digits := n.Int.String()
	neg := strings.HasPrefix(digits, "-")
	if neg {
		digits = digits[1:]
	}
	exp := int(n.Exp)
	var s string
	switch {
	case exp >= 0:
		s = digits + strings.Repeat("0", exp)
	default:
		p := len(digits) + exp
		if p > 0 {
			s = digits[:p] + "." + digits[p:]
		} else {
			s = "0." + strings.Repeat("0", -p) + digits
		}
	}
	if neg {
		s = "-" + s
	}
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-" || s == "-0" {
		return "0"
	}
	return s
}

// treatmentPackageItemResponse renders the typed ref branch:
// {"kind":"<kind>","<fk>":"<id>"} or {"kind":"manual_other","manual":{...}}.
type treatmentPackageItemResponse struct {
	Kind    string
	TypedID uint64
	Manual  json.RawMessage
}

func (r treatmentPackageItemResponse) MarshalJSON() ([]byte, error) {
	if r.Kind == model.TreatmentPackageItemKindManualOther {
		return json.Marshal(map[string]any{
			"kind":   r.Kind,
			"manual": r.Manual,
		})
	}
	field, ok := packageItemKindFKField[r.Kind]
	if !ok {
		return nil, errors.New("unknown package item kind")
	}
	return json.Marshal(map[string]any{
		"kind": r.Kind,
		field:  formatUint64String(r.TypedID),
	})
}

func formatUint64String(v uint64) string {
	return big.NewInt(0).SetUint64(v).String()
}

type treatmentPackageComponentResponse struct {
	ComponentKey string                        `json:"component_key"`
	Item         treatmentPackageItemResponse  `json:"item"`
	Quantity     string                        `json:"quantity"`
}

type treatmentPackageResponse struct {
	ID          string                              `json:"id"`
	ClinicID    string                              `json:"clinic_id"`
	Version     string                              `json:"version"`
	Name        string                              `json:"name"`
	Description string                              `json:"description"`
	IsActive    bool                                `json:"is_active"`
	SortOrder   int                                 `json:"sort_order"`
	Components  []treatmentPackageComponentResponse `json:"components"`
	CreatedAt   string                              `json:"created_at"`
	UpdatedAt   string                              `json:"updated_at"`
}

type treatmentPackageListResponse struct {
	Items        []treatmentPackageResponse `json:"items"`
	OrderVersion string                     `json:"order_version"`
}

func packageItemToResponse(item *model.TreatmentPackageItem) treatmentPackageComponentResponse {
	out := treatmentPackageComponentResponse{
		ComponentKey: item.ComponentKey.String(),
		Quantity:     exactNumericCanonical(item.Quantity),
	}
	out.Item.Kind = item.ItemKind
	switch item.ItemKind {
	case model.TreatmentPackageItemKindConsultation:
		out.Item.TypedID = derefPackageFK(item.ConsultationID)
	case model.TreatmentPackageItemKindProcedure:
		out.Item.TypedID = derefPackageFK(item.ProcedureID)
	case model.TreatmentPackageItemKindMedicine:
		out.Item.TypedID = derefPackageFK(item.MedicineID)
	case model.TreatmentPackageItemKindExaminationType:
		out.Item.TypedID = derefPackageFK(item.ExamTypeID)
	case model.TreatmentPackageItemKindVaccine:
		out.Item.TypedID = derefPackageFK(item.VaccineID)
	case model.TreatmentPackageItemKindCheckupType:
		out.Item.TypedID = derefPackageFK(item.CheckupTypeID)
	case model.TreatmentPackageItemKindMerchandiseItem:
		out.Item.TypedID = derefPackageFK(item.MerchandiseItemID)
	case model.TreatmentPackageItemKindManualOther:
		if item.ManualDefinition != nil {
			out.Item.Manual = *item.ManualDefinition
		} else {
			out.Item.Manual = json.RawMessage("null")
		}
	}
	return out
}

func derefPackageFK(id *uint64) uint64 {
	if id == nil {
		return 0
	}
	return *id
}

// TreatmentPackageView は package 定義と live 構成品の read 形。
type TreatmentPackageView struct {
	Package model.TreatmentPackage
	Items   []model.TreatmentPackageItem
}

func toTreatmentPackageResponse(v *TreatmentPackageView) treatmentPackageResponse {
	components := make([]treatmentPackageComponentResponse, 0, len(v.Items))
	for i := range v.Items {
		components = append(components, packageItemToResponse(&v.Items[i]))
	}
	return treatmentPackageResponse{
		ID:          formatUint64String(v.Package.ID),
		ClinicID:    formatUint64String(v.Package.ClinicID),
		Version:     formatInt64String(v.Package.Version),
		Name:        v.Package.Name,
		Description: v.Package.Description,
		IsActive:    v.Package.IsActive,
		SortOrder:   v.Package.SortOrder,
		Components:  components,
		CreatedAt:   httpapi.LocalTimeRFC3339(v.Package.CreatedAt),
		UpdatedAt:   httpapi.LocalTimeRFC3339(v.Package.UpdatedAt),
	}
}

func formatInt64String(v int64) string {
	return big.NewInt(v).String()
}

// respondTreatmentPackageError は TreatmentPackageError 契約の flat envelope
// （error＋code）を返す。SQL・他医院名・現在価格を露出しない。
func respondTreatmentPackageError(c *gin.Context, err error) {
	status, message, _ := httpapi.ResolveErrorResponse(err)
	if err != nil && status >= http.StatusInternalServerError {
		_ = c.Error(err)
	}
	c.JSON(status, gin.H{
		"error": message,
		"code":  treatmentPackageErrorCode(status),
	})
}

func treatmentPackageErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid-input"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "not-authorized"
	case http.StatusNotFound:
		return "scope-not-found"
	case http.StatusConflict:
		return "version-conflict"
	case http.StatusUnprocessableEntity:
		return "validation"
	case http.StatusServiceUnavailable:
		return "temporarily-unavailable"
	default:
		return "internal-error"
	}
}
