# BUG-186: JWT アクセストークンがログインレスポンスボディに漏洩

| 項目 | 内容 |
|------|------|
| 優先度 | **Critical** |
| CWE | CWE-200: Exposure of Sensitive Information |
| OWASP | A02:2021 Cryptographic Failures |

## 概要

ログインレスポンスの JSON ボディに生の JWT アクセストークン文字列が含まれている。
httpOnly Cookie で設定しているにもかかわらず、レスポンスボディにも含めているため、
XSS 脆弱性が存在した場合にトークンが窃取可能。フロントエンドはこのフィールドを使用していない。

## 現状コード

### `backend/internal/handler/auth_response.go:4-9`
```go
type LoginResponse struct {
	Token         string      `json:"token"`      // ← JWT が JSON に含まれる
	ExpiresAt     int64       `json:"expires_at"`  // ← 同上
	IsSystemAdmin bool        `json:"is_system_admin"`
	User          *MeResponse `json:"user"`
}
```

### `backend/internal/handler/auth_handler.go:261-266`
```go
c.JSON(http.StatusOK, LoginResponse{
	Token:         accessTokenStr,  // ← ここで JWT をレスポンスに含めている
	ExpiresAt:     expiresAt.Unix(),
	IsSystemAdmin: account.IsSystemAdmin,
	User:          buildMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
})
```

### `frontend/src/features/auth/api/login.ts:27-30`
```typescript
// JWT トークンはバックエンドが httpOnly Cookie で設定するため
// フロントエンド側での sessionStorage 保存は不要
const user = mapMeToAuthUser(data.user);  // data.token は未使用
```

## 修正方針

### 1. `backend/internal/handler/auth_response.go`
```go
type LoginResponse struct {
	IsSystemAdmin bool        `json:"is_system_admin"`
	User          *MeResponse `json:"user"`
}
```

### 2. `backend/internal/handler/auth_handler.go:261-266`
```go
c.JSON(http.StatusOK, LoginResponse{
	IsSystemAdmin: account.IsSystemAdmin,
	User:          buildMeResponse(staff, account, mainClinicID, clinicNameMap, allClinics, permMap),
})
```

## 準拠すべきプロジェクト規約

### `.claude/rules/security.md`
> Never log sensitive data (passwords, tokens)

JWT トークンをレスポンスボディに含めることは、ログ記録やブラウザ DevTools で露出するリスクがある。

## 優先度
**Critical** — httpOnly Cookie の保護を完全に無効化しており、XSS 存在時にセッションハイジャックが可能。
