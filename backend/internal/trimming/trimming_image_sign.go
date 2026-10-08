package trimming

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// signTrimmingResponse signs style_image and completed_image when they are
// object keys owned by the record's clinic. Every signing failure degrades to
// an unsigned echo of the stored value (warn-logged) rather than failing the
// response — writes are the strict side (normalizeTrimmingImageForPersist
// rejects foreign keys), while a stray legacy/imported value or a transient
// signer outage must not hide records on read, and create/update must not
// report failure for a record that already committed. The persisted model is
// unchanged.
func (h *Handler) signTrimmingResponse(ctx context.Context, clinicID uint64, appt *model.Reservation) TrimmingResponse {
	if appt == nil {
		return TrimmingResponse{}
	}
	resp := toTrimmingResponse(appt)
	if appt.TrimmingDetail == nil {
		return resp
	}
	ownerClinic := trimmingRecordClinicID(appt, clinicID)
	resp.StyleImage = h.signStoredTrimmingImageURLOrEcho(ctx, ownerClinic, resp.StyleImage)
	resp.CompletedImage = h.signStoredTrimmingImageURLOrEcho(ctx, ownerClinic, resp.CompletedImage)
	return resp
}

func (h *Handler) signedTrimmingResponses(ctx context.Context, clinicID uint64, appts []model.Reservation) []TrimmingResponse {
	responses := make([]TrimmingResponse, 0, len(appts))
	for i := range appts {
		responses = append(responses, h.signTrimmingResponse(ctx, clinicID, &appts[i]))
	}
	return responses
}

// signStoredTrimmingImageURLOrEcho signs one stored image value for display.
// Owned object keys become signed URLs; other values echo unsigned. Any
// failure — malformed or foreign key, missing uploader, signer error — degrades
// to echoing the stored value so one bad field cannot take down a record or
// list response.
func (h *Handler) signStoredTrimmingImageURLOrEcho(ctx context.Context, clinicID uint64, stored string) string {
	signed, err := h.signStoredTrimmingImageURL(ctx, clinicID, stored)
	if err != nil {
		slog.WarnContext(ctx, "trimming image signing failed; echoing stored value",
			slog.String("error", err.Error()),
			slog.Uint64("clinic_id", clinicID))
		return stored
	}
	return signed
}

func (h *Handler) signStoredTrimmingImageURL(ctx context.Context, clinicID uint64, stored string) (string, error) {
	key, isObject, err := resolveTrimmingImageStorageKey(stored)
	if err != nil {
		return "", err
	}
	if !isObject {
		return stored, nil
	}
	if !trimmingImageKeyBelongsToClinic(key, clinicID) {
		// Foreign http(s) values echo unsigned directly; bare foreign keys error
		// here but the caller degrades to echo — either way they are never signed.
		if trimmingImageHasHTTPScheme(stored) {
			return stored, nil
		}
		return "", errTrimmingImageStorageKeyUnresolved
	}
	if h.imageUploader == nil {
		return "", fmt.Errorf("file uploader is not configured")
	}
	signed, err := h.imageUploader.GetSignedURL(ctx, key, trimmingImageSignedURLTTL)
	if err != nil {
		return "", err
	}
	if signed == "" {
		return "", fmt.Errorf("signed url is empty")
	}
	return signed, nil
}

func trimmingStoredImages(appt *model.Reservation) (style, completed string, ok bool) {
	if appt == nil || appt.TrimmingDetail == nil {
		return "", "", false
	}
	return appt.TrimmingDetail.StyleImage, appt.TrimmingDetail.CompletedImage, true
}

func trimmingRecordClinicID(appt *model.Reservation, fallback uint64) uint64 {
	if appt != nil && appt.ClinicID != 0 {
		return appt.ClinicID
	}
	return fallback
}

func normalizeCreateTrimmingImages(clinicID uint64, input *CreateTrimmingInput) error {
	if input == nil {
		return nil
	}
	style, err := normalizeTrimmingImageForPersist(clinicID, input.StyleImage)
	if err != nil {
		return err
	}
	completed, err := normalizeTrimmingImageForPersist(clinicID, input.CompletedImage)
	if err != nil {
		return err
	}
	input.StyleImage = style
	input.CompletedImage = completed
	return nil
}

func normalizeUpdateTrimmingImages(clinicID uint64, input *UpdateTrimmingInput) error {
	if input == nil {
		return nil
	}
	if err := assignNormalizedTrimmingImage(clinicID, input.StyleImage); err != nil {
		return err
	}
	return assignNormalizedTrimmingImage(clinicID, input.CompletedImage)
}

func assignNormalizedTrimmingImage(clinicID uint64, value *string) error {
	if value == nil {
		return nil
	}
	normalized, err := normalizeTrimmingImageForPersist(clinicID, *value)
	if err != nil {
		return err
	}
	*value = normalized
	return nil
}

// normalizeTrimmingImageForPersist returns the value that may be stored.
// Owned object keys, including those embedded in an http(s) or /uploads/ URL,
// become the canonical key. Bare http(s) values and stray /uploads/ values are rejected.
func normalizeTrimmingImageForPersist(clinicID uint64, stored string) (string, error) {
	trimmed := strings.TrimSpace(stored)
	if trimmed == "" {
		return "", nil
	}
	key, isObject, err := resolveTrimmingImageStorageKey(trimmed)
	if err != nil {
		return "", apperrors.WrapInvalidInput("invalid trimming image reference")
	}
	if isObject {
		if !trimmingImageKeyBelongsToClinic(key, clinicID) {
			return "", apperrors.WrapInvalidInput("trimming image does not belong to this clinic")
		}
		return key, nil
	}
	if trimmingImageHasHTTPScheme(trimmed) ||
		strings.HasPrefix(trimmed, trimmingImageLocalURLPrefix) ||
		strings.HasPrefix(trimmed, trimmingImageStorageKeyPrefix) {
		return "", apperrors.WrapInvalidInput("trimming image must be stored as an object key")
	}
	return trimmed, nil
}
