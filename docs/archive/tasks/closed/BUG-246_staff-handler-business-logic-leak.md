# BUG-246: staff_handler に bcrypt/Account操作が漏出 + エラー無視 + 非トランザクション

## 概要

`staff_handler.go` が handler → service → repository のレイヤードアーキテクチャを逸脱し、
handler 層から直接 `h.repos.Account` を操作している。パスワードハッシュ化（bcrypt）も handler で実行されており、
セキュリティロジックのテスタビリティが低下している。さらに以下の3つの付随問題がある:

1. **エラー無視**: `existing, _ := h.repos.Account.FindByEmail(...)` でDB接続エラーを握りつぶし
2. **非トランザクション**: `SetStaffClinicAssignments` が Delete + 複数 Create をトランザクションなしで実行
3. **直接 repo アクセス**: `CreateStaff` / `UpdateStaff` で `h.repos.Account` を直接呼び出し

## 現状コード

### `backend/internal/handler/staff_handler.go:55-57` — エラー無視
```go
existing, _ := h.repos.Account.FindByEmail(ctx, req.Email)
// ↑ DB接続エラー時でもアカウント作成が続行される
```

### `backend/internal/handler/staff_handler.go:62-68` — handler 層で bcrypt
```go
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    RespondError(c, apperrors.Wrap(err, "failed to hash password"))
    return
}
```

### `backend/internal/handler/staff_handler.go:308-325` — 非トランザクション
```go
// Delete → 複数 Create がトランザクション外
if err := h.svc.StaffClinicAssignment.DeleteByStaffID(ctx, staffID); err != nil { ... }
for _, assignmentClinicID := range req.ClinicIDs {
    if _, err := h.svc.StaffClinicAssignment.Create(ctx, ...); err != nil { ... }
}
// Delete 成功後に Create が失敗 → データ不整合
```

### 比較: 正しい実装（handler → service パターン）
```go
// owner_handler.go:88-93 — handler は service のみ呼び出す
func (h *Handler) CreateOwner(c *gin.Context) {
    ...
    owner, err := h.svc.Owner.Create(c.Request.Context(), clinicID, input)
    ...
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/staff_handler.go:52-113` | CreateStaff — Account.Create + bcrypt | 未修正 |
| `backend/internal/handler/staff_handler.go:143-188` | UpdateStaff — Account.Update + bcrypt | 未修正 |
| `backend/internal/handler/staff_handler.go:55-57` | FindByEmail エラー無視 (`_ =`) | 未修正 |
| `backend/internal/handler/staff_handler.go:308-325` | SetStaffClinicAssignments — 非トランザクション | 未修正 |

## 修正方針

### 1. StaffService.Create にアカウント作成を統合
```go
// backend/internal/service/staff_service.go
type CreateStaffInput struct {
    Name     string
    Email    string
    Password string
    ClinicID uint64
    // ...
}

func (s *staffService) Create(ctx context.Context, input CreateStaffInput) (*model.Staff, error) {
    // メール重複チェック
    existing, err := s.accountRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to check email uniqueness")
    }
    if existing != nil {
        return nil, apperrors.WrapConflict("email already exists")
    }

    // bcrypt ハッシュ化
    hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to hash password")
    }

    // トランザクション内で Account + Staff + Assignment を作成
    // ...
}
```

### 2. SetStaffClinicAssignments をトランザクション化
```go
// backend/internal/service/staff_service.go
func (s *staffService) SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffClinicAssignment{}).Error; err != nil {
            return apperrors.Wrap(err, "failed to delete existing assignments")
        }
        for _, cid := range clinicIDs {
            assignment := &model.StaffClinicAssignment{StaffID: staffID, ClinicID: cid}
            if err := tx.Create(assignment).Error; err != nil {
                return apperrors.Wrap(err, "failed to create assignment")
            }
        }
        return nil
    })
}
```

### 3. Handler はサービスのみ呼び出す
```go
func (h *Handler) CreateStaff(c *gin.Context) {
    var req createStaffRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    staff, err := h.svc.Staff.Create(c.Request.Context(), service.CreateStaffInput{...})
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toStaffResponse(staff))
}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — アーキテクチャ規約
> handler → service → repository の軽量レイヤードを徹底。

### `.claude/rules/code-style.md` — レイヤー分離
> Handler は Service のみを呼び出す。Repository への直接アクセスは禁止。

### `.claude/rules/go-language.md` — エラーハンドリング
> `_ = err` 相当のパターンは禁止。すべてのエラーを明示的にハンドリングすること。

## 優先度
**Critical** — エラー無視によりDB障害時にアカウント重複作成が可能。非トランザクション処理によるデータ不整合リスク。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）

## 関連ファイル
- `backend/internal/handler/staff_handler.go:52-325` — 全修正対象
- `backend/internal/service/staff_service.go` — 移動先
