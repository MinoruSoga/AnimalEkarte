# CODE-QUALITY-230: FindAll ページネーション欠落 + ORDER BY 不統一

## 概要

`procedure_repository.go` / `vaccine_repository.go` の `FindAll` にページネーションがなく、
件数増大時のメモリ・レスポンス問題が発生しうる。
また `payment_method_master_repository.go` のソート列名が他マスタと不統一。

---

## 問題1（MEDIUM）: procedure / vaccine FindAll にページネーションなし

**ファイル:**
- `backend/internal/repository/procedure_repository.go` — `FindAll(ctx, clinicID)`
- `backend/internal/repository/vaccine_repository.go` — `FindAll(ctx, clinicID, species *string)`

### 現状コード（procedure）

```go
func (r *procedureRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Procedure, error) {
    var procedures []model.Procedure
    if err := r.db.WithContext(ctx).
        Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
        Order("sort_order ASC, name ASC").
        Find(&procedures).Error; err != nil {
        return nil, apperrors.FromGORM(err, "procedure", "")
    }
    return procedures, nil
}
```

### 問題

- `LIMIT` なし。clinic によって処置マスタが数百件になり得る。
- vaccine は `species` フィルタがあるが件数上限は変わらない。
- 他の類似マスタ（medicine 等）も同様の可能性があるが、
  少なくとも procedure / vaccine はユーザー数が多いほど増加するリスクが高い。

### 影響

- 大量データ返却 → フロントエンドのリスト描画が重くなる。
- メモリ使用量が予測困難。

### 修正案

短期: `Limit(500)` など上限を設けてログ警告を出す。

```go
if err := r.db.WithContext(ctx).
    Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
    Order("sort_order ASC, name ASC").
    Limit(500).
    Find(&procedures).Error; err != nil {
    ...
}
```

中期: handler/service に `pagination` を追加し、全マスタで統一。
（ただし、フロントエンドの select 系 UI はページネーションなしで全件取得が前提の箇所があるため、
API 設計変更は UI 側との協議が必要。）

---

## 問題2（LOW）: payment_method_master の ORDER BY が不統一

**ファイル:** `backend/internal/repository/payment_method_master_repository.go:35`

### 現状コード

```go
Order("display_order ASC, id ASC")
```

### 問題

他のマスタ（procedure, vaccine, medicine, exam_type, checkup_type, trimming_course 等）はすべて

```go
Order("sort_order ASC, name ASC")
```

を使用している。`display_order` カラムは `sort_order` に相当するものと思われるが、
カラム名・セカンダリキーの両方が異なる。

### 確認事項

1. `payment_method_masters` テーブルに `sort_order` カラムが存在するか、それとも `display_order` のみか。
2. `name` カラムが存在するか（ない場合はセカンダリキーの変更は不要）。

### 修正案（テーブル確認後）

- `display_order` → `sort_order` にカラム名を統一する（マイグレーション必要）か、
- ORDER BY のセカンダリキーを `name ASC` に変更する（`id` → `name`）

---

## 修正優先度

| 問題 | 優先度 | 理由 |
|------|--------|------|
| procedure/vaccine FindAll ページネーション欠落 | MEDIUM | 件数増大時のパフォーマンス劣化リスク |
| payment_method_master ORDER BY 不統一 | LOW | 機能上の問題なし。一貫性のための修正 |
