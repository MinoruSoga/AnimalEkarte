// React/Framework
import { useState } from "react";

// Internal
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>差し戻し理由</DialogTitle>
          <DialogDescription>
            差し戻しの理由を入力してください。
          </DialogDescription>
        </DialogHeader>
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
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isPending}
          >
            キャンセル
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!reason.trim() || isPending}
            className={`${C.bgDanger} hover:bg-[#EB5757]/90 text-white`}
          >
            差し戻す
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
