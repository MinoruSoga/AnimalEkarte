---
name: security-checklist
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

// ✅ VITE_* はブラウザへ公開される設定値だけに使う
const apiBaseURL = import.meta.env.VITE_API_BASE_URL
```

`VITE_*` はビルド時にクライアントへ埋め込まれる公開設定であり、API キー・トークン・パスワードなどのシークレット保管場所ではない。シークレットはサーバー側の環境変数またはシークレットマネージャーでのみ扱う。

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
// ✅ middleware で全ルートに認証チェック（実在 API: backend/internal/middleware/auth.go:30）
// clinic_id スコープは Auth middleware 内で c.Set("clinic_id", ...) される
// （ClinicScope という middleware は存在しない）
router.Use(middleware.Auth(secret, isProduction, auditSvc, staffSvc))

// ✅ Handler で clinic_id を middleware から取得（ユーザー入力信用禁止）
clinicID := c.GetString("clinic_id") // context 値は string（auth.go が c.Set("clinic_id", string) している）。uint64 が必要なら strconv.ParseUint
userID := c.GetString("user_id")     // middleware が設定した値

// ❌ ユーザー入力から clinic_id を取得（禁止）
clinicID, _ := strconv.ParseUint(c.Query("clinic_id"), 10, 64)
```

## 認可・検証ロジックのエラーパスは fail-closed

```go
// ❌ 検証用データの取得エラーを握り潰すと検証が vacuous pass する（CRITICAL 実例）
ids, err := repo.FindAllGroupIDsByStaffID(ctx, staffID)
if err != nil { ids = nil } // → 自己参照ガードが空集合で素通り
// ✅ 取得失敗はエラーで拒否（fail-closed）
if err != nil { return apperrors.Wrap(...) }
```

チェック: 認可・ガード判定の入力取得が error 時に「空集合＋続行」になっていないか。同型として、parse 失敗の `continue`（fail-open）もバリデーション素通りを起こす。

（出典: memory be_second_lens_audit_20260630 / F-3 CRITICAL 修正 76cb562f）

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

// ✅ 最小限の非PII識別子だけを記録
slog.InfoContext(ctx, "user login", "staff_id", staffID)
```

パスワード・トークンだけでなく、メールアドレス、氏名、電話番号、住所、カルテ内容などの PII をログへ記録しない。調査には最小限の非 PII 識別子とリクエスト相関 ID を用いる。

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
docker compose exec backend gosec ./...  # 未導入 — 導入後のみ実行可

# TypeScript: 依存関係の脆弱性
docker compose exec frontend pnpm audit --audit-level=high

# シークレット検出（コミット前）: 一致行や値は表示せず、ファイル名と件数だけを出力
rg -l --glob='*.go' --glob='*.ts' --glob='*.tsx' 'sk-|api_key|password\s*=' backend/ frontend/src/ \
  | awk 'BEGIN { count = 0 } { print "match_file=" $0; count++ } END { print "match_file_count=" count }'
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
