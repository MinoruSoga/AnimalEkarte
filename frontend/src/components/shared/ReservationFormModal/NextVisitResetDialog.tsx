import { useState } from "react";

import { CalendarClock } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C, ICON } from "@/lib/design-tokens";
import { todayJSTISO, toJSTWallDate } from "@/lib/jst-date";
import { nextVisitIntentMaxDate, validateNextVisitIntentDate } from "./next-visit-intent-date";

interface NextVisitResetDialogProps {
  open: boolean;
  /** 対象患者名（説明文に使う） */
  petName?: string;
  /** 直前の意図日（YYYY-MM-DD）。再設定のヒントとして説明文に出す */
  previousIntendedDate?: string;
  /** 入力欄の初期値（YYYY-MM-DD）。未指定は空 */
  initialDate?: string;
  /**
   * YYYY-MM-DD を受け取り PUT /v1/reservations/pets/:petId/next-visit-intent 等を呼ぶ。
   * resolve で成功 → ダイアログを閉じる。reject で失敗 → 開いたままインラインエラー表示
   * （キャンセル自体は完了済みなので失敗を握りつぶさず再試行可能にする）。
   */
  onSubmit: (date: string) => Promise<void>;
  onClose: () => void;
  /**
   * 送信中フラグ。呼び出し側で setIntentMutation.isPending を渡す
   * （手動 useState ではなく React Query の mutation 状態を使う規約）。
   */
  isPending?: boolean;
}

/**
 * EMR-308: 「次回予定を再設定」共有ダイアログ。
 * tentative 由来の予約キャンセル後（had_intended_date=true）や期限超過 tentative の
 * 再設定導線から使う。自動作成はしない — スタッフが日付を選んで明示確定した時だけ
 * next-visit-intent PUT が走る。
 */
export function NextVisitResetDialog({
  open,
  petName,
  previousIntendedDate,
  initialDate = "",
  onSubmit,
  onClose,
  isPending = false,
}: NextVisitResetDialogProps) {
  const [date, setDate] = useState(initialDate);
  const [error, setError] = useState<string | null>(null);
  // open 遷移でフォームをリセット（render 中の派生 state リセットパターン）
  const [prevOpen, setPrevOpen] = useState(open);
  if (open !== prevOpen) {
    setPrevOpen(open);
    if (open) {
      setDate(initialDate);
      setError(null);
    }
  }

  const handleSubmit = async () => {
    const message = date === "" ? "日付を選択してください" : validateNextVisitIntentDate(date);
    if (message) {
      setError(message);
      return;
    }
    try {
      await onSubmit(date);
      onClose();
    } catch {
      // mutation 側の onError(handleApiError) が toast 済み。ダイアログは開いたまま
      // インラインでも失敗を示す（「取り消しは成功したが再設定は失敗」を混同させない）。
      setError("次回予定の再設定に失敗しました。日付を確認して再度お試しください。");
    }
  };

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <DialogContent className={`sm:max-w-[400px] ${C.bgWhite}`}>
        <DialogHeader>
          <div className="flex items-center gap-2">
            <CalendarClock className={`${ICON.action} ${C.textActionPrimary}`} />
            <DialogTitle className={`text-sm font-bold ${C.text}`}>
              次回予定を再設定しますか？
            </DialogTitle>
          </div>
          <DialogDescription className={`text-xs ${C.text60}`}>
            {petName ? `${petName}ちゃんの` : "このペットの"}次回予定
            {previousIntendedDate ? `（前回: ${previousIntendedDate}）` : ""}
            を新しい日付で設定できます。不要な場合は閉じてください。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2 px-1">
          <Label htmlFor="next-visit-reset-date" className={`font-semibold ${C.text}`}>
            予定日
          </Label>
          <DatePicker
            id="next-visit-reset-date"
            ariaLabel="予定日"
            value={date}
            disabledDays={{
              before: toJSTWallDate(new Date(`${todayJSTISO()}T00:00:00+09:00`)),
              after: toJSTWallDate(new Date(`${nextVisitIntentMaxDate()}T00:00:00+09:00`)),
            }}
            onChange={(value) => {
              setDate(value);
              setError(null);
            }}
            disabled={isPending}
          />
          <FormFieldError message={error ?? undefined} />
        </div>

        <DialogFooter className="flex-row justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={isPending}
            className="h-11 text-sm"
          >
            設定しない
          </Button>
          <Button
            type="button"
            onClick={() => void handleSubmit()}
            disabled={isPending}
            className={`h-11 text-sm ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid}`}
          >
            {isPending ? "設定中..." : "設定する"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
