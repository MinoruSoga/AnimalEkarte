package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LstepFriendAttributeSnapshotRepository は lstep_friend_attribute_snapshots テーブルの永続化インターフェース。
type LstepFriendAttributeSnapshotRepository interface {
	// Create は新規友だち属性スナップショットを作成する。
	Create(ctx context.Context, snapshot *model.LstepFriendAttributeSnapshot) error
	// BulkCreate は CSV インポート用に複数スナップショットを一括作成する。UNIQUE 制約衝突時はスキップ。
	BulkCreate(ctx context.Context, snapshots []*model.LstepFriendAttributeSnapshot) error
	// FindLatestByOwner はクリニックスコープで指定 LINE ユーザーの最新スナップショットを返す。
	FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error)
	// ListByClinicAndDateRange はクリニックスコープで期間内スナップショット一覧を返す。
	ListByClinicAndDateRange(ctx context.Context, clinicID uint64, since, until time.Time) ([]*model.LstepFriendAttributeSnapshot, error)
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

func (r *lstepFriendAttributeSnapshotRepository) BulkCreate(ctx context.Context, snapshots []*model.LstepFriendAttributeSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(snapshots, 100).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "bulk_create")
	}
	return nil
}

func (r *lstepFriendAttributeSnapshotRepository) FindLatestByOwner(ctx context.Context, clinicID uint64, lineUserID string) (*model.LstepFriendAttributeSnapshot, error) {
	var snapshot model.LstepFriendAttributeSnapshot
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND line_user_id = ?", clinicID, lineUserID).
		Order("snapshot_taken_at DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", lineUserID)
	}
	return &snapshot, nil
}

func (r *lstepFriendAttributeSnapshotRepository) ListByClinicAndDateRange(ctx context.Context, clinicID uint64, since, until time.Time) ([]*model.LstepFriendAttributeSnapshot, error) {
	var snapshots []*model.LstepFriendAttributeSnapshot
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND snapshot_taken_at >= ? AND snapshot_taken_at < ?", clinicID, since, until).
		Order("snapshot_taken_at ASC").
		Find(&snapshots).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_friend_attribute_snapshot", "list_by_date_range")
	}
	return snapshots, nil
}
