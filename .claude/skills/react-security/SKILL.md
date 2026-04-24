---
name: react-security
description: React セキュリティ分析・XSS対応（dangerouslySetInnerHTML、CSRF、token管理）
---

# React Security Analysis

React アプリケーションのセキュリティを OWASP + React セキュリティベストプラクティスで分析します。

## 実行スコープ

### 1. XSS（クロスサイトスクリプティング）検出

#### ❌ 危険パターン
```typescript
// 危険: ユーザー入力を直接HTMLに
<div dangerouslySetInnerHTML={{ __html: userInput }} />

// 危険: eval相当
const fn = new Function('return ' + untrustedCode)
eval(untrustedCode)

// 危険: HTML属性に直接バインド
<img src={userProvidedUrl} />
```

#### ✅ 安全パターン
```typescript
// 安全: テキストとして扱う（自動エスケープ）
<div>{userInput}</div>

// 安全: JSON パース + 検証後に使用
const data = JSON.parse(userInput)

// 安全: URL検証
const safeUrl = userInput.startsWith('http') ? userInput : '/default'
<img src={sanitize(userProvidedUrl)} />
```

### 2. CSRF（クロスサイト要求偽造）対策

```typescript
// ✅ 実装例: axios に自動CSRF token付与
const api = axios.create({
  withCredentials: true,
  headers: {
    'X-CSRF-Token': document.querySelector('meta[name="csrf-token"]')?.content
  }
})

// または React Query + Cookie-based
const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false },
    mutations: {
      onError: (error) => {
        if (error.response?.status === 403) {
          // CSRF token expired
          location.reload()
        }
      }
    }
  }
})
```

### 3. Token / 認証情報管理

#### ❌ 危険
```typescript
// localStorage に token 保存 → XSS で盗まれる
localStorage.setItem('token', jwtToken)
```

#### ✅ 安全
```typescript
// Backend: httpOnly + secure Cookie
// Set-Cookie: authToken=XXX; HttpOnly; Secure; SameSite=Strict

// Frontend: 自動的にリクエストに付与（withCredentials）
const api = axios.create({
  withCredentials: true
})

// Token はメモリのみ（リロードで削除）
let authToken: string | null = null
```

### 4. Content Security Policy (CSP)

```typescript
// index.html に設定
<meta http-equiv="Content-Security-Policy"
  content="
    default-src 'self';
    script-src 'self' https://trusted-cdn.com;
    style-src 'self' 'unsafe-inline';
    img-src 'self' data: https:;
    connect-src 'self' https://api.example.com;
    frame-ancestors 'none'
  " />
```

### 5. Dependency チェック

```bash
docker compose exec frontend pnpm audit
docker compose exec frontend pnpm audit --production
```

**脆弱性対応:**
- Critical: 即時修正
- High: 1週間以内
- Medium: 1ヶ月以内

### 6. コンポーネントセキュリティ

#### React.memo でのProps検証
```typescript
interface Props {
  content: string
  onClick?: (id: string) => void
}

// ✅ 型安全
export const Card = React.memo(({ content, onClick }: Props) => {
  return <div onClick={() => onClick?.(id)}>{content}</div>
})
```

#### フォーム入力のサニタイズ
```typescript
import DOMPurify from 'dompurify'

const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
  const cleanInput = DOMPurify.sanitize(e.target.value)
  setInput(cleanInput)
}
```

#### 外部スクリプト読み込みの防止
```typescript
// ❌ 危険
<script src={userProvidedUrl} />

// ✅ 安全: ホワイトリスト化
const ALLOWED_SCRIPTS = [
  'https://trusted-analytics.com/script.js'
]
const isAllowed = ALLOWED_SCRIPTS.includes(scriptUrl)
```

## チェックリスト

### XSS 対策
- [ ] dangerouslySetInnerHTML 使用なし
- [ ] ユーザー入力は自動エスケープ
- [ ] DOMPurify で HTML フィルタリング
- [ ] iframe の信頼済みソースのみ

### CSRF 対策
- [ ] CSRF token 実装
- [ ] SameSite Cookie 設定
- [ ] Origin ヘッダー検証

### 認証・トークン
- [ ] JWT は httpOnly Cookie に保存
- [ ] withCredentials: true 設定
- [ ] Token 有効期限確認
- [ ] リフレッシュトークンメカニズム

### 依存関係
- [ ] pnpm audit 実行
- [ ] 脆弱性パッチ適用
- [ ] 定期更新スケジュール

### ログ・デバッグ
- [ ] console.log で機密情報出力禁止
- [ ] Error stack trace マスク
- [ ] 本番環境で debug mode 無効

## 出力形式

```markdown
## React Security Analysis Report

### 🔴 Critical Issues
- **XSS Vulnerability** at features/owners/routes/OwnersList.tsx:45
  - Issue: `dangerouslySetInnerHTML` detected
  - Fix: Remove and use safe JSX rendering

### 🟠 High Issues
- **Token Management** - localStorage에서 제거
- **CSRF Token** - X-CSRF-Token header 추가

### 🟡 Medium Issues
- Missing CSP headers
- pnpm audit warnings: 2

### ✅ Passed
- XSS protection: ✅
- CSRF token: ✅
- Dependency audit: 0 critical

### 推奨修正リスト
1. [Critical] dangerouslySetInnerHTML 削除 (30分)
2. [High] CSRF token 実装 (1時間)
3. [Medium] CSP ヘッダー設定 (30分)
```

## 関連スキル

- `go-security` - Backend 認証・認可
- `error-handling-patterns` - エラーメッセージのセキュア化
