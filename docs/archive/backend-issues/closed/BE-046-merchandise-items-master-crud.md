**Status**: Closed

# BE-046: 物販・その他マスタテーブル追加 + CRUD API

## 背景

会計の「物販・その他追加」モーダルで、現在は品目名・単価を手動入力している。
これをマスタ化し、マスタから選択する方式に変更する。

## 要件

### 1. DB: `merchandise_items` テーブル追加（`001_init.sql` に直接追記）

```sql
CREATE TABLE merchandise_items (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        text        NOT NULL DEFAULT '',
    category    item_category NOT NULL DEFAULT 'goods',
    unit_price  numeric     NOT NULL DEFAULT 0,
    tax_rate    numeric     NOT NULL DEFAULT 0.10,
    is_active   boolean     NOT NULL DEFAULT true,
    sort_order  integer     NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE INDEX idx_merchandise_items_clinic ON merchandise_items(clinic_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_merchandise_items_category ON merchandise_items(clinic_id, category) WHERE deleted_at IS NULL;
```

### 2. Go Model: `backend/internal/model/merchandise_item.go`

```go
type MerchandiseItem struct {
    ID        uint64         `gorm:"primaryKey" json:"id"`
    ClinicID  uint64         `json:"clinic_id"`
    Name      string         `json:"name"`
    Category  ItemCategory   `json:"category" gorm:"type:item_category"`
    UnitPrice float64        `json:"unit_price"`
    TaxRate   float64        `json:"tax_rate"`
    IsActive  bool           `json:"is_active"`
    SortOrder int            `json:"sort_order"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
```

### 3. CRUD API エンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | `/v1/merchandise-items` | 一覧取得（clinic_id フィルタ必須、category フィルタ任意） |
| GET | `/v1/merchandise-items/:id` | 詳細取得 |
| POST | `/v1/merchandise-items` | 新規作成 |
| PATCH | `/v1/merchandise-items/:id` | 更新（ポインタ型 + buildUpdateFields） |
| DELETE | `/v1/merchandise-items/:id` | 論理削除 |
| POST | `/v1/merchandise-items/reorder` | 並べ替え |

### 4. Handler / Service / Repository

- 既存のマスタ CRUD パターン（medicines, procedures 等）に準拠
- handler: `*_request.go` でバインド → `service.XxxInput` に変換
- service: バリデーション + ビジネスロジック
- repository: GORM クエリ（clinic_id フィルタ必須）

### 5. Seed データ（`002_seed_master.sql` に追記）

```sql
INSERT INTO merchandise_items (clinic_id, name, category, unit_price, tax_rate, sort_order) VALUES
(1, 'ロイヤルカナン 消化器サポート 1kg', 'food', 2800, 0.10, 1),
(1, 'ヒルズ k/d 2kg', 'food', 3500, 0.10, 2),
(1, 'ペット用歯ブラシセット', 'goods', 1200, 0.10, 3),
(1, 'エリザベスカラー（S）', 'goods', 800, 0.10, 4),
(1, 'ノミ・ダニ予防首輪', 'goods', 1500, 0.10, 5),
(1, '文書料', 'other', 3000, 0.10, 6),
(1, '時間外診療費', 'other', 5000, 0.10, 7)
ON CONFLICT DO NOTHING;
```

### 6. `make codegen` 実行

モデル追加後に `models.ts` を再生成すること。

## 実装パターン参照

- `backend/internal/handler/medicine_handler.go`（マスタCRUDの参照実装）
- `backend/internal/model/medicine.go`
- `backend/CLAUDE.md`（レイヤードアーキテクチャルール）

## 受入条件

- [x] merchandise_items テーブルが 001_init.sql に追加
- [x] Go Model 作成
- [x] 全6エンドポイント実装
- [x] clinic_id フィルタ必須
- [x] Seed データ投入
- [x] `make codegen` で models.ts 更新

## クローズ情報

- **Closed At**: 2026-03-19
- **変更ファイル**:
  - `backend/migrations/001_init.sql` — merchandise_items テーブル + インデックス追加
  - `backend/migrations/002_seed_master.sql` — 7件のシードデータ追加
  - `backend/internal/model/merchandise_item.go` — Go モデル（新規）
  - `backend/internal/handler/merchandise_item_handler.go` — 6 エンドポイント（新規）
  - `backend/internal/handler/merchandise_item_request.go` — リクエスト型（新規）
  - `backend/internal/handler/merchandise_item_response.go` — レスポンス型（新規）
  - `backend/internal/service/merchandise_item_service.go` — サービス層（新規）
  - `backend/internal/repository/merchandise_item_repository.go` — リポジトリ層（新規）
  - `backend/internal/service/service.go` — DI 登録追加
  - `backend/internal/repository/repositories.go` — DI 登録追加
  - `backend/internal/handler/staff_handler.go` — ルート登録追加
  - `frontend/src/types/generated/models.ts` — codegen 更新
