import { useActionState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, STYLE } from "@/lib/design-tokens";
import { getFormString } from "@/lib/form-data";
import { getForbiddenPrefix } from "@/constants/lstep-auto-tag-prefixes";
import { useCreateOwnerTag } from "../api/create-owner-tag";

interface LstepTagAddDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerId: string;
}

interface FormState {
  error: string | null;
  success: boolean;
}

const INITIAL_STATE: FormState = { error: null, success: false };

export function LstepTagAddDialog({ open, onOpenChange, ownerId }: LstepTagAddDialogProps) {
  const { mutateAsync } = useCreateOwnerTag(ownerId);

  const [state, formAction] = useActionState(
    async (_prevState: FormState, formData: FormData): Promise<FormState> => {
      const tagName = getFormString(formData, "tag_name").trim();

      if (!tagName) {
        return { error: "タグ名を入力してください", success: false };
      }

      const forbidden = getForbiddenPrefix(tagName);
      if (forbidden !== null) {
        return {
          error: `「${forbidden}」から始まるタグは自動管理タグのため手動設定できません`,
          success: false,
        };
      }

      try {
        await mutateAsync({ tag_name: tagName });
        return { error: null, success: true };
      } catch {
        // useCreateOwnerTag の onError が handleApiError 済み。ここでは再通知しない。
        return { error: "タグの追加に失敗しました", success: false };
      }
    },
    INITIAL_STATE,
  );

  useEffect(() => {
    if (state.success) {
      onOpenChange(false);
    }
  }, [state.success, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>タグを追加</DialogTitle>
          <DialogDescription>
            手動タグ名を入力してください。自動管理タグのプレフィックスは使用できません。
          </DialogDescription>
        </DialogHeader>

        <form action={formAction} className="flex flex-col gap-4 pt-2">
          <div className="flex flex-col gap-1.5">
            <label htmlFor="tag_name" className={STYLE.formLabel}>
              タグ名
            </label>
            <input
              id="tag_name"
              name="tag_name"
              type="text"
              autoFocus
              placeholder="例: 要注意, VIP顧客"
              className={`${STYLE.formInput} w-full rounded-md px-3`}
            />
            {state.error !== null ? <p className={`text-sm ${C.danger}`}>{state.error}</p> : null}
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className={`px-4 py-2 text-sm rounded-md ${C.text60} ${C.hoverBgLight} transition-colors`}
            >
              キャンセル
            </button>
            <SubmitButton loadingText="付与中..." colorVariant="primary">
              付与する
            </SubmitButton>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
