# CODE-QUALITY-004: マスタ Handler レスポンス・実装パターン統一

## 概要

マスタ系 Handler 層に複数の実装パターン不統一がある。  
`nilIfEmpty` ヘルパーの未使用、条件分岐内のレスポンス生成重複、godoc コメント不正確など。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 問題 |
|---------|-----|
| `backend/internal/handler/vaccine_handler.go` | L74-76: nilIfEmpty 未使用 |
| `backend/internal/handler/diagnosis_handler.go` | L181-200: レスポンス生成重複 |
| `backend/internal/handler/reservation_type_handler.go` | L222: godoc コメント不一致 |
| `backend/internal/handler/animal_species_handler.go` | L97-109: 設計意図コメント欠落 |

---

## 問題一覧

### 1. `vaccine_handler.go:74-76` — nilIfEmpty 未使用（実装不統一）

```go
// 現状: 手動 if 分岐
if input.Species != "" {
    svcInput.Species = &input.Species
}

// consultation_handler.go では nilIfEmpty ヘルパーを使用
svcInput.TaxType = nilIfEmpty(input.TaxType)
```

同一プロジェクト内で同じ変換ロジックに異なる実装が混在している。

**修正方針**: `nilIfEmpty(input.Species)` に統一する。

---

### 2. `diagnosis_handler.go:181-200` — 条件分岐内のレスポンス生成重複

```go
if typeIDStr := c.Query("type_id"); typeIDStr != "" {
    // ... 処理 ...
    c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(names, toDiagnosisNameResponse), total, page, limit))
} else {
    // ... 処理 ...
    c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(names, toDiagnosisNameResponse), total, page, limit))
    // ↑ 全く同じコード
}
```

**修正方針**: 早期 return パターンに変更してレスポンス生成を1箇所に集約。

```go
var names []model.DiagnosisName
var total int64
var err error

if typeIDStr := c.Query("type_id"); typeIDStr != "" {
    catID, parseErr := strconv.ParseUint(typeIDStr, 10, 64)
    if parseErr != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
        return
    }
    names, total, err = h.svc.DiagnosisName.ListByCategoryID(c.Request.Context(), clinicID, catID, page, limit)
} else {
    names, total, err = h.svc.DiagnosisName.List(c.Request.Context(), clinicID, page, limit)
}
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(names, toDiagnosisNameResponse), total, page, limit))
```

---

### 3. `reservation_type_handler.go:222` — godoc コメントのメソッド名不一致

```go
// ListOccupations godoc  ← 誤
func (h *Handler) ListReservationTypeOccupations(c *gin.Context) {

// 正しくは:
// ListReservationTypeOccupations godoc
```

---

### 4. `animal_species_handler.go:97-109` — グローバルマスタ設計の意図コメント欠落

`AnimalSpecies` は `clinic_id` を持たないグローバルマスタであるため、`Reorder` ハンドラが `extractClinicID` を呼ばない。  
他マスタと異なり、コードを読む人が「clinic_id の取得漏れ」と誤解する可能性がある。

**修正方針**: ハンドラメソッドに以下のコメントを追加。
```go
// ReorderAnimalSpecies は動物種マスタの表示順を更新する。
// AnimalSpecies はシステム共通マスタ（clinic_id なし）のため clinicID パラメータは不要。
func (h *Handler) ReorderAnimalSpecies(c *gin.Context) {
```

---

## 規約参照

- `.claude/rules/go-language.md`: 命名規則
- `.claude/CLAUDE.md`: コードの統一性

## テスト

- `vaccine_handler.go` の Species が空文字の場合に nil として扱われることを検証
- `diagnosis_handler.go` の type_id あり/なし両ケースで正常にページネーションレスポンスが返ることを検証
