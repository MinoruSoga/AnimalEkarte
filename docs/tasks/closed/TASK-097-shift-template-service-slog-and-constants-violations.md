# TASK-097: shift_template_service — slog フィールド順序・命名 + buildUpdateFields 定数未使用

## 優先度

MEDIUM

---

## 概要

`shift_template_service.go` で 3 種類の規約違反が存在する。

1. slog の clinic_id フィールド順序違反（Create のみ逆順）
2. slog のエンティティ ID フィールド名が `"id"` で非準拠（全操作）
3. `buildShiftTemplateUpdateFields` が裸の文字列リテラルを使用（定数未定義）

---

## 問題箇所

### 1. slog フィールド順序・命名（L96, L127, L139）

```go
// ❌ L96 (Create): clinic_id が 2番目、フィールド名 "id" は非準拠
slog.InfoContext(ctx, "shift template created",
    slog.Uint64("id", tpl.ID),          // ← entity ID が FIRST（規約違反）
    slog.Uint64("clinic_id", clinicID)) // ← clinic_id が SECOND

// ❌ L127 (Update): clinic_id は FIRST だが "id" は非準拠
slog.InfoContext(ctx, "shift template updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("id", id))              // ← "shift_template_id" であるべき

// ❌ L139 (Delete): 同上
slog.InfoContext(ctx, "shift template deleted",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("id", id))              // ← "shift_template_id" であるべき
```

規約: `clinic_id` は ALWAYS FIRST、エンティティ ID は `{ドメイン名}_id` 形式。

---

### 2. Reorder に slog がない（L143-151）

```go
// ❌ Reorder 完了ログなし
func (s *shiftTemplateService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    // ...
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil { ... }
    return nil  // ← slog.InfoContext なし
}
```

---

### 3. buildShiftTemplateUpdateFields: 裸の文字列リテラル（L153-177）

```go
// ❌ 全フィールドが裸のリテラル
func buildShiftTemplateUpdateFields(input *UpdateShiftTemplateInput) map[string]any {
    fields := map[string]any{}
    if input.Name != nil {
        fields["name"] = *input.Name          // ← 定数なし
    }
    if input.ShiftType != nil {
        fields["shift_type"] = *input.ShiftType // ← 定数なし
    }
    if input.StartTime != nil {
        fields["start_time"] = ...             // ← 定数なし
    }
    // ... end_time, notes, sort_order, is_active も同様
```

---

## 修正方針

```go
// ✅ 定数を追加
const (
    colShiftTemplateName      = "name"
    colShiftTemplateShiftType = "shift_type"
    colShiftTemplateStartTime = "start_time"
    colShiftTemplateEndTime   = "end_time"
    colShiftTemplateNotes     = "notes"
    colShiftTemplateSortOrder = "sort_order"
    colShiftTemplateIsActive  = "is_active"
)

// ✅ slog 修正後
slog.InfoContext(ctx, "shift template created",
    slog.Uint64("clinic_id", clinicID),          // FIRST
    slog.Uint64("shift_template_id", tpl.ID))    // SECOND

// ✅ Reorder に slog 追加
slog.InfoContext(ctx, "shift templates reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `service/shift_template_service.go` | `const col...` ブロック追加、slog の clinic_id を FIRST に、フィールド名を `"shift_template_id"` に統一、Reorder slog 追加 |

---

## 関連

- TASK-080: permission_group_service の同種問題（bare string literals）
- TASK-089: slog エンティティ ID フィールド命名不統一
