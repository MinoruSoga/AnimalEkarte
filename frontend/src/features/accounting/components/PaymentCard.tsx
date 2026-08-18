import { memo, useCallback } from "react";
import { Plus, Save } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { PaymentMethod } from "../types";
import { PAYMENT_METHOD_LABELS } from "@/constants/payment-method";
import { isPaymentSubmitDisabled } from "./accounting-detail-model";

export interface PaymentSplitDraft {
  method: PaymentMethod;
  amount: string;
  receivedAmount: string;
  // #188: お釣り直接上書きモード（レジ実機の誤差吸収）。changeOverride=true の時のみ changeAmount を使う。
  changeOverride?: boolean;
  changeAmount?: string;
}

const PAYMENT_METHODS: PaymentMethod[] = ["cash", "credit_card", "electronic_money", "bank_transfer"];

interface PaymentCardProps {
  billingAmount: number;
  paymentSplits: PaymentSplitDraft[];
  onSplitsChange: (splits: PaymentSplitDraft[]) => void;
  isCompleted: boolean;
  canEdit: boolean;
  canCreate: boolean;
  isEditMode: boolean;
}

export const PaymentCard = memo(function PaymentCard({
  billingAmount,
  paymentSplits,
  onSplitsChange,
  isCompleted,
  canEdit,
  canCreate,
  isEditMode,
}: PaymentCardProps) {
  const canSubmit = isEditMode ? canEdit : canCreate;

  const splitTotal = paymentSplits.reduce((sum, s) => sum + parseInt(s.amount || "0", 10), 0);
  const remaining = billingAmount - splitTotal;

  const isDisabled = isPaymentSubmitDisabled(billingAmount, paymentSplits);

  const handleMethodChange = useCallback((idx: number, method: PaymentMethod) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, method } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleAmountChange = useCallback((idx: number, value: string) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, amount: value } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleReceivedChange = useCallback((idx: number, value: string) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, receivedAmount: value } : s));
  }, [paymentSplits, onSplitsChange]);

  // #188: お釣り手動修正モードの ON/OFF。ON 時は現在の派生値（max(0, received-amount)）を初期値に置く。
  const handleToggleChangeOverride = useCallback((idx: number) => {
    onSplitsChange(paymentSplits.map((s, i) => {
      if (i !== idx) return s;
      if (s.changeOverride) {
        // 自動計算に戻す: 上書きフィールドを除いた基本ドラフトへ戻す
        return { method: s.method, amount: s.amount, receivedAmount: s.receivedAmount };
      }
      const amt = parseInt(s.amount || "0", 10);
      const rec = parseInt(s.receivedAmount || "0", 10);
      return { ...s, changeOverride: true, changeAmount: Math.max(0, rec - amt).toString() };
    }));
  }, [paymentSplits, onSplitsChange]);

  const handleChangeAmountChange = useCallback((idx: number, value: string) => {
    onSplitsChange(paymentSplits.map((s, i) => i === idx ? { ...s, changeAmount: value } : s));
  }, [paymentSplits, onSplitsChange]);

  const handleRemoveSplit = useCallback((idx: number) => {
    onSplitsChange(paymentSplits.filter((_, i) => i !== idx));
  }, [paymentSplits, onSplitsChange]);

  const handleAddSplit = useCallback(() => {
    const rem = billingAmount - paymentSplits.reduce((sum, s) => sum + parseInt(s.amount || "0", 10), 0);
    onSplitsChange([...paymentSplits, { method: "cash", amount: rem > 0 ? rem.toString() : "", receivedAmount: "" }]);
  }, [paymentSplits, onSplitsChange, billingAmount]);

  return (
    <Card className="flex-1">
      <CardHeader className="py-3 px-4 border-b">
        <CardTitle className="text-base font-medium">決済情報</CardTitle>
      </CardHeader>
      <CardContent className="p-6 space-y-6">
        <div className="text-center space-y-1">
          <p className={`text-sm ${C.text50}`}>今回の請求金額</p>
          <p className={`text-heading-1 font-bold ${C.text}`}>
            {formatCurrency(billingAmount)}
          </p>
        </div>

        <Separator />

        {canSubmit ? (
          <div className="space-y-4">
            {paymentSplits.map((split, idx) => {
              const parsedAmount = parseInt(split.amount || "0", 10);
              const parsedReceived = parseInt(split.receivedAmount || "0", 10);
              const splitChange = split.method === "cash" ? parsedReceived - parsedAmount : 0;

              return (
                <div key={idx} className="border rounded-lg p-3 space-y-3">
                  <div className="flex items-center justify-between">
                    <Label className="text-xs font-medium">支払方法</Label>
                    {paymentSplits.length > 1 ? (
                      <DeleteIconButton
                        onClick={() => handleRemoveSplit(idx)}
                        aria-label={`支払${idx + 1}を削除`}
                      />
                    ) : null}
                  </div>
                  <div className="grid grid-cols-3 gap-2">
                    {PAYMENT_METHODS.map((m) => (
                      <Button
                        key={m}
                        type="button"
                        variant={split.method === m ? "default" : "outline"}
                        onClick={() => handleMethodChange(idx, m)}
                        className="h-10 text-sm"
                      >
                        {PAYMENT_METHOD_LABELS[m]}
                      </Button>
                    ))}
                  </div>
                  <div className="space-y-1">
                    <Label htmlFor={`payment-split-${idx}-amount`} className="text-xs">
                      支払{idx + 1}の金額
                    </Label>
                    <NumberInput
                      id={`payment-split-${idx}-amount`}
                      className="h-12 text-xl font-bold"
                      value={split.amount}
                      onChange={(v) => handleAmountChange(idx, v)}
                      suffix="円"
                      align="right"
                    />
                  </div>
                  {split.method === "cash" ? (
                    <>
                      <div className="space-y-1">
                        <Label htmlFor={`payment-split-${idx}-received`} className="text-xs">
                          支払{idx + 1}のお預かり金額
                        </Label>
                        <NumberInput
                          id={`payment-split-${idx}-received`}
                          className="h-12 text-xl font-bold"
                          value={split.receivedAmount}
                          onChange={(v) => handleReceivedChange(idx, v)}
                          suffix="円"
                          align="right"
                        />
                        <div className="flex gap-2 justify-end">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, parsedAmount.toString())}
                          >
                            丁度
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, (Math.ceil(parsedAmount / 1000) * 1000).toString())}
                          >
                            千円単位
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => handleReceivedChange(idx, (Math.ceil(parsedAmount / 10000) * 10000).toString())}
                          >
                            一万単位
                          </Button>
                        </div>
                      </div>
                      <div className={`${C.bgPrimary5} p-3 rounded-lg space-y-2`}>
                        <div className="flex justify-between items-center">
                          <span className={`text-sm font-bold ${C.text60}`}>お釣り</span>
                          <button
                            type="button"
                            onClick={() => handleToggleChangeOverride(idx)}
                            className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-xs px-2 text-xs underline ${C.text50} ${C.hoverText}`}
                          >
                            {split.changeOverride ? "自動計算に戻す" : "手動修正"}
                          </button>
                        </div>
                        {split.changeOverride ? (
                          <>
                            <NumberInput
                              className="h-10 text-xl font-bold"
                              value={split.changeAmount ?? ""}
                              onChange={(v) => handleChangeAmountChange(idx, v)}
                              suffix="円"
                              align="right"
                              min={0}
                            />
                            <p className={`text-xs text-right ${C.text40}`}>
                              レジ実機の実際のお釣りに合わせて手動修正中
                            </p>
                          </>
                        ) : (
                          <div className="flex justify-end">
                            <span className={`text-xl font-bold ${splitChange < 0 ? C.danger : C.text}`}>
                              {formatCurrency(splitChange)}
                            </span>
                          </div>
                        )}
                      </div>
                    </>
                  ) : null}
                </div>
              );
            })}

            {remaining !== 0 ? (
              <p className={`text-xs text-right ${remaining < 0 ? C.danger : C.text50}`}>
                {remaining > 0
                  ? `残り ${formatCurrency(remaining)} 未入力`
                  : `${formatCurrency(Math.abs(remaining))} 超過`}
              </p>
            ) : null}

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddSplit}
              className="w-full"
            >
              <Plus className={`mr-1 ${ICON.action}`} />
              支払方法を追加
            </Button>
          </div>
        ) : (
          <div className="space-y-3">
            {paymentSplits.map((split, idx) => (
              <div key={idx} className="flex justify-between items-center text-sm">
                <span className={C.text50}>{PAYMENT_METHOD_LABELS[split.method] ?? split.method}</span>
                <span className="font-medium">{formatCurrency(parseInt(split.amount || "0", 10))}</span>
              </div>
            ))}
          </div>
        )}

        {canSubmit ? (
          <SubmitButton
            className="w-full h-14 text-xl font-bold mt-4"
            size="lg"
            disabled={isDisabled}
            loadingText="処理中..."
          >
            <Save className={`mr-2 ${ICON.action}`} />
            {isCompleted ? "修正を保存する" : "会計を確定する"}
          </SubmitButton>
        ) : null}
      </CardContent>
    </Card>
  );
});
