package medicalrecord

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeCertificateStore struct {
	ed            *model.VaccinationCertificateEdition
	series        *model.VaccinationCertificateSeries
	validitySeq   []string
	validityCalls int
	vacc          *model.Vaccination
	listVacc      *model.Vaccination
	seriesFound   bool
	eds           []model.VaccinationCertificateEdition
	total         int64
	maxEdition    int64
	artifact      *model.VaccinationCertificateArtifact
	artifactFound bool
}

func (f *fakeCertificateStore) FindEdition(_ context.Context, _, _ uint64) (*model.VaccinationCertificateEdition, *model.VaccinationCertificateSeries, error) {
	if f.ed == nil || f.series == nil {
		return nil, nil, errCertificateNotFound
	}
	return f.ed, f.series, nil
}

func (f *fakeCertificateStore) EditionValidity(_ context.Context, _ uint64, _ *model.VaccinationCertificateEdition, _ *model.VaccinationCertificateSeries) (string, error) {
	i := f.validityCalls
	f.validityCalls++
	if len(f.validitySeq) == 0 {
		return "current", nil
	}
	if i >= len(f.validitySeq) {
		i = len(f.validitySeq) - 1
	}
	return f.validitySeq[i], nil
}

func (f *fakeCertificateStore) FindVaccinationSource(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
	if f.vacc == nil {
		return nil, errCertificateNotFound
	}
	return f.vacc, nil
}

func (f *fakeCertificateStore) FindSeries(_ context.Context, _, _ uint64, _ string) (*model.VaccinationCertificateSeries, bool, error) {
	if !f.seriesFound || f.series == nil {
		return nil, false, nil
	}
	return f.series, true, nil
}

func (f *fakeCertificateStore) ListEditions(_ context.Context, _, _, _ uint64, _, _ int) (*model.Vaccination, []model.VaccinationCertificateEdition, int64, int64, error) {
	v := f.vacc
	if f.listVacc != nil {
		v = f.listVacc
	}
	if v == nil {
		return nil, nil, 0, 0, errCertificateNotFound
	}
	return v, f.eds, f.total, f.maxEdition, nil
}

func (f *fakeCertificateStore) FindArtifact(_ context.Context, _, _ uint64, _ string) (*model.VaccinationCertificateArtifact, bool, error) {
	return f.artifact, f.artifactFound, nil
}

type fakeArtifactStore struct {
	data []byte
	err  error
}

func (f *fakeArtifactStore) OpenCertificateArtifact(_ context.Context, _ string) (io.ReadCloser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return io.NopCloser(bytes.NewReader(f.data)), nil
}

// ─── Fixtures ────────────────────────────────────────────────────────────────

func testCertificateEditionFixture() (*model.VaccinationCertificateEdition, *model.VaccinationCertificateSeries) {
	series := &model.VaccinationCertificateSeries{ID: 10, ClinicID: 1, VaccinationID: 100, CertificateKind: "normal"}
	ed := &model.VaccinationCertificateEdition{
		ID: 20, ClinicID: 1, SeriesID: 10, Edition: 1, SourceVaccinationVersion: 2,
		Snapshot: json.RawMessage(testCertificateSnapshotJSON),
		IssuedBy: 5, IssuedAt: time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
	}
	return ed, series
}

func newCertCtx(t *testing.T, method, target string, body []byte, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, bytes.NewReader(body))
	c.Params = params
	c.Set("clinic_id", "1")
	c.Set("is_system_admin", true)
	return c, w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]certJSONValue {
	t.Helper()
	obj, err := strictJSONObject(w.Body.Bytes())
	require.NoError(t, err)
	return obj.Obj
}

// ─── Handler tests ───────────────────────────────────────────────────────────

func TestGetVaccinationCertificateEdition(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	vacc := &model.Vaccination{ID: 100, ClinicID: 1, Version: 2}
	h := NewVaccinationCertificateHandler(&fakeCertificateStore{ed: ed, series: series, vacc: vacc}, nil)

	c, w := newCertCtx(t, http.MethodGet, "/", nil, gin.Params{{Key: "edition_id", Value: "20"}})
	h.GetVaccinationCertificateEdition(c)
	assert.Equal(t, http.StatusOK, w.Code)
	body := decodeBody(t, w)
	assert.Equal(t, "20", body["id"].Num.String())
	assert.Equal(t, "100", body["vaccination_id"].Num.String())
	assert.Equal(t, "normal", body["kind"].Str)
	assert.Equal(t, "current", body["validity"].Str)
	assert.Len(t, body["snapshot_sha256"].Str, 64)
	layouts := body["available_layouts"]
	require.Len(t, layouts.Arr, 2)
	assert.Equal(t, "portrait", layouts.Arr[0].Str)
	assert.Equal(t, "landscape", layouts.Arr[1].Str)

	// Cross-clinic / missing edition: identical 404 not_found.
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{}, nil)
	c, w = newCertCtx(t, http.MethodGet, "/", nil, gin.Params{{Key: "edition_id", Value: "999"}})
	h.GetVaccinationCertificateEdition(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "not_found", decodeBody(t, w)["code"].Str)
	assert.NotContains(t, w.Body.String(), "snapshot", "no existence/snapshot leak")

	// Edition exists but the current source relation check fails -> 404.
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: ed, series: series}, nil)
	c, w = newCertCtx(t, http.MethodGet, "/", nil, gin.Params{{Key: "edition_id", Value: "20"}})
	h.GetVaccinationCertificateEdition(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "not_found", decodeBody(t, w)["code"].Str)

	// Malformed / non-positive / int64-overflowing ids -> 400 validation.
	for _, id := range []string{"abc", "0", "-5", "9223372036854775808"} {
		c, w = newCertCtx(t, http.MethodGet, "/", nil, gin.Params{{Key: "edition_id", Value: id}})
		h.GetVaccinationCertificateEdition(c)
		assert.Equal(t, http.StatusBadRequest, w.Code, id)
		assert.Equal(t, "validation", decodeBody(t, w)["code"].Str, id)
	}
}

func TestListVaccinationCertificateEditions(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	vacc := &model.Vaccination{ID: 100, ClinicID: 1, Version: 2}
	store := &fakeCertificateStore{
		ed: ed, series: series, seriesFound: true, vacc: vacc,
		eds: []model.VaccinationCertificateEdition{*ed}, total: 1, maxEdition: 1,
	}
	h := NewVaccinationCertificateHandler(store, nil)

	c, w := newCertCtx(t, http.MethodGet, "/?kind=normal", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusOK, w.Code)
	body := decodeBody(t, w)
	assert.Equal(t, "1", body["total"].Num.String())
	assert.Equal(t, "1", body["page"].Num.String())
	assert.Equal(t, "20", body["limit"].Num.String())
	assert.Len(t, body["data"].Arr, 1)

	// kind required -> 400 validation.
	c, w = newCertCtx(t, http.MethodGet, "/", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	c, w = newCertCtx(t, http.MethodGet, "/?kind=bad", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// (page-1)*limit would overflow int64 -> 400 validation, not a wrapped
	// negative offset into the store.
	c, w = newCertCtx(t, http.MethodGet, "/?kind=normal&page=92233720368547760&limit=100", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "validation", decodeBody(t, w)["code"].Str)

	// Vaccination missing/cross-clinic or failing the current relation
	// scope -> 404.
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{}, nil)
	c, w = newCertCtx(t, http.MethodGet, "/?kind=normal", nil, gin.Params{{Key: "id", Value: "999"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "not_found", decodeBody(t, w)["code"].Str)

	// No series -> 200 empty page with data: [] (not null).
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{vacc: vacc}, nil)
	c, w = newCertCtx(t, http.MethodGet, "/?kind=normal", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	assert.Equal(t, http.StatusOK, w.Code)
	body = decodeBody(t, w)
	assert.Equal(t, "0", body["total"].Num.String())
	assert.Empty(t, body["data"].Arr)
}

// Validity must be projected from the vaccination row read inside the
// ListEditions transaction snapshot, not the earlier existence gate: the
// fake's FindVaccinationSource still reports version 2 while the
// transaction-scoped row reports version 5, so the stored edition
// (source_version 2) must project source_changed, not current.
func TestListVaccinationCertificateEditions_TxSnapshotValidity(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	store := &fakeCertificateStore{
		ed: ed, series: series, seriesFound: true,
		vacc:       &model.Vaccination{ID: 100, ClinicID: 1, Version: 2},
		listVacc:   &model.Vaccination{ID: 100, ClinicID: 1, Version: 5},
		eds:        []model.VaccinationCertificateEdition{*ed},
		total:      1,
		maxEdition: 1,
	}
	h := NewVaccinationCertificateHandler(store, nil)
	c, w := newCertCtx(t, http.MethodGet, "/?kind=normal", nil, gin.Params{{Key: "id", Value: "100"}})
	h.ListVaccinationCertificateEditions(c)
	require.Equal(t, http.StatusOK, w.Code)
	body := decodeBody(t, w)
	require.Len(t, body["data"].Arr, 1)
	assert.Equal(t, "source_changed", body["data"].Arr[0].Obj["validity"].Str)
}

func TestPreviewSavedVaccinationCertificate_BodyValidation(t *testing.T) {
	h := NewVaccinationCertificateHandler(&fakeCertificateStore{}, nil)
	params := gin.Params{{Key: "edition_id", Value: "20"}}
	for _, tc := range []struct {
		name string
		body []byte
	}{
		{"malformed", []byte(`{`)},
		{"unknown field", []byte(`{"layout":"portrait","x":1}`)},
		{"missing layout", []byte(`{}`)},
		{"invalid enum", []byte(`{"layout":"bogus"}`)},
		{"trailing value", []byte(`{"layout":"portrait"} {}`)},
	} {
		c, w := newCertCtx(t, http.MethodPost, "/", tc.body, params)
		h.PreviewSavedVaccinationCertificate(c)
		assert.Equal(t, http.StatusBadRequest, w.Code, tc.name)
		assert.Equal(t, "validation", decodeBody(t, w)["code"].Str, tc.name)
	}
}

// An oversized request body cut by a server-imposed MaxBytesReader limit maps
// to 413 payload-too-large, not the flat 400 validation shape.
func TestPreviewSavedVaccinationCertificate_BodyTooLarge(t *testing.T) {
	h := NewVaccinationCertificateHandler(&fakeCertificateStore{}, nil)
	params := gin.Params{{Key: "edition_id", Value: "20"}}
	c, w := newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	c.Request.Body = http.MaxBytesReader(w, c.Request.Body, 8)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.NotEmpty(t, decodeBody(t, w)["error"].Str)
}

func TestPreviewSavedVaccinationCertificate_RenderPath(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	h := NewVaccinationCertificateHandler(&fakeCertificateStore{
		ed: ed, series: series, vacc: &model.Vaccination{ID: 100, ClinicID: 1, Version: 2},
		validitySeq: []string{"current"},
	}, nil)
	params := gin.Params{{Key: "edition_id", Value: "20"}}

	// Pure re-render is disabled in production while the font asset is
	// non-embedded: the no-artifact normal branch answers 503 render_failed
	// instead of a PDF.
	c, w := newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "render_failed", decodeBody(t, w)["code"].Str)
}

func TestPreviewSavedVaccinationCertificate_ErrorMapping(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	vacc := &model.Vaccination{ID: 100, ClinicID: 1, Version: 2}
	params := gin.Params{{Key: "edition_id", Value: "20"}}
	portrait := []byte(`{"layout":"portrait"}`)

	// Edition exists but the current source relation check fails -> 404.
	h := NewVaccinationCertificateHandler(&fakeCertificateStore{ed: ed, series: series}, nil)
	c, w := newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "not_found", decodeBody(t, w)["code"].Str)

	// Layout kind mismatch (normal edition + rabies layout) -> 422 validation.
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"current"}}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"rabies"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "validation", decodeBody(t, w)["code"].Str)

	// Non-current edition -> 409 dependency-changed.
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"superseded"}}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "dependency-changed", decodeBody(t, w)["code"].Str)

	// Stale between the artifact read and the second validity read -> 409.
	stalePDF := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 16)...)
	staleSum := sha256.Sum256(stalePDF)
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{
		ed: ed, series: series, vacc: vacc, validitySeq: []string{"current", "superseded"},
		artifact: &model.VaccinationCertificateArtifact{
			ID: 30, ClinicID: 1, EditionID: 20, Layout: "portrait",
			TemplateVersion: "normal-portrait/1", RendererVersion: "ekarte-certificate-pdf/1",
			ObjectKey: "k", ContentSHA256: staleSum[:],
		},
		artifactFound: true,
	}, &fakeArtifactStore{data: stalePDF})
	c, w = newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "dependency-changed", decodeBody(t, w)["code"].Str)

	// Unknown schema_version in stored snapshot -> 409 dependency-changed.
	badEd := *ed
	badEd.Snapshot = json.RawMessage(`{"schema_version":2}`)
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: &badEd, series: series, vacc: vacc}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "dependency-changed", decodeBody(t, w)["code"].Str)

	// Snapshot source.version disagreeing with the edition -> 503 result-unavailable.
	staleAnchor := *ed
	staleAnchor.Snapshot = json.RawMessage(bytes.Replace(
		[]byte(testCertificateSnapshotJSON),
		[]byte(`"version": 2`),
		[]byte(`"version": 9`), 1))
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: &staleAnchor, series: series, vacc: vacc, validitySeq: []string{"current"}}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "result-unavailable", decodeBody(t, w)["code"].Str)

	// Rabies edition with no artifact -> 503 artifact_unavailable.
	rabiesSeries := &model.VaccinationCertificateSeries{ID: 11, ClinicID: 1, VaccinationID: 100, CertificateKind: "rabies"}
	rabiesEd := &model.VaccinationCertificateEdition{
		ID: 21, ClinicID: 1, SeriesID: 11, Edition: 1, SourceVaccinationVersion: 2,
		Snapshot: json.RawMessage(testCertificateRabiesSnapshot(100, 2)),
	}
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: rabiesEd, series: rabiesSeries, vacc: vacc, validitySeq: []string{"current"}}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"rabies"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "artifact_unavailable", decodeBody(t, w)["code"].Str)

	// Manifest version drift -> 409 dependency-changed.
	drifted := *ed
	drifted.Snapshot = json.RawMessage(bytes.Replace(
		[]byte(testCertificateSnapshotJSON),
		[]byte(`"renderer_version": "ekarte-certificate-pdf/1"`),
		[]byte(`"renderer_version": "ekarte-certificate-pdf/0"`), 1))
	h = NewVaccinationCertificateHandler(&fakeCertificateStore{ed: &drifted, series: series, vacc: vacc, validitySeq: []string{"current"}}, nil)
	c, w = newCertCtx(t, http.MethodPost, "/", portrait, params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "dependency-changed", decodeBody(t, w)["code"].Str)
}

func TestPreviewSavedVaccinationCertificate_ArtifactPath(t *testing.T) {
	ed, series := testCertificateEditionFixture()
	vacc := &model.Vaccination{ID: 100, ClinicID: 1, Version: 2}
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 16)...)
	sum := sha256.Sum256(pdf)
	// The artifact row's versions must equal the stored snapshot manifest.
	artifact := &model.VaccinationCertificateArtifact{
		ID: 30, ClinicID: 1, EditionID: 20, Layout: "portrait",
		TemplateVersion: "normal-portrait/1", RendererVersion: "ekarte-certificate-pdf/1",
		ObjectKey: "secret/key", ContentSHA256: sum[:],
	}
	h := NewVaccinationCertificateHandler(
		&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"current"}, artifact: artifact, artifactFound: true},
		&fakeArtifactStore{data: pdf})
	params := gin.Params{{Key: "edition_id", Value: "20"}}

	c, w := newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	body := decodeBody(t, w)
	assert.Equal(t, "normal-portrait/1", body["template_version"].Str)
	assert.Equal(t, "ekarte-certificate-pdf/1", body["renderer_version"].Str)

	// Artifact versions disagreeing with the stored manifest -> 503
	// artifact_unavailable (no re-render, no fallback).
	mismatched := *artifact
	mismatched.RendererVersion = "writer/9"
	h = NewVaccinationCertificateHandler(
		&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"current"}, artifact: &mismatched, artifactFound: true},
		&fakeArtifactStore{data: pdf})
	c, w = newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "artifact_unavailable", decodeBody(t, w)["code"].Str)

	// Digest mismatch -> artifact_unavailable (never re-render on artifact row).
	artifact.ContentSHA256 = make([]byte, 32)
	h = NewVaccinationCertificateHandler(
		&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"current"}, artifact: artifact, artifactFound: true},
		&fakeArtifactStore{data: pdf})
	c, w = newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "artifact_unavailable", decodeBody(t, w)["code"].Str)

	// Nil store -> artifact_unavailable.
	artifact.ContentSHA256 = sum[:]
	h = NewVaccinationCertificateHandler(
		&fakeCertificateStore{ed: ed, series: series, vacc: vacc, validitySeq: []string{"current"}, artifact: artifact, artifactFound: true},
		nil)
	c, w = newCertCtx(t, http.MethodPost, "/", []byte(`{"layout":"portrait"}`), params)
	h.PreviewSavedVaccinationCertificate(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// ─── Clinic isolation (testdb) ───────────────────────────────────────────────

func setupCertificateIsolationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.Vaccine{},
		&model.Staff{}, &model.StaffClinicAssignment{}, &model.Vaccination{},
		&model.VaccinationCertificateSeries{},
		&model.VaccinationCertificateEdition{},
		&model.VaccinationCertificateArtifact{},
	))
	db.Exec("TRUNCATE TABLE vaccination_certificate_artifacts CASCADE")
	db.Exec("TRUNCATE TABLE vaccination_certificate_editions CASCADE")
	db.Exec("TRUNCATE TABLE vaccination_certificate_series CASCADE")
	db.Exec("TRUNCATE TABLE vaccinations CASCADE")
	return db
}

// TestVaccinationCertificateReader_ClinicIsolation follows
// vaccination_clinic_isolation_test.go patterns: same-id rows in another
// clinic must produce the identical not_found, never an existence leak.
func TestVaccinationCertificateReader_ClinicIsolation(t *testing.T) {
	db := setupCertificateIsolationDB(t)
	reader := NewVaccinationCertificateReader(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	owner := makeTestOwner(t, db, clinicA, "cert owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "cert pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "cert vaccine")
	doctor := makeDoctor(t, db, clinicA, "cert doctor")
	vacc := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, VaccineID: vaccine.ID,
		DoctorID: &doctor.ID, Date: time.Now(), Version: 1,
	}
	require.NoError(t, db.WithContext(ctx).Create(vacc).Error)

	series := &model.VaccinationCertificateSeries{
		ClinicID: clinicA, VaccinationID: vacc.ID, CertificateKind: "normal",
	}
	require.NoError(t, db.WithContext(ctx).Create(series).Error)
	snapshot := `{"schema_version":1,"kind":"normal","source":{"vaccination_id":` +
		strconv.FormatUint(vacc.ID, 10) + `,"version":1,"vaccine_id":7,` +
		`"vaccine_name":"v","administered_on":"2026-10-01","next_scheduled_on":null,` +
		`"lots":[null,null,null,null],"doctor_name":null},` +
		`"owner":{"name":null,"name_kana":null,"postal_code":null,"address":null,"phone":null},` +
		`"pet":{"name":null,"species":null,"breed":null,"sex":null,"coat_color":null,"age_text":null,"birth_date":null},` +
		`"clinic":{"name":null,"director_name":null,"postal_code":null,"address":null,"phone":null,"fax":null,"email":null,"url":null,"veterinarian_association":null,"designated_veterinarian":null,"veterinarian_name":null},` +
		`"display_issue_date":"2026-10-20",` +
		`"inputs":{"certificate_number":null,"injection_tag_number":null,"dog_registration_tag_number":null,"distinguishing_features":null,"injection_amount_text":null,"body_size":null,"registration_fee":null,"injection_fee":null,"tag_fee":null,"total":null},` +
		`"eligibility":{"rabies_eligible":null,"vaccine_id":1,"vaccine_version":null,"provenance":"master_snapshot"},` +
		`"render_manifest":{"renderer_version":"r","font_asset_version":"f","templates":{"portrait":"p","landscape":"l"}}}`
	ed := &model.VaccinationCertificateEdition{
		ClinicID: clinicA, SeriesID: series.ID, Edition: 1,
		SourceVaccinationVersion: 1, Snapshot: json.RawMessage(snapshot),
		IssuedBy: doctor.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(ed).Error)

	// Owning clinic reads it.
	got, gotSeries, err := reader.FindEdition(ctx, clinicA, ed.ID)
	require.NoError(t, err)
	assert.Equal(t, ed.ID, got.ID)
	assert.Equal(t, series.ID, gotSeries.ID)
	validity, err := reader.EditionValidity(ctx, clinicA, got, gotSeries)
	require.NoError(t, err)
	assert.Equal(t, "current", validity)

	// Foreign clinic: identical not_found for edition, series and vaccination.
	_, _, err = reader.FindEdition(ctx, clinicB, ed.ID)
	assert.Equal(t, "not_found", certErrCode(t, err))
	_, err = reader.FindVaccinationSource(ctx, clinicB, vacc.ID)
	assert.Equal(t, "not_found", certErrCode(t, err))
	_, found, err := reader.FindSeries(ctx, clinicB, vacc.ID, "normal")
	require.NoError(t, err)
	assert.False(t, found)

	// Soft-deleted vaccination remains readable (Unscoped) as source_deleted.
	require.NoError(t, db.WithContext(ctx).Delete(vacc).Error)
	validity, err = reader.EditionValidity(ctx, clinicA, got, gotSeries)
	require.NoError(t, err)
	assert.Equal(t, "source_deleted", validity)
}

// TestVaccinationCertificateReader_EditionValiditySnapshot pins that the
// scoped vaccination read and the newer-edition count share one snapshot:
// a bumped source version alongside a newer edition must project
// source_changed (priority over superseded), not a mixed-read superseded.
func TestVaccinationCertificateReader_EditionValiditySnapshot(t *testing.T) {
	db := setupCertificateIsolationDB(t)
	reader := NewVaccinationCertificateReader(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ensureVaccinationTestClinics(t, db, clinicA)

	owner := makeTestOwner(t, db, clinicA, "validity owner")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "validity pet")
	vaccine := makeVaccineMaster(t, db, clinicA, "validity vaccine")
	doctor := makeDoctor(t, db, clinicA, "validity doctor")
	vacc := &model.Vaccination{
		ClinicID: clinicA, PetID: &pet.ID, VaccineID: vaccine.ID,
		DoctorID: &doctor.ID, Date: time.Now(), Version: 3,
	}
	require.NoError(t, db.WithContext(ctx).Create(vacc).Error)

	series := &model.VaccinationCertificateSeries{
		ClinicID: clinicA, VaccinationID: vacc.ID, CertificateKind: "normal",
	}
	require.NoError(t, db.WithContext(ctx).Create(series).Error)
	ed1 := &model.VaccinationCertificateEdition{
		ClinicID: clinicA, SeriesID: series.ID, Edition: 1,
		SourceVaccinationVersion: 3, Snapshot: json.RawMessage(`{}`),
		IssuedBy: doctor.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(ed1).Error)
	ed2 := &model.VaccinationCertificateEdition{
		ClinicID: clinicA, SeriesID: series.ID, Edition: 2,
		SourceVaccinationVersion: 5, Snapshot: json.RawMessage(`{}`),
		IssuedBy: doctor.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(ed2).Error)
	require.NoError(t, db.WithContext(ctx).Model(vacc).Update("version", 5).Error)

	// One consistent snapshot: current version 5 != stored source 3 and a
	// newer edition exists -> source_changed takes priority over superseded.
	validity, err := reader.EditionValidity(ctx, clinicA, ed1, series)
	require.NoError(t, err)
	assert.Equal(t, "source_changed", validity)
}
