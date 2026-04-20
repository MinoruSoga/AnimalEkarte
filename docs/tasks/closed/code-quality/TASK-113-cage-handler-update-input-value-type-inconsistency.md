# TASK-113: `cage_handler.go` — UpdateCageInput の初期化パターンが他マスタと不統一

## 優先度

**Low** — コードの一貫性の問題。機能には影響しない。

---

## 概要

`cage_handler.go` の `UpdateCage` ハンドラで `service.UpdateCageInput` を
**値型** で初期化してからポインタを取得している。

一方、同ファイルの `CreateCage` ハンドラや他のマスタハンドラ
（`vaccine_handler.go`, `procedure_handler.go`, `exam_type_handler.go` 等）は
`&service.CreateXxxInput{}` でポインタリテラルとして初期化している。

この不統一により「なぜここだけ値型を経由するのか」という混乱が生じる。

---

## 問題箇所

### `handler/cage_handler.go:99-109`

```go
// ❌ 値型で初期化してから & でポインタに変換（他マスタと不統一）
svcInput := service.UpdateCageInput{
    Name:        input.Name,
    CageType:    input.CageType,
    CageSize:    input.CageSize,
    Price:       input.Price,
    IsActive:    input.IsActive,
    Description: input.Description,
    SortOrder:   input.SortOrder,
}
cage, err := h.svc.Cage.Update(c.Request.Context(), clinicID, id, &svcInput)
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/cage_handler.go:64-73 (CreateCage — 同ファイル内が既に正しい)
svcInput := &service.CreateCageInput{
    Name:        input.Name,
    CageType:    input.CageType,
    ...
}
cage, err := h.svc.Cage.Create(c.Request.Context(), clinicID, svcInput)

// ✅ handler/vaccine_handler.go (UpdateVaccine)
input := &service.UpdateVaccineInput{
    Name:    req.Name,
    ...
}
vaccine, err := h.svc.Vaccine.Update(...)

// ✅ handler/exam_type_handler.go (UpdateExaminationType)
svcInput := &service.UpdateExamTypeInput{
    Name:     req.Name,
    ...
}
```

---

## 修正方針

### `handler/cage_handler.go:99-109`

変数初期化と同時にポインタリテラルとして生成する。

```go
// ✅ 修正後
cage, err := h.svc.Cage.Update(c.Request.Context(), clinicID, id, &service.UpdateCageInput{
    Name:        input.Name,
    CageType:    input.CageType,
    CageSize:    input.CageSize,
    Price:       input.Price,
    IsActive:    input.IsActive,
    Description: input.Description,
    SortOrder:   input.SortOrder,
})
```

または変数定義で統一:

```go
// ✅ 代替
svcInput := &service.UpdateCageInput{
    Name:        input.Name,
    CageType:    input.CageType,
    CageSize:    input.CageSize,
    Price:       input.Price,
    IsActive:    input.IsActive,
    Description: input.Description,
    SortOrder:   input.SortOrder,
}
cage, err := h.svc.Cage.Update(c.Request.Context(), clinicID, id, svcInput)
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `handler/cage_handler.go:99-109` | UpdateCage の svcInput 初期化 | ❌ 値型経由 → `&` 変換 |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md`

> 非エクスポート: camelCase

ポインタリテラル `&service.UpdateXxxInput{}` の方が Go 慣用的で他ハンドラとの一貫性が保てる。

### プロジェクト内参照実装

- `handler/cage_handler.go:64-73` — `CreateCage` では既にポインタリテラルを使用（同ファイル内の不統一）
- `handler/vaccine_handler.go` — UpdateVaccine の正しいパターン
