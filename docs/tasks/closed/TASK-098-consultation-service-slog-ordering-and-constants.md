# TASK-098: consultation_service — slog フィールド順序違反 + buildUpdateFields 定数未使用

## 優先度

MEDIUM

---

## 概要

`consultation_service.go` で 2 種類の規約違反が存在する。

1. slog の clinic_id フィールドが FIRST でない操作がある（Create/Delete）
2. `buildConsultationUpdateFields` が裸の文字列リテラルを使用（定数未定義）

---

## 問題箇所

### 1. slog フィールド順序（L88, L120）

```go
// ❌ L88 (Create): clinic_id が SECOND
slog.InfoContext(ctx, "consultation created",
    slog.Uint64("consultation_id", consultation.ID), // ← entity ID が FIRST（違反）
    slog.Uint64("clinic_id", clinicID))

// ✅ L106 (Update): 正しい順序
slog.InfoContext(ctx, "consultation updated",
    slog.Uint64("clinic_id", clinicID),              // ← FIRST ✅
    slog.Uint64("consultation_id", id))

// ❌ L120 (Delete): clinic_id が SECOND
slog.InfoContext(ctx, "consultation deleted",
    slog.Uint64("consultation_id", id),              // ← entity ID が FIRST（違反）
    slog.Uint64("clinic_id", clinicID))
```

Update（L106）は正しいパターンだが、Create と Delete が逆順になっている。

---

### 2. buildConsultationUpdateFields: 裸の文字列リテラル（L152-186）

```go
// ❌ 全フィールドが定数なし
func buildConsultationUpdateFields(input *UpdateConsultationInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil {
        fields["name"] = *input.Name            // ← 定数なし
    }
    if input.Price != nil {
        fields["price"] = *input.Price          // ← 定数なし
    }
    if input.IsActive != nil {
        fields["is_active"] = *input.IsActive   // ← 定数なし
    }
    // ... description, time_condition, duration, parent_id, sort_order, tax_type, tax_rate も同様
```

---

## 修正方針

```go
// ✅ 定数を追加
const (
    colConsultationName          = "name"
    colConsultationPrice         = "price"
    colConsultationIsActive      = "is_active"
    colConsultationDescription   = "description"
    colConsultationTimeCondition = "time_condition"
    colConsultationDuration      = "duration"
    colConsultationParentID      = "parent_id"
    colConsultationSortOrder     = "sort_order"
    colConsultationTaxType       = "tax_type"
    colConsultationTaxRate       = "tax_rate"
)

// ✅ slog 修正後（Create/Delete）
slog.InfoContext(ctx, "consultation created",
    slog.Uint64("clinic_id", clinicID),          // FIRST
    slog.Uint64("consultation_id", consultation.ID))
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `service/consultation_service.go` | `const col...` ブロック追加、Create/Delete slog の clinic_id を FIRST に変更 |

---

## 関連

- TASK-080: permission_group_service の同種問題（bare string literals）
- TASK-097: shift_template_service の同種問題
