package lstep

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SharedFileResponse はGETレスポンス用
type SharedFileResponse struct {
	ID         uint64     `json:"id"`
	ClinicID   uint64     `json:"clinic_id"`
	OwnerID    *uint64    `json:"owner_id"`
	UploadedBy uint64     `json:"uploaded_by"`
	FileType   string     `json:"file_type"`
	FileName   string     `json:"file_name"`
	FileSize   int64      `json:"file_size"`
	Purpose    string     `json:"purpose"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// UploadSharedFileInput はPOSTリクエスト用
type UploadSharedFileInput struct {
	Content     io.Reader
	FileName    string
	ContentType string
	FileType    string
	FileSize    int64
	Purpose     string
	OwnerID     *uint64
}

// SharedFileService は shared_files の管理インターフェース。
type SharedFileService interface {
	Upload(ctx context.Context, clinicID, uploadedBy uint64, input *UploadSharedFileInput) (*SharedFileResponse, error)
	GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error)
	FindAll(ctx context.Context, clinicID uint64) ([]*SharedFileResponse, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// CleanupExpired は7日 - 25時間経過したファイルを削除する（バッチ用）。
	CleanupExpired(ctx context.Context) error
}

type sharedFileService struct {
	repo      SharedFileRepository
	ownerRepo SharedFileOwnerFinder
	storage   SharedFileStorage
}

// SharedFileOwnerFinder verifies that an optional owner belongs to the authenticated clinic.
type SharedFileOwnerFinder interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

// SharedFileStorage is the target-owned object-storage contract.
type SharedFileStorage interface {
	Upload(ctx context.Context, key string, content io.Reader, contentType string) error
	GetSignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// NewSharedFileService は SharedFileService を初期化して返す。
func NewSharedFileService(repo SharedFileRepository, ownerRepo SharedFileOwnerFinder, storage SharedFileStorage) SharedFileService {
	return &sharedFileService{repo: repo, ownerRepo: ownerRepo, storage: storage}
}

func (s *sharedFileService) Upload(ctx context.Context, clinicID, uploadedBy uint64, input *UploadSharedFileInput) (*SharedFileResponse, error) {
	// owner_id が指定された場合、同一クリニックに存在することを検証
	if input.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
			slog.ErrorContext(ctx, "owner not found or different clinic", "error", err, "clinic_id", clinicID, "owner_id", *input.OwnerID)
			return nil, apperrors.Wrap(err, "owner not found")
		}
	}

	key, err := generateSharedFileKey(clinicID, input.FileName)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to generate file key")
	}

	if err := s.storage.Upload(ctx, key, input.Content, input.ContentType); err != nil {
		slog.ErrorContext(ctx, "failed to upload shared file to storage", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to upload file")
	}

	record := &model.SharedFile{
		ClinicID:   clinicID,
		UploadedBy: uploadedBy,
		OwnerID:    input.OwnerID,
		FileType:   input.FileType,
		FileName:   input.FileName,
		FileKey:    key,
		FileSize:   input.FileSize,
		Purpose:    input.Purpose,
	}
	if err := s.repo.Create(ctx, record); err != nil {
		slog.ErrorContext(ctx, "failed to create shared_file record", "error", err, "key", key)
		// ストレージをロールバック
		if delErr := s.storage.Delete(ctx, key); delErr != nil {
			slog.ErrorContext(ctx, "failed to cleanup uploaded file after DB error", "error", delErr, "key", key)
		}
		return nil, apperrors.Wrap(err, "failed to create shared file record")
	}

	return toSharedFileSvcResponse(record), nil
}

func (s *sharedFileService) GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error) {
	record, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find shared file", "error", err, "id", id)
		return "", apperrors.Wrap(err, "failed to find shared file")
	}
	url, err := s.storage.GetSignedURL(ctx, record.FileKey, 24*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate signed URL", "error", err, "key", record.FileKey)
		return "", apperrors.Wrap(err, "failed to generate signed URL")
	}
	return url, nil
}

func (s *sharedFileService) FindAll(ctx context.Context, clinicID uint64) ([]*SharedFileResponse, error) {
	records, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find shared files", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find shared files")
	}
	result := make([]*SharedFileResponse, 0, len(records))
	for _, r := range records {
		result = append(result, toSharedFileSvcResponse(r))
	}
	return result, nil
}

func (s *sharedFileService) Delete(ctx context.Context, clinicID, id uint64) error {
	record, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find shared file for delete", "error", err, "id", id)
		return apperrors.Wrap(err, "failed to find shared file")
	}
	if err := s.storage.Delete(ctx, record.FileKey); err != nil {
		slog.ErrorContext(ctx, "failed to delete file from storage", "error", err, "key", record.FileKey)
		return apperrors.Wrap(err, "failed to delete file from storage")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to soft-delete shared file", "error", err, "id", id)
		return apperrors.Wrap(err, "failed to delete shared file record")
	}
	return nil
}

// CleanupExpired は 7日 - 25時間 = 143時間以上経過したファイルを削除する。
// 24時間有効の署名付きURLとの衝突を防ぐために25時間バッファを設ける。
// G2F-06: FindExpired is hard-capped; drain batches until a short page.
func (s *sharedFileService) CleanupExpired(ctx context.Context) error {
	const expiryHours = 7*24 - 25
	threshold := time.Now().Add(-expiryHours * time.Hour).Unix()

	var firstErr error
	for {
		records, err := s.repo.FindExpired(ctx, threshold)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find expired shared files", "error", err)
			return apperrors.Wrap(err, "failed to find expired shared files")
		}
		if len(records) == 0 {
			break
		}
		progress := 0
		for _, r := range records {
			if err := s.storage.Delete(ctx, r.FileKey); err != nil {
				slog.ErrorContext(ctx, "failed to delete expired file from storage", "error", err, "key", r.FileKey, "id", r.ID)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			if err := s.repo.Delete(ctx, r.ClinicID, r.ID); err != nil {
				slog.ErrorContext(ctx, "failed to soft-delete expired shared file", "error", err, "id", r.ID)
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			progress++
		}
		// Full batch may mean more remain; stop on short page or zero progress
		// (failed deletes would otherwise re-fetch the same rows forever).
		if len(records) < sharedFileExpiredMax || progress == 0 {
			break
		}
	}
	if firstErr != nil {
		return apperrors.Wrap(firstErr, "cleanup partially failed")
	}
	return nil
}

func toSharedFileSvcResponse(r *model.SharedFile) *SharedFileResponse {
	return &SharedFileResponse{
		ID:         r.ID,
		ClinicID:   r.ClinicID,
		OwnerID:    r.OwnerID,
		UploadedBy: r.UploadedBy,
		FileType:   r.FileType,
		FileName:   r.FileName,
		FileSize:   r.FileSize,
		Purpose:    r.Purpose,
		ExpiresAt:  r.ExpiresAt,
		CreatedAt:  r.CreatedAt,
	}
}

// generateSharedFileKey はクリニックID + 乱数16バイトでユニークなS3キーを生成する。
func generateSharedFileKey(clinicID uint64, fileName string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	ext := filepath.Ext(fileName)
	return fmt.Sprintf("shared/%d/%s%s", clinicID, hex.EncodeToString(b), ext), nil
}
