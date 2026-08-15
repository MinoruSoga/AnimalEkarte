package lstep

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/textsearch"
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
	Reason     *string `gorm:"column:reason"`
	Tags       []string
}

// TagEntry は BulkReplaceOwnerTags で渡すタグ1件分のデータ。
type TagEntry struct {
	TagName  string
	Category string
	Reason   string
}

// LstepTagCacheRepository はLステップタグキャッシュの永続化インターフェース。
type LstepTagCacheRepository interface {
	// UpsertTag は (clinic_id, owner_id, tag_name) でUPSERTする。reason が空文字の場合は NULL として保存する。
	UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
	// DeleteTag は特定タグを削除する。
	DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error
	// DeleteAllByOwner は飼い主の全タグキャッシュを削除する。
	DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error
	// FindByOwner は飼い主のタグキャッシュ一覧を返す。
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error)
	// FindByOwners は複数飼い主のタグキャッシュ一覧を owner_id 別にまとめて返す（N+1回避のバッチ版）。
	// ownerIDs が空の場合は空mapを即返す。タグを持たない owner_id はキーとして存在しない。
	FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
	// TagSummary はクリニック内タグ名・カテゴリ別の飼い主数集計を返す。totalOwnersWithLstep はタグを1つ以上持つ飼い主数。
	TagSummary(ctx context.Context, clinicID uint64) (rows []TagSummaryRow, totalOwnersWithLstep int64, err error)
	// FindOwnersByTag は指定タグを持つ飼い主をページネーションで返す。nameQuery が空文字以外のとき飼い主名で部分一致フィルタ。
	FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error)
	// FindOwnerIDsByTag は指定タグを持つ飼い主IDリストを全件返す（FEAT-383 バッチ用）。
	FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error)
}

type lstepTagCacheRepository struct{ db *gorm.DB }

// NewLstepTagCacheRepository は LstepTagCacheRepository を初期化して返す。
func NewLstepTagCacheRepository(db *gorm.DB) LstepTagCacheRepository {
	return &lstepTagCacheRepository{db: db}
}

func (r *lstepTagCacheRepository) UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error {
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	record := &model.LstepTagCache{
		ClinicID: clinicID,
		OwnerID:  ownerID,
		TagName:  tagName,
		Category: category,
		Reason:   reasonPtr,
		SyncedAt: time.Now(),
	}
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "owner_id"}, {Name: "tag_name"}},
			DoUpdates: clause.AssignmentColumns([]string{"category", "reason", "synced_at"}),
		}).
		Create(record).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d tag=%s", ownerID, tagName))
	}
	return nil
}

func (r *lstepTagCacheRepository) DeleteTag(ctx context.Context, clinicID, ownerID uint64, tagName string) error {
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("owner_id = ? AND tag_name = ?", ownerID, tagName).
		Delete(&model.LstepTagCache{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d tag=%s", ownerID, tagName))
	}
	return nil
}

func (r *lstepTagCacheRepository) DeleteAllByOwner(ctx context.Context, clinicID, ownerID uint64) error {
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where("owner_id = ?", ownerID).
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("owner=%d", ownerID))
	}
	return records, nil
}

func (r *lstepTagCacheRepository) FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) {
	result := make(map[uint64][]*model.LstepTagCache, len(ownerIDs))
	if len(ownerIDs) == 0 {
		return result, nil
	}
	var records []*model.LstepTagCache
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("owner_id IN ?", ownerIDs).
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_tag_cache", fmt.Sprintf("clinic=%d batch_owners", clinicID))
	}
	for _, rec := range records {
		result[rec.OwnerID] = append(result[rec.OwnerID], rec)
	}
	return result, nil
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
		Reason     *string `gorm:"column:reason"`
	}
	baseSQL := `FROM owners o
		JOIN lstep_tag_cache tc ON tc.owner_id = o.id AND tc.clinic_id = ?
		WHERE o.clinic_id = ? AND o.deleted_at IS NULL AND tc.tag_name = ?`
	args := []any{clinicID, clinicID, tagName}
	if nameQuery != "" {
		// G2C-01: escape LIKE metacharacters; pair with ESCAPE (owner/pet repository pattern).
		baseSQL += ` AND o.name LIKE ? ESCAPE '\'`
		args = append(args, "%"+textsearch.EscapeLike(nameQuery)+"%")
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
		Raw("SELECT DISTINCT o.id AS owner_id, o.name AS owner_name, o.line_user_id, tc.reason "+baseSQL+` ORDER BY o.id LIMIT ? OFFSET ?`, pageArgs...).
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
		result[i] = TagOwnerRow{OwnerID: s.OwnerID, OwnerName: s.OwnerName, LineUserID: s.LineUserID, Reason: s.Reason, Tags: tags}
	}
	return result, total, nil
}

func (r *lstepTagCacheRepository) FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error) {
	type row struct{ OwnerID uint64 }
	var rows []row
	err := r.db.WithContext(ctx).Table("lstep_tag_cache AS c").
		Joins("JOIN owners AS o ON o.id = c.owner_id AND o.clinic_id = c.clinic_id AND o.deleted_at IS NULL").
		Where("c.clinic_id = ? AND c.tag_name = ?", clinicID, tagName).
		Distinct("c.owner_id").
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
