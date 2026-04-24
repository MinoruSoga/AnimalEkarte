# CODE-QUALITY-212: インターフェース肥大化とデッドメソッド

## 概要

`PermissionGroupService` が8メソッド、`ReservationTypeService` が合成12メソッドを持ち、
「インターフェースは3〜5メソッドに絞る」という規約を大きく超過している。
また `DiagnosisNameService.ListNames` がハンドラから一度も呼ばれていないデッドメソッドになっている。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題 |
|---------|-----|
| `backend/internal/service/permission_group_service.go` | インターフェース 8メソッド超過 |
| `backend/internal/service/reservation_type_service.go` | 合成インターフェース 12メソッド超過 |
| `backend/internal/service/diagnosis_service.go` | `ListNames` がデッドメソッド |

---

## 問題 1: PermissionGroupService の肥大化（8メソッド）

### 現状

```go
type PermissionGroupService interface {
    List(...)
    GetByID(...)
    Create(...)
    Update(...)
    Delete(...)
    SetRules(...)       // PermissionGroup 固有操作
    Reorder(...)
    GetEffectivePermissions(...)  // 認可チェック用 — 別責務
}
```

`GetEffectivePermissions` は `auth_handler.go` と `clinic_handler.go` から呼ばれており実使用されているが、
「スタッフの有効な権限を取得する」責務はパーミッショングループの CRUD とは異なる責務（認可サービス）。

### 修正方針

`GetEffectivePermissions` を独立したインターフェースに切り出す。

```go
// 新設: EffectivePermissionService
type EffectivePermissionService interface {
    GetEffectivePermissions(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
}

// Services 構造体に追加
type Services struct {
    ...
    PermissionGroup    PermissionGroupService
    EffectivePermission EffectivePermissionService  // 追加
}
```

`GetEffectivePermissions` の実装は `permissionGroupService` が兼務しても良いが、
インターフェースとして分離することで DI/テストが容易になる。

---

## 問題 2: ReservationTypeService の合成インターフェース（12メソッド）

### 現状

```go
// 3つのサブインターフェースを合成
type ReservationTypeService interface {
    ReservationTypeCoreService        // 6メソッド
    ReservationTypeUnavailableTimeService  // 3メソッド
    ReservationTypeOccupationService  // 3メソッド
}
```

`handler/reservation_type_handler.go` が1つのフィールドで全12メソッドを使っている。
モックが12メソッド全て実装しなければならずテストコストが高い。

### 修正方針（推奨）

`Services` 構造体で3フィールドに分割し、Handler は3つを個別に受け取る。

```go
type Services struct {
    ...
    ReservationType             ReservationTypeCoreService
    ReservationTypeUnavailableTime ReservationTypeUnavailableTimeService
    ReservationTypeOccupation   ReservationTypeOccupationService
}
```

影響範囲: `service.go`、`reservation_type_handler.go`（`h.svc.ReservationTypeXxx` の参照変更）、`main.go` の DI 配線。

---

## 問題 3: DiagnosisNameService.ListNames がデッドメソッド

### 現状

`DiagnosisNameService` インターフェースに `ListNames` が定義されているが、
`handler/` 内のいかなるファイルからも呼ばれていない。
`FindAllActive` リポジトリメソッドまで実装されているが接続されていない。

```go
// diagnosis_service.go
type DiagnosisNameService interface {
    List(...)
    ListNames(...)  // ← handler から未使用
    GetByID(...)
    ...
}
```

### 修正方針

以下のどちらかを選択する：

- **A（推奨）**: ハンドラに `ListDiagnosisNamesAll` エンドポイントを追加し、ルートに登録する（機能として必要な場合）
- **B**: `ListNames` をインターフェースから削除し、実装とリポジトリの `FindAllActive` も削除する（不要な場合）

選択前にフロントエンドの要件を確認すること。

---

## 規約参照

- `.claude/rules/go-language.md`: 「インターフェース最小化（3-5メソッド）」

## テスト

- `PermissionGroupService` から `GetEffectivePermissions` を分離後、auth_handler / clinic_handler のテストが引き続き通ることを確認
- `ReservationTypeService` 分割後、全ルートの統合テストが通ることを確認
