import { useState, useActionState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { C } from "@/lib/design-tokens";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";

import { correctCreditPayment } from "../api/correct-credit-payment";
import {
  getCorrectableCardPayments,
  type CorrectableCardMethod,
} from "./accounting-detail-model";
import type { Accounting } from "../types";

interface CreditCorrectionDialogProps {
  accounting: Accounting;
  // #189/M-2: 対象会計の予定日がレジ締め確定済み期間か（親の判定を受け取り警告を表示する）
  isPostClose?: boolean;
}

// #189: 確定済み会計のクレジット（カード）金額を確定後に訂正する専用導線。
// 確定済み（completed）かつカード系支払いが存在する場合のみ表示する（getCorrectableCardPayments で判定）。
// 通常の updateAccounting (PATCH) ではなく専用 correctCreditPayment エンドポイントを呼ぶ。
export function CreditCorrectionDialog({ accounting, isPostClose = false }: CreditCorrectionDialogProps) {
  const correctable = getCorrectableCardPayments(accounting);
  const [open, setOpen] = useState(false);
  const [method, setMethod] = useState<CorrectableCardMethod>(
    correctable[0]?.method ?? "credit_card",
  );
  const [amount, setAmount] = useState<string>(String(correctable[0]?.amount ?? ""));
  const [reason, setReason] = useState("");
  const [memo, setMemo] = useState("");
  const queryClient = useQueryClient();

  const [, formAction] = useActionState<null, FormData>(async (_prev: null, _formData: FormData) => {
    const amountNum = parseInt(amount, 10);
    if (!reason.trim()) {
      toast.error("訂正理由を入力してください");
      return null;
    }
    if (!Number.isFinite(amountNum) || amountNum <= 0) {
      toast.error("金額は1円以上で入力してください");
      return null;
    }
    try {
      await correctCreditPayment(accounting.id, accounting.clinicId, {
        method,
        amount: amountNum,
        reason: reason.trim(),
        memo: memo.trim() || undefined,
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accounting.id) });
      // BUG-007: 訂正差額が未納者一覧・飼主未納残高に反映されるよう invalidate
      queryClient.invalidateQueries({ queryKey: queryKeys.accounting.unpaidAll() });
      queryClient.invalidateQueries({
        queryKey: queryKeys.ownerUnpaidBalance(accounting.clinicId, accounting.ownerId),
      });
      toast.success("クレジット金額を訂正しました");
      setReason("");
      setMemo("");
      setOpen(false);
      return null;
    } catch (error) {
      handleApiError(error, "クレジット訂正");
      return null;
    }
  }, null);

  // 確定前・カード支払い無しでは導線を出さない（権限ゲートとは別に状態でガード）。
  if (correctable.length === 0) {
    return null;
  }

  const handleMethodChange = (next: CorrectableCardMethod) => {
    setMethod(next);
    const target = correctable.find((c) => c.method === next);
    if (target) {
      setAmount(String(target.amount));
    }
  };

  // ダイアログを開くたびに最新の保存値へ初期化する（訂正後に再度開いた際の stale 値防止）。
  const handleOpenChange = (next: boolean) => {
    if (next) {
      setMethod(correctable[0].method);
      setAmount(String(correctable[0].amount));
      setReason("");
      setMemo("");
    }
    setOpen(next);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" size="sm">
          クレジット訂正
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>クレジット金額の訂正</DialogTitle>
          <DialogDescription>
            確定済み会計のカード金額を訂正します。訂正理由は監査ログに記録されます。
          </DialogDescription>
        </DialogHeader>
        {isPostClose ? (
          <p className={`rounded-md border ${C.borderDanger20} ${C.bgDanger8} p-2 text-sm font-semibold ${C.danger}`}>
            ⚠ この会計はレジ締め確定済み期間です。訂正すると締め時点の帳票と差異が生じます。
          </p>
        ) : null}
        {/* BUG-008/018: HTML5 required/min が JS toast より先にインターセプトしないよう noValidate */}
        <form action={formAction} noValidate className="space-y-4">
          {correctable.length > 1 ? (
            <div className="space-y-1">
              <Label htmlFor="cc-method">支払い手段</Label>
              <select
                id="cc-method"
                value={method}
                onChange={(e) => {
                  const v = e.target.value;
                  if (v === "credit_card" || v === "electronic_money") {
                    handleMethodChange(v);
                  }
                }}
                className="h-9 w-full rounded-md border px-3 text-sm"
              >
                {correctable.map((c) => (
                  <option key={c.method} value={c.method}>
                    {PAYMENT_METHOD_LABELS[c.method]}
                  </option>
                ))}
              </select>
            </div>
          ) : null}
          <div className="space-y-1">
            <Label htmlFor="cc-amount">訂正後の金額</Label>
            <NumberInput
              id="cc-amount"
              value={amount}
              onChange={setAmount}
              min={1}
              suffix="円"
              align="right"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="cc-reason">訂正理由（必須）</Label>
            <Textarea
              id="cc-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="例: レジ端末への入力金額を打ち間違えたため"
              aria-required="true"
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="cc-memo">メモ（任意）</Label>
            <Textarea
              id="cc-memo"
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              キャンセル
            </Button>
            <SubmitButton>訂正を保存</SubmitButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
