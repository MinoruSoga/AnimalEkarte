import { Link } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

const RESET_PASSWORD_INPUT_BASE = `w-full h-[48px] text-base rounded-xxs ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 ${C.focusRingActionPrimary} focus:border-transparent disabled:opacity-60`;

export function ResetPasswordInvalidLink() {
  return (
    <div className={`min-h-screen flex items-center justify-center ${C.bgPage} p-4`}>
      <div className="w-full max-w-[380px] mx-auto text-center space-y-4">
        <div
          data-testid="reset-password-invalid-brand-mark"
          className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
        >
          <Stethoscope className={`size-[26px] ${C.textWhite}`} />
        </div>
        <h1 className={`text-heading-3 font-bold ${C.text}`}>無効なリンクです</h1>
        <p className={`text-sm ${C.text50}`}>
          パスワードリセットリンクが無効または期限切れです。再度リセットを申請してください。
        </p>
        <Link
          to={paths.auth.forgotPassword.getHref()}
          className={`inline-flex min-h-11 items-center justify-center text-sm ${C.textBrand} hover:underline`}
        >
          パスワードリセットを再申請する
        </Link>
      </div>
    </div>
  );
}

export function ResetPasswordBrandHeader() {
  return (
    <div className="text-center mb-8">
      <div
        data-testid="reset-password-brand-mark"
        className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
      >
        <Stethoscope className={`size-[26px] ${C.textWhite}`} />
      </div>
      <h1 className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>
        新しいパスワードの設定
      </h1>
      <p className={`text-base ${C.text50}`}>8文字以上のパスワードを設定してください</p>
    </div>
  );
}

interface ResetPasswordFieldsProps {
  showPassword: boolean;
  showConfirmPassword: boolean;
  error: string | null;
  onTogglePassword: () => void;
  onToggleConfirmPassword: () => void;
}

export function ResetPasswordFields({
  showPassword,
  showConfirmPassword,
  error,
  onTogglePassword,
  onToggleConfirmPassword,
}: ResetPasswordFieldsProps) {
  return (
    <>
      <div className="space-y-1.5">
        <label htmlFor="reset-password" className={`text-sm block ${C.text65}`}>
          新しいパスワード
        </label>
        <div className="relative">
          <input
            id="reset-password"
            name="reset-password"
            type={showPassword ? "text" : "password"}
            autoComplete="new-password"
            placeholder="8文字以上で入力"
            className={`${RESET_PASSWORD_INPUT_BASE} pl-2.5 pr-12`}
            aria-invalid={error !== null}
            aria-describedby={error ? "reset-error" : undefined}
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

      <div className="space-y-1.5">
        <label htmlFor="reset-confirm-password" className={`text-sm block ${C.text65}`}>
          パスワード（確認）
        </label>
        <div className="relative">
          <input
            id="reset-confirm-password"
            name="reset-confirm-password"
            type={showConfirmPassword ? "text" : "password"}
            autoComplete="new-password"
            placeholder="同じパスワードを入力"
            className={`${RESET_PASSWORD_INPUT_BASE} pl-2.5 pr-12`}
            aria-invalid={error !== null}
            aria-describedby={error ? "reset-error" : undefined}
          />
          <button
            type="button"
            onClick={onToggleConfirmPassword}
            className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
            aria-label={showConfirmPassword ? "確認パスワードを非表示" : "確認パスワードを表示"}
          >
            {showConfirmPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
          </button>
        </div>
      </div>

      <FormFieldError id="reset-error" message={error} />

      <SubmitButton
        colorVariant="brand"
        className="w-full h-[52px]"
        loadingText="設定中..."
      >
        パスワードを設定する
      </SubmitButton>

      <Link
        to={paths.auth.login.getHref()}
        className={`flex min-h-11 items-center justify-center text-center text-sm ${C.textBrand} hover:underline`}
      >
        ログインページに戻る
      </Link>
    </>
  );
}
