import { useActionState } from "react";
import { Link } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { getFormString } from "@/lib/form-data";
import { forgotPassword } from "../api/forgot-password";

const INPUT_BASE = `w-full h-[48px] text-base rounded-xxs ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 ${C.focusRingActionPrimary} focus:border-transparent disabled:opacity-60`;

type ForgotPasswordState =
  | { status: "idle"; error: null }
  | { status: "sent" }
  | { status: "error"; error: string };

const INITIAL_STATE: ForgotPasswordState = { status: "idle", error: null };

export function ForgotPasswordPage() {
  const [state, formAction] = useActionState(
    async (
      _prev: ForgotPasswordState,
      formData: FormData,
    ): Promise<ForgotPasswordState> => {
      const email = getFormString(formData, "forgot-email").trim();
      if (!email) {
        return { status: "error", error: "メールアドレスを入力してください" };
      }

      try {
        await forgotPassword({ email });
      } catch {
        // FE6-6: anti-enumeration — 失敗理由（アドレス不存在・ネットワーク断など）に
        // かかわらず常に成功表示にする。handleApiError の toast はここでは意図的に呼ばない
        // （呼ぶと catch に入った事実自体が UI に可視化され、列挙防止が部分的に破れるため）。
      }
      // セキュリティ上、アドレス不存在でも成功として扱う
      return { status: "sent" };
    },
    INITIAL_STATE,
  );

  const errorMessage = state.status === "error" ? state.error : null;

  return (
    <div className={`min-h-screen flex items-center justify-center ${C.bgPage} p-4`}>
      <div className="w-full max-w-[380px] mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div
            data-testid="forgot-password-brand-mark"
            className={`inline-flex items-center justify-center size-[48px] rounded-xl mb-4 ${C.bgBrandIdentity}`}
          >
            <Stethoscope className={`size-[26px] ${C.textWhite}`} />
          </div>
          <h1 className={`text-heading-3 font-bold leading-tight ${C.text} mb-1`}>
            パスワードのリセット
          </h1>
          <p className={`text-base ${C.text50}`}>
            登録済みのメールアドレスを入力してください
          </p>
        </div>

        {state.status === "sent" ? (
          <div className="space-y-4">
            <div className={`rounded-xxs p-4 ${C.bgStatusGreen} border ${C.borderStatusGreen}`}>
              <p className={`text-sm ${C.text}`}>
                パスワードリセットのリンクをメールに送信しました。メールをご確認ください。
              </p>
            </div>
            <Link
              to={paths.auth.login.getHref()}
              className={`flex min-h-11 items-center justify-center text-center text-sm ${C.textBrand} hover:underline`}
            >
              ログインページに戻る
            </Link>
          </div>
        ) : (
          <form action={formAction} noValidate className="space-y-4">
            <div className="space-y-1.5">
              <label htmlFor="forgot-email" className={`text-sm block ${C.text65}`}>
                メールアドレス
              </label>
              <input
                id="forgot-email"
                name="forgot-email"
                type="email"
                autoComplete="email"
                placeholder="例: user@example.com"
                className={`${INPUT_BASE} px-2.5`}
                aria-invalid={errorMessage !== null}
                aria-describedby={errorMessage ? "forgot-error" : undefined}
              />
            </div>

            <FormFieldError id="forgot-error" message={errorMessage} />

            <SubmitButton
              colorVariant="brand"
              className="w-full h-[52px]"
              loadingText="送信中..."
            >
              リセットリンクを送信
            </SubmitButton>

            <Link
              to={paths.auth.login.getHref()}
              className={`flex min-h-11 items-center justify-center text-center text-sm ${C.textBrand} hover:underline`}
            >
              ログインページに戻る
            </Link>
          </form>
        )}
      </div>
    </div>
  );
}
