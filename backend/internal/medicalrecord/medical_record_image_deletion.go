package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	medicalRecordImageDeletionPending = "pending_storage"
	medicalRecordImageDeletionDeleted = "storage_deleted"
	medicalRecordImageDeletionShared  = "retained_shared"
	medicalRecordImageDeletionErrMax  = 500
)

var (
	errMedicalRecordImageDeletionExists = errors.New("medical record image deletion already recorded")
	errMedicalRecordImageSigningRetired = errors.New("medical record image signing is retired")
)

// medicalRecordImageDeletion は署名停止と storage 削除再試行の正本。
// 監査ログは同じ transaction に参加するが、削除してよいかの判定には使わない。
type medicalRecordImageDeletion struct {
	ClinicID           uint64
	MedicalRecordID    uint64
	ImageID            uint64
	ImageObjectKey     string
	ThumbnailObjectKey string
	StorageStatus      string
	AttemptCount       int
	LastError          string
	ActorStaffID       uint64
}

// medicalRecordImageDeletionStore は削除台帳。未適用の migration で表が無いときはエラーを返し、黙って通さない。
type medicalRecordImageDeletionStore interface {
	Insert(ctx context.Context, row medicalRecordImageDeletion) error
	FindByImageID(ctx context.Context, clinicID, imageID uint64) (*medicalRecordImageDeletion, error)
	// BlockedImageIDs は一覧内の削除台帳を一括で引く（一覧時の N+1 防止）。
	BlockedImageIDs(ctx context.Context, clinicID uint64, imageIDs []uint64) (map[uint64]bool, error)
	UpdateStorageResult(ctx context.Context, clinicID, imageID uint64, status, lastError string) error
}

// medicalRecordImageObjectDeleter は storage オブジェクトの削除だけを行う。
// 署名 URL の発行口とは分け、削除 transaction の外から呼ぶ。
type medicalRecordImageObjectDeleter interface {
	Delete(ctx context.Context, key string) error
}

// ImageDeletionAudit は診療画像削除の監査と削除台帳を composition から渡す束。
type ImageDeletionAudit struct {
	audit AuditTxLogger
	store medicalRecordImageDeletionStore
}

// NewImageDeletionAudit は本番の SQL 台帳と、同じ transaction に参加する監査を束ねる。
func NewImageDeletionAudit(db *gorm.DB, audit AuditTxLogger) ImageDeletionAudit {
	return ImageDeletionAudit{audit: audit, store: newSQLMedicalRecordImageDeletionStore(db)}
}

func (s *medicalRecordImageService) SigningBlocked(ctx context.Context, clinicID, imageID uint64) (bool, error) {
	if s.deletions == nil {
		return true, apperrors.WrapInternalServerError("medical record image deletion dependency is required")
	}
	row, err := s.deletions.FindByImageID(ctx, clinicID, imageID)
	if err != nil {
		return true, err
	}
	return row != nil, nil
}

// BlockedImageIDs は一覧の削除台帳突合を 1 クエリで行う。
func (s *medicalRecordImageService) BlockedImageIDs(ctx context.Context, clinicID uint64, imageIDs []uint64) (map[uint64]bool, error) {
	if s.deletions == nil {
		return nil, apperrors.WrapInternalServerError("medical record image deletion dependency is required")
	}
	return s.deletions.BlockedImageIDs(ctx, clinicID, imageIDs)
}

func (s *medicalRecordImageService) Delete(
	ctx context.Context,
	clinicID, medicalRecordID, imageID, actorStaffID uint64,
	objects medicalRecordImageObjectDeleter,
) error {
	if clinicID == 0 || medicalRecordID == 0 || imageID == 0 {
		return apperrors.WrapInvalidInput("clinic_id, medical_record_id, and image_id are required")
	}
	if actorStaffID == 0 {
		return apperrors.WrapInvalidInput("actor staff id is required")
	}
	if err := s.requireImageDeletionDeps(); err != nil {
		return err
	}

	image, err := s.repo.FindByID(ctx, clinicID, imageID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			return apperrors.Wrap(err, "failed to get record image")
		}
		return s.retryImageStorageDelete(ctx, clinicID, medicalRecordID, imageID, objects)
	}
	if image.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("medical_record_image", "not found in this medical record")
	}

	imageKey, imageKeyErr := medicalRecordImageObjectKey(image.ImageURL)
	thumbKey, thumbKeyErr := medicalRecordImageObjectKey(image.ThumbnailURL)
	resolveErr := errors.Join(imageKeyErr, thumbKeyErr)

	var committed medicalRecordImageDeletion
	txErr := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := lockDraftMedicalRecord(txCtx, s.medRec, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテの画像は削除できません"); err != nil {
			return err
		}
		current, err := s.repo.FindByID(txCtx, clinicID, imageID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return errMedicalRecordImageDeletionExists
			}
			return apperrors.Wrap(err, "failed to get record image")
		}
		if current.MedicalRecordID != medicalRecordID {
			return apperrors.WrapNotFound("medical_record_image", "not found in this medical record")
		}
		row := medicalRecordImageDeletion{
			ClinicID:           clinicID,
			MedicalRecordID:    medicalRecordID,
			ImageID:            imageID,
			ImageObjectKey:     imageKey,
			ThumbnailObjectKey: thumbKey,
			StorageStatus:      medicalRecordImageDeletionPending,
			ActorStaffID:       actorStaffID,
		}
		if err := s.deletions.Insert(txCtx, row); err != nil {
			return err
		}
		if err := s.auditImageDeletion(txCtx, row, current.ImageURL, current.ThumbnailURL); err != nil {
			return err
		}
		if err := s.repo.Delete(txCtx, clinicID, imageID); err != nil {
			return apperrors.Wrap(err, "failed to delete record image")
		}
		committed = row
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, errMedicalRecordImageDeletionExists) {
			return s.retryImageStorageDelete(ctx, clinicID, medicalRecordID, imageID, objects)
		}
		return txErr
	}

	// 署名 URL は発行済みなら最大 15 分残る。storage 削除の失敗では削除台帳を戻さず
	// pending_storage のまま残し、失敗を caller へ返す。204 を返すと frontend は
	// 画像を一覧から消し、再試行経路（同一 imageID への再 DELETE → retryImageStorageDelete）
	// が実質到達不能になって orphan オブジェクトが残るため。
	if err := s.removeImageStorage(ctx, committed, objects, resolveErr); err != nil {
		return err
	}
	slog.InfoContext(ctx, "record image deleted", slog.Uint64("image_id", imageID))
	return nil
}

func (s *medicalRecordImageService) requireImageDeletionDeps() error {
	if s.repo == nil || s.medRec == nil || s.transactor == nil || s.deletions == nil || s.auditTx == nil {
		return apperrors.WrapInternalServerError("medical record image deletion dependencies are required")
	}
	return nil
}

func (s *medicalRecordImageService) auditImageDeletion(ctx context.Context, row medicalRecordImageDeletion, imageURL, thumbnailURL string) error {
	clinicID := row.ClinicID
	actorID := row.ActorStaffID
	imageID := row.ImageID
	err := s.auditTx.LogEntryTx(ctx, &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  auditActorTypeFor(&actorID),
		Action:     model.AuditActionMedicalRecordImageDelete,
		Resource:   model.AuditResourceMedicalRecordImage,
		ResourceID: &imageID,
		OldValue: map[string]any{
			"medical_record_id":    row.MedicalRecordID,
			"image_object_key":     row.ImageObjectKey,
			"thumbnail_object_key": row.ThumbnailObjectKey,
			// legacy 行は presigned URL が image_url に残っていることがある。
			// クエリ署名（token）を監査ログへ永続化しないよう query/fragment は剥ぐ。
			"image_url":     auditSafeMedicalRecordImageURL(imageURL),
			"thumbnail_url": auditSafeMedicalRecordImageURL(thumbnailURL),
		},
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to audit medical record image delete")
	}
	return nil
}

// auditSafeMedicalRecordImageURL は監査向けに URL の query/fragment を除去する。
// object key のような非 URL 値（parse で path 扱い）はそのまま返り、parse 失敗時は
// クエリ以降だけを落とす。
func auditSafeMedicalRecordImageURL(stored string) string {
	if u, err := url.Parse(stored); err == nil {
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}
	if i := strings.IndexByte(stored, '?'); i >= 0 {
		return stored[:i]
	}
	return stored
}

func (s *medicalRecordImageService) retryImageStorageDelete(
	ctx context.Context,
	clinicID, medicalRecordID, imageID uint64,
	objects medicalRecordImageObjectDeleter,
) error {
	// 台帳行は「認可済みの論理削除」（clinic 検証 + draft ロック下の tx で挿入済み）
	// の証跡なので、物理削除の再試行はカルテの draft 状態に依存しない。
	// draft ロックを要求すると確定後のカルテの pending_storage が回収不能になる。
	existing, err := s.deletions.FindByImageID(ctx, clinicID, imageID)
	if err != nil {
		return apperrors.Wrap(err, "failed to read medical record image deletion")
	}
	if existing == nil || existing.ClinicID != clinicID || existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("medical_record_image", fmt.Sprintf("%d", imageID))
	}
	if existing.StorageStatus != medicalRecordImageDeletionPending {
		return nil
	}
	return s.removeImageStorage(ctx, *existing, objects, nil)
}

// removeImageStorage は台帳行に対応する storage オブジェクトを削除し、結果を台帳へ記録する。
// 失敗時は pending_storage に留めて error を返す（caller が HTTP error として表面化する）。
func (s *medicalRecordImageService) removeImageStorage(
	ctx context.Context,
	row medicalRecordImageDeletion,
	objects medicalRecordImageObjectDeleter,
	resolveErr error,
) error {
	if resolveErr != nil {
		slog.ErrorContext(ctx, "medical record image storage key unresolved after delete",
			slog.Uint64("image_id", row.ImageID), slog.String("error", resolveErr.Error()))
		s.recordImageStorageResult(ctx, row, medicalRecordImageDeletionPending, resolveErr)
		return resolveErr
	}
	keys := uniqueMedicalRecordImageKeys(row.ImageObjectKey, row.ThumbnailObjectKey)
	if len(keys) == 0 {
		s.recordImageStorageResult(ctx, row, medicalRecordImageDeletionDeleted, nil)
		return nil
	}

	retained := false
	var failed error
	for _, key := range keys {
		outcome, err := s.deleteOneImageObject(ctx, row, key, objects)
		if err != nil {
			failed = err
			break
		}
		if outcome == medicalRecordImageDeletionShared {
			retained = true
		}
	}
	status := medicalRecordImageDeletionDeleted
	switch {
	case failed != nil:
		status = medicalRecordImageDeletionPending
	case retained:
		status = medicalRecordImageDeletionShared
	}
	s.recordImageStorageResult(ctx, row, status, failed)
	if failed != nil {
		return apperrors.Wrap(failed, "failed to delete medical record image storage object")
	}
	return nil
}

func (s *medicalRecordImageService) deleteOneImageObject(
	ctx context.Context,
	row medicalRecordImageDeletion,
	key string,
	objects medicalRecordImageObjectDeleter,
) (string, error) {
	count, err := s.repo.CountLiveByStorageKey(ctx, row.ClinicID, key)
	if err != nil {
		slog.ErrorContext(ctx, "medical record image live-key count failed",
			slog.Uint64("image_id", row.ImageID), slog.String("key", key), slog.String("error", err.Error()))
		return "", err
	}
	if count > 0 {
		slog.InfoContext(ctx, "medical record image object retained because another registration uses it",
			slog.Uint64("image_id", row.ImageID), slog.String("key", key), slog.Int64("live_registrations", count))
		return medicalRecordImageDeletionShared, nil
	}
	if !medicalRecordImageKeyBelongsToRecord(key, row.MedicalRecordID) {
		err = errors.New("medical record image object key is outside the record")
		slog.ErrorContext(ctx, "medical record image storage delete refused",
			slog.Uint64("image_id", row.ImageID), slog.String("error", err.Error()))
		return "", err
	}
	if objects == nil {
		err = errors.New("medical record image object deleter is not configured")
		slog.ErrorContext(ctx, "medical record image storage delete failed",
			slog.Uint64("image_id", row.ImageID), slog.String("key", key), slog.String("error", err.Error()))
		return "", err
	}
	if err = objects.Delete(ctx, key); err != nil {
		slog.ErrorContext(ctx, "medical record image storage delete failed",
			slog.Uint64("image_id", row.ImageID), slog.String("key", key), slog.String("error", err.Error()))
		return "", err
	}
	return medicalRecordImageDeletionDeleted, nil
}

func (s *medicalRecordImageService) recordImageStorageResult(
	ctx context.Context,
	row medicalRecordImageDeletion,
	status string,
	cause error,
) {
	err := s.deletions.UpdateStorageResult(ctx, row.ClinicID, row.ImageID, status, truncateImageDeletionError(cause))
	if err != nil {
		slog.ErrorContext(ctx, "failed to record medical record image storage result",
			slog.Uint64("image_id", row.ImageID), slog.String("error", err.Error()))
	}
}

func medicalRecordImageObjectKey(stored string) (string, error) {
	key, isObject, err := resolveMedicalRecordImageStorageKey(stored)
	if err != nil {
		return "", err
	}
	if !isObject {
		return "", nil
	}
	return key, nil
}

func uniqueMedicalRecordImageKeys(keys ...string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func truncateImageDeletionError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) <= medicalRecordImageDeletionErrMax {
		return msg
	}
	return msg[:medicalRecordImageDeletionErrMax]
}
