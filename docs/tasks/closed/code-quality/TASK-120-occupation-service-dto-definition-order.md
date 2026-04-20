# TASK-120: `occupation_service.go` — `UpdateOccupationInput` が `CreateOccupationInput` より前に定義されている

## 優先度

**Low** — コードの一貫性の問題。機能には影響しない。

---

## 概要

`occupation_service.go` において `UpdateOccupationInput`（行 17）が
`CreateOccupationInput`（行 25）より前に定義されている。

プロジェクト内の他のマスタサービスファイルはすべて
`CreateXxxInput` → `UpdateXxxInput` の順序で DTO を定義している。

---

## 問題箇所

### `service/occupation_service.go:17-30`

```go
// ❌ UpdateDTO が CreateDTO より前に定義されている
type UpdateOccupationInput struct {  // 行 17
    Name        *string
    Description *string
    SortOrder   *int
    IsActive    *bool
}

type CreateOccupationInput struct {  // 行 25
    Name        string
    Description string
    SortOrder   int
    IsActive    bool
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/cage_service.go — Create → Update の順序
type CreateCageInput struct { ... }   // 先に Create
type UpdateCageInput struct { ... }   // 後に Update

// ✅ service/vaccine_service.go — Create → Update の順序
type CreateVaccineInput struct { ... }
type UpdateVaccineInput struct { ... }
```

---

## 修正方針

### `service/occupation_service.go:17-30`

`UpdateOccupationInput` と `CreateOccupationInput` の定義順序を入れ替える。

```go
// ✅ 修正後
type CreateOccupationInput struct {  // Create を先に
    Name        string
    Description string
    SortOrder   int
    IsActive    bool
}

type UpdateOccupationInput struct {  // Update を後に
    Name        *string
    Description *string
    SortOrder   *int
    IsActive    *bool
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/occupation_service.go:17-30` | UpdateOccupationInput・CreateOccupationInput の定義順 | ❌ Update が Create より前 |

コードの移動のみ。ロジック変更なし。

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `service/cage_service.go` — `CreateCageInput` → `UpdateCageInput` の正しい順序
- `service/vaccine_service.go` — 同上
- `service/medicine_service.go` — 同上
