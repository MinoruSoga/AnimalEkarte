package medicalrecord

// vaccination_certificate_artifact.go — EMR-486 fail-safe private artifact read.
//
// Consumer-side interface only: medicalrecord must not import internal/infra
// (ADR-006 dependency direction), and no storage binding/config exists yet, so
// the production composition root passes nil and every artifact read returns
// 503 artifact_unavailable until the owning writer session binds storage.
//
// Fail-safe rules (never re-render when an artifact row exists, never log the
// object key or contents): nil store, open/read error, oversize blob, sha256
// mismatch against the row, or non-PDF magic bytes all map to
// artifact_unavailable.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"

	"github.com/animal-ekarte/backend/internal/model"
)

// certificateArtifactStore is the narrow read capability this package consumes.
// The writer-side FileStorage adapter implements it in a later session.
type certificateArtifactStore interface {
	OpenCertificateArtifact(ctx context.Context, objectKey string) (io.ReadCloser, error)
}

// certificateArtifactMaxBytes caps a certificate PDF read (16 MiB).
const certificateArtifactMaxBytes = 16 << 20

var certificatePDFMagic = []byte("%PDF-")

// readCertificateArtifact streams the object, enforces the size cap, verifies
// the stored digest, and checks the PDF magic bytes. Every failure mode maps to
// errCertificateArtifactUnavailable without leaking the object key or content.
func readCertificateArtifact(ctx context.Context, store certificateArtifactStore, row *model.VaccinationCertificateArtifact) ([]byte, error) {
	if store == nil || row == nil {
		return nil, errCertificateArtifactUnavailable
	}
	rc, err := store.OpenCertificateArtifact(ctx, row.ObjectKey)
	if err != nil {
		return nil, errCertificateArtifactUnavailable
	}
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(io.LimitReader(rc, certificateArtifactMaxBytes+1))
	if err != nil {
		return nil, errCertificateArtifactUnavailable
	}
	if len(data) > certificateArtifactMaxBytes {
		return nil, errCertificateArtifactUnavailable
	}
	sum := sha256.Sum256(data)
	if !bytes.Equal(sum[:], row.ContentSHA256) {
		return nil, errCertificateArtifactUnavailable
	}
	if !bytes.HasPrefix(data, certificatePDFMagic) {
		return nil, errCertificateArtifactUnavailable
	}
	return data, nil
}
