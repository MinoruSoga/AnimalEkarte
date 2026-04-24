# BUG-382: merchandise_item_handler で全件取得にページネーション引数をハードコード

## 概要
`merchandise_item_handler.go` の `ListMerchandiseItems` は、サービスの `List` メソッドがページネーション引数（page, limit）を要求するにもかかわらず、`page=1, limit=10000` をハードコードして全件取得している。これはサービスインターフェースの設計と実際の使用意図の乖離であり、データが 10,000 件を超えた場合に問題が発生する。また、ページネーション有無の戦略がマスタ間で統一されていない。

## 再現手順
1. `GET /v1/masters/merchandise-items` を実行
2. **結果**: `page=1, limit=10000` で GORM クエリが発行される（DB レベルで LIMIT 10000）
3. 比較: `GET /v1/masters/medicines` → 正規のページネーションで `parsePagination(c)` を使用

## 期待する動作
- マスタデータを全件取得する場合、`List(ctx, clinicID, category)` のようにページネーション引数を持たないシグネチャを使用すること
- または、全マスタで統一されたページネーション方針（全件取得 or ページネーション）を決定し適用すること

## 現状コード

### `backend/internal/handler/merchandise_item_handler.go:25-26`（ハードコード全件取得）
```go
// page=1, limit=10000 で全件取得（マスタデータは件数が限定的）
items, _, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, 1, 10000, category)
```

### 比較: 正しい実装例 A（medicine_handler.go — 正規ページネーション）
```go
// backend/internal/handler/medicine_handler.go:21-32
page, limit, err := parsePagination(c)
if err != nil {
    RespondError(c, err)
    return
}
medicines, total, err := h.svc.Medicine.List(c.Request.Context(), clinicID, page, limit)
```

### 比較: 正しい実装例 B（cage_handler.go — ページネーションなし全件取得）
```go
// backend/internal/handler/cage_handler.go
cages, err := h.svc.Cage.List(c.Request.Context(), clinicID)
// サービスシグネチャ: List(ctx, clinicID) — ページネーション引数なし
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/merchandise_item_handler.go:25-26` | ハードコード全件取得 | 要修正 |
| `backend/internal/service/merchandise_item_service.go` | List シグネチャ（ページネーション引数の要否） | 要検討・修正 |
| `backend/internal/repository/merchandise_item_repository.go` | FindAll シグネチャ | 要検討・修正 |

## 修正方針

マスタデータ（merchandise_item）の件数は業務上限定的（数百件以内）であるため、全件取得シグネチャに統一する。

### 1. `backend/internal/service/merchandise_item_service.go` — シグネチャ変更
```go
// インターフェース変更
type MerchandiseItemService interface {
    List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
    // ページネーション引数・total 戻り値を削除
    // ...
}

// 実装変更
func (s *merchandiseItemService) List(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error) {
    items, err := s.repo.MerchandiseItem.FindAll(ctx, clinicID, category)
    if err != nil {
        return nil, apperrors.Wrap(err, "商品マスタ一覧の取得に失敗しました")
    }
    return items, nil
}
```

### 2. `backend/internal/handler/merchandise_item_handler.go:25-26`
```go
// 修正前
items, _, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, 1, 10000, category)

// 修正後
items, err := h.svc.MerchandiseItem.List(c.Request.Context(), clinicID, category)
```

### 3. `backend/internal/repository/merchandise_item_repository.go` — FindAll シグネチャも整合
```go
// インターフェース
FindAll(ctx context.Context, clinicID uint64, category string) ([]model.MerchandiseItem, error)
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — アーキテクチャ遵守
> handler → service → repository の軽量レイヤードを徹底。インターフェースの設計意図とハンドラの使用方法を一致させる。

### プロジェクト内参照実装
`backend/internal/handler/cage_handler.go` — ページネーションなし全件取得の正しいパターン（サービスシグネチャと一致）

## 優先度
**Medium** — 現状は 10,000 件上限で機能するが、インターフェース設計の乖離がテストと保守を困難にする。サービス層のシグネチャ変更が必要なため、影響範囲の確認が必要。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/merchandise_item_handler.go:25-26` — 問題箇所
- `backend/internal/service/merchandise_item_service.go` — シグネチャ変更対象
- `backend/internal/repository/merchandise_item_repository.go` — シグネチャ変更対象
- `backend/internal/service/merchandise_item_service_test.go` — テスト更新が必要
