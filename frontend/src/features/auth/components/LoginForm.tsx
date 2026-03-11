import { useState, useCallback } from "react";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C, PALETTE } from "@/lib/design-tokens";
import { useAuth } from "../hooks/useAuth";
import { MOCK_USERS } from "../api/mock-data";
import { JOB_TITLE_LABELS, USER_TYPE_LABELS } from "../types";
import type { AuthUser } from "../types";

/* ================================================================== */
/*  Demo Account Panel                                                 */
/* ================================================================== */

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
      className={`w-full text-left px-2.5 py-[7px] rounded-[3px] ${C.hoverBgLight} transition-colors group flex items-center gap-2.5`}
    >
      <div className={`size-6 rounded-full flex items-center justify-center shrink-0 ${C.bgSkeleton}`}>
        <span className={`text-xs ${C.text50}`}>{displayName.charAt(0)}</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <span className={`text-sm ${C.text}`}>{displayName}</span>
          <span className={`text-xs px-1.5 py-px rounded-sm ${C.text50} ${C.bgHover}`}>
            {roleLabel}
          </span>
        </div>
        <span className={`text-xs ${C.text30} block truncate`}>{email}</span>
      </div>
    </button>
  );
}

/* ================================================================== */
/*  Login Form                                                         */
/* ================================================================== */

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
        <div className={`inline-flex items-center justify-center size-10 rounded-[3px] mb-4 ${C.bgBrand}`}>
          <Stethoscope className="size-5 text-white" />
        </div>
        <h1
          className={`${C.text} mb-1`}
          style={{ fontSize: "24px", fontWeight: 700, lineHeight: 1.2 }}
        >
          ノア動物病院
        </h1>
        <p className={`text-sm ${C.text50}`}>管理システムにログイン</p>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {/* Email */}
        <div className="space-y-1">
          <label htmlFor="login-email" className={`text-xs block ${C.text65}`}>
            メールアドレス
          </label>
          <input
            id="login-email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              setError(null);
            }}
            placeholder="例: admin@example.com"
            className={`w-full h-11 px-2.5 text-sm rounded-[3px] outline-none transition-all ${C.textPlaceholderFaint} ${C.bgInputLogin}`}
            style={{ color: PALETTE.primary, border: `1px solid ${PALETTE.borderMedium}` }}
            aria-invalid={error !== null}
            aria-describedby={error ? "login-error" : undefined}
            disabled={isSubmitting}
          />
        </div>

        {/* Password */}
        <div className="space-y-1">
          <label htmlFor="login-password" className={`text-xs block ${C.text65}`}>
            パスワード
          </label>
          <div className="relative">
            <input
              id="login-password"
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setError(null);
              }}
              placeholder="パスワードを入力"
              className={`w-full h-11 px-2.5 pr-10 text-sm rounded-[3px] outline-none transition-all ${C.textPlaceholderFaint} ${C.bgInputLogin}`}
              style={{ color: PALETTE.primary, border: `1px solid ${PALETTE.borderMedium}` }}
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

        {/* Error */}
        <FormFieldError id="login-error" message={error} />

        {/* Submit */}
        <button
          type="submit"
          className="w-full h-11 text-sm rounded-[3px] transition-colors text-white disabled:opacity-60"
          style={{ background: PALETTE.brand, fontWeight: 500 }}
          onMouseEnter={(e) => {
            if (!isSubmitting) e.currentTarget.style.background = PALETTE.brandHover;
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = PALETTE.brand;
          }}
          disabled={isSubmitting}
        >
          {isSubmitting ? "ログイン中..." : "ログイン"}
        </button>
      </form>

      {/* Demo accounts */}
      <div className="mt-8">
        <div className="flex items-center gap-2 mb-2">
          <div className={`h-px flex-1 ${C.bgLight}`} />
          <span className={`text-xs ${C.text35}`}>デモアカウント</span>
          <div className={`h-px flex-1 ${C.bgLight}`} />
        </div>
        <p className={`text-xs text-center mb-2 ${C.text40}`}>パスワード: password</p>
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
