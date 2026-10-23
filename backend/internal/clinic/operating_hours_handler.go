package clinic

// 共通医院時間の読取 HTTP（A03 §2.2/§2.3 の GET 3 本。write/preview/結果照会は別 slice）。
// これらの route が生成する応答（200/400/403/503。route 認可 middleware の拒否を含む）に Cache-Control: no-store を付ける
// （契約 §1・api.yaml）。401 は保護 group の JWT middleware が route より前に書くため、グローバルの
// SecurityHeaders の値（no-store, no-cache, must-revalidate）になる（api.yaml の 401 header との差は別途対応）。

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const operatingHoursUnavailableMessage = "共通医院時間を取得できません"

// OperatingHoursHandler owns the common clinic operating-hours read endpoints.
type OperatingHoursHandler struct {
	reader OperatingHoursReader
	// requirePermission は選択医院の grant を要求する（休診 2 route。旧 /clinic-holidays と同じ）。
	requirePermission PermissionMiddleware
	// requireAssignedClinicPermission は GET/HEAD で所属医院の grant への fallback を許す
	// （URL :clinic_id の route。line-reservation-settings と同じ。対象医院の認可は handler 内で行う）。
	requireAssignedClinicPermission PermissionMiddleware
}

// NewOperatingHoursHandler constructs the read-side HTTP adapter. requirePermission is the strict
// selected-clinic middleware; requireAssignedClinicPermission is auth's
// RequirePermissionAllowingAssignedClinicGrant for the clinic-path route.
func NewOperatingHoursHandler(
	reader OperatingHoursReader,
	requirePermission PermissionMiddleware,
	requireAssignedClinicPermission PermissionMiddleware,
) *OperatingHoursHandler {
	return &OperatingHoursHandler{
		reader:                          reader,
		requirePermission:               requirePermission,
		requireAssignedClinicPermission: requireAssignedClinicPermission,
	}
}

// noStoreMiddleware は後続の middleware／handler（権限拒否を含む）の応答に no-store を付ける。
// 先行する group 単位の middleware（認証の 401 など）の応答には効かない。
// 先頭に置くこと。グローバルの SecurityHeaders が付ける複合値は api.yaml の enum [no-store] と一致しないため上書きする。
func noStoreMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

// RegisterOperatingHoursRoutes registers the three read routes (contract §2.2/§2.3).
func (h *OperatingHoursHandler) RegisterOperatingHoursRoutes(rg *gin.RouterGroup) {
	// 対象医院は URL :clinic_id の所属と hospital_settings grant で認可する（line-reservation-settings と同じ経路）。
	// route middleware は所属医院 grant への fallback を許し、URL 医院の grant は handler 内で検査する。
	rg.GET("/clinics/:clinic_id/operating-hours",
		noStoreMiddleware(),
		h.requireAssignedClinicPermission(string(model.ResourceHospitalSettings), "view"),
		h.GetClinicOperatingHours)
	// 月内の共通休診。医院は認証 context の選択医院（旧 /clinic-holidays と同じ）。
	rg.GET("/clinic-holidays/common",
		noStoreMiddleware(),
		h.requirePermission(string(model.ResourceShifts), "view"),
		h.ListCommonHolidays)
	// closing 経路は closing-settings:view に加え、旧 closing 休日 read が共有 handler で課している
	// shifts:view（選択医院）を ListCommonHolidays 内で同じ形で課す。
	rg.GET("/closing-settings/holidays/common",
		noStoreMiddleware(),
		h.requirePermission(string(model.ResourceClosingSettings), "view"),
		h.ListCommonHolidays)
}

// GetClinicOperatingHours godoc
// GET /v1/clinics/:clinic_id/operating-hours
func (h *OperatingHoursHandler) GetClinicOperatingHours(c *gin.Context) {
	clinicID, ok := canonicalClinicIDParam(c)
	if !ok {
		return
	}
	if !httpapi.AuthorizeClinicIDsForPermission(c, []uint64{clinicID}, string(model.ResourceHospitalSettings), "view") {
		return
	}
	current, err := h.reader.Get(c.Request.Context(), clinicID)
	if err != nil {
		respondOperatingHoursUnavailable(c, err)
		return
	}
	body, err := NewClinicOperatingHoursResponse(clinicID, current)
	if err != nil {
		respondOperatingHoursUnavailable(c, err)
		return
	}
	c.JSON(http.StatusOK, body)
}

// ListCommonHolidays godoc
// GET /v1/clinic-holidays/common?year_month=YYYY-MM
// GET /v1/closing-settings/holidays/common?year_month=YYYY-MM
// 旧 ListClinicHolidays と同じく、選択医院の shifts:view を共有 handler 内で検査する。
func (h *OperatingHoursHandler) ListCommonHolidays(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceShifts), "view") {
		return
	}
	if clinicID == 0 || clinicID > math.MaxInt64 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("invalid clinic context"))
		return
	}
	month, ok := yearMonthQuery(c)
	if !ok {
		return
	}
	current, err := h.reader.Get(c.Request.Context(), clinicID)
	if err != nil {
		respondOperatingHoursUnavailable(c, err)
		return
	}
	body, err := NewClinicCommonHolidayMonthResponse(clinicID, month, current)
	if err != nil {
		respondOperatingHoursUnavailable(c, err)
		return
	}
	c.JSON(http.StatusOK, body)
}

// canonicalClinicIDParam parses :clinic_id as a canonical decimal 1..9223372036854775807.
func canonicalClinicIDParam(c *gin.Context) (uint64, bool) {
	raw := c.Param("clinic_id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || id > math.MaxInt64 || strconv.FormatUint(id, 10) != raw {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("clinic_id は 1 以上の正規十進数で指定してください"))
		return 0, false
	}
	return id, true
}

// yearMonthQuery requires exactly one strict year_month query value.
func yearMonthQuery(c *gin.Context) (YearMonth, bool) {
	values := c.Request.URL.Query()["year_month"]
	if len(values) != 1 {
		httpapi.RespondError(c, apperrors.WrapInvalidInput("year_month は YYYY-MM 形式で1つだけ指定してください"))
		return YearMonth{}, false
	}
	month, err := ParseYearMonth(values[0])
	if err != nil {
		httpapi.RespondError(c, err)
		return YearMonth{}, false
	}
	return month, true
}

// respondOperatingHoursUnavailable は読取失敗・保存データ破損・応答構築不能をすべて 503 にする。
// 未設定や空成功へ変換しない。原因は Gin error として記録し、応答本文へは出さない。
func respondOperatingHoursUnavailable(c *gin.Context, cause error) {
	if cause != nil && !errors.Is(cause, ErrOperatingHoursUnavailable) && !errors.Is(cause, ErrOperatingHoursStoredDataCorrupt) {
		cause = errors.Join(ErrOperatingHoursUnavailable, cause)
	}
	_ = c.Error(cause)
	httpapi.RespondError(c, apperrors.WrapServiceUnavailable(operatingHoursUnavailableMessage))
}
