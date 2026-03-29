import { useState, useCallback, useTransition } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { C, ICON } from "@/lib/design-tokens";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { resetPassword } from "../api/reset-password";

const INPUT_BASE = `w-full h-[48px] text-base rounded-[3px] ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 focus:ring-[#038B94] focus:border-transparent disabled:opacity-60`;

export function ResetPasswordPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    setError(null);
  }, []);

  const handleConfirmPasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setConfirmPassword(e.target.value);
    setError(null);
  }, []);

  const handleSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();

      if (!token) {
        setError("無効なリセットリンクです。再度パスワードリセットを申請してください。");
        return;
      }

      const form = e.currentTarget;
      const passwordValue = (form.elements.namedItem("reset-password") as HTMLInputElement).value;
      const confirmValue = (form.elements.namedItem("reset-confirm-password") as HTMLInputElement).value;

      if (!passwordValue) {
        setError("新しいパスワードを入力してください");
        return;
      }
      if (passwordValue.length < 8) {
        setError("パスワードは8文字以上で入力してください");
        return;
      }
      if (passwordValue !== confirmValue) {
        setError("パスワードが一致しません");
        return;
      }

      startTransition(async () => {
        try {
          await resetPassword({ token, password: passwordValue });
          void navigate("/login");
        } catch {
          setError("パスワードのリセットに失敗しました。リンクの有効期限が切れている可能性があります。");
        }
      });
    },
    [token, navigate],
  );

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#F1F0EE] p-4">
      <div className="w-full max-w-[380px] mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrand}`}>
            <Stethoscope className="size-[26px] text-white" />
          </div>
          <h1 className={`text-[24px] font-bold leading-tight ${C.text} mb-1`}>
            新しいパスワードの設定
          </h1>
          <p className={`text-base ${C.text50}`}>8文字以上のパスワードを設定してください</p>
        </div>

        <form onSubmit={handleSubmit} noValidate className="space-y-4">
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
                value={password}
                onChange={handlePasswordChange}
                placeholder="8文字以上で入力"
                className={`${INPUT_BASE} pl-2.5 pr-10`}
                aria-invalid={error !== null}
                aria-describedby={error ? "reset-error" : undefined}
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
                value={confirmPassword}
                onChange={handleConfirmPasswordChange}
                placeholder="同じパスワードを入力"
                className={`${INPUT_BASE} pl-2.5 pr-10`}
                aria-invalid={error !== null}
                aria-describedby={error ? "reset-error" : undefined}
                disabled={isPending}
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword((prev) => !prev)}
                className={`absolute right-1 top-1/2 -translate-y-1/2 size-8 flex items-center justify-center rounded-[3px] ${C.text35} ${C.hoverText} transition-colors`}
                aria-label={showConfirmPassword ? "確認パスワードを非表示" : "確認パスワードを表示"}
              >
                {showConfirmPassword ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
              </button>
            </div>
          </div>

          <FormFieldError id="reset-error" message={error} />

          <button
            type="submit"
            className="w-full h-[52px] text-base font-medium rounded-[3px] bg-[#038B94] hover:bg-[#027A82] transition-colors text-white disabled:opacity-60"
            disabled={isPending}
          >
            {isPending ? "設定中..." : "パスワードを設定する"}
          </button>

          <Link
            to="/login"
            className={`block text-center text-sm ${C.text50} hover:underline`}
          >
            ログインページに戻る
          </Link>
        </form>
      </div>
    </div>
  );
}
