import { Link } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

const LOGIN_INPUT_BASE = `w-full h-[48px] text-base rounded-xxs ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 ${C.focusRingActionPrimary} focus:border-transparent disabled:opacity-60`;

export function LoginFormBrandHeader() {
  return (
    <div className="text-center mb-8">
      <div
        className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
      >
        <Stethoscope className={`size-[26px] ${C.textWhite}`} />
      </div>
      <h1 className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>ノア動物病院</h1>
      <p className={`text-base ${C.text50}`}>管理システムにログイン</p>
    </div>
  );
}

interface LoginFormCredentialFieldsProps {
  email: string;
  password: string;
  showPassword: boolean;
  isPending: boolean;
  error: string | null;
  onEmailChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onPasswordChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onTogglePassword: () => void;
}

export function LoginFormCredentialFields({
  email,
  password,
  showPassword,
  isPending,
  error,
  onEmailChange,
  onPasswordChange,
  onTogglePassword,
}: LoginFormCredentialFieldsProps) {
  return (
    <>
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
          onChange={onEmailChange}
          placeholder="例: admin@example.com"
          className={`${LOGIN_INPUT_BASE} px-2.5`}
          aria-invalid={error !== null}
          aria-describedby={error ? "login-error" : undefined}
          disabled={isPending}
        />
      </div>

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
            onChange={onPasswordChange}
            placeholder="パスワードを入力"
            minLength={6}
            className={`${LOGIN_INPUT_BASE} pl-2.5 pr-12`}
            aria-invalid={error !== null}
            aria-describedby={error ? "login-error" : undefined}
            disabled={isPending}
          />
          <button
            type="button"
            onClick={onTogglePassword}
            className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
            aria-label={showPassword ? "パスワードを非表示" : "パスワードを表示"}
          >
            {showPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
          </button>
        </div>
      </div>

      <FormFieldError id="login-error" message={error} />

      <SubmitButton colorVariant="brand" className="w-full h-[52px]" loadingText="ログイン中...">
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
    </>
  );
}
