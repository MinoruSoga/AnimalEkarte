package reservation

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// extractSelectedClinicGrant は選択医院の clinic_id を取り出し、その医院への
// view grant を要求する。医院固定 GET 用。書き込みは ExtractClinicID のまま。
func extractSelectedClinicGrant(c *gin.Context, resource string) (uint64, bool) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return 0, false
	}
	if !httpapi.RequireSelectedClinicGrant(c, resource, "view") {
		return 0, false
	}
	return clinicID, true
}
