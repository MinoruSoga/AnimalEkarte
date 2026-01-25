import React, { useState, useMemo, useCallback } from "react";
import { TreatmentTable, TreatmentItem } from "./TreatmentTable";
import { TreatmentDetailedSummary } from "./TreatmentDetailedSummary";

interface TreatmentProps {
  isNewRecord?: boolean;
  planItems?: TreatmentItem[];
  setPlanItems?: React.Dispatch<React.SetStateAction<TreatmentItem[]>>;
  completedItems?: TreatmentItem[];
  setCompletedItems?: React.Dispatch<React.SetStateAction<TreatmentItem[]>>;
}

export function MedicalRecordTreatment({ isNewRecord = false }: TreatmentProps) {
  // Mock data for "Treatment Plan" table
  const [planItems, setPlanItems] = useState<TreatmentItem[]>(
    isNewRecord
      ? []
      : [
          {
            id: 1,
            status: "完了",
            content: "recheck(新料金)1",
            memo: "再診料099",
            insurance: true,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
          {
            id: 2,
            status: "未完了",
            content: "血尿治療Aプラン",
            memo: "血尿治療Aプラン",
            insurance: false,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
          {
            id: 3,
            status: "-",
            content: "血尿治療Aプラン＿採血A",
            memo: "",
            insurance: false,
            unitPrice: 0,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
          {
            id: 4,
            status: "-",
            content: "血尿治療Aプラン＿注射A",
            memo: "",
            insurance: false,
            unitPrice: 0,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
        ]
  );

  // Mock data for "Treatment Completed" table
  const [completedItems, setCompletedItems] = useState<TreatmentItem[]>(
    isNewRecord
      ? []
      : [
          {
            id: 1,
            content: "recheck(新料金)1",
            memo: "再診料099",
            insurance: true,
            unitPrice: 990,
            quantity: 1,
            discountRate: 0,
            discountAmount: 0,
          },
        ]
  );

  const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  const handleAddPlanItem = useCallback(() => {
    setPlanItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        status: "未完了",
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

  const handleRemovePlanItem = useCallback((id: number) => {
    setPlanItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdatePlanItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    setPlanItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }, []);

  const handleAddCompletedItem = useCallback(() => {
    setCompletedItems((prev) => [
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

  const handleRemoveCompletedItem = useCallback((id: number) => {
    setCompletedItems((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const handleUpdateCompletedItem = useCallback((id: number, field: keyof TreatmentItem, value: string | number | boolean) => {
    setCompletedItems((prev) =>
      prev.map((item) => (item.id === id ? { ...item, [field]: value } : item))
    );
  }, []);

  // Calculate Subtotal (simplified)
  const calculateSubtotal = (items: TreatmentItem[]) => {
    return items.reduce((sum, item) => {
      const price = Number(item.unitPrice) || 0;
      const qty = Number(item.quantity) || 0;
      const discount = Number(item.discountAmount) || 0;
      return sum + (price * qty - discount);
    }, 0);
  };

  const { subtotal, tax, total } = useMemo(() => {
    const planSub = calculateSubtotal(planItems);
    const completedSub = calculateSubtotal(completedItems);
    const sub = planSub + completedSub;
    const t = Math.floor(sub * 0.1);
    return { subtotal: sub, tax: t, total: sub + t };
  }, [planItems, completedItems]);

  return (
    <div className="h-[calc(100vh-220px)] min-h-[500px] flex flex-col gap-3 overflow-y-auto pr-1 pb-20">
      
      {/* Treatment Plan Section */}
      <div className="flex flex-col gap-1.5">
        <h2 className="text-sm font-bold text-[#37352F]">治療プラン</h2>
        <TreatmentTable
            items={planItems}
            onUpdate={handleUpdatePlanItem}
            onRemove={handleRemovePlanItem}
            onAddRow={handleAddPlanItem}
            showStatus={true}
        />
      </div>

      {/* Treatment Completed Section */}
      <div className="flex flex-col gap-1.5">
        <h2 className="text-sm font-bold text-[#37352F]">治療済み</h2>
        <TreatmentTable
            items={completedItems}
            onUpdate={handleUpdateCompletedItem}
            onRemove={handleRemoveCompletedItem}
            onAddRow={handleAddCompletedItem}
            showStatus={false}
        />
      </div>

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
    </div>
  );
}
