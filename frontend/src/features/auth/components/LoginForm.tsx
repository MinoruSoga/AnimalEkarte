import { useState, useCallback, memo, useActionState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router";
import { paths } from "@/config/paths";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { isAxiosError } from "axios";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { parseInternalPath } from "@/lib/internal-navigation";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { useAuth } from "../hooks/use-auth";

/* ---- Demo accounts (dev only) ---- */

interface DemoCredential {
  email: string;
  displayName: string;
  occupationLabel: string;
  permissionLabel: string;
  clinicLabel: string;
  isSystemAdmin?: boolean;
}

// M-10 (#91) / SEC-CS2-F01: local Vite DEV only。preview/production では非表示。
// ロジックは computeShowDemoAccounts (show-demo-accounts.ts) と同一 — 定数畳み込み維持のためインライン化。
// export はテスト用（本番バンドルの tree-shake には影響しない — 参照は test 側 dynamic import のみ）。
export const SHOW_DEMO = import.meta.env.DEV;

// staff-attach 実在アカウント（email は stg-staff-{id}@example.test）。
// SHOW_DEMO 分岐で定数畳み込みし、非 DEV バンドルから tree-shake する。
// パスワードは UI / 配列に置かない（repo 外 secrets）。ラベルは roster+handoff 由来。
const DEMO_ACCOUNTS: readonly DemoCredential[] = SHOW_DEMO
  ? [
      { email: "stg-staff-11000021@example.test", displayName: "林 文明", occupationLabel: "獣医師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000003@example.test", displayName: "高橋 純子", occupationLabel: "獣医師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000007@example.test", displayName: "鈴木 諒平", occupationLabel: "獣医師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000008@example.test", displayName: "加藤 茉里", occupationLabel: "獣医師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000025@example.test", displayName: "チャン ハン", occupationLabel: "看護師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000031@example.test", displayName: "近喰 千瞳", occupationLabel: "動物看護師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000034@example.test", displayName: "川野 称希", occupationLabel: "動物看護師", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000005@example.test", displayName: "冨田 美佳", occupationLabel: "VT", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000006@example.test", displayName: "井冨 和美", occupationLabel: "VT", permissionLabel: "一般", clinicLabel: "城東センター病院" },
      { email: "stg-staff-11000009@example.test", displayName: "原 梨吏華", occupationLabel: "スタッフ", permissionLabel: "一般", clinicLabel: "城東センター病院" },
    ]
  : [];

const DemoAccount = memo(function DemoAccount({
  email,
  displayName,
  occupationLabel,
  permissionLabel,
  clinicLabel,
  isSystemAdmin,
  onSelect,
}: DemoCredential & { onSelect: (email: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(email)}
      className={`w-full text-left px-2.5 py-2 rounded-xxs ${C.hoverBgLight} transition-colors flex items-center gap-3`}
    >
      <div className={`size-[36px] rounded-full flex items-center justify-center shrink-0 ${C.bgInactive}`}>
        <span className={`text-sm font-medium ${C.text65}`}>{displayName.charAt(0)}</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5 flex-wrap">
          <span className={`text-sm font-medium ${C.text}`}>{displayName}</span>
          <span className={`text-xs px-1.5 py-px rounded-xxs ${C.text50} ${C.bgInactive}`}>
            {occupationLabel}
          </span>
          <span className={`text-xs px-1.5 py-px rounded-xxs ${C.textBrand} ${C.bgBrand10}`}>
            {permissionLabel}
          </span>
          {isSystemAdmin ? (
            <span className={`text-xs px-1.5 py-px rounded-xxs ${C.danger} ${C.bgDanger8}`}>
              システム管理者
            </span>
          ) : null}
        </div>
        <div className="flex items-center gap-1.5">
          <span className={`text-xs ${C.text35} truncate`}>{email}</span>
          <span className={`text-xs ${C.text35}`}>·</span>
          <span className={`text-xs ${C.text50}`}>{clinicLabel}</span>
        </div>
      </div>
    </button>
  );
});

/* ---- Shared input classes (padding-x set per field to avoid conflict) ---- */
// Figma実測: fontSize=15px, height=~48px, bg=warm neutral 60%透過（PALETTE.hoverBgInput相当の色調）, borderRadius=3px
const INPUT_BASE = `w-full h-[48px] text-base rounded-xxs ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 ${C.focusRingActionPrimary} focus:border-transparent disabled:opacity-60`;

/* ---- Login Form ---- */

export const LoginForm = memo(function LoginForm() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, formData: FormData): Promise<ActionState> => {
      const emailValue = (formData.get("login-email") as string).trim();
      const passwordValue = formData.get("login-password") as string;

      if (!emailValue) return { success: false, error: "メールアドレスを入力してください", timestamp: Date.now() };
      if (!passwordValue) return { success: false, error: "パスワードを入力してください", timestamp: Date.now() };

      try {
        // AuthContext の login() が setUser() を直接呼ぶため、
        // navigate() でそのまま遷移できる（フルリロード不要）。
        await login(emailValue, passwordValue);

        // 1. location.state から取得 (内部遷移)
        // 2. URL クエリパラメータから取得 (Axios インターセプター等からの強制遷移)
        const stateFrom = (location.state as { from?: string })?.from;
        const queryFrom = new URLSearchParams(window.location.search).get("from");
        const from =
          parseInternalPath(stateFrom) ??
          parseInternalPath(queryFrom) ??
          paths.home.getHref();

        navigate(from, { replace: true });
        return { success: true, error: null, timestamp: Date.now() };
      } catch (err) {
        // BUG-047: axios エラーを日本語メッセージに変換
        let msg = "ログインに失敗しました。しばらくしてから再度お試しください";
        if (isAxiosError(err)) {
          if (!err.response) msg = "接続できません。ネットワークをご確認ください";
          else if (err.response.status === 401) msg = "メールアドレスまたはパスワードが違います";
          else if (err.response.status === 403) msg = "このアカウントはアクセスが制限されています";
          else if (err.response.status >= 500) msg = "サーバーエラーが発生しました。しばらくしてからお試しください";
        }
        return { success: false, error: msg, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
  }, []);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
  }, []);

  const handleSelectDemo = useCallback((demoEmail: string) => {
    // 共有デモパスワードは VITE_DEMO_LOGIN_PASSWORD（staff-attach secrets と同一）から。
    // リテラル "password" へフォールバックしない。未設定時は空のまま。
    setEmail(demoEmail);
    const fromEnv =
      typeof import.meta.env.VITE_DEMO_LOGIN_PASSWORD === "string"
        ? import.meta.env.VITE_DEMO_LOGIN_PASSWORD.trim()
        : "";
    setPassword(fromEnv);
  }, []);

  // ログイン済みなら即リダイレクト（直接 /login にアクセスした場合）
  if (isAuthenticated) {
    return <Navigate to={paths.home.getHref()} replace />;
  }

  return (
    <div className="w-full max-w-[380px] mx-auto">
      {/* Header */}
      <div className="text-center mb-8">
        <div className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}>
          <Stethoscope className={`size-[26px] ${C.textWhite}`} />
        </div>
        <h1 className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>
          ノア動物病院
        </h1>
        <p className={`text-base ${C.text50}`}>管理システムにログイン</p>
      </div>

      {/* Form */}
      <form id="login-form" action={formAction} noValidate className="space-y-4">
        {/* Email */}
        <div className="space-y-1.5">
          <label htmlFor="login-email" className={`text-sm block ${C.text65}`}>
            メールアドレス
          </label>
          <input
            id="login-email"
            name="login-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={handleEmailChange}
            placeholder="例: admin@example.com"
            className={`${INPUT_BASE} px-2.5`}
            aria-invalid={formState.error !== null}
            aria-describedby={formState.error ? "login-error" : undefined}
            disabled={isPending}
          />
        </div>

        {/* Password */}
        <div className="space-y-1.5">
          <label htmlFor="login-password" className={`text-sm block ${C.text65}`}>
            パスワード
          </label>
          <div className="relative">
            <input
              id="login-password"
              name="login-password"
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              value={password}
              onChange={handlePasswordChange}
              placeholder="パスワードを入力"
              minLength={6}
              className={`${INPUT_BASE} pl-2.5 pr-12`}
              aria-invalid={formState.error !== null}
              aria-describedby={formState.error ? "login-error" : undefined}
              disabled={isPending}
            />
            <button
              type="button"
              onClick={() => setShowPassword((prev) => !prev)}
              className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
              aria-label={showPassword ? "パスワードを非表示" : "パスワードを表示"}
            >
              {showPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
            </button>
          </div>
        </div>

        <FormFieldError id="login-error" message={formState.error} />

        {/* Submit */}
        <SubmitButton
          colorVariant="brand"
          className="w-full h-[52px]"
          loadingText="ログイン中..."
        >
          ログイン
        </SubmitButton>

        <div className="text-center">
          <Link
            to={paths.auth.forgotPassword.getHref()}
            className={`inline-flex min-h-11 items-center justify-center text-sm ${C.textBrand} hover:underline`}
          >
            パスワードをお忘れですか？
          </Link>
        </div>
      </form>

      {SHOW_DEMO && DEMO_ACCOUNTS.length > 0 ? (
        <div className="mt-8">
          <div className="flex items-center gap-2 mb-2">
            <div className={`h-px flex-1 ${C.bgLight}`} />
            <span className={`text-sm ${C.text35}`}>デモアカウント</span>
            <div className={`h-px flex-1 ${C.bgLight}`} />
          </div>
          <p className={`text-sm text-center mb-2 ${C.text40}`}>
            {typeof import.meta.env.VITE_DEMO_LOGIN_PASSWORD === "string" &&
            import.meta.env.VITE_DEMO_LOGIN_PASSWORD.trim() !== ""
              ? "パスワードは自動入力されます（staff-attach と同一）"
              : "VITE_DEMO_LOGIN_PASSWORD 未設定 — .env.local を staff-attach secrets と揃えてください"}
          </p>
          <div className="max-h-[min(40vh,320px)] overflow-y-auto space-y-px">
            {DEMO_ACCOUNTS.map((cred) => (
              <DemoAccount
                key={cred.email}
                {...cred}
                onSelect={handleSelectDemo}
              />
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
});
