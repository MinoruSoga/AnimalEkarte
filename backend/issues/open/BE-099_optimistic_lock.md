# BE-099: 楽観的ロック（version フィールド）実装

## 概要
複数ユーザーが同一レコードを同時編集した場合の後勝ち問題を解決するため、
楽観的ロック（version フィールド）をカルテ（medical_records）に実装する。

## 変更内容

### 1. migration: version カラム追加

```sql
-- backend/migrations/001_init.sql を直接編集
ALTER TABLE medical_records ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
```

### 2. model 更新

```go
// backend/internal/model/medical_record.go
type MedicalRecord struct {
  // ... 既存フィールド
  Version int `gorm:"default:1" json:"version"`
}
```

### 3. handler: リクエストに version を追加

```go
// backend/internal/handler/medical_record_request.go
type UpdateMedicalRecordRequest struct {
  // ... 既存フィールド
  Version *int `json:"version"` // 楽観的ロック用
}
```

### 4. service: 競合チェック

```go
// backend/internal/service/medical_record_service.go
func (s *MedicalRecordService) Update(ctx context.Context, id uint64, input UpdateMedicalRecordInput) (*model.MedicalRecord, error) {
  record, err := s.repo.GetByID(ctx, id)
  if err != nil {
    return nil, apperrors.Wrap(err, "failed to get medical record")
  }

  // 楽観的ロックチェック
  if input.Version != nil && record.Version != *input.Version {
    return nil, apperrors.NewConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")
  }

  // バージョンインクリメント
  fields["version"] = record.Version + 1
  // ... 以降の更新処理
}
```

### 5. レスポンスに version を含める

GET /v1/medical-records/:id のレスポンスに `version` フィールドを含める。

## フロントエンド対応
BE 完了後に FE 側で:
1. カルテ取得時に `version` を保持
2. 更新リクエストに `version` を付与
3. 409 Conflict レスポンス時に「再読み込みしてください」トーストを表示

## 優先度
Low

## 関連
- docs/tasks/open/feature/BUG-099_optimistic_lock.md
