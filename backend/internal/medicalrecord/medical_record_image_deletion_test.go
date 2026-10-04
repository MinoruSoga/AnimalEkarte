package medicalrecord

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type memImageRepo struct {
	images       map[uint64]*model.MedicalRecordImage
	stagedDelete map[uint64]struct{}
	nextID       uint64
}

func (r *memImageRepo) FindByMedicalRecordID(context.Context, uint64, uint64) ([]model.MedicalRecordImage, error) {
	out := make([]model.MedicalRecordImage, 0, len(r.images))
	for id, image := range r.images {
		if _, gone := r.stagedDelete[id]; gone {
			continue
		}
		out = append(out, *image)
	}
	return out, nil
}

func (r *memImageRepo) Create(_ context.Context, image *model.MedicalRecordImage) error {
	r.nextID++
	image.ID = r.nextID
	copied := *image
	r.images[image.ID] = &copied
	return nil
}

func (r *memImageRepo) Delete(_ context.Context, _, id uint64) error {
	if _, ok := r.images[id]; !ok {
		return apperrors.WrapNotFound("medical_record_image", "missing")
	}
	r.stagedDelete[id] = struct{}{}
	return nil
}

func (r *memImageRepo) FindByID(_ context.Context, _, id uint64) (*model.MedicalRecordImage, error) {
	if _, gone := r.stagedDelete[id]; gone {
		return nil, apperrors.WrapNotFound("medical_record_image", "missing")
	}
	image, ok := r.images[id]
	if !ok {
		return nil, apperrors.WrapNotFound("medical_record_image", "missing")
	}
	copied := *image
	return &copied, nil
}

func (r *memImageRepo) CountLiveByStorageKey(_ context.Context, _ uint64, key string) (int64, error) {
	var count int64
	for id, image := range r.images {
		if _, gone := r.stagedDelete[id]; gone {
			continue
		}
		if medicalRecordImageStoredValueUsesKey(image.ImageURL, key) || medicalRecordImageStoredValueUsesKey(image.ThumbnailURL, key) {
			count++
		}
	}
	return count, nil
}

func (r *memImageRepo) commit() {
	for id := range r.stagedDelete {
		delete(r.images, id)
	}
	r.stagedDelete = map[uint64]struct{}{}
}

func (r *memImageRepo) rollback() {
	r.stagedDelete = map[uint64]struct{}{}
}

type memDeletionStore struct {
	committed map[uint64]medicalRecordImageDeletion
	staged    map[uint64]medicalRecordImageDeletion
	inserts   int
}

func (s *memDeletionStore) Insert(_ context.Context, row medicalRecordImageDeletion) error {
	if _, ok := s.committed[row.ImageID]; ok {
		return errMedicalRecordImageDeletionExists
	}
	if _, ok := s.staged[row.ImageID]; ok {
		return errMedicalRecordImageDeletionExists
	}
	s.staged[row.ImageID] = row
	s.inserts++
	return nil
}

func (s *memDeletionStore) FindByImageID(_ context.Context, clinicID, imageID uint64) (*medicalRecordImageDeletion, error) {
	row, ok := s.committed[imageID]
	if !ok || row.ClinicID != clinicID {
		return nil, nil
	}
	copied := row
	return &copied, nil
}

func (s *memDeletionStore) BlockedImageIDs(_ context.Context, clinicID uint64, imageIDs []uint64) (map[uint64]bool, error) {
	blocked := make(map[uint64]bool, len(imageIDs))
	for _, id := range imageIDs {
		if row, ok := s.committed[id]; ok && row.ClinicID == clinicID {
			blocked[id] = true
		}
	}
	return blocked, nil
}

func (s *memDeletionStore) UpdateStorageResult(_ context.Context, clinicID, imageID uint64, status, lastError string) error {
	row, ok := s.committed[imageID]
	if !ok || row.ClinicID != clinicID {
		return errors.New("medical record image deletion row is missing")
	}
	row.StorageStatus = status
	row.LastError = lastError
	row.AttemptCount++
	s.committed[imageID] = row
	return nil
}

func (s *memDeletionStore) commit() {
	for id, row := range s.staged {
		s.committed[id] = row
	}
	s.staged = map[uint64]medicalRecordImageDeletion{}
}

func (s *memDeletionStore) rollback() {
	s.inserts -= len(s.staged)
	s.staged = map[uint64]medicalRecordImageDeletion{}
}

type imageDeleteTransactor struct {
	repo  *memImageRepo
	store *memDeletionStore
}

func (t *imageDeleteTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if err := fn(ctx); err != nil {
		t.repo.rollback()
		t.store.rollback()
		return err
	}
	t.repo.commit()
	t.store.commit()
	return nil
}

type scriptImageAudit struct {
	err     error
	entries int
	last    *AuditEntry
}

func (a *scriptImageAudit) LogEntryTx(_ context.Context, entry *AuditEntry) error {
	if a.err != nil {
		return a.err
	}
	a.entries++
	a.last = entry
	return nil
}

type memObjectStore struct {
	err     error
	deleted []string
}

func (o *memObjectStore) Delete(_ context.Context, key string) error {
	if o.err != nil {
		return o.err
	}
	o.deleted = append(o.deleted, key)
	return nil
}

type imageDeleteHarness struct {
	repo    *memImageRepo
	store   *memDeletionStore
	medRec  *mockMedicalRecordRepository
	audit   *scriptImageAudit
	objects *memObjectStore
	svc     *medicalRecordImageService
}

func newImageDeleteHarness(images ...model.MedicalRecordImage) *imageDeleteHarness {
	repo := &memImageRepo{
		images:       map[uint64]*model.MedicalRecordImage{},
		stagedDelete: map[uint64]struct{}{},
	}
	for i := range images {
		image := images[i]
		repo.images[image.ID] = &image
		if image.ID > repo.nextID {
			repo.nextID = image.ID
		}
	}
	store := &memDeletionStore{
		committed: map[uint64]medicalRecordImageDeletion{},
		staged:    map[uint64]medicalRecordImageDeletion{},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	audit := &scriptImageAudit{}
	svc := NewMedicalRecordImageService(repo, medRec, &imageDeleteTransactor{repo: repo, store: store}).(*medicalRecordImageService)
	svc.auditTx = audit
	svc.deletions = store
	return &imageDeleteHarness{repo: repo, store: store, medRec: medRec, audit: audit, objects: &memObjectStore{}, svc: svc}
}

func (h *imageDeleteHarness) delete(imageID uint64) error {
	return h.svc.Delete(context.Background(), 1, 5, imageID, 7, h.objects)
}

func TestMedicalRecordImageDelete_StorageFailureDoesNotResumeSigning(t *testing.T) {
	const key = "medical-records/5/original.png"
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key})
	h.objects.err = errors.New("storage unavailable")

	// storage 削除失敗は caller へ error で返る（204 で握り潰すと再試行経路が途切れる）
	err := h.delete(4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage unavailable")

	_, err = h.repo.FindByID(context.Background(), 1, 4)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	row := h.store.committed[4]
	assert.Equal(t, medicalRecordImageDeletionPending, row.StorageStatus)
	assert.Contains(t, row.LastError, "storage unavailable")
	assert.Empty(t, h.objects.deleted)
	blocked, err := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)

	h.objects.err = nil
	require.NoError(t, h.delete(4))
	assert.Equal(t, []string{key}, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
	assert.Equal(t, 1, h.store.inserts)
	blocked, err = h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)
}

// TestMedicalRecordImageDelete_StorageRetryAllowedAfterFinalization は確定済みカルテでも
// 認可済み台帳行の物理削除リトライが拒否されないことを検証する。
func TestMedicalRecordImageDelete_StorageRetryAllowedAfterFinalization(t *testing.T) {
	const key = "medical-records/5/original.png"
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key})
	h.objects.err = errors.New("storage unavailable")

	// draft 中の初回削除: storage 失敗で台帳は pending_storage のまま
	require.Error(t, h.delete(4))
	require.Equal(t, medicalRecordImageDeletionPending, h.store.committed[4].StorageStatus)

	// カルテを確定させてから再 DELETE（= 台帳経由の物理削除リトライ）
	h.medRec.findByIDFn = func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusFinalized}, nil
	}
	h.objects.err = nil
	require.NoError(t, h.delete(4))

	assert.Equal(t, []string{key}, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
	assert.Equal(t, 1, h.store.inserts)
}

func TestMedicalRecordImageDelete_BlocksReregisteredKeyButAllowsOtherLiveImage(t *testing.T) {
	const key = "medical-records/5/shared.png"
	h := newImageDeleteHarness(
		model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key},
		model.MedicalRecordImage{ID: 8, MedicalRecordID: 5, ImageURL: key},
	)

	require.NoError(t, h.delete(4))

	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionShared, h.store.committed[4].StorageStatus)
	blocked, err := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)
	blocked, err = h.svc.SigningBlocked(context.Background(), 1, 8)
	require.NoError(t, err)
	assert.False(t, blocked)

	require.NoError(t, h.delete(8))
	assert.Equal(t, []string{key}, h.objects.deleted)
	h.repo.images[9] = &model.MedicalRecordImage{ID: 9, MedicalRecordID: 5, ImageURL: key}
	blocked, err = h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)
	blocked, err = h.svc.SigningBlocked(context.Background(), 1, 9)
	require.NoError(t, err)
	assert.False(t, blocked)
}

func TestMedicalRecordImageDelete_EquivalentURLKeepsSharedObject(t *testing.T) {
	const key = "medical-records/5/shared.png"
	h := newImageDeleteHarness(
		model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key},
		model.MedicalRecordImage{ID: 8, MedicalRecordID: 5, ImageURL: "https://cdn.example/medical-records/5/shared.png?X-Amz-Signature=abc"},
		model.MedicalRecordImage{ID: 10, MedicalRecordID: 5, ThumbnailURL: "/uploads/medical-records/5/shared.png"},
	)

	require.NoError(t, h.delete(4))

	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionShared, h.store.committed[4].StorageStatus)
	for _, imageID := range []uint64{8, 10} {
		blocked, err := h.svc.SigningBlocked(context.Background(), 1, imageID)
		require.NoError(t, err)
		assert.False(t, blocked, "image %d", imageID)
	}
}

func TestMedicalRecordImageDelete_LongerKeyDoesNotKeepObject(t *testing.T) {
	const key = "medical-records/5/shared.png"
	h := newImageDeleteHarness(
		model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key},
		model.MedicalRecordImage{ID: 8, MedicalRecordID: 5, ImageURL: "https://cdn.example/medical-records/5/shared.png.bak"},
	)

	require.NoError(t, h.delete(4))

	assert.Equal(t, []string{key}, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
}

func TestMedicalRecordImageDelete_DoesNotDeleteKeyOutsideRecord(t *testing.T) {
	const key = "medical-records/5/../../etc/passwd"
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key})

	// 台帳外の key 拒否は error で surface し、台帳は pending_storage に残る
	err := h.delete(4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "outside the record")

	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionPending, h.store.committed[4].StorageStatus)
	assert.Contains(t, h.store.committed[4].LastError, "outside the record")
	blocked, err := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestMedicalRecordImageDelete_LiveRegistrationStaysSignable(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/live.png"})

	blocked, err := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.False(t, blocked)
	_, err = h.repo.FindByID(context.Background(), 1, 4)
	require.NoError(t, err)
}

func TestMedicalRecordImageDelete_AuditFailureRollsBack(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/live.png"})
	h.audit.err = errors.New("audit sink down")

	err := h.delete(4)

	require.Error(t, err)
	_, findErr := h.repo.FindByID(context.Background(), 1, 4)
	require.NoError(t, findErr)
	assert.Empty(t, h.store.committed)
	assert.Empty(t, h.objects.deleted)
	blocked, blockErr := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, blockErr)
	assert.False(t, blocked)
}

func TestMedicalRecordImageDelete_FinalizedRecordDoesNotDelete(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/live.png"})
	h.medRec.findByIDFn = func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusFinalized}, nil
	}

	err := h.delete(4)

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	_, findErr := h.repo.FindByID(context.Background(), 1, 4)
	require.NoError(t, findErr)
	assert.Empty(t, h.store.committed)
	assert.Empty(t, h.objects.deleted)
}

func TestMedicalRecordImageDelete_NilAuditFailsClosed(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/live.png"})
	h.svc.auditTx = nil

	err := h.delete(4)

	require.Error(t, err)
	_, findErr := h.repo.FindByID(context.Background(), 1, 4)
	require.NoError(t, findErr)
	assert.Empty(t, h.objects.deleted)
}

func TestMedicalRecordImageDelete_ExternalURLHasNoObjectDeleteAndStopsSigning(t *testing.T) {
	const external = "https://cdn.example/photo.jpg"
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: external})

	require.NoError(t, h.delete(4))

	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
	assert.Empty(t, h.store.committed[4].ImageObjectKey)
	require.NotNil(t, h.audit.last)
	oldValue, ok := h.audit.last.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, external, oldValue["image_url"])
	blocked, err := h.svc.SigningBlocked(context.Background(), 1, 4)
	require.NoError(t, err)
	assert.True(t, blocked)
}

func TestMedicalRecordImageCreate_RechecksDraftAndClinic(t *testing.T) {
	var gotClinic, gotRecord uint64
	repo := &mockMedicalRecordImageRepository{
		createFn: func(context.Context, *model.MedicalRecordImage) error {
			t.Fatal("finalized record must not persist an image")
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			gotClinic = clinicID
			gotRecord = id
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewMedicalRecordImageService(repo, medRec, &mockCheckupTransactor{})

	_, err := svc.Create(context.Background(), 9, 8, &CreateMedicalRecordImageInput{
		ImageURL:  "medical-records/8/a.png",
		ImageType: model.MedicalImageTypeOther,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, uint64(9), gotClinic)
	assert.Equal(t, uint64(8), gotRecord)
}

func TestListMedicalRecordImages_DoesNotSignDeletedImageID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const stored = "https://cdn.example/already-stored.jpg"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(context.Context, string, time.Duration) (string, error) {
			t.Fatal("non-object stored URL must not reach the signer")
			return "", nil
		},
	}
	handler := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(context.Context, uint64, uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID: 4, MedicalRecordID: 5, ImageURL: stored,
				}}, nil
			},
			blockedImageIDsFn: func(context.Context, uint64, []uint64) (map[uint64]bool, error) {
				return map[uint64]bool{4: true}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	handler.ListMedicalRecordImages(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), stored)
	assert.Empty(t, uploader.signedURLCalls)
}

// TestListMedicalRecordImages_BlocksSigningAfterConcurrentDelete は署名発行と
// 削除 commit が交差しても一括突合で応答を落とすことを検証する（署名後チェック）。
func TestListMedicalRecordImages_BlocksSigningAfterConcurrentDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const signed = "https://signed.example/medical-records/5/uuid.png?sig=x"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, key string, _ time.Duration) (string, error) {
			return signed, nil
		},
	}
	handler := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(context.Context, uint64, uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{{
					ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/uuid.png",
				}}, nil
			},
			blockedImageIDsFn: func(context.Context, uint64, []uint64) (map[uint64]bool, error) {
				return map[uint64]bool{4: true}, nil
			},
		},
		uploader,
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	handler.ListMedicalRecordImages(c)

	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), signed)
}

func TestUploadMedicalRecordImage_SignFailureLeavesNeitherRowNorObject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageDeleteHarness()
	petID := uint64(3)
	h.medRec.findByIDFn = func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{
			ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft, PetID: &petID,
		}, nil
	}
	svc := NewMedicalRecordImageServiceWithRelationValidation(
		h.repo, h.medRec, livingPetFinder(), nil, nil, nil, h.svc.transactor,
	).(*medicalRecordImageService)
	svc.auditTx = h.audit
	svc.deletions = h.store

	uploader := &mockMedicalRecordImageUploader{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
			assert.True(t, len(key) > len("medical-records/5/"))
			return key, nil
		},
		getSignedURLFn: func(context.Context, string, time.Duration) (string, error) {
			return "", errors.New("presign failed")
		},
	}
	handler := newHandlerWithMedicalRecordImageSvc(ownedMedicalRecordImageGetter(), svc, uploader)
	body, contentType := buildImageMultipart(t, "photo.png", "image/png", uploadFixturePNG())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	c.Request.Header.Set("Content-Type", contentType)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	handler.UploadMedicalRecordImage(c)

	assert.NotEqual(t, http.StatusCreated, w.Code)
	assert.Empty(t, h.repo.images)
	require.Len(t, uploader.deleteCalls, 1)
	assert.Contains(t, uploader.deleteCalls[0], "medical-records/5/")
	require.Len(t, h.store.committed, 1)
	for imageID := range h.store.committed {
		blocked, err := svc.SigningBlocked(context.Background(), 1, imageID)
		require.NoError(t, err)
		assert.True(t, blocked)
	}
}
