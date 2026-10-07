package trimming

import (
	"context"
	"fmt"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// signTrimmingResponse signs style_image and completed_image when they are
// object keys owned by the record's clinic. Legacy values pass through.
// Foreign or malformed object keys fail closed. The persisted model is unchanged.
func (h *Handler) signTrimmingResponse(ctx context.Context, clinicID uint64, appt *model.Reservation) (TrimmingResponse, error) {
	if appt == nil {
		return TrimmingResponse{}, nil
	}
	resp := toTrimmingResponse(appt)
	if appt.TrimmingDetail == nil {
		return resp, nil
	}
	ownerClinic := trimmingRecordClinicID(appt, clinicID)
	styleImage, err := h.signStoredTrimmingImageURL(ctx, ownerClinic, resp.StyleImage)
	if err != nil {
		return TrimmingResponse{}, err
	}
	resp.StyleImage = styleImage
	completedImage, err := h.signStoredTrimmingImageURL(ctx, ownerClinic, resp.CompletedImage)
	if err != nil {
		return TrimmingResponse{}, err
	}
	resp.CompletedImage = completedImage
	return resp, nil
}

func (h *Handler) signedTrimmingResponses(ctx context.Context, clinicID uint64, appts []model.Reservation) ([]TrimmingResponse, error) {
	responses := make([]TrimmingResponse, 0, len(appts))
	for i := range appts {
		resp, err := h.signTrimmingResponse(ctx, clinicID, &appts[i])
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
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
