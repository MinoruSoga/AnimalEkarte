// React/Framework
import { useActionState } from "react";

// External
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C } from "@/lib/design-tokens";

// Relative
import { changeMyPassword } from "../api/change-password";

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** 変更成功後に呼ばれるコールバック（再ログイン誘導など） */
  onSuccess?: () => void;
}

type FormState = { error?: string; success?: boolean } | null;

export function ChangePasswordDialog({
  open,
  onOpenChange,
  onSuccess,
}: ChangePasswordDialogProps) {
  const [state, formAction] = useActionState<FormState, FormData>(
    async (_prev, formData) => {
      const currentPassword = formData.get("current_password") as string;
      const newPassword = formData.get("new_password") as string;
      const confirmPassword = formData.get("confirm_password") as string;

      if (!currentPassword || !newPassword || !confirmPassword) {
        return { error: "すべての項目を入力してください" };
      }
      if (newPassword.length < 8) {
        return { error: "新しいパスワードは8文字以上で入力してください" };
      }
      if (newPassword !== confirmPassword) {
        return { error: "新しいパスワードと確認用パスワードが一致しません" };
      }

      try {
        await changeMyPassword({ current_password: currentPassword, new_password: newPassword });
        toast.success("パスワードを変更しました。再度ログインしてください。");
        onOpenChange(false);
        onSuccess?.();
        return { success: true };
      } catch {
        // handleApiError already called inside changeMyPassword
        return { error: "パスワードの変更に失敗しました" };
      }
    },
    null,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>パスワード変更</DialogTitle>
        </DialogHeader>
        <form action={formAction} className="flex flex-col gap-4 py-2">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="current_password" className={`text-sm ${C.text}`}>
              現在のパスワード
            </Label>
            <Input
              id="current_password"
              name="current_password"
              type="password"
              autoComplete="current-password"
              required
              placeholder="現在のパスワードを入力"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="new_password" className={`text-sm ${C.text}`}>
              新しいパスワード
              <span className={`ml-1 text-xs ${C.text40}`}>(8文字以上)</span>
            </Label>
            <Input
              id="new_password"
              name="new_password"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              placeholder="新しいパスワードを入力"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="confirm_password" className={`text-sm ${C.text}`}>
              新しいパスワード（確認）
            </Label>
            <Input
              id="confirm_password"
              name="confirm_password"
              type="password"
              autoComplete="new-password"
              required
              placeholder="もう一度入力"
            />
          </div>
          {state?.error ? (
            <p className="text-sm text-red-600" role="alert">
              {state.error}
            </p>
          ) : null}
          <DialogFooter className="gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              キャンセル
            </Button>
            <SubmitButton>変更する</SubmitButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
