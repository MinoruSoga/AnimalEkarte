import { Printer } from "lucide-react";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/format/number";
import { PERIOD_OPTIONS, PERIOD_LABELS, type CashRegisterPeriod } from "../constants";
import { UnifiedClosingSummaryTable } from "./UnifiedClosingSummaryTable";
import { CashReconciliationCard } from "./CashReconciliationCard";
import { BillingDetailTable } from "./BillingDetailTable";
import { ClosePrintArea } from "./ClosePrintArea";
import type { ClosePreviewResult } from "../api/get-cash-register-preview";

interface CashRegisterCloseTargetSectionProps {
  date: string;
  period: CashRegisterPeriod;
  onDateChange: (value: string) => void;
  onPeriodChange: (value: CashRegisterPeriod) => void;
  onEnablePreview: () => void;
}

export function CashRegisterCloseTargetSection({
  date,
  period,
  onDateChange,
  onPeriodChange,
  onEnablePreview,
}: CashRegisterCloseTargetSectionProps) {
  return (
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
            onChange={(e) => onDateChange(e.target.value)}
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
                onClick={() => onPeriodChange(p)}
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
          onClick={onEnablePreview}
          className={`h-11 px-4 text-base rounded-full ${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} transition-colors`}
        >
          プレビュー
        </button>
      </div>
    </section>
  );
}

interface CashRegisterCloseTaxBreakdownProps {
  taxBreakdown: ClosePreviewResult["aggregate"]["taxBreakdown"];
}

function CashRegisterCloseTaxBreakdown({ taxBreakdown }: CashRegisterCloseTaxBreakdownProps) {
  return (
    <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <h2 className={`text-base font-semibold ${C.text} mb-4`}>消費税内訳</h2>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 max-w-lg">
        <div>
          <p className={`text-xs ${C.text40} mb-1`}>標準税率（10%）</p>
          <div className="flex justify-between text-sm py-1">
            <span className={C.text60}>課税対象額</span>
            <span className={C.text}>
              {formatCurrency(taxBreakdown.standard.taxableAmount)}
            </span>
          </div>
          <div className="flex justify-between text-sm py-1">
            <span className={C.text60}>消費税額</span>
            <span className={C.text}>
              {formatCurrency(taxBreakdown.standard.taxAmount)}
            </span>
          </div>
        </div>
        <div>
          <p className={`text-xs ${C.text40} mb-1`}>軽減税率（8%）</p>
          <div className="flex justify-between text-sm py-1">
            <span className={C.text60}>課税対象額</span>
            <span className={C.text}>
              {formatCurrency(taxBreakdown.reduced.taxableAmount)}
            </span>
          </div>
          <div className="flex justify-between text-sm py-1">
            <span className={C.text60}>消費税額</span>
            <span className={C.text}>
              {formatCurrency(taxBreakdown.reduced.taxAmount)}
            </span>
          </div>
        </div>
      </div>
      <div className={`mt-3 pt-3 border-t ${C.borderLight} flex justify-between text-sm font-medium max-w-lg`}>
        <span className={C.text70}>消費税合計</span>
        <span className={C.text}>
          {formatCurrency(taxBreakdown.standard.taxAmount + taxBreakdown.reduced.taxAmount)}
        </span>
      </div>
    </section>
  );
}

interface CashRegisterCloseExecuteFormProps {
  date: string;
  period: CashRegisterPeriod;
  actualCash: string;
  theoreticalCash: number;
  formAction: (formData: FormData) => void;
  onActualCashChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onCancel: () => void;
}

function CashRegisterCloseExecuteForm({
  date,
  period,
  actualCash,
  theoreticalCash,
  formAction,
  onActualCashChange,
  onCancel,
}: CashRegisterCloseExecuteFormProps) {
  return (
    <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <h2 className={`text-base font-semibold ${C.text} mb-4`}>レジ締め実行</h2>
      <form id="cash-register-close-form" action={formAction} className="space-y-4">
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
            onChange={onActualCashChange}
            className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
            placeholder="0"
            required
            aria-invalid={actualCash !== "" && Number(actualCash) < 0}
          />
          {actualCash !== "" && Number(actualCash) < 0 ? (
            <p role="alert" className={`mt-1 text-sm ${C.danger}`}>
              実際のレジ現金は0以上の金額を入力してください
            </p>
          ) : null}
        </div>
        {actualCash !== "" ? (
          <CashReconciliationCard
            theoreticalCash={theoreticalCash}
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
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            キャンセル
          </Button>
          <SubmitButton colorVariant="primary" loadingText="締め中...">締める</SubmitButton>
        </div>
      </form>
    </section>
  );
}

interface CashRegisterClosePreviewProps {
  preview: ClosePreviewResult;
  period: CashRegisterPeriod;
  periodLabel: string;
  clinicName: string;
  actualCash: string;
  formAction: (formData: FormData) => void;
  onActualCashChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onCancel: () => void;
}

export function CashRegisterClosePreview({
  preview,
  period,
  periodLabel,
  clinicName,
  actualCash,
  formAction,
  onActualCashChange,
  onCancel,
}: CashRegisterClosePreviewProps) {
  return (
    <>
      {preview.isHoliday ? (
        <div className={`p-4 rounded-lg ${C.bgNotice} border ${C.borderNotice}`}>
          <p className={`text-base ${C.textNotice}`}>この日は休診日として設定されています</p>
        </div>
      ) : null}

      {preview.isAlreadyClosed ? (
        <div className={`p-4 rounded-lg ${C.bgStatusGreen} border ${C.borderStatusGreen}`}>
          <p className={`text-base font-medium ${C.textStatusGreen}`}>
            {preview.date} {periodLabel} はすでに締め済みです
          </p>
        </div>
      ) : (
        <>
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

          <CashRegisterCloseTaxBreakdown taxBreakdown={preview.aggregate.taxBreakdown} />

          <CashRegisterCloseExecuteForm
            date={preview.date}
            period={period}
            actualCash={actualCash}
            theoreticalCash={preview.aggregate.theoreticalCash}
            formAction={formAction}
            onActualCashChange={onActualCashChange}
            onCancel={onCancel}
          />

          <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
            <h2 className={`text-base font-semibold ${C.text} mb-4`}>
              個別会計明細（{preview.billingDetails.length}件）
            </h2>
            <BillingDetailTable details={preview.billingDetails} />
          </section>
        </>
      )}
    </>
  );
}
