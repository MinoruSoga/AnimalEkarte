import React, { useState, useMemo, useCallback } from "react";
import { Button } from "../../../components/ui/button";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";

export function MedicalRecordBillCheck({ isNewRecord = false }: { isNewRecord?: boolean }) {
  const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  // Mock data for Bill items
  const [items, setItems] = useState<TreatmentItem[]>(
    isNewRecord
      ? []
      : [
          {
            id: 1,
            content: "recheck(新料金)1",
            memo: "再診科099",
            insurance: true,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
        ]
  );

  const handleAddItem = useCallback(() => {
    setItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        content: "",
        memo: "",
        insurance: false,
        unitPrice: 0,
        quantity: 1,
        discountRate: 0,
        discountAmount: 0,
      },
    ]);
  }, []);

  const handleRemoveItem = useCallback((id: number) => {
    setItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdateItem = useCallback((id: number, field: keyof TreatmentItem, value: any) => {
    setItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }, []);

  const { subtotal, tax, total } = useMemo(() => {
    const sub = items.reduce((sum, item) => {
      const price = Number(item.unitPrice) || 0;
      const qty = Number(item.quantity) || 0;
      const discount = Number(item.discountAmount) || 0;
      return sum + (price * qty - discount);
    }, 0);
    const t = Math.floor(sub * 0.1);
    return { subtotal: sub, tax: t, total: sub + t };
  }, [items]);

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pb-10 pr-1 relative">
      {/* Items Table */}
      <TreatmentTable
        items={items}
        onUpdate={handleUpdateItem}
        onRemove={handleRemoveItem}
        onAddRow={handleAddItem}
        showStatus={false}
      />

      {/* Summary Table */}
      <TreatmentDetailedSummary
        subtotal={subtotal}
        tax={tax}
        total={total}
        discountRate={globalDiscountRate}
        discountAmount={globalDiscountAmount}
        onUpdateDiscountRate={setGlobalDiscountRate}
        onUpdateDiscountAmount={setGlobalDiscountAmount}
      />

      {/* Action Button */}
      <div className="fixed bottom-6 right-6 z-50">
        <Button
          size="sm"
          className="bg-[#37352F] hover:bg-[#37352F]/90 text-white min-w-[100px] shadow-lg h-10 text-sm shadow-sm border-transparent"
        >
          チェック完了
        </Button>
      </div>
    </div>
  );
}
