---
name: security-review
description: セキュリティレビューの統合チェックリスト。認証・入力バリデーション・シークレット管理・OWASP Top 10対応。API実装・認証コード変更・ユーザー入力処理時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# セキュリティレビュースキル

このプロジェクト（Go/Gin + React 19）向けの統合セキュリティチェックリスト。

## 他スキルとの使い分け

- **本スキル**: 認証・入力検証・シークレット管理・OWASP Top10 全般の統合チェックリスト
- **go-security**: Go/Gin/GORM 固有の実装詳細（gosec, SQLi のパラメータ化, bcrypt/JWT の Go 実装）
- **react-security**: React/TSX 固有の実装詳細（dangerouslySetInnerHTML, CSRF token 管理, localStorage）

3スキルとも参照してよいが、本文の重複を避けるため OWASP 一般論は本スキルに集約し、go-security/react-security は言語固有の diff のみを持つ。

## When to Activate

- 認証・認可コードの実装・変更
- 新規 API エンドポイントの追加
- ユーザー入力を処理するコードの変更
- シークレット・機密情報を扱うコードの変更

## 1. シークレット管理

```go
// ❌ 絶対に禁止
const apiKey = "sk-xxx-hardcoded"
db, _ := gorm.Open("postgres://user:password@localhost/db")

// ✅ 環境変数から読み込み
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY is required")
}
```

```typescript
// ❌ 絶対に禁止
const API_KEY = "sk-xxx-hardcoded"

// ✅ 環境変数
const API_KEY = import.meta.env.VITE_API_KEY
```

**チェック:**
- [ ] シークレットが `.env` に格納（`.gitignore` に追加済み）
- [ ] `.env.example` にダミー値のみ
- [ ] ソースコードにキー・パスワードなし

### 環境変数の有無確認で値を漏らさない（実績由来）

- 有無確認に `${VAR:-...}` を使うと設定済みの場合に**値そのものが出力される**。2026-06-25 に GITHUB_TOKEN が平文露出しローテーションに至った実例あり
- 有無だけなら `${VAR:+set}`、長さ確認は `${#VAR}` を使う
- gh CLI の write 操作が env の失効トークンで失敗する場合は `env -u GITHUB_TOKEN gh ...` で keyring トークンを強制使用
（出典: memory feedback_env_var_presence_check_leaks_value / ops_gh_invalid_github_token_env）

## 2. 入力バリデーション（Go）

```go
// ✅ Handler でバインド + バリデーション
func (h *OwnerHandler) Create(c *gin.Context) {
    var req CreateOwnerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    // req は型安全・バリデーション済み
}

// ✅ struct tag でバリデーション
type CreateOwnerRequest struct {
    Name  string `json:"name" binding:"required,min=1,max=100"`
    Email string `json:"email" binding:"required,email"`
    Phone string `json:"phone" binding:"omitempty,e164"`
}
```

## 3. SQL インジェクション防止

```go
// ✅ GORM パラメータ化（自動的に安全）
db.Where("clinic_id = ? AND name = ?", clinicID, name).Find(&owners)

// ❌ 文字列結合（禁止）
db.Where(fmt.Sprintf("name = '%s'", userInput)).Find(&owners)
```

## 4. 認証・認可

```go
// ✅ middleware で全ルートに認証チェック
router.Use(middleware.AuthRequired())
router.Use(middleware.ClinicScope()) // clinic_id によるスコープ

// ✅ Handler で clinic_id を middleware から取得（ユーザー入力信用禁止）
clinicID := c.GetUint64("clinic_id") // middleware が設定した値
ownerID := c.GetUint64("owner_id")   // middleware が設定した値

// ❌ ユーザー入力から clinic_id を取得（禁止）
clinicID, _ := strconv.ParseUint(c.Query("clinic_id"), 10, 64)
```

## 5. XSS 対策（React）

```typescript
// ❌ XSS リスク
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// ✅ テキストとして安全に表示
<div>{userInput}</div>

// ✅ サニタイズが必要な場合は DOMPurify を使用
import DOMPurify from 'dompurify'
<div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(richText) }} />
```

## 6. ログのセキュリティ

```go
// ❌ パスワード・トークンをログに含めない
slog.InfoContext(ctx, "user login", "password", req.Password)  // 禁止

// ✅ センシティブ情報を除外
slog.InfoContext(ctx, "user login", "email", req.Email)
```

## 7. エラーメッセージの情報漏洩

```go
// ❌ 内部エラーをそのまま返す
c.JSON(500, gin.H{"error": err.Error()}) // スタックトレース等が漏洩する可能性

// ✅ RespondError でユーザー向けメッセージに変換
RespondError(c, err) // 内部エラーはログに記録、ユーザーには汎用メッセージ
```

## OWASP Top 10 対応状況チェック

| # | 脅威 | このプロジェクトの対策 |
|---|------|------------------|
| A01 | Broken Access Control | middleware で clinic_id スコープ |
| A02 | Cryptographic Failures | HTTPS 必須, bcrypt パスワードハッシュ |
| A03 | Injection | GORM パラメータ化, ShouldBindJSON |
| A05 | Security Misconfiguration | CORS 設定, Docker non-root |
| A07 | Auth Failures | JWT セッション管理 |
| A09 | Logging Failures | slog 構造化ログ |

## セキュリティスキャンコマンド

**gosec は本プロジェクト未導入**（Dockerfile/Makefile/CI に無し）。実行する場合は導入が先。CI の security-scan.yml は agentshield（エージェント設定監査）のみで Go コードスキャナではない

```bash
# Go: gosec 静的分析
docker compose exec backend gosec ./...

# TypeScript: 依存関係の脆弱性
docker compose exec frontend pnpm audit --audit-level=high

# シークレット検出（コミット前）
grep -rn "sk-\|api_key\|password\s*=" backend/ frontend/src/ --include="*.go" --include="*.ts" --include="*.tsx"
```

## セキュリティチェックリスト

- [ ] シークレットがソースコードにない（環境変数使用）
- [ ] 全入力にバリデーション（binding tag または Zod）
- [ ] GORM パラメータ化クエリ使用
- [ ] middleware で全 API に認証チェック
- [ ] clinic_id は middleware から取得（ユーザー入力不使用）
- [ ] XSS: dangerouslySetInnerHTML に非サニタイズ入力なし
- [ ] エラーメッセージに内部詳細が含まれない
- [ ] ログにパスワード・トークンが含まれない
- [ ] HTTPS 必須設定
- [ ] pnpm audit でクリティカルなし
