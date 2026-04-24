# FE-138: パスワードリセットページ実装（/forgot-password + /reset-password）

**Status**: Open
**Priority**: Medium
**Affects**: features/auth/routes/, app/router.tsx
**Date Created**: 2026-03-29
**Related**: BUG-060, BE-081（先行必須）

---

## Summary

`/forgot-password` と `/reset-password` ページが未実装。
BE-081 完了後に本チケットを実装する。

---

## 実装手順

### 1. `/forgot-password` ページ（`features/auth/routes/ForgotPasswordPage.tsx`）

```typescript
export function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [isSent, setIsSent] = useState(false);
  const [isPending, startTransition] = useTransition();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    startTransition(async () => {
      await forgotPassword({ email });
      setIsSent(true);
    });
  };

  if (isSent) {
    return (
      <div>
        <p>パスワードリセットのリンクをメールに送信しました。</p>
        <p>メールを確認してください（30分以内に使用してください）。</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit}>
      <h1>パスワードをお忘れですか？</h1>
      <label htmlFor="email">メールアドレス</label>
      <input
        id="email"
        type="email"
        value={email}
        onChange={e => setEmail(e.target.value)}
        required
      />
      <Button type="submit" disabled={isPending}>
        {isPending ? "送信中..." : "リセットリンクを送信"}
      </Button>
    </form>
  );
}
```

### 2. `/reset-password` ページ（`features/auth/routes/ResetPasswordPage.tsx`）

```typescript
export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const navigate = useNavigate();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [isPending, startTransition] = useTransition();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirm) {
      toast.error("パスワードが一致しません");
      return;
    }
    startTransition(async () => {
      await resetPassword({ token, password });
      toast.success("パスワードを変更しました");
      navigate("/login");
    });
  };

  if (!token) {
    return <p>無効なリンクです。</p>;
  }

  return (
    <form onSubmit={handleSubmit}>
      <h1>新しいパスワードを設定</h1>
      <label htmlFor="password">新しいパスワード（8文字以上）</label>
      <input id="password" type="password" value={password}
        onChange={e => setPassword(e.target.value)} required minLength={8} />
      <label htmlFor="confirm">パスワード（確認）</label>
      <input id="confirm" type="password" value={confirm}
        onChange={e => setConfirm(e.target.value)} required />
      <Button type="submit" disabled={isPending}>
        {isPending ? "変更中..." : "パスワードを変更"}
      </Button>
    </form>
  );
}
```

### 3. API 関数（`features/auth/api/forgot-password.ts`, `reset-password.ts`）

```typescript
// features/auth/api/forgot-password.ts
export async function forgotPassword(input: { email: string }) {
  const { data } = await axios.post("/api/v1/auth/forgot-password", input);
  return data;
}

// features/auth/api/reset-password.ts
export async function resetPassword(input: { token: string; password: string }) {
  const { data } = await axios.post("/api/v1/auth/reset-password", input);
  return data;
}
```

### 4. ルーター登録（`app/router.tsx`）

```typescript
// 認証不要のパブリックルート
{
  path: "/forgot-password",
  lazy: async () => {
    const { ForgotPasswordPage } = await import("@/features/auth/routes/ForgotPasswordPage");
    return { Component: ForgotPasswordPage };
  },
},
{
  path: "/reset-password",
  lazy: async () => {
    const { ResetPasswordPage } = await import("@/features/auth/routes/ResetPasswordPage");
    return { Component: ResetPasswordPage };
  },
},
```

### 5. ログインページに「パスワードをお忘れですか？」リンク追加

```tsx
// features/auth/routes/LoginPage.tsx
<Link to="/forgot-password" className="text-sm text-muted-foreground hover:underline">
  パスワードをお忘れですか？
</Link>
```

---

## 受入条件

- [ ] ログインページに「パスワードをお忘れですか？」リンクがある
- [ ] `/forgot-password` でメールアドレス送信後「メールを送信しました」が表示される
- [ ] dev 環境では backend ログにトークンが出力される（メール不要）
- [ ] `/reset-password?token=xxx` でパスワード変更後ログインページに遷移
- [ ] 無効トークンの場合はエラートーストが表示される
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
