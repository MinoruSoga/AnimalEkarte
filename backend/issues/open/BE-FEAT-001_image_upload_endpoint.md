# BE-FEAT-001: カルテ画像アップロード API 実装

## 概要
カルテの画像タブに対応するため、画像アップロード・取得・削除エンドポイントを実装する。

## エンドポイント設計

| メソッド | パス | 説明 |
|---------|------|------|
| POST   | /v1/medical-records/:id/images | 画像アップロード（multipart/form-data） |
| GET    | /v1/medical-records/:id/images | 画像一覧取得 |
| DELETE | /v1/medical-records/:id/images/:image_id | 画像削除 |

## テーブル設計

```sql
CREATE TABLE medical_record_images (
  id          BIGSERIAL PRIMARY KEY,
  clinic_id   BIGINT NOT NULL,
  record_id   BIGINT NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  file_name   VARCHAR(255) NOT NULL,
  file_size   BIGINT NOT NULL,
  mime_type   VARCHAR(100) NOT NULL,
  storage_key VARCHAR(500) NOT NULL,  -- ローカルファイルパスまたは S3 キー
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_images_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id)
);
CREATE INDEX idx_images_record ON medical_record_images(record_id);
```

## バリデーション
- ファイルサイズ: 10MB 以下
- 許可 MIME: `image/jpeg`, `image/png`, `image/gif`, `application/pdf`
- 1カルテあたり最大 50 ファイル

## ストレージ戦略
- 開発環境: `/app/uploads/medical-records/{record_id}/` にローカル保存
- 本番環境: S3 バケットへのアップロード（環境変数で切り替え）

## レスポンス型

```go
type MedicalRecordImageResponse struct {
  ID        uint64 `json:"id"`
  FileName  string `json:"file_name"`
  FileSize  int64  `json:"file_size"`
  MimeType  string `json:"mime_type"`
  URL       string `json:"url"`  // ダウンロード用 URL
  CreatedAt string `json:"created_at"`
}
```

## モデルファイル
- `backend/internal/model/medical_record_image.go`（新規）
- `backend/internal/handler/medical_record_image_handler.go`（新規）
- `backend/internal/service/medical_record_image_service.go`（新規）
- `backend/internal/repository/medical_record_image_repository.go`（新規）

## 優先度
Medium

## 関連
- docs/tasks/open/feature/FEAT-001_image_upload.md
