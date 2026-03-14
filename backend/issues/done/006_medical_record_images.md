# 診療画像 CRUD 実装

## 概要
外来カルテへの診療画像（レントゲン・エコー・写真等）添付・取得・削除 API を実装する。
`model/record_image.go` の `RecordImage` struct は実装済み。handler・service・repository が未実装。

ファイル実体の保存先（ローカルストレージ or S3）は未決定のため、まず URL ベース（`image_url` をクライアントが指定する）で実装する。

## 優先度
medium

## 関連テーブル
- `record_images` (id, medical_record_id NOT NULL, image_url text NOT NULL DEFAULT '', thumbnail_url text DEFAULT '', file_name text DEFAULT '', file_size bigint, mime_type text DEFAULT '', image_type medical_image_type DEFAULT 'other', description text DEFAULT '', taken_at timestamp, exam_id, staff_id, sort_order int DEFAULT 0, created_at, updated_at)
- `medical_records` (親テーブル)
- `examination_records` (exam_id の参照先)
- `staffs` (staff_id の参照先)

## 実装内容

### モデル
`model/record_image.go` は実装済み。変更不要。

`MedicalImageType` enum: `xray` / `echo` / `photo` / `endoscope` / `ct` / `mri` / `microscope` / `other`

### リポジトリ
新規ファイル `repository/record_image_repository.go`:
```go
type RecordImageRepository interface {
    ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error)
    Create(ctx context.Context, image *model.RecordImage) error
    Delete(ctx context.Context, id uint64) error
    FindByID(ctx context.Context, id uint64) (*model.RecordImage, error)
}
```
- `ListByMedicalRecordID`: `ORDER BY sort_order ASC, created_at ASC`
- `Delete`: 物理削除（record_images テーブルに deleted_at なし）

`repository/repositories.go` に `RecordImage RecordImageRepository` を追加。

### サービス
新規ファイル `service/record_image_service.go`:
```go
type CreateRecordImageInput struct {
    ImageURL     string
    ThumbnailURL string
    FileName     string
    FileSize     int64
    MimeType     string
    ImageType    model.MedicalImageType
    Description  string
    TakenAt      *time.Time
    ExamID       *uint64
    StaffID      *uint64
    SortOrder    int
}

type RecordImageService interface {
    List(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error)
    Create(ctx context.Context, medicalRecordID uint64, input *CreateRecordImageInput) (*model.RecordImage, error)
    Delete(ctx context.Context, medicalRecordID, imageID uint64) error
}
```

`service/validators.go` に `validateMedicalImageType(imageType string) error` を追加。

### ハンドラ
新規ファイル `handler/record_image_handler.go`:
```go
func (h *Handler) ListRecordImages(c *gin.Context)
func (h *Handler) CreateRecordImage(c *gin.Context)
func (h *Handler) DeleteRecordImage(c *gin.Context)
func (h *Handler) RegisterRecordImageRoutes(rg *gin.RouterGroup)
```

新規ファイル `handler/record_image_request.go`:
```go
type createRecordImageRequest struct {
    ImageURL     string     `json:"image_url"     binding:"required"`
    ThumbnailURL string     `json:"thumbnail_url"`
    FileName     string     `json:"file_name"`
    FileSize     int64      `json:"file_size"`
    MimeType     string     `json:"mime_type"`
    ImageType    string     `json:"image_type"`
    Description  string     `json:"description"`
    TakenAt      *time.Time `json:"taken_at"`
    ExamID       *uint64    `json:"exam_id"`
    StaffID      *uint64    `json:"staff_id"`
    SortOrder    int        `json:"sort_order"`
}
```

新規ファイル `handler/record_image_response.go`:
```go
type recordImageResponse struct {
    ID              uint64     `json:"id"`
    MedicalRecordID uint64     `json:"medical_record_id"`
    ImageURL        string     `json:"image_url"`
    ThumbnailURL    string     `json:"thumbnail_url"`
    FileName        string     `json:"file_name"`
    FileSize        int64      `json:"file_size"`
    MimeType        string     `json:"mime_type"`
    ImageType       string     `json:"image_type"`
    Description     string     `json:"description"`
    TakenAt         *time.Time `json:"taken_at,omitempty"`
    ExamID          *uint64    `json:"exam_id,omitempty"`
    StaffID         *uint64    `json:"staff_id,omitempty"`
    SortOrder       int        `json:"sort_order"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    Staff           *staffSummaryResponse `json:"staff,omitempty"`
}
```

### ルート登録
`cmd/api/main.go` の医療記録グループに追加:
```go
medRecords.GET("/:id/images",             h.ListRecordImages)
medRecords.POST("/:id/images",            h.CreateRecordImage)
medRecords.DELETE("/:id/images/:imageId", h.DeleteRecordImage)
```

## 完了条件
- `GET /v1/medical-records/:id/images` が画像一覧を `sort_order`, `created_at` 昇順で返す
- `POST /v1/medical-records/:id/images` で URL ベースの画像レコードを作成できる（`image_url` 必須）
- `DELETE /v1/medical-records/:id/images/:imageId` で画像レコードを削除できる
- 存在しない `medical_record_id` に対しては 404 を返す
- `imageId` が指定の medical record に属さない場合も 404 を返す
- 不正な `image_type` は 400 を返す
