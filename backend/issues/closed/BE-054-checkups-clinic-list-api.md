# BE-054: 定期健診 クリニック横断一覧 API 実装

## 種類
機能追加（新規エンドポイント）

## 関連バグ
OPEN-BUG-017: 定期健診 `/checkups` ページが「ページが見つかりません」

## 背景
現在 checkup の取得 API は `/v1/medical-records/:id/checkups`（カルテ紐づき）のみ。
フロントエンドの `/checkups` ページでクリニック全体の定期健診一覧を表示するには、
クリニック横断で全健診レコードを取得できるエンドポイントが必要。

## 要件

### エンドポイント
```
GET /v1/checkups
```

### クエリパラメータ
| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `start_date` | string (YYYY-MM-DD) | 実施日 絞込 開始 |
| `end_date` | string (YYYY-MM-DD) | 実施日 絞込 終了 |
| `next_start_date` | string (YYYY-MM-DD) | 次回予定 絞込 開始 |
| `next_end_date` | string (YYYY-MM-DD) | 次回予定 絞込 終了 |

### レスポンス
```json
{
  "data": [
    {
      "id": "1",
      "medical_record_id": "10",
      "checkup_type_id": "2",
      "date": "2026-01-15",
      "next_date": "2027-01-15",
      "result": "異常なし",
      "checkup_type": { "id": "2", "name": "年次健診" },
      "medical_record": {
        "id": "10",
        "pet": {
          "id": "3",
          "name": "チョコ",
          "owner": { "id": "5", "name": "田中美咲" }
        }
      }
    }
  ]
}
```

## 実装指針

### handler
- `GET /v1/checkups` を `RegisterCheckupRoutes` (or 新規 `RegisterGlobalCheckupRoutes`) に追加
- clinic_id は JWT/session から取得（マルチテナント必須）
- クエリパラメータをパースして service 層に渡す

### service
```go
type ListCheckupsByClinicInput struct {
  ClinicID      uint64
  StartDate     *string
  EndDate       *string
  NextStartDate *string
  NextEndDate   *string
}

func (s *CheckupService) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, error)
```

### repository
```go
func (r *checkupRepository) ListByClinic(ctx context.Context, clinicID uint64, filters CheckupFilters) ([]model.Checkup, error)
```

GORM Preload:
- `CheckupType`
- `MedicalRecord.Pet.Owner`

### response 追加
`checkup_response.go` に `medical_record` ネスト（Pet.Owner まで）を追加。

## 影響ファイル
- `backend/internal/handler/checkup_handler.go` — 新ハンドラ関数追加
- `backend/internal/handler/checkup_response.go` — nested pet/owner フィールド追加
- `backend/internal/service/checkup_service.go` — ListByClinic 追加
- `backend/internal/repository/checkup_repository.go` — ListByClinic 追加
- `backend/internal/handler/handler.go` — ルート登録追加
