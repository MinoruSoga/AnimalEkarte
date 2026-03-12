import { useState, useCallback } from "react";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C } from "@/lib/design-tokens";
import { useAuth } from "../hooks/useAuth";
import { MOCK_USERS } from "../api/mock-data";
import { JOB_TITLE_LABELS, USER_TYPE_LABELS } from "../types";
import type { AuthUser } from "../types";

/* ---- Demo Account Row ---- */

interface DemoAccountProps {
  email: string;
  displayName: string;
  user: AuthUser;
  onSelect: (email: string) => void;
}

function DemoAccount({ email, displayName, user, onSelect }: DemoAccountProps) {
  const roleLabel =
    user.userType === "system_admin"
      ? USER_TYPE_LABELS.system_admin
      : user.userType === "clinic_admin"
        ? USER_TYPE_LABELS.clinic_admin
        : user.jobTitle
          ? JOB_TITLE_LABELS[user.jobTitle] || user.jobTitle
          : "スタッフ";

  return (
    <button
      type="button"
      onClick={() => onSelect(email)}
      className={`w-full text-left px-2.5 py-[7px] rounded-[3px] ${C.hoverBgLight} transition-colors flex items-center gap-2.5`}
    >
      <div className={`size-[26px] rounded-full flex items-center justify-center shrink-0 ${C.bgSkeleton}`}>
        <span className={`text-sm ${C.text50}`}>{displayName.charAt(0)}</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className={`text-base ${C.text}`}>{displayName}</span>
          <span className={`text-sm px-1.5 py-px rounded ${C.text50} ${C.bgHover}`}>
            {roleLabel}
          </span>
        </div>
        <span className={`text-sm ${C.text30} block truncate`}>{email}</span>
      </div>
    </button>
  );
}

/* ---- Shared input classes (padding-x set per field to avoid conflict) ---- */
// Figma実測: fontSize=15px, height=~48px, bg=rgba(242,241,238,0.6), borderRadius=3px
const INPUT_BASE = `w-full h-[48px] text-base rounded-[3px] ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 focus:ring-[#038B94] focus:border-transparent disabled:opacity-60`;

/* ---- Login Form ---- */

export function LoginForm() {
  const { login } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSelectDemo = useCallback((demoEmail: string) => {
    setEmail(demoEmail);
    setPassword("password");
    setError(null);
  }, []);

  const handleSubmit = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      setError(null);

      if (!email.trim()) {
        setError("メールアドレスを入力してください");
        return;
      }
      if (!password) {
        setError("パスワードを入力してください");
        return;
      }

      setIsSubmitting(true);
      try {
        await login(email.trim(), password);
      } catch (err) {
        const message = err instanceof Error ? err.message : "ログインに失敗しました";
        setError(message);
      } finally {
        setIsSubmitting(false);
      }
    },
    [email, password, login],
  );

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
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {/* Email */}
        <div className="space-y-1.5">
          <label htmlFor="login-email" className={`text-sm block ${C.text65}`}>
            メールアドレス
          </label>
          <input
            id="login-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => { setEmail(e.target.value); setError(null); }}
            placeholder="例: admin@example.com"
            className={`${INPUT_BASE} px-2.5`}
            aria-invalid={error !== null}
            aria-describedby={error ? "login-error" : undefined}
            disabled={isSubmitting}
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
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              value={password}
              onChange={(e) => { setPassword(e.target.value); setError(null); }}
              placeholder="パスワードを入力"
              className={`${INPUT_BASE} pl-2.5 pr-10`}
              aria-invalid={error !== null}
              aria-describedby={error ? "login-error" : undefined}
              disabled={isSubmitting}
            />
            <button
              type="button"
              onClick={() => setShowPassword((prev) => !prev)}
              className={`absolute right-1 top-1/2 -translate-y-1/2 size-8 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} transition-colors`}
              aria-label={showPassword ? "パスワードを非表示" : "パスワードを表示"}
            >
              {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
            </button>
          </div>
        </div>

        <FormFieldError id="login-error" message={error} />

        {/* Submit */}
        <button
          type="submit"
          className="w-full h-[48px] text-base font-medium rounded-[3px] bg-[#038B94] hover:bg-[#027A82] transition-colors text-white disabled:opacity-60"
          disabled={isSubmitting}
        >
          {isSubmitting ? "ログイン中..." : "ログイン"}
        </button>
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
          {MOCK_USERS.map((cred) => (
            <DemoAccount
              key={cred.email}
              email={cred.email}
              displayName={cred.user.displayName}
              user={cred.user}
              onSelect={handleSelectDemo}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
