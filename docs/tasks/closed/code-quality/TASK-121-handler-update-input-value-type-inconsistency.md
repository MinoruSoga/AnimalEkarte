# TASK-121: 複数ハンドラ — UpdateInput の値型初期化パターンが他マスタと不統一

## 優先度

**Low** — コードの一貫性の問題。機能には影響しない。

---

## 概要

以下の 3 ハンドラで `service.UpdateXxxInput` を **値型** で初期化してからポインタを取得している。
TASK-113（`cage_handler.go`）と同一のパターン違反が複数ファイルに存在する。

他のマスタハンドラ（`vaccine_handler.go`, `exam_type_handler.go` 等）は
`&service.UpdateXxxInput{}` でポインタリテラルとして初期化している。

---

## 問題箇所

### `handler/trimming_master_handler.go:96-106`（UpdateTrimmingCourse）

```go
// ❌ 値型で初期化してから & でポインタに変換
svcInput := service.UpdateTrimmingCourseInput{
    Name:        req.Name,
    Price:       req.Price,
    IsActive:    req.IsActive,
    Description: req.Description,
    TargetSize:  req.TargetSize,
    Duration:    req.Duration,
    SortOrder:   req.SortOrder,
}
course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, &svcInput)
```

### `handler/trimming_master_handler.go:231` 付近（UpdateTrimmingOption）

```go
// ❌ 同様のパターン（UpdateTrimmingOptionInput）
svcInput := service.UpdateTrimmingOptionInput{ ... }
option, err := h.svc.TrimmingOption.Update(c.Request.Context(), clinicID, id, &svcInput)
```

### `handler/merchandise_item_handler.go:65-75`（CreateMerchandiseItem）

```go
// ❌ Create でも値型初期化
input := service.CreateMerchandiseItemInput{
    Name:      req.Name,
    ...
}
item, err := h.svc.MerchandiseItem.Create(c.Request.Context(), clinicID, &input)
```

### `handler/merchandise_item_handler.go:101-111`（UpdateMerchandiseItem）

```go
// ❌ Update でも値型初期化
input := service.UpdateMerchandiseItemInput{
    Name:      req.Name,
    ...
}
item, err := h.svc.MerchandiseItem.Update(c.Request.Context(), clinicID, id, &input)
```

### `handler/insurance_handler.go:96-105`（UpdateInsurance）

```go
// ❌ Update は値型初期化（Create は正しい実装）
svcInput := service.UpdateInsuranceInput{
    Name:         req.Name,
    ...
}
insurance, err := h.svc.Insurance.Update(c.Request.Context(), clinicID, id, &svcInput)
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/insurance_handler.go:62 (CreateInsurance — 同ファイル内が正しい)
svcInput := &service.CreateInsuranceInput{
    Name:         req.Name,
    ...
}
insurance, err := h.svc.Insurance.Create(c.Request.Context(), clinicID, svcInput)

// ✅ handler/vaccine_handler.go (UpdateVaccine)
svcInput := &service.UpdateVaccineInput{
    Name: req.Name,
    ...
}
```

---

## 修正方針

各箇所で `&service.UpdateXxxInput{...}` のポインタリテラル形式に統一する。

```go
// ✅ 修正後（trimming_master_handler.go）
course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, &service.UpdateTrimmingCourseInput{
    Name:        req.Name,
    Price:       req.Price,
    IsActive:    req.IsActive,
    Description: req.Description,
    TargetSize:  req.TargetSize,
    Duration:    req.Duration,
    SortOrder:   req.SortOrder,
})

// ✅ または変数形式を統一
svcInput := &service.UpdateTrimmingCourseInput{ ... }
course, err := h.svc.TrimmingCourse.Update(c.Request.Context(), clinicID, id, svcInput)
```

---

## 影響範囲

| ファイル | 行 | 対象メソッド | 状態 |
|---------|---|------------|------|
| `handler/trimming_master_handler.go:96` | UpdateTrimmingCourse | ❌ 値型 → `&` 変換 |
| `handler/trimming_master_handler.go:231` 付近 | UpdateTrimmingOption | ❌ 値型 → `&` 変換 |
| `handler/merchandise_item_handler.go:65` | CreateMerchandiseItem | ❌ 値型 → `&` 変換 |
| `handler/merchandise_item_handler.go:101` | UpdateMerchandiseItem | ❌ 値型 → `&` 変換 |
| `handler/insurance_handler.go:96` | UpdateInsurance | ❌ 値型 → `&` 変換 |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `handler/insurance_handler.go:62` — CreateInsurance では正しいポインタリテラルを使用（同ファイル内の不統一）
- `handler/vaccine_handler.go` — UpdateVaccine の正しいパターン
- TASK-113: `handler/cage_handler.go` — 同一パターンの先行チケット
