# TASK-027: その他マスタ HIGH 問題 2件（Reorder slog 欠落 / extractClinicID 順序逆転）

## 優先度

HIGH

---

## 問題 1: Reorder の slog.InfoContext が 4 ドメインで欠落

### ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/service/animal_species_service.go` | L125-132（Reorder 後ログなし） |
| `backend/internal/service/occupation_service.go` | L108-116（Reorder 後ログなし） |
| `backend/internal/service/insurance_service.go` | L86-94（Reorder 後ログなし） |
| `backend/internal/service/consultation_service.go` | L89-97（Reorder 後ログなし） |

### 問題
Create / Update / Delete には `slog.InfoContext` があるが、Reorder だけ全ドメインでログなし。他ミューテーション操作との一貫性が破れている。

### 修正案（統一パターン）
```go
func (s *xxxService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder xxx")
    }
    slog.InfoContext(ctx, "xxx reordered",
        slog.Uint64("clinic_id", clinicID),
        slog.Int("count", len(ids)))
    return nil
}
```

**注**: exam_type / checkup_type / diagnosis_type / chief_complaint_type の Reorder も同様にログなし（TASK-029 参照）。全マスタ系 Reorder へ一斉に追加することを推奨。

---

## 問題 2: UpdateTrimmingCourse / UpdateTrimmingOption の extractClinicID と parseIDParam の呼び出し順序が逆

### ファイル
`backend/internal/handler/trimming_master_handler.go:84-117`（UpdateTrimmingCourse）
`backend/internal/handler/trimming_master_handler.go:220-251`（UpdateTrimmingOption）

### 問題
```go
// 現状: id を先に取得（他ハンドラと逆順）
id, ok := parseIDParam(c, "id")
if !ok { return }
clinicID, ok := extractClinicID(c)
if !ok { return }
```

全ての他ハンドラは `clinicID → id` の順で統一されている。この2メソッドだけが逆順で、コードレビュー時の混乱とコピーペーストバグの温床になる。

### 修正案
```go
// 修正後: clinicID を先に（全ハンドラ統一）
clinicID, ok := extractClinicID(c)
if !ok { return }
id, ok := parseIDParam(c, "id")
if !ok { return }
```
