import { useState, useCallback, memo, useActionState } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { isAxiosError } from "axios";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON } from "@/lib/design-tokens";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
// AuthProvider はこのページを囲まないため useAuth() は使用不可。
// login API を直接呼び出し、成功後に navigate("/") で保護ルート側に遷移する。
import { login as loginApi } from "../api/login";

/* ---- Demo accounts (dev only) ---- */

interface DemoCredential {
  email: string;
  displayName: string;
  roleLabel: string;
  permissionLabel: string;
}

const DEMO_ACCOUNTS: readonly DemoCredential[] = [
  { email: "admin@example.com",     displayName: "田中 太郎",  roleLabel: "医院管理者", permissionLabel: "全権限"     },
  { email: "manager@example.com",   displayName: "渡辺 院長",  roleLabel: "管理者",     permissionLabel: "管理者グループ" },
  { email: "exec@example.com",      displayName: "小林 部長",  roleLabel: "執行",       permissionLabel: "執行グループ"  },
  { email: "vet@example.com",       displayName: "山田 花子",  roleLabel: "医師",       permissionLabel: "一般グループ"  },
  { email: "nurse@example.com",     displayName: "佐藤 美咲",  roleLabel: "看護師",     permissionLabel: "一般グループ"  },
  { email: "reception@example.com", displayName: "鈴木 一郎",  roleLabel: "受付",       permissionLabel: "一般グループ"  },
  { email: "trimmer@example.com",   displayName: "高橋 さくら", roleLabel: "トリマー",   permissionLabel: "一般グループ"  },
  { email: "system@example.com",    displayName: "本部 管理者", roleLabel: "運営管理者", permissionLabel: "全権限"     },
];

const DemoAccount = memo(function DemoAccount({
  email,
  displayName,
  roleLabel,
  permissionLabel,
  onSelect,
}: DemoCredential & { onSelect: (email: string) => void }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(email)}
      className={`w-full text-left px-2.5 py-2 rounded-[3px] ${C.hoverBgLight} transition-colors flex items-center gap-3`}
    >
      <div className={`size-[36px] rounded-full flex items-center justify-center shrink-0 ${C.bgInactive}`}>
        <span className={`text-sm font-medium ${C.text65}`}>{displayName.charAt(0)}</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className={`text-sm font-medium ${C.text}`}>{displayName}</span>
          <span className={`text-xs px-1.5 py-px rounded-[3px] ${C.text50} ${C.bgInactive}`}>
            {roleLabel}
          </span>
          <span className={`text-xs px-1.5 py-px rounded-[3px] text-[#038B94] bg-[#038B94]/10`}>
            {permissionLabel}
          </span>
        </div>
        <span className={`text-xs ${C.text35} block truncate`}>{email}</span>
      </div>
    </button>
  );
});

/* ---- Shared input classes (padding-x set per field to avoid conflict) ---- */
// Figma実測: fontSize=15px, height=~48px, bg=rgba(242,241,238,0.6), borderRadius=3px
const INPUT_BASE = `w-full h-[48px] text-base rounded-[3px] ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 focus:ring-[#038B94] focus:border-transparent disabled:opacity-60`;

/* ---- Login Form ---- */

export function LoginForm() {
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
      await loginApi(emailValue, passwordValue);

      // 1. location.state から取得 (内部遷移)
      // 2. URL クエリパラメータから取得 (Axios インターセプター等からの強制遷移)
      const searchParams = new URLSearchParams(window.location.search);
      const queryFrom = searchParams.get("from");
      const from = (location.state as { from?: string })?.from || queryFrom || "/";

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
    setEmail(demoEmail);
    setPassword("password");
  }, []);

  return (
    <div className="w-full max-w-[380px] mx-auto">
      {/* Header */}
      <div className="text-center mb-8">
        <div className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrand}`}>
          <Stethoscope className="size-[26px] text-white" />
        </div>
        <h1 className={`text-[24px] font-bold leading-tight ${C.text} mb-1`}>
          ノア動物病院
        </h1>
        <p className={`text-base ${C.text50}`}>管理システムにログイン</p>
      </div>

      {/* Form */}
      <form action={formAction} noValidate className="space-y-4">
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
              className={`${INPUT_BASE} pl-2.5 pr-10`}
              aria-invalid={formState.error !== null}
              aria-describedby={formState.error ? "login-error" : undefined}
              disabled={isPending}
            />
            <button
              type="button"
              onClick={() => setShowPassword((prev) => !prev)}
              className={`absolute right-1 top-1/2 -translate-y-1/2 size-8 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} transition-colors`}
              aria-label={showPassword ? "パスワードを非表示" : "パスワードを表示"}
            >
              {showPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
            </button>
          </div>
        </div>

        <FormFieldError id="login-error" message={formState.error} />

        {/* Submit */}
        <SubmitButton
          className="w-full h-[52px] text-base font-medium rounded-[3px] bg-[#038B94] hover:bg-[#027A82] transition-colors text-white"
          loadingText="ログイン中..."
        >
          ログイン
        </SubmitButton>

        <div className="text-center">
          <Link
            to="/forgot-password"
            className={`text-sm ${C.text50} hover:underline`}
          >
            パスワードをお忘れですか？
          </Link>
        </div>
      </form>

      {/* Demo accounts */}
      <div className="mt-8">
        <div className="flex items-center gap-2 mb-2">
          <div className={`h-px flex-1 ${C.bgLight}`} />
          <span className={`text-sm ${C.text35}`}>デモアカウント</span>
          <div className={`h-px flex-1 ${C.bgLight}`} />
        </div>
        <p className={`text-sm text-center mb-2 ${C.text40}`}>パスワード: password</p>
        <div className="space-y-px">
          {DEMO_ACCOUNTS.map((cred) => (
            <DemoAccount
              key={cred.email}
              email={cred.email}
              displayName={cred.displayName}
              roleLabel={cred.roleLabel}
              permissionLabel={cred.permissionLabel}
              onSelect={handleSelectDemo}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
