package medicalrecord

// vaccination_certificate_handler.go — EMR-486 HTTP boundary for the three
// read-only certificate operations:
//
//	GET  /vaccination-certificate-editions/:edition_id
//	GET  /vaccinations/:id/certificate-editions?kind=&page=&limit=
//	POST /vaccination-certificate-editions/:edition_id/preview
//
// All three are vaccinations:view reads (POST preview is still a read). Errors
// are the flat {"error","code"} shape with codes only from the
// CertificateError enum; no `status` is added. DB failures go through the
// existing generic 500 path without SQL details. Snapshots, object keys and
// personal data never appear in errors or logs.

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// certificateError carries an HTTP status plus a CertificateError-enum code.
type certificateError struct {
	status  int
	code    string
	message string
}

func (e *certificateError) Error() string { return e.message }

var (
	errCertificateNotFound            = &certificateError{http.StatusNotFound, "not_found", "certificate resource not found"}
	errCertificateDependencyChanged   = &certificateError{http.StatusConflict, "dependency-changed", "certificate dependency has changed"}
	errCertificateResultUnavailable   = &certificateError{http.StatusServiceUnavailable, "result-unavailable", "certificate result is unavailable"}
	errCertificateRenderFailed        = &certificateError{http.StatusServiceUnavailable, "render_failed", "certificate rendering failed"}
	errCertificateArtifactUnavailable = &certificateError{http.StatusServiceUnavailable, "artifact_unavailable", "certificate artifact is unavailable"}
)

func certificateValidationError(message string) error {
	return &certificateError{http.StatusBadRequest, "validation", message}
}

func certificateUnprocessableError(message string) error {
	return &certificateError{http.StatusUnprocessableEntity, "validation", message}
}

// writeCertificateError emits the flat {"error","code"} body for certificate
// errors and defers everything else (e.g. DB failures) to the existing generic
// path, which yields a detail-free 500.
func writeCertificateError(c *gin.Context, err error) {
	var ce *certificateError
	if errors.As(err, &ce) {
		c.JSON(ce.status, gin.H{"error": ce.message, "code": ce.code})
		return
	}
	httpapi.RespondError(c, err)
}

type certificateEditionDTO struct {
	ID                uint64          `json:"id"`
	SeriesID          uint64          `json:"series_id"`
	VaccinationID     uint64          `json:"vaccination_id"`
	Kind              string          `json:"kind"`
	Edition           int64           `json:"edition"`
	SourceVersion     int64           `json:"source_version"`
	PreviousEditionID *uint64         `json:"previous_edition_id"`
	IssuedAt          time.Time       `json:"issued_at"`
	IssuedBy          uint64          `json:"issued_by"`
	Validity          string          `json:"validity"`
	Snapshot          json.RawMessage `json:"snapshot"`
	SnapshotSHA256    string          `json:"snapshot_sha256"`
	AvailableLayouts  []string        `json:"available_layouts"`
}

type certificateEditionPageDTO struct {
	Data  []certificateEditionDTO `json:"data"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

type certificateSavedPreviewDTO struct {
	EditionID       uint64 `json:"edition_id"`
	Layout          string `json:"layout"`
	PDFBase64       string `json:"pdf_base64"`
	MediaType       string `json:"media_type"`
	TemplateVersion string `json:"template_version"`
	RendererVersion string `json:"renderer_version"`
	ContentSHA256   string `json:"content_sha256"`
}

// VaccinationCertificateHandler serves the certificate read boundary.
type VaccinationCertificateHandler struct {
	store     certificateStore
	artifacts certificateArtifactStore
}

func NewVaccinationCertificateHandler(store certificateStore, artifacts certificateArtifactStore) *VaccinationCertificateHandler {
	return &VaccinationCertificateHandler{store: store, artifacts: artifacts}
}

// parseCertificateIDParam accepts only positive int64 path ids; values above
// 9223372036854775807 or non-positive/zero are 400 validation.
func parseCertificateIDParam(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		writeCertificateError(c, certificateValidationError("invalid "+key))
		return 0, false
	}
	return uint64(id), true
}

// writeCertificateBodyDecodeError maps an oversized body (server-imposed
// MaxBytesReader limit) to 413 payload-too-large, following the
// staff/http_binding.go pattern; any other decode failure is 400 validation.
func writeCertificateBodyDecodeError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		httpapi.RespondError(c, apperrors.WrapPayloadTooLarge("request body exceeds size limit"))
		return
	}
	writeCertificateError(c, certificateValidationError("invalid request body"))
}

// GetVaccinationCertificateEdition handles GET /vaccination-certificate-editions/:edition_id.
func (h *VaccinationCertificateHandler) GetVaccinationCertificateEdition(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceVaccinations), "view") {
		return
	}
	editionID, ok := parseCertificateIDParam(c, "edition_id")
	if !ok {
		return
	}
	ctx := c.Request.Context()
	ed, series, err := h.store.FindEdition(ctx, clinicID, editionID)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	// Current patient/clinic/vaccination/original-record relation check on
	// every request: a source that fails the scoped lookup is not_found.
	if _, err := h.store.FindVaccinationSource(ctx, clinicID, series.VaccinationID); err != nil {
		writeCertificateError(c, err)
		return
	}
	validity, err := h.store.EditionValidity(ctx, clinicID, ed, series)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	dto, err := buildCertificateEditionDTO(ed, series, validity)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto)
}

// ListVaccinationCertificateEditions handles
// GET /vaccinations/:id/certificate-editions?kind=&page=&limit=.
// The path wildcard is :id to match the existing /vaccinations/:id routes.
func (h *VaccinationCertificateHandler) ListVaccinationCertificateEditions(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceVaccinations), "view") {
		return
	}
	vaccinationID, ok := parseCertificateIDParam(c, "id")
	if !ok {
		return
	}
	kind := c.Query("kind")
	if kind != "normal" && kind != "rabies" {
		writeCertificateError(c, certificateValidationError("kind must be normal or rabies"))
		return
	}
	page, limit, err := httpapi.ParsePaginationWithMax(c, 100)
	if err != nil {
		writeCertificateError(c, certificateValidationError("invalid pagination parameters"))
		return
	}
	// ParsePaginationWithMax leaves page unbounded above; reject the request
	// rather than let (page-1)*limit overflow the int64 offset.
	if page-1 > math.MaxInt64/limit {
		writeCertificateError(c, certificateValidationError("invalid pagination parameters"))
		return
	}
	ctx := c.Request.Context()
	// Current relation/existence gate only: version and deleted_at used for
	// validity must come from the single ListEditions snapshot below, not
	// this earlier read.
	if _, err := h.store.FindVaccinationSource(ctx, clinicID, vaccinationID); err != nil {
		writeCertificateError(c, err)
		return
	}
	series, found, err := h.store.FindSeries(ctx, clinicID, vaccinationID, kind)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	if !found {
		c.JSON(http.StatusOK, certificateEditionPageDTO{
			Data:  []certificateEditionDTO{},
			Total: 0,
			Page:  page,
			Limit: limit,
		})
		return
	}
	vacc, eds, total, maxEdition, err := h.store.ListEditions(ctx, clinicID, vaccinationID, series.ID, (page-1)*limit, limit)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	deleted := vacc.DeletedAt.Valid
	currentVersion := int64(vacc.Version)
	data := make([]certificateEditionDTO, 0, len(eds))
	for i := range eds {
		validity := certificateValidity(deleted, currentVersion, eds[i].SourceVaccinationVersion, eds[i].Edition < maxEdition)
		dto, err := buildCertificateEditionDTO(&eds[i], series, validity)
		if err != nil {
			writeCertificateError(c, err)
			return
		}
		data = append(data, *dto)
	}
	c.JSON(http.StatusOK, certificateEditionPageDTO{Data: data, Total: total, Page: page, Limit: limit})
}

// PreviewSavedVaccinationCertificate handles
// POST /vaccination-certificate-editions/:edition_id/preview.
// Read-only: no artifact row, receipt or audit is created; existing artifacts
// are digest-verified reads, otherwise a normal edition is purely re-rendered
// from its stored snapshot when the manifest matches this server's versions.
func (h *VaccinationCertificateHandler) PreviewSavedVaccinationCertificate(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	if !httpapi.RequireSelectedClinicGrant(c, string(model.ResourceVaccinations), "view") {
		return
	}
	editionID, ok := parseCertificateIDParam(c, "edition_id")
	if !ok {
		return
	}
	var body struct {
		Layout string `json:"layout"`
	}
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeCertificateBodyDecodeError(c, err)
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeCertificateBodyDecodeError(c, err)
		return
	}
	if body.Layout != "portrait" && body.Layout != "landscape" && body.Layout != "rabies" {
		writeCertificateError(c, certificateValidationError("invalid layout"))
		return
	}
	ctx := c.Request.Context()
	ed, series, err := h.store.FindEdition(ctx, clinicID, editionID)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	// Current patient/clinic/vaccination/original-record relation check on
	// every request, before the snapshot is even decoded.
	if _, err := h.store.FindVaccinationSource(ctx, clinicID, series.VaccinationID); err != nil {
		writeCertificateError(c, err)
		return
	}
	snap, err := decodeCertificateSnapshot(ed.Snapshot)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	if err := checkCertificateSnapshotAnchors(snap, ed, series); err != nil {
		writeCertificateError(c, err)
		return
	}
	validity, err := h.store.EditionValidity(ctx, clinicID, ed, series)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	if validity != "current" {
		writeCertificateError(c, errCertificateDependencyChanged)
		return
	}
	if !certificateLayoutAllowed(series.CertificateKind, body.Layout) {
		writeCertificateError(c, certificateUnprocessableError("layout does not match certificate kind"))
		return
	}
	if !slices.Contains(availableCertificateLayouts(series.CertificateKind, snap.Templates), body.Layout) {
		writeCertificateError(c, errCertificateDependencyChanged)
		return
	}
	var pdf []byte
	var templateVersion, rendererVersion string
	artifact, found, err := h.store.FindArtifact(ctx, clinicID, ed.ID, body.Layout)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	if found {
		// An artifact row exists: verified read only, never a re-render. The
		// row's versions must equal the stored snapshot manifest versions.
		if artifact.TemplateVersion != snap.Templates[body.Layout] ||
			artifact.RendererVersion != snap.RendererVersion {
			writeCertificateError(c, errCertificateArtifactUnavailable)
			return
		}
		pdf, err = readCertificateArtifact(ctx, h.artifacts, artifact)
		if err != nil {
			writeCertificateError(c, err)
			return
		}
		templateVersion = artifact.TemplateVersion
		rendererVersion = artifact.RendererVersion
	} else if series.CertificateKind == "normal" {
		// Pure re-render of the stored snapshot is allowed only when the saved
		// manifest equals this server's supported versions.
		if snap.RendererVersion != certificateRendererVersion ||
			snap.FontAssetVersion != certificateFontAssetVersion ||
			snap.Templates[body.Layout] != certificateTemplateForLayout(body.Layout) {
			writeCertificateError(c, errCertificateDependencyChanged)
			return
		}
		if !certificateRendererProductionEnabled {
			writeCertificateError(c, errCertificateRenderFailed)
			return
		}
		pdf, err = renderCertificatePDF(snap, body.Layout)
		if err != nil {
			writeCertificateError(c, err)
			return
		}
		templateVersion = snap.Templates[body.Layout]
		rendererVersion = snap.RendererVersion
	} else {
		writeCertificateError(c, errCertificateArtifactUnavailable)
		return
	}
	// Second read: validity must still be current right before responding.
	validity, err = h.store.EditionValidity(ctx, clinicID, ed, series)
	if err != nil {
		writeCertificateError(c, err)
		return
	}
	if validity != "current" {
		writeCertificateError(c, errCertificateDependencyChanged)
		return
	}
	sum := sha256.Sum256(pdf)
	c.Header("Cache-Control", "private, no-store")
	c.JSON(http.StatusOK, certificateSavedPreviewDTO{
		EditionID:       ed.ID,
		Layout:          body.Layout,
		PDFBase64:       base64.StdEncoding.EncodeToString(pdf),
		MediaType:       "application/pdf",
		TemplateVersion: templateVersion,
		RendererVersion: rendererVersion,
		ContentSHA256:   hex.EncodeToString(sum[:]),
	})
}
