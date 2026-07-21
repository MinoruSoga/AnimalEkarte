package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LstepTagCacheRepository は internal/lstep への移行facade（BE9-2C L②・BE9-2F削除予定）。
type (
	LstepTagCacheRepository = lstep.LstepTagCacheRepository
	// TagSummaryRow はタグ集計行 DTO（残留 consumer=lstep_tag_summary_service=L③ 用 alias）。
	TagSummaryRow = lstep.TagSummaryRow
	TagOwnerRow   = lstep.TagOwnerRow
	TagEntry      = lstep.TagEntry
)

// NewLstepTagCacheRepository は internal/lstep の実装を返す（BE9-2C L② facade）。
func NewLstepTagCacheRepository(db *gorm.DB) LstepTagCacheRepository {
	return lstep.NewLstepTagCacheRepository(db)
}
