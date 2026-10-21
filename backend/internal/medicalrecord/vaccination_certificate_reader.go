package medicalrecord

// vaccination_certificate_reader.go — EMR-486 clinic-scoped read path for
// certificate series/editions/artifacts plus the current-validity projection.
//
// Every query is scoped by the authenticated clinic_id. The vaccination lookup
// is Unscoped() because soft-deleted rows must remain readable to project
// source_deleted. Missing or foreign-clinic rows report one identical
// not_found (no existence leak). The reader never fills missing snapshot
// values from current master/owner/pet data.

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// certificateStore is the read capability the certificate handler consumes.
// *VaccinationCertificateReader implements it against GORM; tests substitute
// fakes.
type certificateStore interface {
	FindEdition(ctx context.Context, clinicID, editionID uint64) (*model.VaccinationCertificateEdition, *model.VaccinationCertificateSeries, error)
	EditionValidity(ctx context.Context, clinicID uint64, ed *model.VaccinationCertificateEdition, series *model.VaccinationCertificateSeries) (string, error)
	FindVaccinationSource(ctx context.Context, clinicID, vaccinationID uint64) (*model.Vaccination, error)
	FindSeries(ctx context.Context, clinicID, vaccinationID uint64, kind string) (*model.VaccinationCertificateSeries, bool, error)
	ListEditions(ctx context.Context, clinicID, vaccinationID, seriesID uint64, offset, limit int) (*model.Vaccination, []model.VaccinationCertificateEdition, int64, int64, error)
	FindArtifact(ctx context.Context, clinicID, editionID uint64, layout string) (*model.VaccinationCertificateArtifact, bool, error)
}

// VaccinationCertificateReader is the GORM-backed read side for certificates.
type VaccinationCertificateReader struct {
	db *gorm.DB
}

func NewVaccinationCertificateReader(db *gorm.DB) *VaccinationCertificateReader {
	return &VaccinationCertificateReader{db: db}
}

// FindEdition loads one edition and its series under clinic scope. Missing
// edition or series (including foreign clinic) is a single not_found.
func (r *VaccinationCertificateReader) FindEdition(ctx context.Context, clinicID, editionID uint64) (*model.VaccinationCertificateEdition, *model.VaccinationCertificateSeries, error) {
	var ed model.VaccinationCertificateEdition
	err := r.db.WithContext(ctx).Where("clinic_id = ? AND id = ?", clinicID, editionID).Take(&ed).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errCertificateNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	var series model.VaccinationCertificateSeries
	err = r.db.WithContext(ctx).Where("clinic_id = ? AND id = ?", clinicID, ed.SeriesID).Take(&series).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errCertificateNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	return &ed, &series, nil
}

// FindVaccinationSource loads a vaccination including soft-deleted rows (a
// soft-deleted source must remain readable to project source_deleted), while
// the current patient/clinic/vaccination/original-record relation scope is
// still enforced. Missing or relation-failing rows -> not_found.
func (r *VaccinationCertificateReader) FindVaccinationSource(ctx context.Context, clinicID, vaccinationID uint64) (*model.Vaccination, error) {
	var vacc model.Vaccination
	err := r.db.WithContext(ctx).Unscoped().
		Select("id", "clinic_id", "version", "deleted_at").
		Scopes(vaccinationPatientRelationsScope(clinicID)).
		Where("id = ? AND clinic_id = ?", vaccinationID, clinicID).
		Take(&vacc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errCertificateNotFound
	}
	if err != nil {
		return nil, err
	}
	return &vacc, nil
}

// EditionValidity projects current validity for one edition:
// source_deleted > source_changed > superseded > current. The scoped
// vaccination read and the newer-edition count run inside one read-only
// REPEATABLE READ transaction so the projection uses a single snapshot: a
// concurrent correction committed between otherwise independent reads could
// hide source_changed behind superseded. A row missing under the current
// relation scope is not_found, never silently treated as deleted.
func (r *VaccinationCertificateReader) EditionValidity(ctx context.Context, clinicID uint64, ed *model.VaccinationCertificateEdition, series *model.VaccinationCertificateSeries) (string, error) {
	var vacc model.Vaccination
	var newer int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Select("id", "clinic_id", "version", "deleted_at").
			Scopes(vaccinationPatientRelationsScope(clinicID)).
			Where("id = ? AND clinic_id = ?", series.VaccinationID, clinicID).
			Take(&vacc).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errCertificateNotFound
		}
		if err != nil {
			return err
		}
		return tx.Model(&model.VaccinationCertificateEdition{}).
			Where("clinic_id = ? AND series_id = ? AND edition > ?", clinicID, ed.SeriesID, ed.Edition).
			Count(&newer).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return "", err
	}
	return certificateValidity(vacc.DeletedAt.Valid, int64(vacc.Version), ed.SourceVaccinationVersion, newer > 0), nil
}

// FindSeries returns the (clinic, vaccination, kind) series, or found=false.
func (r *VaccinationCertificateReader) FindSeries(ctx context.Context, clinicID, vaccinationID uint64, kind string) (*model.VaccinationCertificateSeries, bool, error) {
	var series model.VaccinationCertificateSeries
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND vaccination_id = ? AND certificate_kind = ?", clinicID, vaccinationID, kind).
		Take(&series).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &series, true, nil
}

// ListEditions returns one page ordered by edition DESC plus total count and
// the maximum edition number in the series (for superseded projection). The
// scoped vaccination read and the three series queries run inside one
// read-only REPEATABLE READ transaction so validity is projected from a
// single snapshot: an independent earlier read could mix a stale source
// version/deleted_at with a newer page and mark two editions current.
// Missing or relation-failing vaccination rows -> not_found.
func (r *VaccinationCertificateReader) ListEditions(ctx context.Context, clinicID, vaccinationID, seriesID uint64, offset, limit int) (*model.Vaccination, []model.VaccinationCertificateEdition, int64, int64, error) {
	var total, maxEdition int64
	var eds []model.VaccinationCertificateEdition
	var vacc model.Vaccination
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().
			Select("id", "clinic_id", "version", "deleted_at").
			Scopes(vaccinationPatientRelationsScope(clinicID)).
			Where("id = ? AND clinic_id = ?", vaccinationID, clinicID).
			Take(&vacc).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errCertificateNotFound
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&model.VaccinationCertificateEdition{}).
			Where("clinic_id = ? AND series_id = ?", clinicID, seriesID).
			Count(&total).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.VaccinationCertificateEdition{}).
			Where("clinic_id = ? AND series_id = ?", clinicID, seriesID).
			Select("COALESCE(MAX(edition),0)").
			Scan(&maxEdition).Error; err != nil {
			return err
		}
		eds = make([]model.VaccinationCertificateEdition, 0, limit)
		return tx.
			Where("clinic_id = ? AND series_id = ?", clinicID, seriesID).
			Order("edition DESC").Offset(offset).Limit(limit).
			Find(&eds).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, nil, 0, 0, err
	}
	return &vacc, eds, total, maxEdition, nil
}

// FindArtifact returns the (clinic, edition, layout) artifact row, or
// found=false when none exists.
func (r *VaccinationCertificateReader) FindArtifact(ctx context.Context, clinicID, editionID uint64, layout string) (*model.VaccinationCertificateArtifact, bool, error) {
	var row model.VaccinationCertificateArtifact
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND edition_id = ? AND layout = ?", clinicID, editionID, layout).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &row, true, nil
}

// certificateValidity is the pure priority projection:
// source_deleted > source_changed > superseded > current.
func certificateValidity(sourceDeleted bool, currentVersion, sourceVersion int64, superseded bool) string {
	switch {
	case sourceDeleted:
		return "source_deleted"
	case currentVersion != sourceVersion:
		return "source_changed"
	case superseded:
		return "superseded"
	default:
		return "current"
	}
}

// availableCertificateLayouts lists the layouts allowed for kind that also
// carry a template key in render_manifest.templates, in fixed
// portrait/landscape/rabies order. Kinds are never mixed.
func availableCertificateLayouts(kind string, templates map[string]string) []string {
	want := []string{"portrait", "landscape"}
	if kind == "rabies" {
		want = []string{"rabies"}
	}
	out := make([]string, 0, len(want))
	for _, l := range want {
		if _, ok := templates[l]; ok {
			out = append(out, l)
		}
	}
	return out
}

// checkCertificateSnapshotAnchors verifies that a decoded snapshot is anchored
// to the edition and series that produced it. Any mismatch is 503
// result-unavailable (never leaking snapshot contents).
func checkCertificateSnapshotAnchors(snap *certificateSnapshot, ed *model.VaccinationCertificateEdition, series *model.VaccinationCertificateSeries) error {
	if snap.Kind != series.CertificateKind {
		return errCertificateResultUnavailable
	}
	if snap.Source.VaccinationID != int64(series.VaccinationID) {
		return errCertificateResultUnavailable
	}
	if snap.Source.Version != ed.SourceVaccinationVersion {
		return errCertificateResultUnavailable
	}
	return nil
}

// buildCertificateEditionDTO decodes and cross-checks the stored snapshot and
// produces the wire DTO. A structurally invalid snapshot is 503
// result-unavailable; schema_version != 1 is 409 dependency-changed (inside
// decodeCertificateSnapshot).
func buildCertificateEditionDTO(ed *model.VaccinationCertificateEdition, series *model.VaccinationCertificateSeries, validity string) (*certificateEditionDTO, error) {
	snap, err := decodeCertificateSnapshot(ed.Snapshot)
	if err != nil {
		return nil, err
	}
	if err := checkCertificateSnapshotAnchors(snap, ed, series); err != nil {
		return nil, err
	}
	sha, err := certificateSnapshotSHA256(snap.Doc)
	if err != nil {
		return nil, errCertificateResultUnavailable
	}
	return &certificateEditionDTO{
		ID:                ed.ID,
		SeriesID:          ed.SeriesID,
		VaccinationID:     series.VaccinationID,
		Kind:              series.CertificateKind,
		Edition:           ed.Edition,
		SourceVersion:     ed.SourceVaccinationVersion,
		PreviousEditionID: ed.PreviousEditionID,
		IssuedAt:          ed.IssuedAt.UTC(),
		IssuedBy:          ed.IssuedBy,
		Validity:          validity,
		Snapshot:          ed.Snapshot,
		SnapshotSHA256:    sha,
		AvailableLayouts:  availableCertificateLayouts(series.CertificateKind, snap.Templates),
	}, nil
}
