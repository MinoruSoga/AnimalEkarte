# TASK-087: permission_group_handler — handler 層が repository パッケージを直接 import（レイヤー違反）

## 優先度

HIGH

---

## 概要

`permission_group_handler.go` が `repository` パッケージを直接 import し、
`repository.MarshalAuditJSON()` を handler 層から直接呼び出している。

Clean Architecture の handler → service → repository の依存方向に違反する。
handler が repository 実装詳細に結合し、テスト困難・層境界崩壊を招く。

---

## 問題箇所

### permission_group_handler.go:13

```go
// ❌ handler 層が repository パッケージを直接 import
import (
    "github.com/animal-ekarte/backend/internal/repository"
    // ...
)
```

### permission_group_handler.go:85, 136, 178, 284

```go
// ❌ repository 関数を handler から直接呼び出し
repository.MarshalAuditJSON(pg)
```

---

## 影響

- handler が repository の実装詳細に直接依存 → 層境界の崩壊
- `MarshalAuditJSON` をモック不可 → handler の単体テストが困難
- repository の変更が handler に伝播するリスク

---

## 参照実装（正しいパターン）

`MarshalAuditJSON` の呼び出しは **service 層** で行い、
handler は service メソッドの戻り値を受け取るだけにする。

```go
// ✅ 修正後の設計
// service 層で監査ログ生成
func (s *permissionGroupService) Update(ctx context.Context, ...) (*model.PermissionGroup, error) {
    pg, err := s.repo.UpdateFields(ctx, ...)
    // ...
    auditData, err := s.repo.MarshalAuditJSON(pg)  // service から repo を呼ぶのは正常
    // ...
}

// handler 層: repository を import しない
func (h *permissionGroupHandler) Update(c *gin.Context) {
    pg, err := h.service.Update(ctx, ...)
    // ...
}
```

---

## 修正方針

1. `repository.MarshalAuditJSON` の呼び出しを **service 層に移動**
2. handler の import から `repository` パッケージを削除
3. service インターフェースに監査ログ生成を含める、または service の Update/Create/Delete がログを内部で処理する

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/permission_group_handler.go` | `repository` import 削除、`repository.MarshalAuditJSON()` 呼び出しを削除 |
| `service/permission_group_service.go` | 監査ログ生成ロジックを service 層に移動 |
