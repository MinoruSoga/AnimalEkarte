package model

import "time"

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
	ID               uint64         `gorm:"primaryKey;autoIncrement"                 json:"id"`
	Category         ManualCategory `gorm:"type:text;not null"                       json:"category"`
	Slug             string         `gorm:"not null"                                 json:"slug"`
	Title            string         `gorm:"not null"                                 json:"title"`
	OrderValue       float64        `gorm:"column:order_value;not null;default:9999" json:"order_value"`
	Section          string         `gorm:"not null"                                 json:"section"`
	BodyMarkdown     string         `gorm:"not null;type:text"                       json:"body_markdown"`
	UpdatedByStaffID *uint64        `gorm:"column:updated_by_staff_id"               json:"updated_by_staff_id,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"                           json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"                           json:"updated_at"`
}

// TableName は ManualArticle のテーブル名
func (ManualArticle) TableName() string { return "manual_articles" }

// ManualArticleVersion はマニュアル編集履歴（編集ごとのスナップショット）
type ManualArticleVersion struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"     json:"id"`
	ArticleID       uint64    `gorm:"not null;index"               json:"article_id"`
	Title           string    `gorm:"not null"                     json:"title"`
	OrderValue      float64   `gorm:"column:order_value;not null"  json:"order_value"`
	Section         string    `gorm:"not null"                     json:"section"`
	BodyMarkdown    string    `gorm:"not null;type:text"           json:"body_markdown"`
	EditedByStaffID *uint64   `gorm:"column:edited_by_staff_id"    json:"edited_by_staff_id,omitempty"`
	EditedAt        time.Time `gorm:"autoCreateTime;column:edited_at" json:"edited_at"`
}

// TableName は ManualArticleVersion のテーブル名
func (ManualArticleVersion) TableName() string { return "manual_article_versions" }
