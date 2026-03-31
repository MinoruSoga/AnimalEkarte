// React/Framework
import { useState } from "react";

// Internal
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { C } from "@/lib/design-tokens";

interface ReturnReasonDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (reason: string) => void;
  isPending: boolean;
}

export function ReturnReasonDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: ReturnReasonDialogProps) {
  const [reason, setReason] = useState("");

  const handleSubmit = () => {
    if (!reason.trim()) return;
    onSubmit(reason.trim());
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) setReason("");
    onOpenChange(nextOpen);
  };

  return (
    <FormDialog
      open={open}
      onClose={() => handleOpenChange(false)}
      title="差し戻し理由"
      description="差し戻しの理由を入力してください。"
      onSave={handleSubmit}
      saveLabel="差し戻す"
      saveClassName={`${C.bgDanger} hover:bg-[#EB5757]/90 text-white`}
      isPending={isPending}
      isSaveDisabled={!reason.trim()}
    >
      <div className="py-2 space-y-2">
        <Label htmlFor="return-reason">
          差し戻し理由 <span className={C.textRequired}>*</span>
        </Label>
        <Textarea
          id="return-reason"
          placeholder="差し戻しの理由を入力してください"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          rows={4}
        />
      </div>
    </FormDialog>
  );
}
