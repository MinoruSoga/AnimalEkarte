# TASK-067: permission_group_response — ID / ClinicID が string 型（他のマスタは uint64）

## 優先度

LOW

---

## 概要

`permission_group_response.go` の `permissionGroupResponse` と `permissionGroupRuleResponse` が、
`ID` / `ClinicID` / `GroupID` を `string` 型（`strconv.FormatUint` 変換）で返している。

全マスタ response（`occupationResponse`, `vaccineResponse`, `cageResponse` 等）は `ID uint64` を使っており、
permission_group のみが例外的に string 型を使用している。

---

## 問題箇所

### backend/internal/handler/permission_group_response.go

```go
// ❌ 現状: ID が string 型
type permissionGroupResponse struct {
    ID       string `json:"id"`        // ← string
    ClinicID string `json:"clinic_id"` // ← string
    // ...
}

type permissionGroupRuleResponse struct {
    ID      string `json:"id"`       // ← string
    GroupID string `json:"group_id"` // ← string
    // ...
}

func toPermissionGroupResponse(pg *model.PermissionGroup) permissionGroupResponse {
    return permissionGroupResponse{
        ID:       strconv.FormatUint(pg.ID, 10),       // ← 変換
        ClinicID: strconv.FormatUint(pg.ClinicID, 10), // ← 変換
        // ...
    }
}
```

---

## 修正方針

他のマスタ response と同様に `uint64` 型に統一する。

```go
// ✅ 修正後
type permissionGroupResponse struct {
    ID       uint64 `json:"id"`
    ClinicID uint64 `json:"clinic_id"`
    // ...
}

func toPermissionGroupResponse(pg *model.PermissionGroup) permissionGroupResponse {
    return permissionGroupResponse{
        ID:       pg.ID,
        ClinicID: pg.ClinicID,
        // ...
    }
}
```

---

## 確認事項

- フロントエンド側が現在 `id` / `clinic_id` を **string として扱っている**場合、この変更は破壊的変更になる
- 変更前にフロントエンドコードの `permission_group` の型定義を確認し、影響範囲を把握すること
- `frontend/src/types/generated/models.ts` の `PermissionGroup` 型と整合を確認

---

## 備考

- JavaScript の `JSON.parse()` は `uint64` の大きな値で精度を失うリスクがあるが、
  このプロジェクトの ID は BIGSERIAL（最大 9.2 × 10^18）であり、
  他のマスタも同様の uint64 を返しているため、特別扱いの根拠が不明
- `staff_response.go` は `ID uint64` を使っており、同じスタッフ系でも不統一
