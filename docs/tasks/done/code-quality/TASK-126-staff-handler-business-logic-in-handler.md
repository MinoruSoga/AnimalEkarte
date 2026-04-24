# TASK-126: `staff_handler.go` — Handler 層にバリデーション・デフォルト値設定が混在

## 優先度

**High** — バリデーション関数呼び出しとデフォルト値設定が Handler 層に存在し、責務分離が崩れている。

---

## 概要

`staff_handler.go` の `CreateStaff` ハンドラ（行 56-100）に以下のビジネスロジックが混在している:

1. **バリデーション呼び出し**（行 61）: `validatePassword(req.Password)` をハンドラから直接呼び出し
2. **デフォルト値設定の重複**（行 65-68 と 84-87）: `reservationVisible` のデフォルト値設定が 2 箇所に重複

これらはすべて Service 層（`staff_service.go`）の責務である。

---

## 問題箇所

### 問題 1: `handler/staff_handler.go:56-64` — バリデーション呼び出し

```go
// ❌ Handler でバリデーション関数を直接呼び出している
if email != "" {
    if req.Password == "" {
        RespondError(c, apperrors.WrapInvalidInput("password is required when email is provided"))
        return
    }
    if err := validatePassword(req.Password); err != nil {  // ← Handler からバリデーション直呼び出し
        RespondError(c, err)
        return
    }
```

### 問題 2: `handler/staff_handler.go:65-68 と 84-87` — デフォルト値設定の重複

```go
// ❌ email あり分岐（行 65-68）
reservationVisible := true
if req.ReservationVisible != nil {
    reservationVisible = *req.ReservationVisible
}

// ❌ email なし分岐（行 84-87）— まったく同じコードが重複
reservationVisible := true
if req.ReservationVisible != nil {
    reservationVisible = *req.ReservationVisible
}
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/vaccine_handler.go — Handler はリクエスト解析と委譲のみ
func (h *Handler) CreateVaccine(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    var req createVaccineRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    vaccine, err := h.svc.Vaccine.Create(c.Request.Context(), clinicID, &service.CreateVaccineInput{
        Name: req.Name, ...
    })
    ...
}

// ✅ service/vaccine_service.go — バリデーションは Service で実行
func (s *vaccineService) Create(...) (*model.Vaccine, error) {
    if err := validateRequiredName(input.Name); err != nil { return nil, err }
    ...
}
```

---

## 修正方針

### 問題 1: `validatePassword` を Service に移動

Handler は raw string を Service にそのまま渡す。

```go
// ✅ 修正後（handler）
staff, err = h.svc.Staff.CreateWithAccount(ctx, &service.CreateStaffWithAccountInput{
    ...
    Password: req.Password,  // raw string をそのまま渡す
    ...
})

// ✅ 修正後（service/staff_service.go の CreateWithAccount 内）
func (s *staffService) CreateWithAccount(...) (*model.Staff, error) {
    if input.Email != "" && input.Password == "" {
        return nil, apperrors.WrapInvalidInput("password is required when email is provided")
    }
    if input.Password != "" {
        if err := validatePassword(input.Password); err != nil {  // ← Service でバリデーション
            return nil, err
        }
    }
    ...
}
```

### 問題 2: デフォルト値設定の重複を解消

Handler から重複ロジックを除去し、Service Input にポインタのまま渡す。
Service 側でデフォルト値を設定する。

```go
// ✅ 修正後（handler）
staff, err = h.svc.Staff.CreateWithAccount(ctx, &service.CreateStaffWithAccountInput{
    ...
    ReservationVisible: req.ReservationVisible,  // nil のまま渡す（デフォルトは Service で設定）
    ...
})

// ✅ 修正後（service/staff_service.go）
reservationVisible := true  // デフォルト
if input.ReservationVisible != nil {
    reservationVisible = *input.ReservationVisible
}
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `handler/staff_handler.go:61` | `validatePassword` 呼び出し | ❌ Handler にバリデーション |
| `handler/staff_handler.go:65-68` | `reservationVisible` デフォルト設定 | ❌ 重複ロジック（1 箇所目） |
| `handler/staff_handler.go:84-87` | `reservationVisible` デフォルト設定 | ❌ 重複ロジック（2 箇所目） |
| `service/staff_service.go` | CreateWithAccount | バリデーション・デフォルト値設定の追加が必要 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — 依存関係の方向

> handler (プレゼンテーション層) → service (ビジネスロジック層) → repository (データアクセス層)
> Handler はリクエスト解析と Service への委譲のみを担う。

### プロジェクト内参照実装

- `handler/vaccine_handler.go` — バリデーションを Service に完全委譲した正しいパターン
- TASK-109: `billing_item_handler.go` の同一パターン先行チケット
