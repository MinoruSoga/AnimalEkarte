import { useActionState, useState, useCallback } from "react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { handleApiError } from "@/lib/handle-api-error";
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
import { useGetCashRegisterPreview } from "../api/get-cash-register-preview";
import { useCreateCashRegisterClose } from "../api/create-cash-register-close";
import { CategoryPaymentMatrix } from "../components/CategoryPaymentMatrix";
import { CashReconciliationCard } from "../components/CashReconciliationCard";
import { BillingDetailTable } from "../components/BillingDetailTable";
import { useCashRegisterCloseForm } from "../hooks/use-cash-register-close-form";

export function CashRegisterClosePage() {
  const { date, period, previewEnabled, handleDateChange, handlePeriodChange, enablePreview } =
    useCashRegisterCloseForm();
  const [actualCash, setActualCash] = useState<string>("");
  const [showConfirm, setShowConfirm] = useState(false);
  const [pendingFormData, setPendingFormData] = useState<FormData | null>(null);

  const { data: preview, isLoading: previewLoading } = useGetCashRegisterPreview(
    date,
    period,
    previewEnabled,
  );
  const createMutation = useCreateCashRegisterClose();

  const handleActualCashChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setActualCash(e.target.value);
  }, []);

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    setPendingFormData(formData);
    setShowConfirm(true);
    return null;
  }, null);

  const handleConfirmClose = useCallback(async () => {
    if (!pendingFormData) return;
    try {
      await createMutation.mutateAsync({
        date: pendingFormData.get("close_date") as string,
        period: pendingFormData.get("period") as "am" | "pm",
        actual_cash: Number(pendingFormData.get("actual_cash")),
        memo: (pendingFormData.get("memo") as string) || undefined,
      });
      toast.success("締めを実行しました");
      setShowConfirm(false);
      setPendingFormData(null);
      setActualCash("");
    } catch (error) {
      handleApiError(error, "締め実行");
      setShowConfirm(false);
    }
  }, [pendingFormData, createMutation]);

  const handleCancelConfirm = useCallback(() => {
    setShowConfirm(false);
    setPendingFormData(null);
  }, []);

  const periodLabel = period === "am" ? "午前" : "午後";

  return (
    <div className="flex-1 overflow-y-auto px-6 py-6 space-y-6">
      <h1 className={`text-xl font-bold ${C.text}`}>レジ締め</h1>

      {/* 対象日・区分選択 */}
      <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
        <h2 className={`text-base font-semibold ${C.text} mb-4`}>対象日時の選択</h2>
        <div className="flex flex-wrap items-end gap-4">
          <div>
            <label htmlFor="target_date" className={STYLE.formLabel}>
              対象日
            </label>
            <input
              id="target_date"
              type="date"
              value={date}
              onChange={(e) => handleDateChange(e.target.value)}
              className={`${STYLE.formInput} mt-1 rounded-[4px] border px-3 block`}
            />
          </div>
          <div>
            <p className={`${STYLE.formLabel} mb-1`}>区分</p>
            <div className="flex gap-2">
              {(["am", "pm"] as const).map((p) => (
                <button
                  key={p}
                  type="button"
                  onClick={() => handlePeriodChange(p)}
                  className={`px-4 h-11 text-base rounded-[4px] border transition-colors ${
                    period === p
                      ? `${C.bgAccent} text-white border-transparent`
                      : `bg-white ${C.borderMedium} ${C.text} ${C.hoverBgLight}`
                  }`}
                >
                  {p === "am" ? "午前" : "午後"}
                </button>
              ))}
            </div>
          </div>
          <button
            type="button"
            onClick={enablePreview}
            className={`h-11 px-5 text-base rounded-[4px] ${C.bgBrand} text-white ${C.hoverBgBrand} transition-colors`}
          >
            プレビュー
          </button>
        </div>
      </section>

      {previewLoading ? (
        <div className="flex items-center justify-center py-8">
          <p className={`text-base ${C.text50}`}>読み込み中...</p>
        </div>
      ) : null}

      {preview !== undefined && !previewLoading ? (
        <>
          {preview.is_holiday ? (
            <div className={`p-4 rounded-lg ${C.bgNotice} border ${C.borderNotice}`}>
              <p className={`text-base ${C.textNotice}`}>この日は休診日として設定されています</p>
            </div>
          ) : null}

          {preview.is_already_closed ? (
            <div className={`p-4 rounded-lg ${C.bgStatusGreen} border ${C.borderStatusGreen}`}>
              <p className={`text-base font-medium ${C.textStatusGreen}`}>
                {date} {periodLabel} はすでに締め済みです
              </p>
            </div>
          ) : (
            <>
              {/* カテゴリ×支払方法マトリクス */}
              <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
                <h2 className={`text-base font-semibold ${C.text} mb-4`}>部門別集計</h2>
                <CategoryPaymentMatrix
                  categories={preview.aggregate.categories}
                  paymentMethods={preview.aggregate.payment_methods}
                />
              </section>

              {/* 締めフォーム */}
              <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
                <h2 className={`text-base font-semibold ${C.text} mb-4`}>レジ締め実行</h2>
                <form action={formAction} className="space-y-4">
                  <input type="hidden" name="close_date" value={date} />
                  <input type="hidden" name="period" value={period} />
                  <div className="max-w-sm">
                    <label htmlFor="actual_cash" className={STYLE.formLabel}>
                      実際のレジ現金（円）
                    </label>
                    <input
                      id="actual_cash"
                      name="actual_cash"
                      type="number"
                      min="0"
                      value={actualCash}
                      onChange={handleActualCashChange}
                      className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
                      placeholder="0"
                      required
                    />
                  </div>
                  {actualCash !== "" ? (
                    <CashReconciliationCard
                      theoreticalCash={preview.aggregate.theoretical_cash}
                      actualCash={Number(actualCash)}
                    />
                  ) : null}
                  <div>
                    <label htmlFor="close_memo" className={STYLE.formLabel}>
                      メモ（任意）
                    </label>
                    <input
                      id="close_memo"
                      name="memo"
                      type="text"
                      className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
                      placeholder="特記事項があれば入力"
                    />
                  </div>
                  <div className="flex justify-end">
                    <SubmitButton loadingText="締め中...">締める</SubmitButton>
                  </div>
                </form>
              </section>

              {/* 個別会計明細 */}
              <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
                <h2 className={`text-base font-semibold ${C.text} mb-4`}>
                  個別会計明細（{preview.billing_details.length}件）
                </h2>
                <BillingDetailTable details={preview.billing_details} />
              </section>
            </>
          )}
        </>
      ) : null}

      <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>締めを実行しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              {date} {periodLabel} の締めを実行します。この操作は取り消せません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={handleCancelConfirm}>キャンセル</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmClose}
              className={`${C.bgAccent} text-white ${C.bgAccentHover}`}
            >
              締める
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
