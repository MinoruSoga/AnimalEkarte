# CODE-QUALITY-207: マスタ系 軽微パターン不統一まとめ

## 概要

各ドメインで発見された比較的軽微なパターン不統一・冗長コード・命名問題をまとめる。
個別に起票するほどではないが、蓄積するとコードベースの一貫性が損なわれる問題群。

## 優先度

MEDIUM（全項目）

---

## 項目一覧

### 1. cage_service.go:108-113 — CageType/CageSize バリデーションの二重チェック

**ファイル**: `backend/internal/service/cage_service.go`

`createCageRequest` の `binding:"required,oneof=..."` で Gin が Handler 段階で既にバリデーション済みにもかかわらず、Service の `Create` でも `validateCageType` / `validateCageSize` を再度呼んでいる。

Handler から来たリクエストは必ず binding 済みのため、Service の二重チェックは冗長。

**修正方針**: Service の `validateCageType` / `validateCageSize` 呼び出しを削除するか、
「Service は外部から直接呼ばれる可能性があるため防御的チェックを維持する」という意図のコメントを追加する。

---

### 2. chief_complaint_service.go — Reorder と Delete の実装順が interface と逆転

**ファイル**: `backend/internal/service/chief_complaint_service.go`

Interface の宣言順: `List, GetByID, Create, Update, Delete, Reorder`
実装の順序: `List, GetByID, Create, Update, Reorder(L129), Delete(L142)`

他の全ドメインは interface 宣言順に実装を並べている。

**修正方針**: `Reorder` と `Delete` の実装ブロックを入れ替える。

---

### 3. occupation_service.go — buildOccupationUpdateFields がカラム名リテラル直書き

**ファイル**: `backend/internal/service/occupation_service.go`

```go
// 現状: リテラル直書き
fields["name"] = *input.Name
fields["description"] = *input.Description
fields["is_active"] = *input.IsActive
```

`animal_species_service.go`（定数方式）と `occupation_service.go`（リテラル方式）が混在。

**修正方針**: 定数を定義して typo を防ぐ。
```go
const (
    colOccupationName        = "name"
    colOccupationDescription = "description"
    colOccupationIsActive    = "is_active"
    colOccupationSortOrder   = "sort_order"
)
```

---

### 4. consultation_handler.go:70 — nilIfEmpty 変換が handler 層に存在

**ファイル**: `backend/internal/handler/consultation_handler.go`

```go
// handler:70 — nilIfEmpty によるビジネス変換が handler に存在
TaxType: nilIfEmpty(input.TaxType),
```

`nilIfEmpty` は「空文字列を nil に変換する」意味的な変換。
Handler はリクエスト型から Input DTO への構造変換のみ担当すべき。
`consultation_service.go:130` で既に `if input.TaxType != nil && *input.TaxType != ""` チェックが存在するため二重処理にもなっている。

**修正方針**: handler から `nilIfEmpty` を削除し `*string` のまま渡す。service 側のチェックに一本化。

---

### 5. reservation_type_handler.go:184-186 — エラーメッセージが英語

**ファイル**: `backend/internal/handler/reservation_type_handler.go`

```go
RespondError(c, apperrors.WrapInvalidInput("specific_date must be YYYY-MM-DD"))
```

プロジェクト全体で日本語メッセージで統一されているが、このメッセージのみ英語。

**修正方針**:
```go
RespondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
```

---

### 6. trimming_course_repository.go:85-94 — clinicScope 未使用

**ファイル**: `backend/internal/repository/trimming_master_repository.go`

```go
// 現状: 手動 WHERE
Where("course_id = ? AND clinic_id = ?", courseID, clinicID).Count(&count)

// 規約: clinicScope を使用
Scopes(clinicScope(clinicID)).Where("course_id = ?", courseID).Count(&count)
```

**修正方針**: `clinicScope(clinicID)` を使用するパターンに統一。

---

### 7. reservation_type_service.go:408-417 — LinkOccupation の FindAll + 線形探索

**ファイル**: `backend/internal/service/reservation_type_service.go`

```go
// 作成後に全件 FindAll → ループで線形探索
items, err := s.occupationRepo.FindAll(ctx, clinicID, reservationTypeID)
for i := range items {
    if items[i].OccupationID == occupationID {
        return &items[i], nil
    }
}
return o, nil  // Occupation が Preload されていない fallback
```

作成後に全件取得してから線形探索している。N+1 的な無駄があり、
fallback の `return o, nil` は `o.Occupation` が Preload されていない状態で返る可能性がある。

**修正方針**: Repository に `FindByOccupationID(ctx, clinicID, reservationTypeID, occupationID)` を追加し、
Create 後に単件取得する。

---

### 8. permission_group_handler.go:162-177 — 削除前の GetByID 失敗時の WarnContext 条件

**ファイル**: `backend/internal/handler/permission_group_handler.go`

```go
oldPG, getErr := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
if getErr != nil {
    slog.WarnContext(...)  // 404 でも warn が出る
}
```

削除対象が存在しない（404）場合でも `slog.WarnContext` が記録されてしまう。
404 は正常なエラーケースなのでノイズになる。

**修正方針**:
```go
oldPG, getErr := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
if getErr != nil && !errors.Is(getErr, apperrors.ErrNotFound) {
    slog.WarnContext(c.Request.Context(), "failed to fetch old permission group for audit",
        slog.String("error", getErr.Error()))
}
```

---

### 9. inquiry_template_repository.go — CountUsageByInquiryTemplateID のスタブコメント改善

**ファイル**: `backend/internal/repository/inquiry_template_repository.go`

`CountUsageByInquiryTemplateID` が常に `0, nil` を返すスタブになっているが、
コメントが将来の実装指示として不明確。

**修正方針**: TODO 形式に変更。
```go
// TODO: inquiry_answers テーブル追加時に以下の COUNT クエリを実装すること。
// 現スキーマには inquiry_template_id FK を持つテーブルが存在しないため常に 0 を返す。
func (r *inquiryTemplateRepository) CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error) {
    return 0, nil
}
```

---

### 10. medicine_response.go:30-38 — 不要な型変換コピー

**ファイル**: `backend/internal/handler/medicine_response.go`

```go
// 現状: 冗長な中間変数
if m.DosageForm != nil {
    s := string(*m.DosageForm)   // 中間 string 変数
    dosageForm = &s
}
```

`model.DosageForm` は `string` の型エイリアスのため、中間変数は不要。

**修正方針**: ヘルパー関数を使うか、直接変換する。
```go
if m.DosageForm != nil {
    v := string(*m.DosageForm)
    dosageForm = &v  // 変数名を1文字にする（Go の慣習に従い）
}
// または response 型のフィールドを *model.DosageForm に変更
```

---

## 規約参照

- `.claude/CLAUDE.md`: handler → service の責任分離
- `.claude/rules/go-language.md`: 命名規則・コード構造

## テスト

- 各修正は既存の handler/service テストでカバーされること
- slog メッセージ変更・コメント変更は動作に影響しないため追加テスト不要
