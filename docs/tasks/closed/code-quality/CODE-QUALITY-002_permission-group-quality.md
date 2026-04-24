# CODE-QUALITY-002: permission_group 全レイヤー品質修正

## 概要

`permission_group` の Handler / Service / Repository に複数の規約違反がある。  
エラー無視、バリデーション欠落、Soft Delete の非統一が主な問題。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/handler/permission_group_handler.go` | L172, L227, L250 |
| `backend/internal/service/permission_group_service.go` | L37, L70 |
| `backend/internal/repository/permission_group_repository.go` | L92-104 |

---

## 問題一覧

### [Handler] 1. `_ = err` によるエラー無視（規約禁止事項）

```go
// L172: エラーを明示的に無視（禁止パターン）
oldPG, _ := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
```

監査ログ用のデータ取得であり「ベストエフォート」の意図は理解できるが、  
`_ = err` はプロジェクト規約で明示的に禁止されている。

**修正方針**:
```go
// 取得失敗は警告ログで記録し、処理は続行する
oldPG, getErr := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
if getErr != nil {
    slog.WarnContext(c.Request.Context(), "failed to fetch old permission group for audit",
        slog.String("error", getErr.Error()))
}
```

---

### [Handler] 2. `extractStaffID` の二重呼び出し（冗長）

`SetPermissionGroupRules` 内で L227 で取得した `staffID` を使わず、L250 で再度 `extractStaffID(c)` を呼び出している。  
現状は同じ値だが、将来コンテキスト変更があった場合に監査ログと実行ユーザーが乖離するリスクがある。

**修正方針**: L227 で取得した `staffID` を監査ログブロックでもそのまま使用する。

---

### [Service] 3. `Create` の引数が値型（他サービスと不統一）

```go
// 現状: 値渡し
func (s *permissionGroupService) Create(ctx context.Context, clinicID uint64, input CreatePermissionGroupInput) ...

// 他の全サービス: ポインタ渡し
func (s *xxxService) Create(ctx context.Context, clinicID uint64, input *CreateXxxInput) ...
```

**修正方針**: インターフェースと実装を `input *CreatePermissionGroupInput` に統一。

---

### [Service] 4. `Create` の `validateRequiredName` 未呼び出し

他のすべてのサービスの `Create` では `validateRequiredName(input.Name)` が最初に実行されるが、  
`permission_group_service.go` では Name バリデーションが欠落している。  
空文字の権限グループ名が登録できてしまう。

**修正方針**:
```go
func (s *permissionGroupService) Create(...) (*model.PermissionGroup, error) {
    if err := validateRequiredName(input.Name); err != nil {
        return nil, err
    }
    // ...
}
```

---

### [Repository] 5. `Delete` が GORM soft delete を bypass

```go
// 現状: 手動で deleted_at を更新（GORM soft delete フック bypass）
r.db.WithContext(ctx).
    Model(&model.PermissionGroup{}).
    Where("clinic_id = ? AND id = ?", clinicID, id).
    Update("deleted_at", gorm.Expr("now()"))

// 他の全リポジトリ: GORM soft delete を使用
r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).
    Delete(&model.PermissionGroup{})
```

手動更新では `BeforeDelete` / `AfterDelete` GORM フックが発火しない（将来のフック追加時のリスク）。

**修正方針**: 意図的な回避（例: コールバック無効化が必要）なら理由をコメントで明記。  
そうでなければ `.Delete(&model.PermissionGroup{})` に統一する。

---

## 規約参照

- `.claude/rules/go-language.md`: 「Ignoring errors (`_ = err`)」禁止事項
- `.claude/CLAUDE.md`: Service 層のバリデーションパターン
- `.claude/rules/go-language.md`: GORM PATCH パターン

## テスト

- `Create` に空文字名を渡した場合に 400 エラーが返ることを検証
- `Delete` 後に `deleted_at` が設定され、GORM フックが動作することを検証
