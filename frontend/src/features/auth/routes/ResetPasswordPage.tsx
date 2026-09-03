import { useActionState, useState, useCallback, useEffect } from "react";
import { useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { getFormString } from "@/lib/form-data";
import { resetPassword } from "../api/reset-password";
import {
  ResetPasswordBrandHeader,
  ResetPasswordFields,
  ResetPasswordInvalidLink,
} from "./ResetPasswordPageSections";

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
      const password = getFormString(formData, "reset-password");
      const confirmPassword = getFormString(formData, "reset-confirm-password");

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

  if (!token) {
    return <ResetPasswordInvalidLink />;
  }

  return (
    <div className={`min-h-screen flex items-center justify-center ${C.bgPage} p-4`}>
      <div className="w-full max-w-[380px] mx-auto">
        <ResetPasswordBrandHeader />

        <form action={formAction} noValidate className="space-y-4">
          <ResetPasswordFields
            showPassword={showPassword}
            showConfirmPassword={showConfirmPassword}
            error={state.error}
            onTogglePassword={handleTogglePassword}
            onToggleConfirmPassword={handleToggleConfirmPassword}
          />
        </form>
      </div>
    </div>
  );
}
