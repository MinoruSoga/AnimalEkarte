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

func (r *memImageRepo) CountLiveByStorageKey(_ context.Context, key string) (int64, error) {
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
			return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
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

// TestMedicalRecordImageDelete_StorageRetryAllowedAfterFinalization は、親カルテが後から変わっても
// （EMR-495: 旧 finalized 判定は廃止）認可済み台帳行の物理削除リトライが拒否されないことを検証する。
func TestMedicalRecordImageDelete_StorageRetryAllowedAfterFinalization(t *testing.T) {
	const key = "medical-records/5/original.png"
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: key})
	h.objects.err = errors.New("storage unavailable")

	// draft 中の初回削除: storage 失敗で台帳は pending_storage のまま
	require.Error(t, h.delete(4))
	require.Equal(t, medicalRecordImageDeletionPending, h.store.committed[4].StorageStatus)

	// 親カルテを再取得させてから再 DELETE（= 台帳経由の物理削除リトライ）
	h.medRec.findByIDFn = func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
		return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
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

// EMR-495 P1: 旧 RechecksDraftAndClinic（確定済みなら Conflict で画像を作らない）は廃止した。
// 親行ロックが clinic スコープで取得され、画像が作成されることだけを検証する。
func TestMedicalRecordImageCreate_RechecksClinicParentLock(t *testing.T) {
	var gotClinic, gotRecord uint64
	created := false
	repo := &mockMedicalRecordImageRepository{
		createFn: func(context.Context, *model.MedicalRecordImage) error {
			created = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			gotClinic = clinicID
			gotRecord = id
			return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
		},
	}
	svc := NewMedicalRecordImageService(repo, medRec, &mockCheckupTransactor{})

	_, err := svc.Create(context.Background(), 9, 8, &CreateMedicalRecordImageInput{
		ImageURL:  "medical-records/8/a.png",
		ImageType: model.MedicalImageTypeOther,
	})

	require.NoError(t, err)
	assert.True(t, created)
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
			ID: id, ClinicID: clinicID, PetID: &petID,
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

// signingObjectDeleter は GetSignedURL を持つ object deleter fake。
// trusted storage origin の host を署名 URL から自己申告する本番 uploader を模す。
type signingObjectDeleter struct {
	memObjectStore
	host string
}

func (o *signingObjectDeleter) GetSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://" + o.host + "/" + key + "?sig=x", nil
}

// TestAuditSafeMedicalRecordImageURL は監査永続化前の URL sanitizer を検証する。
// presigned query（=認証情報）と HTTP userinfo は audit_logs に残してはいけない。
func TestAuditSafeMedicalRecordImageURL(t *testing.T) {
	tests := []struct {
		name   string
		stored string
		want   string
	}{
		{
			name:   "presigned URL strips query and fragment",
			stored: "https://cdn.example/medical-records/5/a.png?X-Amz-Signature=abc&X-Amz-Credential=x#frag",
			want:   "https://cdn.example/medical-records/5/a.png",
		},
		{
			name:   "userinfo URL strips credentials",
			stored: "https://user:password@cdn.example/medical-records/5/a.png?sig=1",
			want:   "https://cdn.example/medical-records/5/a.png",
		},
		{
			name:   "plain object key is unchanged",
			stored: "medical-records/5/a.png",
			want:   "medical-records/5/a.png",
		},
		{
			name:   "uploads path is unchanged",
			stored: "/uploads/medical-records/5/a.png",
			want:   "/uploads/medical-records/5/a.png",
		},
		{
			name:   "unparseable value drops query only",
			stored: "weird%zz?token=1",
			want:   "weird%zz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auditSafeMedicalRecordImageURL(tt.stored)
			assert.Equal(t, tt.want, got)
			assert.NotContains(t, got, "password")
			assert.NotContains(t, got, "X-Amz-Signature")
		})
	}
}

// TestMedicalRecordImageDelete_AuditStripsUserinfoAndQuery は削除フロー経由で
// audit_logs に書かれる値に userinfo / 署名クエリが残らないことを検証する。
func TestMedicalRecordImageDelete_AuditStripsUserinfoAndQuery(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{
		ID: 4, MedicalRecordID: 5,
		ImageURL: "https://user:secret@cdn.example/photo.jpg?X-Amz-Signature=tok#frag",
	})

	require.NoError(t, h.delete(4))

	require.NotNil(t, h.audit.last)
	oldValue, ok := h.audit.last.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "https://cdn.example/photo.jpg", oldValue["image_url"])
}

// TestMedicalRecordImageDelete_RetryPreservesKeyResolutionFailure は key 解決に
// 失敗した pending_storage 行のリトライが storage_deleted に進まないことを検証する。
// 画像行は既に無く再解決もできないため、失敗を保持して返し続ける。
func TestMedicalRecordImageDelete_RetryPreservesKeyResolutionFailure(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/"})

	// 初回: key 解決失敗 → error + pending_storage（key 空で記録）
	require.Error(t, h.delete(4))
	row := h.store.committed[4]
	require.Equal(t, medicalRecordImageDeletionPending, row.StorageStatus)
	assert.Empty(t, row.ImageObjectKey)
	assert.Empty(t, row.ThumbnailObjectKey)
	assert.Contains(t, row.LastError, medicalRecordImageKeyUnresolvedMark)

	// リトライ: storage_deleted にせず pending を維持し、失敗を返す
	err := h.delete(4)
	require.Error(t, err)
	assert.Contains(t, err.Error(), medicalRecordImageKeyUnresolvedMark)
	row = h.store.committed[4]
	assert.Equal(t, medicalRecordImageDeletionPending, row.StorageStatus)
	assert.NotEqual(t, medicalRecordImageDeletionDeleted, row.StorageStatus)
	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, 1, h.store.inserts)
}

// TestMedicalRecordImageDelete_RetryPreservesPartialResolutionFailure は片方の URL だけ
// 解決できた部分失敗の行でも、リトライが「解決失敗」を一時失敗と混同しないことを検証する。
func TestMedicalRecordImageDelete_RetryPreservesPartialResolutionFailure(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{
		ID: 4, MedicalRecordID: 5,
		ImageURL:     "medical-records/5/img.png",
		ThumbnailURL: "medical-records/",
	})

	require.Error(t, h.delete(4))
	row := h.store.committed[4]
	require.Equal(t, medicalRecordImageDeletionPending, row.StorageStatus)
	assert.Equal(t, "medical-records/5/img.png", row.ImageObjectKey)
	assert.Empty(t, row.ThumbnailObjectKey)
	assert.Contains(t, row.LastError, medicalRecordImageKeyUnresolvedMark)

	// リトライ: 解決済み key を持っていても解決失敗マーカーの行は storage_deleted にせず、
	// オブジェクト削除にも進まない（orphan 側を黙って見捨てない）。
	err := h.delete(4)
	require.Error(t, err)
	row = h.store.committed[4]
	assert.Equal(t, medicalRecordImageDeletionPending, row.StorageStatus)
	assert.Empty(t, h.objects.deleted)
}

// TestMedicalRecordImageDelete_ExternalHostURLIsNotOwned は外部 host の URL が
// path に key 形を含んでも自ストレージの object key にならないことを検証する。
func TestMedicalRecordImageDelete_ExternalHostURLIsNotOwned(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{
		ID: 4, MedicalRecordID: 5,
		ImageURL: "https://evil.example/medical-records/5/x.png?sig=z",
	})
	objects := &signingObjectDeleter{host: "trusted.example"}

	require.NoError(t, h.svc.Delete(context.Background(), 1, 5, 4, 7, objects))

	assert.Empty(t, objects.deleted)
	row := h.store.committed[4]
	assert.Equal(t, medicalRecordImageDeletionDeleted, row.StorageStatus)
	assert.Empty(t, row.ImageObjectKey)
}

// TestMedicalRecordImageDelete_TrustedHostURLDeletesOwnedObject は uploader の
// storage origin と host が一致する URL だけが削除対象 key になることを検証する。
func TestMedicalRecordImageDelete_TrustedHostURLDeletesOwnedObject(t *testing.T) {
	const key = "medical-records/5/x.png"
	h := newImageDeleteHarness(model.MedicalRecordImage{
		ID: 4, MedicalRecordID: 5,
		ImageURL: "https://trusted.example/" + key + "?sig=z",
	})
	objects := &signingObjectDeleter{host: "trusted.example"}

	require.NoError(t, h.svc.Delete(context.Background(), 1, 5, 4, 7, objects))

	assert.Equal(t, []string{key}, objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
}

// TestMedicalRecordImageDelete_NoSignerKeepsHTTPNonOwned は署名能力を持たない
// deleter では HTTP(S) URL を一切 owned key にしない（fail-closed）ことを検証する。
func TestMedicalRecordImageDelete_NoSignerKeepsHTTPNonOwned(t *testing.T) {
	h := newImageDeleteHarness(model.MedicalRecordImage{
		ID: 4, MedicalRecordID: 5,
		ImageURL: "https://cdn.example/medical-records/5/x.png?sig=z",
	})

	require.NoError(t, h.delete(4))

	assert.Empty(t, h.objects.deleted)
	assert.Equal(t, medicalRecordImageDeletionDeleted, h.store.committed[4].StorageStatus)
}

// TestListMedicalRecordImages_UsesOneBatchedBlockedCheck は一覧の削除台帳突合が
// 画像数に依らず 1 クエリ（BlockedImageIDs 一括）で行われることを検証する（N+1 防止）。
func TestListMedicalRecordImages_UsesOneBatchedBlockedCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var blockedCalls int
	var gotIDs []uint64
	handler := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			listFn: func(context.Context, uint64, uint64) ([]model.MedicalRecordImage, error) {
				return []model.MedicalRecordImage{
					{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/a.png"},
					{ID: 6, MedicalRecordID: 5, ImageURL: "medical-records/5/b.png"},
					{ID: 8, MedicalRecordID: 5, ImageURL: "medical-records/5/c.png"},
				}, nil
			},
			blockedImageIDsFn: func(_ context.Context, _ uint64, ids []uint64) (map[uint64]bool, error) {
				blockedCalls++
				gotIDs = append([]uint64(nil), ids...)
				return map[uint64]bool{}, nil
			},
		},
		&mockMedicalRecordImageUploader{},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/medical-records/5/images", http.NoBody)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)

	handler.ListMedicalRecordImages(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, blockedCalls, "blocked check must be a single batched query, not per-image")
	assert.ElementsMatch(t, []uint64{4, 6, 8}, gotIDs)
}

// TestSignedMedicalRecordImageResponse_DropsSignedURLWhenDeletionCommitted は
// 署名発行→台帳確認の順序で、削除 commit が間に入った場合に発行済み URL を
// レスポンスに含めないことを検証する（単体署名経路の ordering 保証）。
func TestSignedMedicalRecordImageResponse_DropsSignedURLWhenDeletionCommitted(t *testing.T) {
	const signed = "https://signed.example/medical-records/5/u.png?sig=x"
	uploader := &mockMedicalRecordImageUploader{
		getSignedURLFn: func(_ context.Context, _ string, _ time.Duration) (string, error) {
			return signed, nil
		},
	}
	handler := newHandlerWithMedicalRecordImageSvc(
		ownedMedicalRecordImageGetter(),
		&mockMedicalRecordImageService{
			signingBlockedFn: func(context.Context, uint64, uint64) (bool, error) {
				return true, nil
			},
		},
		uploader,
	)

	resp, err := handler.signedMedicalRecordImageResponse(
		context.Background(), 1,
		&model.MedicalRecordImage{ID: 4, MedicalRecordID: 5, ImageURL: "medical-records/5/u.png"},
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errMedicalRecordImageSigningRetired))
	assert.Empty(t, resp.ImageURL, "blocked image must not leak the issued signed URL")
	assert.Empty(t, resp.ThumbnailURL)
	require.Len(t, uploader.signedURLCalls, 1, "sign runs first, then the ledger check drops the result")
}

// uploadSignFailureHarness は署名失敗→補償削除系の handler テスト共通セットアップ。
// svc は createFn で画像を返し、deleteFn / signingBlockedFn で補償結果を制御する。
func uploadSignFailureHandler(
	t *testing.T,
	deleteErr error,
	signingBlockedFn func(context.Context, uint64, uint64) (bool, error),
) (*MedicalRecordImageHandler, *gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := &mockMedicalRecordImageService{
		createFn: func(_ context.Context, _, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
			return &model.MedicalRecordImage{ID: 9, MedicalRecordID: medicalRecordID, ImageURL: input.ImageURL}, nil
		},
		deleteFn: func(context.Context, uint64, uint64, uint64, uint64, medicalRecordImageObjectDeleter) error {
			return deleteErr
		},
		signingBlockedFn: signingBlockedFn,
	}
	uploader := &mockMedicalRecordImageUploader{
		uploadFn: func(_ context.Context, key string, _ io.Reader, _ string) (string, error) {
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
	return handler, c, w
}

// TestUploadMedicalRecordImage_SignFailureCommittedRetirementResponse は、補償削除が
// error を返しても台帳 commit 済み（retirement 成立）の場合に「保存済み・一覧再確認」
// と誤認させる応答を返さないことを検証する。
func TestUploadMedicalRecordImage_SignFailureCommittedRetirementResponse(t *testing.T) {
	handler, c, w := uploadSignFailureHandler(t,
		errors.New("storage unreachable"),
		func(context.Context, uint64, uint64) (bool, error) { return true, nil },
	)

	handler.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "was retired")
	assert.Contains(t, w.Body.String(), "re-upload")
	assert.NotContains(t, w.Body.String(), "retry listing")
}

// TestUploadMedicalRecordImage_SignFailureImageRemainsResponse は、補償削除自体が
// 失敗して画像行が残る場合に「一覧確認（再 upload ではない）」を案内することを検証する。
func TestUploadMedicalRecordImage_SignFailureImageRemainsResponse(t *testing.T) {
	handler, c, w := uploadSignFailureHandler(t,
		errors.New("record finalized before cleanup"),
		func(context.Context, uint64, uint64) (bool, error) { return false, nil },
	)

	handler.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "was saved but could not be signed")
	assert.Contains(t, w.Body.String(), "retry listing images instead of re-uploading")
}

// TestUploadMedicalRecordImage_SignFailureUnknownOutcomeResponse は、補償削除の
// 成否自体が判定不能な場合にどちらとも断定しない応答を返すことを検証する。
func TestUploadMedicalRecordImage_SignFailureUnknownOutcomeResponse(t *testing.T) {
	handler, c, w := uploadSignFailureHandler(t,
		errors.New("storage unreachable"),
		func(context.Context, uint64, uint64) (bool, error) {
			return false, errors.New("ledger read failed")
		},
	)

	handler.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "outcome is uncertain")
	assert.NotContains(t, w.Body.String(), "retry listing images instead of re-uploading")
}
