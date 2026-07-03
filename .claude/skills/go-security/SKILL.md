---
name: go-security
description: Go セキュリティ分析・OWASP対応（gosec、SQLインジェクション、パスワードハッシング）
---

# Go Security Analysis

Go アプリケーションのセキュリティを OWASP Top 10 と gosec に基づいて分析します。

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

OWASP Top 10 の一般論・脅威説明・チェックリストは `security-review` スキルの「OWASP Top 10 対応状況チェック」を参照。ここでは Go 実装固有の差分のみ扱う。

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

// ハッシング
hashedPassword, _ := bcrypt.GenerateFromPassword(
  []byte(password), bcrypt.DefaultCost,
)

// 検証
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(inputPassword))
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
tokenString, _ := token.SignedString([]byte(secretKey))

// 検証（middleware で実装）
token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
  return []byte(secretKey), nil
})
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

SQLインジェクション対策の詳細は上記 A1 のパラメータ化例を参照。LIKE 検索も同様に安全化する:
```go
db.Where("name LIKE ?", "%"+name+"%").Find(&users)
```

## 出力形式

```markdown
## Go Security Analysis Report

### 🔴 Critical Issues
- **SQL Injection** at handler/owner_handler.go:45
  - Line: `db.Raw(query).Scan()`
  - Fix: Use parameterized queries

### 🟠 High Issues
- **Hardcoded Secret** at config/db.go:10
- **Missing Input Validation** at service/owner_service.go

### 🟡 Medium Issues
- Weak Password Hashing Algorithm
- Missing CORS Headers

### ✅ Passed
- gosec warnings: 0
- OWASP Coverage: 90%

### 推奨修正リスト
1. [Critical] SQLインジェクション修正 (1時間)
2. [High] 認証情報削除 (30分)
3. [Medium] Input バリデーション追加 (2時間)
```

## 関連スキル

- `security-review` - OWASP一般論・シークレット管理・統合チェックリスト
- `database-indexing` - クエリパフォーマンスと安全性
