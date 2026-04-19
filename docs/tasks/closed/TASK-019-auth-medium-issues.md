# TASK-019: 認証系 MEDIUM 問題 4件（LINE エラー漏洩 / middleware JSON スキーマ / validators 言語不統一 / liff StaffID=0）

## 概要

auth/LIFF/handler 系で発見された MEDIUM 優先度の問題を一括管理する。

## 優先度

MEDIUM

---

## 問題 1: liff_auth.go — LINE API エラーメッセージをレスポンスに直接含めている

### ファイル
`backend/internal/middleware/liff_auth.go:114-118`

### 問題
```go
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ID token: " + err.Error()})
```
`err.Error()` の内容によっては LINE API の内部メッセージや ID Token の一部がクライアントに露出する。

### 修正案
```go
slog.WarnContext(c.Request.Context(), "invalid LINE ID token", slog.String("error", err.Error()))
c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ID token"})
```

---

## 問題 2: ミドルウェア層のエラーレスポンス JSON スキーマが handler 層と不統一

### ファイル
- `backend/internal/middleware/auth.go:52, 64`
- `backend/internal/middleware/liff_auth.go`
- `backend/internal/middleware/rate_limit.go`

### 問題
ミドルウェアのエラーは `{"error": "..."}` だが、ハンドラ層の `RespondError` は `{"code": ..., "message": ..., "timestamp": ...}` を返す。フロントエンドの `handleApiError` がスキーマを統一して処理できない。

### 修正案
middleware パッケージ内に共通レスポンス関数を追加:
```go
// middleware/response.go（新規）
func respondError(c *gin.Context, status int, msg string) {
    c.AbortWithStatusJSON(status, gin.H{
        "code":      status,
        "message":   msg,
        "timestamp": time.Now(),
    })
}
```

---

## 問題 3: validators.go のエラーメッセージ言語不統一

### ファイル
`backend/internal/service/validators.go:111`

### 問題
`validateDiscountRate` のみ英語 `"discount_rate must be between 0 and 100"`。他のバリデーションは日本語。

### 修正案
```go
// Before
return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
// After
return apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
```

---

## 問題 4: liff_handler.go — StaffID=0 の場合も checkDoctorClinicAssignment を呼んでいる

### ファイル
`backend/internal/handler/liff_handler.go:204`

### 問題
`req.StaffID == 0` は「指名なし」を意味するが、ガードなしに `checkDoctorClinicAssignment` を呼んでいる。実装次第では `staffID=0` で不正な DB クエリが発生する可能性がある。

### 修正案
```go
if req.StaffID != 0 {
    if err := h.checkDoctorClinicAssignment(c.Request.Context(), clinicID, req.StaffID); err != nil {
        RespondError(c, err)
        return
    }
}
```

---

## 問題 5: RefreshToken で mainClinicID を旧 claims から引き継いでいる

### ファイル
`backend/internal/handler/auth_handler.go:391-394`

### 問題
`assignments` を再取得しているのに `mainClinicID` は旧 claims の値を使い続ける。スタッフがクリニックを変更した場合に反映されない。

### 修正案
```go
// Before
mainClinicID := claims.ClinicID
// After
mainClinicID, clinicIDs := resolveClinicInfo(assignments)
```
