---
name: go-security
description: Go セキュリティ分析・OWASP対応（SQLインジェクション、パスワードハッシング。gosec は本プロジェクト未導入）
---

# Go Security Analysis

Go アプリケーションのセキュリティを OWASP Top 10 に基づいて分析します（gosec は本プロジェクト未導入のため手動レビュー中心。導入すれば併用可能）。

## 実行スコープ

### 1. gosec 静的分析

**gosec は本プロジェクト未導入**（Dockerfile/Makefile/CI に無し）。実行する場合は導入が先。CI の security-scan.yml は agentshield（エージェント設定監査）のみで Go コードスキャナではない

```bash
docker compose exec backend gosec -json ./... | jq
```

**検出対象:**
- SQL インジェクション
- 脆弱な暗号化
- ハードコードされた認証情報
- パストラバーサル
- クロスサイトスクリプティング (XSS)

### 2. OWASP Top 10 — Go固有差分

OWASP Top 10 の一般論・脅威説明・チェックリストは `security-checklist` スキルの「OWASP Top 10 対応状況チェック」を参照。ここでは Go 実装固有の差分のみ扱う。

#### A1: Injection — GORM実装
```go
// ❌ 危険: 文字列連結
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)
db.Raw(query).Scan(&result)

// ✅ 安全: パラメータ化
db.Where("id = ?", userID).Find(&result)
```

#### A5: Broken Access Control — clinic_id検証実装
```go
// clinic_id マルチテナント検証
if visitRecord.ClinicID != clinicID {
  return ErrAccessDenied // 必須
}
```

#### A7: XSS — Goテンプレート
- Go で生成: template/html で自動エスケープ（React 側の対応は `react-security` 参照）

#### A9: Using Components with Known Vulnerabilities — Go依存関係スキャン

**nancy は本プロジェクト未導入**

```bash
docker compose exec backend go list -json -m all | nancy sleuth
```

## セキュリティ実装パターン（Go固有）

### パスワードハッシング
```go
import "golang.org/x/crypto/bcrypt"

// ハッシング — エラーは必ず伝播する（無視禁止）
hashedPassword, err := bcrypt.GenerateFromPassword(
  []byte(password), bcrypt.DefaultCost,
)
if err != nil {
  return apperrors.Wrap(err, "hash password")
}

// 検証
err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(inputPassword))
if err != nil {
  return apperrors.WrapUnauthorized("invalid password") // fail-closed: 検証失敗は必ず拒否
}
```

### JWT トークン (認証用)
```go
import "github.com/golang-jwt/jwt/v5"

type Claims struct {
  UserID   uint
  ClinicID uint
  jwt.RegisteredClaims
}

// トークン生成
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(secretKey))
if err != nil {
  return apperrors.Wrap(err, "sign jwt")
}

// 検証（middleware で実装）— エラーを無視すると偽トークンを通過させる致命的脆弱性になる
parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
  return []byte(secretKey), nil
})
if err != nil || !parsed.Valid {
  return apperrors.WrapUnauthorized("invalid token") // fail-closed: パース/検証失敗は必ず拒否
}
```

### Input バリデーション
```go
import "github.com/go-playground/validator/v10"

type CreateOwnerRequest struct {
  Name  string `validate:"required,min=1,max=100"`
  Phone string `validate:"required,e164"`
  Email string `validate:"required,email"`
}

validate := validator.New()
err := validate.Struct(req)
```

### CookieAuth / CSRF（現行）

認証は HttpOnly Cookie。CSRF は meta `csrf-token` ではなく、`middleware.RequireXRequestedWith` が POST/PATCH/DELETE に `X-Requested-With` を強制する。FE は `frontend/src/lib/axios.ts` で全リクエストに付与する。forgot-password など明示除外以外で CSRF middleware を外さない。

SQLインジェクション対策の詳細は上記 A1 のパラメータ化例を参照。LIKE 検索も同様に安全化する:
```go
db.Where("name LIKE ?", "%"+escapeLike(name)+"%").Find(&users)  // escapeLike で % / _ をエスケープ
```

## 出力形式

```markdown
## Go Security Analysis Report

### 🔴 Critical Issues
- **SQL Injection** at internal/owner/handler.go:45
  - Line: `db.Raw(query).Scan()`
  - Fix: Use parameterized queries

### 🟠 High Issues
- **Hardcoded Secret** at config/db.go:10
- **Missing Input Validation** at internal/owner/service.go

### 🟡 Medium Issues
- Weak Password Hashing Algorithm
- Missing CORS Headers

### ✅ Passed
- gosec: 未導入のため実行なし（手動レビューでの検出結果のみ記載）
- OWASP Coverage: 未測定。割合を捏造しない

### 推奨修正リスト
1. [Critical] SQLインジェクション修正 (1時間)
2. [High] 認証情報削除 (30分)
3. [Medium] Input バリデーション追加 (2時間)
```

## 関連スキル

- `security-checklist` - OWASP一般論・シークレット管理・統合チェックリスト
- `database-indexing` - クエリパフォーマンスと安全性
