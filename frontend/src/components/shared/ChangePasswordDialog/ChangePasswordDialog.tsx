// React/Framework
import { memo, useActionState, useState } from "react";

// External
import { isAxiosError } from "axios";
import Eye from "lucide-react/dist/esm/icons/eye";
import EyeOff from "lucide-react/dist/esm/icons/eye-off";
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { getFormString } from "@/lib/form-data";
import { extractApiErrorMessage, handleApiError } from "@/lib/handle-api-error";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, STYLE } from "@/lib/design-tokens";

interface ChangePasswordInput {
  current_password: string;
  new_password: string;
}

const changeMyPassword = async (input: ChangePasswordInput): Promise<void> => {
  await axios.put("/v1/users/me/password", input);
};

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** 変更成功後に呼ばれるコールバック（再ログイン誘導など） */
  onSuccess?: () => void;
}

type FormState = { error?: string; success?: boolean; currentPassword?: string } | null;

/**
 * パスワード入力フィールド + 表示/非表示トグル。
 * LoginForm の Eye/EyeOff パターンと同じ発想で、フィールドごとに独立した state を持つ。
 */
function PasswordField({
  id,
  label,
  hint,
  autoComplete,
  placeholder,
  minLength,
  show,
  onToggle,
  fieldNameForAria,
  defaultValue,
}: {
  id: string;
  label: string;
  hint?: string;
  autoComplete: "current-password" | "new-password";
  placeholder: string;
  minLength?: number;
  show: boolean;
  onToggle: () => void;
  fieldNameForAria: string;
  defaultValue?: string;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id} className={`text-sm ${C.text}`}>
        {label}
        {hint ? <span className={`ml-1 text-xs ${C.text40}`}>{hint}</span> : null}
      </Label>
      <div className="relative">
        <Input
          id={id}
          name={id}
          type={show ? "text" : "password"}
          autoComplete={autoComplete}
          required
          minLength={minLength}
          placeholder={placeholder}
          defaultValue={defaultValue}
          className="pr-12"
        />
        <button
          type="button"
          onClick={onToggle}
          className={`absolute right-1 top-1/2 -translate-y-1/2 ${STYLE.iconBtn32} ${C.text35} ${C.hoverText}`}
          aria-label={show ? `${fieldNameForAria}を非表示` : `${fieldNameForAria}を表示`}
          aria-pressed={show}
          aria-controls={id}
        >
          {show ? <EyeOff className={ICON.action} /> : <Eye className={ICON.action} />}
        </button>
      </div>
    </div>
  );
}

export const ChangePasswordDialog = memo(function ChangePasswordDialog({
  open,
  onOpenChange,
  onSuccess,
}: ChangePasswordDialogProps) {
  // フィールドごとの表示/非表示。current を晒さずに new/confirm の一致だけ
  // 確認したいケースに対応するため、3 つ独立した state とする。
  const [showCurrent, setShowCurrent] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const [state, formAction] = useActionState<FormState, FormData>(async (_prev, formData) => {
    const currentPassword = getFormString(formData, "current_password");
    const newPassword = getFormString(formData, "new_password");
    const confirmPassword = getFormString(formData, "confirm_password");

    if (!currentPassword || !newPassword || !confirmPassword) {
      return { error: "すべての項目を入力してください", currentPassword };
    }
    if (newPassword.length < 8) {
      return { error: "新しいパスワードは8文字以上で入力してください", currentPassword };
    }
    if (newPassword !== confirmPassword) {
      return { error: "新しいパスワードと確認用パスワードが一致しません", currentPassword };
    }

    try {
      await changeMyPassword({ current_password: currentPassword, new_password: newPassword });
      toast.success("パスワードを変更しました。再度ログインしてください。");
      onOpenChange(false);
      onSuccess?.();
      return { success: true };
    } catch (error) {
      // Axios.create() instance has no isAxiosError static (BUG-016).
      if (isAxiosError(error)) {
        const status = error.response?.status;
        if (status === 401) {
          return { error: "現在のパスワードが正しくありません", currentPassword };
        }
        if (status === 400) {
          return { error: extractApiErrorMessage(error, "パスワード変更"), currentPassword };
        }
      }
      handleApiError(error, "パスワード変更");
      return { error: "パスワードの変更に失敗しました" };
    }
  }, null);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>パスワード変更</DialogTitle>
          <DialogDescription>
            現在のパスワードと新しいパスワードを入力してください。
          </DialogDescription>
        </DialogHeader>
        <form action={formAction} noValidate className="flex flex-col gap-4 py-2">
          <PasswordField
            id="current_password"
            label="現在のパスワード"
            autoComplete="current-password"
            placeholder="現在のパスワードを入力"
            show={showCurrent}
            onToggle={() => setShowCurrent((prev) => !prev)}
            fieldNameForAria="現在のパスワード"
            defaultValue={state?.currentPassword}
          />
          <PasswordField
            id="new_password"
            label="新しいパスワード"
            hint="(8文字以上)"
            autoComplete="new-password"
            placeholder="新しいパスワードを入力"
            minLength={8}
            show={showNew}
            onToggle={() => setShowNew((prev) => !prev)}
            fieldNameForAria="新しいパスワード"
          />
          <PasswordField
            id="confirm_password"
            label="新しいパスワード（確認）"
            autoComplete="new-password"
            placeholder="もう一度入力"
            show={showConfirm}
            onToggle={() => setShowConfirm((prev) => !prev)}
            fieldNameForAria="確認用パスワード"
          />
          {state?.error ? (
            <p className={`text-sm ${C.danger}`} role="alert">
              {state.error}
            </p>
          ) : null}
          <DialogFooter className="gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              キャンセル
            </Button>
            <SubmitButton>変更する</SubmitButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
});
