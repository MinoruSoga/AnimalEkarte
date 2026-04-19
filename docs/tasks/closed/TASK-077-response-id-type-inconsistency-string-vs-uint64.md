# TASK-077: レスポンス ID 型不統一 — occupation / inquiry_template / chief_complaint が string 使用

## 優先度

LOW

---

## 概要

参照実装 `medicine_response.go` の ID/ClinicID は `uint64` であるが、
一部のマスタドメインで `string` 型（`strconv.FormatUint` で変換）が使われている。

TASK-067（`permission_group_response.go` が string 型を使用している問題）と同じパターンが
他ドメインにも存在する。

---

## 対象ドメイン

| ファイル | 現在の型 | 正しい型 |
|---------|---------|---------|
| `occupation_response.go` | `ID string`, `ClinicID string` | `ID uint64`, `ClinicID uint64` |
| `inquiry_template_response.go` | `ID string`, `ClinicID string` | `ID uint64`, `ClinicID uint64` |
| `chief_complaint_response.go` (要確認) | `string` の可能性 | `uint64` |

参照実装：
- `medicine_response.go`: `ID uint64`, `ClinicID uint64` ✅
- `cage_response.go`: `ID uint64`, `ClinicID uint64` ✅
- `reservation_type_group_response.go`: `ID uint64`, `ClinicID uint64` ✅

---

## 修正方針

```go
// ❌ 修正前（occupation_response.go）
type occupationResponse struct {
    ID       string    `json:"id"`
    ClinicID string    `json:"clinic_id"`
    // ...
}

func toOccupationResponse(o *model.Occupation) occupationResponse {
    return occupationResponse{
        ID:       strconv.FormatUint(o.ID, 10),
        ClinicID: strconv.FormatUint(o.ClinicID, 10),
    }
}

// ✅ 修正後
type occupationResponse struct {
    ID       uint64    `json:"id"`
    ClinicID uint64    `json:"clinic_id"`
    // ...
}

func toOccupationResponse(o *model.Occupation) occupationResponse {
    return occupationResponse{
        ID:       o.ID,
        ClinicID: o.ClinicID,
    }
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `occupation_response.go` | ID/ClinicID を `string` → `uint64` に変更、strconv 削除 |
| `inquiry_template_response.go` | 同上 |
| `chief_complaint_response.go` | 要確認・同パターンであれば同上 |

**注意**: フロントエンドが現在 string 型として受け取っている場合、API 破壊的変更になる可能性がある。
変更前に FE 側の影響確認を行うこと。
