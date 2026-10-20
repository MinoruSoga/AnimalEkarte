package model

import (
	"time"
)

// ManualCategory は取扱説明書のカテゴリ
type ManualCategory string

const (
	ManualCategoryScreens   ManualCategory = "screens"
	ManualCategoryWorkflows ManualCategory = "workflows"
)

// IsValidManualCategory は ManualCategory が有効値か判定
func IsValidManualCategory(c string) bool {
	return c == string(ManualCategoryScreens) || c == string(ManualCategoryWorkflows)
}

// ManualArticle は取扱説明書のオーバーライド版（DB に保存された編集後マニュアル）
//
// 設計方針:
//   - フロントエンドが MD ファイルをデフォルト（バンドル）として保持
//   - DB にはオーバーライド版を保存
//   - 読み込み時: DB に該当 slug があればそれを優先、なければ MD ファイル
//   - 編集時: 該当 slug を DB に upsert
//   - マニュアルは医院共通の情報のため clinic_id は持たない
type ManualArticle struct {
	ID                 uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Category           ManualCategory `gorm:"column:category;type:text;not null" json:"category"`
	Slug               string         `gorm:"column:slug;type:text;not null" json:"slug"`
	CreatedAt          time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	Version            int64          `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CurrentVersionID   *uint64        `gorm:"column:current_version_id;type:bigint" json:"-"`
	PublishedVersionID *uint64        `gorm:"column:published_version_id;type:bigint" json:"-"`
	HiddenAt           *time.Time     `gorm:"column:hidden_at;type:timestamptz" json:"-"`
}

// TableName は ManualArticle のテーブル名
func (ManualArticle) TableName() string { return "manual_articles" }

// ManualArticleVersion はマニュアル編集履歴（編集ごとのスナップショット）
type ManualArticleVersion struct {
	ID              uint64       `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ArticleID       uint64       `gorm:"column:article_id;type:bigint;not null" json:"article_id"`
	Title           string       `gorm:"column:title;type:text;not null" json:"title"`
	OrderValue      ExactNumeric `gorm:"column:order_value;type:numeric(10, 2);not null" json:"order_value"`
	Section         string       `gorm:"column:section;type:text;not null" json:"section"`
	BodyMarkdown    string       `gorm:"column:body_markdown;type:text;not null" json:"body_markdown"`
	EditedByStaffID *uint64      `gorm:"column:edited_by_staff_id;type:bigint" json:"edited_by_staff_id,omitempty"`
	EditedAt        time.Time    `gorm:"column:edited_at;type:timestamptz;not null;default:now()" json:"edited_at"`
	ContentDigest   *[]byte      `gorm:"column:content_digest;type:bytea" json:"-"`
	ContentRevision *int64       `gorm:"column:content_revision;type:bigint" json:"-"`
	RestoredFromID  *uint64      `gorm:"column:restored_from_id;type:bigint" json:"-"`
	SourceKind      *string      `gorm:"column:source_kind;type:text" json:"-"`
}

// TableName は ManualArticleVersion のテーブル名
func (ManualArticleVersion) TableName() string { return "manual_article_versions" }
