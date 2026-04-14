// External
import { memo, useMemo } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";

// Internal
import { C } from "@/lib/design-tokens";

// Types
import type { Accounting, PaymentInfo } from "../types";

type DocumentPaymentInfo = Pick<
  PaymentInfo,
  "totalAmount" | "insuranceAmount" | "billingAmount" | "receivedAmount" | "changeAmount"
>;

interface ClinicInfo {
  name?: string;
  postalCode?: string;
  address?: string;
  phoneNumber?: string;
  registrationNumber?: string;
  invoiceRegistrationNumber?: string;
  /** BUG-367: インボイス帳票の軽減税率判定に使用 */
  standardTaxRate?: number;
  reducedTaxRate?: number;
}

interface AccountingDocumentProps {
  accounting: Accounting;
  paymentInfo: DocumentPaymentInfo;
  clinic: ClinicInfo | null;
}

// BUG-367: 軽減税率判定の浮動小数誤差吸収
const TAX_RATE_EPSILON = 0.0001;
function approxEqual(a: number, b: number): boolean {
  return Math.abs(a - b) < TAX_RATE_EPSILON;
}

/**
 * BUG-367: 明細兼領収書（A4 統合帳票）
 * 旧: 領収書 (80mm) / 診療明細書 (A4) の 2 帳票を廃止し、
 * 適格請求書要件（消費税法第 57 条の 4）を満たす A4 1 枚に統合。
 */
export const AccountingDocument = memo(function AccountingDocument({ accounting, paymentInfo, clinic }: AccountingDocumentProps) {
  const currentDate = useMemo(
    () => format(new Date(), "yyyy年MM月dd日", { locale: ja }),
    [],
  );

  const reducedRate = clinic?.reducedTaxRate ?? 0.08;
  const standardRate = clinic?.standardTaxRate ?? 0.1;

  const taxBreakdown = useMemo(() => {
    const stdItems = accounting.items.filter(i => approxEqual(i.taxRate, standardRate));
    const redItems = accounting.items.filter(i => approxEqual(i.taxRate, reducedRate));

    const stdBase = stdItems.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);
    const redBase = redItems.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);

    return {
      standardBase: stdBase,
      reducedBase: redBase,
      standardAmount: Math.floor(stdBase * standardRate),
      reducedAmount: Math.floor(redBase * reducedRate),
      standardRatePercent: Math.round(standardRate * 100),
      reducedRatePercent: Math.round(reducedRate * 100),
    };
  }, [accounting.items, standardRate, reducedRate]);

  const registrationNumber = clinic?.invoiceRegistrationNumber?.trim() ?? "";
  const hasRegistrationNumber = registrationNumber !== "";

  return (
    <div className="bg-white p-8 text-sm font-sans flex flex-col gap-6 border mx-auto max-w-2xl print:max-w-none print:w-full print:border-none print:p-0">
      {/* AC-6: 登録番号未設定警告 */}
      {!hasRegistrationNumber ? (
        <div className={`border border-red-300 bg-red-50 text-red-700 p-2 text-xs rounded print:hidden`}>
          登録番号が未設定です。病院設定から登録してください。適格請求書として無効となります。
        </div>
      ) : null}

      <div className="flex justify-between items-end border-b pb-4">
        <div>
          <h1 className="text-2xl font-bold mb-2">明細兼領収書</h1>
          <p className={C.text60}>No. {accounting.id}</p>
          <p className={C.text60}>発行日: {currentDate}</p>
        </div>
        <div className={`text-right text-xs ${C.text60}`}>
          <p className={`font-bold text-base ${C.text} mb-1`}>{clinic?.name}</p>
          <p>〒{clinic?.postalCode} {clinic?.address}</p>
          <p>TEL: {clinic?.phoneNumber}</p>
          <p>登録番号: {hasRegistrationNumber ? registrationNumber : "未設定"}</p>
        </div>
      </div>

      <div className="flex justify-between items-start">
        <div className="space-y-1">
          <div className="text-xl border-b border-black mb-2 pb-1 inline-block min-w-[250px]">
            {accounting.ownerName} 様
          </div>
          <p>ペット名: {accounting.petName} ({accounting.petSpecies})</p>
        </div>
      </div>

      <table className="w-full text-left border-collapse">
        <thead>
          <tr className="border-b-2 border-black">
            <th className="py-2">項目</th>
            <th className="py-2 text-right w-16">税率</th>
            <th className="py-2 text-right">単価</th>
            <th className="py-2 text-center">数量</th>
            <th className="py-2 text-right">金額</th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {accounting.items.map((item) => {
            const isReduced = approxEqual(item.taxRate, reducedRate);
            const ratePercent = Math.round(item.taxRate * 100);
            return (
              <tr key={item.id}>
                <td className="py-2">
                  <div className="font-medium">
                    {isReduced ? <span className="mr-1">※</span> : null}{item.name}
                  </div>
                  {item.category ? <span className={`text-xs ${C.text50}`}>{item.category}</span> : null}
                </td>
                <td className="py-2 text-right text-xs">{ratePercent}%{isReduced ? "※" : ""}</td>
                <td className="py-2 text-right">¥{item.unitPrice.toLocaleString()}</td>
                <td className="py-2 text-center">{item.quantity}</td>
                <td className="py-2 text-right">¥{(item.unitPrice * item.quantity).toLocaleString()}</td>
              </tr>
            );
          })}
        </tbody>
      </table>

      <p className={`text-xs ${C.text50}`}>※ は軽減税率（{taxBreakdown.reducedRatePercent}%）対象品</p>

      <div className="flex justify-end mt-4">
        <div className="w-72 space-y-2">
          <div className="flex justify-between font-bold border-b pb-1">
            <span>合計金額 (税込)</span>
            <span>¥{paymentInfo.totalAmount.toLocaleString()}</span>
          </div>

          <div className={`text-xs ${C.text60} space-y-1 pt-2`}>
            <div className="flex justify-between">
              <span>{taxBreakdown.standardRatePercent}%対象</span>
              <span>¥{taxBreakdown.standardBase.toLocaleString()}（内 消費税 ¥{taxBreakdown.standardAmount.toLocaleString()}）</span>
            </div>
            {taxBreakdown.reducedBase > 0 ? (
              <div className="flex justify-between">
                <span>{taxBreakdown.reducedRatePercent}%対象 ※軽減税率</span>
                <span>¥{taxBreakdown.reducedBase.toLocaleString()}（内 消費税 ¥{taxBreakdown.reducedAmount.toLocaleString()}）</span>
              </div>
            ) : null}
          </div>

          {paymentInfo.insuranceAmount < 0 ? (
            <div className={`flex justify-between ${C.textStatusGreen} border-b pb-1 pt-2`}>
              <span>保険適用</span>
              <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>
            </div>
          ) : null}

          <div className="flex justify-between font-bold text-lg pt-2 border-t border-black">
            <span>請求金額</span>
            <span>¥{paymentInfo.billingAmount.toLocaleString()}</span>
          </div>

          <div className={`flex justify-between text-xs ${C.text60} pt-1`}>
            <span>お預かり</span>
            <span>¥{paymentInfo.receivedAmount.toLocaleString()}</span>
          </div>
          <div className={`flex justify-between text-xs ${C.text60}`}>
            <span>お釣り</span>
            <span>¥{paymentInfo.changeAmount.toLocaleString()}</span>
          </div>
        </div>
      </div>
    </div>
  );
});
