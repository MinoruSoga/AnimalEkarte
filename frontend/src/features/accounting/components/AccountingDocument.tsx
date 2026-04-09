// External
import { useMemo } from "react";
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
}

interface AccountingDocumentProps {
  type: "receipt" | "statement";
  accounting: Accounting;
  paymentInfo: DocumentPaymentInfo;
  clinic: ClinicInfo | null;
}

export function AccountingDocument({ type, accounting, paymentInfo, clinic }: AccountingDocumentProps) {
  const currentDate = useMemo(
    () => format(new Date(), "yyyy年MM月dd日", { locale: ja }),
    [],
  );

  const taxBreakdown = useMemo(() => {
    const tax10Items = accounting.items.filter(i => i.taxRate === 0.1);
    const tax8Items = accounting.items.filter(i => i.taxRate === 0.08);

    const tax10Base = tax10Items.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);
    const tax8Base = tax8Items.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);

    return {
      tax10Base,
      tax8Base,
      tax10Amount: Math.floor(tax10Base * 0.1),
      tax8Amount: Math.floor(tax8Base * 0.08),
    };
  }, [accounting.items]);

  if (type === "receipt") {
    return (
      <div className="w-[80mm] bg-white p-4 text-sm font-sans flex flex-col gap-4 border mx-auto print:mx-0 print:border-none">
        <div className="text-center border-b pb-4">
          <h1 className="text-xl font-bold mb-2">領 収 書</h1>
          <p className={`text-xs ${C.text50}`}>{currentDate}</p>
        </div>

        <div className="space-y-1">
          <p className="text-lg font-medium border-b border-black inline-block min-w-[200px]">
            {accounting.ownerName} 様
          </p>
          <p className={`text-xs ${C.text50}`}>ペット名: {accounting.petName}</p>
        </div>

        <div className={`py-6 text-center ${C.bgPage} my-2`}>
          <p className={`text-xs ${C.text50} mb-1`}>領収金額</p>
          <p className="text-2xl font-bold">¥{paymentInfo.billingAmount.toLocaleString()}-</p>
        </div>

        <div className="space-y-2 text-xs">
          <div className="flex justify-between border-b border-dotted pb-1">
            <span>診療費等合計</span>
            <span>¥{paymentInfo.totalAmount.toLocaleString()}</span>
          </div>
          {paymentInfo.insuranceAmount < 0 ? (
            <div className={`flex justify-between border-b border-dotted pb-1 ${C.textStatusGreen}`}>
              <span>保険負担額</span>
              <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>
            </div>
          ) : null}
          <div className="flex justify-between border-b border-dotted pb-1 font-bold">
            <span>請求金額</span>
            <span>¥{paymentInfo.billingAmount.toLocaleString()}</span>
          </div>
          <div className="flex justify-between pt-1">
            <span>お預かり</span>
            <span>¥{paymentInfo.receivedAmount.toLocaleString()}</span>
          </div>
          <div className="flex justify-between">
            <span>お釣り</span>
            <span>¥{paymentInfo.changeAmount.toLocaleString()}</span>
          </div>
        </div>

        <div className={`mt-4 pt-4 border-t text-xs text-center leading-relaxed ${C.text60}`}>
          <p className={`font-bold text-sm ${C.text} mb-1`}>{clinic?.name}</p>
          <p>〒{clinic?.postalCode} {clinic?.address}</p>
          <p>TEL: {clinic?.phoneNumber}</p>
          {clinic?.invoiceRegistrationNumber ? <p>登録番号: {clinic.invoiceRegistrationNumber}</p> : null}
        </div>
      </div>
    );
  }

  // Statement (A4)
  return (
    <div className="bg-white p-8 text-sm font-sans flex flex-col gap-6 border mx-auto max-w-2xl print:max-w-none print:w-full print:border-none print:p-0">
      <div className="flex justify-between items-end border-b pb-4">
        <div>
          <h1 className="text-2xl font-bold mb-2">診療明細書</h1>
          <p className={C.text60}>No. {accounting.id}</p>
          <p className={C.text60}>発行日: {currentDate}</p>
        </div>
        <div className={`text-right text-xs ${C.text60}`}>
          <p className={`font-bold text-base ${C.text} mb-1`}>{clinic?.name}</p>
          <p>〒{clinic?.postalCode} {clinic?.address}</p>
          <p>TEL: {clinic?.phoneNumber}</p>
          {clinic?.invoiceRegistrationNumber ? <p>登録番号: {clinic.invoiceRegistrationNumber}</p> : null}
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
            <th className="py-2 text-right">単価</th>
            <th className="py-2 text-center">数量</th>
            <th className="py-2 text-right">金額</th>
          </tr>
        </thead>
        <tbody className="divide-y">
          {accounting.items.map((item) => (
            <tr key={item.id}>
              <td className="py-2">
                <div className="font-medium">{item.name}</div>
                {item.category ? <span className={`text-xs ${C.text50}`}>{item.category}</span> : null}
              </td>
              <td className="py-2 text-right">¥{item.unitPrice.toLocaleString()}</td>
              <td className="py-2 text-center">{item.quantity}</td>
              <td className="py-2 text-right">¥{(item.unitPrice * item.quantity).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="flex justify-end mt-4">
        <div className="w-64 space-y-2">
          <div className="flex justify-between font-bold border-b pb-1">
            <span>合計金額 (税込)</span>
            <span>¥{paymentInfo.totalAmount.toLocaleString()}</span>
          </div>

          <div className={`text-xs ${C.text50} space-y-1 pt-2`}>
            <div className="flex justify-between">
              <span>10%対象額</span>
              <span>¥{taxBreakdown.tax10Base.toLocaleString()} (税 ¥{taxBreakdown.tax10Amount.toLocaleString()})</span>
            </div>
            {taxBreakdown.tax8Base > 0 ? (
              <div className="flex justify-between">
                <span>8%対象額</span>
                <span>¥{taxBreakdown.tax8Base.toLocaleString()} (税 ¥{taxBreakdown.tax8Amount.toLocaleString()})</span>
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
        </div>
      </div>
    </div>
  );
}
