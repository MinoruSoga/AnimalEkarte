package support

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// manual_cache.go — support_manual_embeddings 派生キャッシュのデータアクセス。
// 共有マニュアル由来の再生成可能データであり clinic_id は持たない（R3）。
// 本文・質問・スタッフ情報は保存しない。

// supportManualEmbeddingRow は support_manual_embeddings の1行。
// 派生キャッシュのため support パッケージローカルに置く（global model 変更なし）。
type supportManualEmbeddingRow struct {
	CacheKey       string         `gorm:"column:cache_key;primaryKey"`
	Model          string         `gorm:"column:model"`
	TaskType       string         `gorm:"column:task_type"`
	ChunkerVersion int            `gorm:"column:chunker_version"`
	Dimensions     int            `gorm:"column:dimensions"`
	Embedding      embeddingArray `gorm:"column:embedding;type:real[]"`
	CreatedAt      time.Time      `gorm:"column:created_at;default:now()"`
}

func (supportManualEmbeddingRow) TableName() string { return "support_manual_embeddings" }

// manualEmbeddingMaxKeys は 1 要求の Load キー上限（R3: 2048 chunks）。
const manualEmbeddingMaxKeys = 2048

// ManualEmbeddingCache は cache_key → 正規化済みベクトルの参照/保存。
type ManualEmbeddingCache interface {
	// Load は指定 keys のうち存在する行だけを返す（exact key intersection）。
	Load(ctx context.Context, keys []string) (map[string][]float64, error)
	// StoreWithKey は検証済みベクトルを明示 cache_key で保存する（ON CONFLICT DO NOTHING）。
	StoreWithKey(ctx context.Context, cacheKey, model, taskType string, chunkerVersion, dimensions int, embedding []float64) error
}

type manualEmbeddingCache struct {
	db    *gorm.DB
	model string
}

// NewManualEmbeddingCache は ManualEmbeddingCache を初期化する。
// model はこのキャッシュが扱う唯一のモデル名（他モデルの行は採用しない）。
func NewManualEmbeddingCache(db *gorm.DB, model string) ManualEmbeddingCache {
	return &manualEmbeddingCache{db: db, model: model}
}

func (r *manualEmbeddingCache) Load(ctx context.Context, keys []string) (map[string][]float64, error) {
	if r.db == nil {
		return nil, apperrors.WrapInvalidInput("manual embedding cache requires a db")
	}
	if len(keys) == 0 {
		return map[string][]float64{}, nil
	}
	if len(keys) > manualEmbeddingMaxKeys {
		return nil, apperrors.WrapInvalidInput("too many embedding cache keys")
	}
	uniq := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		uniq[k] = struct{}{}
	}
	var rows []supportManualEmbeddingRow
	if err := persistence.DBOrTx(ctx, r.db).WithContext(ctx).
		Where("cache_key IN ?", keys).Find(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load manual embeddings")
	}
	return validatedCacheRows(rows, uniq, r.model), nil
}

func (r *manualEmbeddingCache) StoreWithKey(ctx context.Context, cacheKey, model, taskType string, chunkerVersion, dimensions int, embedding []float64) error {
	if r.db == nil {
		return apperrors.WrapInvalidInput("manual embedding cache requires a db")
	}
	if !sha256HexRe.MatchString(cacheKey) {
		return apperrors.WrapInvalidInput("invalid cache key")
	}
	if model == "" || model != r.model ||
		taskType != manualarticle.TaskTypeRetrievalDocument ||
		chunkerVersion != manualarticle.ChunkerVersion ||
		dimensions != manualarticle.EmbeddingDimensions {
		return apperrors.WrapInvalidInput("invalid embedding metadata")
	}
	norm, err := normalizeEmbedding(embedding)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid embedding vector")
	}
	vec := make(embeddingArray, len(norm))
	var f32norm float64
	for i, v := range norm {
		vec[i] = float32(v)
		f32norm += float64(vec[i]) * float64(vec[i])
	}
	if !(f32norm > 0) {
		return apperrors.WrapInvalidInput("invalid embedding vector")
	}
	row := supportManualEmbeddingRow{
		CacheKey:       cacheKey,
		Model:          model,
		TaskType:       taskType,
		ChunkerVersion: chunkerVersion,
		Dimensions:     dimensions,
		Embedding:      vec,
	}
	err = persistence.DBOrTx(ctx, r.db).WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
	if err != nil {
		return apperrors.Wrap(err, "failed to store manual embedding")
	}
	return nil
}
