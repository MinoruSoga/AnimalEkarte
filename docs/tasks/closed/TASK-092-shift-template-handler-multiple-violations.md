# TASK-092: shift_template_handler — 複数の実装違反（GET欠落・PUT/PATCH・ID型・デッドコード）

## 優先度

MEDIUM

---

## 概要

`shift_template_handler.go` に 4 種類の規約違反が混在している。

---

## 問題箇所

### 1. GetShiftTemplate ハンドラが存在しない

全マスタドメインは List/Get/Create/Update/Delete/Reorder の 6 ハンドラを持つ。
`shift_template_handler.go` は以下 5 つのみで **Get（単一取得）が欠落**している：

```
✅ ListShiftTemplates   GET  ""
❌ GetShiftTemplate が存在しない（GET /:id ルートなし）
✅ CreateShiftTemplate  POST ""
✅ UpdateShiftTemplate  PATCH /:id
✅ DeleteShiftTemplate  DELETE /:id
✅ ReorderShiftTemplates
```

ルート登録（L236-243）にも `GET "/:id"` が存在しない。
Service/Repository は `FindByID` を実装済みのため、handler 層だけが欠落している。

---

### 2. Reorder に PUT を使用（L240）

```go
// ❌ L240: PUT — 全マスタは PATCH を使用
g.PUT("/reorder", h.RequirePermission(...), h.ReorderShiftTemplates)

// ✅ 全マスタの標準（例: exam_type_handler.go）
g.PATCH("/reorder", h.RequirePermission(...), h.ReorderExamTypes)
```

---

### 3. レスポンス struct の ID/ClinicID が string 型（L57-62）

TASK-077 と同パターン。

```go
// ❌ L57-62
type shiftTemplateResponse struct {
    ID        string `json:"id"`       // ← uint64 であるべき
    ClinicID  string `json:"clinic_id"` // ← uint64 であるべき
    // ...
}

// ❌ shiftTemplateBreakResponse も同様
type shiftTemplateBreakResponse struct {
    ID         string `json:"id"` // ← uint64 であるべき
    // ...
}

// ✅ 参照実装（medicine_response.go 等）
type medicineResponse struct {
    ID       uint64 `json:"id"`
    ClinicID uint64 `json:"clinic_id"`
}
```

---

### 4. 未使用の reorderShiftTemplateRequest struct（L44-47）

```go
// ❌ L44-47: 定義されているが一度も使われていない（デッドコード）
type reorderShiftTemplateRequest struct {
    IDs []uint64 `json:"ids" binding:"required"`
}

// 実際のハンドラ L223 では共通の reorderRequest を使用
var req reorderRequest
```

---

## 修正方針

1. `GetShiftTemplate` ハンドラを追加し、`GET "/:id"` ルートを登録
2. `PUT "/reorder"` → `PATCH "/reorder"` に変更
3. `shiftTemplateResponse.ID/ClinicID` を `uint64` に変更、`shiftTemplateBreakResponse.ID` も同様
4. `reorderShiftTemplateRequest` struct を削除

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/shift_template_handler.go` | `GetShiftTemplate` 追加、`PUT`→`PATCH` 変更、ID/ClinicID 型を `uint64` に変更、デッドコード削除 |

---

## 関連

- TASK-077: occupation/inquiry_template/chief_complaint の ID が string 型（同パターン）
