---
name: go-security
description: Go セキュリティ分析・OWASP対応（gosec、SQLインジェクション、パスワードハッシング）
---

# Go Security Analysis

Go アプリケーションのセキュリティを OWASP Top 10 と gosec に基づいて分析します。

## 実行スコープ

### 1. gosec 静的分析
```bash
docker compose exec backend gosec -json ./... | jq
```

**検出対象:**
- SQL インジェクション
- 脆弱な暗号化
- ハードコードされた認証情報
- パストラバーサル
- クロスサイトスクリプティング (XSS)

### 2. OWASP Top 10 チェック

#### A1: Injection（インジェクション）
```go
// ❌ 危険: 文字列連結
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)
db.Raw(query).Scan(&result)

// ✅ 安全: パラメータ化
db.Where("id = ?", userID).Find(&result)
```

#### A2: Broken Authentication（認証の破綻）
- [ ] パスワードハッシング: bcrypt/argon2使用
- [ ] セッション管理: httpOnly Cookie + secure flag
- [ ] MFA対応
- [ ] パスワードリセット: トークン有効期限設定

#### A3: Sensitive Data Exposure（機密データ露出）
```go
// ❌ 危険: パスワードをログ
slog.Info("user login", "password", password)

// ✅ 安全: マスク
slog.Info("user login", "user_id", userID)
```

#### A4: XML External Entities (XXE)（XML外部エンティティ）
- [ ] XML パーサー設定確認
- [ ] DTD 無効化

#### A5: Broken Access Control（アクセス制御の破綻）
```go
// clinic_id マルチテナント検証
if visitRecord.ClinicID != clinicID {
  return ErrAccessDenied // 必須
}
```

#### A6: Security Misconfiguration（セキュリティ設定ミス）
- [ ] Debug mode 本番環境で無効
- [ ] セキュリティヘッダー設定
- [ ] CORS ホワイトリスト設定

#### A7: XSS（クロスサイトスクリプティング）
- Go で生成: template/html で自動エスケープ

#### A8: Insecure Deserialization（不安全な逆シリアル化）
- [ ] JSON アンマーシャリング時の型検証

#### A9: Using Components with Known Vulnerabilities（既知の脆弱性を持つコンポーネント）
```bash
docker compose exec backend go list -json -m all | nancy sleuth
```

#### A10: Insufficient Logging & Monitoring（不十分なログ記録・監視）
- [ ] エラーログに詳細情報記録
- [ ] 監査ログ実装

## セキュリティ実装パターン

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

### SQL インジェクション対策
```go
// ❌ 危険
query := "SELECT * FROM users WHERE name = '" + name + "'"

// ✅ 安全 (GORM自動)
db.Where("name = ?", name).Find(&users)
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

- `error-handling-patterns` - エラーメッセージのセキュア化
- `database-indexing` - クエリパフォーマンスと安全性
