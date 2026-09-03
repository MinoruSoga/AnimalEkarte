// React/Framework
import { useState } from "react";

// Internal
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

// Relative
import { C } from "@/lib/design-tokens";
import { H_STYLES } from "../lib/styles";

interface DischargeAlertDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (navigateToAccounting: boolean) => void;
}

export function DischargeAlertDialog({ open, onOpenChange, onConfirm }: DischargeAlertDialogProps) {
  const [navigateToAccounting, setNavigateToAccounting] = useState(false);

  const handleOpenChange = (next: boolean) => {
    if (!next) setNavigateToAccounting(false);
    onOpenChange(next);
  };

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>退院処理を行いますか？</AlertDialogTitle>
          <AlertDialogDescription>
            ステータスを「退院済」に変更し、退院日を本日に設定します。
            <br />
            この操作は取り消せません。
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="flex items-center gap-2 py-2">
          <Checkbox
            id="navigate-to-accounting"
            checked={navigateToAccounting}
            onCheckedChange={(checked) => setNavigateToAccounting(checked === true)}
          />
          <Label htmlFor="navigate-to-accounting" className="text-sm cursor-pointer">
            退院後、そのまま会計画面へ進む
          </Label>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel className={H_STYLES.button.action}>キャンセル</AlertDialogCancel>
          <AlertDialogAction
            onClick={() => onConfirm(navigateToAccounting)}
            className={`${C.bgDanger} ${C.hoverBgDanger90} ${C.textWhite} ${H_STYLES.button.action}`}
          >
            退院処理を実行
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
