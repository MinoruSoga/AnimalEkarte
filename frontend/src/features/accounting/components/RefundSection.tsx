import { memo, useCallback, useMemo, useState } from "react";
import { Plus, RotateCcw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useGetPaymentMethods } from "@/features/master";
import { C, ICON } from "@/lib/design-tokens";

import { useGetRefunds } from "../api/get-refunds";

const NO_PAYMENT_METHOD = "none";

interface RefundSectionProps {
  accountingId: string;
  totalAmount: number;
  isRefunding: boolean;
  onRefund: (amount: number, reason: string, paymentMethodId?: number) => void;
  canEdit: boolean;
}

export const RefundSection = memo(function RefundSection({
  accountingId,
  totalAmount,
  isRefunding,
  onRefund,
  canEdit,
}: RefundSectionProps) {
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const [refundPaymentMethodId, setRefundPaymentMethodId] = useState(NO_PAYMENT_METHOD);
  const { data: refunds = [] } = useGetRefunds(accountingId);
  const { data: paymentMethods = [] } = useGetPaymentMethods();

  const activePaymentMethods = useMemo(
    () => paymentMethods.filter((m) => m.isActive).sort((a, b) => a.displayOrder - b.displayOrder),
    [paymentMethods],
  );
  // 返金一覧の支払方法表示用: id -> name
  const paymentMethodNameById = useMemo(
    () => new Map(paymentMethods.map((m) => [m.id, m.name])),
    [paymentMethods],
  );

  const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);
  const refundableAmount = totalAmount - totalRefunded;

  const handleSubmit = useCallback(() => {
    const amount = parseInt(refundAmount, 10);
    if (!amount || amount <= 0) return;
    const paymentMethodId =
      refundPaymentMethodId !== NO_PAYMENT_METHOD ? Number(refundPaymentMethodId) : undefined;
    onRefund(amount, refundReason, paymentMethodId);
    setRefundDialogOpen(false);
    setRefundAmount("");
    setRefundReason("");
    setRefundPaymentMethodId(NO_PAYMENT_METHOD);
  }, [refundAmount, refundReason, refundPaymentMethodId, onRefund]);

  return (
    <Card>
      <CardHeader className={`py-3 px-4 ${C.bgSubtle} border-b`}>
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium flex items-center gap-2">
            <RotateCcw className={`${ICON.action} ${C.textDiscount}`} />
            返金管理
            <span className={`text-xs font-normal ${C.text50}`}>
              残額 ¥{refundableAmount.toLocaleString()}
            </span>
            {totalRefunded > 0 ? (
              <span className={`text-xs font-normal ${C.textDiscount} ${C.bgDiscountLight} px-2 py-0.5 rounded`}>
                合計 ¥{totalRefunded.toLocaleString()} 返金済
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
                  disabled={refundableAmount <= 0}
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
                <div className="space-y-4 py-2">
                  <div className="space-y-2">
                    <Label>返金金額（円）</Label>
                    <Input
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
                    <Label>返金理由（任意）</Label>
                    <Input
                      value={refundReason}
                      onChange={(e) => setRefundReason(e.target.value)}
                      placeholder="返金理由を入力..."
                      className="h-10"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>支払方法（任意）</Label>
                    <Select value={refundPaymentMethodId} onValueChange={setRefundPaymentMethodId}>
                      <SelectTrigger data-testid="refund-payment-method-trigger" className="h-10">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value={NO_PAYMENT_METHOD}>指定なし</SelectItem>
                        {activePaymentMethods.map((m) => (
                          <SelectItem key={m.id} value={m.id}>
                            {m.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setRefundDialogOpen(false)}>
                    キャンセル
                  </Button>
                  <Button
                    type="button"
                    onClick={handleSubmit}
                    disabled={!refundAmount || parseInt(refundAmount, 10) <= 0 || isRefunding}
                  >
                    {isRefunding ? "処理中..." : "登録する"}
                  </Button>
                </DialogFooter>
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
                <th className="px-3 py-2 text-left font-medium">日時</th>
                <th className="px-3 py-2 text-left font-medium">処理者</th>
                <th className="px-3 py-2 text-right font-medium">金額</th>
                <th className="px-3 py-2 text-left font-medium">支払方法</th>
                <th className="px-3 py-2 text-left font-medium">理由</th>
              </tr>
            </thead>
            <tbody>
              {refunds.map((r) => (
                <tr key={r.id} className="border-b last:border-0">
                  <td className={`px-3 py-2 font-mono text-xs ${C.text50}`}>
                    {new Date(r.refundedAt).toLocaleDateString("ja-JP")}
                  </td>
                  <td className={`px-3 py-2 text-xs ${C.text50}`}>
                    {r.refundedByName || "-"}
                  </td>
                  <td className={`px-3 py-2 text-right font-medium ${C.textDiscount}`}>
                    ¥{r.amount.toLocaleString()}
                  </td>
                  <td className={`px-3 py-2 text-xs ${C.text50}`}>
                    {r.paymentMethodId ? (paymentMethodNameById.get(r.paymentMethodId) ?? "-") : "-"}
                  </td>
                  <td className={`px-3 py-2 ${C.text50} truncate max-w-[120px]`}>
                    {r.reason || "-"}
                  </td>
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
