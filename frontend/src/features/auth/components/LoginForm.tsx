import { useState, useCallback, memo, useActionState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { isAxiosError } from "axios";
import { C } from "@/lib/design-tokens";
import { parseInternalPath } from "@/lib/internal-navigation";
import { getFormString } from "@/lib/form-data";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { useAuth } from "@/hooks/use-auth";
import { LoginFormBrandHeader, LoginFormCredentialFields } from "./LoginFormSections";

interface DemoCredential {
  email: string;
  displayName: string;
  occupationLabel: string;
  permissionLabel: string;
  clinicLabel: string;
  isSystemAdmin?: boolean;
}

export const SHOW_DEMO =
  import.meta.env.DEV ||
  __VERCEL_ENV__ === "preview" ||
  import.meta.env.VITE_VERCEL_ENV === "preview";

const DEMO_ACCOUNTS: readonly DemoCredential[] = SHOW_DEMO
  ? [
      {
        email: "stg-staff-10000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000003@example.test",
        displayName: "高橋 純子",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000007@example.test",
        displayName: "鈴木 諒平",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000008@example.test",
        displayName: "加藤 茉里",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000025@example.test",
        displayName: "チャン ハン",
        occupationLabel: "看護師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000031@example.test",
        displayName: "近喰 千瞳",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000034@example.test",
        displayName: "川野 称希",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000005@example.test",
        displayName: "冨田 美佳",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000006@example.test",
        displayName: "井冨 和美",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-10000009@example.test",
        displayName: "原 梨吏華",
        occupationLabel: "スタッフ",
        permissionLabel: "一般",
        clinicLabel: "八王子病院",
      },
      {
        email: "stg-staff-20000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000003@example.test",
        displayName: "高橋 純子",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000007@example.test",
        displayName: "鈴木 諒平",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000008@example.test",
        displayName: "加藤 茉里",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000025@example.test",
        displayName: "チャン ハン",
        occupationLabel: "看護師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000031@example.test",
        displayName: "近喰 千瞳",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000034@example.test",
        displayName: "川野 称希",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000005@example.test",
        displayName: "冨田 美佳",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000006@example.test",
        displayName: "井冨 和美",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-20000009@example.test",
        displayName: "原 梨吏華",
        occupationLabel: "スタッフ",
        permissionLabel: "一般",
        clinicLabel: "城東センター病院",
      },
      {
        email: "stg-staff-30000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000003@example.test",
        displayName: "高橋 純子",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000007@example.test",
        displayName: "鈴木 諒平",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000008@example.test",
        displayName: "加藤 茉里",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000025@example.test",
        displayName: "チャン ハン",
        occupationLabel: "看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000031@example.test",
        displayName: "近喰 千瞳",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000034@example.test",
        displayName: "川野 称希",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000005@example.test",
        displayName: "冨田 美佳",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000006@example.test",
        displayName: "井冨 和美",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-30000009@example.test",
        displayName: "原 梨吏華",
        occupationLabel: "スタッフ",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　敷島病院",
      },
      {
        email: "stg-staff-40000021@example.test",
        displayName: "林 文明",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000003@example.test",
        displayName: "高橋 純子",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000007@example.test",
        displayName: "鈴木 諒平",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000008@example.test",
        displayName: "加藤 茉里",
        occupationLabel: "獣医師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000025@example.test",
        displayName: "チャン ハン",
        occupationLabel: "看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000031@example.test",
        displayName: "近喰 千瞳",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000034@example.test",
        displayName: "川野 称希",
        occupationLabel: "動物看護師",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000005@example.test",
        displayName: "冨田 美佳",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000006@example.test",
        displayName: "井冨 和美",
        occupationLabel: "VT",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
      {
        email: "stg-staff-40000009@example.test",
        displayName: "原 梨吏華",
        occupationLabel: "スタッフ",
        permissionLabel: "一般",
        clinicLabel: "ノア動物病院　Hako bu neco",
      },
    ]
  : [];

function readDemoLoginPassword(): string {
  if (!SHOW_DEMO) {
    return "";
  }
  return "password";
}

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
      <div
        className={`size-[36px] rounded-full flex items-center justify-center shrink-0 ${C.bgInactive}`}
      >
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

export const LoginForm = memo(function LoginForm() {
  const { login, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, formData: FormData): Promise<ActionState> => {
      const emailValue = getFormString(formData, "login-email").trim();
      const passwordValue = getFormString(formData, "login-password");

      if (!emailValue)
        return { success: false, error: "メールアドレスを入力してください", timestamp: Date.now() };
      if (!passwordValue)
        return { success: false, error: "パスワードを入力してください", timestamp: Date.now() };

      try {
        await login(emailValue, passwordValue);

        const stateFrom = (location.state as { from?: string })?.from;
        const queryFrom = new URLSearchParams(window.location.search).get("from");
        const from =
          parseInternalPath(stateFrom) ?? parseInternalPath(queryFrom) ?? paths.home.getHref();

        navigate(from, { replace: true });
        return { success: true, error: null, timestamp: Date.now() };
      } catch (err) {
        let msg = "ログインに失敗しました。しばらくしてから再度お試しください";
        if (isAxiosError(err)) {
          if (!err.response) msg = "接続できません。ネットワークをご確認ください";
          else if (err.response.status === 401) msg = "メールアドレスまたはパスワードが違います";
          else if (err.response.status === 403) msg = "このアカウントはアクセスが制限されています";
          else if (err.response.status >= 500)
            msg = "サーバーエラーが発生しました。しばらくしてからお試しください";
        }
        return { success: false, error: msg, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
  }, []);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
  }, []);

  const handleSelectDemo = useCallback((demoEmail: string) => {
    setEmail(demoEmail);
    setPassword(readDemoLoginPassword());
  }, []);

  if (isAuthenticated) {
    return <Navigate to={paths.home.getHref()} replace />;
  }

  return (
    <div className="w-full mx-auto flex flex-col items-center">
      <div className="w-full max-w-[380px]">
        <LoginFormBrandHeader />

        <form id="login-form" action={formAction} noValidate className="space-y-4">
          <LoginFormCredentialFields
            email={email}
            password={password}
            showPassword={showPassword}
            isPending={isPending}
            error={formState.error ?? null}
            onEmailChange={handleEmailChange}
            onPasswordChange={handlePasswordChange}
            onTogglePassword={() => setShowPassword((prev) => !prev)}
          />
        </form>
      </div>

      {SHOW_DEMO && DEMO_ACCOUNTS.length > 0 ? (
        <div data-testid="demo-accounts" className="w-full max-w-[760px] mt-8">
          <div className="flex items-center gap-2 mb-2">
            <div className={`h-px flex-1 ${C.bgLight}`} />
            <span className={`text-sm ${C.text35}`}>デモアカウント</span>
            <div className={`h-px flex-1 ${C.bgLight}`} />
          </div>
          <p className={`text-sm text-center mb-2 ${C.text40}`}>
            {readDemoLoginPassword() !== ""
              ? "パスワードは自動入力されます（全デモ共通）"
              : "デモ用パスワード未設定"}
          </p>
          <div className="max-h-[min(40vh,320px)] overflow-y-auto space-y-px">
            {DEMO_ACCOUNTS.map((cred) => (
              <DemoAccount key={cred.email} {...cred} onSelect={handleSelectDemo} />
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
});
