# BUG-207: デモ認証情報が環境チェックなしで本番フロントエンドに含まれる

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| CWE | CWE-798: Use of Hard-coded Credentials |
| OWASP | A05:2021 Security Misconfiguration |

## 概要

LoginForm に 8 つのデモアカウント（email + パスワード "password"）がソースコードに直接記載され、
`import.meta.env.DEV` 等の環境チェックなしで常に表示される。

## 現状コード

### `frontend/src/features/auth/components/LoginForm.tsx:25-38`
```typescript
const DEMO_ACCOUNTS: readonly DemoCredential[] = [
  { email: "admin@noavet.jp",        displayName: "システム管理 太郎", ... },
  { email: "admin@example.com",      displayName: "執行 太郎", ... },
  { email: "vet@example.com",        displayName: "一般 花子", ... },
  // ... 8 accounts total
];
```

### `frontend/src/features/auth/components/LoginForm.tsx:237-254`
```typescript
{/* Demo accounts — 環境チェックなしで常に表示 */}
<div className="mt-8">
  <p className={`text-sm text-center mb-2 ${C.text40}`}>パスワード: password</p>
  <div className="space-y-px">
    {DEMO_ACCOUNTS.map((cred) => (
      <DemoAccount key={cred.email} {...cred} onSelect={handleSelectDemo} />
    ))}
  </div>
</div>
```

## 修正方針

### `frontend/src/features/auth/components/LoginForm.tsx`
```typescript
{import.meta.env.DEV ? (
  <div className="mt-8">
    {/* ... demo accounts section ... */}
  </div>
) : null}
```

## 優先度
**High** — 本番環境でデモアカウントが存在する場合、誰でもパスワード "password" でログインできる。
