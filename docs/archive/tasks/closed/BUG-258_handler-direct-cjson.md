# BUG-258: Handler で c.JSON 直接使用（RespondError 迂回）

## 概要

Handler 層で `RespondError(c, err)` を使わずに直接 `c.JSON(status, gin.H{"error": ...})` を
呼んでいる箇所が残存。エラーレスポンスの統一性が損なわれている。

## 影響範囲

| ファイル | 行番号 | 内容 |
|---------|--------|------|
| `handler/response.go` | :165-222 | `extractStaffID`/`extractClinicID`/`extractIsSystemAdmin`/`extractClinicIDFromParam` が直接 `c.JSON(401/400, ...)` |
| `handler/liff_handler.go` | :200-205 | `ReservationLimitError` を直接 `c.JSON(409, ...)` で返却（`redirect_step` フィールド付き） |
| `handler/reservation_course_handler.go` | :161 | `c.JSON(501, gin.H{"error": "not implemented"})` |
| `handler/reservation_staff_handler.go` | :175 | `c.JSON(501, gin.H{"error": "not implemented"})` |

## 修正方針

### response.go の extractXxx ヘルパー

```go
// Before
c.JSON(http.StatusUnauthorized, gin.H{"error": "missing clinic context"})
return 0, false

// After
RespondError(c, apperrors.WrapUnauthorized("missing clinic context"))
return 0, false
```

### liff_handler.go の ReservationLimitError

`RespondError` を拡張して `ReservationLimitError` 型を認識するか、
例外的パスとしてコメントで明示する。

### not implemented エンドポイント

```go
// Before
c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})

// After — apperrors に WrapNotImplemented がない場合はコメントで例外明示
// TODO: apperrors.WrapNotImplemented 追加後に RespondError に統一
c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
```

## 優先度

**High** — response.go の extractXxx は全ハンドラから呼ばれるため影響範囲が広い。

## 関連チケット

- BUG-253: 親チケット
