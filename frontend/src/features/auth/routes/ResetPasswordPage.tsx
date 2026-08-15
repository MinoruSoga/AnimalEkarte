import { useActionState, useState, useCallback, useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { toast } from "sonner";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { resetPassword } from "../api/reset-password";

const INPUT_BASE = `w-full h-[48px] text-base rounded-xxs ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 ${C.focusRingActionPrimary} focus:border-transparent disabled:opacity-60`;

type ResetPasswordState = { error: string | null };

const INITIAL_STATE: ResetPasswordState = { error: null };

function resetTokenFromLocation(search: string, hash: string): string {
  const fragment = hash.startsWith("#") ? hash.slice(1) : hash;
  const fragmentToken = new URLSearchParams(fragment).get("token");
  if (fragmentToken) return fragmentToken;
  return new URLSearchParams(search).get("token") ?? "";
}

export function ResetPasswordPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const [token] = useState(() =>
    resetTokenFromLocation(location.search, location.hash),
  );

  useEffect(() => {
    if (!token || (location.search === "" && location.hash === "")) return;
    void navigate(paths.auth.resetPassword.getHref(), { replace: true });
  }, [location.hash, location.search, navigate, token]);

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const handleTogglePassword = useCallback(() => {
    setShowPassword((prev) => !prev);
  }, []);

  const handleToggleConfirmPassword = useCallback(() => {
    setShowConfirmPassword((prev) => !prev);
  }, []);

  const [state, formAction] = useActionState(
    async (
      _prev: ResetPasswordState,
      formData: FormData,
    ): Promise<ResetPasswordState> => {
      const password = formData.get("reset-password") as string;
      const confirmPassword = formData.get("reset-confirm-password") as string;

      if (!password) {
        return { error: "新しいパスワードを入力してください" };
      }
      if (password.length < 8) {
        return { error: "パスワードは8文字以上で入力してください" };
      }
      if (password !== confirmPassword) {
        return { error: "パスワードが一致しません" };
      }

      try {
        await resetPassword({ token, password });
        toast.success("パスワードを変更しました");
        void navigate(paths.auth.login.getHref());
        return { error: null };
      } catch {
        return { error: "パスワードのリセットに失敗しました。リンクの有効期限が切れている可能性があります。" };
      }
    },
    INITIAL_STATE,
  );

  // token がない場合は無効なリンクとして表示
  if (!token) {
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

  return (
    <div className={`min-h-screen flex items-center justify-center ${C.bgPage} p-4`}>
      <div className="w-full max-w-[380px] mx-auto">
        {/* Header */}
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

        <form action={formAction} noValidate className="space-y-4">
          {/* New Password */}
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
                className={`${INPUT_BASE} pl-2.5 pr-12`}
                aria-invalid={state.error !== null}
                aria-describedby={state.error ? "reset-error" : undefined}
              />
              <button
                type="button"
                onClick={handleTogglePassword}
                className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
                aria-label={showPassword ? "パスワードを非表示" : "パスワードを表示"}
              >
                {showPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
              </button>
            </div>
          </div>

          {/* Confirm Password */}
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
                className={`${INPUT_BASE} pl-2.5 pr-12`}
                aria-invalid={state.error !== null}
                aria-describedby={state.error ? "reset-error" : undefined}
              />
              <button
                type="button"
                onClick={handleToggleConfirmPassword}
                className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
                aria-label={showConfirmPassword ? "確認パスワードを非表示" : "確認パスワードを表示"}
              >
                {showConfirmPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
              </button>
            </div>
          </div>

          <FormFieldError id="reset-error" message={state.error} />

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
        </form>
      </div>
    </div>
  );
}
