import { memo, useActionState, useMemo, useState } from "react";
import { Plus, RotateCcw } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TableCell, TableHead } from "@/components/ui/table";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON } from "@/lib/design-tokens";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { formatJSTDate } from "@/lib/jst-date";
import { formatCurrency } from "@/lib/format/number";

import { useGetRefunds } from "../api/get-refunds";
import type { PaymentMethod, PaymentSplitInfo } from "../types";

const NO_PAYMENT_METHOD = "none";

interface RefundSectionProps {
  accountingId: string;
  totalAmount: number;
  /** この会計で実際に使われた支払方法の内訳。返金の支払方法選択肢の出所（#60）。 */
  paymentSplits: PaymentSplitInfo[];
  isRefunding: boolean;
  onRefund: (amount: number, reason: string, paymentMethod?: PaymentMethod) => void;
  canEdit: boolean;
}

export const RefundSection = memo(function RefundSection({
  accountingId,
  totalAmount,
  paymentSplits,
  isRefunding,
  onRefund,
  canEdit,
}: RefundSectionProps) {
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const [refundPaymentMethod, setRefundPaymentMethod] = useState(NO_PAYMENT_METHOD);
  const { data: refunds = [] } = useGetRefunds(accountingId);

  // この会計で使われた支払方法（ENUM）のユニーク一覧。返金の選択肢にする。
  const usedPaymentMethods = useMemo(() => {
    const seen = new Set<PaymentMethod>();
    for (const s of paymentSplits) {
      if (s.method) seen.add(s.method);
    }
    return [...seen];
  }, [paymentSplits]);

  const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);
  const refundableAmount = totalAmount - totalRefunded;
  const recordedNegative = totalAmount < 0;

  const [, formAction] = useActionState<null, FormData>(
    async (_prev: null, _formData: FormData) => {
      const amount = parseInt(refundAmount, 10);
      if (!amount || amount <= 0) return null;
      if (amount > refundableAmount) {
        toast.error(`返金額は残額 ${formatCurrency(refundableAmount)} 以下で入力してください`);
        return null;
      }
      const paymentMethod =
        refundPaymentMethod !== NO_PAYMENT_METHOD
          ? (refundPaymentMethod as PaymentMethod)
          : undefined;
      onRefund(amount, refundReason, paymentMethod);
      setRefundDialogOpen(false);
      setRefundAmount("");
      setRefundReason("");
      setRefundPaymentMethod(NO_PAYMENT_METHOD);
      return null;
    },
    null,
  );

  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgSubtle} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <RotateCcw className={`${ICON.action} ${C.textDiscount}`} />
            返金管理
            <span id="refund-recorded-amount" className={`text-xs font-normal ${C.text50}`}>
              {recordedNegative
                ? `記録金額 ${formatCurrency(totalAmount)}`
                : `残額 ${formatCurrency(refundableAmount)}`}
            </span>
            {totalRefunded > 0 ? (
              <span
                className={`text-xs font-normal ${C.textDiscount} ${C.bgDiscountLight} px-2 py-0.5 rounded`}
              >
                合計 {formatCurrency(totalRefunded)} 返金済
              </span>
            ) : null}
          </CardTitle>
          {canEdit ? (
            <Dialog open={refundDialogOpen} onOpenChange={setRefundDialogOpen}>
              <DialogTrigger asChild>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8 text-xs"
                  disabled={recordedNegative || refundableAmount <= 0}
                  aria-describedby="refund-recorded-amount"
                >
                  <Plus className={`mr-1 ${ICON.action}`} />
                  返金を登録
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-sm">
                <DialogHeader>
                  <DialogTitle>返金を登録</DialogTitle>
                  <DialogDescription>返金金額と理由を入力してください。</DialogDescription>
                </DialogHeader>
                {/* HTML5 required/min が JS toast より先にインターセプトしないよう noValidate */}
                <form action={formAction} noValidate>
                  <div className="space-y-4 py-2">
                    <div className="space-y-2">
                      <Label htmlFor="refund-amount">返金金額（円）</Label>
                      <Input
                        id="refund-amount"
                        type="number"
                        step={1}
                        min={1}
                        value={refundAmount}
                        onChange={(e) => setRefundAmount(e.target.value)}
                        placeholder="0"
                        className="h-10"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="refund-reason">返金理由（任意）</Label>
                      <Input
                        id="refund-reason"
                        value={refundReason}
                        onChange={(e) => setRefundReason(e.target.value)}
                        placeholder="返金理由を入力..."
                        className="h-10"
                      />
                    </div>
                    {usedPaymentMethods.length > 0 ? (
                      <div className="space-y-2">
                        <Label htmlFor="refund-payment-method-trigger">支払方法（任意）</Label>
                        <Select value={refundPaymentMethod} onValueChange={setRefundPaymentMethod}>
                          <SelectTrigger
                            id="refund-payment-method-trigger"
                            data-testid="refund-payment-method-trigger"
                            className="h-10"
                          >
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value={NO_PAYMENT_METHOD}>指定なし</SelectItem>
                            {usedPaymentMethods.map((m) => (
                              <SelectItem key={m} value={m}>
                                {PAYMENT_METHOD_LABELS[m] ?? m}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    ) : null}
                  </div>
                  <DialogFooter>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setRefundDialogOpen(false)}
                    >
                      キャンセル
                    </Button>
                    <SubmitButton
                      disabled={!refundAmount || parseInt(refundAmount, 10) <= 0 || isRefunding}
                    >
                      {isRefunding ? "処理中..." : "登録する"}
                    </SubmitButton>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          ) : null}
        </div>
      </CardHeader>
      {refunds.length > 0 ? (
        <CardContent className="p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className={`border-b ${C.bgPage30} text-xs`}>
                <TableHead>日時</TableHead>
                <TableHead>処理者</TableHead>
                <TableHead className="text-right">金額</TableHead>
                <TableHead>支払方法</TableHead>
                <TableHead>理由</TableHead>
              </tr>
            </thead>
            <tbody>
              {refunds.map((r) => (
                <tr key={r.id} className="border-b last:border-0">
                  <TableCell className={`font-mono ${C.text50}`}>
                    {formatJSTDate(r.refundedAt)}
                  </TableCell>
                  <TableCell className={C.text50}>{r.refundedByName || "-"}</TableCell>
                  <TableCell className={`text-right font-medium ${C.textDiscount}`}>
                    {formatCurrency(r.amount)}
                  </TableCell>
                  <TableCell className={C.text50}>
                    {r.paymentMethod
                      ? (PAYMENT_METHOD_LABELS[r.paymentMethod as PaymentMethod] ?? r.paymentMethod)
                      : "-"}
                  </TableCell>
                  <TableCell className={`${C.text50} max-w-[120px] truncate`}>
                    {r.reason || "-"}
                  </TableCell>
                </tr>
              ))}
            </tbody>
          </table>
        </CardContent>
      ) : (
        <CardContent className={`p-4 text-center text-sm ${C.text50}`}>
          返金記録はありません
        </CardContent>
      )}
    </Card>
  );
});
