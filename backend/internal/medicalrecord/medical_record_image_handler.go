package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordImageHandler serves the medical-record-image (診療画像) HTTP boundary nested under a
// medical record. It holds the MedicalRecordImageService, a medicalRecordGetter for the
// tenant-ownership guard (pre-move verifyMedicalRecordOwnership), and a fileUploader for the
// multipart upload endpoint (S3 or local storage).
type MedicalRecordImageHandler struct {
	service       MedicalRecordImageService
	medicalRecord medicalRecordGetter
	uploader      fileUploader
	quota         medicalRecordImageUploadQuotaStore
}

// NewMedicalRecordImageHandler initializes a MedicalRecordImageHandler.
// quota is required for uploads; nil fails closed. Optional trailing arg keeps
// route-snapshot call sites compiling during migration.
func NewMedicalRecordImageHandler(
	service MedicalRecordImageService,
	medicalRecord medicalRecordGetter,
	uploader fileUploader,
	quota ...medicalRecordImageUploadQuotaStore,
) *MedicalRecordImageHandler {
	var q medicalRecordImageUploadQuotaStore
	if len(quota) > 0 {
		q = quota[0]
	}
	return &MedicalRecordImageHandler{
		service:       service,
		medicalRecord: medicalRecord,
		uploader:      uploader,
		quota:         q,
	}
}

// verifyOwnership は clinicID + medicalRecordID を検証しテナント分離を保証する（pre-move
// internal/handler.Handler.verifyMedicalRecordOwnership のローカル移植）。検証済みの
// MedicalRecord と成否を返す。
func (h *MedicalRecordImageHandler) verifyOwnership(c *gin.Context, clinicID, medicalRecordID uint64) bool {
	_, err := h.medicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return false
	}
	return true
}

// ListMedicalRecordImages godoc
// GET /medical-records/:id/images
func (h *MedicalRecordImageHandler) ListMedicalRecordImages(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceMedicalRecords), "view") {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	images, err := h.service.List(c.Request.Context(), clinicID, medicalRecordID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// 署名は全件発行してから削除台帳と一括突合する。先に blocked 判定すると、
	// 判定後の署名発行中に削除 commit が割り込み retired 画像の署名 URL を返しうる
	// （TOCTOU）。署名後チェックで「発行済み URL が削除済みだった」ケースを検出する。
	items := make([]medicalRecordImageResponse, 0, len(images))
	imageIDs := make([]uint64, 0, len(images))
	for i := range images {
		item, signErr := h.signMedicalRecordImageResponse(c.Request.Context(), &images[i])
		if signErr != nil {
			httpapi.RespondError(c, apperrors.Wrap(signErr, "failed to sign medical record image url"))
			return
		}
		items = append(items, item)
		imageIDs = append(imageIDs, images[i].ID)
	}
	blocked, err := h.service.BlockedImageIDs(c.Request.Context(), clinicID, imageIDs)
	if err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to check medical record image deletion"))
		return
	}
	if len(blocked) > 0 {
		httpapi.RespondError(c, errMedicalRecordImageSigningRetired)
		return
	}
	c.JSON(http.StatusOK, items)
}

// DeleteMedicalRecordImage godoc
// DELETE /medical-records/:id/images/:imageId
func (h *MedicalRecordImageHandler) DeleteMedicalRecordImage(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	medicalRecordID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	imageID, ok := httpapi.ParseIDParam(c, "imageId")
	if !ok {
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, imageID, staffID, h.uploader); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadMedicalRecordImage godoc
// POST /medical-records/:id/images/upload
// Accepts multipart/form-data with field name "file".
func (h *MedicalRecordImageHandler) UploadMedicalRecordImage(c *gin.Context) {
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
	if !h.verifyOwnership(c, clinicID, medicalRecordID) {
		return
	}

	if c.Request.ContentLength > medicalRecordImageMaxRequestSize {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("medical record image upload request exceeds size limit"))
		return
	}

	// SEC-CS-F08-R1: acquire shared quota BEFORE FormFile / MaxBytesReader body work.
	if h.quota == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "upload quota unavailable"})
		return
	}
	declaredBytes := c.Request.ContentLength
	if declaredBytes <= 0 {
		// Chunked / unknown length: reserve per-file upper bound conservatively.
		declaredBytes = medicalRecordImageMaxUploadSize
	}
	release, err := h.quota.Acquire(c.Request.Context(), clinicID, staffID, declaredBytes)
	if err != nil {
		respondMedicalRecordImageUploadQuotaError(c, err)
		return
	}
	// Release must not inherit a canceled request context (lease would stick until stale TTL).
	defer release(context.WithoutCancel(c.Request.Context()))

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, medicalRecordImageMaxRequestSize)
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("medical record image upload request exceeds size limit"))
			return
		}
		httpapi.RespondError(c, apperrors.WrapInvalidInput("file field is required"))
		return
	}
	defer file.Close() //nolint:errcheck // multipart ファイルのクローズ失敗は復旧不可のため無視

	uploadMeta, err := newMedicalRecordImageUploadRequest(fileHeader).validate()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	// SEC-CS3-O6: 宣言・拡張子の検証に加え、実バイトの content-sniff を照合する。
	if err := uploadMeta.verifySniffedContent(file); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	storedName, err := uploadMeta.newStoredName(time.Now())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	// Upload via FileUploader (S3 or local). Persist the object key, never a public URL.
	key := uploadMeta.uploadKey(medicalRecordID, storedName)
	if _, err := h.uploader.Upload(c.Request.Context(), key, file, uploadMeta.mimeType); err != nil {
		httpapi.RespondError(c, apperrors.Wrap(err, "failed to upload file"))
		return
	}

	cleanupUpload := func() {
		if delErr := h.uploader.Delete(c.Request.Context(), key); delErr != nil {
			slog.WarnContext(c.Request.Context(), "failed to delete uploaded image on service error (best-effort)", "error", delErr, "key", key)
		}
	}

	now := time.Now()
	input := uploadMeta.toUploadedInput(key, now).toServiceInput()
	image, err := h.service.Create(c.Request.Context(), clinicID, medicalRecordID, input)
	if err != nil {
		cleanupUpload()
		httpapi.RespondError(c, err)
		return
	}

	resp, err := h.signedMedicalRecordImageResponse(c.Request.Context(), clinicID, image)
	if err != nil {
		h.retireUploadAfterSignFailure(c, clinicID, medicalRecordID, image.ID, staffID, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/medical-records/%d/images/%d", medicalRecordID, image.ID))
	c.JSON(http.StatusCreated, resp)
}

func (h *MedicalRecordImageHandler) retireUploadAfterSignFailure(
	c *gin.Context,
	clinicID, medicalRecordID, imageID, staffID uint64,
	signErr error,
) {
	retireErr := h.service.Delete(c.Request.Context(), clinicID, medicalRecordID, imageID, staffID, h.uploader)
	if retireErr == nil {
		httpapi.RespondError(c, apperrors.Wrap(signErr, "failed to sign medical record image url"))
		return
	}
	// 補償削除が error を返した場合、retirement が commit 済みか（台帳あり・画像行なし、
	// storage 掃除のみ未了）か、補償自体が失敗して画像行が残っているかを分けて返す。
	// どちらにも「保存成功」と誤解させる応答はしない。
	slog.ErrorContext(c.Request.Context(), "failed to retire medical record image after sign failure",
		slog.Uint64("image_id", imageID), slog.String("error", retireErr.Error()))
	retired, blockedErr := h.service.SigningBlocked(c.Request.Context(), clinicID, imageID)
	switch {
	case blockedErr != nil:
		// commit 状態を判定できない。再 upload が重複になるか不明なので断定しない。
		slog.ErrorContext(c.Request.Context(), "medical record image retirement state could not be determined",
			slog.Uint64("image_id", imageID), slog.String("error", blockedErr.Error()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf(
			"medical record image %d upload outcome is uncertain; check the image list before re-uploading",
			imageID)})
	case retired:
		// 台帳 commit 済み（storage 掃除のみ pending_storage で残る）: 画像は一覧に出ない。
		// 「保存済みだから一覧を見よ」とは言わず、再 upload が重複にならないことだけを伝える。
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf(
			"medical record image %d could not be signed and was retired; it will not appear in the image list and may be safely re-uploaded",
			imageID)})
	default:
		// 台帳なし = 補償失敗で画像行は残存。再 upload は重複になるため一覧確認を促す。
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf(
			"medical record image %d was saved but could not be signed; retry listing images instead of re-uploading",
			imageID)})
	}
}

// signedMedicalRecordImageResponse は署名発行→削除台帳確認の順で返す。
// 先に blocked 判定してから署名すると、判定後に削除 commit が割り込んで
// retired 画像の署名 URL を発行しうる（TOCTOU）。署名後に確認すれば
// 「発行済みだが削除済み」ケースを確実に検出できる（発行自体の競合窓は
// 署名 API 側の仕組みでは塞げないため、残存窓は台帳側で検出して応答を落とす）。
func (h *MedicalRecordImageHandler) signedMedicalRecordImageResponse(ctx context.Context, clinicID uint64, img *model.MedicalRecordImage) (medicalRecordImageResponse, error) {
	resp, err := h.signMedicalRecordImageResponse(ctx, img)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	blocked, err := h.service.SigningBlocked(ctx, clinicID, img.ID)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	if blocked {
		return medicalRecordImageResponse{}, errMedicalRecordImageSigningRetired
	}
	return resp, nil
}

// signMedicalRecordImageResponse は削除台帳を見ずに署名だけを行う
// （一覧時は署名後に BlockedImageIDs で一括突合する）。
func (h *MedicalRecordImageHandler) signMedicalRecordImageResponse(ctx context.Context, img *model.MedicalRecordImage) (medicalRecordImageResponse, error) {
	resp := toMedicalRecordImageResponse(img)
	imageURL, err := h.signStoredMedicalRecordImageURL(ctx, img.MedicalRecordID, resp.ImageURL)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	resp.ImageURL = imageURL
	thumbURL, err := h.signStoredMedicalRecordImageURL(ctx, img.MedicalRecordID, resp.ThumbnailURL)
	if err != nil {
		return medicalRecordImageResponse{}, err
	}
	resp.ThumbnailURL = thumbURL
	return resp, nil
}

func (h *MedicalRecordImageHandler) signStoredMedicalRecordImageURL(ctx context.Context, medicalRecordID uint64, stored string) (string, error) {
	key, isObject, err := resolveMedicalRecordImageStorageKey(stored)
	if err != nil {
		return "", err
	}
	if !isObject {
		return stored, nil
	}
	if !medicalRecordImageKeyBelongsToRecord(key, medicalRecordID) {
		if medicalRecordImageHasHTTPScheme(stored) {
			return stored, nil
		}
		return "", errMedicalRecordImageStorageKeyUnresolved
	}
	if h.uploader == nil {
		return "", fmt.Errorf("file uploader is not configured")
	}
	signed, err := h.uploader.GetSignedURL(ctx, key, medicalRecordImageSignedURLTTL)
	if err != nil {
		return "", err
	}
	if signed == "" {
		return "", fmt.Errorf("signed url is empty")
	}
	return signed, nil
}

// respondMedicalRecordImageUploadQuotaError maps quota errors to stable HTTP 429 messages
// (or fail-closed 500 when the quota store is unavailable).
func respondMedicalRecordImageUploadQuotaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errMedicalRecordImageUploadConcurrency):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadConcurrency.Error(),
		})
	case errors.Is(err, errMedicalRecordImageUploadRateLimit):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadRateLimit.Error(),
		})
	case errors.Is(err, errMedicalRecordImageUploadByteBudget):
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": errMedicalRecordImageUploadByteBudget.Error(),
		})
	default:
		// Infrastructure failure → fail-closed (do not allow unlimited upload).
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "upload quota unavailable",
		})
	}
}
