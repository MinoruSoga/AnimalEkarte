import { useActionState, useState, useCallback } from "react";
import { Calculator, Printer } from "lucide-react";
import { toast } from "sonner";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { handleApiError } from "@/lib/handle-api-error";
import { useCurrentClinicName } from "@/hooks/use-current-clinic-name";
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
import { PERIOD_OPTIONS, PERIOD_LABELS, type CashRegisterPeriod } from "../constants";
import { UnifiedClosingSummaryTable } from "../components/UnifiedClosingSummaryTable";
import { CashReconciliationCard } from "../components/CashReconciliationCard";
import { BillingDetailTable } from "../components/BillingDetailTable";
import { ClosePrintArea } from "../components/ClosePrintArea";
import { useCashRegisterCloseForm } from "../hooks/use-cash-register-close-form";
import { ResourceCashRegisterClose } from "@/types/generated/models";
import { formatCurrency } from "@/lib/format/number";

export function CashRegisterClosePage() {
  const { date, period, previewEnabled, handleDateChange, handlePeriodChange, enablePreview } =
    useCashRegisterCloseForm();
  const clinicName = useCurrentClinicName();
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
        period: pendingFormData.get("period") as CashRegisterPeriod,
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

  const periodLabel = PERIOD_LABELS[period];

  return (
    <PageLayout
      title="レジ締め"
      resource={ResourceCashRegisterClose}
      icon={<Calculator className={`${ICON.page} ${C.text}`} />}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="space-y-6">
        {/* 対象日・区分選択 */}
        <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
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
                className={`${STYLE.formInput} mt-1 rounded-xs border px-3 block`}
              />
            </div>
            <div>
              <p className={`${STYLE.formLabel} mb-1`}>区分</p>
              <div className="flex gap-2">
                {PERIOD_OPTIONS.map((p) => (
                  <button
                    key={p}
                    type="button"
                    onClick={() => handlePeriodChange(p)}
                    className={`px-4 h-11 text-base rounded-xs border transition-colors ${
                      period === p
                        ? `${C.bgBrand} ${C.textOnBrand} border-transparent`
                        : `${C.bgWhite} ${C.borderMedium} ${C.text} ${C.hoverBgLight}`
                    }`}
                  >
                    {PERIOD_LABELS[p]}
                  </button>
                ))}
              </div>
            </div>
            <button
              type="button"
              onClick={enablePreview}
              className={`h-11 px-4 text-base rounded-full ${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} transition-colors`}
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
            {preview.isHoliday ? (
              <div className={`p-4 rounded-lg ${C.bgNotice} border ${C.borderNotice}`}>
                <p className={`text-base ${C.textNotice}`}>この日は休診日として設定されています</p>
              </div>
            ) : null}

            {preview.isAlreadyClosed ? (
              <div className={`p-4 rounded-lg ${C.bgStatusGreen} border ${C.borderStatusGreen}`}>
                <p className={`text-base font-medium ${C.textStatusGreen}`}>
                  {date} {periodLabel} はすでに締め済みです
                </p>
              </div>
            ) : (
              <>
                {/* 印刷 / PDF 出力導線（#153: #184 と同じ印刷基盤を再利用）*/}
                <div className="flex justify-end print:hidden" data-testid="close-actions">
                  <button
                    type="button"
                    onClick={() => window.print()}
                    data-testid="close-print-button"
                    className={`flex min-h-11 items-center gap-2 px-4 text-base rounded-xs ${C.bgWhite} border ${C.borderMedium} ${C.text} ${C.hoverBgLight} transition-colors`}
                  >
                    <Printer className="size-4" />
                    印刷 / PDF出力
                  </button>
                </div>

                {/* 統合テーブル: 部門別集計（件数＋支払方法別金額＋合計）*/}
                <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
                  <h2 className={`text-base font-semibold ${C.text} mb-4`}>部門別集計</h2>
                  <UnifiedClosingSummaryTable
                    categories={preview.aggregate.categories}
                    paymentMethods={preview.aggregate.paymentMethods}
                    billingDetails={preview.billingDetails}
                    unclassifiedOtherCount={preview.aggregate.unclassifiedOtherCount}
                    categoryCounts={preview.aggregate.categoryCounts}
                  />
                </section>

                {/* 印刷 / PDF 出力ビュー（印刷時のみ表示）*/}
                <ClosePrintArea
                  date={preview.date}
                  period={period}
                  clinicName={clinicName}
                  categories={preview.aggregate.categories}
                  paymentMethods={preview.aggregate.paymentMethods}
                  billingDetails={preview.billingDetails}
                  taxBreakdown={preview.aggregate.taxBreakdown}
                  theoreticalCash={preview.aggregate.theoreticalCash}
                  actualCash={actualCash !== "" ? Number(actualCash) : null}
                  unclassifiedOtherCount={preview.aggregate.unclassifiedOtherCount}
                  categoryCounts={preview.aggregate.categoryCounts}
                />

                {/* 消費税内訳 */}
                <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
                  <h2 className={`text-base font-semibold ${C.text} mb-4`}>消費税内訳</h2>
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 max-w-lg">
                    <div>
                      <p className={`text-xs ${C.text40} mb-1`}>標準税率（10%）</p>
                      <div className="flex justify-between text-sm py-1">
                        <span className={C.text60}>課税対象額</span>
                        <span className={C.text}>
                          {formatCurrency(preview.aggregate.taxBreakdown.standard.taxableAmount)}
                        </span>
                      </div>
                      <div className="flex justify-between text-sm py-1">
                        <span className={C.text60}>消費税額</span>
                        <span className={C.text}>
                          {formatCurrency(preview.aggregate.taxBreakdown.standard.taxAmount)}
                        </span>
                      </div>
                    </div>
                    <div>
                      <p className={`text-xs ${C.text40} mb-1`}>軽減税率（8%）</p>
                      <div className="flex justify-between text-sm py-1">
                        <span className={C.text60}>課税対象額</span>
                        <span className={C.text}>
                          {formatCurrency(preview.aggregate.taxBreakdown.reduced.taxableAmount)}
                        </span>
                      </div>
                      <div className="flex justify-between text-sm py-1">
                        <span className={C.text60}>消費税額</span>
                        <span className={C.text}>
                          {formatCurrency(preview.aggregate.taxBreakdown.reduced.taxAmount)}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div className={`mt-3 pt-3 border-t ${C.borderLight} flex justify-between text-sm font-medium max-w-lg`}>
                    <span className={C.text70}>消費税合計</span>
                    <span className={C.text}>
                      {formatCurrency(
                        preview.aggregate.taxBreakdown.standard.taxAmount +
                          preview.aggregate.taxBreakdown.reduced.taxAmount,
                      )}
                    </span>
                  </div>
                </section>

                {/* 締めフォーム */}
                <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
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
                        className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                        placeholder="0"
                        required
                      />
                    </div>
                    {actualCash !== "" ? (
                      <CashReconciliationCard
                        theoreticalCash={preview.aggregate.theoreticalCash}
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
                        className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                        placeholder="特記事項があれば入力"
                      />
                    </div>
                    <div className="flex justify-end">
                      <SubmitButton colorVariant="primary" loadingText="締め中...">締める</SubmitButton>
                    </div>
                  </form>
                </section>

                {/* 個別会計明細 */}
                <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
                  <h2 className={`text-base font-semibold ${C.text} mb-4`}>
                    個別会計明細（{preview.billingDetails.length}件）
                  </h2>
                  <BillingDetailTable details={preview.billingDetails} />
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
                className={`${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} rounded-full`}
              >
                締める
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </PageLayout>
  );
}
