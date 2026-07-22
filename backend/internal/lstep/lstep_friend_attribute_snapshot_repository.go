package lstep

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepFriendAttributeSnapshotRepository は lstep_friend_attribute_snapshots テーブルの永続化インターフェース。
type LstepFriendAttributeSnapshotRepository interface {
	// Create は新規友だち属性スナップショットを作成する。
	Create(ctx context.Context, snapshot *model.LstepFriendAttributeSnapshot) error
	// FindLatestByOwner はクリニックスコープで指定 LINE ユーザーの最新スナップショットを返す。
	FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error)
}

type lstepFriendAttributeSnapshotRepository struct{ db *gorm.DB }

// NewLstepFriendAttributeSnapshotRepository は LstepFriendAttributeSnapshotRepository を初期化して返す。
func NewLstepFriendAttributeSnapshotRepository(db *gorm.DB) LstepFriendAttributeSnapshotRepository {
	return &lstepFriendAttributeSnapshotRepository{db: db}
}

func (r *lstepFriendAttributeSnapshotRepository) Create(ctx context.Context, snapshot *model.LstepFriendAttributeSnapshot) error {
	if err := r.db.WithContext(ctx).Create(snapshot).Error; err != nil {
		return apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "")
	}
	return nil
}

func (r *lstepFriendAttributeSnapshotRepository) FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error) {
	var snapshot model.LstepFriendAttributeSnapshot
	err := r.db.WithContext(ctx).
		Table("lstep_friend_attribute_snapshots AS snapshot").
		Where("snapshot.clinic_id = ? AND snapshot.line_user_id = ?", clinicID, lineUserID).
		Where(`snapshot.csv_import_id IS NULL OR EXISTS (
			SELECT 1
			FROM lstep_csv_imports AS csv_import
			WHERE csv_import.id = snapshot.csv_import_id
			  AND csv_import.clinic_id = snapshot.clinic_id
		)`).
		Order("snapshot.snapshot_taken_at DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", lineUserID)
	}
	return &snapshot, nil
}
