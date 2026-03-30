import { useState, useCallback, useTransition } from "react";
import { Link } from "react-router";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { C } from "@/lib/design-tokens";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { forgotPassword } from "../api/forgot-password";

const INPUT_BASE = `w-full h-[48px] text-base rounded-[3px] ${C.bgInputLogin} border ${C.borderMedium} ${C.text} ${C.textPlaceholder} outline-none transition-all focus:ring-2 focus:ring-[#038B94] focus:border-transparent disabled:opacity-60`;

export function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);
  const [isPending, startTransition] = useTransition();

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    setError(null);
  }, []);

  const handleSubmit = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();

      const form = e.currentTarget;
      const emailValue = (form.elements.namedItem("forgot-email") as HTMLInputElement).value.trim();

      if (!emailValue) {
        setError("メールアドレスを入力してください");
        return;
      }

      startTransition(async () => {
        try {
          await forgotPassword({ email: emailValue });
          setSent(true);
        } catch {
          // セキュリティ上、存在しないアドレスでも成功として扱う
          setSent(true);
        }
      });
    },
    [],
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
            パスワードのリセット
          </h1>
          <p className={`text-base ${C.text50}`}>
            登録済みのメールアドレスを入力してください
          </p>
        </div>

        {sent ? (
          <div className="space-y-4">
            <div className={`rounded-[3px] p-4 bg-[#038B94]/10 border border-[#038B94]/20`}>
              <p className={`text-sm ${C.text}`}>
                パスワードリセットのリンクをメールに送信しました。メールをご確認ください。
              </p>
            </div>
            <Link
              to="/login"
              className={`block text-center text-sm ${C.text50} hover:underline`}
            >
              ログインページに戻る
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} noValidate className="space-y-4">
            <div className="space-y-1.5">
              <label htmlFor="forgot-email" className={`text-sm block ${C.text65}`}>
                メールアドレス
              </label>
              <input
                id="forgot-email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={handleEmailChange}
                placeholder="例: user@example.com"
                className={`${INPUT_BASE} px-2.5`}
                aria-invalid={error !== null}
                aria-describedby={error ? "forgot-error" : undefined}
                disabled={isPending}
              />
            </div>

            <FormFieldError id="forgot-error" message={error} />

            <SubmitButton
              className="w-full h-[52px] text-base font-medium rounded-[3px] bg-[#038B94] hover:bg-[#027A82] transition-colors text-white"
              loadingText="送信中..."
            >
              リセットリンクを送信
            </SubmitButton>

            <Link
              to="/login"
              className={`block text-center text-sm ${C.text50} hover:underline`}
            >
              ログインページに戻る
            </Link>
          </form>
        )}
      </div>
    </div>
  );
}
