package billing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingConfirmationHandler は BillingConfirmationService の HTTP handler。
type BillingConfirmationHandler struct {
	svc               BillingConfirmationService
	requirePermission PermissionMiddleware
}

// NewBillingConfirmationHandler は BillingConfirmationHandler を構築する。
func NewBillingConfirmationHandler(svc BillingConfirmationService, requirePermission PermissionMiddleware) *BillingConfirmationHandler {
	return &BillingConfirmationHandler{svc: svc, requirePermission: requirePermission}
}

// GetBillingConfirmation は指定カルテIDの会計医師確認を取得または初期化して返す
// GET /medical-records/:id/billing-confirmation
func (h *BillingConfirmationHandler) GetBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceAccounting), "view") {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	review, err := h.svc.GetOrCreate(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ConfirmBillingConfirmation は会計を医師確認済みにする
// POST /medical-records/:id/billing-confirmation/confirm
func (h *BillingConfirmationHandler) ConfirmBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req confirmBillingConfirmationRequest
	if err := bindBillingConfirmationJSON(c, &req, "memo"); err != nil {
		respondBillingConfirmationRequestError(c, err)
		return
	}
	normalizedReq, err := req.normalizeAndValidate()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	review, err := h.svc.Confirm(c.Request.Context(), clinicID, medicalRecordID, normalizedReq.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

// ReturnBillingConfirmation は会計を差し戻す
// POST /medical-records/:id/billing-confirmation/return
func (h *BillingConfirmationHandler) ReturnBillingConfirmation(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}

	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	var req returnBillingConfirmationRequest
	if err := bindBillingConfirmationJSON(c, &req, "return_reason", "memo"); err != nil {
		respondBillingConfirmationRequestError(c, err)
		return
	}
	normalizedReq, err := req.normalizeAndValidate()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	review, err := h.svc.Return(c.Request.Context(), clinicID, medicalRecordID, normalizedReq.toServiceInput(staffID))
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toBillingConfirmationResponse(review))
}

var errBillingConfirmationUnsupportedMediaType = errors.New(
	billingConfirmationUnsupportedMediaTypeMessage,
)

func bindBillingConfirmationJSON(c *gin.Context, destination any, allowedFields ...string) error {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errBillingConfirmationUnsupportedMediaType
	}

	if c.Request.ContentLength > billingConfirmationJSONBodyMaxBytes {
		return apperrors.WrapPayloadTooLarge(billingConfirmationBodyTooLargeMessage)
	}

	boundedBody := http.MaxBytesReader(c.Writer, c.Request.Body, billingConfirmationJSONBodyMaxBytes)
	defer func() {
		_ = boundedBody.Close()
	}()

	body, err := io.ReadAll(boundedBody)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return apperrors.WrapPayloadTooLarge(billingConfirmationBodyTooLargeMessage)
		}
		return apperrors.WrapInvalidInput(billingConfirmationInvalidBodyMessage)
	}

	trimmedBody := bytes.TrimSpace(body)
	if len(trimmedBody) == 0 || trimmedBody[0] != '{' {
		return apperrors.WrapInvalidInput(billingConfirmationInvalidBodyMessage)
	}
	if !hasOnlyExactStringFields(body, allowedFields) {
		return apperrors.WrapInvalidInput(billingConfirmationInvalidBodyMessage)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return apperrors.WrapInvalidInput(billingConfirmationInvalidBodyMessage)
	}

	var trailingValue json.RawMessage
	if err := decoder.Decode(&trailingValue); !errors.Is(err, io.EOF) {
		return apperrors.WrapInvalidInput(billingConfirmationInvalidBodyMessage)
	}
	return nil
}

func hasOnlyExactStringFields(body []byte, allowedFields []string) bool {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		return false
	}

	allowed := make(map[string]struct{}, len(allowedFields))
	for _, field := range allowedFields {
		allowed[field] = struct{}{}
	}
	for field, rawValue := range fields {
		if _, ok := allowed[field]; !ok {
			return false
		}
		trimmedValue := bytes.TrimSpace(rawValue)
		if len(trimmedValue) == 0 || trimmedValue[0] != '"' {
			return false
		}
	}
	return true
}

func respondBillingConfirmationRequestError(c *gin.Context, err error) {
	if errors.Is(err, errBillingConfirmationUnsupportedMediaType) {
		c.JSON(
			http.StatusUnsupportedMediaType,
			gin.H{"error": billingConfirmationUnsupportedMediaTypeMessage},
		)
		return
	}
	httpapi.RespondError(c, err)
}
