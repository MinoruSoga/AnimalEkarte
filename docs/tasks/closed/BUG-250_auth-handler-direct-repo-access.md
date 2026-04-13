# BUG-250: auth_handler の直接 repository アクセス

## 概要

`auth_handler.go` が handler 層から直接 `h.repos.Staff` / `h.repos.Account` を呼び出しており、
handler → service → repository のレイヤードアーキテクチャに違反している。
認証関連のビジネスロジックが handler に散在しており、テスタビリティとメンテナビリティが低い。

## 現状コード

### `backend/internal/handler/auth_handler.go` — 直接 repo アクセス箇所
```go
// 行131
staff, err := h.repos.Staff.FindByAccountID(ctx, account.ID)

// 行329
staff, err := h.repos.Staff.FindByID(ctx, staffID)

// 行439
staff, err := h.repos.Staff.FindByID(ctx, staffID)

// 行449
account, err := h.repos.Account.GetByID(ctx, accountID)

// 行500
staff, err := h.repos.Staff.FindByID(ctx, staffID)
```

### 比較: 正しい実装（他ハンドラ）
```go
// owner_handler.go — service 経由のアクセスのみ
owner, err := h.svc.Owner.GetByID(c.Request.Context(), clinicID, id)
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/auth_handler.go:131` | FindByAccountID — Login後のstaff取得 | 未修正 |
| `backend/internal/handler/auth_handler.go:329` | FindByID — RefreshToken内 | 未修正 |
| `backend/internal/handler/auth_handler.go:439` | FindByID — GetMe内 | 未修正 |
| `backend/internal/handler/auth_handler.go:449` | Account.GetByID — GetMe内 | 未修正 |
| `backend/internal/handler/auth_handler.go:500` | FindByID — UpdateMe内 | 未修正 |

## 修正方針

### 1. AuthService を新設し、認証関連のビジネスロジックを集約

```go
// backend/internal/service/auth_service.go
type AuthService interface {
    GetStaffByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
    GetStaffWithAccount(ctx context.Context, staffID uint64) (*model.Staff, *model.Account, error)
    GetCurrentUser(ctx context.Context, staffID, accountID uint64) (*model.Staff, *model.Account, error)
}

type authService struct {
    staffRepo   repository.StaffRepository
    accountRepo repository.AccountRepository
}

func (s *authService) GetStaffByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
    staff, err := s.staffRepo.FindByAccountID(ctx, accountID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get staff by account ID")
    }
    return staff, nil
}
```

### 2. Handler を service 経由に変更

```go
// auth_handler.go — 修正後
staff, err := h.svc.Auth.GetStaffByAccountID(ctx, account.ID)
if err != nil {
    RespondError(c, err)
    return
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — レイヤー分離
> Handler は Service のみを呼び出す。Repository への直接アクセスは禁止。

### `.claude/CLAUDE.md` — アーキテクチャ規約
> handler → service → repository の軽量レイヤードを徹底。

## 優先度
**High** — アーキテクチャ規約の根幹に関わる違反。認証ドメインのテスタビリティが著しく低い。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）
- BUG-246: staff_handler も同種の問題（Account 直接操作）

## 関連ファイル
- `backend/internal/handler/auth_handler.go:131,329,439,449,500` — 全修正対象
- `backend/internal/service/auth_service.go` — 新設（AuthService）
