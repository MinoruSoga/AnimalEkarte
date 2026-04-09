# TASK: 画像アップロードを S3 に移行（staging 環境）

## 概要

現在、診療画像のアップロードはコンテナ内ローカルファイルシステム (`/app/uploads/medical-records/`) に保存している。
ECS Fargate はエフェメラルストレージであり、**タスク再起動のたびに画像が全て消失する**。
staging 環境でまず S3 へ移行し、本番デプロイ前に検証する。

## 現状の問題

1. **データ消失**: ECS タスクの再デプロイ・再起動で `/app/uploads/` が初期化される
2. **スケール不可**: Fargate タスクを複数起動すると、タスク間でファイルが共有されない
3. **バックアップ不可**: ローカルストレージはバックアップ対象外

## ユーザーフロー

カルテ画面の「画像」タブから画像をアップロードする。

1. カルテ詳細画面 → 画像タブ（`MedicalRecordImage` コンポーネント）
2. `ImageGalleryFilter` の「ファイルを選択」ボタンでファイルを選択
3. `useUploadImages` → `POST /v1/medical-records/:id/images/upload`（multipart/form-data）
4. バックエンド `UploadRecordImage` がファイルをローカル保存 + DB レコード作成
5. 画像一覧を再取得（`useGetRecordImages`）→ `img.image_url` で `<img>` に表示

## 現状コード

### フロントエンド（変更不要）

| ファイル | 役割 |
|---------|------|
| `frontend/src/features/medical-records/components/MedicalRecordImage.tsx` | 画像タブ本体。アップロード・削除・フィルタ |
| `frontend/src/features/medical-records/api/record-images.ts:13-25` | `uploadImage` — `FormData` で `POST /v1/.../images/upload` |
| `frontend/src/features/medical-records/api/record-images.ts:46-53` | `deleteImage` — `DELETE /v1/.../images/:imageId` |
| `frontend/src/features/medical-records/api/get-record-images.ts` | 画像一覧取得。`img.image_url` を `src` に設定 |
| `frontend/src/features/medical-records/components/ImageGalleryGroup.tsx` | 画像グループ表示 |
| `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx` | 検索・アップロードボタン |

### バックエンド

| ファイル | 役割 |
|---------|------|
| `backend/internal/handler/record_image_handler.go:22-25` | `uploadsBaseDir = "/app/uploads/medical-records"` 定数 |
| `backend/internal/handler/record_image_handler.go:152-250` | `UploadRecordImage` — `c.SaveUploadedFile()` でローカル保存 |
| `backend/internal/handler/record_image_handler.go:228` | URL 生成: `/uploads/medical-records/{id}/{filename}` |
| `backend/internal/service/record_image_service.go` | RecordImage CRUD（ストレージ層に依存しない） |
| `backend/internal/model/record_image.go` | `ImageURL string` フィールド |

### AWS

| リソース | 現状 |
|---------|------|
| S3 バケット（画像用） | **未作成** |
| AWS SDK for Go | **未導入**（go.mod に `aws-sdk-go` なし） |
| ECS タスクロール | S3 権限なし |

## 修正方針

### Phase 1: S3 バケット作成 + IAM 設定

```bash
# S3 バケット作成
aws s3 mb s3://animalekarte-stg-uploads --region us-east-1

# バケットポリシー: パブリックアクセスブロック（全て有効）
aws s3api put-public-access-block \
  --bucket animalekarte-stg-uploads \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

# CORS 設定（フロントエンドからの直接表示用）
aws s3api put-bucket-cors --bucket animalekarte-stg-uploads --cors-configuration '{
  "CORSRules": [{
    "AllowedOrigins": ["https://stg.noah-karte.com", "http://localhost:3000"],
    "AllowedMethods": ["GET"],
    "AllowedHeaders": ["*"],
    "MaxAgeSeconds": 86400
  }]
}'
```

### Phase 2: ECS タスクロールに S3 権限追加

```json
{
  "Effect": "Allow",
  "Action": [
    "s3:PutObject",
    "s3:GetObject",
    "s3:DeleteObject"
  ],
  "Resource": "arn:aws:s3:::animalekarte-stg-uploads/*"
}
```

### Phase 3: バックエンド実装変更

#### 3a. AWS SDK 導入

```bash
docker compose exec backend go get github.com/aws/aws-sdk-go-v2
docker compose exec backend go get github.com/aws/aws-sdk-go-v2/config
docker compose exec backend go get github.com/aws/aws-sdk-go-v2/service/s3
```

#### 3b. S3 アップローダー作成

```go
// internal/infra/s3_uploader.go
package infra

type FileUploader interface {
    Upload(ctx context.Context, key string, body io.Reader, contentType string) (url string, err error)
    Delete(ctx context.Context, key string) error
}

type S3Uploader struct {
    client *s3.Client
    bucket string
    region string
}

func (u *S3Uploader) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
    _, err := u.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      &u.bucket,
        Key:         &key,
        Body:        body,
        ContentType: &contentType,
    })
    if err != nil {
        return "", fmt.Errorf("s3 upload failed: %w", err)
    }
    return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", u.bucket, u.region, key), nil
}
```

#### 3c. UploadRecordImage ハンドラー修正

```go
// Before: c.SaveUploadedFile(fileHeader, storedPath)
// After:
key := fmt.Sprintf("medical-records/%d/%s", medicalRecordID, storedName)
imageURL, err := h.uploader.Upload(c.Request.Context(), key, file, mimeType)
```

#### 3d. 環境変数

```env
# .env.staging に追加
S3_BUCKET=animalekarte-stg-uploads
S3_REGION=us-east-1
```

### Phase 4: 画像削除時の S3 オブジェクト削除

`DeleteRecordImage` ハンドラーに `h.uploader.Delete()` を追加する。
現状はDB レコードの論理削除のみで、ファイル自体は削除していない。

### Phase 5: フロントエンド

`image_url` が S3 の完全 URL になるため、フロントエンド側の変更は不要。
`img.image_url` がそのまま `<img src={...}>` で表示される。

## 環境別設計

| 環境 | ストレージ | 設定 |
|------|-----------|------|
| ローカル開発 | ローカルファイルシステム（現状維持） | `STORAGE_TYPE=local` |
| staging | S3 | `STORAGE_TYPE=s3`, `S3_BUCKET=animalekarte-stg-uploads` |
| production | S3 | `STORAGE_TYPE=s3`, `S3_BUCKET=animalekarte-prod-uploads` |

`STORAGE_TYPE` 環境変数で `FileUploader` インターフェースの実装を切り替える（DI）。

## テスト計画

- [ ] S3 バケット作成 + パブリックアクセスブロック確認
- [ ] ECS タスクロールに S3 権限追加確認
- [ ] ステージングで画像アップロード → S3 に保存されることを確認
- [ ] アップロード後の画像が `<img src="https://...s3...">` で表示されることを確認
- [ ] 画像削除時に S3 オブジェクトも削除されることを確認
- [ ] ECS タスク再起動後も画像が消えないことを確認
- [ ] ローカル開発環境が引き続きローカルストレージで動作することを確認

## 優先度

**High** — ステージング環境でデプロイ毎に画像が消失する実害がある。

## 影響範囲

| 対象 | 変更 |
|------|------|
| `backend/internal/handler/record_image_handler.go` | S3 アップロード/削除に変更 |
| `backend/internal/infra/s3_uploader.go` | 新規作成 |
| `backend/internal/infra/local_uploader.go` | 新規作成（ローカル用） |
| `backend/cmd/api/main.go` | FileUploader の DI 配線 |
| `backend/go.mod` | aws-sdk-go-v2 追加 |
| `backend/.env.staging` | S3_BUCKET, S3_REGION, STORAGE_TYPE 追加 |
| AWS: S3 バケット | 新規作成 |
| AWS: ECS タスクロール | S3 権限追加 |
| フロントエンド | **変更不要** |
