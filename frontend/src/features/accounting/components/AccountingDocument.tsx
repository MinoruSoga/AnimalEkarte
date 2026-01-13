import { Accounting, AccountingItem } from "../types";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { useClinicInfo } from "../../clinic/hooks/useClinicInfo";

interface AccountingDocumentProps {
  type: "receipt" | "statement";
  accounting: Accounting;
  paymentInfo: any; // Using simplified type for now
}

export const AccountingDocument = ({ type, accounting, paymentInfo }: AccountingDocumentProps) => {
  const { clinicInfo } = useClinicInfo();

  const currentDate = format(new Date(), "yyyy年MM月dd日", { locale: ja });
  
  // Calculate tax breakdown
  const tax10Items = accounting.items.filter(i => i.taxRate === 0.1);
  const tax8Items = accounting.items.filter(i => i.taxRate === 0.08);
  
  const tax10Base = tax10Items.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);
  const tax8Base = tax8Items.reduce((sum, i) => sum + (i.unitPrice * i.quantity), 0);
  
  const tax10Amount = Math.floor(tax10Base * 0.1);
  const tax8Amount = Math.floor(tax8Base * 0.08);

  if (type === "receipt") {
    return (
      <div className="w-[80mm] bg-white p-4 text-sm font-sans flex flex-col gap-4 border mx-auto print:mx-0 print:border-none">
        <div className="text-center border-b pb-4">
          <h1 className="text-xl font-bold mb-2">領 収 書</h1>
          <p className="text-xs text-gray-500">{currentDate}</p>
        </div>

        <div className="space-y-1">
          <p className="text-lg font-medium border-b border-black inline-block min-w-[200px]">
            {accounting.ownerName} 様
          </p>
          <p className="text-xs text-gray-500">ペット名: {accounting.petName}</p>
        </div>

        <div className="py-6 text-center bg-gray-50 my-2">
            <p className="text-xs text-gray-500 mb-1">領収金額</p>
            <p className="text-2xl font-bold">¥{paymentInfo.billingAmount.toLocaleString()}-</p>
        </div>

        <div className="space-y-2 text-xs">
          <div className="flex justify-between border-b border-dotted pb-1">
            <span>診療費等合計</span>
            <span>¥{paymentInfo.totalAmount.toLocaleString()}</span>
          </div>
          {paymentInfo.insuranceAmount < 0 && (
            <div className="flex justify-between border-b border-dotted pb-1 text-green-700">
              <span>保険負担額</span>
              <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>
            </div>
          )}
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

        <div className="mt-4 pt-4 border-t text-xs text-center leading-relaxed text-gray-600">
          <p className="font-bold text-sm text-black mb-1">{clinicInfo.name}</p>
          <p>〒{clinicInfo.postalCode} {clinicInfo.address}</p>
          <p>TEL: {clinicInfo.phoneNumber}</p>
          {clinicInfo.registrationNumber && <p>登録番号: {clinicInfo.registrationNumber}</p>}
        </div>
      </div>
    );
  }

  // Statement (A4 or A5 usually, but responsive here)
  return (
    <div className="bg-white p-8 text-sm font-sans flex flex-col gap-6 border mx-auto max-w-2xl print:max-w-none print:w-full print:border-none print:p-0">
      <div className="flex justify-between items-end border-b pb-4">
        <div>
           <h1 className="text-2xl font-bold mb-2">診療明細書</h1>
           <p className="text-gray-600">No. {accounting.id}</p>
           <p className="text-gray-600">発行日: {currentDate}</p>
        </div>
        <div className="text-right text-xs text-gray-600">
          <p className="font-bold text-base text-black mb-1">{clinicInfo.name}</p>
          <p>〒{clinicInfo.postalCode} {clinicInfo.address}</p>
          <p>TEL: {clinicInfo.phoneNumber}</p>
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
          {accounting.items.map((item, i) => (
            <tr key={i}>
              <td className="py-2">
                <div className="font-medium">{item.name}</div>
                {item.category && <span className="text-xs text-gray-500">{item.category}</span>}
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
            
            <div className="text-xs text-gray-500 space-y-1 pt-2">
                <div className="flex justify-between">
                    <span>10%対象額</span>
                    <span>¥{tax10Base.toLocaleString()} (税 ¥{tax10Amount.toLocaleString()})</span>
                </div>
                {tax8Base > 0 && (
                    <div className="flex justify-between">
                        <span>8%対象額</span>
                        <span>¥{tax8Base.toLocaleString()} (税 ¥{tax8Amount.toLocaleString()})</span>
                    </div>
                )}
            </div>

            {paymentInfo.insuranceAmount < 0 && (
                <div className="flex justify-between text-green-700 border-b pb-1 pt-2">
                    <span>保険適用</span>
                    <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>
                </div>
            )}
            
            <div className="flex justify-between font-bold text-lg pt-2 border-t border-black">
                <span>請求金額</span>
                <span>¥{paymentInfo.billingAmount.toLocaleString()}</span>
            </div>
        </div>
      </div>
    </div>
  );
};
