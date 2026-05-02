package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TagSummaryRow はタグ集計クエリの結果行。
type TagSummaryRow struct {
	TagName    string `gorm:"column:tag_name"`
	Category   string `gorm:"column:category"`
	OwnerCount int64  `gorm:"column:owner_count"`
}

// TagOwnerRow はタグ別飼い主検索クエリの結果行。
type TagOwnerRow struct {
	OwnerID    uint64  `gorm:"column:owner_id"`
	OwnerName  string  `gorm:"column:owner_name"`
	LineUserID *string `gorm:"column:line_user_id"`
	Tags       []string
}

// TagEntry は BulkReplaceOwnerTags で渡すタグ1件分のデータ。
type TagEntry struct {
	TagName  string
	Category string
}

// LstepTagCacheRepository はLステップタグキャッシュの永続化インターフェース。
type LstepTagCacheRepository interface {
	// UpsertTag は (clinic_id, owner_id, tag_name) でUPSERTする。
	UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category string) error
	// DeleteTag は特定タグを削除する。
	DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	// DeleteAllByOwner は飼い主の全タグキャッシュを削除する。
	DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error
	// FindByOwner は飼い主のタグキャッシュ一覧を返す。
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	// CountByTag はクリニック内で指定タグを持つ飼い主数を返す。
	CountByTag(ctx context.Context, clinicID uint64, tagName string) (int64, error)
	// TagSummary はクリニック内タグ名・カテゴリ別の飼い主数集計を返す。totalOwnersWithLstep はタグを1つ以上持つ飼い主数。
	TagSummary(ctx context.Context, clinicID uint64) (rows []TagSummaryRow, totalOwnersWithLstep int64, err error)
	// FindOwnersByTag は指定タグを持つ飼い主をページネーションで返す。nameQuery が空文字以外のとき飼い主名で部分一致フィルタ。
	FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error)
	// BulkReplaceOwnerTags は飼い主の全タグを削除してから指定タグを一括挿入する（BE-018用）。
	BulkReplaceOwnerTags(ctx context.Context, clinicID, ownerID uint64, tags []TagEntry) error
	// FindOwnerIDsByTag は指定タグを持つ飼い主IDリストを全件返す（FEAT-383 バッチ用）。
	FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error)
}

type lstepTagCacheRepository struct{ db *gorm.DB }

// NewLstepTagCacheRepository は LstepTagCacheRepository を初期化して返す。
func NewLstepTagCacheRepository(db *gorm.DB) LstepTagCacheRepository {
	return &lstepTagCacheRepository{db: db}
}

func (r *lstepTagCacheRepository) UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category string) error {
	record := &model.LstepTagCache{
		ClinicID: clinicID,
		OwnerID:  ownerID,
		TagName:  tagName,
		Category: category,
		SyncedAt: time.Now(),
	}
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "owner_id"}, {Name: "tag_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"category", "synced_at"}),
		}).
		Create(record).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d tag=%s", ownerID, tagName))
	}
	return nil
}

func (r *lstepTagCacheRepository) DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error {
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ? AND tag_name = ?", ownerID, tagName).
		Delete(&model.LstepTagCache{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d tag=%s", ownerID, tagName))
	}
	return nil
}

func (r *lstepTagCacheRepository) DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error {
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ?", ownerID).
		Delete(&model.LstepTagCache{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d", ownerID))
	}
	return nil
}

func (r *lstepTagCacheRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	var records []*model.LstepTagCache
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("owner_id = ?", ownerID).
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d", ownerID))
	}
	return records, nil
}

func (r *lstepTagCacheRepository) CountByTag(ctx context.Context, clinicID uint64, tagName string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.LstepTagCache{}).
		Scopes(clinicScope(clinicID)).
		Where("tag_name = ?", tagName).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("tag=%s", tagName))
	}
	return count, nil
}

func (r *lstepTagCacheRepository) TagSummary(ctx context.Context, clinicID uint64) ([]TagSummaryRow, int64, error) {
	var rows []TagSummaryRow
	err := r.db.WithContext(ctx).
		Raw(`SELECT tag_name, category, COUNT(DISTINCT owner_id) AS owner_count
		     FROM lstep_tag_cache
		     WHERE clinic_id = ?
		     GROUP BY tag_name, category
		     ORDER BY owner_count DESC`, clinicID).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d summary", clinicID))
	}
	var total int64
	if err := r.db.WithContext(ctx).
		Raw(`SELECT COUNT(DISTINCT owner_id) FROM lstep_tag_cache WHERE clinic_id = ?`, clinicID).
		Scan(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d total_owners", clinicID))
	}
	return rows, total, nil
}

func (r *lstepTagCacheRepository) FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error) {
	type ownerStub struct {
		OwnerID    uint64  `gorm:"column:owner_id"`
		OwnerName  string  `gorm:"column:owner_name"`
		LineUserID *string `gorm:"column:line_user_id"`
	}
	baseSQL := `FROM owners o
		JOIN lstep_tag_cache tc ON tc.owner_id = o.id AND tc.clinic_id = ?
		WHERE o.clinic_id = ? AND o.deleted_at IS NULL AND tc.tag_name = ?`
	args := []any{clinicID, clinicID, tagName}
	if nameQuery != "" {
		baseSQL += ` AND o.name LIKE ?`
		args = append(args, "%"+nameQuery+"%")
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(DISTINCT o.id) "+baseSQL, args...).
		Scan(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d tag=%s count", clinicID, tagName))
	}

	pageArgs := make([]any, 0, len(args)+2)
	pageArgs = append(pageArgs, args...)
	pageArgs = append(pageArgs, limit, offset)
	var stubs []ownerStub
	if err := r.db.WithContext(ctx).
		Raw("SELECT DISTINCT o.id AS owner_id, o.name AS owner_name, o.line_user_id "+baseSQL+` ORDER BY o.id LIMIT ? OFFSET ?`, pageArgs...).
		Scan(&stubs).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d tag=%s page", clinicID, tagName))
	}
	if len(stubs) == 0 {
		return []TagOwnerRow{}, total, nil
	}

	ownerIDs := make([]uint64, len(stubs))
	for i, s := range stubs {
		ownerIDs[i] = s.OwnerID
	}
	type tagRow struct {
		OwnerID uint64 `gorm:"column:owner_id"`
		TagName string `gorm:"column:tag_name"`
	}
	var tagRows []tagRow
	if err := r.db.WithContext(ctx).
		Raw(`SELECT owner_id, tag_name FROM lstep_tag_cache WHERE clinic_id = ? AND owner_id IN ? ORDER BY owner_id, tag_name`, clinicID, ownerIDs).
		Scan(&tagRows).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d batch_tags", clinicID))
	}

	tagMap := make(map[uint64][]string, len(stubs))
	for _, tr := range tagRows {
		tagMap[tr.OwnerID] = append(tagMap[tr.OwnerID], tr.TagName)
	}
	result := make([]TagOwnerRow, len(stubs))
	for i, s := range stubs {
		tags := tagMap[s.OwnerID]
		if tags == nil {
			tags = []string{}
		}
		result[i] = TagOwnerRow{OwnerID: s.OwnerID, OwnerName: s.OwnerName, LineUserID: s.LineUserID, Tags: tags}
	}
	return result, total, nil
}

func (r *lstepTagCacheRepository) FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error) {
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).Model(&model.LstepTagCache{}).
		Where("clinic_id = ? AND tag_name = ?", clinicID, tagName).
		Distinct("owner_id").
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d tag=%s", clinicID, tagName))
	}
	ids := make([]uint64, len(rows))
	for i, r := range rows {
		ids[i] = r.OwnerID
	}
	return ids, nil
}

func (r *lstepTagCacheRepository) BulkReplaceOwnerTags(ctx context.Context, clinicID, ownerID uint64, tags []TagEntry) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
			Delete(&model.LstepTagCache{}).Error; err != nil {
			return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d delete", ownerID))
		}
		if len(tags) == 0 {
			return nil
		}
		now := time.Now()
		records := make([]*model.LstepTagCache, len(tags))
		for i, t := range tags {
			records[i] = &model.LstepTagCache{
				ClinicID: clinicID, OwnerID: ownerID,
				TagName: t.TagName, Category: t.Category, SyncedAt: now,
			}
		}
		if err := tx.Create(&records).Error; err != nil {
			return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d insert", ownerID))
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "lstep tag cache bulk replace")
	}
	return nil
}
