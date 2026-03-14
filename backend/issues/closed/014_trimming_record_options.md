# トリミング記録オプション（junction table）の動作確認と修正

## 概要
`trimming_record_options` テーブル（TrimmingRecord と TrimmingOption の多対多中間テーブル）が、トリミング記録の作成・更新時に正しく管理されているか確認し、未実装であれば実装する。

`model/trimming.go` には `TrimmingRecordOption` struct および `TrimmingRecord.Options []TrimmingOption`（many2many）が定義済み。リクエストで `option_ids: []string` を受け付けてオプションを設定できるかを確認する。

## 優先度
medium

## 関連テーブル
- `trimming_record_options` (id, trimming_record_id NOT NULL, option_id NOT NULL, sort_order int DEFAULT 0)
  - UNIQUE(trimming_record_id, option_id)（推奨）
- `trimming_records` (親テーブル)
- `trimming_options` (オプションマスタ)

## 実装内容

### モデル
`model/trimming.go` は実装済み。変更不要。

```go
type TrimmingRecord struct {
    ...
    Options []TrimmingOption `gorm:"many2many:trimming_record_options;joinForeignKey:TrimmingRecordID;joinReferences:OptionID" json:"options,omitempty"`
}

type TrimmingRecordOption struct {
    ID               uint64
    TrimmingRecordID uint64
    OptionID         uint64
    SortOrder        int
}
```

### リポジトリ
`repository/trimming_repository.go` の `CreateTrimmingRecord` / `UpdateTrimmingRecord` を確認し、以下を検証・追加:
- 作成時に `option_ids []uint64` を受け取り、`trimming_record_options` へ挿入する
- 更新時に既存の `trimming_record_options` を全削除して再挿入する（`DELETE WHERE trimming_record_id = ?` → 再 INSERT）

GORM の `many2many` アソシエーションを使う場合は `db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&record)` は避け、明示的に `Association("Options").Replace(options)` または手動 DELETE/INSERT を行う。

`repository/trimming_repository.go` に以下を追加（存在しない場合）:
```go
func (r *trimmingRepository) SetOptions(ctx context.Context, trimmingRecordID uint64, optionIDs []uint64) error
```
- `DELETE FROM trimming_record_options WHERE trimming_record_id = ?`
- `INSERT INTO trimming_record_options (trimming_record_id, option_id, sort_order) VALUES ...`（バッチ挿入、sort_order はスライスのインデックス）
- トランザクション内で実行

### サービス
`service/trimming_service.go` の `CreateTrimmingRecord` / `UpdateTrimmingRecord` を確認し:
- `CreateTrimmingInput.OptionIDs []uint64` が存在するか確認・追加
- `UpdateTrimmingInput.OptionIDs *[]uint64`（nil = 変更なし、空スライス = 全削除）が存在するか確認・追加
- `Create` / `Update` 後に `repo.Trimming.SetOptions()` を呼び出す

### ハンドラ
`handler/trimming_request.go` を確認し:
- `createTrimmingRecordRequest.OptionIDs []uint64` が存在するか確認・追加
- `updateTrimmingRecordRequest.OptionIDs *[]uint64` が存在するか確認・追加

`handler/trimming_response.go` を確認し:
- `trimmingRecordResponse.Options []trimmingOptionSummaryResponse` が含まれているか確認・追加

`handler/trimming_handler.go` の `CreateTrimmingRecord` / `UpdateTrimmingRecord` で `option_ids` を service input に正しく渡しているか確認。

### ルート登録
変更不要。既存のトリミングルートを使用。

## 完了条件
- `POST /v1/trimming-records` で `option_ids: [1, 2, 3]` を渡すとオプションが関連付けられる
- `PATCH /v1/trimming-records/:id` で `option_ids: [4, 5]` を渡すと既存オプションが置き換えられる
- `PATCH /v1/trimming-records/:id` で `option_ids: []` を渡すと全オプションが削除される
- `PATCH /v1/trimming-records/:id` で `option_ids` を送らない場合、既存オプションは変更されない
- `GET /v1/trimming-records/:id` のレスポンスに `options` フィールドが含まれる
- 存在しない `option_id` を指定すると 400 エラーを返す（バリデーション推奨）
