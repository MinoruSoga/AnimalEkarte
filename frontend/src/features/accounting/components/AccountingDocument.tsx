// External
import { memo, useMemo } from "react";

// Internal
import { C, STYLE } from "@/lib/design-tokens";
import { useClinicTaxRates } from "@/hooks/use-clinic-tax-rates";
import { formatDate } from "@/lib/format/date";
import { formatCurrency } from "@/lib/format/number";
import { todayJSTISO } from "@/lib/jst-date";
// #190: セクション定数 — R-F2-S9 で src/config/ へ抽出
import {
  DOCUMENT_SECTION_KEYS,
  type DocumentSectionKey,
} from "@/config/accounting-document-sections";

// Types
import type { Accounting, PaymentInfo } from "../types";
import { calcTaxBreakdown, recordedLineNet } from "../lib/tax-breakdown";

type DocumentPaymentInfo = Pick<
  PaymentInfo,
  "totalAmount" | "insuranceAmount" | "billingAmount" | "receivedAmount" | "changeAmount"
>;

export interface ClinicInfo {
  name?: string;
  postalCode?: string;
  address?: string;
  phoneNumber?: string;
  registrationNumber?: string;
  invoiceRegistrationNumber?: string;
  logoUrl?: string | null;
  /** BUG-367: インボイス帳票の軽減税率判定に使用 */
  standardTaxRate?: number;
  reducedTaxRate?: number;
  accountingDocumentShowLogo?: boolean;
  accountingDocumentShowRegistrationWarning?: boolean;
  accountingDocumentShowItemCategory?: boolean;
  accountingDocumentFooterNote?: string;
  // #190: セクション表示/非表示トグルと表示順
  accountingDocumentShowClinicHeader?: boolean;
  accountingDocumentShowOwnerPetInfo?: boolean;
  accountingDocumentShowItemsTable?: boolean;
  accountingDocumentShowPaymentSummary?: boolean;
  accountingDocumentSectionOrder?: string[];
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
 * #179 ①(#190): 帳票ロゴ src 用に http(s) スキームのみ許可する防御ガード。
 * logo_url は病院設定で編集可能なため、javascript:/data: 等の危険スキームや
 * 不正値が <img src> へ流入しないよう描画時に検証する。不正・空は undefined を返す。
 */
function safeImageUrl(url: string | null | undefined): string | undefined {
  if (!url) return undefined;
  try {
    const parsed = new URL(url, window.location.origin);
    return parsed.protocol === "https:" || parsed.protocol === "http:" ? parsed.href : undefined;
  } catch {
    return undefined;
  }
}

/**
 * BUG-367: 明細兼領収書（A4 統合帳票）
 * #190: セクション表示/非表示トグルと表示順を clinic 設定から取得。
 *       accountingDocumentSectionOrder が空 → デフォルト順。
 *       show_xxx が未設定 (undefined) → true（後方互換）。
 */
export const AccountingDocument = memo(function AccountingDocument({
  accounting,
  paymentInfo,
  clinic,
}: AccountingDocumentProps) {
  const currentDate = useMemo(() => formatDate(todayJSTISO()), []);

  const { standardTaxRate: standardRate, reducedTaxRate: reducedRate } = useClinicTaxRates();

  const taxBreakdown = useMemo(
    () => calcTaxBreakdown(accounting.items, standardRate, reducedRate),
    [accounting.items, standardRate, reducedRate],
  );

  const registrationNumber = clinic?.invoiceRegistrationNumber?.trim() ?? "";
  const hasRegistrationNumber = registrationNumber !== "";
  const showRegistrationWarning = clinic?.accountingDocumentShowRegistrationWarning ?? true;
  const showItemCategory = clinic?.accountingDocumentShowItemCategory ?? true;
  const logoSrc = safeImageUrl(clinic?.logoUrl);
  const showLogo = Boolean(clinic?.accountingDocumentShowLogo && logoSrc);
  const footerNote = clinic?.accountingDocumentFooterNote?.trim() ?? "";

  // #190: セクション表示/非表示（未設定は true = 後方互換）
  const showClinicHeader = clinic?.accountingDocumentShowClinicHeader ?? true;
  const showOwnerPetInfo = clinic?.accountingDocumentShowOwnerPetInfo ?? true;
  const showItemsTable = clinic?.accountingDocumentShowItemsTable ?? true;
  const showPaymentSummary = clinic?.accountingDocumentShowPaymentSummary ?? true;

  // #190: 有効なセクション順の計算
  //   rawOrder に既知キーが 1 つ以上あれば採用し、漏れたキーを末尾に追加。
  //   空の場合はデフォルト順 (DOCUMENT_SECTION_KEYS) をそのまま使用する。
  const effectiveOrder = useMemo<DocumentSectionKey[]>(() => {
    const rawOrder = clinic?.accountingDocumentSectionOrder ?? [];
    const validKeys = rawOrder.filter((k): k is DocumentSectionKey =>
      (DOCUMENT_SECTION_KEYS as readonly string[]).includes(k),
    );
    if (validKeys.length === 0) return [...DOCUMENT_SECTION_KEYS];
    const missing = DOCUMENT_SECTION_KEYS.filter((k) => !validKeys.includes(k));
    return [...validKeys, ...missing];
  }, [clinic?.accountingDocumentSectionOrder]);

  // #190: セクション表示フラグマップ
  const sectionVisible: Record<DocumentSectionKey, boolean> = {
    clinic_header: showClinicHeader,
    owner_pet_info: showOwnerPetInfo,
    items_table: showItemsTable,
    payment_summary: showPaymentSummary,
    footer_note: Boolean(footerNote),
  };

  return (
    <div
      className={`${C.bgWhite} p-8 text-sm font-sans flex flex-col gap-6 border mx-auto max-w-2xl print:max-w-none print:w-full print:border-none print:p-0`}
    >
      {/* AC-6: 登録番号未設定警告（セクション順に非依存） */}
      {!hasRegistrationNumber && showRegistrationWarning ? (
        <div
          className={`border ${C.borderRed300} ${C.bgRed50} ${C.textRed700} p-2 text-xs rounded print:hidden`}
        >
          登録番号が未設定です。病院設定から登録してください。適格請求書として無効となります。
        </div>
      ) : null}

      {/* #190: セクションを effectiveOrder 順に描画（非表示はスキップ） */}
      {effectiveOrder.map((key) => {
        if (!sectionVisible[key]) return null;

        if (key === "clinic_header") {
          return (
            <div key="clinic_header" className="flex justify-between items-end border-b pb-4">
              <div>
                <h1 className="text-heading-3 font-bold mb-2">明細兼領収書</h1>
                <p className={C.text60}>No. {accounting.id}</p>
                <p className={C.text60}>発行日: {currentDate}</p>
              </div>
              <div className={`text-right text-xs ${C.text60}`}>
                {showLogo ? (
                  <img
                    src={logoSrc}
                    alt={`${clinic?.name ?? "病院"}ロゴ`}
                    className="max-h-12 ml-auto mb-2 object-contain"
                  />
                ) : null}
                <p className={`font-bold text-base ${C.text} mb-1`}>{clinic?.name}</p>
                <p>
                  〒{clinic?.postalCode} {clinic?.address}
                </p>
                <p>TEL: {clinic?.phoneNumber}</p>
                <p>登録番号: {hasRegistrationNumber ? registrationNumber : "未設定"}</p>
              </div>
            </div>
          );
        }

        if (key === "owner_pet_info") {
          return (
            <div key="owner_pet_info" className="flex justify-between items-start">
              <div className="space-y-1">
                {/* BUG-374 TC-367-03: 飼主名が空のとき「様」のみ表示されないようフォールバック */}
                <div
                  className={`text-xl border-b ${STYLE.printBorder} mb-2 pb-1 inline-block min-w-[250px]`}
                >
                  {accounting.ownerName ? `${accounting.ownerName} 様` : "（飼主名未取得）"}
                </div>
                <p>
                  ペット名: {accounting.petName} ({accounting.petSpecies})
                </p>
              </div>
            </div>
          );
        }

        if (key === "items_table") {
          return (
            <div key="items_table">
              <table className="w-full text-left border-collapse">
                <thead>
                  <tr className={`border-b-2 ${STYLE.printBorder}`}>
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
                            {isReduced ? <span className="mr-1">※</span> : null}
                            {item.name}
                          </div>
                          {item.category && showItemCategory ? (
                            <span className={`text-xs ${C.text50}`}>{item.category}</span>
                          ) : null}
                          {item.discountAmount > 0 ? (
                            <span className={`block text-xs ${C.text50}`}>
                              割引 −{formatCurrency(item.discountAmount)}
                            </span>
                          ) : null}
                        </td>
                        <td className="py-2 text-right text-xs">
                          {ratePercent}%{isReduced ? "※" : ""}
                        </td>
                        <td className="py-2 text-right">{formatCurrency(item.unitPrice)}</td>
                        <td className="py-2 text-center">{item.quantity}</td>
                        <td className="py-2 text-right">{formatCurrency(recordedLineNet(item))}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              <p className={`text-xs ${C.text50} mt-2`}>
                ※ は軽減税率（{taxBreakdown.reducedRatePercent}%）対象品
              </p>
            </div>
          );
        }

        if (key === "payment_summary") {
          return (
            <div key="payment_summary" className="flex justify-end">
              <div className="w-72 space-y-2">
                <div className="flex justify-between font-bold border-b pb-1">
                  <span>合計金額 (税込)</span>
                  <span>{formatCurrency(paymentInfo.totalAmount)}</span>
                </div>

                <div className={`text-xs ${C.text60} space-y-1 pt-2`}>
                  <div className="flex justify-between">
                    <span>{taxBreakdown.standardRatePercent}%対象</span>
                    <span>
                      {formatCurrency(taxBreakdown.standardBase)}（内 消費税{" "}
                      {formatCurrency(taxBreakdown.standardAmount)}）
                    </span>
                  </div>
                  {taxBreakdown.reducedBase !== 0 ? (
                    <div className="flex justify-between">
                      <span>{taxBreakdown.reducedRatePercent}%対象 ※軽減税率</span>
                      <span>
                        {formatCurrency(taxBreakdown.reducedBase)}（内 消費税{" "}
                        {formatCurrency(taxBreakdown.reducedAmount)}）
                      </span>
                    </div>
                  ) : null}
                </div>

                {paymentInfo.insuranceAmount < 0 ? (
                  <div className={`flex justify-between ${C.textStatusGreen} border-b pb-1 pt-2`}>
                    <span>保険適用</span>
                    <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>
                  </div>
                ) : null}

                <div
                  className={`flex justify-between font-bold text-xl pt-2 border-t ${STYLE.printBorder}`}
                >
                  <span>請求金額</span>
                  <span>{formatCurrency(paymentInfo.billingAmount)}</span>
                </div>

                <div className={`flex justify-between text-xs ${C.text60} pt-1`}>
                  <span>お預かり</span>
                  <span>{formatCurrency(paymentInfo.receivedAmount)}</span>
                </div>
                <div className={`flex justify-between text-xs ${C.text60}`}>
                  <span>お釣り</span>
                  <span>{formatCurrency(paymentInfo.changeAmount)}</span>
                </div>
              </div>
            </div>
          );
        }

        if (key === "footer_note") {
          return (
            <p
              key="footer_note"
              className={`text-xs ${C.text60} border-t pt-3 whitespace-pre-wrap`}
            >
              {footerNote}
            </p>
          );
        }

        return null;
      })}
    </div>
  );
});
