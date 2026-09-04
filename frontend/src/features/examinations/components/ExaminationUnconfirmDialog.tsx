import { useActionState, useState } from "react";

import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";

const MAX_UNCONFIRM_REASON_LENGTH = 500;

interface ExaminationUnconfirmDialogProps {
  onUnconfirm: (reason: string) => Promise<boolean>;
}

interface UnconfirmFormProps extends ExaminationUnconfirmDialogProps {
  onClose: () => void;
}

interface UnconfirmFormState {
  error: string | null;
}

function UnconfirmForm({ onUnconfirm, onClose }: UnconfirmFormProps) {
  const [reasonLength, setReasonLength] = useState(0);
  const [state, formAction] = useActionState(
    async (_previous: UnconfirmFormState, formData: FormData): Promise<UnconfirmFormState> => {
      const reasonEntry = formData.get("reason");
      const reason = typeof reasonEntry === "string" ? reasonEntry.trim() : "";
      if (!reason) return { error: "確定解除理由は必須です" };
      if (reason.length > MAX_UNCONFIRM_REASON_LENGTH) {
        return {
          error: `確定解除理由は${MAX_UNCONFIRM_REASON_LENGTH}文字以内で入力してください`,
        };
      }

      try {
        if (await onUnconfirm(reason)) {
          onClose();
          return { error: null };
        }
      } catch {
        // The mutation boundary reports the detailed API error; keep this dialog fail-closed.
      }
      return { error: "確定解除に失敗しました" };
    },
    { error: null },
  );

  return (
    <form action={formAction} className="space-y-4">
      {state.error ? (
        <p
          id="examination-unconfirm-error"
          role="alert"
          className={`rounded-xs border px-3 py-2 text-sm ${C.danger} ${C.bgDanger8} ${C.borderDanger20}`}
        >
          {state.error}
        </p>
      ) : null}
      <div className="space-y-1.5">
        <label
          htmlFor="examination-unconfirm-reason"
          className={`block text-sm font-medium ${C.text}`}
        >
          確定解除理由
          <span aria-hidden="true" className={`ml-1 ${C.danger}`}>
            *
          </span>
        </label>
        <Textarea
          id="examination-unconfirm-reason"
          name="reason"
          required
          aria-required="true"
          aria-invalid={Boolean(state.error)}
          aria-describedby="examination-unconfirm-reason-count examination-unconfirm-error"
          maxLength={MAX_UNCONFIRM_REASON_LENGTH}
          rows={4}
          onChange={(event) => setReasonLength(event.target.value.length)}
          className={`text-base ${C.bgWhite} ${C.borderMedium}`}
        />
        <p
          id="examination-unconfirm-reason-count"
          aria-live="polite"
          className={`text-right text-sm ${reasonLength > MAX_UNCONFIRM_REASON_LENGTH ? C.danger : C.text60}`}
        >
          {reasonLength} / {MAX_UNCONFIRM_REASON_LENGTH}
        </p>
      </div>
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose} className="h-11 min-w-11">
          キャンセル
        </Button>
        <SubmitButton loadingText="解除中...">確定を解除する</SubmitButton>
      </div>
    </form>
  );
}

export function ExaminationUnconfirmDialog({ onUnconfirm }: ExaminationUnconfirmDialogProps) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" className={`h-11 min-w-11 text-sm ${C.danger}`}>
          確定解除
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>検査記録の確定解除</DialogTitle>
          <DialogDescription>
            確定時点の記録を保持したまま編集可能な状態へ戻します。理由は監査履歴に記録されます。
          </DialogDescription>
        </DialogHeader>
        {open ? <UnconfirmForm onUnconfirm={onUnconfirm} onClose={() => setOpen(false)} /> : null}
      </DialogContent>
    </Dialog>
  );
}
