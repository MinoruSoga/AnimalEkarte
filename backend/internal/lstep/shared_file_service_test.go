package lstep

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockSharedFileRepository struct {
	SharedFileRepository
	createFn      func(ctx context.Context, f *model.SharedFile) error
	findByIDFn    func(ctx context.Context, clinicID, id uint64) (*model.SharedFile, error)
	findAllFn     func(ctx context.Context, clinicID uint64) ([]*model.SharedFile, error)
	deleteFn      func(ctx context.Context, clinicID, id uint64) error
	findExpiredFn func(ctx context.Context, thresholdUnix int64) ([]*model.SharedFile, error)
}

type mockSharedFileOwnerRepository struct {
	findByIDFn func(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

func (m *mockSharedFileOwnerRepository) FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, ownerID)
	}
	return &model.Owner{ID: ownerID, ClinicID: clinicID}, nil
}

func (m *mockSharedFileRepository) Create(ctx context.Context, f *model.SharedFile) error {
	if m.createFn != nil {
		return m.createFn(ctx, f)
	}
	return nil
}

func (m *mockSharedFileRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.SharedFile, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
}

func (m *mockSharedFileRepository) FindAll(ctx context.Context, clinicID uint64) ([]*model.SharedFile, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []*model.SharedFile{}, nil
}

func (m *mockSharedFileRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockSharedFileRepository) FindExpired(ctx context.Context, thresholdUnix int64) ([]*model.SharedFile, error) {
	if m.findExpiredFn != nil {
		return m.findExpiredFn(ctx, thresholdUnix)
	}
	return []*model.SharedFile{}, nil
}

type mockFileStorage struct {
	uploadFn       func(ctx context.Context, key string, content io.Reader, contentType string) error
	getSignedURLFn func(ctx context.Context, key string, ttl time.Duration) (string, error)
	deleteFn       func(ctx context.Context, key string) error
}

func (m *mockFileStorage) Upload(ctx context.Context, key string, content io.Reader, contentType string) error {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, key, content, contentType)
	}
	return nil
}

func (m *mockFileStorage) GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if m.getSignedURLFn != nil {
		return m.getSignedURLFn(ctx, key, ttl)
	}
	return "http://dummy-url", nil
}

func (m *mockFileStorage) Delete(ctx context.Context, key string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

func TestSharedFileService_Upload(t *testing.T) {
	ctx := context.Background()

	t.Run("success with owner verification", func(t *testing.T) {
		ownerID := uint64(500)
		ownerRepo := &mockSharedFileOwnerRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Owner, error) {
				return &model.Owner{ID: id, ClinicID: clinicID}, nil
			},
		}
		repo := &mockSharedFileRepository{
			createFn: func(_ context.Context, f *model.SharedFile) error {
				f.ID = 100
				return nil
			},
		}
		storage := &mockFileStorage{}
		svc := NewSharedFileService(repo, ownerRepo, storage)
		input := &UploadSharedFileInput{
			Content:     bytes.NewBufferString("dummy"),
			FileName:    "test.txt",
			ContentType: "text/plain",
			FileType:    "text",
			FileSize:    5,
			Purpose:     "general",
			OwnerID:     &ownerID,
		}
		res, err := svc.Upload(ctx, 1, 10, input)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), res.ID)
	})

	t.Run("owner not found", func(t *testing.T) {
		ownerID := uint64(500)
		ownerRepo := &mockSharedFileOwnerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewSharedFileService(nil, ownerRepo, nil)
		input := &UploadSharedFileInput{
			OwnerID: &ownerID,
		}
		_, err := svc.Upload(ctx, 1, 10, input)
		assert.Error(t, err)
	})

	t.Run("storage upload error", func(t *testing.T) {
		ownerRepo := &mockSharedFileOwnerRepository{}
		storage := &mockFileStorage{
			uploadFn: func(_ context.Context, _ string, _ io.Reader, _ string) error {
				return errors.New("storage error")
			},
		}
		svc := NewSharedFileService(nil, ownerRepo, storage)
		input := &UploadSharedFileInput{
			FileName: "test.txt",
		}
		_, err := svc.Upload(ctx, 1, 10, input)
		assert.Error(t, err)
	})

	t.Run("db create error, rollback storage", func(t *testing.T) {
		ownerRepo := &mockSharedFileOwnerRepository{}
		storage := &mockFileStorage{
			deleteFn: func(_ context.Context, key string) error {
				assert.Contains(t, key, "shared/1/")
				return nil
			},
		}
		repo := &mockSharedFileRepository{
			createFn: func(_ context.Context, _ *model.SharedFile) error {
				return errors.New("db error")
			},
		}
		svc := NewSharedFileService(repo, ownerRepo, storage)
		input := &UploadSharedFileInput{
			FileName: "test.txt",
		}
		_, err := svc.Upload(ctx, 1, 10, input)
		assert.Error(t, err)
	})
}

func TestSharedFileService_GetSignedURL(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.SharedFile, error) {
				return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
			},
		}
		storage := &mockFileStorage{
			getSignedURLFn: func(_ context.Context, key string, ttl time.Duration) (string, error) {
				assert.Equal(t, "dummy-key", key)
				assert.Equal(t, 24*time.Hour, ttl)
				return "https://signed-url.com", nil
			},
		}
		svc := NewSharedFileService(repo, nil, storage)
		url, err := svc.GetSignedURL(ctx, 1, 100)
		assert.NoError(t, err)
		assert.Equal(t, "https://signed-url.com", url)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.SharedFile, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewSharedFileService(repo, nil, nil)
		_, err := svc.GetSignedURL(ctx, 1, 100)
		assert.Error(t, err)
	})
}

func TestSharedFileService_FindAll(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findAllFn: func(_ context.Context, clinicID uint64) ([]*model.SharedFile, error) {
				return []*model.SharedFile{{ID: 1, ClinicID: clinicID}}, nil
			},
		}
		svc := NewSharedFileService(repo, nil, nil)
		res, err := svc.FindAll(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findAllFn: func(_ context.Context, _ uint64) ([]*model.SharedFile, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewSharedFileService(repo, nil, nil)
		_, err := svc.FindAll(ctx, 1)
		assert.Error(t, err)
	})
}

func TestSharedFileService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.SharedFile, error) {
				return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
			},
			deleteFn: func(_ context.Context, clinicID, id uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(100), id)
				return nil
			},
		}
		storage := &mockFileStorage{
			deleteFn: func(_ context.Context, key string) error {
				assert.Equal(t, "dummy-key", key)
				return nil
			},
		}
		svc := NewSharedFileService(repo, nil, storage)
		err := svc.Delete(ctx, 1, 100)
		assert.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.SharedFile, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewSharedFileService(repo, nil, nil)
		err := svc.Delete(ctx, 1, 100)
		assert.Error(t, err)
	})

	t.Run("storage delete error", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.SharedFile, error) {
				return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
			},
		}
		storage := &mockFileStorage{
			deleteFn: func(_ context.Context, _ string) error {
				return errors.New("storage delete error")
			},
		}
		svc := NewSharedFileService(repo, nil, storage)
		err := svc.Delete(ctx, 1, 100)
		assert.Error(t, err)
	})
}

func TestSharedFileService_Upload_RollbackDeleteAlsoFails(t *testing.T) {
	ctx := context.Background()
	ownerRepo := &mockSharedFileOwnerRepository{}
	rollbackCalled := false
	storage := &mockFileStorage{
		deleteFn: func(_ context.Context, _ string) error {
			rollbackCalled = true
			return errors.New("rollback failed")
		},
	}
	repo := &mockSharedFileRepository{
		createFn: func(_ context.Context, _ *model.SharedFile) error {
			return errors.New("db error")
		},
	}
	svc := NewSharedFileService(repo, ownerRepo, storage)
	input := &UploadSharedFileInput{FileName: "test.txt"}

	_, err := svc.Upload(ctx, 1, 10, input)

	assert.Error(t, err)
	assert.True(t, rollbackCalled, "storage.Delete によるロールバックが試行されること")
}

func TestSharedFileService_GetSignedURL_StorageError(t *testing.T) {
	ctx := context.Background()
	repo := &mockSharedFileRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.SharedFile, error) {
			return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
		},
	}
	storage := &mockFileStorage{
		getSignedURLFn: func(_ context.Context, _ string, _ time.Duration) (string, error) {
			return "", errors.New("storage error")
		},
	}
	svc := NewSharedFileService(repo, nil, storage)

	_, err := svc.GetSignedURL(ctx, 1, 100)

	assert.Error(t, err)
}

func TestSharedFileService_Delete_RepoDeleteError(t *testing.T) {
	ctx := context.Background()
	repo := &mockSharedFileRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.SharedFile, error) {
			return &model.SharedFile{ID: id, ClinicID: clinicID, FileKey: "dummy-key"}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	storage := &mockFileStorage{}
	svc := NewSharedFileService(repo, nil, storage)

	err := svc.Delete(ctx, 1, 100)

	assert.Error(t, err)
}

func TestSharedFileService_CleanupExpired_RepoDeleteErrorAfterStorageSuccess(t *testing.T) {
	ctx := context.Background()
	repo := &mockSharedFileRepository{
		findExpiredFn: func(_ context.Context, _ int64) ([]*model.SharedFile, error) {
			return []*model.SharedFile{
				{ID: 10, ClinicID: 1, FileKey: "key10"},
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return errors.New("db error")
		},
	}
	storage := &mockFileStorage{}
	svc := NewSharedFileService(repo, nil, storage)

	err := svc.CleanupExpired(ctx)

	assert.Error(t, err)
}

func TestGenerateSharedFileKey(t *testing.T) {
	key1, err := generateSharedFileKey(1, "photo.jpg")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(key1, "shared/1/"))
	assert.True(t, strings.HasSuffix(key1, ".jpg"))

	key2, err := generateSharedFileKey(1, "photo.jpg")
	assert.NoError(t, err)
	assert.NotEqual(t, key1, key2, "各呼び出しでユニークなキーが生成されること")

	keyNoExt, err := generateSharedFileKey(2, "noext")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(keyNoExt, "shared/2/"))
	assert.False(t, strings.Contains(keyNoExt, "."))
}

func TestSharedFileService_CleanupExpired(t *testing.T) {
	ctx := context.Background()

	t.Run("success cleanup", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findExpiredFn: func(_ context.Context, _ int64) ([]*model.SharedFile, error) {
				return []*model.SharedFile{
					{ID: 10, ClinicID: 1, FileKey: "key10"},
					{ID: 20, ClinicID: 1, FileKey: "key20"},
				}, nil
			},
			deleteFn: func(_ context.Context, clinicID, id uint64) error {
				assert.Equal(t, uint64(1), clinicID)
				return nil
			},
		}
		storage := &mockFileStorage{
			deleteFn: func(_ context.Context, key string) error {
				return nil
			},
		}
		svc := NewSharedFileService(repo, nil, storage)
		err := svc.CleanupExpired(ctx)
		assert.NoError(t, err)
	})

	t.Run("find expired error", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findExpiredFn: func(_ context.Context, _ int64) ([]*model.SharedFile, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewSharedFileService(repo, nil, nil)
		err := svc.CleanupExpired(ctx)
		assert.Error(t, err)
	})

	t.Run("partial error during cleanup", func(t *testing.T) {
		repo := &mockSharedFileRepository{
			findExpiredFn: func(_ context.Context, _ int64) ([]*model.SharedFile, error) {
				return []*model.SharedFile{
					{ID: 10, ClinicID: 1, FileKey: "key10"},
				}, nil
			},
		}
		storage := &mockFileStorage{
			deleteFn: func(_ context.Context, _ string) error {
				return errors.New("storage delete error")
			},
		}
		svc := NewSharedFileService(repo, nil, storage)
		err := svc.CleanupExpired(ctx)
		assert.Error(t, err)
	})
}
